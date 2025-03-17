package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

func CreateProposalCmd() *cobra.Command {
	proposalCmd := &cobra.Command{
		Use:   "proposal",
		Short: "List proposals for a Safe",
		Long:  `List all queued proposals for a Safe, including pending transactions and their status.`,
	}

	proposalCmd.AddCommand(createListProposalCmd())
	return proposalCmd
}

func createListProposalCmd() *cobra.Command {
	var safe string

	listProposalCmd := &cobra.Command{
		Use:   "list",
		Short: "List all queued proposals for a Safe",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(safe) {
				return fmt.Errorf("invalid safe address: %s", safe)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				return fmt.Errorf("failed to connect to RPC: %v", err)
			}

			chainID, err := client.ChainID(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get chain ID: %v", err)
			}

			if safeAPIURL == "" {
				safeAPIURL = fmt.Sprintf("https://safe-client.safe.global/v1/chains/%d/safes/%s/transactions/queued", chainID.Int64(), safe)
				fmt.Println("safe-api is not set, using default:", safeAPIURL)
			}

			var cursor *string
			var allTransactions []TransactionItem
			httpClient := &http.Client{Timeout: 60 * time.Second}

			for {
				url := safeAPIURL
				if cursor != nil {
					url = fmt.Sprintf("%s?cursor=%s", safeAPIURL, *cursor)
				}

				resp, err := httpClient.Get(url)
				if err != nil {
					return fmt.Errorf("failed to fetch proposals: %v", err)
				}

				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				}

				var proposalResponse SafeProposalResponse
				if err := json.NewDecoder(resp.Body).Decode(&proposalResponse); err != nil {
					resp.Body.Close()
					return fmt.Errorf("failed to decode response: %v", err)
				}
				resp.Body.Close()

				for _, item := range proposalResponse.Results {
					if item.Type == "TRANSACTION" && item.Transaction != nil {
						allTransactions = append(allTransactions, item)
					}
				}

				if proposalResponse.Next == nil {
					break
				}
				cursor = proposalResponse.Next
			}

			if len(allTransactions) == 0 {
				return fmt.Errorf("no proposals found")
			}

			for _, item := range allTransactions {
				tx := item.Transaction
				cmd.Printf("Transaction ID: %s\n", tx.ID)
				cmd.Printf("Status: %s\n", tx.TxStatus)
				cmd.Printf("Type: %s\n", tx.TxInfo.Type)
				cmd.Printf("Nonce: %d\n", tx.ExecutionInfo.Nonce)
				cmd.Printf("Confirmations: %d/%d\n",
					tx.ExecutionInfo.ConfirmationsSubmitted,
					tx.ExecutionInfo.ConfirmationsRequired)

				if len(tx.ExecutionInfo.MissingSigners) > 0 {
					cmd.Println("Missing Signatures from:")
					for _, signer := range tx.ExecutionInfo.MissingSigners {
						cmd.Printf("  - %s\n", signer.Value)
					}
				}

				if tx.SafeAppInfo != nil {
					cmd.Printf("Created by: %s\n", tx.SafeAppInfo.Name)
				}

				cmd.Println("---")
			}

			return nil
		},
	}

	listProposalCmd.Flags().StringVar(&safe, "safe", "", "Safe address")
	listProposalCmd.Flags().StringVar(&rpcURL, "rpc", "", "RPC URL to retrieve chain ID")
	listProposalCmd.Flags().StringVar(&safeAPIURL, "safe-api", "", "Override default Safe API URL")
	listProposalCmd.MarkFlagRequired("safe")
	listProposalCmd.MarkFlagRequired("rpc")

	return listProposalCmd
}

// Response types
type Address struct {
	Value   string  `json:"value"`
	Name    *string `json:"name"`
	LogoUri *string `json:"logoUri"`
}

type TransferInfo struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type TxInfo struct {
	Type             string        `json:"type"`
	HumanDescription *string       `json:"humanDescription"`
	Sender           *Address      `json:"sender,omitempty"`
	Recipient        *Address      `json:"recipient,omitempty"`
	To               *Address      `json:"to,omitempty"`
	Direction        string        `json:"direction,omitempty"`
	TransferInfo     *TransferInfo `json:"transferInfo,omitempty"`
	DataSize         string        `json:"dataSize,omitempty"`
	Value            string        `json:"value"`
	MethodName       *string       `json:"methodName"`
	ActionCount      *string       `json:"actionCount"`
	IsCancellation   bool          `json:"isCancellation"`
}

type ExecutionInfo struct {
	Type                   string    `json:"type"`
	Nonce                  int       `json:"nonce"`
	ConfirmationsRequired  int       `json:"confirmationsRequired"`
	ConfirmationsSubmitted int       `json:"confirmationsSubmitted"`
	MissingSigners         []Address `json:"missingSigners"`
}

type SafeAppInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	LogoUri string `json:"logoUri"`
}

type Transaction struct {
	TxInfo        TxInfo        `json:"txInfo"`
	ID            string        `json:"id"`
	Timestamp     int64         `json:"timestamp"`
	TxStatus      string        `json:"txStatus"`
	ExecutionInfo ExecutionInfo `json:"executionInfo"`
	SafeAppInfo   *SafeAppInfo  `json:"safeAppInfo"`
	TxHash        *string       `json:"txHash"`
}

type TransactionItem struct {
	Type         string       `json:"type"`
	Label        string       `json:"label,omitempty"`
	Transaction  *Transaction `json:"transaction,omitempty"`
	ConflictType string       `json:"conflictType,omitempty"`
}

type SafeProposalResponse struct {
	Count    int               `json:"count"`
	Next     *string           `json:"next"`
	Previous *string           `json:"previous"`
	Results  []TransactionItem `json:"results"`
}
