# i18nedt Project Context

## Project Overview
**i18nedt** is an AI-ready command-line interface (CLI) tool designed to streamline the editing of internationalization (i18n) JSON files. It allows developers to use their preferred text editors (Vim, VS Code, etc.) to modify translations across multiple locale files simultaneously.

The tool converts complex JSON structures into a simplified, flat Markdown-like format for editing. This format is specifically designed to be "AI-friendly," enabling users to easily copy context into LLMs for automated translation or expansion.

### Key Features
- **Multi-file Editing:** Edit keys across `zh-CN.json`, `en-US.json`, etc., at once.
- **Nested Key Support:** Handles deep JSON structures (e.g., `home.welcome.title`).
- **Editor Agnostic:** Works with any editor defined via `$EDITOR`.
- **AI-Ready Format:** Generates temporary files with clear prompts for AI translation assistance.
- **Glob Support:** Flexible file selection using standard glob patterns.

## Technology Stack
- **Language:** Go (v1.24.0)
- **CLI Parsing:** `github.com/alexflint/go-arg`
- **JSON Manipulation:** `github.com/tidwall/gjson` and `sjson`
- **TUI/Styling:** `github.com/charmbracelet/bubbletea`, `lipgloss` (likely for output formatting/interaction)
- **File Matching:** `github.com/bmatcuk/doublestar`

## Architecture & Directory Structure

### Core Components
- **`cmd/i18nedt/main.go`**: Application entry point and orchestration.
- **`internal/i18n/`**:
    - **`parser.go`**: Handles parsing of JSON files.
    - **`key.go`**: Manages dot-notation key paths and value retrieval/setting.
    - **`locale.go`**: Detects locales from filenames.
- **`internal/editor/`**:
    - **`tempfile.go`**: Generates the temporary `.md` file for editing and parses the result back into JSON updates.
    - **`editor.go`**: Handles invoking the external editor command.
- **`pkg/types/`**: Defines shared domain types like `Config`, `I18nFile`, and `TempFile`.

### Data Flow
1.  **Input**: User provides file patterns and target keys via CLI args.
2.  **Load**: JSON files are identified and loaded into memory.
3.  **Flatten**: Requested keys (and children) are extracted.
4.  **Edit**: A temporary Markdown file is generated and opened in `$EDITOR`.
5.  **Sync**: User saves and quits; the tool parses the Markdown and updates the original JSON files.

## Building and Running

### Prerequisites
- Go 1.24+
- Make (optional, for convenience commands)

### Commands
*   **Build Binary**:
    ```bash
    make build
    # OR
    go build -o bin/i18nedt cmd/i18nedt/main.go
    ```
*   **Run Tests**:
    ```bash
    make test
    # OR
    go test -v ./...
    ```
*   **Run Binary**:
    ```bash
    ./bin/i18nedt src/locales/*.json -k home
    ```
*   **Install**:
    ```bash
    go install github.com/kikyous/i18nedt/cmd/i18nedt@latest
    ```

## Development Conventions

### Code Style
- Standard Go formatting (`gofmt`).
- Error handling is explicit; almost all operations return `error`.
- Modular package design separating CLI concerns (`cmd`), internal logic (`internal`), and shared types (`pkg`).

### Testing
- Tests are co-located with source files (e.g., `key_test.go`).
- `internal/testutil/` provides helpers for mocking or setup.
- `test-locales/` contains real JSON files used for integration testing or manual verification.

### Editor Format Spec
The temporary file format is critical logic. It follows this structure:
```markdown
# key.path
* locale-code
Translation Content
```
- Lines starting with `#` denote keys.
- Lines starting with `*` denote locales.
- Lines starting with `#-` denote keys to be deleted.
