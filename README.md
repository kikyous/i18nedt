# i18nedt - AI ready i18n command line Editor


<img src="https://vhs.charm.sh/vhs-2CTMxsib9GPXzJH64Bxsjg.gif" alt="Made with VHS">



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

When you run i18nedt, it opens your editor (defined by `$EDITOR` environment variable, defaults to `vim`) with a temporary file (for example `.i18nedt-1764129261.md` in current dir) in this format:

```md
you are a md file translator, add missing translations to this file.
key start with # and language start with *.
do not read or edit other file.(this is a tip for ai)

# home.welcome    <-- key and translations that already exist in the JSON file
* zh-CN
欢迎

* zh-TW
歡迎

* en-US
Welcome

# home.start    <-- new key, translations is empty
* zh-CN

* zh-TW

* en-US
```
you can edit this file, add translation or add new key, when you exit the editor, you change will be apply(write to json file).

### JSON value
JSON value is supported.
```bash
i18nedt src/locales/*.json -k home
```

```md
you are a md file translator, add missing translations to this file.
key start with # and language start with * or +.
do not read or edit other file.(this is a tip for ai)

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

+ zh-TW
{
  "start": "开始",
  "welcome": "你好"
}
```

## Key syntax
```bash
i18nedt src/locales/*.json -k home.start
i18nedt src/locales/*.json -k "home.st*"
i18nedt src/locales/*.json -k "*.start"
i18nedt src/locales/*.json -k array-key.0
i18nedt src/locales/*.json -k array-key.1
```

support most gjson syntax, check out [GJSON Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md)


## AI ready
as you can see, the heading of temp md file is an ai prompt, so you can add a few translations and submit it to ai to fill the missings.

![AI Ready](ai-ready.png)


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
export EDITOR=nano            # Use nano instead of vim
export EDITOR="code --wait"   # Use VS Code
export EDITOR="zed --wait"    # Use Zed
export EDITOR=emacs           # Use Emacs
```

or set it inline:
```bash
EDITOR="zed --wait" i18nedt locales/{zh-CN,zh-TW,en-US}.json -k test
```

## File Path Patterns

You can use glob patterns to specify multiple files:

```bash
# Using brace expansion
i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k welcome

# Using wildcards
i18nedt src/locales/*.json -k welcome

# Deep wildcards
i18nedt src/locales/**/*/common.json -k welcome

# Specific files
i18nedt locales/zh-CN.json locales/en-US.json -k welcome
```

Alternatively, you can set the `I18NEDT_FILES` environment variable with a glob pattern for your i18n files. This allows you to omit the file paths from the command line:

```bash
export I18NEDT_FILES="src/locales/*.json"
i18nedt -k welcome
```

recommend using `direnv` to automatic set different I18NEDT_FILES for different projects.


### Fuzzy Finding with fzf

You can integrate `i18nedt` with `fzf` for an interactive key selection experience. First, ensure you have `fzf` installed. Then, add the following alias to your shell configuration (e.g., `.bashrc`, `.zshrc`):

```bash
export I18NEDT_FILES="src/locales/*.json"   ## this ENV is required

alias fi18n="i18nedt -f | fzf \
     --bind 'enter:become:i18nedt -k {1}' \
     --bind 'ctrl-o:execute:i18nedt -k {1}' \
     --bind 'ctrl-x:become:i18nedt -k {q}' \
     --delimiter = --preview 'i18nedt -p -a -k {1}' \
     --preview-window '<80(up):wrap' --bind '?:toggle-preview'"
```



This alias allows you to:
-   Interactively fuzzy search keys.
-   Press `Enter` to edit the selected key.
-   Press `Ctrl-o` to edit the selected key and immediately open your editor.
-   Press `Ctrl-x` to edit a custom key (using your `fzf` query as the key).
-   Preview the content of the selected key (`i18nedt -p -a -k {1}`).


## Working with Namespaces

`i18nedt` supports working with internationalization files organized by namespaces, which often correspond to different modules or sections of an application. To enable namespace support, your file path patterns must include the `{{ns}}` (or `{{namespace}}`) placeholder in addition to the `{{language}}` (or `{{locale}}`) placeholder.

For example:

```bash
# Pattern to discover files like:
# locales/en-US/common.json
# locales/zh-CN/auth.json
i18nedt "locales/{{language}}/{{ns}}.json" -k common:home.title
```

When specifying keys, you can prepend the namespace to the key using a colon (`:`):

```bash
# Edit the 'home.title' key within the 'common' namespace
i18nedt "locales/{{language}}/{{ns}}.json" -k common:home.title

# Edit the 'welcome' key within the 'auth' namespace
i18nedt "locales/{{language}}/{{ns}}.json" -k auth:welcome
```

### Automatic Namespace Creation

If you specify a key for a namespace that does not yet have corresponding files, `i18nedt` will automatically detect this and prepare to create the necessary files.

For example, if you run:

```bash
i18nedt "locales/{{language}}/{{ns}}.json" -k newModule:title.greeting
```

And files for `newModule` (e.g., `locales/en-US/newModule.json`, `locales/zh-CN/newModule.json`) do not exist, `i18nedt` will:
1. Print a message indicating it's "Creating new namespace: newModule".
2. Include the `newModule:title.greeting` key in the temporary editor file.
3. Upon saving your changes in the editor, `i18nedt` will create the `newModule.json` files in the appropriate locale directories (`locales/en-US/newModule.json`, `locales/zh-CN/newModule.json`, etc.) with the new key and its translated value.

This functionality relies on having a pattern with both `{{language}}` (or `{{locale}}`) and `{{ns}}` (or `{{namespace}}`) placeholders. If no such pattern is provided, automatic namespace creation will not be possible.

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
