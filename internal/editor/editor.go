package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenEditor opens the system editor with the specified file
func OpenEditor(filePath, editorName string) error {
	if editorName == "" {
		return fmt.Errorf("editor name cannot be empty")
	}

	// Check if editor exists
	if _, err := exec.LookPath(editorName); err != nil {
		return fmt.Errorf("editor '%s' not found: %w", editorName, err)
	}

	// Create command to run editor
	cmd := exec.Command(editorName, filePath)

	// Set up standard I/O to connect with terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the editor and wait for it to exit
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	return nil
}

// ValidateEditor checks if the specified editor is available
func ValidateEditor(editorName string) error {
	if editorName == "" {
		return fmt.Errorf("editor name cannot be empty")
	}

	if _, err := exec.LookPath(editorName); err != nil {
		return fmt.Errorf("editor '%s' not found in PATH: %w", editorName, err)
	}

	return nil
}

// GetDefaultEditor returns the default editor based on environment
func GetDefaultEditor() string {
	// Check EDITOR environment variable first
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Check VISUAL environment variable
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Default to vim
	return "vim"
}