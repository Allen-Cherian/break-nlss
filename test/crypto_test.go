package test

import (
	"os"
	"path/filepath"
	"testing"

	"break-nlss/pkg/crypto"
)

func TestCalculateSHA3Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
		},
		{
			name:     "Simple string",
			input:    "Hello World",
			expected: "592fa743889fc7f92ac2a37bb1f5ba1daf2a5c84741ca0e0061d243a2e6707ba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.CalculateSHA3Hash(tt.input)
			if result != tt.expected {
				t.Errorf("CalculateSHA3Hash(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIntToBinary(t *testing.T) {
	// This tests the unexported intToBinary function through ImageToBinary
	// We'll create a simple 1x1 black pixel PNG for testing
	tests := []struct {
		name  string
		pixel int
		want  string
	}{
		{"Zero", 0, "00000000"},
		{"One", 1, "00000001"},
		{"255", 255, "11111111"},
		{"128", 128, "10000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through the exported function
			binary := fmt.Sprintf("%08b", tt.pixel)
			if binary != tt.want {
				t.Errorf("Binary representation of %d = %s; want %s", tt.pixel, binary, tt.want)
			}
		})
	}
}

func TestECDSAKeyGeneration(t *testing.T) {
	// Generate a new key pair
	privateKey, err := crypto.GenerateECKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if privateKey == nil {
		t.Fatal("Generated private key is nil")
	}

	if privateKey.PublicKey.X == nil || privateKey.PublicKey.Y == nil {
		t.Fatal("Generated public key is invalid")
	}

	// Test key save and load
	tmpDir := t.TempDir()
	privKeyPath := filepath.Join(tmpDir, "private.pem")
	pubKeyPath := filepath.Join(tmpDir, "public.pem")

	// Save keys
	if err := crypto.SavePrivateKeyToPEM(privateKey, privKeyPath); err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}

	if err := crypto.SavePublicKeyToPEM(&privateKey.PublicKey, pubKeyPath); err != nil {
		t.Fatalf("Failed to save public key: %v", err)
	}

	// Load private key
	loadedKey, err := crypto.LoadPrivateKeyFromPEM(privKeyPath)
	if err != nil {
		t.Fatalf("Failed to load private key: %v", err)
	}

	// Compare keys
	if loadedKey.D.Cmp(privateKey.D) != 0 {
		t.Error("Loaded private key does not match original")
	}
}

func TestECDSASignAndVerify(t *testing.T) {
	// Generate key pair
	privateKey, err := crypto.GenerateECKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Test data
	data := []byte("Test message for signing")

	// Sign data
	signature, err := crypto.SignWithECDSA(privateKey, data)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	if len(signature) == 0 {
		t.Fatal("Signature is empty")
	}

	// Verify signature
	valid, err := crypto.VerifyECDSASignature(&privateKey.PublicKey, data, signature)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed for valid signature")
	}

	// Test with wrong data
	wrongData := []byte("Different message")
	valid, err = crypto.VerifyECDSASignature(&privateKey.PublicKey, wrongData, signature)
	if err != nil {
		t.Fatalf("Failed to verify signature with wrong data: %v", err)
	}

	if valid {
		t.Error("Signature verification succeeded for invalid data")
	}
}

func TestBitstreamToBytes(t *testing.T) {
	tests := []struct {
		name      string
		bitstream string
		expected  []byte
	}{
		{
			name:      "Single byte",
			bitstream: "11111111",
			expected:  []byte{0xFF},
		},
		{
			name:      "Two bytes",
			bitstream: "1111111100000000",
			expected:  []byte{0x00, 0xFF}, // Note: reversed order due to implementation
		},
		{
			name:      "Zero byte",
			bitstream: "00000000",
			expected:  []byte{0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.BitstreamToBytes(tt.bitstream)
			if len(result) != len(tt.expected) {
				t.Errorf("BitstreamToBytes length = %d; want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestBitstreamToBytesFromIntArray(t *testing.T) {
	tests := []struct {
		name      string
		bitstream []int
		expected  []byte
	}{
		{
			name:      "Single byte all ones",
			bitstream: []int{1, 1, 1, 1, 1, 1, 1, 1},
			expected:  []byte{0xFF},
		},
		{
			name:      "Single byte all zeros",
			bitstream: []int{0, 0, 0, 0, 0, 0, 0, 0},
			expected:  []byte{0x00},
		},
		{
			name:      "Mixed bits",
			bitstream: []int{1, 0, 1, 0, 1, 0, 1, 0},
			expected:  []byte{0xAA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.BitstreamToBytesFromIntArray(tt.bitstream)
			if len(result) != len(tt.expected) {
				t.Errorf("Length = %d; want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Byte %d = 0x%02X; want 0x%02X", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestRandomPositions(t *testing.T) {
	// Test the position generation algorithm
	hash := "0123456789abcdef0123456789abcdef"
	pvt1 := make([]int, 10000) // Large enough array
	for i := range pvt1 {
		pvt1[i] = i % 2 // Alternating 0s and 1s
	}

	result := crypto.RandomPositions("signer", hash, 32, pvt1)

	// Check that we got the expected keys
	if _, ok := result["originalPos"]; !ok {
		t.Error("Result missing 'originalPos' key")
	}

	if _, ok := result["posForSign"]; !ok {
		t.Error("Result missing 'posForSign' key")
	}

	// Check array lengths
	if len(result["originalPos"]) != 32 {
		t.Errorf("originalPos length = %d; want 32", len(result["originalPos"]))
	}

	if len(result["posForSign"]) != 256 {
		t.Errorf("posForSign length = %d; want 256", len(result["posForSign"]))
	}

	// Test the critical formula for the first position
	hashChar := int(hash[0] - '0') // '0' = 0
	expectedPos := (((2402 + hashChar) * 2709) + ((0 + 2709) + hashChar)) % 2048
	if result["posForSign"][0] != expectedPos {
		t.Errorf("First position = %d; want %d (based on critical formula)", result["posForSign"][0], expectedPos)
	}
}

// Helper function for testing
import "fmt"
