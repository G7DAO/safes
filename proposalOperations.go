package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func SignAndApproveProposal(hash string, key *keystore.Key, apiURL string) error {
	hashBytes := common.HexToHash(hash).Bytes()
	signature, err := crypto.Sign(hashBytes, key.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign SafeTxHash: %w", err)
	}

	signature[64] += 27
	senderSignature := "0x" + common.Bytes2Hex(signature)

	payload := map[string]string{
		"signature": senderSignature,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error! status: %d, body: %s", resp.StatusCode, string(body))
	}

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
