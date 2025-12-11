package flatten

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// FlattenJSON flattens JSON data and returns key-value pairs
func FlattenJSON(data []byte, namespace string) (map[string]string, error) {
	// Parse JSON into interface{}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Determine prefix
	prefix := ""
	if namespace != "" {
		prefix = namespace + ":"
	}

	// Start recursive traversal and output
	flat := make(map[string]string)
	traverse(result, "", prefix, flat)
	return flat, nil
}

// traverse recursively traverses JSON structure and prints paths
func traverse(data interface{}, path, prefix string, result map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Handle object (Map)
		// Sort keys for stable output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := v[k]
			newPath := k
			if path != "" {
				newPath = path + "." + k
			}
			traverse(val, newPath, prefix, result)
		}

	case []interface{}:
		// Handle array (Slice)
		for i, val := range v {
			newPath := fmt.Sprintf("%d", i)
			if path != "" {
				newPath = path + "." + newPath
			}
			traverse(val, newPath, prefix, result)
		}

	default:
		// Handle basic types (String, Number, Bool, Null)
		// Use json.Marshal to convert value back to JSON format string
		// This way strings will have quotes (e.g., "Start"), numbers remain as-is
		valBytes, err := json.Marshal(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format value: %v\n", err)
			return
		}
		fullPath := prefix + path
		result[fullPath] = string(valBytes)
	}
}
