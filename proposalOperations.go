package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/G7DAO/safes/bindings/Safe"
	"github.com/G7DAO/seer/bindings/GnosisSafe"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CreateSafeProposal(safeAddress common.Address, to string, value string, calldata string, safeOperationType Safe.SafeOperationType, chainID *big.Int, key *keystore.Key, client *ethclient.Client, safeApi string) error {
	safeInstance, err := GnosisSafe.NewGnosisSafe(safeAddress, client)
	if err != nil {
		return fmt.Errorf("failed to create GnosisSafe instance: %w", err)
	}

	nonce, err := safeInstance.Nonce(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("failed to fetch nonce: %w", err)
	}

	txData := Safe.SafeTransactionData{
		To:             to,
		Value:          value,
		Data:           calldata,
		Operation:      safeOperationType,
		SafeTxGas:      0,
		BaseGas:        0,
		GasPrice:       "0",
		GasToken:       Safe.NativeTokenAddress,
		RefundReceiver: Safe.NativeTokenAddress,
		Nonce:          nonce,
	}

	// Compute the hash of the transaction for signing
	safeTxHash, err := Safe.CalculateSafeTxHash(safeAddress, txData, chainID)
	if err != nil {
		return fmt.Errorf("failed to calculate SafeTxHash: %w", err)
	}

	// Sign the hash with the user's private key
	signature, err := crypto.Sign(safeTxHash.Bytes(), key.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign SafeTxHash: %w", err)
	}

	// Adjust the V value for Ethereum signature replay protection
	signature[64] += 27
	senderSignature := "0x" + common.Bytes2Hex(signature)

	requestBody := map[string]interface{}{
		"to":             txData.To,
		"value":          txData.Value,
		"data":           "0x" + txData.Data,
		"operation":      int(txData.Operation),
		"safeTxGas":      fmt.Sprintf("%d", txData.SafeTxGas),
		"baseGas":        fmt.Sprintf("%d", txData.BaseGas),
		"gasPrice":       txData.GasPrice,
		"gasToken":       txData.GasToken,
		"refundReceiver": txData.RefundReceiver,
		"nonce":          fmt.Sprintf("%d", txData.Nonce),
		"safeTxHash":     safeTxHash.Hex(),
		"sender":         key.Address.Hex(),
		"signature":      senderSignature,
		"origin":         fmt.Sprintf("{\"url\":\"%s\",\"name\":\"SafeProposal Creation\"}", safeApi),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", safeApi, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		var jsonErr interface{}
		if err := json.Unmarshal(body, &jsonErr); err != nil {
			return fmt.Errorf("HTTP %d, failed to parse error body: %s", resp.StatusCode, string(body))
		}
		formatted, _ := json.MarshalIndent(jsonErr, "", "  ")
		return fmt.Errorf("HTTP %d, error response:\n%s", resp.StatusCode, formatted)
	}

	fmt.Println("Safe proposal created successfully")
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
