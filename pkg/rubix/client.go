package rubix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Rubix blockchain HTTP client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Rubix client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitiateTransfer initiates a token transfer and returns the hash to sign
// Reference: /Users/allen/Professional/sky/lib/native_interaction/rubix/rubix_platform_calls.dart:113-152
func (c *Client) InitiateTransfer(req InitiateTransferRequest) (*InitiateTransferResponse, error) {
	url := fmt.Sprintf("http://%s/api/initiate-rbt-transfer", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response InitiateTransferResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("initiate transfer failed: %s", response.Message)
	}

	return &response, nil
}

// SubmitSignature submits the signatures to complete the transfer
// Reference: /Users/allen/Professional/sky/lib/native_interaction/rubix/rubix_platform_calls.dart:202-229
func (c *Client) SubmitSignature(req SignatureRequest) (*SignatureResponse, error) {
	url := fmt.Sprintf("http://%s/api/signature-response", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response SignatureResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("signature submission failed: %s", response.Message)
	}

	return &response, nil
}

// GetBalance retrieves the account balance for a DID
// Reference: /Users/allen/Professional/sky/lib/native_interaction/rubix/rubix_platform_calls.dart:231-261
func (c *Client) GetBalance(did string) (*GetBalanceResponse, error) {
	url := fmt.Sprintf("http://%s/api/get-account-info?did=%s", c.BaseURL, did)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response GetBalanceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("get balance failed: %s", response.Message)
	}

	if len(response.AccountInfo) == 0 {
		return nil, fmt.Errorf("no account info found for DID: %s", did)
	}

	return &response, nil
}

// GetAllDID retrieves all DIDs from the node
func (c *Client) GetAllDID() (*GetAllDIDResponse, error) {
	url := fmt.Sprintf("http://%s/api/getalldid", c.BaseURL)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response GetAllDIDResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Status {
		return nil, fmt.Errorf("get all DID failed: %s", response.Message)
	}

	return &response, nil
}
