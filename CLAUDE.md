# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Core Commands
- `make build` - Build the i18nedt binary to `bin/i18nedt`
- `go build -o bin/i18nedt cmd/i18nedt/main.go` - Alternative build command
- `make test` - Run all tests with verbose output
- `go test -v ./...` - Alternative test command
- `make test-coverage` - Run tests with coverage report (generates coverage.html)
- `make clean` - Clean build artifacts
- `make install` - Install binary to `/usr/local/bin`

### Running the Tool
- `./bin/i18nedt -k home.welcome src/locales/{zh-CN,en-US}.json` - Edit a key across multiple locales
- `EDITOR=vim ./bin/i18nedt -k nav.menu.title test-locales/*.json` - Run with specific editor

## Architecture Overview

### Project Structure
This is a Go CLI tool for editing i18n JSON files. The main components are:

- **cmd/i18nedt/main.go** - Entry point and orchestration logic
- **pkg/types/** - Core data structures (Config, I18nFile, TempFile)
- **internal/cli/** - Command line argument parsing and validation
- **internal/i18n/** - JSON file operations, key management, locale parsing
- **internal/editor/** - Temporary file creation, editor integration, change application

### Key Data Flow
1. CLI parsing → validation → file loading
2. Key expansion (non-leaf keys become all their child keys)
3. Temporary file creation with current values
4. Editor opens temporary file
5. Parse edited content → apply changes → save files

### Core Types
- `Config` - CLI configuration (files, keys, editor)
- `I18nFile` - Represents a locale file with path and parsed JSON data
- `TempFile` - Temporary editing session with keys, locales, content, and deletions

### Key Management
The tool supports nested key structures (e.g., `nav.menu.home.title`). The `internal/i18n/key.go` package handles:
- Dot-separated key path parsing (`ParseKeyPath`)
- Value retrieval and setting in nested maps (`GetValue`, `SetValue`)
- Key expansion for non-leaf keys (`ExpandKeys`)
- Automatic collection of all child keys under a prefix

### Temporary File Format
The editor uses a markdown-like format:
```
# home.welcome
* zh-CN
欢迎

* en-US
Welcome

#- old.key.to.delete
```

### Locale Detection
Locales are automatically extracted from filenames:
- `zh-CN.json` → `zh-CN`
- `locales/en/messages.json` → `en`

### Testing
- Test files use `_test.go` suffix
- `internal/testutil/` contains test utilities
- Uses standard Go testing framework

## Development Notes

### Error Handling
- All operations return explicit errors
- Main.go handles errors with user-friendly messages and exit codes
- Temporary files are cleaned up with defer

### File Operations
- JSON files are loaded into `map[string]interface{}` for flexibility
- Non-existent files are created automatically
- Empty maps are cleaned up after deletions

### Editor Integration
- Uses `$EDITOR` environment variable, defaults to `vim`
- Creates temporary files in current directory with `.i18nedt-{timestamp}.md` pattern
- Validates editor availability before launching