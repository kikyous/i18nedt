package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/alexflint/go-arg"
	"github.com/kikyous/i18nedt/internal/doctor"
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
	Keys      []string `arg:"-k,--key,separate" help:"Key to edit (can be specified multiple times)"`
	PrintOnly bool     `arg:"-p,--print" help:"Print temporary file content without launching editor"`
	NoTips    bool     `arg:"-a,--no-tips,env" help:"Exclude AI tips from temporary file content"`
	Doctor    bool     `arg:"-d,--doctor" help:"Check for missing and empty keys"`
	Flatten   bool     `arg:"-f,--flatten" help:"Flatten JSON files to key=value format"`
	Separator string   `arg:"-s,--separator,env:SEPARATOR" default:":" help:"Namespace separator (default: ':')"`
	Version   bool     `arg:"-v,--version" help:"Show version information"`
	Files     []string `arg:"positional" help:"Target file paths [env: I18NEDT_FILES]"`
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
		Files:     flatFiles, // We keep this for Flatten logic which iterates simple paths
		Keys:      args.Keys,
		Editor:    os.Getenv("EDITOR"),
		PrintOnly: args.PrintOnly,
		NoTips:    args.NoTips,
		Flatten:   args.Flatten,
		Doctor:    args.Doctor,
		Separator: args.Separator,
	}
	if config.Editor == "" {
		config.Editor = "vim"
	}

	// Handle doctor mode
	if config.Doctor {
		runDoctor(sources, config.Flatten, config.Separator)
		return
	}

	// Handle flatten mode
	if config.Flatten {
		runFlatten(sources, config.Separator)
		return
	}

	// Validate editor
	if err := editor.ValidateEditor(config.Editor); err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %v\n", err)
		os.Exit(1)
	}

	runEditor(config, sources)
}

func runDoctor(sources []types.FileSource, simple bool, separator string) {
	// Load all i18n files
	files, err := i18n.LoadAllFiles(sources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading files: %v\n", err)
		os.Exit(1)
	}

	foundIssues, err := doctor.Run(files, simple, separator)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running doctor check: %v\n", err)
		os.Exit(1)
	}

	if foundIssues {
		os.Exit(1)
	}
}

func runFlatten(sources []types.FileSource, separator string) {
	// Load all i18n files
	files, err := i18n.LoadAllFiles(sources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading files: %v\n", err)
		os.Exit(1)
	}

	// Flatten each file
	for _, file := range files {
		flat, err := flatten.FlattenJSON([]byte(file.Data), file.Namespace, separator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error flattening file %s: %v\n", file.Path, err)
			os.Exit(1)
		}

		// Sort keys for consistent output
		keys := make([]string, 0, len(flat))
		for k := range flat {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Printf("%s = %s\n", k, flat[k])
		}
	}
}

func runEditor(config *types.Config, sources []types.FileSource) {
	// Load all i18n files
	files, err := i18n.LoadAllFiles(sources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading files: %v\n", err)
		os.Exit(1)
	}

	// Check for requested namespaces that don't exist and create them if possible
	files, createdNs, err := i18n.CreateMissingNamespaces(files, sources, config.Keys, config.Separator)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	for _, ns := range createdNs {
		fmt.Printf("Creating new namespace: %s\n", ns)
	}

	// Use keys directly without expansion - user explicitly specifies what to edit
	tempFile, err := editor.CreateTempFile(files, config.Keys, config.Separator)
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
	savedCount, err := i18n.SaveAllFiles(files)
	if err != nil {
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

	fmt.Printf("Successfully updated %d files\n", savedCount)
}
