# PODX

<p align="center">
  <b>🔐 Secure Encryption CLI for Teams</b><br>
  Encrypt secrets, share with team, commit safely to Git
</p>

<p align="center">
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#commands">Commands</a> •
  <a href="#project-workflow">Project Workflow</a>
</p>

---

## Features

- 🔑 **Age & GPG Encryption** — Modern X25519 or traditional GPG
- 🔐 **Password Encryption** — AES-256-GCM with Argon2id key derivation
- 📁 **Project Workspaces** — Per-project `.podx.yaml` config
- 👥 **Multi-Recipient** — Share secrets with team members
- 📝 **Format-Preserving** — `.env` files stay readable (`KEY=ENC[...]`)
- 🖥️ **Interactive TUI** — Beautiful terminal interface with file browser
- 🛡️ **Security Checks** — Pre-commit hook, secret scanning, gitignore validation
- 🔄 **Self-Update** — Built-in update command
- 🌍 **Cross-Platform** — Linux, macOS, Windows

---

## Installation

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
```

### macOS (Apple Silicon)

```bash
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
```

### Windows (PowerShell as Admin)

```powershell
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex
```

### Build from Source

```bash
git clone https://github.com/dwirx/podx
cd podx
go build -o podx .
sudo mv podx /usr/local/bin/
```

### Uninstall

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/uninstall.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/uninstall.ps1 | iex
```

---

## Quick Start

```bash
# 1. Generate encryption key
podx keygen -t age

# 2. Initialize project
cd your-project
podx init

# 3. Create secrets
echo "API_KEY=secret123" > .env

# 4. Encrypt (deletes .env, creates .env.podx)
podx encrypt-all

# 5. Commit encrypted files
git add .podx.yaml .env.podx
git commit -m "Add encrypted secrets"

# 6. After clone, decrypt
podx decrypt-all
```

---

## Interactive TUI

Launch the interactive terminal interface by running `podx` without arguments:

```bash
podx
```

### TUI Features

The TUI provides four tabs:

| Tab | Description |
|-----|-------------|
| **Dashboard** | Project overview, encryption status, recipients, quick actions |
| **Commands** | Interactive menu to run any podx command |
| **Security** | Live security check results with findings |
| **Files** | File browser with encryption/decryption capabilities |

### Navigation Keys

| Key | Action |
|-----|--------|
| `Tab` / `1,2,3,4` | Switch between tabs |
| `↑↓` / `j,k` | Navigate up/down |
| `←→` / `h,l` | Navigate back/forward, select |
| `Enter` | Confirm action |
| `r` | Refresh data |
| `?` | Show help overlay |
| `q` / `Esc` | Quit |

### Files Tab

The Files tab provides a powerful file browser for encryption operations:

| Key | Action |
|-----|--------|
| `Space` | Toggle file selection |
| `a` | Select/deselect all files |
| `e` | Encrypt selected files |
| `d` | Decrypt selected files |
| `p` | Toggle preview panel |
| `g` | Go to path (enter directory path) |
| `/` | Filter files by name |

### Encryption Methods

When pressing `e` to encrypt, you can choose between two methods:

#### 🔐 Password Encryption (AES-GCM)

- Uses AES-256-GCM symmetric encryption
- Password-based with Argon2id key derivation
- Requires password to decrypt
- Good for personal files or when sharing password securely

#### 🔑 Age Key Encryption

- Uses Age X25519 asymmetric encryption
- Encrypts to configured recipients in `.podx.yaml`
- No password needed (uses key pairs)
- Good for team sharing via Git

### File Browser Features

- 📁 **Directory Navigation** — Browse directories with Enter/l, go back with h
- 🔍 **File Filtering** — Type `/` to filter files by name
- 📂 **Go-to Path** — Press `g` to jump to any directory
- 👁️ **File Preview** — Toggle preview panel with `p`
- ✓ **Multi-select** — Select multiple files with Space, select all with `a`
- 🎨 **Color Coding** — Files colored by type (Go=blue, Python=blue, config=green)
- 🔒 **Encryption Status** — Encrypted files shown with lock icon

---

## Commands

### Project Commands

| Command | Description |
|---------|-------------|
| `podx init` | Initialize project, create `.podx.yaml` |
| `podx add-recipient -n NAME -k KEY` | Add team member |
| `podx encrypt-all` | Encrypt all secrets, delete originals |
| `podx decrypt-all` | Decrypt all secrets |
| `podx status` | Show project info |

### File Commands

| Command | Description |
|---------|-------------|
| `podx encrypt -a ALGO -i FILE -o OUT` | Encrypt single file |
| `podx decrypt -i FILE -o OUT` | Decrypt single file |
| `podx env encrypt -i .env -o .env.enc` | Encrypt .env (format-preserving) |
| `podx env decrypt -i .env.enc -o .env` | Decrypt .env |

### Key Management

| Command | Description |
|---------|-------------|
| `podx keygen -t age` | Generate Age key pair |
| `podx keygen -t gpg -n NAME -e EMAIL` | Generate GPG key |

### Other

| Command | Description |
|---------|-------------|
| `podx update` | Self-update to latest version |
| `podx version` | Show version and platform |
| `podx help` | Show help |

### Security Commands

| Command | Description |
|---------|-------------|
| `podx check` | Run all security checks |
| `podx check --fix` | Auto-fix gitignore issues |
| `podx check --pre-commit` | Silent mode for git hooks |
| `podx hook install` | Install pre-commit hook |
| `podx hook uninstall` | Remove pre-commit hook |
| `podx hook status` | Check if hook is installed |

---

## Project Workflow

### 1. Initialize Project

```bash
podx init
```

Creates `.podx.yaml`:

```yaml
version: 1
backend: age
recipients:
  - name: Owner
    key: age1xc2ttxdm60507...
secrets:
  - .env
```

### 2. Add Team Members

```bash
# Team member generates their key
podx keygen -t age
# Output: age1abc123...

# Project owner adds them
podx add-recipient -n "Alice" -k age1abc123...
```

### 3. Encrypt Secrets

```bash
podx encrypt-all
```

**Before:**
```
.env          # Plain text (will be deleted)
```

**After:**
```
.env.podx     # Encrypted (commit this)
```

**Format-preserving .env.podx:**
```env
API_KEY=ENC[age:YWdlLWVuY3J5cH...]
DB_PASS=ENC[age:YWdlLWVuY3J5cH...]
# This comment is preserved
DEBUG=ENC[age:YWdlLWVuY3J5cH...]
```

### 4. Commit to Git

```bash
git add .podx.yaml .env.podx
git commit -m "Add encrypted secrets"
git push
```

### 5. Team Member Decrypts

```bash
git clone <repo>
cd <repo>
podx decrypt-all
# .env is now restored
```

---

## Encryption Algorithms

### Symmetric (Password-based)

| Algorithm | Description |
|-----------|-------------|
| `aes-gcm` | AES-256-GCM (default, hardware accelerated) |
| `chacha20` | ChaCha20-Poly1305 (ARM-friendly) |

```bash
podx encrypt -a aes-gcm -i file.txt -o file.enc
podx encrypt -a chacha20 -i file.txt -o file.enc
```

### Asymmetric (Key-based)

| Backend | Description |
|---------|-------------|
| `age` | Modern X25519 encryption |
| `gpg` | Traditional GPG/PGP |

```bash
# Age (used by encrypt-all)
podx keygen -t age

# GPG
podx keygen -t gpg -n "Name" -e "email@example.com"
```

---

## Configuration

### Key Storage

```
~/.config/podx/
├── age-keys.txt           # Private keys
└── age-recipients/
    └── default.txt        # Public key
```

### Project Config (.podx.yaml)

```yaml
version: 1
backend: age

# Who can decrypt
recipients:
  - name: Owner
    key: age1xc2ttxdm60507q6wqqmsk695arqxn4x3zpq43dkstwxhaxkccaxspz47na
  - name: Alice
    key: age1abc123...

# Files to encrypt
secrets:
  - .env
  - .env.production
  - config/secrets.yaml
```

---

## Self-Update

```bash
# Check version
podx version

# Update to latest
podx update
```

---

## Security

### Security Checks

PODX includes comprehensive security scanning:

| Check | Description |
|-------|-------------|
| **Encryption Status** | Verifies all secrets are encrypted |
| **Gitignore Validation** | Ensures sensitive files are excluded |
| **Pattern Scan** | Detects hardcoded secrets (API keys, passwords) |

#### Pattern Detection

Scans for common secret patterns:
- AWS Access Keys (`AKIA...`)
- Private Keys (`-----BEGIN...PRIVATE KEY-----`)
- API Keys (`api[_-]?key`, `apikey`)
- Passwords in URLs (`password=...`)
- Connection Strings (`mongodb://...`, `postgres://...`)
- JWT Tokens (`eyJ...`)

#### Pre-Commit Hook

Automatically run security checks before each commit:

```bash
# Install hook
podx hook install

# Check status
podx hook status

# Uninstall
podx hook uninstall
```

When the hook detects issues, the commit is blocked with details about what needs to be fixed.

### Cryptography

- **Argon2id** for password-based key derivation
- **AEAD** encryption (authenticated encryption)
- **Random nonces** for each encryption
- **No key in ciphertext** — keys stored separately

---

## Releasing (Maintainers)

```bash
# Tag and push
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions builds all platforms automatically
```

Or use manual workflow dispatch from GitHub Actions tab.

---

## License

MIT © [dwirx](https://github.com/dwirx)
