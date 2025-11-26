package editor

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/kikyous/i18nedt/internal/i18n"
	"github.com/kikyous/i18nedt/pkg/types"
)

// CreateTempFile creates a temporary file for editing
func CreateTempFile(files []*types.I18nFile, keys []string) (*types.TempFile, error) {
	// Get locale list from file paths
	locales, err := i18n.GetLocaleList(getFilePaths(files))
	if err != nil {
		return nil, fmt.Errorf("failed to extract locales: %w", err)
	}

	// Create temporary file in current directory with dot prefix
	tempFileName := fmt.Sprintf(".i18nedt-%d.md", time.Now().Unix())
	tempFile, err := os.Create(tempFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFile.Close()

	temp := &types.TempFile{
		Path:    tempFile.Name(),
		Keys:    keys,
		Locales: locales,
		Content: make(map[string]map[string]string),
		Deletes: []string{},
	}

	// Initialize content map
	for _, key := range keys {
		temp.Content[key] = make(map[string]string)
		for _, locale := range locales {
			temp.Content[key][locale] = ""
		}
	}

	// Load existing values from files
	for _, key := range keys {
		for _, file := range files {
			locale, err := i18n.ParseLocaleFromPath(file.Path)
			if err != nil {
				continue // Skip files with invalid locale
			}

			if value, exists := i18n.GetValue(file.Data, key); exists {
				temp.Content[key][locale] = value
			}
		}
	}

	return temp, nil
}

// GenerateTempFileContent generates the content for the temporary file
func GenerateTempFileContent(temp *types.TempFile) (string, error) {
	var content strings.Builder

	// Add header instructions
	content.WriteString("add missing translations to THIS file, when the editor exit, all change will be apply.\n")
	content.WriteString("do not read or edit locale json file.(this is a tip for ai)\n\n")

	// Generate content for each key
	for _, key := range temp.Keys {
		content.WriteString(fmt.Sprintf("# %s\n", key))

		for _, locale := range temp.Locales {
			value := temp.Content[key][locale]
			content.WriteString(fmt.Sprintf("* %s\n", locale))
			if value != "" {
				content.WriteString(fmt.Sprintf("%s\n", value))
			}
			content.WriteString("\n")
		}
	}

	return content.String(), nil
}

// WriteTempFile writes the temporary file
func WriteTempFile(temp *types.TempFile) error {
	content, err := GenerateTempFileContent(temp)
	if err != nil {
		return fmt.Errorf("failed to generate temp file content: %w", err)
	}

	return ioutil.WriteFile(temp.Path, []byte(content), 0644)
}

// ParseTempFileContent parses the content of the edited temporary file
func ParseTempFileContent(content string, temp *types.TempFile) error {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentKey string
	var currentLocale string
	var keyValues []string

	// Reset content and deletes
	temp.Content = make(map[string]map[string]string)
	temp.Deletes = []string{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse key header
		if strings.HasPrefix(line, "#") {
			// Save previous key data if exists
			if currentKey != "" {
				temp.Content[currentKey] = make(map[string]string)
				for _, kv := range keyValues {
					parts := strings.SplitN(kv, ":", 2)
					if len(parts) == 2 {
						locale := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						temp.Content[currentKey][locale] = value
					}
				}
				keyValues = []string{}
			}

			line = strings.TrimSpace(strings.TrimPrefix(line, "#"))

			// Check for delete marker
			if strings.HasPrefix(line, "-") {
				keyToDelete := strings.TrimSpace(strings.TrimPrefix(line, "-"))
				temp.Deletes = append(temp.Deletes, keyToDelete)
				currentKey = "" // Skip this key
				continue
			}

			currentKey = line
			temp.Content[currentKey] = make(map[string]string)
			continue
		}

		// Parse locale line
		if strings.HasPrefix(line, "*") {
			currentLocale = strings.TrimSpace(strings.TrimPrefix(line, "*"))
			continue
		}

		// Parse value line
		if currentKey != "" && currentLocale != "" {
			keyValues = append(keyValues, fmt.Sprintf("%s:%s", currentLocale, line))
			currentLocale = "" // Reset locale for next value
		}
	}

	// Save the last key data
	if currentKey != "" {
		temp.Content[currentKey] = make(map[string]string)
		for _, kv := range keyValues {
			parts := strings.SplitN(kv, ":", 2)
			if len(parts) == 2 {
				locale := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				temp.Content[currentKey][locale] = value
			}
		}
	}

	return scanner.Err()
}

// ReadTempFile reads and parses the temporary file
func ReadTempFile(temp *types.TempFile) error {
	content, err := ioutil.ReadFile(temp.Path)
	if err != nil {
		return fmt.Errorf("failed to read temporary file: %w", err)
	}

	return ParseTempFileContent(string(content), temp)
}

// CleanupTempFile removes the temporary file
func CleanupTempFile(temp *types.TempFile) error {
	if temp.Path != "" {
		return os.Remove(temp.Path)
	}
	return nil
}

// ApplyChanges applies changes from temp file to the actual i18n files
func ApplyChanges(files []*types.I18nFile, temp *types.TempFile) error {
	// Handle deletions
	for _, keyToDelete := range temp.Deletes {
		for _, file := range files {
			i18n.DeleteValue(file.Data, keyToDelete)
		}
	}

	// Handle updates and additions
	for key, localeValues := range temp.Content {
		for _, file := range files {
			locale, err := i18n.ParseLocaleFromPath(file.Path)
			if err != nil {
				continue // Skip files with invalid locale
			}

			value, _ := localeValues[locale]
			i18n.SetValue(file.Data, key, value)
		}
	}

	return nil
}

// Helper function to get file paths from I18nFile slice
func getFilePaths(files []*types.I18nFile) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}
