package types

import "os"

// Config represents the application configuration
type Config struct {
	Files  []string
	Keys   []string
	Editor string
}

// I18nFile represents a single i18n JSON file
type I18nFile struct {
	Path string
	Data map[string]interface{}
}

// TempFile represents the temporary edit file
type TempFile struct {
	Path    string
	Keys    []string
	Locales []string
	Content map[string]map[string]string // key -> locale -> value
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