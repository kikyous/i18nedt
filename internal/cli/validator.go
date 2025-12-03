package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kikyous/i18nedt/pkg/types"
)

// ValidateConfig validates the configuration and returns any errors
func ValidateConfig(config *types.Config) error {
	if len(config.Files) == 0 {
		return fmt.Errorf("no files specified")
	}

	// In flatten mode, keys are not required
	if !config.Flatten && len(config.Keys) == 0 {
		return fmt.Errorf("no keys specified")
	}

	// Validate each file path
	for _, file := range config.Files {
		if err := validateFilePath(file); err != nil {
			return fmt.Errorf("invalid file path %s: %w", file, err)
		}
	}

	// Validate each key (only if not in flatten mode)
	if !config.Flatten {
		for _, key := range config.Keys {
			if err := validateKey(key); err != nil {
				return fmt.Errorf("invalid key %s: %w", key, err)
			}
		}
	}

	return nil
}

// validateFilePath checks if a file path is valid
func validateFilePath(path string) error {
	// Check if path is absolute
	if !filepath.IsAbs(path) {
		// Convert to absolute path for validation
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("cannot resolve absolute path: %w", err)
		}
		path = abs
	}

	// Check if directory exists
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if file has .json extension
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		return fmt.Errorf("file must have .json extension")
	}

	return nil
}

// validateKey checks if a key is valid
func validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	return nil
}