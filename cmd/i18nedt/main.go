package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chen/i18nedt/internal/cli"
	"github.com/chen/i18nedt/internal/editor"
	"github.com/chen/i18nedt/internal/i18n"
)

func main() {
	// Parse command line arguments
	config, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cli.ValidateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate editor
	if err := editor.ValidateEditor(config.Editor); err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %v\n", err)
		os.Exit(1)
	}

	// Load all i18n files
	files, err := i18n.LoadAllFiles(config.Files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading files: %v\n", err)
		os.Exit(1)
	}

	// Expand keys to include all child keys for non-leaf keys
	// Use the first file's data as reference for key expansion
	var allData map[string]interface{}
	if len(files) > 0 {
		allData = files[0].Data
	} else {
		allData = make(map[string]interface{})
	}

	expandedKeys := i18n.ExpandKeys(allData, config.Keys)

	// Report if keys were expanded
	if len(expandedKeys) > len(config.Keys) {
		fmt.Printf("Expanded %d key(s) to %d key(s)\n", len(config.Keys), len(expandedKeys))
	}

	// Create temporary file for editing
	tempFile, err := editor.CreateTempFile(files, expandedKeys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
		os.Exit(1)
	}

	// Ensure cleanup on exit
	defer func() {
		if err := editor.CleanupTempFile(tempFile); err != nil {
			log.Printf("Warning: failed to cleanup temporary file: %v", err)
		}
	}()

	// Write initial content to temporary file
	if err := editor.WriteTempFile(tempFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing temporary file: %v\n", err)
		os.Exit(1)
	}

	// Open editor
	if err := editor.OpenEditor(tempFile.Path, config.Editor); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
		os.Exit(1)
	}

	// Parse edited content
	if err := editor.ReadTempFile(tempFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing edited file: %v\n", err)
		os.Exit(1)
	}

	// Apply changes to the actual files
	if err := editor.ApplyChanges(files, tempFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error applying changes: %v\n", err)
		os.Exit(1)
	}

	// Save all files
	if err := i18n.SaveAllFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving files: %v\n", err)
		os.Exit(1)
	}

	// Report summary
	if len(tempFile.Deletes) > 0 {
		fmt.Printf("Deleted %d keys\n", len(tempFile.Deletes))
	}

	updatedCount := 0
	for _, localeValues := range tempFile.Content {
		for _, value := range localeValues {
			if value != "" {
				updatedCount++
				break
			}
		}
	}

	if updatedCount > 0 {
		fmt.Printf("Updated %d keys\n", updatedCount)
	}

	fmt.Printf("Successfully updated %d files\n", len(files))
}