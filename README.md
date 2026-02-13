# outenv

A CLI tool for directory-based environment variable management.
Environment variables are automatically switched when you `cd`.

## Features

- Hierarchical merging of env files (parent directories are overridden by children)
- Shell hook for automatic apply and cleanup
- AES-256-GCM encryption for sensitive values
- Supports bash, zsh, and fish

## Installation

```bash
go install github.com/eji/outenv@latest
```

## Setup

Add the following to your shell configuration:

```bash
# bash (~/.bashrc)
eval "$(outenv hook bash)"

# zsh (~/.zshrc)
eval "$(outenv hook zsh)"

# fish (~/.config/fish/config.fish)
outenv hook fish | source
```

## Usage

```bash
# Create an env file for the current directory
outenv init

# Edit the env file
outenv edit

# Manually export variables (without shell hook)
eval "$(outenv export)"

# Encrypt a value and paste it into the env file
outenv encrypt "secret_value"
```

## Env file format

```bash
# Comments
DATABASE_URL=postgres://localhost/mydb
APP_TITLE="spaces are ok"

# Encrypted values
SECRET_KEY=ENC:base64...
```

## Directory hierarchy merging

Env files are merged from the root directory down to the current directory.
Child directories override parent values.

```
~/.local/share/outenv/
└── home/alice/
    ├── env                  # EDITOR=vim
    └── projects/myapp/
        └── env              # DATABASE_URL=postgres://...
```

When you `cd /home/alice/projects/myapp`, both `EDITOR` and `DATABASE_URL` are set.
When you leave the directory, the previous values are automatically restored.

## Encryption

Sensitive values can be stored encrypted:

```bash
$ outenv encrypt "my_password"
ENC:base64encodedstring...
```

Paste the output into the env file:

```bash
DB_PASSWORD=ENC:base64encodedstring...
```

Values are automatically decrypted during `export` and shell hook execution.

- Algorithm: AES-256-GCM
- Key file: `$XDG_CONFIG_HOME/outenv/key` (default `~/.config/outenv/key`)
- The key is auto-generated on the first `outenv encrypt` invocation
- Key file permissions are set to `0600`

## File locations

Follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/).

| Type | Path | Environment variable |
|------|------|----------------------|
| Env data | `~/.local/share/outenv/` | `$XDG_DATA_HOME` |
| Encryption key | `~/.config/outenv/key` | `$XDG_CONFIG_HOME` |

## Commands

| Command | Description |
|---------|-------------|
| `outenv init` | Create an env file for the current directory |
| `outenv edit` | Open the env file in `$EDITOR` |
| `outenv export` | Print merged environment variables |
| `outenv encrypt <value>` | Encrypt a value |
| `outenv hook <shell>` | Print shell hook code |

## License

MIT
