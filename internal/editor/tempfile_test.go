package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kikyous/i18nedt/internal/i18n"
	"github.com/kikyous/i18nedt/pkg/types"
)

func TestCreateTempFile(t *testing.T) {
	files := []*types.I18nFile{
		{Path: "zh-CN.json", Data: `{"welcome": "欢迎"}`, Locale: "zh-CN"},
		{Path: "en-US.json", Data: `{"welcome": "Welcome"}`, Locale: "en-US"},
	}

	keys := []string{"home.welcome", "nav.home"}

	temp, err := CreateTempFile(files, keys)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}

	// Cleanup
	defer os.Remove(temp.Path)

	if temp.Path == "" {
		t.Error("CreateTempFile() Path should not be empty")
	}

	// Verify the file is created in current directory and starts with dot and ends with .md
	filename := filepath.Base(temp.Path)
	if !strings.HasPrefix(filename, ".i18nedt-") {
		t.Errorf("CreateTempFile() Path should start with .i18nedt-, got %s", filename)
	}
	if !strings.HasSuffix(filename, ".md") {
		t.Errorf("CreateTempFile() Path should end with .md, got %s", filename)
	}

	// Verify the file is in current directory
	if filepath.Dir(temp.Path) != "." {
		t.Errorf("CreateTempFile() should create file in current directory, got %s", filepath.Dir(temp.Path))
	}

	if len(temp.Keys) != 2 {
		t.Errorf("CreateTempFile() Keys length = %v, want %v", len(temp.Keys), 2)
	}

	if len(temp.Locales) != 2 {
		t.Errorf("CreateTempFile() Locales length = %v, want %v", len(temp.Locales), 2)
	}

	if temp.Content == nil {
		t.Error("CreateTempFile() Content should not be nil")
	}

	if len(temp.Deletes) != 0 {
		t.Errorf("CreateTempFile() Deletes length = %v, want %v", len(temp.Deletes), 0)
	}
}

func TestGenerateTempFileContent(t *testing.T) {
	temp := &types.TempFile{
		Keys:    []string{"home.welcome", "nav.home"},
		Locales: []string{"zh-CN", "en-US"},
		Content: map[string]map[string]*types.Value{
			"home.welcome": {
				"zh-CN": types.NewStringValue("欢迎"),
				"en-US": types.NewStringValue("Welcome"),
			},
			"nav.home": {
				"zh-CN": types.NewStringValue("首页"),
				"en-US": types.NewStringValue("Home"),
			},
		},
	}

	content, err := GenerateTempFileContent(temp)
	if err != nil {
		t.Fatalf("GenerateTempFileContent() error = %v", err)
	}

	// Check that the content contains expected parts
	expectedParts := []string{
		"# home.welcome",
		"* zh-CN",
		"欢迎",
		"* en-US",
		"Welcome",
		"# nav.home",
		"首页",
		"Home",
	}

	for _, part := range expectedParts {
		if !strings.Contains(string(content), part) {
			t.Errorf("GenerateTempFileContent() missing expected part: %s", part)
		}
	}
}

func TestWriteTempFile(t *testing.T) {
	temp := &types.TempFile{
		Path:    "/tmp/test-i18nedt.txt",
		Keys:    []string{"home.welcome"},
		Locales: []string{"zh-CN"},
		Content: map[string]map[string]*types.Value{
			"home.welcome": {
				"zh-CN": types.NewStringValue("欢迎"),
			},
		},
	}

	err := WriteTempFile(temp)
	if err != nil {
		t.Fatalf("WriteTempFile() error = %v", err)
	}

	// Cleanup
	defer os.Remove(temp.Path)

	// Verify file was created
	if _, err := os.Stat(temp.Path); os.IsNotExist(err) {
		t.Error("WriteTempFile() file was not created")
	}

	// Verify file content
	data, err := os.ReadFile(temp.Path)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "欢迎") {
		t.Error("WriteTempFile() file doesn't contain expected content")
	}
}

func TestParseTempFileContent(t *testing.T) {
	content := `# home.welcome
* zh-CN
欢迎

* en-US
Welcome

# nav.home
* zh-CN
首页

* en-US
Home`

	locales := []string{"zh-CN", "en-US"}
	temp, err := ParseTempFileContent(content, locales)
	if err != nil {
		t.Fatalf("ParseTempFileContent() error = %v", err)
	}

	// Verify parsed content
	if temp.Content["home.welcome"]["zh-CN"].Value != "欢迎" {
		t.Errorf("ParseTempFileContent() zh-CN welcome = %v, want %v", temp.Content["home.welcome"]["zh-CN"].Value, "欢迎")
	}

	if temp.Content["home.welcome"]["en-US"].Value != "Welcome" {
		t.Errorf("ParseTempFileContent() en-US welcome = %v, want %v", temp.Content["home.welcome"]["en-US"].Value, "Welcome")
	}

	if temp.Content["nav.home"]["zh-CN"].Value != "首页" {
		t.Errorf("ParseTempFileContent() zh-CN home = %v, want %v", temp.Content["nav.home"]["zh-CN"].Value, "首页")
	}

	if temp.Content["nav.home"]["en-US"].Value != "Home" {
		t.Errorf("ParseTempFileContent() en-US home = %v, want %v", temp.Content["nav.home"]["en-US"].Value, "Home")
	}
}

func TestParseTempFileContentWithDeletes(t *testing.T) {
	content := `#- old.key
# new.key
* zh-CN
新的键

* en-US
New key

# existing.key
* zh-CN
现有值

* en-US
Existing value`

	locales := []string{"zh-CN", "en-US"}
	temp, err := ParseTempFileContent(content, locales)
	if err != nil {
		t.Fatalf("ParseTempFileContent() error = %v", err)
	}

	// Verify delete markers
	if len(temp.Deletes) != 1 {
		t.Errorf("ParseTempFileContent() Deletes length = %v, want %v", len(temp.Deletes), 1)
	}

	if temp.Deletes[0] != "old.key" {
		t.Errorf("ParseTempFileContent() Deletes[0] = %v, want %v", temp.Deletes[0], "old.key")
	}

	// Verify new content
	if temp.Content["new.key"]["zh-CN"].Value != "新的键" {
		t.Errorf("ParseTempFileContent() new key zh-CN = %v, want %v", temp.Content["new.key"]["zh-CN"].Value, "新的键")
	}

	if temp.Content["existing.key"]["en-US"].Value != "Existing value" {
		t.Errorf("ParseTempFileContent() existing key en-US = %v, want %v", temp.Content["existing.key"]["en-US"].Value, "Existing value")
	}
}

func TestReadTempFile(t *testing.T) {
	content := `# home.welcome
* zh-CN
欢迎

* en-US
Welcome`

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-read-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test content
	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	temp := &types.TempFile{
		Path:    tmpFile.Name(),
		Keys:    []string{"home.welcome"},
		Locales: []string{"zh-CN", "en-US"},
		Content: make(map[string]map[string]*types.Value),
		Deletes: []string{},
	}

	err = ReadTempFile(temp)
	if err != nil {
		t.Fatalf("ReadTempFile() error = %v", err)
	}

	// Verify parsed content
	if temp.Content["home.welcome"]["zh-CN"].Value != "欢迎" {
		t.Errorf("ReadTempFile() zh-CN welcome = %v, want %v", temp.Content["home.welcome"]["zh-CN"].Value, "欢迎")
	}
}

func TestCleanupTempFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-cleanup-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	temp := &types.TempFile{
		Path: tmpFile.Name(),
	}

	// Verify file exists
	if _, err := os.Stat(temp.Path); os.IsNotExist(err) {
		t.Fatalf("Test temp file does not exist")
	}

	err = CleanupTempFile(temp)
	if err != nil {
		t.Errorf("CleanupTempFile() error = %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(temp.Path); !os.IsNotExist(err) {
		t.Error("CleanupTempFile() file was not deleted")
	}
}

func TestApplyChanges(t *testing.T) {
	files := []*types.I18nFile{
		{Path: "zh-CN.json", Data: `{"old": "旧值"}`, Locale: "zh-CN"},
		{Path: "en-US.json", Data: `{"old": "Old value"}`, Locale: "en-US"},
	}

	temp := &types.TempFile{
		Content: map[string]map[string]*types.Value{
			"new": {
				"zh-CN": types.NewStringValue("新值"),
				"en-US": types.NewStringValue("New value"),
			},
		},
		Deletes: []string{"old"},
	}

	err := ApplyChanges(files, temp)
	if err != nil {
		t.Fatalf("ApplyChanges() error = %v", err)
	}

	// Verify old keys were deleted
	if value, err := i18n.GetValue(files[0].Data, "old"); err == nil && value != "" {
		t.Errorf("ApplyChanges() old key still exists in zh-CN: %v", value)
	}

	// Verify new keys were added
	if value, err := i18n.GetValue(files[0].Data, "new"); err != nil || value != "新值" {
		t.Errorf("ApplyChanges() new key in zh-CN = %v, want %v", value, "新值")
	}

	if value, err := i18n.GetValue(files[1].Data, "new"); err != nil || value != "New value" {
		t.Errorf("ApplyChanges() new key in en-US = %v, want %v", value, "New value")
	}

	// Verify dirty flag
	if !files[0].Dirty {
		t.Error("ApplyChanges() zh-CN file should be dirty")
	}
	if !files[1].Dirty {
		t.Error("ApplyChanges() en-US file should be dirty")
	}
}

func TestGetFilePaths(t *testing.T) {
	files := []*types.I18nFile{
		{Path: "zh-CN.json", Data: "{}"},
		{Path: "en-US.json", Data: "{}"},
		{Path: "ja-JP.json", Data: "{}"},
	}

	paths := getFilePaths(files)
	expected := []string{"zh-CN.json", "en-US.json", "ja-JP.json"}

	if len(paths) != len(expected) {
		t.Errorf("getFilePaths() length = %v, want %v", len(paths), len(expected))
	}

	for i, path := range paths {
		if path != expected[i] {
			t.Errorf("getFilePaths()[%d] = %v, want %v", i, path, expected[i])
		}
	}
}