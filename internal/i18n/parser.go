package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/chen/i18nedt/pkg/types"
)

// LoadFile loads and parses an i18n JSON file
func LoadFile(filePath string) (*types.I18nFile, error) {
	file := &types.I18nFile{
		Path: filePath,
		Data: make(map[string]interface{}),
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, return empty file
		return file, nil
	}

	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse JSON
	if len(data) > 0 {
		if err := json.Unmarshal(data, &file.Data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON in file %s: %w", filePath, err)
		}
	}

	return file, nil
}

// SaveFile saves an i18n file to disk
func SaveFile(file *types.I18nFile) error {
	// Clean up empty maps
	file.Data = CleanEmptyMaps(file.Data)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(file.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for file %s: %w", file.Path, err)
	}

	// Ensure directory exists
	dir := GetDirectory(file.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to temporary file first
	tempFile := file.Path + ".tmp"
	if err := ioutil.WriteFile(tempFile, jsonData, 0644); err != nil {
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
func SaveAllFiles(files []*types.I18nFile) error {
	for _, file := range files {
		if err := SaveFile(file); err != nil {
			return err
		}
	}
	return nil
}

// GetDirectory returns the directory part of a file path
func GetDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

// LoadAllFiles loads multiple i18n files
func LoadAllFiles(filePaths []string) ([]*types.I18nFile, error) {
	files := make([]*types.I18nFile, 0, len(filePaths))

	for _, filePath := range filePaths {
		file, err := LoadFile(filePath)
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
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := ioutil.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	return nil
}