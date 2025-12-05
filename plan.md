# Implementation Plan: Namespace Support for i18nedt

## Objective
Support multi-file namespaces for locales (e.g., `locales/en/common.json`, `locales/en/auth.json`) instead of assuming a single file per locale.

## Proposed Changes

### 1. Update Data Structures
**File:** `pkg/types/types.go`
- Add `Namespace` field to `I18nFile` struct.

### 2. Update Locale & Namespace Detection
**File:** `internal/i18n/locale.go`
- Add `ParseNamespace(filePath, locale string) string` function.
    - Logic: If filename (without ext) equals locale, namespace is empty. Otherwise, namespace is the filename.

### 3. Update File Loading
**File:** `internal/i18n/parser.go`
- In `LoadFile`:
    - After calling `ParseLocaleFromPath`, call `ParseNamespace`.
    - Populate the `Namespace` field in the returned `I18nFile`.

### 4. Update Editor Logic (Reading/Merging)
**File:** `internal/editor/tempfile.go`
- Modify `CreateTempFile`:
    - When iterating through files to populate `temp.Content`:
        - Construct a **composite key**:
            - If `file.Namespace` is empty: use original `key`.
            - If `file.Namespace` is not empty: use `namespace:key`.
        - Store value in `temp.Content[compositeKey][file.Locale]`.
    - This ensures keys from different namespaces (e.g., `common:title`, `auth:title`) do not collide and are editable separately.

### 5. Update Editor Logic (Writing/Saving)
**File:** `internal/editor/tempfile.go`
- Modify `ApplyChanges`:
    - Iterate through `temp.Content` (which now has composite keys).
    - Parse the key to extract `namespace` and `realKey`.
        - If key contains `:`, split into `ns` and `key`.
        - Else `ns` is empty, `key` is original.
    - When finding the target file to update:
        - Iterate `files`.
        - Check if `file.Locale` matches AND `file.Namespace` matches.
        - Only update if both match.

## Verification
- Create a test case with:
    - `locales/en/common.json`: `{"hello": "world"}`
    - `locales/en/auth.json`: `{"login": "sign in"}`
- Run `i18nedt` and verify the editor shows keys like `# common:hello` and `# auth:login`.
- Verify saving updates the correct files.