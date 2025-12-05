package i18n

import (
	"fmt"
	"strings"

	"github.com/kikyous/i18nedt/pkg/types"
)

// CreateMissingNamespaces checks for requested namespaces that don't exist and creates them if possible.
// It returns the updated list of files and a list of created namespace names.
func CreateMissingNamespaces(files []*types.I18nFile, sources []types.FileSource, keys []string) ([]*types.I18nFile, []string, error) {
	// Identify existing namespaces
	existingNs := make(map[string]bool)
	for _, f := range files {
		existingNs[f.Namespace] = true
	}

	// Identify missing namespaces from requested keys
	missingNs := make(map[string]bool)
	for _, key := range keys {
		if strings.Contains(key, ":") {
			parts := strings.SplitN(key, ":", 2)
			ns := parts[0]
			if ns != "" && !existingNs[ns] {
				missingNs[ns] = true
			}
		}
	}

	if len(missingNs) == 0 {
		return files, nil, nil
	}

	// Find a suitable pattern for creating new files
	var templatePattern string
	for _, src := range sources {
		if HasNamespacePlaceholder(src.Pattern) && HasLocalePlaceholder(src.Pattern) {
			templatePattern = src.Pattern
			break
		}
	}

	if templatePattern == "" {
		return files, nil, fmt.Errorf("cannot create new namespaces because no pattern with {{ns}} and {{language}} placeholders was found")
	}

	// Get existing locales to create files for
	locales, _ := GetLocaleList(files)
	if len(locales) == 0 {
		return files, nil, fmt.Errorf("cannot create new namespaces because no existing locales found")
	}

	var createdNs []string
	for ns := range missingNs {
		createdNs = append(createdNs, ns)
		for _, loc := range locales {
			path := ConstructPathFromMetadata(templatePattern, loc, ns)
			newFile := &types.I18nFile{
				Path:      path,
				Data:      "{}",
				Locale:    loc,
				Namespace: ns,
				Dirty:     false, // Will be set to true if edited later
			}
			files = append(files, newFile)
		}
	}

	return files, createdNs, nil
}
