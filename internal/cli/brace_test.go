package cli

import (
	"strings"
	"testing"
)

func TestExpandBraces(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "simple_brace_expansion",
			pattern:  "{a,b,c}",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "brace_with_prefix",
			pattern:  "prefix_{zh,en}.json",
			expected: []string{"prefix_zh.json", "prefix_en.json"},
		},
		{
			name:     "brace_with_suffix",
			pattern:  "file.{json,yml}",
			expected: []string{"file.json", "file.yml"},
		},
		{
			name:     "brace_with_prefix_and_suffix",
			pattern:  "src/locales/{zh-CN,en-US}.json",
			expected: []string{"src/locales/zh-CN.json", "src/locales/en-US.json"},
		},
		{
			name:     "multiple_braces",
			pattern:  "{a,b}.{x,y}",
			expected: []string{"a.x", "a.y", "b.x", "b.y"},
		},
		{
			name:     "empty_alternative",
			pattern:  "{a,}.txt",
			expected: []string{"a.txt", ".txt"},
		},
		{
			name:     "single_alternative",
			pattern:  "{a}.txt",
			expected: []string{"a.txt"},
		},
		{
			name:     "no_braces",
			pattern:  "normal_file.txt",
			expected: []string{"normal_file.txt"},
		},
		{
			name:     "unmatched_open_brace",
			pattern:  "file_{unmatched.txt",
			expected: []string{"file_{unmatched.txt"},
		},
		{
			name:     "unmatched_close_brace",
			pattern:  "file}_matched.txt",
			expected: []string{"file}_matched.txt"},
		},
		{
			name:     "empty_braces",
			pattern:  "file{}.txt",
			expected: []string{"file.txt"},
		},
		{
			name:     "simple_nested_braces",
			pattern:  "prefix_{a,b}.txt",
			expected: []string{"prefix_a.txt", "prefix_b.txt"},
		},
		{
			name:     "complex_nested_braces",
			pattern:  "{a,b}{c,d}",
			expected: []string{"ac", "ad", "bc", "bd"},
		},
		{
			name:     "three_alternatives",
			pattern:  "{zh-CN,zh-TW,en-US}",
			expected: []string{"zh-CN", "zh-TW", "en-US"},
		},
		{
			name:     "real_world_example",
			pattern:  "assets/js/vue/i18n/{zh-CN,zh-TW,zh-MO,en-US}.json",
			expected: []string{
				"assets/js/vue/i18n/zh-CN.json",
				"assets/js/vue/i18n/zh-TW.json",
				"assets/js/vue/i18n/zh-MO.json",
				"assets/js/vue/i18n/en-US.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandBraces(tt.pattern)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d results, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got:      %v", result)
				return
			}

			// Convert to sets for comparison
			expectedMap := make(map[string]bool)
			for _, item := range tt.expected {
				expectedMap[item] = true
			}

			resultMap := make(map[string]bool)
			for _, item := range result {
				resultMap[item] = true
			}

			for item := range expectedMap {
				if !resultMap[item] {
					t.Errorf("Missing expected result: %s", item)
				}
			}

			for item := range resultMap {
				if !expectedMap[item] {
					t.Errorf("Unexpected result: %s", item)
				}
			}
		})
	}
}

func TestExpandFilePathsWithBraces(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		expectError bool
		errorMsg    string
	}{
		{
			name:  "brace_expansion_with_existing_files",
			paths: []string{"test-locales/{zh-CN,zh-TW,zh-MO,en-US}.json"},
		},
		{
			name:  "brace_expansion_with_nonexistent_files",
			paths: []string{"nonexistent{a,b,c}.txt"},
		},
		{
			name:        "brace_and_glob_with_no_matches",
			paths:       []string{"nonexistent{a,b,c}*.txt"},
			expectError: true,
			errorMsg:    "no files match pattern",
		},
		{
			name:  "simple_brace_patterns",
			paths: []string{"test-locales/{zh-CN,en-US}.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandFilePaths(tt.paths)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) == 0 && !tt.expectError {
				t.Errorf("Expected at least one result")
			}
		})
	}
}