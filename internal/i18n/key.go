package i18n

import (
	"fmt"
	"strings"

	"github.com/kikyous/i18nedt/pkg/types"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ParseKeyPath splits a dot-separated key path into individual components
func ParseKeyPath(key string) []string {
	return strings.Split(key, ".")
}

// GetValue retrieves a value from JSON string using gjson
func GetValue(jsonStr, key string) (string, error) {
	if !gjson.Valid(jsonStr) {
		return "", fmt.Errorf("invalid JSON string")
	}

	result := gjson.Get(jsonStr, key)
	if !result.Exists() {
		return "", nil // Return empty string for non-existent keys
	}

	return result.String(), nil
}

// SetValue sets a value in JSON string using sjson
func SetValue(jsonStr, key, value string) (string, error) {
	if !gjson.Valid(jsonStr) {
		return "", fmt.Errorf("invalid JSON string")
	}

	newJson, err := sjson.Set(jsonStr, key, value)
	if err != nil {
		return "", fmt.Errorf("failed to set key '%s': %w", key, err)
	}

	return newJson, nil
}

// DeleteValue removes a key from JSON string using sjson
func DeleteValue(jsonStr, key string) (string, error) {
	if !gjson.Valid(jsonStr) {
		return "", fmt.Errorf("invalid JSON string")
	}

	newJson, err := sjson.Delete(jsonStr, key)
	if err != nil {
		return "", fmt.Errorf("failed to delete key '%s': %w", key, err)
	}

	return newJson, nil
}

// ValidateJSON checks if a string is valid JSON
func ValidateJSON(jsonStr string) error {
	if !gjson.Valid(jsonStr) {
		return fmt.Errorf("invalid JSON string")
	}
	return nil
}

// GetValueTyped retrieves a value with type information
func GetValueTyped(jsonStr, key string) (*types.Value, error) {
	if !gjson.Valid(jsonStr) {
		return nil, fmt.Errorf("invalid JSON string")
	}

	result := gjson.Get(jsonStr, key)
	if !result.Exists() {
		return types.NewStringValue(""), nil
	}

	switch result.Type {
	case gjson.JSON:
		return types.NewJSONValue(result.Raw), nil
	default:
		return types.NewStringValue(result.String()), nil
	}
}

// SetValueTyped sets a value with proper type handling
func SetValueTyped(jsonStr, key string, value *types.Value) (string, error) {
	if !gjson.Valid(jsonStr) {
		return "", fmt.Errorf("invalid JSON string")
	}

	var result string
	var err error

	switch value.Type {
	case types.ValueTypeJSON:
		// For JSON values, we need to parse and set as object/array
		if !gjson.Valid(value.Value) {
			return "", fmt.Errorf("invalid JSON content for key '%s'", key)
		}
		result, err = sjson.Set(jsonStr, key, gjson.Parse(value.Value).Value())
	default:
		// For string values, set as string
		result, err = sjson.Set(jsonStr, key, value.Value)
	}

	if err != nil {
		return "", fmt.Errorf("failed to set key '%s': %w", key, err)
	}

	return result, nil
}
