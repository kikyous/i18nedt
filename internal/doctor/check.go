package doctor

import (
	"fmt"
	"sort"

	"github.com/kikyous/i18nedt/internal/flatten"
	"github.com/kikyous/i18nedt/pkg/types"
)

// CheckResult holds the result of a check for a single file/locale
type CheckResult struct {
	File        *types.I18nFile
	MissingKeys []string
	EmptyKeys   []string
}

// Run executes the doctor check on the provided files and prints the report
// Returns true if issues were found, false otherwise
func Run(files []*types.I18nFile, simple bool) (bool, error) {
	results, err := Check(files)
	if err != nil {
		return false, err
	}

	if simple {
		// Collect unique keys
		keySet := make(map[string]bool)
		for _, res := range results {
			for _, k := range res.MissingKeys {
				keySet[k] = true
			}
			for _, k := range res.EmptyKeys {
				keySet[k] = true
			}
		}

		if len(keySet) == 0 {
			return false, nil
		}

		// Sort
		var keys []string
		for k := range keySet {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// Print
		for _, k := range keys {
			fmt.Println(k)
		}
		return true, nil
	}

	hasIssues := false

	// Sort keys (file paths) for consistent output
	var paths []string
	for p := range results {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		res := results[path]
		if len(res.MissingKeys) > 0 || len(res.EmptyKeys) > 0 {
			hasIssues = true
			fmt.Printf("File: %s (Locale: %s, Namespace: %s)\n", res.File.Path, res.File.Locale, res.File.Namespace)

			if len(res.MissingKeys) > 0 {
				fmt.Println("  Missing Keys:")
				for _, k := range res.MissingKeys {
					fmt.Printf("    - %s\n", k)
				}
			}

			if len(res.EmptyKeys) > 0 {
				fmt.Println("  Empty Keys:")
				for _, k := range res.EmptyKeys {
					fmt.Printf("    - %s\n", k)
				}
			}
			fmt.Println()
		}
	}

	if !hasIssues {
		fmt.Println("No issues found! All keys are present and non-empty.")
		return false, nil
	}

	return true, nil
}

// Check performs the analysis and returns results
func Check(files []*types.I18nFile) (map[string]CheckResult, error) {
	results := make(map[string]CheckResult)

	// Group files by Namespace
	// Map: Namespace -> Locale -> File
	groups := make(map[string]map[string]*types.I18nFile)

	for _, file := range files {
		ns := file.Namespace
		if groups[ns] == nil {
			groups[ns] = make(map[string]*types.I18nFile)
		}
		groups[ns][file.Locale] = file
	}

	// Iterate over each namespace
	for _, localeFiles := range groups {
		// 1. Flatten all files in this namespace and collect ALL keys
		allKeys := make(map[string]bool)
		fileFlats := make(map[string]map[string]string) // Locale -> FlatMap

		for locale, file := range localeFiles {
			flat, err := flatten.FlattenJSON([]byte(file.Data), file.Namespace)
			if err != nil {
				return nil, fmt.Errorf("failed to flatten file %s: %w", file.Path, err)
			}
			fileFlats[locale] = flat
			for k := range flat {
				allKeys[k] = true
			}
		}

		sortedKeys := make([]string, 0, len(allKeys))
		for k := range allKeys {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// 2. Check each locale against allKeys
		for locale, file := range localeFiles {
			flat := fileFlats[locale]
			var missing []string
			var empty []string

			// Check missing
			for _, k := range sortedKeys {
				if _, exists := flat[k]; !exists {
					missing = append(missing, k)
				}
			}

			// Check empty
			for k, v := range flat {
				if v == "\"\"" {
					empty = append(empty, k)
				}
			}

			results[file.Path] = CheckResult{
				File:        file,
				MissingKeys: missing,
				EmptyKeys:   empty,
			}
		}
	}

	return results, nil
}
