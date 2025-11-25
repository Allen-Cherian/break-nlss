package rubix

// InitiateTransferRequest represents the request to initiate a token transfer
// Reference: /Users/allen/Professional/sky/lib/native_interaction/rubix/rubix_platform_calls.dart:121-127
type InitiateTransferRequest struct {
	Receiver   string  `json:"receiver"`
	Sender     string  `json:"sender"`
	TokenCount float64 `json:"tokenCOunt"` // Note: Capital 'O' - this is how Rubix API expects it
	Comment    string  `json:"comment"`
	Type       int     `json:"type"` // Usually 2
}

// InitiateTransferResponse represents the response from initiate transfer API
type InitiateTransferResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Result  struct {
		ID   string `json:"id"`
		Hash string `json:"hash"` // Base64 encoded hash
	} `json:"result"`
}

// SignatureData represents the signature payload
type SignatureData struct {
	Signature []byte `json:"Signature"` // ECDSA signature (ASN.1 DER encoded)
	Pixels    []byte `json:"Pixels"`    // Image-based signature
}

// SignatureRequest represents the request to submit signatures
// Reference: /Users/allen/Professional/sky/lib/native_interaction/rubix/rubix_platform_calls.dart:204-212
type SignatureRequest struct {
	ID        string        `json:"id"`
	Signature SignatureData `json:"signature"`
}

// SignatureResponse represents the response from signature submission
type SignatureResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// GetBalanceRequest represents the request to get account balance
type GetBalanceRequest struct {
	DID string `json:"did"`
}

// AccountInfo represents account information from the balance API
type AccountInfo struct {
	DID        string  `json:"did"`
	DIDType    int     `json:"did_type"`
	RBTAmount  float64 `json:"rbt_amount"`
	PledgedRBT float64 `json:"pledged_rbt"`
	LockedRBT  float64 `json:"locked_rbt"`
	PinnedRBT  float64 `json:"pinned_rbt"`
}

// GetBalanceResponse represents the response from balance API
// Reference: /Users/allen/Professional/sky/lib/native_interaction/rubix/rubix_platform_calls.dart:246-261
type GetBalanceResponse struct {
	Status      bool          `json:"status"`
	Message     string        `json:"message"`
	AccountInfo []AccountInfo `json:"account_info"`
}

// GetAllDIDResponse represents the response from get all DID API
type GetAllDIDResponse struct {
	Status      bool          `json:"status"`
	Message     string        `json:"message"`
	Result      interface{}   `json:"result"`
	AccountInfo []AccountInfo `json:"account_info"`
}
