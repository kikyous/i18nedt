package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kikyous/i18nedt/pkg/types"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "i18nedt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

// CreateTempFile creates a temporary file with given content
func CreateTempFile(t *testing.T, content string) string {
	tmpFile, err := ioutil.TempFile("", "i18nedt-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != "" {
		_, err = tmpFile.WriteString(content)
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	tmpFile.Close()
	return tmpFile.Name()
}

// CreateTestFiles creates test i18n files with given content
func CreateTestFiles(t *testing.T, tmpDir string, files map[string]string) []*types.I18nFile {
	i18nFiles := make([]*types.I18nFile, 0, len(files))

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		i18nFile := &types.I18nFile{
			Path: filePath,
			Data: make(map[string]interface{}),
		}
		i18nFiles = append(i18nFiles, i18nFile)
	}

	return i18nFiles
}

// CleanupFiles removes test files
func CleanupFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

// AssertEqualMaps compares two maps for equality
func AssertEqualMaps(t *testing.T, a, b map[string]interface{}) {
	if len(a) != len(b) {
		t.Errorf("Maps have different lengths: %d vs %d", len(a), len(b))
		return
	}

	for k, v := range a {
		if bv, exists := b[k]; !exists {
			t.Errorf("Key %s exists in first map but not second", k)
		} else if !EqualInterface(v, bv) {
			t.Errorf("Key %s values differ: %v vs %v", k, v, bv)
		}
	}
}

// EqualInterface compares two interface values for equality
func EqualInterface(a, b interface{}) bool {
	switch a := a.(type) {
	case map[string]interface{}:
		b, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return EqualMaps(a, b)
	case []interface{}:
		b, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if !EqualInterface(a[i], b[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// EqualMaps compares two maps for equality (helper for EqualInterface)
func EqualMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, exists := b[k]; !exists || !EqualInterface(v, bv) {
			return false
		}
	}
	return true
}

// CreateTestJSONContent creates JSON content for test files
func CreateTestJSONContent(data map[string]interface{}) string {
	content := "{\n"
	for k, v := range data {
		switch v := v.(type) {
		case string:
			content += `  "` + k + `": "` + v + `",` + "\n"
		default:
			content += `  "` + k + `": ` + `"placeholder"` + "," + "\n"
		}
	}
	content = content[:len(content)-2] + "\n}" // Remove last comma and closing
	return content
}

// CreateNestedTestData creates nested test data structure
func CreateNestedTestData() map[string]interface{} {
	return map[string]interface{}{
		"home": map[string]interface{}{
			"welcome": "Welcome",
			"goodbye": "Goodbye",
		},
		"nav": map[string]interface{}{
			"menu": map[string]interface{}{
				"home": "Home",
				"about": "About",
			},
		},
		"simple": "Simple value",
	}
}

// CreateSimpleTestData creates simple flat test data structure
func CreateSimpleTestData() map[string]interface{} {
	return map[string]interface{}{
		"welcome": "Welcome",
		"goodbye": "Goodbye",
		"home":    "Home",
		"about":   "About",
	}
}