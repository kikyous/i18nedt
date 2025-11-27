package i18n

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Helper function to convert map to JSON string for testing
func mapToJSONString(data map[string]interface{}) string {
	if data == nil || len(data) == 0 {
		return "{}"
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func TestParseKeyPath(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected []string
	}{
		{
			name:     "simple key",
			key:      "welcome",
			expected: []string{"welcome"},
		},
		{
			name:     "nested key",
			key:      "home.welcome",
			expected: []string{"home", "welcome"},
		},
		{
			name:     "deeply nested key",
			key:      "nav.menu.home.title",
			expected: []string{"nav", "menu", "home", "title"},
		},
		{
			name:     "single character key",
			key:      "a",
			expected: []string{"a"},
		},
		{
			name:     "empty key",
			key:      "",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseKeyPath(tt.key)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseKeyPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		want     string
		wantOk   bool
	}{
		{
			name: "simple key exists",
			data: map[string]interface{}{
				"welcome": "Welcome",
			},
			key:    "welcome",
			want:   "Welcome",
			wantOk: true,
		},
		{
			name: "simple key doesn't exist",
			data: map[string]interface{}{
				"goodbye": "Goodbye",
			},
			key:    "welcome",
			want:   "",
			wantOk: false,
		},
		{
			name: "nested key exists",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome Home",
				},
			},
			key:    "home.welcome",
			want:   "Welcome Home",
			wantOk: true,
		},
		{
			name: "deeply nested key exists",
			data: map[string]interface{}{
				"nav": map[string]interface{}{
					"menu": map[string]interface{}{
						"home": map[string]interface{}{
							"title": "Home",
						},
					},
				},
			},
			key:    "nav.menu.home.title",
			want:   "Home",
			wantOk: true,
		},
		{
			name: "nested key path doesn't exist",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome",
				},
			},
			key:    "home.goodbye",
			want:   "",
			wantOk: false, // Key doesn't exist, should return empty string and nil error
		},
		{
			name: "intermediate path is not a map",
			data: map[string]interface{}{
				"home": "not a map",
			},
			key:    "home.welcome",
			want:   "",
			wantOk: false, // Key doesn't exist, should return empty string and nil error
		},
		{
			name: "non-string value",
			data: map[string]interface{}{
				"count": 42,
			},
			key:    "count",
			want:   "42",
			wantOk: true,
		},
		{
			name: "boolean value",
			data: map[string]interface{}{
				"enabled": true,
			},
			key:    "enabled",
			want:   "true",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert map to JSON string for testing
			jsonStr := mapToJSONString(tt.data)
			got, err := GetValue(jsonStr, tt.key)
			wantOk := err == nil && got != ""
			if wantOk != tt.wantOk {
				t.Errorf("GetValue() ok = %v, want %v", wantOk, tt.wantOk)
				return
			}
			if got != tt.want {
				t.Errorf("GetValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		value    string
		expected map[string]interface{}
	}{
		{
			name: "set simple key",
			data: map[string]interface{}{},
			key:  "welcome",
			value: "Welcome",
			expected: map[string]interface{}{
				"welcome": "Welcome",
			},
		},
		{
			name: "set nested key in empty data",
			data: map[string]interface{}{},
			key:  "home.welcome",
			value: "Welcome Home",
			expected: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome Home",
				},
			},
		},
		{
			name: "set deeply nested key",
			data: map[string]interface{}{},
			key:  "nav.menu.home.title",
			value: "Home",
			expected: map[string]interface{}{
				"nav": map[string]interface{}{
					"menu": map[string]interface{}{
						"home": map[string]interface{}{
							"title": "Home",
						},
					},
				},
			},
		},
		{
			name: "overwrite existing simple key",
			data: map[string]interface{}{
				"welcome": "Old Welcome",
			},
			key:  "welcome",
			value: "New Welcome",
			expected: map[string]interface{}{
				"welcome": "New Welcome",
			},
		},
		{
			name: "overwrite existing nested key",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Old Welcome",
				},
			},
			key:  "home.welcome",
			value: "New Welcome",
			expected: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "New Welcome",
				},
			},
		},
		{
			name: "replace non-map with map",
			data: map[string]interface{}{
				"home": "not a map",
			},
			key:  "home.welcome",
			value: "Welcome",
			expected: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome",
				},
			},
		},
		{
			name: "add to existing nested structure",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome",
				},
			},
			key:  "home.goodbye",
			value: "Goodbye",
			expected: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome",
					"goodbye": "Goodbye",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := mapToJSONString(tt.data)
			newJson, err := SetValue(jsonStr, tt.key, tt.value)
			if err != nil {
				t.Errorf("SetValue() error = %v", err)
				return
			}
			// Parse both JSON objects to compare their content regardless of key order
			var resultData, expectedData map[string]interface{}
			json.Unmarshal([]byte(newJson), &resultData)
			json.Unmarshal([]byte(mapToJSONString(tt.expected)), &expectedData)

			if !reflect.DeepEqual(resultData, expectedData) {
				t.Errorf("SetValue() result = %v, want %v", resultData, expectedData)
			}
		})
	}
}

func TestDeleteValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected map[string]interface{}
		wantOk   bool
	}{
		{
			name: "delete existing simple key",
			data: map[string]interface{}{
				"welcome": "Welcome",
				"goodbye": "Goodbye",
			},
			key: "welcome",
			expected: map[string]interface{}{
				"goodbye": "Goodbye",
			},
			wantOk: true,
		},
		{
			name: "delete existing nested key",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"welcome": "Welcome",
					"goodbye": "Goodbye",
				},
			},
			key: "home.welcome",
			expected: map[string]interface{}{
				"home": map[string]interface{}{
					"goodbye": "Goodbye",
				},
			},
			wantOk: true,
		},
		{
			name: "delete non-existent simple key",
			data: map[string]interface{}{
				"goodbye": "Goodbye",
			},
			key: "welcome",
			expected: map[string]interface{}{
				"goodbye": "Goodbye",
			},
			wantOk: true, // sjson.Delete doesn't return error for non-existent keys
		},
		{
			name: "delete non-existent nested key",
			data: map[string]interface{}{
				"home": map[string]interface{}{
					"goodbye": "Goodbye",
				},
			},
			key: "home.welcome",
			expected: map[string]interface{}{
				"home": map[string]interface{}{
					"goodbye": "Goodbye",
				},
			},
			wantOk: true, // sjson.Delete doesn't return error for non-existent keys
		},
		{
			name: "delete from non-existent path",
			data: map[string]interface{}{
				"welcome": "Welcome",
			},
			key: "home.welcome",
			expected: map[string]interface{}{
				"welcome": "Welcome",
			},
			wantOk: true, // sjson.Delete doesn't return error for non-existent keys
		},
		{
			name: "delete deeply nested key",
			data: map[string]interface{}{
				"nav": map[string]interface{}{
					"menu": map[string]interface{}{
						"home": map[string]interface{}{
							"title": "Home",
							"desc":  "Description",
						},
						"about": map[string]interface{}{
							"title": "About",
						},
					},
				},
			},
			key: "nav.menu.home.title",
			expected: map[string]interface{}{
				"nav": map[string]interface{}{
					"menu": map[string]interface{}{
						"home": map[string]interface{}{
							"desc": "Description",
						},
						"about": map[string]interface{}{
							"title": "About",
						},
					},
				},
			},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := mapToJSONString(tt.data)
			newJson, err := DeleteValue(jsonStr, tt.key)
			ok := err == nil
			if ok != tt.wantOk {
				t.Errorf("DeleteValue() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			expectedJson := mapToJSONString(tt.expected)
			if newJson != expectedJson {
				t.Errorf("DeleteValue() result = %v, want %v", newJson, expectedJson)
			}
		})
	}
}

// GetAllKeys function removed as it doesn't exist in the codebase

// IsEmptyMap function removed as it doesn't exist in the codebase

// CleanEmptyMaps function removed as it doesn't exist in the codebase