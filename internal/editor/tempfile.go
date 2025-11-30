package editor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kikyous/i18nedt/internal/i18n"
	"github.com/kikyous/i18nedt/pkg/types"
	"github.com/tidwall/gjson"
)

// CreateTempFile creates a temporary file for editing
func CreateTempFile(files []*types.I18nFile, keys []string) (*types.TempFile, error) {
	// Get locale list from I18nFile structs
	locales, err := i18n.GetLocaleList(files)
	if err != nil {
		return nil, fmt.Errorf("failed to extract locales: %w", err)
	}

	// Create temporary file in current directory with dot prefix
	tempFileName := fmt.Sprintf(".i18nedt-%d.md", time.Now().Unix())

	temp := &types.TempFile{
		Path:    tempFileName,
		Keys:    keys,
		Locales: locales,
		Content: make(map[string]map[string]*types.Value),
		Deletes: []string{},
	}

	// Initialize content map
	for _, key := range keys {
		temp.Content[key] = make(map[string]*types.Value)
		for _, locale := range locales {
			temp.Content[key][locale] = types.NewStringValue("")
		}
	}

	// Load existing values from files
	for _, key := range keys {
		for _, file := range files {
			if value, err := i18n.GetValueTyped(file.Data, key); err == nil {
				temp.Content[key][file.Locale] = value
			}
		}
	}

	return temp, nil
}

// GenerateTempFileContent generates the content for the temporary file
func GenerateTempFileContent(temp *types.TempFile) ([]byte, error) {
	return GenerateTempFileContentWithOptions(temp, false)
}

// GenerateTempFileContentWithOptions generates the content for the temporary file with options
func GenerateTempFileContentWithOptions(temp *types.TempFile, noTips bool) ([]byte, error) {
	var builder strings.Builder

	// Add header comments
	if !noTips {
		builder.WriteString("you are a md file translator, add missing translations to this file.\n")
		builder.WriteString("key start with # and language start with * or +.\n")
		builder.WriteString("do not read or edit other file.(this is a tip for ai)\n\n")
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(temp.Content))
	for k := range temp.Content {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("# %s\n", key))

		localeValues := temp.Content[key]

		// Sort locales for consistent output
		locales := make([]string, 0, len(localeValues))
		for locale := range localeValues {
			locales = append(locales, locale)
		}
		sort.Strings(locales)

		for _, locale := range locales {
			value := localeValues[locale]

			// Use appropriate marker based on value type
			var marker string
			switch value.Type {
			case types.ValueTypeJSON:
				marker = "+"
			default:
				marker = "*"
			}

			builder.WriteString(fmt.Sprintf("%s %s\n", marker, locale))

			// For JSON values, format with proper indentation
			if value.Type == types.ValueTypeJSON {
				var formattedJSON []byte
				if gjson.Valid(value.Value) {
					formattedJSON, _ = json.MarshalIndent(gjson.Parse(value.Value).Value(), "", "  ")
				} else {
					formattedJSON = []byte(value.Value)
				}
				builder.WriteString(string(formattedJSON))
				builder.WriteString("\n")
			} else {
				builder.WriteString(value.Value)
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
	}

	// Add deletion markers
	for _, deleteKey := range temp.Deletes {
		builder.WriteString(fmt.Sprintf("#- %s\n", deleteKey))
	}

	return []byte(builder.String()), nil
}

// WriteTempFile writes the temporary file
func WriteTempFile(temp *types.TempFile) error {
	return WriteTempFileWithOptions(temp, false)
}

// WriteTempFileWithOptions writes the temporary file with options
func WriteTempFileWithOptions(temp *types.TempFile, noTips bool) error {
	content, err := GenerateTempFileContentWithOptions(temp, noTips)
	if err != nil {
		return fmt.Errorf("failed to generate temp file content: %w", err)
	}

	return ioutil.WriteFile(temp.Path, content, 0644)
}

// ParseTempFileContent parses the content of the edited temporary file
func ParseTempFileContent(content string, locales []string) (*types.TempFile, error) {
	lines := strings.Split(content, "\n")
	temp := &types.TempFile{
		Content: make(map[string]map[string]*types.Value),
		Deletes: []string{},
	}

	var currentKey string
	var currentLocale string
	var currentValue strings.Builder
	var isJSONValue bool

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Key line
		if strings.HasPrefix(line, "#") {
			// Save previous value if any
			if currentKey != "" && currentLocale != "" {
				if err := saveValue(temp, currentKey, currentLocale, currentValue.String(), isJSONValue); err != nil {
					return nil, fmt.Errorf("line %d: %w", i, err)
				}
			}

			// Handle deletion marker
			if strings.HasPrefix(line, "#-") {
				deleteKey := strings.TrimSpace(line[2:])
				temp.Deletes = append(temp.Deletes, deleteKey)
				currentKey = ""
				currentLocale = ""
				continue
			}

			// New key
			currentKey = strings.TrimSpace(line[1:])
			if temp.Content[currentKey] == nil {
				temp.Content[currentKey] = make(map[string]*types.Value)
			}
			currentLocale = ""
			currentValue.Reset()
			continue
		}

		// Locale line
		if strings.HasPrefix(line, "*") || strings.HasPrefix(line, "+") {
			// Save previous value if any
			if currentKey != "" && currentLocale != "" {
				if err := saveValue(temp, currentKey, currentLocale, currentValue.String(), isJSONValue); err != nil {
					return nil, fmt.Errorf("line %d: %w", i, err)
				}
			}

			// New locale
			parts := strings.Fields(line[1:])
			if len(parts) == 0 {
				return nil, fmt.Errorf("line %d: invalid locale format", i)
			}

			currentLocale = parts[0]
			isJSONValue = strings.HasPrefix(line, "+")
			currentValue.Reset()
			continue
		}

		// Value line
		if currentKey != "" && currentLocale != "" {
			if currentValue.Len() > 0 {
				currentValue.WriteString("\n")
			}
			currentValue.WriteString(line)
		}
	}

	// Save last value
	if currentKey != "" && currentLocale != "" {
		if err := saveValue(temp, currentKey, currentLocale, currentValue.String(), isJSONValue); err != nil {
			return nil, fmt.Errorf("end of file: %w", err)
		}
	}

	return temp, nil
}

func saveValue(temp *types.TempFile, key, locale, value string, isJSON bool) error {
	var v *types.Value
	if isJSON {
		// Validate JSON content
		if !gjson.Valid(value) {
			return fmt.Errorf("invalid JSON content for key '%s', locale '%s'", key, locale)
		}
		v = types.NewJSONValue(value)
	} else {
		v = types.NewStringValue(value)
	}

	temp.Content[key][locale] = v
	return nil
}

// ReadTempFile reads and parses the temporary file
func ReadTempFile(temp *types.TempFile) error {
	content, err := ioutil.ReadFile(temp.Path)
	if err != nil {
		return fmt.Errorf("failed to read temporary file: %w", err)
	}

	parsedTemp, err := ParseTempFileContent(string(content), temp.Locales)
	if err != nil {
		return fmt.Errorf("failed to parse temp file content: %w", err)
	}

	// Update the original temp with the parsed content
	// Note: Don't overwrite temp.Path as ParseTempFileContent doesn't set it
	temp.Keys = parsedTemp.Keys
	temp.Locales = parsedTemp.Locales
	temp.Content = parsedTemp.Content
	temp.Deletes = parsedTemp.Deletes

	return nil
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
			newData, err := i18n.DeleteValue(file.Data, keyToDelete)
			if err == nil {
				file.Data = newData
			}
		}
	}

	// Handle updates and additions
	for key, localeValues := range temp.Content {
		for _, file := range files {
			value, exists := localeValues[file.Locale]
			if !exists {
				continue // Skip if no value for this locale
			}

			newData, err := i18n.SetValueTyped(file.Data, key, value)
			if err == nil {
				file.Data = newData
			}
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
