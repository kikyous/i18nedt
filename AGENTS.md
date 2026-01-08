# Agentic Coding Guidelines for i18nedt

This repository contains a Go CLI tool for editing i18n JSON files. This document provides guidelines for agentic coding agents to operate effectively within this codebase.

## 1. Build & Test Commands

Use these commands to verify your changes.

- **Build:**
  ```bash
  make build
  # Output: bin/i18nedt
  ```

- **Run All Tests:**
  ```bash
  make test
  # Equivalent to: go test -v ./...
  ```

- **Run Single Test:**
  To run a specific test case (e.g., `TestGetValue`), identify the package (e.g., `./internal/i18n`) and run:
  ```bash
  go test -v -run TestGetValue ./internal/i18n
  ```

- **Check Coverage:**
  ```bash
  make test-coverage
  # Generates coverage.out and coverage.html
  ```

- **Clean:**
  ```bash
  make clean
  ```

- **Install:**
  ```bash
  make install
  # Installs to /usr/local/bin
  ```

## 2. Code Style & Conventions

Adhere strictly to standard Go conventions and existing project patterns.

### General
- **Formatting:** Always run `gofmt` (or let the IDE/tools handle it).
- **Module Path:** `github.com/kikyous/i18nedt`
- **Imports:** Group imports: Standard library first, third-party libraries second, internal/project imports last.
- **Entry Point:** `cmd/i18nedt/main.go` handles CLI orchestration.
- **Core Logic:** `internal/` contains business logic. `pkg/` is for shared types that might be imported by other projects (though this is primarily a CLI tool).

### Types & JSON Handling
- **Dynamic JSON:** This project intentionally uses `map[string]interface{}` and libraries like `github.com/tidwall/gjson` (read) and `github.com/tidwall/sjson` (write) for manipulating JSON files.
  - **Do NOT** attempt to define rigid structs for the user's i18n files, as their structure is variable.
  - **Do** use `pkg/types` for internal application structures (e.g., `Config`, `TempFile`).

### Error Handling
- **Return Errors:** Functions in `internal/` and `pkg/` must return `error` as the last return value.
- **Wrapping:** Wrap errors using `fmt.Errorf("context: %w", err)` to preserve the error chain.
- **CLI Exit:** Only `cmd/i18nedt/main.go` should call `os.Exit()`. Library code should return errors to the caller.

### Naming Conventions
- **Exported:** `PascalCase` (e.g., `ParseKeyPath`, `Config`).
- **Internal:** `camelCase` (e.g., `runEditor`, `sources`).
- **Tests:** `TestFunctionName` for unit tests.
- **Variables:** Short, idiomatic Go names (e.g., `f` for file, `err` for error, `ctx` for context).

### Testing
- **Table-Driven:** Use table-driven tests for logic with multiple input/output scenarios (see `internal/i18n/key_test.go` for examples).
- **Subtests:** Use `t.Run("description", func(t *testing.T) { ... })` within loops.
- **Deep Equality:** Use `reflect.DeepEqual` for comparing complex structures (maps, slices).
- **Location:** Test files must reside next to the source file (e.g., `key.go` -> `key_test.go`) and belong to the same package (e.g., `package i18n`).

## 3. Architecture Overview

- **cmd/i18nedt:** Main application entry point. Parses args and calls internal packages.
- **pkg/types:** Common data structures (`Config`, `I18nFile`, `TempFile`).
- **internal/i18n:** Core logic for reading/writing JSON, parsing keys (`gjson`/`sjson` wrappers).
- **internal/editor:** Manages the temporary editing session (generating the "markdown-like" edit format and parsing it back).
- **internal/doctor:** Logic for the `doctor` command (health checks).
- **internal/flatten:** Logic for flattening nested JSON to dot-notation keys.

## 4. Key Libraries

- **CLI:** `github.com/alexflint/go-arg` for argument parsing.
- **JSON:** `github.com/tidwall/gjson` (getting values) and `github.com/tidwall/sjson` (setting values).

## 5. Agent Workflow Notes

- **Modifying JSON Logic:** If modifying how keys are read/written, check `internal/i18n/key.go`. Ensure you handle nested keys (`grandparent.parent.child`) correctly using the dot-notation helper functions.
- **Editor Format:** The temporary file format is custom (Markdown-like). If changing it, update `internal/editor/tempfile.go` and ensure both generation and parsing logic are synchronized.
