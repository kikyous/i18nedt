package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// parseEditorCommand splits an editor command string into the executable name and arguments
func parseEditorCommand(editorStr string) (string, []string) {
	parts := strings.Fields(editorStr)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// OpenEditor opens the system editor with the specified file
func OpenEditor(filePath, editorName string) error {
	if editorName == "" {
		return fmt.Errorf("editor name cannot be empty")
	}

	// Parse editor command to separate executable from arguments
	executable, args := parseEditorCommand(editorName)

	// Check if editor exists
	if _, err := exec.LookPath(executable); err != nil {
		return fmt.Errorf("editor '%s' not found: %w", executable, err)
	}

	// Create command to run editor with arguments + file path
	allArgs := append(args, filePath)
	cmd := exec.Command(executable, allArgs...)

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

	// Parse editor command to separate executable from arguments
	executable, _ := parseEditorCommand(editorName)

	if _, err := exec.LookPath(executable); err != nil {
		return fmt.Errorf("editor '%s' not found in PATH: %w", executable, err)
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