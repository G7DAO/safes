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

	proposalCmd.AddCommand(createExecuteProposalCmd())
	proposalCmd.SetOut(os.Stdout)

	return proposalCmd
}

func createExecuteProposalCmd() *cobra.Command {
	var (
		safe     string
		hash     string
		executor string
		keyfile  string
		password string
	)
	executeProposalCmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute proposal for a safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			if !IsValidHex(hash) {
				return fmt.Errorf("invalid hash")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if safeAPIURL == "" {
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v1/chains/%s/transactions/%s", chainID.String(), hash)
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			} else {
				fmt.Println("Using custom safe-api URL: ", safeAPIURL)
			}
			err = ExecuteProposalCmd(safe, safeAPIURL, key, chainID, client)
			if err != nil {
				cmd.Printf("Error Executing proposal: %v\n", err)
				return fmt.Errorf("error executing proposal: %v", err)
			}

			fmt.Println("Proposal executed to:", safeAPIURL)
			return nil
		},
	}

	executeProposalCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	executeProposalCmd.Flags().StringVar(&hash, "hash", "h", "Safe tx hash")
	executeProposalCmd.Flags().StringVarP(&keyfile, "keyfile", "k", "", "Path to the keystore file")
	executeProposalCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the keystore file")
	executeProposalCmd.Flags().StringVar(&executor, "executor", "", "Executor Address")
	executeProposalCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	executeProposalCmd.MarkFlagRequired("keyfile")
	executeProposalCmd.MarkFlagRequired("rpc")
	executeProposalCmd.MarkFlagRequired("hash")
	executeProposalCmd.MarkFlagRequired("safe")
	executeProposalCmd.MarkFlagRequired("executor")

	return executeProposalCmd
}
