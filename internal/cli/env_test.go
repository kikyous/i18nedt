package cli

import (
	"os"
	"strings"
	"testing"
)

func TestParseFlagsWithEnvVar(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		args           []string
		expectedFiles  []string
		expectedKeys   []string
		expectError    bool
		errorContains  string
	}{
		{
			name:          "env_var_with_space_separated_files",
			envValue:      "test1.json test2.json",
			args:          []string{"-k", "home.welcome"},
			expectedFiles: []string{"test1.json", "test2.json"},
			expectedKeys:  []string{"home.welcome"},
			expectError:   false,
		},
		{
			name:          "env_var_with_colon_separated_files",
			envValue:      "test1.json:test2.json",
			args:          []string{"-k", "home.welcome"},
			expectedFiles: []string{"test1.json", "test2.json"},
			expectedKeys:  []string{"home.welcome"},
			expectError:   false,
		},
		{
			name:          "env_var_with_semicolon_separated_files",
			envValue:      "test1.json;test2.json",
			args:          []string{"-k", "home.welcome"},
			expectedFiles: []string{"test1.json", "test2.json"},
			expectedKeys:  []string{"home.welcome"},
			expectError:   false,
		},
		{
			name:          "env_var_with_mixed_separators",
			envValue:      "test1.json test2.json:test3.json;test4.json",
			args:          []string{"-k", "home.welcome"},
			expectedFiles: []string{"test1.json", "test2.json", "test3.json", "test4.json"},
			expectedKeys:  []string{"home.welcome"},
			expectError:   false,
		},
		{
			name:          "command_line_overrides_env_var",
			envValue:      "env1.json env2.json",
			args:          []string{"-k", "home.welcome", "cmd1.json", "cmd2.json"},
			expectedFiles: []string{"cmd1.json", "cmd2.json"},
			expectedKeys:  []string{"home.welcome"},
			expectError:   false,
		},
		{
			name:         "empty_env_var_should_fail",
			envValue:     "",
			args:         []string{"-k", "home.welcome"},
			expectError:  true,
			errorContains: "at least one file must be specified",
		},
		{
			name:         "no_env_var_should_fail",
			args:         []string{"-k", "home.welcome"},
			expectError:  true,
			errorContains: "at least one file must be specified",
		},
		{
			name:          "env_var_with_specific_files",
			envValue:      "test-locales/zh-CN.json test-locales/en-US.json",
			args:          []string{"-k", "home.welcome"},
			expectedFiles: []string{"test-locales/zh-CN.json", "test-locales/en-US.json"},
			expectedKeys:  []string{"home.welcome"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("I18NEDT_FILES", tt.envValue)
				defer os.Unsetenv("I18NEDT_FILES")
			} else {
				// Ensure env var is not set
				os.Unsetenv("I18NEDT_FILES")
			}

			// Temporarily replace os.Args for testing
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Construct test args
			testArgs := append([]string{"i18nedt"}, tt.args...)
			os.Args = testArgs

			config, err := ParseFlags()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check keys
			if len(config.Keys) != len(tt.expectedKeys) {
				t.Errorf("Expected %d keys, got %d", len(tt.expectedKeys), len(config.Keys))
				return
			}
			for i, key := range tt.expectedKeys {
				if config.Keys[i] != key {
					t.Errorf("Expected key %d to be '%s', got '%s'", i, key, config.Keys[i])
				}
			}

			// Check files
			if len(tt.expectedFiles) > 0 {
				// Check exact file match when expectedFiles is specified
				if len(config.Files) != len(tt.expectedFiles) {
					t.Errorf("Expected %d files, got %d", len(tt.expectedFiles), len(config.Files))
					return
				}
				for i, file := range tt.expectedFiles {
					if config.Files[i] != file {
						t.Errorf("Expected file %d to be '%s', got '%s'", i, file, config.Files[i])
					}
				}
			} else {
				// Just check that we have some files when expectedFiles is empty
				if len(config.Files) == 0 {
					t.Errorf("Expected at least one file, got none")
				}
			}
		})
	}
}

func TestParseFlagsEnvVarEdgeCases(t *testing.T) {
	// Test edge cases for environment variable parsing
	tests := []struct {
		name          string
		envValue      string
		expectedFiles []string
	}{
		{
			name:          "multiple_spaces",
			envValue:      "  test1.json   test2.json  ",
			expectedFiles: []string{"test1.json", "test2.json"},
		},
		{
			name:          "multiple_colons",
			envValue:      "test1.json::test2.json",
			expectedFiles: []string{"test1.json", "test2.json"},
		},
		{
			name:          "multiple_semicolons",
			envValue:      "test1.json;;test2.json",
			expectedFiles: []string{"test1.json", "test2.json"},
		},
		{
			name:          "mixed_separators_with_spaces",
			envValue:      "test1.json : test2.json ; test3.json",
			expectedFiles: []string{"test1.json", "test2.json", "test3.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("I18NEDT_FILES", tt.envValue)
			defer os.Unsetenv("I18NEDT_FILES")

			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			testArgs := []string{"i18nedt", "-k", "test.key"}
			os.Args = testArgs

			config, err := ParseFlags()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(config.Files) != len(tt.expectedFiles) {
				t.Errorf("Expected %d files, got %d", len(tt.expectedFiles), len(config.Files))
				return
			}
			for i, file := range tt.expectedFiles {
				if config.Files[i] != file {
					t.Errorf("Expected file %d to be '%s', got '%s'", i, file, config.Files[i])
				}
			}
		})
	}
}