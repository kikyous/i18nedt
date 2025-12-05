package main

import (
	"fmt"
	"log"
	"os"
	"github.com/alexflint/go-arg"
	"github.com/kikyous/i18nedt/internal/editor"
	"github.com/kikyous/i18nedt/internal/flatten"
	"github.com/kikyous/i18nedt/internal/i18n"
	"github.com/kikyous/i18nedt/pkg/types"
)

// Version information
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// args struct for go-arg
var args struct {
	Keys         []string `arg:"-k,--key,separate" help:"Key to edit (can be specified multiple times)"`
	PrintOnly    bool     `arg:"-p,--print" help:"Print temporary file content without launching editor"`
	NoTips       bool     `arg:"-a,--no-tips,env" help:"Exclude AI tips from temporary file content"`
	PathAsLocale bool     `arg:"-P,--path-as-locale" help:"Use file path as locale identifier"`
	Flatten      bool     `arg:"-f,--flatten" help:"Flatten JSON files to key=value format"`
	Version      bool     `arg:"-v,--version" help:"Show version information"`
	Files        []string `arg:"positional" help:"Target file paths [env: I18NEDT_FILES]"`
}

func main() {
	p, err := arg.NewParser(arg.Config{
		EnvPrefix: "I18NEDT_",
	}, &args)
	// Parse command line arguments
	p.MustParse(os.Args[1:])

	// Handle version flag
	if args.Version {
		fmt.Printf("i18nedt version %s\n", Version)
		if Commit != "unknown" {
			fmt.Printf("commit: %s\n", Commit)
		}
		if Date != "unknown" {
			fmt.Printf("built: %s\n", Date)
		}
		os.Exit(0)
	}

	// Handle file expansion (globbing) via discovery module
	sources, flatFiles, err := i18n.DiscoverFiles(args.Files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Construct Config
	config := &types.Config{
		Files:        flatFiles, // We keep this for Flatten logic which iterates simple paths
		Keys:         args.Keys,
		Editor:       os.Getenv("EDITOR"),
		PrintOnly:    args.PrintOnly,
		NoTips:       args.NoTips,
		PathAsLocale: args.PathAsLocale,
		Flatten:      args.Flatten,
	}
	if config.Editor == "" {
		config.Editor = "vim"
	}

	// Handle flatten mode
	if config.Flatten {
		// Flatten each file
		for _, file := range config.Files {
			if err := flatten.FlattenJSON(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error flattening file %s: %v\n", file, err)
				os.Exit(1)
			}
		}
		return
	}

	// Validate editor
	if err := editor.ValidateEditor(config.Editor); err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %v\n", err)
		os.Exit(1)
	}

	// Load all i18n files
	files, err := i18n.LoadAllFiles(sources, config.PathAsLocale)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading files: %v\n", err)
		os.Exit(1)
	}

	// Use keys directly without expansion - user explicitly specifies what to edit
	tempFile, err := editor.CreateTempFile(files, config.Keys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
		os.Exit(1)
	}

	// If print only mode, generate content and print to stdout
	if config.PrintOnly {
		content, err := editor.GenerateTempFileContentWithOptions(tempFile, config.NoTips)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating temporary file content: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(content))
		return
	}

	// Ensure cleanup on exit
	defer func() {
		if err := editor.CleanupTempFile(tempFile); err != nil {
			log.Printf("Warning: failed to cleanup temporary file: %v", err)
		}
	}()

	// Write initial content to temporary file
	if err := editor.WriteTempFileWithOptions(tempFile, config.NoTips); err != nil {
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
		for _, key := range tempFile.Deletes {
			fmt.Printf("  %s\n", key)
		}
	}

	updatedCount := 0
	for _, localeValues := range tempFile.Content {
		for _, value := range localeValues {
			if value.Value != "" {
				updatedCount++
			}
		}
	}

	if updatedCount > 0 {
		fmt.Printf("Updated %d keys\n", updatedCount)
	}

	fmt.Printf("Successfully updated %d files\n", len(files))
}
