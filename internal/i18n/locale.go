package i18n

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kikyous/i18nedt/pkg/types"
)

// ParseLocaleFromPath extracts locale code from file path
func ParseLocaleFromPath(filePath string) (string, error) {
	base := filepath.Base(filePath)

	// Remove extension
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Common patterns for locale extraction
	patterns := []string{
		`^([a-z]{2}-[A-Z]{2})$`,      // zh-CN, en-US
		`^([a-z]{2}-[A-Z][a-z]+)$`,    // zh-Hans, zh-Hant
		`^([a-z]{2})$`,               // en, zh
		`^([a-z]{2}-[A-Z]{2})\..*`,   // zh-CN.json, en-US.json
		`^([a-z]{2}-[A-Z][a-z]+)\..*`, // zh-Hans.json, zh-Hant.json
		`^([a-z]{2})\..*`,            // en.json, zh.json
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Try to extract from the full path
	// Look for patterns like src/locales/zh-CN/ or locales/zh/
	pathParts := strings.Split(filepath.Dir(filePath), string(filepath.Separator))
	for i := len(pathParts) - 1; i >= 0; i-- {
		part := pathParts[i]
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(part)
			if len(matches) > 1 {
				return matches[1], nil
			}
		}
	}

	// Default to filename without extension if no locale pattern found
	return name, nil
}


// GetLocaleList extracts locale codes from I18nFile structs
func GetLocaleList(files []*types.I18nFile) ([]string, error) {
	locales := make([]string, 0, len(files))

	for _, file := range files {
		locales = append(locales, file.Locale)
	}

	return locales, nil
}

// FindFileByLocale finds the I18nFile for a given locale
func FindFileByLocale(files []*types.I18nFile, locale string) *types.I18nFile {
	for _, file := range files {
		if file.Locale == locale {
			return file
		}
	}
	return nil
}
