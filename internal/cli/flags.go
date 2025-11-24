package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chen/i18nedt/pkg/types"
)

// ParseFlags parses command line arguments and returns a Config
func ParseFlags() (*types.Config, error) {
	config := types.NewConfig()

	args := os.Args[1:] // Skip program name
	var keys []string
	var files []string

	i := 0
	for i < len(args) {
		arg := args[i]

		if strings.HasPrefix(arg, "-k") {
			// Handle -k flag
			if len(arg) == 2 {
				// -k space value format
				i++
				if i >= len(args) {
					return nil, fmt.Errorf("-k flag requires a value")
				}
				keys = append(keys, args[i])
			} else {
				// -kvalue format (without space)
				keys = append(keys, arg[2:])
			}
		} else if strings.HasPrefix(arg, "--key") {
			// Handle --key long form
			if len(arg) == 5 {
				// --key space value format
				i++
				if i >= len(args) {
					return nil, fmt.Errorf("--key flag requires a value")
				}
				keys = append(keys, args[i])
			} else {
				// --key=value format
				if arg[5] != '=' {
					return nil, fmt.Errorf("invalid --key format, expected --key=value")
				}
				keys = append(keys, arg[6:])
			}
		} else if arg == "-h" || arg == "--help" {
			// Show help
			PrintUsage()
			os.Exit(0)
		} else if strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		} else {
			// File path
			files = append(files, arg)
		}
		i++
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one key must be specified with -k or --key")
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("at least one file must be specified")
	}

	// Expand file paths
	expandedFiles, err := expandFilePaths(files)
	if err != nil {
		return nil, fmt.Errorf("failed to expand file paths: %w", err)
	}

	config.Keys = keys
	config.Files = expandedFiles
	return config, nil
}

// PrintUsage prints the usage information
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: i18nedt [options] files...\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  -k key, --key=key     Key to edit (can be specified multiple times)\n")
	fmt.Fprintf(os.Stderr, "  -h, --help            Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k home.welcome\n")
	fmt.Fprintf(os.Stderr, "  i18nedt -k home.welcome -k home.start src/locales/*.json\n")
	fmt.Fprintf(os.Stderr, "  i18nedt --key=nav.menu --key=footer src/locales/*.json\n")
	fmt.Fprintf(os.Stderr, "  i18nedt src/locales/*.json -k home  # -k can be anywhere\n")
}

// expandFilePaths expands file patterns (like glob) to actual file paths
func expandFilePaths(paths []string) ([]string, error) {
	var expanded []string

	for _, path := range paths {
		// Handle glob patterns
		if strings.Contains(path, "*") || strings.Contains(path, "{") {
			matches, err := filepath.Glob(path)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern %s: %w", path, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("no files match pattern: %s", path)
			}
			expanded = append(expanded, matches...)
		} else {
			expanded = append(expanded, path)
		}
	}

	return expanded, nil
}