package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"math/big"

	"github.com/G7DAO/seer/bindings/GnosisSafe"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Signature struct {
	Signer struct {
		Value string `json:"value"`
	} `json:"signer"`
	Signature string `json:"signature"`
}

type TxData struct {
	HexData string `json:"hexData"`
	To      struct {
		Value string `json:"value"`
	} `json:"to"`
	Value     string `json:"value"`
	Operation int    `json:"operation"`
}

type ExecuteTransactionParams struct {
	To             string
	Value          string
	Data           string
	Operation      uint8
	SafeTxGas      string
	BaseGas        string
	GasPrice       string
	GasToken       string
	RefundReceiver string
	Signatures     string
}

type DetailedExecutionInfo struct {
	ConfirmationsRequired int         `json:"confirmationsRequired"`
	Confirmations         []Signature `json:"confirmations"`
	SafeTxGas             string      `json:"safeTxGas"`
	BaseGas               string      `json:"baseGas"`
	GasPrice              string      `json:"gasPrice"`
	GasToken              string      `json:"gasToken"`
	RefundReceiver        struct {
		Value string `json:"value"`
	} `json:"refundReceiver"`
}

type TxInfoResponse struct {
	DetailedExecutionInfo DetailedExecutionInfo `json:"detailedExecutionInfo"`
	TxData                TxData                `json:"txData"`
}

func ExecuteProposalCmd(safe, apiURL string, key *keystore.Key, chainID *big.Int, client *ethclient.Client) error {
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("invalid Safe API URL: %w", err)
	}

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return fmt.Errorf("failed to fetch transaction info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var txInfo TxInfoResponse
	if err := json.Unmarshal(body, &txInfo); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	confirmations := txInfo.DetailedExecutionInfo.Confirmations
	required := txInfo.DetailedExecutionInfo.ConfirmationsRequired

	if len(confirmations) < required {
		return fmt.Errorf("not enough confirmations: got %d, need %d", len(confirmations), required)
	}

	sort.SliceStable(confirmations, func(i, j int) bool {
		return strings.ToLower(confirmations[i].Signer.Value) < strings.ToLower(confirmations[j].Signer.Value)
	})

	var combined string
	for _, conf := range confirmations {
		combined += strings.TrimPrefix(conf.Signature, "0x")
	}
	finalSigs := "0x" + combined

	value := new(big.Int)
	value.SetString(txInfo.TxData.Value, 10)

	safeTxGas := new(big.Int)
	safeTxGas.SetString(txInfo.DetailedExecutionInfo.SafeTxGas, 10)

	baseGas := new(big.Int)
	baseGas.SetString(txInfo.DetailedExecutionInfo.BaseGas, 10)

	gasPrice := new(big.Int)
	gasPrice.SetString(txInfo.DetailedExecutionInfo.GasPrice, 10)

	data, err := hex.DecodeString(strings.TrimPrefix(txInfo.TxData.HexData, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode tx data: %w", err)
	}

	signatures, err := hex.DecodeString(strings.TrimPrefix(finalSigs, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode signatures: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(key.PrivateKey, chainID)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	safeInstance, err := GnosisSafe.NewGnosisSafe(common.HexToAddress(safe), client)
	if err != nil {
		return fmt.Errorf("failed to create GnosisSafe instance: %w", err)
	}

	tx, err := safeInstance.ExecTransaction(
		auth,
		common.HexToAddress(txInfo.TxData.To.Value),
		value,
		data,
		uint8(txInfo.TxData.Operation),
		safeTxGas,
		baseGas,
		gasPrice,
		common.HexToAddress(txInfo.DetailedExecutionInfo.GasToken),
		common.HexToAddress(txInfo.DetailedExecutionInfo.RefundReceiver.Value),
		signatures,
	)
	if err != nil {
		return fmt.Errorf("ExecTransaction failed: %w", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	return nil
}

func IsValidHex(s string) bool {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}

	if len(s) == 0 {
		return false
	}

	if len(s)%2 != 0 {
		return false
	}

	_, err := hex.DecodeString(s)
	return err == nil
}
