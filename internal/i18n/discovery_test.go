package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create test structure
	// locales/en/common.json
	// locales/en/auth.json
	// locales/zh/common.json
	dirs := []string{
		filepath.Join(tmpDir, "locales", "en"),
		filepath.Join(tmpDir, "locales", "zh"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	files := []string{
		filepath.Join(tmpDir, "locales", "en", "common.json"),
		filepath.Join(tmpDir, "locales", "en", "auth.json"),
		filepath.Join(tmpDir, "locales", "zh", "common.json"),
	}
	for _, f := range files {
		if err := os.WriteFile(f, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tests := []struct {
		name           string
		patterns       []string
		envVar         string
		wantSourcesLen int
		wantFilesLen   int
		wantErr        bool
	}{
		{
			name:           "glob pattern",
			patterns:       []string{filepath.Join(tmpDir, "locales", "*", "*.json")},
			wantSourcesLen: 3,
			wantFilesLen:   3,
			wantErr:        false,
		},
		{
			name:           "custom pattern placeholder",
			patterns:       []string{filepath.Join(tmpDir, "locales", "{{language}}", "{{ns}}.json")},
			wantSourcesLen: 3,
			wantFilesLen:   3,
			wantErr:        false,
		},
		{
			name:           "no files found (but pattern valid)",
			patterns:       []string{filepath.Join(tmpDir, "locales", "*.txt")},
			wantSourcesLen: 1, // Adds the pattern itself if no match
			wantFilesLen:   1,
			wantErr:        false,
		},
		{
			name:           "empty pattern list with env var",
			patterns:       []string{},
			envVar:         filepath.Join(tmpDir, "locales", "*", "common.json"),
			wantSourcesLen: 2,
			wantFilesLen:   2,
			wantErr:        false,
		},
		{
			name:           "empty pattern list no env var",
			patterns:       []string{},
			envVar:         "",
			wantSourcesLen: 0,
			wantFilesLen:   0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("I18NEDT_FILES", tt.envVar)
			} else {
				os.Unsetenv("I18NEDT_FILES")
			}

			sources, flatFiles, err := DiscoverFiles(tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(sources) != tt.wantSourcesLen {
				t.Errorf("DiscoverFiles() sources len = %v, want %v", len(sources), tt.wantSourcesLen)
			}

			if len(flatFiles) != tt.wantFilesLen {
				t.Errorf("DiscoverFiles() flatFiles len = %v, want %v", len(flatFiles), tt.wantFilesLen)
			}
		})
	}
}
