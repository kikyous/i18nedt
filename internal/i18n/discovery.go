package i18n

import (
	"fmt"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// DiscoverFiles finds all files matching the given patterns or globs.
// It returns a slice of FileSource for loading and a slice of flat file paths for legacy support (flatten).
func DiscoverFiles(patterns []string) ([]FileSource, []string, error) {
	var sources []FileSource
	var flatFiles []string

	// If no patterns provided, check environment variable
	if len(patterns) == 0 {
		if envFiles := os.Getenv("I18NEDT_FILES"); envFiles != "" {
			patterns = strings.Fields(envFiles)
		}
	}

	// Check again if we have patterns
	if len(patterns) == 0 {
		return nil, nil, fmt.Errorf("at least one file must be specified (use command line arguments or I18NEDT_FILES environment variable)")
	}

	for _, arg := range patterns {
		var globPattern string
		var extractionPattern string

		if strings.Contains(arg, "{{") {
			// It's a pattern with placeholders
			globPattern = PatternToGlob(arg)
			extractionPattern = arg
		} else {
			// It's a standard glob or file path
			globPattern = arg
			extractionPattern = ""
		}

		// Use doublestar for file globbing
		matches, err := doublestar.FilepathGlob(globPattern)
		if err == nil && len(matches) > 0 {
			for _, match := range matches {
				sources = append(sources, FileSource{
					Path:    match,
					Pattern: extractionPattern,
				})
				flatFiles = append(flatFiles, match)
			}
		} else {
			// If no match, add the original arg (globPattern) as path
			sources = append(sources, FileSource{
				Path:    globPattern,
				Pattern: extractionPattern,
			})
			flatFiles = append(flatFiles, globPattern)
		}
	}

	return sources, flatFiles, nil
}
