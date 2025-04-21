package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

func CreateSafeProposalCmd() *cobra.Command {
	proposalCmd := &cobra.Command{
		Use:   "proposal",
		Short: "Manage proposals for a Safe",
		Long:  `Manage Safe proposals — create new ones, list existing ones, approve pending ones, and execute approved proposals.`,
	}

	proposalCmd.AddCommand(createApproveProposalsCmd())
	proposalCmd.SetOut(os.Stdout)

	return proposalCmd
}

func createApproveProposalsCmd() *cobra.Command {
	var (
		hash     string
		keyfile  string
		password string
	)
	approveProposalsCmd := &cobra.Command{
		Use:   "approve",
		Short: "approve a proposal for a safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v1/chains/%s/transactions/%s/confirmations/", chainID.String(), hash)
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			}

			err = SignAndApproveProposal(hash, key, safeAPIURL)
			if err != nil {
				return fmt.Errorf("error signing and approving proposal: %v", err)
			}
			cmd.Println("Successfully signed and approved proposal")
			return nil
		},
	}
	approveProposalsCmd.Flags().StringVar(&hash, "hash", "h", "Safe tx hash")
	approveProposalsCmd.Flags().StringVarP(&keyfile, "keyfile", "k", "", "Path to the keystore file")
	approveProposalsCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the keystore file")
	approveProposalsCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	approveProposalsCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	approveProposalsCmd.MarkFlagRequired("keyfile")
	approveProposalsCmd.MarkFlagRequired("hash")
	approveProposalsCmd.MarkFlagRequired("rpc")

	return approveProposalsCmd
}
