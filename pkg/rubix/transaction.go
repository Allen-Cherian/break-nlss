package rubix

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"path/filepath"

	"break-nlss/pkg/crypto"
)

// TransferParams contains all parameters needed for a token transfer
type TransferParams struct {
	RubixNodeURL  string
	SenderDID     string
	ReceiverDID   string
	Amount        float64
	Comment       string
	NLSSOutputDir string // Output directory where pvtShare.png files are stored
}

// TransferTokens performs a complete two-phase token transfer
// Phase 1: Initiate transfer and get hash
// Phase 2: Sign hash and submit signatures
func TransferTokens(params TransferParams) error {
	client := NewClient(params.RubixNodeURL)

	// ============================================
	// PHASE 1: Initiate Transfer
	// ============================================
	fmt.Println("Phase 1: Initiating transfer...")

	initiateReq := InitiateTransferRequest{
		Receiver:   params.ReceiverDID,
		Sender:     params.SenderDID, // Use sender DID from params
		TokenCount: params.Amount,
		Comment:    params.Comment,
		Type:       2, // Type 2 for RBT transfer
	}

	initiateResp, err := client.InitiateTransfer(initiateReq)
	if err != nil {
		return fmt.Errorf("failed to initiate transfer: %w", err)
	}

	requestID := initiateResp.Result.ID
	hashBase64 := initiateResp.Result.Hash

	fmt.Printf("✓ Transaction initiated\n")
	fmt.Printf("  Request ID: %s\n", requestID)
	fmt.Printf("  Hash (Base64): %s\n", hashBase64)

	// ============================================
	// PHASE 2: Sign and Complete
	// ============================================
	fmt.Println("\nPhase 2: Generating signatures...")

	// 2.1: Decode hash from Base64
	hashBytes, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}
	hash := string(hashBytes)
	fmt.Printf("✓ Decoded hash: %s\n", hash)

	// 2.2: Generate image-based signature
	// Construct path to private share: ./output/{sender_did}/pvtShare.png
	pvtSharePath := filepath.Join(params.NLSSOutputDir, params.SenderDID, "pvtShare.png")
	fmt.Printf("  Generating image signature from: %s\n", pvtSharePath)

	imgSignBytes, err := crypto.Sign(pvtSharePath, hash)
	if err != nil {
		return fmt.Errorf("failed to generate image signature: %w", err)
	}
	fmt.Printf("✓ Image signature generated (%d bytes)\n", len(imgSignBytes))
	fmt.Println("The image sign here :", imgSignBytes)

	// 2.3: Submit signatures
	fmt.Println("\nPhase 3: Submitting signatures...")

	// Use empty byte array for ECDSA signature (not required)
	var pvtBytes []byte

	signReq := SignatureRequest{
		ID: requestID,
		Signature: SignatureData{
			Signature: pvtBytes,
			Pixels:    imgSignBytes,
		},
	}
	fmt.Println("The signReq :", signReq)
	signResp, err := client.SubmitSignature(signReq)
	if err != nil {
		return fmt.Errorf("failed to submit signature: %w", err)
	}

	fmt.Printf("\n✓ Transaction completed successfully!\n")
	fmt.Printf("  Message: %s\n", signResp.Message)

	return nil
}

// GetAccountBalance retrieves the balance for a DID
func GetAccountBalance(rubixNodeURL, did string) (float64, error) {
	client := NewClient(rubixNodeURL)

	response, err := client.GetBalance(did)
	if err != nil {
		return 0, err
	}

	if len(response.AccountInfo) == 0 {
		return 0, fmt.Errorf("no account info found")
	}

	return response.AccountInfo[0].RBTAmount, nil
}

// GenerateAndSaveKeys generates a new EC key pair and saves to files
func GenerateAndSaveKeys(privateKeyPath, publicKeyPath string) (*ecdsa.PrivateKey, error) {
	// Generate key pair
	privateKey, err := crypto.GenerateECKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Save private key
	if err := crypto.SavePrivateKeyToPEM(privateKey, privateKeyPath); err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	if err := crypto.SavePublicKeyToPEM(&privateKey.PublicKey, publicKeyPath); err != nil {
		return nil, fmt.Errorf("failed to save public key: %w", err)
	}

	return privateKey, nil
}
