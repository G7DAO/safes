package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/G7DAO/safes/bindings/Safe"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

func CreateSafeProposalCmd() *cobra.Command {
	proposalCmd := &cobra.Command{
		Use:   "proposal",
		Short: "Manage proposals for a Safe",
		Long:  `Manage Safe proposals — create new ones, list existing ones, approve pending ones, and execute approved proposals.`,
	}

	proposalCmd.AddCommand(createSafeProposalCmd())
	proposalCmd.SetOut(os.Stdout)

	return proposalCmd
}

func createSafeProposalCmd() *cobra.Command {
	var (
		calldata          string
		safe              string
		safeOperationType uint8
		to                string
		value             string
		keyfile           string
		password          string
	)

	createProposalCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new proposal for a Safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if safe == "" {
				return fmt.Errorf("--safe not specified")
			} else if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			if to == "" {
				return fmt.Errorf("--to not specified")
			} else if !common.IsHexAddress(to) {
				return fmt.Errorf("invalid to address: %s", to)
			}
			if calldata != "" {
				if !strings.HasPrefix(calldata, "0x") {
					calldata = "0x" + calldata
				}
				if _, err := hex.DecodeString(strings.TrimPrefix(calldata, "0x")); err != nil {
					return fmt.Errorf("invalid calldata hex: %v", err)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			toAddr := common.HexToAddress(to).Hex()
			safeAddr := common.HexToAddress(safe).Hex()

			key, keyErr := KeyFromFile(keyfile, password)
			if keyErr != nil {
				return keyErr
			}

			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				return fmt.Errorf("failed to connect to the Ethereum client: %v", err)
			}

			chainID, err := client.ChainID(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get chain ID: %v", err)
			}

			parsedValue := new(big.Int)
			if _, ok := parsedValue.SetString(value, 10); !ok {
				return fmt.Errorf("invalid value: %s", value)
			}

			if safeAPIURL == "" {
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v1/chains/%s/transactions/%s/propose", chainID.String(), safeAddr)
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			} else {
				fmt.Println("Using custom safe-api URL: ", safeAPIURL)
			}

			err = CreateSafeProposal(common.HexToAddress(safeAddr), toAddr, parsedValue.String(), calldata, Safe.SafeOperationType(safeOperationType), chainID, key, client, safeAPIURL)
			if err != nil {
				cmd.Printf("Error creating proposal: %v\n", err)
				return fmt.Errorf("error creating proposal: %v", err)
			}

			fmt.Println("Proposal submitted to:", safeAPIURL)
			return nil
		},
	}

	createProposalCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	createProposalCmd.Flags().StringVar(&to, "to", "", "Recipient address")
	createProposalCmd.Flags().StringVarP(&keyfile, "keyfile", "k", "", "Path to the keystore file")
	createProposalCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the keystore file")
	createProposalCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	createProposalCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	createProposalCmd.Flags().StringVar(&value, "value", "", "Value to send with the transaction")
	createProposalCmd.Flags().StringVar(&calldata, "calldata", "", "Hex-encoded ABI calldata to be sent with the transaction (e.g., function selector and arguments).")
	createProposalCmd.Flags().Uint8Var(&safeOperationType, "safe-operation", 0, "Safe operation type: 0 (Call) or 1 (DelegateCall)")
	createProposalCmd.MarkFlagRequired("keyfile")
	createProposalCmd.MarkFlagRequired("safe")

	return createProposalCmd
}
