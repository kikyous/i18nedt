package i18n

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chen/i18nedt/pkg/types"
)

// ParseLocaleFromPath extracts locale code from file path
func ParseLocaleFromPath(filePath string) (string, error) {
	base := filepath.Base(filePath)

	// Remove extension
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Common patterns for locale extraction
	patterns := []string{
		`^([a-z]{2}-[A-Z]{2})$`,     // zh-CN, en-US
		`^([a-z]{2})$`,              // en, zh
		`^([a-z]{2}-[A-Z]{2})\..*`,  // zh-CN.json, en-US.json
		`^([a-z]{2})\..*`,           // en.json, zh.json
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

// CreateI18nFiles creates I18nFile structs from file paths
func CreateI18nFiles(filePaths []string) ([]*types.I18nFile, error) {
	files := make([]*types.I18nFile, 0, len(filePaths))

	for _, filePath := range filePaths {
		file := &types.I18nFile{
			Path: filePath,
			Data: make(map[string]interface{}),
		}
		files = append(files, file)
	}

	return files, nil
}

// GetLocaleList extracts locale codes from file paths
func GetLocaleList(filePaths []string) ([]string, error) {
	locales := make([]string, 0, len(filePaths))

	for _, filePath := range filePaths {
		locale, err := ParseLocaleFromPath(filePath)
		if err != nil {
			return nil, err
		}
		locales = append(locales, locale)
	}

	return locales, nil
}

// FindFileByLocale finds the I18nFile for a given locale
func FindFileByLocale(files []*types.I18nFile, locale string) *types.I18nFile {
	for _, file := range files {
		fileLocale, err := ParseLocaleFromPath(file.Path)
		if err == nil && fileLocale == locale {
			return file
		}
	}
	return nil
}