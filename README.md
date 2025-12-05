# i18nedt - AI Ready i18n CLI Editor

<img src="https://vhs.charm.sh/vhs-2CTMxsib9GPXzJH64Bxsjg.gif" alt="Made with VHS">

**i18nedt** is a command-line tool designed to streamline the editing of internationalization (i18n) JSON files. It converts complex JSON structures into a simplified, Markdown-like format, allowing you to edit translations across multiple locales simultaneously using your favorite text editor.

It is specifically designed to be **AI-friendly**, making it effortless to generate translations using LLMs.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [The Editing Format](#the-editing-format)
    - [Basic Editing](#basic-editing)
    - [JSON Values](#json-values)
    - [Deleting Keys](#deleting-keys)
    - [Renaming Keys](#renaming-keys)
- [Key Selection Syntax](#key-selection-syntax)
- [AI Workflow](#ai-workflow)
- [Advanced Configuration](#advanced-configuration)
    - [Editor Configuration](#editor-configuration)
    - [File Selection & Glob Patterns](#file-selection--glob-patterns)
    - [Working with Namespaces](#working-with-namespaces)
- [CLI Reference](#cli-reference)
- [Integrations](#integrations)
    - [Fuzzy Finding with fzf](#fuzzy-finding-with-fzf)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Multi-file Editing**: specific keys across `zh-CN.json`, `en-US.json`, etc., in a single view.
- **Deep Nesting Support**: Handles complex JSON structures (e.g., `home.welcome.title`) effortlessly.
- **AI-Ready**: Generates a context-rich format optimized for prompting LLMs to fill in missing translations.
- **Namespace Support**: Works with folder-based structures (e.g., `locales/en/common.json`).
- **Editor Agnostic**: Uses your `$EDITOR` (Vim, VS Code, Nano, Zed, etc.).
- **Safety**: Automatically creates non-existent keys and files.

## Installation

### From Source (Recommended for developers)

```bash
git clone https://github.com/kikyous/i18nedt.git
cd i18nedt
make build
# Binary will be in ./bin/i18nedt
```

### Using Go Install

```bash
go install github.com/kikyous/i18nedt/cmd/i18nedt@latest
```

## Quick Start

**Edit a single key across all locales:**
```bash
i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k home.welcome
```

**Edit multiple keys:**
```bash
i18nedt src/locales/*.json -k home.welcome -k home.start
```

**Edit a parent key (and all its children):**
```bash
i18nedt src/locales/*.json -k home
```

## The Editing Format

When `i18nedt` opens your editor, you will see a temporary `.md` file. This format is designed to be human-readable and AI-parseable.

### Basic Editing

```markdown
you are a md file translator, add missing translations to this file.
key start with # and language start with *.
do not read or edit other file.(this is a tip for ai)

# home.welcome    <-- Existing key
* zh-CN
欢迎
* en-US
Welcome

# home.start      <-- New key (values empty)
* zh-CN
* en-US
```

Simply edit the text under the locale codes. When you save and exit, `i18nedt` updates the JSON files.

### JSON Values

You can also provide raw JSON objects or arrays as values with `+` mark.

```markdown
# home
+ en-US
{
  "start": "Start",
  "welcome": "Hello"
}

+ zh-CN
{
  "start": "开始",
  "welcome": "你好"
}
```

### Deleting Keys

To delete a key, change the `#` to `#-`.

```markdown
#- home.deprecated_key
```

### Renaming Keys

Since `i18nedt` focuses on a diff-like editing experience, renaming a key is a manual two-step process within the temporary editing file:

1.  **Delete the old key**: Mark the old key for deletion using `#-`.
2.  **Create the new key**: Add the desired new key name using `#`.

Example: To rename `home.welcome` to `home.greeting` and migrate its translations:

```markdown
#- home.welcome   <-- Mark old key for deletion
# home.greeting   <-- New key with migrated translations
* zh-CN
欢迎词

* en-US
Greetings
```
This approach ensures that the historical content is preserved for the new key.

## Key Selection Syntax

`i18nedt` supports flexible key selection, powered by [GJSON Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md).

```bash
# Standard dot notation
i18nedt ... -k home.start

# Wildcards
i18nedt ... -k "home.st*"
i18nedt ... -k "*.start"

# Array access
i18nedt ... -k array-key.0
```

## AI Workflow

The temporary file format is optimized for interaction with AI.

**Using AI Code Editors (Cursor, Windsurf, VS Code)**
If you are using an editor with built-in AI, you don't need to copy-paste.
1. Run `i18nedt` to open the temporary file.
2. In your AI chat, reference the temporary file (e.g., type `@.i18nedt-xxx.md`).
3. the AI will understand the context and generate the translations for you.

**Standard LLMs**
The file header acts as a system prompt. You can also copy the entire content into ChatGPT/Claude/Gemini and paste the result back.

![AI Ready](ai-ready.png)

## Advanced Configuration

### Editor Configuration

`i18nedt` uses the `$EDITOR` environment variable.

```bash
# Set explicitly
export EDITOR="code --wait"   # VS Code
export EDITOR="zed --wait"    # Zed
export EDITOR="vim"           # Vim (Default)

# One-off usage
EDITOR=nano i18nedt ...
```

### File Selection & Glob Patterns

You can specify files using arguments or an environment variable.

**Using Arguments:**
```bash
# Brace expansion
i18nedt src/locales/{zh-CN,en-US}.json -k welcome

# Deep wildcards
i18nedt src/locales/**/*/common.json -k welcome
```

**Using Environment Variable (`I18NEDT_FILES`):**
This allows you to omit the file path in every command. Highly recommended to set this via `.env` or `direnv`.

```bash
export I18NEDT_FILES="src/locales/*.json"
i18nedt -k welcome
```

### Working with Namespaces

If your project uses namespaces (e.g., `locales/en/common.json`, `locales/en/auth.json`), use the `{{ns}}` (or `{{namespace}}`) and `{{language}}` (or `{{locale}}`) placeholders in your pattern.

**1. Define the pattern:**
```bash
# Pattern matches: locales/en-US/common.json
i18nedt "locales/{{language}}/{{ns}}.json" -k common:home.title
```

**2. Select keys with `namespace:key` syntax:**
```bash
# Edit 'home.title' in 'common.json'
i18nedt ... -k common:home.title

# Edit 'login' in 'auth.json'
i18nedt ... -k auth:login
```

**Automatic Namespace Creation:**
If you reference a namespace that doesn't exist (e.g., `-k newPage:title`), `i18nedt` will automatically create the corresponding JSON files (e.g., `locales/en/newPage.json`) upon saving.

## CLI Reference

```text
Usage: i18nedt [--key KEY] [--print] [--no-tips] [--flatten] [--version] [FILES]

Positional arguments:
  FILES                  Target file paths [env: I18NEDT_FILES]

Options:
  --key KEY, -k KEY      Key to edit (can be specified multiple times)
  --print, -p            Print temporary file content without launching editor
  --no-tips, -a          Exclude AI tips from temporary file content
  --flatten, -f          Flatten JSON files to key=value format
  --version, -v          Show version information
  --help, -h             display this help and exit
```

## Integrations

### Fuzzy Finding with fzf

Combine `i18nedt` with `fzf` for an interactive translation experience. Add this alias to your shell (`.bashrc` or `.zshrc`):

```bash
# Requirement: I18NEDT_FILES must be set
export I18NEDT_FILES="src/locales/*.json"

alias fi18n="i18nedt -f | fzf \
     --bind 'enter:become:i18nedt -k {1}' \
     --bind 'ctrl-o:execute:i18nedt -k {1}' \
     --bind 'ctrl-x:become:i18nedt -k {q}' \
     --delimiter = --preview 'i18nedt -p -a -k {1}' \
     --preview-window '<80(up):wrap' --bind '?:toggle-preview'"
```

**Usage:**
- Run `fi18n` to search keys.
- `Enter`: Edit selected key.
- `Ctrl-x`: Create/Edit a new key using your search query.
- `?`: Toggle preview of values.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
