package types

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		envVal  string
		want    string
	}{
		{
			name:   "default editor when EDITOR is not set",
			want:   "vim",
		},
		{
			name:   "use EDITOR environment variable",
			envVar: "EDITOR",
			envVal: "nano",
			want:   "nano",
		},
		{
			name:   "use EDITOR environment variable with code",
			envVar: "EDITOR",
			envVal: "code",
			want:   "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment before test
			os.Unsetenv("EDITOR")
			os.Unsetenv("VISUAL")

			// Set up environment
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			config := NewConfig()
			if config.Editor != tt.want {
				t.Errorf("NewConfig().Editor = %v, want %v", config.Editor, tt.want)
			}
		})
	}
}

func TestI18nFile(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "simple path",
			path: "test.json",
		},
		{
			name: "complex path",
			path: "/path/to/locales/zh-CN.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &I18nFile{
				Path: tt.path,
				Data: "",
			}

			if file.Path != tt.path {
				t.Errorf("I18nFile.Path = %v, want %v", file.Path, tt.path)
			}

			// Empty string is fine for I18nFile.Data - it represents empty JSON
		})
	}
}

func TestTempFile(t *testing.T) {
	keys := []string{"home.welcome", "nav.home"}
	locales := []string{"zh-CN", "en-US"}

	temp := &TempFile{
		Path:    "/tmp/test.txt",
		Keys:    keys,
		Locales: locales,
		Content: make(map[string]map[string]*Value),
		Deletes: []string{},
	}

	// Test Keys
	if len(temp.Keys) != 2 {
		t.Errorf("TempFile.Keys length = %v, want %v", len(temp.Keys), 2)
	}

	// Test Locales
	if len(temp.Locales) != 2 {
		t.Errorf("TempFile.Locales length = %v, want %v", len(temp.Locales), 2)
	}

	// Test Content initialization
	if temp.Content == nil {
		t.Error("TempFile.Content should not be nil")
	}

	// Test Deletes
	if len(temp.Deletes) != 0 {
		t.Errorf("TempFile.Deletes length = %v, want %v", len(temp.Deletes), 0)
	}
}

func TestKeyOperation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		delete  bool
	}{
		{
			name:   "regular key operation",
			key:    "home.welcome",
			value:  "Welcome",
			delete: false,
		},
		{
			name:   "delete key operation",
			key:    "old.key",
			value:  "",
			delete: true,
		},
		{
			name:   "empty value key operation",
			key:    "empty.key",
			value:  "",
			delete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := KeyOperation{
				Key:    tt.key,
				Value:  tt.value,
				Delete: tt.delete,
			}

			if op.Key != tt.key {
				t.Errorf("KeyOperation.Key = %v, want %v", op.Key, tt.key)
			}

			if op.Value != tt.value {
				t.Errorf("KeyOperation.Value = %v, want %v", op.Value, tt.value)
			}

			if op.Delete != tt.delete {
				t.Errorf("KeyOperation.Delete = %v, want %v", op.Delete, tt.delete)
			}
		})
	}
}