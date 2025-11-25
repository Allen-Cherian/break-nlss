package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

func RandomPositions(role string, hash string, numOfPositions int, pvt1 []int) *RandPos {
	var u, l, m int = 0, 0, 0

	hashCharacters := make([]int, 32)
	randomPositions := make([]int, 32)
	randPos := make([]int, 256)
	var finalPositions, pos []int
	originalPos := make([]int, 32)
	posForSign := make([]int, 32*8)

	for k := 0; k < numOfPositions; k++ {

		temp, err := strconv.ParseInt(string(hash[k]), 16, 32)
		if err != nil {
			return nil
		}
		hashCharacters[k] = int(temp)
		randomPositions[k] = (((2402 + hashCharacters[k]) * 2709) + ((k + 2709) + hashCharacters[(k)])) % 2048
		originalPos[k] = (randomPositions[k] / 8) * 8

		pos = make([]int, 32)
		pos[k] = originalPos[k]
		randPos[k] = pos[k]

		finalPositions = make([]int, 8)

		for p := 0; p < 8; p++ {

			posForSign[u] = randPos[k]
			randPos[k]++
			u++

			finalPositions[l] = pos[k]
			pos[k]++
			l++

			if l == 8 {
				l = 0
			}
		}
		if role == "signer" {
			var p1 []int = GetPrivatePositions(finalPositions, pvt1)
			hash = HexToStr(CalculateHash([]byte(hash+IntArraytoStr(originalPos)+IntArraytoStr(p1)), "SHA3-256"))

		} else {
			p1 := make([]int, 8)
			for i := 0; i < 8; i++ {
				p1[i] = pvt1[m]
				m++
			}
			hash = HexToStr(CalculateHash([]byte(hash+IntArraytoStr(originalPos)+IntArraytoStr(p1)), "SHA3-256"))
		}
	}
	return &RandPos{
		OriginalPos: originalPos, PosForSign: posForSign}
}

func HexToStr(d []byte) string {
	dst := make([]byte, hex.EncodedLen(len(d)))
	hex.Encode(dst, d)

	return string(dst)
}

func CalculateHash(data []byte, method string) []byte {
	switch method {
	case "SHA3-256":
		h := sha3.New256()
		h.Write(data)
		return h.Sum(nil)
	default:
		return nil
	}
}

func GetPNGImagePixels(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	pixels := make([]byte, 0, w*h*3)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixels = append(pixels, byte(r>>8))
			pixels = append(pixels, byte(g>>8))
			pixels = append(pixels, byte(b>>8))
		}
	}
	return pixels, nil
}

func GetPrivatePositions(positions []int, privateArray []int) []int {

	privatePositions := make([]int, len(positions))

	for k := 0; k < len(positions); k++ {
		var a int = positions[k]
		var b int = privateArray[a]

		privatePositions[k] = b
	}

	return privatePositions
}

func IntArraytoStr(intArray []int) string {
	var result bytes.Buffer
	for i := 0; i < len(intArray); i++ {
		if intArray[i] == 1 {
			result.WriteString("1")
		} else {
			result.WriteString("0")
		}
	}
	return result.String()
}

type RandPos struct {
	OriginalPos []int `json:"originalPos"`
	PosForSign  []int `json:"posForSign"`
}

func Sign(pvtSharePath string, hash string) ([]byte, error) {
	byteImg, err := GetPNGImagePixels(pvtSharePath)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	ps := ByteArraytoIntArray(byteImg)

	randPosObject := RandomPositions("signer", hash, 32, ps)

	finalPos := randPosObject.PosForSign
	pvtPos := GetPrivatePositions(finalPos, ps)
	pvtPosStr := IntArraytoStr(pvtPos)
	bs := BitstreamToBytes(pvtPosStr)
	if err != nil {
		return nil, err
	}
	return bs, err
}

func ByteArraytoIntArray(byteArray []byte) []int {

	result := make([]int, len(byteArray)*8)
	for i, b := range byteArray {
		for j := 0; j < 8; j++ {
			result[i*8+j] = int(b >> uint(7-j) & 0x01)
		}
	}
	return result
}

// ImageToBinary converts a PNG image to a binary string
// Each pixel's RGB values are converted to 8-bit binary strings and concatenated
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:11-54
func ImageToBinary(imagePath string) (string, error) {
	// Open and decode the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	var binaryBuilder strings.Builder
	// Pre-allocate approximate size (3 colors * 8 bits * pixels)
	binaryBuilder.Grow(width * height * 24)

	// Process each pixel row by row (top to bottom, left to right)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get pixel color
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert from 16-bit to 8-bit (RGBA returns 16-bit values)
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Convert each color component to 8-bit binary string
			binaryBuilder.WriteString(intToBinary(int(r8)))
			binaryBuilder.WriteString(intToBinary(int(g8)))
			binaryBuilder.WriteString(intToBinary(int(b8)))
			// Note: Alpha channel is ignored, as per Dart implementation comment
		}
	}

	return binaryBuilder.String(), nil
}

// intToBinary converts an integer (0-255) to an 8-bit binary string
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:116-122
func intToBinary(pixel int) string {
	binary := strconv.FormatInt(int64(pixel), 2)
	// Pad with leading zeros to ensure 8 bits
	for len(binary) < 8 {
		binary = "0" + binary
	}
	return binary
}

// RandomPositions generates deterministic positions from a hash
// This is the CRITICAL algorithm that must match exactly with the Dart implementation
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:56-114
// func RandomPositions(role string, hash string, numberOfPositions int, pvt1 []int) map[string][]int {
// 	hashCharacters := make([]int, 32)
// 	randomPositions := make([]int, 32)
// 	originalPos := make([]int, 32)
// 	posForSign := make([]int, 32*8) // 256 positions

// 	u := 0
// 	m := 0

// 	for k := 0; k < numberOfPositions; k++ {
// 		// 1. Parse hex character from hash
// 		hashVar := string(hash[k])
// 		hashChar, _ := strconv.ParseInt(hashVar, 16, 64)
// 		hashCharacters[k] = int(hashChar)

// 		// 2. CRITICAL FORMULA - Must match exactly
// 		// Formula: (((2402 + hashChar) * 2709) + ((k + 2709) + hashChar)) % 2048
// 		randomPositions[k] = (((2402 + hashCharacters[k]) * 2709) +
// 			((k + 2709) + hashCharacters[k])) % 2048

// 		// 3. Calculate byte-aligned position
// 		originalPos[k] = (randomPositions[k] / 8) * 8

// 		// 4. Extract 8 consecutive bit positions starting from randomPositions[k]
// 		randPos := randomPositions[k]
// 		for p := 0; p < 8; p++ {
// 			posForSign[u] = randPos
// 			randPos++
// 			u++
// 		}

// 		// 5. Update hash for next iteration (only for "signer" role)
// 		if role == "signer" {
// 			// Get 8 positions for this iteration
// 			finalPositions := make([]int, 8)
// 			for i := 0; i < 8; i++ {
// 				finalPositions[i] = originalPos[k] + i
// 			}

// 			// Get private values at these positions
// 			p1 := getPrivatePosition(finalPositions, pvt1)

// 			// Calculate new hash
// 			hash = CalculateSHA3Hash(hash + intArrayToStr(originalPos) + intArrayToStr(p1))
// 		} else {
// 			// For non-signer role
// 			p1 := make([]int, 8)
// 			for i := 0; i < 8; i++ {
// 				if m < len(pvt1) {
// 					p1[i] = pvt1[m]
// 					m++
// 				}
// 			}
// 			hash = CalculateSHA3Hash(hash + intArrayToStr(originalPos) + intArrayToStrJoin(p1))
// 		}
// 	}

// 	return map[string][]int{
// 		"originalPos": originalPos, // 32 byte-aligned positions
// 		"posForSign":  posForSign,  // 256 bit positions
// 	}
// }

// getPrivatePosition extracts values from privateArray at specified positions
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:172-185
func getPrivatePosition(positions []int, privateArray []int) []int {
	privatePosition := make([]int, len(positions))

	for k := 0; k < len(positions); k++ {
		a := positions[k]
		if a < len(privateArray) {
			privatePosition[k] = privateArray[a]
		} else {
			privatePosition[k] = 0
		}
	}

	return privatePosition
}

// intArrayToStr converts an int array of 0s and 1s to a string
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:124-134
func intArrayToStr(inputArray []int) string {
	var builder strings.Builder
	for _, val := range inputArray {
		if val == 1 {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
	}
	return builder.String()
}

// intArrayToStrJoin is an alternative conversion using join (for compatibility)
func intArrayToStrJoin(inputArray []int) string {
	strArray := make([]string, len(inputArray))
	for i, val := range inputArray {
		strArray[i] = strconv.Itoa(val)
	}
	return strings.Join(strArray, "")
}

// BitstreamToBytes converts a bitstream (string of 0s and 1s) to bytes
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/dependencies.dart:141-160
func BitstreamToBytes(bitstream string) []byte {
	var result []byte
	str := bitstream

	for str != "" {
		var l int
		if len(str) > 8 {
			l = len(str) - 8
		} else {
			l = 0
		}

		// Parse the last 8 bits (or remaining bits)
		temp, _ := strconv.ParseInt(str[l:], 2, 64)
		result = append([]byte{byte(temp)}, result...) // Prepend to result

		if l == 0 {
			break
		}
		str = str[:l]
	}

	return result
}

// BitstreamToBytesFromIntArray converts an int array of 0s and 1s to bytes
func BitstreamToBytesFromIntArray(bitstream []int) []byte {
	var bytes []byte

	// Process 8 bits at a time
	for i := 0; i < len(bitstream); i += 8 {
		var b byte = 0
		for j := 0; j < 8 && i+j < len(bitstream); j++ {
			if bitstream[i+j] == 1 {
				// MSB first (most significant bit first)
				b |= 1 << uint(7-j)
			}
		}
		bytes = append(bytes, b)
	}

	return bytes
}

// GenerateSignatureFromShares generates a complete image-based signature
// This is the main function that combines all the pieces
// Reference: /Users/allen/Professional/sky/fexr-flutter/lib/signature/gen_sign.dart:11-20
// func GenerateSignatureFromShares(imagePath string, hash string) ([]byte, error) {
// 	// 1. Convert image to binary string
// 	binaryStr, err := ImageToBinary(imagePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to convert image to binary: %w", err)
// 	}

// 	// 2. Convert binary string to int array (each '0'/'1' char → 0/1 int)
// 	privateIntArray := make([]int, len(binaryStr))
// 	for i, char := range binaryStr {
// 		if char == '1' {
// 			privateIntArray[i] = 1
// 		} else {
// 			privateIntArray[i] = 0
// 		}
// 	}

// 	// 3. Generate deterministic positions from hash
// 	positions := RandomPositions("signer", hash, 32, privateIntArray)
// 	finalPos := positions["posForSign"] // 256 positions

// 	// 4. Extract signature bits at calculated positions
// 	signatureBits := getPrivatePosition(finalPos, privateIntArray)

// 	// 5. Convert bitstream (int array of 0s and 1s) to bytes
// 	signatureBytes := BitstreamToBytesFromIntArray(signatureBits)

// 	return signatureBytes, nil
// }
