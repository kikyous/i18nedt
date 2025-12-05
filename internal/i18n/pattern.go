package i18n

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// PatternToGlob converts a pattern string with placeholders to a glob string
func PatternToGlob(pattern string) string {
	// Replace placeholders with *
	// We support {{language}}, {{locale}}, {{namespace}}, {{ns}}
	// and simplistic {language} style if we wanted, but let's stick to {{...}} per user request
	
	// Use simple string replacement for known placeholders
	// Note: order matters if one is prefix of another, but here they are distinct
	glob := pattern
	glob = strings.ReplaceAll(glob, "{{language}}", "*")
	glob = strings.ReplaceAll(glob, "{{locale}}", "*")
	glob = strings.ReplaceAll(glob, "{{namespace}}", "*")
	glob = strings.ReplaceAll(glob, "{{ns}}", "*")
	
	return glob
}

// ExtractMetadataFromPath extracts locale and namespace from a path using a pattern
func ExtractMetadataFromPath(path, pattern string) (string, string, error) {
	// Normalize path separators
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Create regex from pattern
	// Escape special regex characters in the pattern
	regexPattern := regexp.QuoteMeta(pattern)

	// Replace the escaped placeholders with named capture groups
	// Note: We need to replace the *escaped* versions of {{...}}
	// regexp.QuoteMeta escapes {, }, so {{ becomes {{
	
	replacements := map[string]string{
		"\\{\\{language\\}\\}":  "(?P<locale>[^/]+)",
		"\\{\\{locale\\}\\}":    "(?P<locale>[^/]+)",
		"\\{\\{namespace\\}\\}": "(?P<namespace>[^/]+)",
		"\\{\\{ns\\}\\}":        "(?P<namespace>[^/]+)",
	}

	for old, new := range replacements {
		regexPattern = strings.ReplaceAll(regexPattern, old, new)
	}
	
	// Handle wildcards that might have been in the original pattern (if any)
	// If the user put '*' in the pattern, QuoteMeta escaped it to '\*'.
	// We might want to revert that if we support mix of * and {{}}.
	// For now, let's assume the user uses * either as a literal (unlikely in filenames) or as a wildcard.
	// If they meant wildcard, they passed it in CLI.
	// But QuoteMeta escapes it.
	// Let's unescape \* back to [^/]+ (non-recursive match) or .*?
	// Given doublestar is used for finding files, a * in the pattern usually means "match anything in this segment".
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", "[^/]+")

	// Anchor full match
	regexPattern = "^" + regexPattern + "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return "", "", fmt.Errorf("invalid pattern regex: %w", err)
	}

	match := re.FindStringSubmatch(path)
	if match == nil {
		return "", "", fmt.Errorf("path %s does not match pattern %s", path, pattern)
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return result["locale"], result["namespace"], nil
}

// ConstructPathFromMetadata constructs a file path from a pattern and metadata
func ConstructPathFromMetadata(pattern, locale, namespace string) string {
	path := pattern
	path = strings.ReplaceAll(path, "{{language}}", locale)
	path = strings.ReplaceAll(path, "{{locale}}", locale)
	path = strings.ReplaceAll(path, "{{namespace}}", namespace)
	path = strings.ReplaceAll(path, "{{ns}}", namespace)
	return path
}

// HasNamespacePlaceholder checks if the pattern contains a namespace placeholder
func HasNamespacePlaceholder(pattern string) bool {
	return strings.Contains(pattern, "{{namespace}}") || strings.Contains(pattern, "{{ns}}")
}

// HasLocalePlaceholder checks if the pattern contains a locale placeholder
func HasLocalePlaceholder(pattern string) bool {
	return strings.Contains(pattern, "{{language}}") || strings.Contains(pattern, "{{locale}}")
}
