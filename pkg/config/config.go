package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	RubixNodeURL     string // e.g., "localhost:20006"
	PresetFolder     string // Path to preset folder
	PrivateKeyPath   string // Path to privatekey.pem
	PrivateSharePath string // Path to PrivateShare.png
	PublicSharePath  string // Path to PublicShare.png
	DIDImagePath     string // Path to DID.png
	SenderPeerID     string // e.g., "peer789"
	SenderDID        string // e.g., "DID012"

	// NLSS Configuration
	NLSSBasePath     string // e.g., "/mnt/storage/bulkset/set1"
	NLSSNodeName     string // e.g., "bulk011"
	NLSSDIDImageName string // e.g., "did.png" (default)
	NLSSPubShareName string // e.g., "pubShare.png" (default)
	NLSSOutputDir    string // e.g., "./output" (default)
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	// Get preset folder from env or use default
	presetFolder := os.Getenv("PRESET_FOLDER")
	if presetFolder == "" {
		// Try to use ./preset in current directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		presetFolder = filepath.Join(cwd, "preset")
	}

	// Get Rubix node URL from env or use default
	rubixNodeURL := os.Getenv("RUBIX_NODE_URL")
	if rubixNodeURL == "" {
		rubixNodeURL = "localhost:20006"
	}

	// Get NLSS configuration from env with defaults
	nlssBasePath := os.Getenv("NLSS_BASE_PATH")
	nlssNodeName := os.Getenv("NLSS_NODE_NAME")
	nlssDIDImageName := os.Getenv("NLSS_DID_IMAGE_NAME")
	if nlssDIDImageName == "" {
		nlssDIDImageName = "did.png"
	}
	nlssPubShareName := os.Getenv("NLSS_PUB_SHARE_NAME")
	if nlssPubShareName == "" {
		nlssPubShareName = "pubShare.png"
	}
	nlssOutputDir := os.Getenv("NLSS_OUTPUT_DIR")
	if nlssOutputDir == "" {
		cwd, _ := os.Getwd()
		nlssOutputDir = filepath.Join(cwd, "output")
	}

	config := &Config{
		RubixNodeURL:     rubixNodeURL,
		PresetFolder:     presetFolder,
		PrivateKeyPath:   filepath.Join(presetFolder, "pvtKey.pem"),
		PrivateSharePath: filepath.Join(presetFolder, "pvtShare.png"),
		PublicSharePath:  filepath.Join(presetFolder, "PublicShare.png"),
		DIDImagePath:     filepath.Join(presetFolder, "did.png"),
		SenderPeerID:     os.Getenv("SENDER_PEER_ID"),
		SenderDID:        os.Getenv("SENDER_DID"),
		NLSSBasePath:     nlssBasePath,
		NLSSNodeName:     nlssNodeName,
		NLSSDIDImageName: nlssDIDImageName,
		NLSSPubShareName: nlssPubShareName,
		NLSSOutputDir:    nlssOutputDir,
	}

	return config, nil
}

// LoadConfigWithOverrides loads configuration with command-line overrides
func LoadConfigWithOverrides(rubixNode, presetFolder, senderPeerID, senderDID string) (*Config, error) {
	// Start with environment-based config
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Override with command-line arguments if provided
	if rubixNode != "" {
		config.RubixNodeURL = rubixNode
	}

	if presetFolder != "" {
		config.PresetFolder = presetFolder
		config.PrivateKeyPath = filepath.Join(presetFolder, "privatekey.pem")
		config.PrivateSharePath = filepath.Join(presetFolder, "PrivateShare.png")
		config.PublicSharePath = filepath.Join(presetFolder, "PublicShare.png")
		config.DIDImagePath = filepath.Join(presetFolder, "DID.png")
	}

	if senderPeerID != "" {
		config.SenderPeerID = senderPeerID
	}

	if senderDID != "" {
		config.SenderDID = senderDID
	}

	return config, nil
}

// Validate checks if all required configuration is present and files exist
func (c *Config) Validate() error {
	// if c.RubixNodeURL == "" {
	// 	return fmt.Errorf("RUBIX_NODE_URL is required")
	// }

	// if c.SenderPeerID == "" {
	// 	return fmt.Errorf("SENDER_PEER_ID is required")
	// }

	// if c.SenderDID == "" {
	// 	return fmt.Errorf("SENDER_DID is required")
	// }

	// Check if preset folder exists
	if _, err := os.Stat(c.PresetFolder); os.IsNotExist(err) {
		return fmt.Errorf("preset folder does not exist: %s", c.PresetFolder)
	}

	// Check if required files exist
	requiredFiles := map[string]string{
		"Private Key":   c.PrivateKeyPath,
		"Private Share": c.PrivateSharePath,
	}

	for name, path := range requiredFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("%s file does not exist: %s", name, path)
		}
	}

	return nil
}

// PrintConfig prints the current configuration
func (c *Config) PrintConfig() {
	fmt.Println("Configuration:")
	fmt.Printf("  Rubix Node URL: %s\n", c.RubixNodeURL)
	fmt.Printf("  Preset Folder: %s\n", c.PresetFolder)
	fmt.Printf("  Sender Peer ID: %s\n", c.SenderPeerID)
	fmt.Printf("  Sender DID: %s\n", c.SenderDID)
	fmt.Printf("  Private Key: %s\n", c.PrivateKeyPath)
	fmt.Printf("  Private Share: %s\n", c.PrivateSharePath)
}

// GetNLSSImagePaths constructs the full paths for DID and public share images
// based on the configured base path, node name, and DID
// Path format: {basePath}/{nodeName}/Rubix/{did}/{imageName}
func (c *Config) GetNLSSImagePaths(did string) (didPath, pubSharePath string, err error) {
	if c.NLSSBasePath == "" {
		return "", "", fmt.Errorf("NLSS_BASE_PATH not configured in .env file")
	}
	if c.NLSSNodeName == "" {
		return "", "", fmt.Errorf("NLSS_NODE_NAME not configured in .env file")
	}

	// Construct base directory: /mnt/storage/bulkset/set1/bulk011/Rubix/{did}/
	baseDir := filepath.Join(
		c.NLSSBasePath,
		c.NLSSNodeName,
		"Rubix",
		did,
	)

	// Construct full paths to images
	didPath = filepath.Join(baseDir, c.NLSSDIDImageName)
	pubSharePath = filepath.Join(baseDir, c.NLSSPubShareName)

	return didPath, pubSharePath, nil
}

// GetNLSSOutputPath constructs the output path for the private share
// Output format: {outputDir}/{did}/pvtShare.png
func (c *Config) GetNLSSOutputPath(did string) (string, error) {
	if c.NLSSOutputDir == "" {
		return "", fmt.Errorf("NLSS_OUTPUT_DIR not configured")
	}

	// Create output directory structure: ./output/{did}/
	outputDir := filepath.Join(c.NLSSOutputDir, did)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return filepath.Join(outputDir, "pvtShare.png"), nil
}
