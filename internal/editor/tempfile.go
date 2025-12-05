package editor

import (
	"encoding/json"
	"fmt"
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
		Keys:    keys, // Note: These are original requested keys
		Locales: locales,
		Content: make(map[string]map[string]*types.Value),
		Deletes: []string{},
	}

	// Iterate over requested keys
	for _, key := range keys {
		// Check if the requested key implies a specific namespace
		reqNs, reqKey := splitNamespaceKey(key)

		for _, file := range files {
			// If user requested a specific namespace, skip files that don't match
			if reqNs != "" && file.Namespace != reqNs {
				continue
			}

			// Determine the key to display in the editor
			// If the file has a namespace, prepend it to the key
			displayKey := reqKey
			if file.Namespace != "" {
				displayKey = file.Namespace + ":" + reqKey
			}

			// Initialize the content map for this display key if not present
			if _, ok := temp.Content[displayKey]; !ok {
				temp.Content[displayKey] = make(map[string]*types.Value)
				// Initialize with empty values for all locales to ensure matrix is complete
				for _, l := range locales {
					temp.Content[displayKey][l] = types.NewStringValue("")
				}
			}

			// Try to retrieve value from file
			// We look up using the 'reqKey' (which is the key without namespace prefix if one was provided)
			// But if reqNs was empty, reqKey is just the key.
			// Correct logic: The key inside the file is always the "key part".
			if value, err := i18n.GetValueTyped(file.Data, reqKey); err == nil {
				// Only set if we found something? Or strictly set?
				// GetValueTyped returns NewStringValue("") if not found (and no error if Valid JSON).
				// But we want to know if it *exists*?
				// Actually GetValueTyped uses gjson.Get.
				// If we want to distinguish "empty string" from "missing", we might need check.
				// But for now, just overwriting with what we found is fine.
				// If it returns empty string for missing, we effectively propose adding it.
				
				// However, GetValueTyped implementation:
				// result := gjson.Get(jsonStr, key)
				// if !result.Exists() { return types.NewStringValue(""), nil }
				// So it returns empty string if missing.
				
				// If we have multiple files for same locale (e.g. different namespaces), 
				// loop ensures we pick the right one because we check file.Namespace above.
				// But wait, 'locales' list contains ALL locales across ALL files.
				// temp.Content[displayKey] has entries for ALL locales.
				// We are assigning to temp.Content[displayKey][file.Locale].
				// This is correct.
				temp.Content[displayKey][file.Locale] = value
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

	return os.WriteFile(temp.Path, content, 0644)
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
	content, err := os.ReadFile(temp.Path)
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
		targetNs, targetKey := splitNamespaceKey(keyToDelete)

		for _, file := range files {
			// Check if file matches namespace (empty targetNs matches empty file.Namespace)
			if file.Namespace != targetNs {
				continue
			}

			newData, err := i18n.DeleteValue(file.Data, targetKey)
			if err == nil && newData != file.Data {
				file.Data = newData
				file.Dirty = true
			}
		}
	}

	// Handle updates and additions
	for key, localeValues := range temp.Content {
		targetNs, targetKey := splitNamespaceKey(key)

		for _, file := range files {
			// Check if file matches namespace
			if file.Namespace != targetNs {
				continue
			}

			value, exists := localeValues[file.Locale]
			if !exists {
				continue // Skip if no value for this locale
			}

			// Check if value actually changed to avoid marking file as dirty unnecessarily
			currentVal, err := i18n.GetValueTyped(file.Data, targetKey)
			if err == nil && currentVal.Value == value.Value && currentVal.Type == value.Type {
				continue
			}

			newData, err := i18n.SetValueTyped(file.Data, targetKey, value)
			if err == nil && newData != file.Data {
				file.Data = newData
				file.Dirty = true
			}
		}
	}

	return nil
}

// Helper function to split "namespace:key" into "namespace" and "key"
// If no namespace, returns empty string and original key
func splitNamespaceKey(compositeKey string) (string, string) {
	parts := strings.SplitN(compositeKey, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", compositeKey
}

// Helper function to get file paths from I18nFile slice
func getFilePaths(files []*types.I18nFile) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}