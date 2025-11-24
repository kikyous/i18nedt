package i18n

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/chen/i18nedt/pkg/types"
)

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
				err := ioutil.WriteFile(filePath, jsonData, 0644)
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
				err := ioutil.WriteFile(filePath, []byte{}, 0644)
				return filePath, err
			},
			wantErr: false,
			wantData: map[string]interface{}{},
		},
		{
			name: "load invalid JSON",
			setup: func() (string, error) {
				filePath := filepath.Join(tmpDir, "invalid.json")
				err := ioutil.WriteFile(filePath, []byte("{invalid json}"), 0644)
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

			file, err := LoadFile(filePath)
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

			if !equalMaps(file.Data, tt.wantData) {
				t.Errorf("LoadFile().Data = %v, want %v", file.Data, tt.wantData)
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
				Data: map[string]interface{}{
					"welcome": "Welcome",
				},
			},
			wantErr: false,
		},
		{
			name: "save nested data",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "nested.json"),
				Data: map[string]interface{}{
					"home": map[string]interface{}{
						"title": "Home",
						"desc":  "Description",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "save to non-existent directory",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "subdir", "new.json"),
				Data: map[string]interface{}{
					"test": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "save empty data",
			file: &types.I18nFile{
				Path: filepath.Join(tmpDir, "empty.json"),
				Data: map[string]interface{}{},
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
			data, err := ioutil.ReadFile(tt.file.Path)
			if err != nil {
				t.Errorf("SaveFile() failed to read saved file: %v", err)
				return
			}

			var loadedData map[string]interface{}
			if err := json.Unmarshal(data, &loadedData); err != nil {
				t.Errorf("SaveFile() saved invalid JSON: %v", err)
				return
			}

			// Clean empty maps before comparison
			cleanedData := CleanEmptyMaps(tt.file.Data)
			if !equalMaps(loadedData, cleanedData) {
				t.Errorf("SaveFile() saved data = %v, want %v", loadedData, cleanedData)
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
			Data: map[string]interface{}{
				"welcome": "欢迎",
			},
		},
		{
			Path: filepath.Join(tmpDir, "en-US.json"),
			Data: map[string]interface{}{
				"welcome": "Welcome",
			},
		},
	}

	err := SaveAllFiles(files)
	if err != nil {
		t.Errorf("SaveAllFiles() error = %v", err)
		return
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
		err := ioutil.WriteFile(filePath, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test loading all files
	filePaths := []string{
		filepath.Join(tmpDir, "zh-CN.json"),
		filepath.Join(tmpDir, "en-US.json"),
		filepath.Join(tmpDir, "nonexistent.json"), // This should not cause error
	}

	files, err := LoadAllFiles(filePaths)
	if err != nil {
		t.Errorf("LoadAllFiles() error = %v", err)
		return
	}

	if len(files) != len(filePaths) {
		t.Errorf("LoadAllFiles() length = %v, want %v", len(files), len(filePaths))
	}

	for i, file := range files {
		if file.Path != filePaths[i] {
			t.Errorf("LoadAllFiles()[%d].Path = %v, want %v", i, file.Path, filePaths[i])
		}

		if i < len(testFiles) {
			// Verify loaded data for existing files
			if welcome, exists := GetValue(file.Data, "welcome"); !exists || welcome != "Welcome" {
				t.Errorf("LoadAllFiles()[%d] incorrect data, got welcome = %v", i, welcome)
			}
		} else {
			// Verify empty data for non-existent files
			if len(file.Data) != 0 {
				t.Errorf("LoadAllFiles()[%d] expected empty data for non-existent file", i)
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
				ioutil.WriteFile(filePath, jsonData, 0644)
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
					originalData, _ := ioutil.ReadFile(filePath)
					backupData, _ := ioutil.ReadFile(backupPath)
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