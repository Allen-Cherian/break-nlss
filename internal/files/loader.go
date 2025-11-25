package files

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileInfo represents information about a loaded file
type FileInfo struct {
	Path   string
	Exists bool
	Size   int64
}

// CheckFile checks if a file exists and returns its info
func CheckFile(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &FileInfo{
			Path:   path,
			Exists: false,
			Size:   0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &FileInfo{
		Path:   path,
		Exists: true,
		Size:   info.Size(),
	}, nil
}

// ValidatePresetFolder validates that all required files exist in the preset folder
func ValidatePresetFolder(presetFolder string) error {
	requiredFiles := []string{
		"privatekey.pem",
		"PrivateShare.png",
	}

	optionalFiles := []string{
		"PublicShare.png",
		"DID.png",
	}

	// Check if preset folder exists
	if _, err := os.Stat(presetFolder); os.IsNotExist(err) {
		return fmt.Errorf("preset folder does not exist: %s", presetFolder)
	}

	// Check required files
	for _, filename := range requiredFiles {
		path := filepath.Join(presetFolder, filename)
		info, err := CheckFile(path)
		if err != nil {
			return fmt.Errorf("error checking %s: %w", filename, err)
		}
		if !info.Exists {
			return fmt.Errorf("required file missing: %s", path)
		}
	}

	// Check optional files (just warn, don't error)
	for _, filename := range optionalFiles {
		path := filepath.Join(presetFolder, filename)
		info, _ := CheckFile(path)
		if !info.Exists {
			fmt.Printf("Warning: Optional file not found: %s\n", path)
		}
	}

	return nil
}

// EnsurePresetFolder creates the preset folder if it doesn't exist
func EnsurePresetFolder(presetFolder string) error {
	if err := os.MkdirAll(presetFolder, 0755); err != nil {
		return fmt.Errorf("failed to create preset folder: %w", err)
	}
	return nil
}

// ListPresetFiles lists all files in the preset folder
func ListPresetFiles(presetFolder string) ([]string, error) {
	entries, err := os.ReadDir(presetFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset folder: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
