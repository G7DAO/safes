package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Confirmation struct {
	Owner           string  `json:"owner"`
	SubmissionDate  string  `json:"submissionDate"`
	TransactionHash *string `json:"transactionHash"`
	Signature       string  `json:"signature"`
	SignatureType   string  `json:"signatureType"`
}

type DataDecoded struct {
	Method     string      `json:"method"`
	Parameters []Parameter `json:"parameters"`
	Accuracy   string      `json:"accuracy"`
}

type Parameter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type ProposalResponse struct {
	Safe                  string         `json:"safe"`
	To                    string         `json:"to"`
	Value                 string         `json:"value"`
	Data                  *string        `json:"data"`
	Operation             int            `json:"operation"`
	GasToken              string         `json:"gasToken"`
	SafeTxGas             int            `json:"safeTxGas"`
	BaseGas               int            `json:"baseGas"`
	GasPrice              string         `json:"gasPrice"`
	RefundReceiver        string         `json:"refundReceiver"`
	Nonce                 int            `json:"nonce"`
	ExecutionDate         string         `json:"executionDate"`
	SubmissionDate        string         `json:"submissionDate"`
	Modified              string         `json:"modified"`
	BlockNumber           int            `json:"blockNumber"`
	TransactionHash       string         `json:"transactionHash"`
	SafeTxHash            string         `json:"safeTxHash"`
	Proposer              string         `json:"proposer"`
	Executor              string         `json:"executor"`
	IsExecuted            bool           `json:"isExecuted"`
	IsSuccessful          bool           `json:"isSuccessful"`
	EthGasPrice           string         `json:"ethGasPrice"`
	MaxFeePerGas          string         `json:"maxFeePerGas"`
	MaxPriorityFeePerGas  string         `json:"maxPriorityFeePerGas"`
	GasUsed               int            `json:"gasUsed"`
	Fee                   string         `json:"fee"`
	Origin                string         `json:"origin"`
	DataDecoded           *DataDecoded   `json:"dataDecoded"`
	ConfirmationsRequired int            `json:"confirmationsRequired"`
	Confirmations         []Confirmation `json:"confirmations"`
	Trusted               bool           `json:"trusted"`
	Signatures            string         `json:"signatures"`
}

func GetProposals(apiURL string) ([]ProposalResponse, error) {
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var response struct {
		Count    int                `json:"count"`
		Next     *string            `json:"next"`
		Previous *string            `json:"previous"`
		Results  []ProposalResponse `json:"results"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	return response.Results, nil
}

func nullableString(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
