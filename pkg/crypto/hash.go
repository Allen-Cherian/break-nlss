package crypto

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// CalculateSHA3Hash calculates SHA3-256 hash of the input string
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:162-170
func CalculateSHA3Hash(input string) string {
	hasher := sha3.New256()
	hasher.Write([]byte(input))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// CalculateSHA3HashBytes calculates SHA3-256 hash and returns bytes
func CalculateSHA3HashBytes(input []byte) []byte {
	hasher := sha3.New256()
	hasher.Write(input)
	return hasher.Sum(nil)
}
