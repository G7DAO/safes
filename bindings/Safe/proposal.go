package Safe

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func CreateApproveProposalCommand() *cobra.Command {
	var safeAddress string
	var proposalID string
	var keyfile string
	var password string
	var rpcURL string
	var nonce string
	var gasPrice string
	var maxFeePerGas string
	var maxPriorityFeePerGas string
	var gasLimit uint64
	var safeNonce string
	var safeOperationType uint8

	approveProposalCmd := &cobra.Command{
		Use:   "approve-proposal",
		Short: "Approve a proposal on a Safe",
		Long:  `Approve a specific proposal on a Safe by providing the Safe address and proposal ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if safeAddress == "" || proposalID == "" || keyfile == "" {
				return fmt.Errorf("--safe, --proposal-id, and --keyfile are required")
			}

			// Call the function to approve the proposal
			err := ApproveProposal(safeAddress, proposalID, keyfile, password, rpcURL, nonce, gasPrice, maxFeePerGas, maxPriorityFeePerGas, gasLimit, safeNonce, safeOperationType)
			if err != nil {
				return err
			}

			fmt.Println("Proposal approved successfully")
			return nil
		},
	}

	// Add flags
	approveProposalCmd.Flags().StringVar(&safeAddress, "safe", "", "Address of the Safe")
	approveProposalCmd.Flags().StringVar(&proposalID, "proposal-id", "", "ID of the proposal to approve")
	approveProposalCmd.Flags().StringVar(&keyfile, "keyfile", "", "Path to the keystore file to use for the transaction")
	approveProposalCmd.Flags().StringVar(&password, "password", "", "Password to use to unlock the keystore (if not specified, you will be prompted for the password when the command executes)")
	approveProposalCmd.Flags().StringVar(&rpcURL, "rpc", "", "URL of the JSONRPC API to use")
	approveProposalCmd.Flags().StringVar(&nonce, "nonce", "", "Nonce to use for the transaction")
	approveProposalCmd.Flags().StringVar(&gasPrice, "gas-price", "", "Gas price to use for the transaction")
	approveProposalCmd.Flags().StringVar(&maxFeePerGas, "max-fee-per-gas", "", "Maximum fee per gas to use for the (EIP-1559) transaction")
	approveProposalCmd.Flags().StringVar(&maxPriorityFeePerGas, "max-priority-fee-per-gas", "", "Maximum priority fee per gas to use for the (EIP-1559) transaction")
	approveProposalCmd.Flags().Uint64Var(&gasLimit, "gas-limit", 0, "Gas limit for the transaction")
	approveProposalCmd.Flags().StringVar(&safeNonce, "safe-nonce", "", "Safe nonce overrider for the transaction (optional)")
	approveProposalCmd.Flags().Uint8Var(&safeOperationType, "safe-operation", 0, "Safe operation type: 0 (Call) or 1 (DelegateCall)")

	// Mark flags as required
	approveProposalCmd.MarkFlagRequired("safe")
	approveProposalCmd.MarkFlagRequired("proposal-id")
	approveProposalCmd.MarkFlagRequired("keyfile")

	return approveProposalCmd
}

func ApproveProposal(safeAddress, proposalID, keyfile, password, rpcURL, nonce, gasPrice, maxFeePerGas, maxPriorityFeePerGas string, gasLimit uint64, safeNonce string, safeOperationType uint8) error {
	// Create an Ethereum client
	client, err := NewClient(rpcURL)
	if errors.Is(err, ErrNoRPCURL) {
		return fmt.Errorf("no RPC URL provided -- please pass an RPC URL from the command line or set the SAFE_RPC_URL environment variable")
	} else if err != nil {
		return fmt.Errorf("failed to create Ethereum client: %v", err)
	}

	// Load the Safe contract
	safeContract, err := NewSafe(common.HexToAddress(safeAddress), client)
	if err != nil {
		return fmt.Errorf("failed to load Safe contract: %v", err)
	}

	// Read the keyfile
	keyJSON, err := os.ReadFile(keyfile)
	if err != nil {
		return fmt.Errorf("failed to read keyfile: %v", err)
	}

	// Unlock the keyfile
	_, err = keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt keyfile: %v", err)
	}

	// Create a transactor
	auth, err := bind.NewTransactorWithChainID(bytes.NewReader(keyJSON), password, big.NewInt(1)) // Replace `1` with the correct chain ID
	if err != nil {
		return fmt.Errorf("failed to create transactor: %v", err)
	}

	// Set gas-related parameters
	if gasLimit > 0 {
		auth.GasLimit = gasLimit
	}
	if gasPrice != "" {
		gasPriceWei, err := ParseEther(gasPrice) // Use the custom ParseEther function
		if err != nil {
			return fmt.Errorf("invalid gas price: %v", err)
		}
		auth.GasPrice = gasPriceWei
	}
	if maxFeePerGas != "" && maxPriorityFeePerGas != "" {
		maxFeeWei, err := ParseEther(maxFeePerGas) // Use the custom ParseEther function
		if err != nil {
			return fmt.Errorf("invalid max fee per gas: %v", err)
		}
		maxPriorityFeeWei, err := ParseEther(maxPriorityFeePerGas) // Use the custom ParseEther function
		if err != nil {
			return fmt.Errorf("invalid max priority fee per gas: %v", err)
		}
		auth.GasFeeCap = maxFeeWei
		auth.GasTipCap = maxPriorityFeeWei
	}

	// Set nonce
	if nonce != "" {
		nonceInt, err := strconv.ParseUint(nonce, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid nonce: %v", err)
		}
		auth.Nonce = big.NewInt(int64(nonceInt))
	}

	// Approve the proposal (using the correct method, e.g., approveHash)
	tx, err := safeContract.ApproveHash(auth, common.HexToHash(proposalID))
	if err != nil {
		return fmt.Errorf("failed to approve proposal: %v", err)
	}

	fmt.Printf("Proposal approved, transaction hash: %s\n", tx.Hash().Hex())
	return nil
}

// ParseEther converts a string representation of Ether (e.g., "1.5") into wei (*big.Int).
func ParseEther(ether string) (*big.Int, error) {
	// Split the input into whole and fractional parts
	parts := strings.Split(ether, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid ether value: %s", ether)
	}

	// Parse the whole part
	whole, ok := new(big.Int).SetString(parts[0], 10)
	if !ok {
		return nil, fmt.Errorf("invalid whole part: %s", parts[0])
	}
	whole = new(big.Int).Mul(whole, big.NewInt(1e18))

	// If there's no fractional part, return the whole part
	if len(parts) == 1 {
		return whole, nil
	}

	// Parse the fractional part
	fractional := parts[1]
	if len(fractional) > 18 {
		return nil, fmt.Errorf("fractional part too long: %s", fractional)
	}
	fractional += strings.Repeat("0", 18-len(fractional)) // Pad with zeros
	fractionalInt, ok := new(big.Int).SetString(fractional, 10)
	if !ok {
		return nil, fmt.Errorf("invalid fractional part: %s", fractional)
	}

	// Add the whole and fractional parts
	return new(big.Int).Add(whole, fractionalInt), nil
}
