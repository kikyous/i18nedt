# i18nedt - i18n Command Line Editor

A command-line tool for editing i18n JSON files with your favorite editor.

## Features

- Edit multiple i18n JSON files simultaneously
- Support for nested key structures (e.g., `home.welcome.title`)
- Delete keys with a simple syntax (`#- key.name`)
- Use your preferred editor via `$EDITOR` environment variable
- Automatic file creation for non-existent files
- Support for glob patterns in file paths

## Installation

```bash
go install github.com/kikyous/i18nedt/cmd/i18nedt@latest
```

## Usage

### Basic Usage

```bash
i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k home.welcome
```

### Edit Multiple Keys

```bash
i18nedt src/locales/*.json -k home.welcome -k home.start
```

### Nested Keys

```bash
i18nedt src/locales/*.json -k home
```

## Editor Format

When you run i18nedt, it opens your editor (defined by `$EDITOR` environment variable, defaults to `vim`) with a temporary file in this format:

```md
# home.welcome
* zh-CN
欢迎

* zh-TW
歡迎

* en-US
Welcome

# home.start
* zh-CN
开始

* zh-TW
開始

* en-US
Start
```
you can edit this file, add translation or add new key, when you exit the editor, you change will be apply.

## Deleting Keys

To delete a key, add `-` after the `#`:

```md
#- home.welcome
# home.welcomeNew
* zh-CN
新的欢迎

* zh-TW
新的歡迎

* en-US
New Welcome
```

In this example, `home.welcome` will be deleted and `home.welcomeNew` will be created.

## File Structure

i18nedt expects JSON files with this structure:

```json
{
  "home": {
    "welcome": "Welcome",
    "start": "Get Started"
  },
  "nav": {
    "menu": {
      "home": "Home",
      "about": "About"
    }
  }
}
```

## Configuration

The tool uses the `$EDITOR` environment variable to determine which editor to use. If not set, it defaults to `vim`.

```bash
export EDITOR=nano    # Use nano instead of vim
export EDITOR=code    # Use VS Code
export EDITOR=emacs   # Use Emacs
```

## File Path Patterns

You can use glob patterns to specify multiple files:

```bash
# Using brace expansion
i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k welcome

# Using wildcards
i18nedt src/locales/*.json -k welcome

# Specific files
i18nedt locales/zh-CN.json locales/en-US.json -k welcome
```

## Locale Detection

i18nedt automatically detects locale codes from file names:

- `zh-CN.json` → `zh-CN`
- `en-US.json` → `en-US`
- `locales/zh-CN/app.json` → `zh-CN`
- `locales/en/messages.json` → `en`

## Error Handling

- If a JSON file doesn't exist, it will be created automatically
- Invalid JSON syntax will be reported with file location
- Nested key structures are validated for correctness

## Examples

### Edit a single key across multiple languages

```bash
i18nedt src/locales/{zh-CN,en-US,ja-JP}.json -k app.title
```

### Edit multiple nested keys

```bash
i18nedt src/locales/*.json -k settings.profile.name -k settings.profile.email
```

### Delete old keys and add new ones

```
# Edit the temporary file:
#- old.deprecated.key
# new.modern.key
* zh-CN
现代化的内容

* en-US
Modern content
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
