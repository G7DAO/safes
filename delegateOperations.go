package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"os"

	"io"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"golang.org/x/crypto/ssh/terminal"
)

type DelegateResponse struct {
	Safe      string `json:"safe"`
	Delegate  string `json:"delegate"`
	Delegator string `json:"delegator"`
	Label     string `json:"label"`
}

func AddDelegate(safeAddress, delegateAddress, label string, chainID *big.Int, key *keystore.Key, apiURL string) error {
	// Generate TOTP (Time-based One-Time Password)
	totp := big.NewInt(time.Now().Unix() / 3600)

	// Convert addresses to checksum format
	checksumSafe := common.HexToAddress(safeAddress).Hex()
	checksumDelegate := common.HexToAddress(delegateAddress).Hex()
	checksumSigner := key.Address.Hex()

	// Create EIP-712 message
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"Delegate": []apitypes.Type{
				{Name: "delegateAddress", Type: "address"},
				{Name: "totp", Type: "uint256"},
			},
		},
		PrimaryType: "Delegate",
		Domain: apitypes.TypedDataDomain{
			Name:    "Safe Transaction Service",
			Version: "1.0",
			ChainId: (*math.HexOrDecimal256)(chainID),
		},
		Message: apitypes.TypedDataMessage{
			"delegateAddress": checksumDelegate,
			"totp":            totp.String(),
		},
	}

	typedDataHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return fmt.Errorf("failed to hash typed data: %v", err)
	}

	// Sign the typedDataHash
	signature, err := crypto.Sign(common.BytesToHash(typedDataHash).Bytes(), key.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign typed data hash: %v", err)
	}

	// Adjust V value for Ethereum's replay protection
	signature[64] += 27

	// Convert signature to hex
	senderSignature := "0x" + common.Bytes2Hex(signature)

	// Create the request payload
	payload := map[string]string{
		"safe":      checksumSafe,
		"delegate":  checksumDelegate,
		"delegator": checksumSigner,
		"signature": senderSignature,
		"label":     label,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Delegate added successfully.")

	return nil
}

func GetDelegates(safe, delegate, delegator, label string, limit, offset int, chainID *big.Int, apiURL string) ([]DelegateResponse, error) {
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	params := url.Values{}
	params.Add("safe", safe)
	if delegate != "" {
		params.Add("delegate", delegate)
	}
	if delegator != "" {
		params.Add("delegator", delegator)
	}
	if label != "" {
		params.Add("label", label)
	}
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", offset))
	}

	baseURL.RawQuery = params.Encode()

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
		Results  []DelegateResponse `json:"results"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Results, nil
}

func RemoveDelegate(safeAddress, delegateAddress string, chainID *big.Int, key *keystore.Key, apiURL string) error {
	// Generate TOTP (Time-based One-Time Password)
	totp := big.NewInt(time.Now().Unix() / 3600)

	// Convert addresses to checksum format
	checksumSafe := common.HexToAddress(safeAddress).Hex()
	checksumDelegate := common.HexToAddress(delegateAddress).Hex()
	checksumSigner := key.Address.Hex()

	// Create EIP-712 message
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"Delegate": []apitypes.Type{
				{Name: "delegateAddress", Type: "address"},
				{Name: "totp", Type: "uint256"},
			},
		},
		PrimaryType: "Delegate",
		Domain: apitypes.TypedDataDomain{
			Name:    "Safe Transaction Service",
			Version: "1.0",
			ChainId: (*math.HexOrDecimal256)(chainID),
		},
		Message: apitypes.TypedDataMessage{
			"delegateAddress": checksumDelegate,
			"totp":            totp.String(),
		},
	}

	typedDataHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return fmt.Errorf("failed to hash typed data: %v", err)
	}

	// Sign the SafeTxHash
	signature, err := crypto.Sign(common.BytesToHash(typedDataHash).Bytes(), key.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign SafeTxHash: %v", err)
	}

	// Adjust V value for Ethereum's replay protection
	signature[64] += 27

	// Convert signature to hex
	senderSignature := "0x" + common.Bytes2Hex(signature)

	// Create the request payload
	payload := map[string]string{
		"safe":      checksumSafe,
		"delegator": checksumSigner,
		"signature": senderSignature,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("DELETE", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Change this part
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Delegate removed successfully.")

	return nil
}

func KeyFromFile(keystoreFile string, password string) (*keystore.Key, error) {
	var emptyKey *keystore.Key
	keystoreContent, readErr := os.ReadFile(keystoreFile)
	if readErr != nil {
		return emptyKey, readErr
	}

	// If password is "", prompt user for password.
	if password == "" {
		fmt.Printf("Please provide a password for keystore (%s): ", keystoreFile)
		passwordRaw, inputErr := terminal.ReadPassword(int(os.Stdin.Fd()))
		if inputErr != nil {
			return emptyKey, fmt.Errorf("error reading password: %s", inputErr.Error())
		}
		fmt.Print("\n")
		password = string(passwordRaw)
	}

	key, err := keystore.DecryptKey(keystoreContent, password)
	return key, err
}
