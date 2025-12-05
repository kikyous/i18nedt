package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kikyous/i18nedt/pkg/types"
)

// Helper function to convert map to JSON string
func mapToJSON(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func TestLoadFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() (string, error)
		wantErr  bool
		wantData map[string]interface{}
	}{
		{
			name: "load existing file",
			setup: func() (string, error) {
				filePath := filepath.Join(tmpDir, "test.json")
				data := map[string]interface{}{
					"welcome": "Welcome",
					"home": map[string]interface{}{
						"title": "Home",
					},
				}
				jsonData, _ := json.Marshal(data)
				err := os.WriteFile(filePath, jsonData, 0644)
				return filePath, err
			},
			wantErr: false,
			wantData: map[string]interface{}{
				"welcome": "Welcome",
				"home": map[string]interface{}{
					"title": "Home",
				},
			},
		},
		{
			name: "load non-existent file",
			setup: func() (string, error) {
				filePath := filepath.Join(tmpDir, "nonexistent.json")
				return filePath, nil
			},
			wantErr: false,
			wantData: map[string]interface{}{},
		},
		{
			name: "load empty file",
			setup: func() (string, error) {
				filePath := filepath.Join(tmpDir, "empty.json")
				err := os.WriteFile(filePath, []byte{}, 0644)
				return filePath, err
			},
			wantErr: false,
			wantData: map[string]interface{}{},
		},
		{
			name: "load invalid JSON",
			setup: func() (string, error) {
				filePath := filepath.Join(tmpDir, "invalid.json")
				err := os.WriteFile(filePath, []byte("{invalid json}"), 0644)
				return filePath, err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := tt.setup()
			if err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			// Use the absolute file path as the pattern to ensure a match
			// Since these tests focus on JSON loading, not metadata extraction
			pattern := filePath
			file, err := LoadFile(filePath, pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if file.Path != filePath {
				t.Errorf("LoadFile().Path = %v, want %v", file.Path, filePath)
			}

			// Parse file.Data as JSON and compare with expected data
			var fileData map[string]interface{}
			if err := json.Unmarshal([]byte(file.Data), &fileData); err != nil {
				t.Errorf("LoadFile().Data is not valid JSON: %v", err)
				return
			}
			if !equalMaps(fileData, tt.wantData) {
				t.Errorf("LoadFile().Data = %v, want %v", fileData, tt.wantData)
			}
		})
	}
}

func TestSaveFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		file    *types.I18nFile
		wantErr bool
	}{
		{
			name: "save simple data",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "simple.json"),
				Data: `{"welcome": "Welcome"}`,
			},
			wantErr: false,
		},
		{
			name: "save nested data",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "nested.json"),
				Data: mapToJSON(map[string]interface{}{
					"home": map[string]interface{}{
						"title": "Home",
						"desc":  "Description",
					},
				}),
			},
			wantErr: false,
		},
		{
			name: "save to non-existent directory",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "subdir", "new.json"),
				Data: mapToJSON(map[string]interface{}{
					"test": "value",
				}),
			},
			wantErr: false,
		},
		{
			name: "save empty data",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "empty.json"),
				Data: mapToJSON(map[string]interface{}{}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Verify file was created and has correct content
			if _, err := os.Stat(tt.file.Path); os.IsNotExist(err) {
				t.Errorf("SaveFile() file was not created")
				return
			}

			// Load and verify content
			data, err := os.ReadFile(tt.file.Path)
			if err != nil {
				t.Errorf("SaveFile() failed to read saved file: %v", err)
				return
			}

			var loadedData map[string]interface{}
			if err := json.Unmarshal(data, &loadedData); err != nil {
				t.Errorf("SaveFile() saved invalid JSON: %v", err)
				return
			}

			// Parse the original file data and compare with loaded data
			var originalData map[string]interface{}
			if err := json.Unmarshal([]byte(tt.file.Data), &originalData); err != nil {
				t.Errorf("Failed to parse original data: %v", err)
				return
			}
			if !equalMaps(loadedData, originalData) {
				t.Errorf("SaveFile() saved data = %v, want %v", loadedData, originalData)
			}
		})
	}
}

func TestSaveAllFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	files := []*types.I18nFile{
		{
			Path: filepath.Join(tmpDir, "zh-CN.json"),
			Data: mapToJSON(map[string]interface{}{
				"welcome": "欢迎",
			}),
			Dirty: true,
		},
		{
			Path: filepath.Join(tmpDir, "en-US.json"),
			Data: mapToJSON(map[string]interface{}{
				"welcome": "Welcome",
			}),
			Dirty: true,
		},
	}

	count, err := SaveAllFiles(files)
	if err != nil {
		t.Errorf("SaveAllFiles() error = %v", err)
		return
	}

	if count != 2 {
		t.Errorf("SaveAllFiles() count = %d, want 2", count)
	}

	// Verify all files were created
	for _, file := range files {
		if _, err := os.Stat(file.Path); os.IsNotExist(err) {
			t.Errorf("SaveAllFiles() file %s was not created", file.Path)
		}
	}
}

func TestLoadAllFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"zh-CN.json", "en-US.json"}
	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		data := map[string]interface{}{
			"welcome": "Welcome",
		}
		jsonData, _ := json.Marshal(data)
		err := os.WriteFile(filePath, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test loading all files
	sources := []FileSource{
		{Path: filepath.Join(tmpDir, "zh-CN.json"), Pattern: filepath.Join(tmpDir, "{{language}}.json")},
		{Path: filepath.Join(tmpDir, "en-US.json"), Pattern: filepath.Join(tmpDir, "{{language}}.json")},
		{Path: filepath.Join(tmpDir, "nonexistent.json"), Pattern: filepath.Join(tmpDir, "nonexistent.json")},
	}

	files, err := LoadAllFiles(sources)
	if err != nil {
		t.Errorf("LoadAllFiles() error = %v", err)
		return
	}

	if len(files) != len(sources) {
		t.Errorf("LoadAllFiles() length = %v, want %v", len(files), len(sources))
	}

	for i, file := range files {
		if file.Path != sources[i].Path {
			t.Errorf("LoadAllFiles()[%d].Path = %v, want %v", i, file.Path, sources[i].Path)
		}

		if i < len(testFiles) {
			// Verify loaded data for existing files
			if welcome, err := GetValue(file.Data, "welcome"); err != nil || welcome != "Welcome" {
				t.Errorf("LoadAllFiles()[%d] incorrect data, got welcome = %v", i, welcome)
			}
		} else {
			// Verify empty data for non-existent files (should be "{}")
			if file.Data != "{}" {
				t.Errorf("LoadAllFiles()[%d] expected empty data for non-existent file, got %s", i, file.Data)
			}
		}
	}
}

func TestBackupFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "backup existing file",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "original.json")
				data := map[string]interface{}{
					"welcome": "Welcome",
				}
				jsonData, _ := json.Marshal(data)
				os.WriteFile(filePath, jsonData, 0644)
				return filePath
			},
			wantErr: false,
		},
		{
			name: "backup non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			wantErr: false, // Should not error on non-existent file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			err := BackupFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("BackupFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// For existing files, verify backup was created
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				backupPath := filePath + ".backup"
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Errorf("BackupFile() backup file was not created")
				} else {
					// Verify backup content matches original
					originalData, _ := os.ReadFile(filePath)
					backupData, _ := os.ReadFile(backupPath)
					if string(originalData) != string(backupData) {
						t.Errorf("BackupFile() backup content differs from original")
					}
				}
			}
		})
	}
}

func TestGetDirectory(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "simple file",
			filePath: "test.json",
			want:     ".",
		},
		{
			name:     "file in subdirectory",
			filePath: "locales/test.json",
			want:     "locales",
		},
		{
			name:     "file with absolute path",
			filePath: "/home/user/project/test.json",
			want:     "/home/user/project",
		},
		{
			name:     "file in deeply nested directory",
			filePath: "src/i18n/locales/zh-CN.json",
			want:     "src/i18n/locales",
		},
		{
			name:     "file in root directory",
			filePath: "/test.json",
			want:     "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDirectory(tt.filePath)
			if got != tt.want {
				t.Errorf("GetDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare maps
func equalMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, exists := b[k]; !exists || !equalInterface(v, bv) {
			return false
		}
	}
	return true
}

func equalInterface(a, b interface{}) bool {
	switch a := a.(type) {
	case map[string]interface{}:
		b, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return equalMaps(a, b)
	case []interface{}:
		b, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if !equalInterface(a[i], b[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}