package i18n

import (
	"fmt"
	"reflect"
	"strings"
)

// ParseKeyPath splits a dot-separated key path into individual components
func ParseKeyPath(key string) []string {
	return strings.Split(key, ".")
}

// GetValue retrieves a value from a nested map using dot-separated key path
func GetValue(data map[string]interface{}, key string) (string, bool) {
	parts := ParseKeyPath(key)
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// This is the final part, return the value
			if value, exists := current[part]; exists {
				if str, ok := value.(string); ok {
					return str, true
				}
				return fmt.Sprintf("%v", value), true
			}
			return "", false
		}

		// Navigate to nested object
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}

	return "", false
}

// SetValue sets a value in a nested map using dot-separated key path
func SetValue(data map[string]interface{}, key, value string) {
	parts := ParseKeyPath(key)
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// This is the final part, set the value
			current[part] = value
			return
		}

		// Navigate to or create nested object
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// Replace with new map
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			}
		} else {
			// Create new nested map
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}
}

// DeleteValue removes a value from a nested map using dot-separated key path
func DeleteValue(data map[string]interface{}, key string) bool {
	parts := ParseKeyPath(key)
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// This is the final part, delete the value
			if _, exists := current[part]; exists {
				delete(current, part)
				return true
			}
			return false
		}

		// Navigate to nested object
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return false
			}
		} else {
			return false
		}
	}

	return false
}

// GetAllKeys recursively collects all keys from a nested map
func GetAllKeys(data map[string]interface{}, prefix string) []string {
	var keys []string

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if valueMap, ok := value.(map[string]interface{}); ok {
			// Recursively collect keys from nested objects
			keys = append(keys, GetAllKeys(valueMap, fullKey)...)
		} else {
			keys = append(keys, fullKey)
		}
	}

	return keys
}

// GetKeysUnderPrefix collects all keys that start with the given prefix
// If the prefix itself is a leaf key (no children), it returns only that key
func GetKeysUnderPrefix(data map[string]interface{}, prefix string) []string {
	// First, check if the prefix exists and is a leaf key
	if _, exists := GetValue(data, prefix); exists {
		// Check if this prefix has any children
		prefixParts := ParseKeyPath(prefix)
		current := data

		// Navigate to the prefix location
		for i, part := range prefixParts {
			if i == len(prefixParts)-1 {
				// This is the final part
				if nextMap, ok := current[part].(map[string]interface{}); ok {
					// This prefix has children, so collect all child keys
					return GetAllKeys(nextMap, prefix)
				} else {
					// This is a leaf key, return only itself
					return []string{prefix}
				}
			}

			if nextMap, ok := current[part].(map[string]interface{}); ok {
				current = nextMap
			} else {
				// Path doesn't exist or is not a map
				break
			}
		}

		// If we get here, check if the value exists
		if exists {
			return []string{prefix}
		}
	}

	// If prefix doesn't exist as a leaf, try to find it as a parent key
	prefixParts := ParseKeyPath(prefix)
	current := data

	// Navigate to the parent of the prefix
	for i, part := range prefixParts {
		if nextMap, ok := current[part].(map[string]interface{}); ok {
			if i == len(prefixParts)-1 {
				// Found the prefix, collect all its children
				return GetAllKeys(nextMap, prefix)
			}
			current = nextMap
		} else {
			// Path doesn't exist
			break
		}
	}

	// Prefix doesn't exist, return empty
	return []string{}
}

// ExpandKeys expands a list of keys by finding all child keys for non-leaf keys
func ExpandKeys(data map[string]interface{}, keys []string) []string {
	var expandedKeys []string
	seenKeys := make(map[string]bool) // To avoid duplicates

	for _, key := range keys {
		childKeys := GetKeysUnderPrefix(data, key)
		for _, childKey := range childKeys {
			if !seenKeys[childKey] {
				expandedKeys = append(expandedKeys, childKey)
				seenKeys[childKey] = true
			}
		}
	}

	return expandedKeys
}

// IsEmptyMap checks if a map is empty (has no string values)
func IsEmptyMap(data map[string]interface{}) bool {
	return len(getStringValues(data)) == 0
}

// CleanEmptyMaps removes empty maps from the data structure
func CleanEmptyMaps(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if valueMap, ok := value.(map[string]interface{}); ok {
			cleaned := CleanEmptyMaps(valueMap)
			if len(cleaned) > 0 {
				result[key] = cleaned
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// getStringValues recursively collects all string values from a nested map
func getStringValues(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if valueMap, ok := value.(map[string]interface{}); ok {
			nested := getStringValues(valueMap)
			for nestedKey, nestedValue := range nested {
				fullKey := key + "." + nestedKey
				result[fullKey] = nestedValue
			}
		} else if !isZeroValue(value) {
			result[key] = value
		}
	}

	return result
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.String() == ""
	case reflect.Map, reflect.Slice, reflect.Array:
		return rv.Len() == 0
	default:
		return false
	}
}