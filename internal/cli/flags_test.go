package cli

import (
	"os"
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		wantKeys []string
		wantLen  int
	}{
		{
			name:    "no keys provided",
			args:    []string{"i18nedt", "file1.json"},
			wantErr: true,
		},
		{
			name:    "no files provided",
			args:    []string{"i18nedt", "-k", "key1"},
			wantErr: true,
		},
		{
			name:     "keys before files (original format)",
			args:     []string{"i18nedt", "-k", "home.welcome", "file1.json"},
			wantErr:  false,
			wantKeys: []string{"home.welcome"},
			wantLen:  1,
		},
		{
			name:     "files before keys (new format)",
			args:     []string{"i18nedt", "file1.json", "-k", "home.welcome"},
			wantErr:  false,
			wantKeys: []string{"home.welcome"},
			wantLen:  1,
		},
		{
			name:     "mixed files and keys",
			args:     []string{"i18nedt", "file1.json", "-k", "home.welcome", "file2.json", "-k", "nav.home"},
			wantErr:  false,
			wantKeys: []string{"home.welcome", "nav.home"},
			wantLen:  2,
		},
		{
			name:     "multiple keys and files - keys first",
			args:     []string{"i18nedt", "-k", "home.welcome", "-k", "nav.home", "file1.json", "file2.json"},
			wantErr:  false,
			wantKeys: []string{"home.welcome", "nav.home"},
			wantLen:  2,
		},
		{
			name:     "multiple keys and files - files first",
			args:     []string{"i18nedt", "file1.json", "file2.json", "-k", "home.welcome", "-k", "nav.home"},
			wantErr:  false,
			wantKeys: []string{"home.welcome", "nav.home"},
			wantLen:  2,
		},
		{
			name:     "--key long form with equals",
			args:     []string{"i18nedt", "--key=home.welcome", "file1.json"},
			wantErr:  false,
			wantKeys: []string{"home.welcome"},
			wantLen:  1,
		},
		{
			name:     "--key long form with space",
			args:     []string{"i18nedt", "--key", "home.welcome", "file1.json"},
			wantErr:  false,
			wantKeys: []string{"home.welcome"},
			wantLen:  1,
		},
		{
			name:     "-k without space",
			args:     []string{"i18nedt", "-khome.welcome", "file1.json"},
			wantErr:  false,
			wantKeys: []string{"home.welcome"},
			wantLen:  1,
		},
		{
			name:    "invalid flag",
			args:    []string{"i18nedt", "-x", "value", "file1.json"},
			wantErr: true,
		},
		{
			name:    "-k without value",
			args:    []string{"i18nedt", "-k", "file1.json"},
			wantErr: true,
		},
		{
			name:    "--key without value",
			args:    []string{"i18nedt", "--key", "file1.json"},
			wantErr: true,
		},
		{
			name:    "invalid --key format",
			args:    []string{"i18nedt", "--keyvalue", "file1.json"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			config, err := ParseFlags()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFlags() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFlags() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(config.Keys, tt.wantKeys) {
				t.Errorf("ParseFlags().Keys = %v, want %v", config.Keys, tt.wantKeys)
			}

			if len(config.Files) != tt.wantLen {
				t.Errorf("ParseFlags().Files length = %v, want %v", len(config.Files), tt.wantLen)
			}
		})
	}
}

func TestExpandFilePaths(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		wantErr  bool
		expected []string
	}{
		{
			name:     "regular files",
			paths:    []string{"file1.json", "file2.json"},
			wantErr:  false,
			expected: []string{"file1.json", "file2.json"},
		},
		{
			name:    "glob pattern with no matches",
			paths:   []string{"*.nonexistent"},
			wantErr: true,
		},
		{
			name:    "mixed paths",
			paths:   []string{"file1.json", "*.json"},
			wantErr: true, // Glob pattern with no matches should error
		},
		{
			name:     "brace expansion pattern",
			paths:    []string{"{zh-CN,en-US}.json"},
			wantErr:  true, // This pattern doesn't work with filepath.Glob
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandFilePaths(tt.paths)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expandFilePaths() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expandFilePaths() unexpected error: %v", err)
				return
			}

			if len(tt.expected) > 0 && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expandFilePaths() = %v, want %v", result, tt.expected)
			}
		})
	}
}