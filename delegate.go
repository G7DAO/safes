package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

var rpcURL string
var safeAPIURL string

func CreateDelegateCmd() *cobra.Command {
	delegateCmd := &cobra.Command{
		Use:   "delegate",
		Short: "Manage delegates for a Safe",
		Long:  `Manage delegates for a Safe by adding, removing, or retrieving existing ones.`,
	}

	delegateCmd.AddCommand(createAddDelegateCmd())
	delegateCmd.AddCommand(createListDelegatesCmd())
	delegateCmd.AddCommand(createRemoveDelegateCmd()) // Add this line

	return delegateCmd
}

func createAddDelegateCmd() *cobra.Command {
	var (
		safe     string
		delegate string
		label    string
		keyfile  string
		password string
	)

	addDelegateCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new delegate to a Safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			if !common.IsHexAddress(delegate) {
				return fmt.Errorf("invalid delegate address: %s", delegate)
			}
			if label == "" {
				return fmt.Errorf("label is required")
			}

			if keyfile == "" {
				return fmt.Errorf("--keyfile not specified (this should be a path to an Ethereum account keystore file)")
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
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v2/chains/%d/delegates/", chainID.Int64())
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			} else {
				fmt.Println("Using custom safe-api URL: ", safeAPIURL)
			}

			err = AddDelegate(safe, delegate, label, chainID, key, safeAPIURL)
			if err != nil {
				cmd.Printf("Error adding delegate: %v\n", err)
				return fmt.Errorf("error adding delegate: %v", err)
			}
			cmd.Printf("Successfully added delegate %s for Safe %s\n", delegate, safe)
			return nil
		},
	}

	addDelegateCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	addDelegateCmd.Flags().StringVar(&delegate, "delegate", "", "Delegate address")
	addDelegateCmd.Flags().StringVarP(&label, "label", "l", "", "Label for the delegate")
	addDelegateCmd.Flags().StringVarP(&keyfile, "keyfile", "k", "", "Path to the keystore file")
	addDelegateCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the keystore file")
	addDelegateCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	addDelegateCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	addDelegateCmd.MarkFlagRequired("keyfile")
	addDelegateCmd.MarkFlagRequired("safe")
	addDelegateCmd.MarkFlagRequired("delegate")

	return addDelegateCmd
}

func createListDelegatesCmd() *cobra.Command {
	var (
		safe      string
		delegate  string
		delegator string
		label     string
		limit     int
		offset    int
	)

	listDelegatesCmd := &cobra.Command{
		Use:   "list",
		Short: "List delegates for a Safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				cmd.PrintErrf("Error connecting to RPC: %v\n", err)
				return fmt.Errorf("error connecting to RPC: %v", err)
			}
			chainID, err := client.ChainID(context.Background())
			if err != nil {
				cmd.PrintErrf("Error retrieving chain ID: %v\n", err)
				return fmt.Errorf("error retrieving chain ID: %v", err)
			}

			if safeAPIURL == "" {
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v2/chains/%d/delegates/", chainID.Int64())
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			}

			delegates, err := GetDelegates(safe, delegate, delegator, label, limit, offset, chainID, safeAPIURL)
			if err != nil {
				return fmt.Errorf("error retrieving delegates: %v", err)
			}
			if len(delegates) == 0 {
				return fmt.Errorf("no delegates found")
			} else {
				for _, d := range delegates {
					cmd.Printf("Safe: %s, Delegate: %s, Delegator: %s, Label: %s\n", d.Safe, d.Delegate, d.Delegator, d.Label)
				}
			}
			return nil
		},
	}

	listDelegatesCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	listDelegatesCmd.Flags().StringVar(&delegate, "delegate", "", "Filter by delegate address")
	listDelegatesCmd.Flags().StringVar(&delegator, "delegator", "", "Filter by delegator address")
	listDelegatesCmd.Flags().StringVarP(&label, "label", "l", "", "Filter by label")
	listDelegatesCmd.Flags().IntVar(&limit, "limit", 0, "Limit the number of results")
	listDelegatesCmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	listDelegatesCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	listDelegatesCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	listDelegatesCmd.MarkFlagRequired("rpc")
	listDelegatesCmd.MarkFlagRequired("safe")

	return listDelegatesCmd
}

func createRemoveDelegateCmd() *cobra.Command {
	var (
		safe     string
		delegate string
		keyfile  string
		password string
	)

	removeDelegateCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a delegate",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			if !common.IsHexAddress(delegate) {
				return fmt.Errorf("invalid delegate address: %s", delegate)
			}

			if keyfile == "" {
				return fmt.Errorf("--keyfile not specified (this should be a path to an Ethereum account keystore file)")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			checksumSafe := common.HexToAddress(safe).Hex()
			checksumDelegate := common.HexToAddress(delegate).Hex()

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
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v2/chains/%d/delegates", chainID.Int64())
				fmt.Println("safe-api is not set, using default: ", safeAPIURL)
			}

			err = RemoveDelegate(checksumSafe, checksumDelegate, chainID, key, safeAPIURL)
			if err != nil {
				return fmt.Errorf("error removing delegate: %v", err)
			}
			cmd.Printf("Successfully removed delegate %s from Safe %s\n", checksumDelegate, checksumSafe)
			return nil
		},
	}

	removeDelegateCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	removeDelegateCmd.Flags().StringVar(&delegate, "delegate", "", "Delegate address to remove")
	removeDelegateCmd.Flags().StringVarP(&keyfile, "keyfile", "k", "", "Path to the keystore file")
	removeDelegateCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the keystore file")
	removeDelegateCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	removeDelegateCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	removeDelegateCmd.MarkFlagRequired("safe")
	removeDelegateCmd.MarkFlagRequired("keyfile")
	removeDelegateCmd.MarkFlagRequired("rpc")
	removeDelegateCmd.MarkFlagRequired("delegate")

	return removeDelegateCmd
}
