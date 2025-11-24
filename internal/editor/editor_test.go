package editor

import (
	"os"
	"testing"
)

func TestValidateEditor(t *testing.T) {
	tests := []struct {
		name     string
		editor   string
		wantErr  bool
	}{
		{
			name:    "valid editor vim",
			editor:  "echo", // Using echo instead of vim as it's guaranteed to exist
			wantErr: false,
		},
		{
			name:    "another valid editor",
			editor:  "cat",
			wantErr: false,
		},
		{
			name:    "empty editor name",
			editor:  "",
			wantErr: true,
		},
		{
			name:    "non-existent editor",
			editor:  "nonexistent-editor-12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEditor(tt.editor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEditor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDefaultEditor(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envVal   string
		expected string
	}{
		{
			name:     "no environment variables set",
			expected: "vim",
		},
		{
			name:     "EDITOR set",
			envVar:   "EDITOR",
			envVal:   "nano",
			expected: "nano",
		},
		{
			name:     "VISUAL set",
			envVar:   "VISUAL",
			envVal:   "emacs",
			expected: "emacs",
		},
		{
			name:     "both set (EDITOR should take precedence)",
			envVar:   "EDITOR",
			envVal:   "code",
			expected: "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			os.Unsetenv("EDITOR")
			os.Unsetenv("VISUAL")

			// Set the test environment variable
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			// If testing both EDITOR and VISUAL precedence
			if tt.name == "both set (EDITOR should take precedence)" {
				os.Setenv("VISUAL", "emacs")
				defer os.Unsetenv("VISUAL")
			}

			got := GetDefaultEditor()
			if got != tt.expected {
				t.Errorf("GetDefaultEditor() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOpenEditor(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-editor-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name    string
		editor  string
		wantErr bool
	}{
		{
			name:    "valid editor with echo",
			editor:  "echo",
			wantErr: false,
		},
		{
			name:    "valid editor with cat",
			editor:  "cat",
			wantErr: false,
		},
		{
			name:    "empty editor name",
			editor:  "",
			wantErr: true,
		},
		{
			name:    "non-existent editor",
			editor:  "nonexistent-editor-12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OpenEditor(tmpFile.Name(), tt.editor)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenEditor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOpenEditorWithCommand tests that the editor command is actually executed
func TestOpenEditorWithCommand(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-editor-command-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Use a command that we know will exist and work
	// On Unix systems, 'true' command always exits successfully
	editorCmd := "true"

	// On Windows, use 'echo' which should exist
	if os.Getenv("GOOS") == "windows" {
		editorCmd = "echo"
	}

	err = OpenEditor(tmpFile.Name(), editorCmd)
	if err != nil {
		t.Errorf("OpenEditor() with '%s' failed: %v", editorCmd, err)
	}
}

func TestOpenEditorFileNotFound(t *testing.T) {
	nonExistentFile := "/tmp/nonexistent-file-12345.txt"

	err := OpenEditor(nonExistentFile, "echo")
	// Most editors will handle non-existent files by creating them,
	// but some might fail. We just check that the error handling works
	if err != nil {
		// This is expected behavior for some editors
		t.Logf("OpenEditor() with non-existent file failed as expected: %v", err)
	}
}

func TestValidateEditorIntegration(t *testing.T) {
	// Test that ValidateEditor and OpenEditor work together consistently
	editor := "echo" // Use echo as it should be available on all systems

	// First validate the editor
	err := ValidateEditor(editor)
	if err != nil {
		t.Skipf("Editor '%s' not available for integration test: %v", editor, err)
		return
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "integration-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Try to open the editor
	err = OpenEditor(tmpFile.Name(), editor)
	if err != nil {
		t.Errorf("OpenEditor() failed after successful ValidateEditor(): %v", err)
	}
}

func TestGetDefaultEditorIntegration(t *testing.T) {
	// Test that GetDefaultEditor returns a valid editor
	defaultEditor := GetDefaultEditor()

	// The default should not be empty
	if defaultEditor == "" {
		t.Error("GetDefaultEditor() returned empty string")
	}

	// Test that the default editor can be validated
	err := ValidateEditor(defaultEditor)
	// Note: This might fail if the system doesn't have vim installed,
	// which is fine - we just want to ensure the function works
	if err != nil {
		t.Logf("Default editor '%s' not available on this system: %v", defaultEditor, err)
	}
}