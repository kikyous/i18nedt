package i18n

import (
	"path/filepath"
	"strings"

	"github.com/kikyous/i18nedt/pkg/types"
	"golang.org/x/text/language"
)

// extractBCP47TagStrict extracts BCP47 language tag from file path
func extractBCP47TagStrict(filePath string) (string, bool) {
	// 先去掉文件扩展名
	pathWithoutExt := strings.TrimSuffix(filePath, filepath.Ext(filePath))

	// 按路径分隔符分割
	pathParts := strings.Split(pathWithoutExt, string(filepath.Separator))

	// 反向遍历：从最深层开始向根目录遍历
	for i := len(pathParts) - 1; i >= 0; i-- {
		pathPart := pathParts[i]
		if pathPart == "" {
			continue
		}

		// 对每个路径部分，再按点号分割，也反向遍历
		subParts := strings.Split(pathPart, ".")
		for j := len(subParts) - 1; j >= 0; j-- {
			subPart := subParts[j]
			if subPart == "" {
				continue
			}

			tag, err := language.Parse(subPart)
			if err == nil {
				// 验证是否为有效的语言标签（非任意字符串）
				region, conf := tag.Region()
				// 只有当地区不是未知地区(ZZ)并且置信度不是No时才接受
				// 这确保我们不会把像"app"这样的任意字符串当作语言标签
				if region.String() != "ZZ" && conf != language.No {
					// 找到有效的语言标签
					return tag.String(), true
				}
			}
		}
	}

	return "", false
}

// ParseLocaleFromPath extracts locale code from file path using BCP47
func ParseLocaleFromPath(filePath string) (string, error) {
	if bcp47Tag, found := extractBCP47TagStrict(filePath); found {
		return bcp47Tag, nil
	}

	// Fallback to filename without extension if no BCP47 tag found
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return name, nil
}

// ParseNamespace extracts namespace from file path
func ParseNamespace(filePath, locale string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// If the filename (without extension) is the same as the locale,
	// it's a root locale file (e.g. en.json), so namespace is empty.
	if name == locale {
		return ""
	}

	// Otherwise, the filename is the namespace (e.g. common.json in en/common.json)
	return name
}

// GetLocaleList extracts unique locale codes from I18nFile structs
func GetLocaleList(files []*types.I18nFile) ([]string, error) {
	localeMap := make(map[string]bool)
	locales := make([]string, 0)

	for _, file := range files {
		if !localeMap[file.Locale] {
			localeMap[file.Locale] = true
			locales = append(locales, file.Locale)
		}
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
