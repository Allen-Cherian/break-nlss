// Package nlss provides NLSS (Non-Linear Secret Sharing Scheme) cryptographic operations
// for reconstructing private shares from DID and public share images.
package nlss

import (
	"bytes"
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"
)

// RandPos represents random positions for signing and verification
type RandPos struct {
	OriginalPos []int `json:"originalPos"`
	PosForSign  []int `json:"posForSign"`
}

// BreakNLSS reconstructs a private share from DID and public share bytes
func BreakNLSS(didBytes, pubBytes []byte) ([]byte, error) {
	didBits := ConvertToBitString(didBytes)
	pubBits := ConvertToBitString(pubBytes)

	if len(pubBits) < 8*len(didBits) {
		return nil, fmt.Errorf("pubBits too small: got %d, need %d", len(pubBits), 8*len(didBits))
	}

	privateBytes := make([]byte, len(pubBytes))
	temp := ""

	fmt.Printf("[DEBUG] Processing %d bits...\n", len(didBits))

	for i := 0; i < len(didBits); i++ {
		// Progress indicator every 50000 iterations
		if i > 0 && i%50000 == 0 {
			fmt.Printf("[DEBUG] Progress: %d/%d (%.0f%%)\n", i, len(didBits), float64(i)/float64(len(didBits))*100)
		}
		didBit := didBits[i]
		temp = ""

		start := 8 * i
		end := start + 8

		for k := start; k < end; k++ {
			sum := (int(didBit-'0') * int(pubBits[k]-'0')) % 2
			temp = ConvertString(temp, sum)
		}

		pvtCandidate := ConvertBitString(temp)[0]
		pubShareValue := pubBytes[i]

		for {
			x := pubShareValue & pvtCandidate
			cnt := 0
			for y := x; y != 0; y &= (y - 1) {
				cnt++
			}
			computed := cnt % 2
			expected := int(didBit - '0')

			if computed == expected {
				privateBytes[i] = pvtCandidate
				break
			}
			pvtCandidate++
		}
	}

	fmt.Printf("[DEBUG] Processing complete!\n")

	return privateBytes, nil
}

// BreakNLSSFromFiles reconstructs a private share from DID and public share image files
func BreakNLSSFromFiles(didPath, pubSharePath, outputPath string) error {
	didBytes, err := GetPNGImagePixels(didPath)
	if err != nil {
		return fmt.Errorf("failed to load DID image: %w", err)
	}

	pubBytes, err := GetPNGImagePixels(pubSharePath)
	if err != nil {
		return fmt.Errorf("failed to load public share image: %w", err)
	}

	fmt.Printf("Loaded DID bytes: %d\n", len(didBytes))
	fmt.Printf("Loaded public share bytes: %d\n", len(pubBytes))

	pvtBytes, err := BreakNLSS(didBytes, pubBytes)
	if err != nil {
		return fmt.Errorf("BreakNLSS failed: %w", err)
	}

	fmt.Printf("Generated private share bytes: %d\n", len(pvtBytes))

	verified := VerifyPVT(didBytes, pubBytes, pvtBytes)
	fmt.Printf("Verification: %v\n", verified)

	if !verified {
		return fmt.Errorf("private share verification failed")
	}

	// Create PNG output (1024x512 is standard size)
	err = CreatePNGImage(pvtBytes, 1024, 512, outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output PNG: %w", err)
	}

	fmt.Printf("Private share saved to: %s\n", outputPath)
	return nil
}

// VerifyPVT verifies that the private share correctly reconstructs the DID
func VerifyPVT(didBytes, pubBytes, pvtBytes []byte) bool {
	didBits := ConvertToBitString(didBytes)

	for i := 0; i < len(didBits); i++ {
		didBit := didBits[i]
		pubMask := pubBytes[i]
		pvtByte := pvtBytes[i]

		x := pubMask & pvtByte
		cnt := 0
		for y := x; y != 0; y &= (y - 1) {
			cnt++
		}
		computed := cnt % 2
		expected := int(didBit - '0')

		if computed != expected {
			fmt.Printf("Mismatch at bit %d: expected %d got %d\n", i, expected, computed)
			return false
		}
	}

	return true
}

// Combine2Shares combines two shares using XOR-like operation
func Combine2Shares(pvt []byte, pub []byte) []byte {
	pvtString := ConvertToBitString(pvt)
	pubString := ConvertToBitString(pub)
	if len(pvtString) != len(pubString) {
		return nil
	}
	var sum int
	var temp string = ""
	for i := 0; i < len(pvtString); i = i + 8 {
		sum = 0
		for j := i; j < i+8; j++ {
			sum = sum + (int(pvtString[j]-0x30) * int(pubString[j]-0x30))
		}
		sum = sum % 2
		temp = ConvertString(temp, sum)
	}
	return (ConvertBitString(temp))
}

// Sign generates a signature from the private share
func Sign(pvtSharePath string, hash string) ([]byte, error) {
	byteImg, err := GetPNGImagePixels(pvtSharePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private share: %w", err)
	}

	ps := ByteArraytoIntArray(byteImg)
	randPosObject := RandomPositions("signer", hash, 32, ps)
	finalPos := randPosObject.PosForSign
	pvtPos := GetPrivatePositions(finalPos, ps)
	pvtPosStr := IntArraytoStr(pvtPos)
	bs := BitstreamToBytes(pvtPosStr)

	return bs, nil
}

// NlssVerify verifies an NLSS signature
func NlssVerify(didPath, pubSharePath string, hash string, pvtShareSig []byte) (bool, error) {
	didImg, err := GetPNGImagePixels(didPath)
	if err != nil {
		return false, err
	}
	pubImg, err := GetPNGImagePixels(pubSharePath)
	if err != nil {
		return false, err
	}

	pSig := BytesToBitstream(pvtShareSig)
	ps := StringToIntArray(pSig)

	didBin := ByteArraytoIntArray(didImg)
	pubBin := ByteArraytoIntArray(pubImg)
	pubPos := RandomPositions("verifier", hash, 32, ps)
	pubPosInt := GetPrivatePositions(pubPos.PosForSign, pubBin)
	pubStr := IntArraytoStr(pubPosInt)
	orgPos := make([]int, len(pubPos.OriginalPos))
	for i := range pubPos.OriginalPos {
		orgPos[i] = pubPos.OriginalPos[i] / 8
	}
	didPosInt := GetPrivatePositions(orgPos, didBin)
	didStr := IntArraytoStr(didPosInt)
	cb := Combine2Shares(ConvertBitString(pSig), ConvertBitString(pubStr))

	db := ConvertBitString(didStr)
	if !bytes.Equal(cb, db) {
		return false, fmt.Errorf("failed to verify signature")
	}

	return true, nil
}

// GetPNGImagePixels extracts RGB pixel data from a PNG file
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

// CreatePNGImage creates a PNG file from RGB pixel data
func CreatePNGImage(pixels []byte, width int, height int, file string) error {
	if len(pixels) != width*height*3 {
		return fmt.Errorf("invalid pixel buffer: got %d bytes, expected %d", len(pixels), width*height*3)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	offset := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: pixels[offset],
				G: pixels[offset+1],
				B: pixels[offset+2],
				A: 255,
			})
			offset = offset + 3
		}
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return err
	}
	return nil
}

// ConvertToBitString converts bytes to binary string
func ConvertToBitString(data []byte) string {
	var bits string = ""
	for i := 0; i < len(data); i++ {
		bits = bits + fmt.Sprintf("%08b", data[i])
	}
	return bits
}

// ConvertString appends bit to string
func ConvertString(str string, s int) string {
	if s == 1 {
		str = str + "1"
	} else {
		str = str + "0"
	}
	return str
}

// ConvertBitString converts binary string to bytes
func ConvertBitString(bitStr string) []byte {
	if len(bitStr)%8 != 0 {
		return nil
	}
	n := len(bitStr) / 8
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		var b byte = 0
		for j := 0; j < 8; j++ {
			if bitStr[i*8+j] == '1' {
				b |= 1 << uint(7-j)
			}
		}
		out[i] = b
	}
	return out
}

// ByteArraytoIntArray converts byte array to bit array
func ByteArraytoIntArray(byteArray []byte) []int {
	result := make([]int, len(byteArray)*8)
	for i, b := range byteArray {
		for j := 0; j < 8; j++ {
			result[i*8+j] = int(b >> uint(7-j) & 0x01)
		}
	}
	return result
}

// StringToIntArray converts binary string to int array
func StringToIntArray(data string) []int {
	result := make([]int, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == '1' {
			result[i] = 1
		} else {
			result[i] = 0
		}
	}
	return result
}

// GetPrivatePositions extracts values from array at specified positions
func GetPrivatePositions(positions []int, privateArray []int) []int {
	privatePositions := make([]int, len(positions))
	for k := 0; k < len(positions); k++ {
		a := positions[k]
		if a < len(privateArray) {
			privatePositions[k] = privateArray[a]
		} else {
			privatePositions[k] = 0
		}
	}
	return privatePositions
}

// IntArraytoStr converts int array to binary string
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

// BytesToBitstream converts bytes to binary string
func BytesToBitstream(data []byte) string {
	var str string
	for _, d := range data {
		str = str + fmt.Sprintf("%08b", d)
	}
	return str
}

// ImageToBinary converts image to binary string
func ImageToBinary(imagePath string) (string, error) {
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
	binaryBuilder.Grow(width * height * 24)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			binaryBuilder.WriteString(intToBinary(int(r8)))
			binaryBuilder.WriteString(intToBinary(int(g8)))
			binaryBuilder.WriteString(intToBinary(int(b8)))
		}
	}

	return binaryBuilder.String(), nil
}

// intToBinary converts an integer (0-255) to an 8-bit binary string
func intToBinary(pixel int) string {
	binary := strconv.FormatInt(int64(pixel), 2)
	for len(binary) < 8 {
		binary = "0" + binary
	}
	return binary
}

// BitstreamToBytes converts bitstream string to bytes
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

		temp, _ := strconv.ParseInt(str[l:], 2, 64)
		result = append([]byte{byte(temp)}, result...)

		if l == 0 {
			break
		}
		str = str[:l]
	}

	return result
}

// RandomPositions generates deterministic positions from a hash
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
		OriginalPos: originalPos,
		PosForSign:  posForSign,
	}
}

// HexToStr converts bytes to hex string
func HexToStr(d []byte) string {
	dst := make([]byte, hex.EncodedLen(len(d)))
	hex.Encode(dst, d)
	return string(dst)
}

// CalculateHash calculates hash using specified method
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

// CalculateSHA3Hash calculates SHA3-256 hash and returns hex string
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
