package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chen/i18nedt/pkg/types"
)

func TestValidateConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  *types.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &types.Config{
				Files: []string{filepath.Join(tmpDir, "test.json")},
				Keys:  []string{"home.welcome"},
			},
			wantErr: false,
		},
		{
			name: "empty files",
			config: &types.Config{
				Files: []string{},
				Keys:  []string{"home.welcome"},
			},
			wantErr: true,
		},
		{
			name: "empty keys",
			config: &types.Config{
				Files: []string{filepath.Join(tmpDir, "test.json")},
				Keys:  []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid file extension",
			config: &types.Config{
				Files: []string{filepath.Join(tmpDir, "test.txt")},
				Keys:  []string{"home.welcome"},
			},
			wantErr: true,
		},
		{
			name: "directory doesn't exist",
			config: &types.Config{
				Files: []string{"/nonexistent/directory/test.json"},
				Keys:  []string{"home.welcome"},
			},
			wantErr: true,
		},
		{
			name: "invalid key format",
			config: &types.Config{
				Files: []string{filepath.Join(tmpDir, "test.json")},
				Keys:  []string{"1invalid.key"}, // starts with number
			},
			wantErr: true,
		},
		{
			name: "empty key",
			config: &types.Config{
				Files: []string{filepath.Join(tmpDir, "test.json")},
				Keys:  []string{""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid json file",
			path:    filepath.Join(tmpDir, "test.json"),
			wantErr: false,
		},
		{
			name:    "valid nested json file",
			path:    filepath.Join(tmpDir, "subdir", "test.json"),
			wantErr: false,
		},
		{
			name:    "invalid extension",
			path:    filepath.Join(tmpDir, "test.txt"),
			wantErr: true,
		},
		{
			name:    "uppercase json extension",
			path:    filepath.Join(tmpDir, "test.JSON"),
			wantErr: false,
		},
		{
			name:    "directory doesn't exist",
			path:    "/nonexistent/directory/test.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subdirectory if needed
			if tt.name == "valid nested json file" {
				err := os.MkdirAll(filepath.Dir(tt.path), 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			}

			err := validateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid simple key",
			key:     "welcome",
			wantErr: false,
		},
		{
			name:    "valid nested key",
			key:     "home.welcome",
			wantErr: false,
		},
		{
			name:    "valid key with underscore",
			key:     "home_welcome",
			wantErr: false,
		},
		{
			name:    "valid key with hyphen",
			key:     "home-welcome",
			wantErr: false,
		},
		{
			name:    "valid key starting with underscore",
			key:     "_private.key",
			wantErr: false,
		},
		{
			name:    "invalid key starting with number",
			key:     "1invalid.key",
			wantErr: true,
		},
		{
			name:    "invalid empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "invalid key with space",
			key:     "home welcome",
			wantErr: true,
		},
		{
			name:    "invalid key with special chars",
			key:     "home@welcome",
			wantErr: true,
		},
		{
			name:    "valid deeply nested key",
			key:     "nav.menu.home.title",
			wantErr: false,
		},
		{
			name:    "valid key with numbers",
			key:     "nav1.menu2.home3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}