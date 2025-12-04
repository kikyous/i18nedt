package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/bmatcuk/doublestar/v4"
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
	Files        []string `arg:"positional,env" help:"Target file paths"`
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

	// Handle file expansion (globbing)
	var finalFiles []string

	for _, pattern := range args.Files {
		// Use doublestar for file globbing (supports {a,b} and **)
		matches, err := doublestar.FilepathGlob(pattern)
		if err == nil && len(matches) > 0 {
			finalFiles = append(finalFiles, matches...)
		} else {
			// If no match or error, keep original (might be a new file or specific path)
			// But if it contains glob characters and failed to match, maybe we shouldn't add it if we want strict behavior?
			// The original code: if err == nil && len(matches) > 0 { append matches } else { append pattern }
			// So if I pass "*.json" and it matches nothing, it appends "*.json".
			// Then LoadAllFiles will try to open "*.json" and fail.
			// But here "at least one file must be specified" check is after this loop.
			// So finalFiles should not be empty if args.Files is not empty.
			finalFiles = append(finalFiles, pattern)
		}
	}

	if len(finalFiles) == 0 {
		fmt.Fprintf(os.Stderr, "Error: at least one file must be specified (use command line arguments or I18NEDT_FILES environment variable)\n")
		os.Exit(1)
	}

	// Construct Config
	config := &types.Config{
		Files:        finalFiles,
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
	files, err := i18n.LoadAllFiles(config.Files, config.PathAsLocale)
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
