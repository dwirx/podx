# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PODX is a secure encryption CLI tool for teams to manage secrets and share them safely via Git. It supports symmetric (password-based) and asymmetric (key-based) encryption with format-preserving `.env` file handling.

## Build Commands

```bash
make build          # Build binary for current platform
make install        # Build and install to /usr/local/bin/
make clean          # Remove build artifacts
make release        # Cross-compile for all platforms (linux/darwin/windows × amd64/arm64)
```

Individual platform builds:
```bash
make linux-amd64    # GOOS=linux GOARCH=amd64
make darwin-arm64   # GOOS=darwin GOARCH=arm64 (Apple Silicon)
make windows-amd64  # GOOS=windows GOARCH=amd64
```

## Testing

```bash
go test ./...                    # Run all tests
go test ./crypto -v              # Run crypto tests with verbose output
go test ./parser -v              # Run parser tests
go test ./... -cover             # Run with coverage
go test ./crypto -run TestAESGCM # Run specific test
```

Test files follow Go conventions:
- `crypto/aes_gcm_test.go` - AES-256-GCM encryption tests
- `crypto/chacha_test.go` - ChaCha20-Poly1305 tests
- `crypto/kdf_test.go` - Argon2id key derivation tests
- `crypto/crypto_test.go` - Encryptor interface tests
- `parser/env_test.go` - .env file parsing and encryption tests

## TUI (Terminal User Interface)

```bash
podx                          # Launch interactive TUI
```

The TUI provides four tabs:
- **Dashboard** - Project status, encryption status, recipients, quick actions
- **Commands** - Interactive menu to run any podx command
- **Security** - Live security check with detailed results
- **Files** - File browser with encryption/decryption capabilities

Navigation:
- `Tab` / `1,2,3,4` - Switch tabs
- `↑↓` / `j,k` - Navigate up/down
- `←→` / `h,l` - Back/Select
- `Enter` - Confirm action
- `r` - Refresh data
- `?` - Help overlay
- `q` - Quit

Files Tab:
- `Space` - Toggle file selection
- `a` - Select/deselect all files
- `e` - Encrypt selected files
- `d` - Decrypt selected files
- `p` - Toggle preview panel
- `g` - Go to path (directory input)
- `/` - Filter files

The `tui/` package uses:
- `bubbletea` - TUI framework (Elm architecture)
- `lipgloss` - Styling and layout
- `bubbles` - Pre-built components (list, viewport)

## Security Features

```bash
podx check                    # Run all security checks
podx check --fix              # Auto-fix gitignore issues
podx check --pre-commit       # Silent mode for git hooks
podx hook install             # Install pre-commit hook
podx hook uninstall           # Remove pre-commit hook
podx hook status              # Check if hook is installed
```

The `security/` package provides:
- `patterns.go` - Secret pattern detection (AWS keys, passwords, API keys, connection strings)
- `scanner.go` - File scanning with binary detection and directory exclusions
- `gitignore.go` - Gitignore validation and automatic fixing
- `check.go` - Main check command logic combining all validators
- `hook.go` - Pre-commit hook installation and management

## Running the CLI

```bash
go run . <command>           # Run directly
./podx <command>             # After building
```

Key commands for development:
- `podx init` - Initialize project with `.podx.yaml`
- `podx encrypt-all` / `podx decrypt-all` - Batch encrypt/decrypt
- `podx keygen -t age` - Generate Age key pair

## Architecture

```
main.go              CLI entry point, command routing, flag parsing
├── crypto/          Encryption implementations
│   ├── crypto.go    Encryptor interface and factory (NewEncryptor)
│   ├── aes_gcm.go   AES-256-GCM symmetric encryption
│   ├── chacha.go    ChaCha20-Poly1305 symmetric encryption
│   ├── age.go       Age X25519 asymmetric encryption
│   ├── gpg.go       GPG/PGP backend
│   └── kdf.go       Argon2id key derivation (DeriveKey, DeriveKeyWithSalt)
├── project/         Project workspace management
│   └── project.go   .podx.yaml config, EncryptAll/DecryptAll, recipient management
├── parser/          Format-preserving .env parsing
│   └── env.go       Parse .env → EnvEntry[], encrypt values only (KEY=ENC[...])
├── keygen/          Key generation utilities
│   └── keygen.go    Age/GPG key generation, key storage (~/.config/podx/)
├── updater/         Self-update mechanism
│   └── updater.go   GitHub releases check/download
└── security/        Security checks and pre-commit hook
    ├── patterns.go  Secret pattern detection regex
    ├── scanner.go   File scanning for secrets
    ├── gitignore.go Gitignore validation/fixing
    ├── check.go     Check command logic
    └── hook.go      Pre-commit hook management
├── tui/             Interactive terminal UI
│   ├── tui.go       Main model and update loop
│   ├── dashboard.go Dashboard tab (project info, quick actions)
│   ├── commands.go  Commands tab (interactive menu)
│   ├── security.go  Security tab (check results)
│   ├── files.go     Files tab (file browser, encrypt/decrypt)
│   ├── styles.go    Lipgloss styles (Dracula-inspired palette)
│   └── keys.go      Key bindings
```

## Key Patterns

**Encryptor interface** (`crypto/crypto.go`): All symmetric algorithms implement `Encrypt(plaintext, key []byte)` and `Decrypt(ciphertext, key []byte)`. Use `crypto.NewEncryptor(algo)` to get an implementation.

**Format-preserving encryption**: `.env` files are parsed line-by-line. Only values are encrypted to `ENC[age:base64]` format while keys and comments stay readable.

**File format for symmetric encryption**: `[salt (16 bytes)][algo (1 byte)][ciphertext]`

**Project workflow**: `.podx.yaml` stores recipients (public keys) and secret file patterns. `encrypt-all` encrypts files to `.podx` extension and deletes originals.

## Dependencies

- `filippo.io/age` - Age X25519 encryption
- `golang.org/x/crypto` - Argon2id KDF
- `gopkg.in/yaml.v3` - YAML config parsing
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/charmbracelet/bubbles` - TUI components

## Version Injection

Version info is injected at build time via ldflags:
```bash
go build -ldflags "-X main.Version=v1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Key Storage Locations

- `~/.config/podx/age-keys.txt` - Private keys
- `~/.config/podx/age-recipients/default.txt` - Public key
