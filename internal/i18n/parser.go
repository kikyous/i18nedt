package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kikyous/i18nedt/pkg/types"
	"github.com/tidwall/gjson"
)

// LoadFile loads and parses an i18n JSON file
func LoadFile(filePath string, pattern string) (*types.I18nFile, error) {
	// Determine locale and namespace
	var locale, namespace string
	var err error

	if pattern != "" {
		// Use pattern-based extraction
		locale, namespace, err = ExtractMetadataFromPath(filePath, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to extract metadata from path %s using pattern %s: %w", filePath, pattern, err)
		}
	} else {
		// Non-NS mode: extract locale using BCP47 parsing or fallback
		locale, err = ParseLocaleFromPath(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract locale from path %s: %w", filePath, err)
		}
		namespace = ""
	}

	file := &types.I18nFile{
		Path:      filePath,
		Data:      "{}", // Default empty JSON object
		Locale:    locale,
		Namespace: namespace,
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, return empty file
		return file, nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Validate JSON content
	jsonStr := string(data)
	if jsonStr == "" {
		file.Data = "{}"
	} else if !gjson.Valid(jsonStr) {
		return nil, fmt.Errorf("invalid JSON in file %s", filePath)
	} else {
		file.Data = jsonStr
	}

	return file, nil
}

// SaveFile saves an i18n file to disk
func SaveFile(file *types.I18nFile) error {
	// Ensure JSON is valid
	if !gjson.Valid(file.Data) {
		return fmt.Errorf("invalid JSON data for file %s", file.Path)
	}

	// Ensure directory exists
	dir := GetDirectory(file.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Format JSON with proper indentation
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, []byte(file.Data), "", "  "); err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	// Write to temporary file first
	tempFile := file.Path + ".tmp"
	if err := os.WriteFile(tempFile, formatted.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tempFile, err)
	}

	// Rename to final file (atomic operation)
	if err := os.Rename(tempFile, file.Path); err != nil {
		// Clean up temp file if rename fails
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temporary file to %s: %w", file.Path, err)
	}

	return nil
}

// SaveAllFiles saves multiple i18n files
func SaveAllFiles(files []*types.I18nFile) (int, error) {
	count := 0
	for _, file := range files {
		// Only save if file is dirty (modified)
		if file.Dirty {
			if err := SaveFile(file); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// GetDirectory returns the directory part of a file path
func GetDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

// FileSource represents a file to load and the pattern used to find it (if any)
type FileSource struct {
	Path    string
	Pattern string
}

// LoadAllFiles loads multiple i18n files
func LoadAllFiles(sources []FileSource) ([]*types.I18nFile, error) {
	files := make([]*types.I18nFile, 0, len(sources))

	for _, src := range sources {
		file, err := LoadFile(src.Path, src.Pattern)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// BackupFile creates a backup of the specified file
func BackupFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, no need to backup
		return nil
	}

	backupPath := filePath + ".backup"
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	return nil
}
