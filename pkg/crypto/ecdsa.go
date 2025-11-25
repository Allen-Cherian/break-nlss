package crypto

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// GenerateECKeyPair generates a new EC key pair using P-256 curve
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/key_gen.dart:9-30
func GenerateECKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// UnSeal decrypts data using AES-GCM with password-derived key
// Reference: /Users/allen/Professional/rubixgoplatform/crypto/seal.go:32-51
func UnSeal(password string, data []byte) ([]byte, error) {
	// Hash password to get AES key
	h := sha256.New()
	h.Write([]byte(password))
	key := h.Sum(nil)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("invalid data")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// LoadPrivateKeyFromPEM loads an EC private key from a PEM file
// Supports encrypted PKCS8 format (matching rubixgoplatform implementation)
// Reference: /Users/allen/Professional/rubixgoplatform/crypto/crypto.go:89-129
func LoadPrivateKeyFromPEM(filepath string) (*ecdsa.PrivateKey, error) {
	return LoadPrivateKeyFromPEMWithPassword(filepath, "")
}

// LoadPrivateKeyFromPEMWithPassword loads an EC private key with optional password
// Reference: /Users/allen/Professional/rubixgoplatform/crypto/crypto.go:89-129
func LoadPrivateKeyFromPEMWithPassword(filepath, password string) (*ecdsa.PrivateKey, error) {
	pemData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PEM file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	var keyBytes []byte

	// Check if encrypted (matching rubixgoplatform logic)
	if block.Type == "ENCRYPTED PRIVATE KEY" {
		if password == "" {
			// Try to get password from environment
			password = os.Getenv("PRIVATE_KEY_PASSWORD")
			if password == "" {
				return nil, errors.New("key is encrypted but no password provided (set PRIVATE_KEY_PASSWORD env var)")
			}
		}

		// Decrypt using AES-GCM (same as rubixgoplatform)
		keyBytes, err = UnSeal(password, block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key (check password): %w", err)
		}
	} else {
		keyBytes = block.Bytes
	}

	// Parse as PKCS8 (matching rubixgoplatform)
	cryptoPrivKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	// Convert to ECDSA private key
	ecKey, ok := cryptoPrivKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an EC private key (found %T)", cryptoPrivKey)
	}

	return ecKey, nil
}

// SavePrivateKeyToPEM saves an EC private key to a PEM file
func SavePrivateKeyToPEM(privateKey *ecdsa.PrivateKey, filepath string) error {
	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509Encoded,
	})

	return os.WriteFile(filepath, pemEncoded, 0600)
}

// SavePublicKeyToPEM saves an EC public key to a PEM file
func SavePublicKeyToPEM(publicKey *ecdsa.PublicKey, filepath string) error {
	x509Encoded, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509Encoded,
	})

	return os.WriteFile(filepath, pemEncoded, 0644)
}

// SignWithECDSA signs data with an ECDSA private key
// Uses crypto.Signer interface with SHA256 (matching rubixgoplatform)
// Reference: /Users/allen/Professional/rubixgoplatform/crypto/crypto.go:131-133
func SignWithECDSA(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	// Use crypto.Signer interface (same as rubixgoplatform)
	// This automatically handles SHA256 hashing and ASN.1 DER encoding
	signer := crypto.Signer(privateKey)
	signature, err := signer.Sign(rand.Reader, data, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	return signature, nil
}

// VerifyECDSASignature verifies an ECDSA signature
// Uses ecdsa.VerifyASN1 (matching rubixgoplatform)
// Reference: /Users/allen/Professional/rubixgoplatform/crypto/crypto.go:135-139
func VerifyECDSASignature(publicKey *ecdsa.PublicKey, data []byte, signatureBytes []byte) (bool, error) {
	// Use VerifyASN1 (same as rubixgoplatform)
	// This function expects the signature in ASN.1 DER format
	valid := ecdsa.VerifyASN1(publicKey, data, signatureBytes)
	return valid, nil
}
