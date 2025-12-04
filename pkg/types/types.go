package types

import (
	"os"
)

// ValueType represents the type of a value
type ValueType string

const (
	ValueTypeString ValueType = "string"
	ValueTypeJSON   ValueType = "json"
)

// Value represents a typed value that can be string or JSON
type Value struct {
	Type  ValueType `json:"type"`
	Value string    `json:"value"`
}

// Helper constructors
func NewStringValue(v string) *Value {
	return &Value{Type: ValueTypeString, Value: v}
}

func NewJSONValue(v string) *Value {
	return &Value{Type: ValueTypeJSON, Value: v}
}

// Config represents the application configuration
type Config struct {
	Files        []string
	Keys         []string
	Editor       string
	PrintOnly    bool // -p flag: print temp file content without launching editor
	NoTips       bool // -a flag: exclude AI tips from temp file content
	PathAsLocale bool // -P flag: use file path as locale instead of extracting BCP47 tag
	Flatten      bool // --flatten flag: flatten JSON files to key=value format
}

// I18nFile represents a single i18n JSON file
type I18nFile struct {
	Path   string
	Data   string // Raw JSON string
	Locale string // Locale identifier extracted from filename
}

// TempFile represents the temporary edit file
type TempFile struct {
	Path    string
	Keys    []string
	Locales []string
	Content map[string]map[string]*Value // key -> locale -> *Value
	Deletes []string                     // keys to delete
}

// KeyOperation represents an operation to perform on a key
type KeyOperation struct {
	Key    string
	Value  string
	Delete bool
}

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	return &Config{
		Editor: editor,
	}
}
