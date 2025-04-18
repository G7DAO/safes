package main

import (
	"context"
	"fmt"
	"os"

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

	proposalCmd.AddCommand(createListProposalsCmd())
	proposalCmd.SetOut(os.Stdout)

	return proposalCmd
}

func createListProposalsCmd() *cobra.Command {
	var safe string
	listProposalsCmd := &cobra.Command{
		Use:   "list",
		Short: "List proposals for a safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				return fmt.Errorf("failed to connect to the Ethereum client: %v", err)
			}

			chainID, err := client.ChainID(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get chain ID: %v", err)
			}
			if safeAPIURL == "" {
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v1/chains/%s/safes/%s/multisig-transactions/raw", chainID.String(), safe)
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			}

			proposals, err := GetProposals(safeAPIURL)
			if err != nil {
				return fmt.Errorf("error retrieving delegates: %v", err)
			}
			if len(proposals) == 0 {
				return fmt.Errorf("no proposals found")
			} else {
				var count int64
				for _, tx := range proposals {
					count++
					cmd.Println("================================================================================================")
					cmd.Printf("Proposal count #%d\n", count)
					cmd.Printf("Safe Address: %s\n", tx.Safe)
					cmd.Printf("To: %s\n", tx.To)
					cmd.Printf("Value: %s\n", tx.Value)
					cmd.Printf("Data:            %s\n", nullableString(tx.Data))
					cmd.Printf("Operation: %d\n", tx.Operation)
					cmd.Printf("Gas Token: %s\n", tx.GasToken)
					cmd.Printf("SafeTxGas: %d\n", tx.SafeTxGas)
					cmd.Printf("BaseGas: %d\n", tx.BaseGas)
					cmd.Printf("Gas Price: %s\n", tx.GasPrice)
					cmd.Printf("Refund Receiver: %s\n", tx.RefundReceiver)
					cmd.Printf("Nonce: %d\n", tx.Nonce)
					cmd.Printf("Execution Date: %s\n", tx.ExecutionDate)
					cmd.Printf("Submission Date: %s\n", tx.SubmissionDate)
					cmd.Printf("Modified: %s\n", tx.Modified)
					cmd.Printf("Block Number: %d\n", tx.BlockNumber)
					cmd.Printf("Transaction Hash: %s\n", tx.TransactionHash)
					cmd.Printf("SafeTxHash: %s\n", tx.SafeTxHash)
					cmd.Printf("Proposer: %s\n", tx.Proposer)
					cmd.Printf("Executor: %s\n", tx.Executor)
					cmd.Printf("Is Executed: %v\n", tx.IsExecuted)
					cmd.Printf("Is Successful: %v\n", tx.IsSuccessful)
					cmd.Printf("ETH Gas Price: %s\n", tx.EthGasPrice)
					cmd.Printf("Max Fee Per Gas: %s\n", tx.MaxFeePerGas)
					cmd.Printf("Max Priority Fee Per Gas: %s\n", tx.MaxPriorityFeePerGas)
					cmd.Printf("Gas Used: %d\n", tx.GasUsed)
					cmd.Printf("Fee: %s\n", tx.Fee)
					cmd.Printf("Origin: %s\n", tx.Origin)
					cmd.Printf("Data Decoded:    %s\n", tx.DataDecoded)
					cmd.Printf("Confirmations Required: %d\n", tx.ConfirmationsRequired)
					cmd.Printf("Trusted: %v\n", tx.Trusted)
					cmd.Printf("Signatures: %s\n", tx.Signatures)

					cmd.Println("\nConfirmations:")
					if len(tx.Confirmations) == 0 {
						cmd.Println("  None")
					} else {
						for i, c := range tx.Confirmations {
							cmd.Printf("  [%d] Owner: %s\n", i+1, c.Owner)
							cmd.Printf("      Submission Date: %s\n", c.SubmissionDate)
							if c.TransactionHash != nil {
								cmd.Printf("      Transaction Hash: %s\n", nullableString(c.TransactionHash))
							} else {
								cmd.Println("      Transaction Hash: <nil>")
							}
							cmd.Printf("      Signature: %s\n", c.Signature)
							cmd.Printf("      Signature Type: %s\n", c.SignatureType)
						}
					}
					cmd.Println("================================================================================================")
				}
			}
			return nil
		},
	}
	listProposalsCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	listProposalsCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	listProposalsCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	listProposalsCmd.MarkFlagRequired("safe")
	listProposalsCmd.MarkFlagRequired("rpc")

	return listProposalsCmd
}
