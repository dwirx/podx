# PODX

<p align="center">
  <img src="https://img.shields.io/github/v/release/dwirx/podx?style=flat-square" alt="Release">
  <img src="https://img.shields.io/github/actions/workflow/status/dwirx/podx/ci.yml?branch=main&style=flat-square" alt="CI">
  <img src="https://img.shields.io/github/license/dwirx/podx?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue?style=flat-square" alt="Platform">
</p>

<p align="center">
  <b>Secure Encryption CLI for Teams</b><br>
  Encrypt secrets, share with team members, commit safely to Git
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#interactive-tui">TUI</a> •
  <a href="#commands">Commands</a> •
  <a href="#security">Security</a>
</p>

---

## Features

### Encryption
- **Age & GPG Encryption** — Modern X25519 asymmetric or traditional GPG
- **Password Encryption** — AES-256-GCM with Argon2id key derivation
- **XChaCha20-Poly1305** — Extended nonce ChaCha20 for safer random nonces
- **Paranoid Mode** — XChaCha20 + Serpent cascade encryption with HMAC-SHA3
- **Format-Preserving** — `.env` files stay readable (`KEY=ENC[...]`)
- **Streaming Encryption** — Automatic chunked encryption for files >100MB

### Team Collaboration
- **Multi-Recipient** — Encrypt to multiple team members
- **Project Workspaces** — Per-project `.podx.yaml` configuration
- **Git-Friendly** — Commit encrypted files safely

### Security
- **Pre-commit Hook** — Block commits with exposed secrets
- **Secret Scanning** — Detect API keys, passwords, tokens
- **Gitignore Validation** — Ensure sensitive files are excluded

### User Experience
- **Interactive TUI** — Beautiful terminal interface with file browser
- **Self-Update** — Built-in update with rollback capability
- **Cross-Platform** — Linux, macOS, Windows (AMD64 & ARM64)

---

## Installation

### Linux / macOS

```bash
# Install latest stable version
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash

# Install beta version
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash -s -- --beta

# Install specific version
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash -s -- --version v1.0.2

# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash -s -- --dir ~/.local/bin
```

### Windows (PowerShell as Administrator)

```powershell
# Install latest stable version
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex
```

### Build from Source

```bash
# Clone repository
git clone https://github.com/dwirx/podx
cd podx

# Build
go build -o podx .

# Install
sudo mv podx /usr/local/bin/

# Verify
podx version
```

### Supported Platforms

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | AMD64 | `podx-linux-amd64` |
| Linux | ARM64 | `podx-linux-arm64` |
| Linux | ARM (32-bit) | `podx-linux-arm` |
| macOS | Intel | `podx-darwin-amd64` |
| macOS | Apple Silicon | `podx-darwin-arm64` |
| Windows | AMD64 | `podx-windows-amd64.exe` |
| Windows | ARM64 | `podx-windows-arm64.exe` |

### Uninstall

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/uninstall.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/uninstall.ps1 | iex
```

---

## Quick Start

### 1. Generate Encryption Key

```bash
podx keygen -t age
```

Output:
```
Age Key Pair Generated

Public Key:  age1xc2ttxdm60507q6wqqmsk695arqxn4x3zpq43dkstwxhaxkccaxspz47na
Private Key: AGE-SECRET-KEY-1QQQQQQQQ...

Keys saved to:
  ~/.config/podx/age-keys.txt
  ~/.config/podx/age-recipients/default.txt
```

### 2. Initialize Project

```bash
cd your-project
podx init
```

Creates `.podx.yaml`:
```yaml
version: 1
backend: age
recipients:
  - name: Owner
    key: age1xc2ttxdm60507q6wqqmsk695arqxn4x3zpq43dkstwxhaxkccaxspz47na
secrets:
  - .env
```

### 3. Create Secrets

```bash
echo "API_KEY=secret123" > .env
echo "DB_PASSWORD=mypassword" >> .env
```

### 4. Encrypt Secrets

```bash
podx encrypt-all
```

**Before:**
```
.env              # Plain text (will be deleted)
```

**After:**
```
.env.podx         # Encrypted (safe to commit)
```

**Encrypted .env.podx format:**
```env
API_KEY=ENC[age:YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+...]
DB_PASSWORD=ENC[age:YWdlLWVuY3J5cHRpb24ub3JnL3YxCi...]
```

### 5. Commit to Git

```bash
git add .podx.yaml .env.podx
git commit -m "Add encrypted secrets"
git push
```

### 6. Team Member Decrypts

```bash
git clone <repo>
cd <repo>
podx decrypt-all
# .env is restored!
```

---

## Interactive TUI

Launch the interactive terminal interface:

```bash
podx
```

### Tabs

| Tab | Icon | Description |
|-----|------|-------------|
| **Dashboard** | `[*]` | Project overview, encryption status, quick actions |
| **Commands** | `[>]` | Interactive command menu |
| **Security** | `[#]` | Live security check results |
| **Files** | `[@]` | File browser with encryption capabilities |

### Global Navigation

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1` `2` `3` `4` | Jump to tab |
| `r` | Refresh data |
| `?` | Toggle help overlay |
| `q` / `Esc` | Quit |

### Dashboard Tab

- **Project Info** — Path, backend, recipients, secret patterns
- **Security Status** — Encryption check, gitignore validation, hook status
- **Quick Actions** — Encrypt All, Decrypt All, Run Check, Install Hook
- **Update Notification** — Shows when new version available

### Files Tab

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `l` / `→` / `Enter` | Enter directory / Select |
| `h` / `←` | Go to parent directory |
| `Space` | Toggle file selection |
| `a` | Select/deselect all |
| `e` | Encrypt selected files |
| `d` | Decrypt selected files |
| `p` | Toggle preview panel |
| `g` | Go to path |
| `/` | Filter files |

### Encryption Dialog

When pressing `e` to encrypt, choose between:

#### Password Encryption (AES-GCM)
- AES-256-GCM symmetric encryption
- Argon2id key derivation (memory-hard)
- Requires password to decrypt
- Best for: Personal files, password-based sharing

#### Age Key Encryption
- X25519 asymmetric encryption
- Uses recipients from `.podx.yaml`
- No password needed
- Best for: Team collaboration via Git

---

## Commands

### Project Commands

```bash
# Initialize project
podx init

# Add team member
podx add-recipient -n "Alice" -k age1abc123...

# Encrypt all secrets (deletes originals)
podx encrypt-all

# Decrypt all secrets
podx decrypt-all

# Show project status
podx status
```

### File Commands

```bash
# Encrypt single file with password (normal mode, AES-GCM)
podx encrypt -i secret.txt -o secret.enc

# Encrypt with XChaCha20
podx encrypt -c xchacha20 -i secret.txt -o secret.enc

# Encrypt with paranoid mode (XChaCha20 + Serpent cascade)
podx encrypt -m paranoid -i secret.txt -o secret.enc

# Encrypt with specific cipher
podx encrypt -m normal -c xchacha20 -i secret.txt -o secret.enc

# Decrypt file (auto-detects encryption format)
podx decrypt -i secret.enc -o secret.txt

# Encrypt .env file (format-preserving)
podx env encrypt -i .env -o .env.podx

# Decrypt .env file
podx env decrypt -i .env.podx -o .env
```

### Encryption Modes

| Mode | Cipher | MAC | KDF | Use Case |
|------|--------|-----|-----|----------|
| `normal` | AES-GCM or XChaCha20 | BLAKE2b | Argon2id (4 passes, 256MB) | Standard security |
| `paranoid` | XChaCha20 + Serpent | HMAC-SHA3-512 | Argon2id (8 passes, 512MB) | Maximum security |

### Key Management

```bash
# Generate Age key pair
podx keygen -t age

# Generate GPG key pair
podx keygen -t gpg -n "Your Name" -e "email@example.com"
```

### Security Commands

```bash
# Run all security checks
podx check

# Auto-fix gitignore issues
podx check --fix

# Silent mode for CI/pre-commit
podx check --pre-commit

# Install pre-commit hook
podx hook install

# Check hook status
podx hook status

# Uninstall hook
podx hook uninstall
```

### Update Commands

```bash
# Check current version
podx version

# Update to latest stable version
podx update

# Update to beta version
podx update --beta

# Rollback to previous version (after update)
podx rollback
```

---

## Security

### Encryption Algorithms

#### Symmetric (Password-based)

| Algorithm | Key Size | Description |
|-----------|----------|-------------|
| `aes-gcm` | 256-bit | AES-GCM (default, hardware-accelerated on modern CPUs) |
| `chacha20` | 256-bit | ChaCha20-Poly1305 (faster on ARM, no AES-NI) |

#### Asymmetric (Key-based)

| Backend | Algorithm | Description |
|---------|-----------|-------------|
| `age` | X25519 | Modern, simple, secure. Recommended. |
| `gpg` | RSA/EdDSA | Traditional PGP. Wide compatibility. |

### Key Derivation

Password-based encryption uses **Argon2id** with secure parameters:
- Memory: 64 MB
- Iterations: 3
- Parallelism: 4
- Salt: 16 bytes (random)

### Security Checks

PODX scans for common secret patterns:

| Pattern | Examples |
|---------|----------|
| AWS Keys | `AKIA...`, `aws_secret_access_key` |
| API Keys | `api_key=...`, `apikey=...` |
| Private Keys | `-----BEGIN RSA PRIVATE KEY-----` |
| Passwords | `password=...`, `passwd=...` |
| Tokens | `token=...`, `bearer ...` |
| Connection Strings | `mongodb://...`, `postgres://...` |
| JWT | `eyJ...` (base64 JSON) |

### Pre-commit Hook

The pre-commit hook prevents committing exposed secrets:

```bash
# Install
podx hook install

# What it checks:
# 1. All secrets in .podx.yaml are encrypted
# 2. .gitignore includes sensitive patterns
# 3. No hardcoded secrets in staged files
```

When issues are detected, the commit is blocked with actionable fixes.

### Best Practices

1. **Never commit `.env`** — Only commit `.env.podx`
2. **Use Age over GPG** — Simpler, modern, fewer footguns
3. **Install pre-commit hook** — Catch mistakes before push
4. **Rotate keys periodically** — Generate new keys, re-encrypt
5. **Audit recipients** — Remove ex-team members promptly

---

## Configuration

### Key Storage

```
~/.config/podx/
├── age-keys.txt           # Private keys (one per line)
└── age-recipients/
    └── default.txt        # Your public key
```

### Project Config (.podx.yaml)

```yaml
version: 1
backend: age  # or "gpg"

# Team members who can decrypt
recipients:
  - name: Owner
    key: age1xc2ttxdm60507q6wqqmsk695arqxn4x3zpq43dkstwxhaxkccaxspz47na
  - name: Alice
    key: age1abc123...
  - name: Bob
    key: age1def456...

# Files to encrypt (glob patterns supported)
secrets:
  - .env
  - .env.production
  - .env.staging
  - config/secrets.yaml
  - config/*.key
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PODX_CONFIG_DIR` | Config directory | `~/.config/podx` |
| `PODX_KEY_FILE` | Age private key file | `~/.config/podx/age-keys.txt` |

---

## Self-Update

PODX includes a built-in updater with automatic rollback:

```bash
# Check current version and available updates
podx version

# Update to latest stable
podx update

# Update to beta (pre-release)
podx update --beta

# If update causes issues, rollback
podx rollback
```

### Update Process

1. **Check** — Fetches latest release from GitHub
2. **Download** — Downloads binary with progress indicator
3. **Verify** — Tests downloaded binary works
4. **Backup** — Saves current binary as `.bak`
5. **Install** — Replaces current with new binary
6. **Cleanup** — Removes backup on success

If installation fails, the backup is automatically restored.

### Permissions

On Linux/macOS, if write permission is denied, PODX automatically requests elevated privileges via `sudo`.

On Windows, run PowerShell as Administrator.

---

## Team Workflow

### Initial Setup (Project Owner)

```bash
# 1. Generate your key
podx keygen -t age

# 2. Initialize project
cd project
podx init

# 3. Create secrets
echo "API_KEY=secret" > .env

# 4. Encrypt and commit
podx encrypt-all
git add .podx.yaml .env.podx
git commit -m "Add encrypted secrets"
git push
```

### Adding Team Members

```bash
# Team member generates their key
podx keygen -t age
# Shares their PUBLIC key: age1abc123...

# Project owner adds them
podx add-recipient -n "Alice" -k age1abc123...

# Re-encrypt secrets for new recipient
podx decrypt-all
podx encrypt-all
git add .podx.yaml .env.podx
git commit -m "Add Alice to recipients"
git push
```

### Team Member Decrypts

```bash
git clone <repo>
cd <repo>
podx decrypt-all
# .env is now available locally
```

### Removing Team Members

Edit `.podx.yaml` and remove the recipient, then re-encrypt:

```bash
# Edit .podx.yaml to remove the recipient
vim .podx.yaml

# Re-encrypt with remaining recipients
podx decrypt-all
podx encrypt-all
git add .podx.yaml .env.podx
git commit -m "Remove ex-team-member from recipients"
git push
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install PODX
        run: curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash

      - name: Setup Age Key
        run: |
          mkdir -p ~/.config/podx
          echo "${{ secrets.AGE_SECRET_KEY }}" > ~/.config/podx/age-keys.txt

      - name: Decrypt Secrets
        run: podx decrypt-all

      - name: Deploy
        run: ./deploy.sh
```

### GitLab CI

```yaml
deploy:
  image: golang:1.22
  before_script:
    - curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
    - mkdir -p ~/.config/podx
    - echo "$AGE_SECRET_KEY" > ~/.config/podx/age-keys.txt
  script:
    - podx decrypt-all
    - ./deploy.sh
```

---

## Troubleshooting

### "No project found"

```bash
# Initialize project first
podx init
```

### "No recipients configured"

```bash
# Generate key first
podx keygen -t age

# Then reinitialize
podx init
```

### "Failed to decrypt: no identity matched"

Your private key doesn't match any recipient. Either:
1. You're not a recipient — ask project owner to add you
2. Wrong key — check `~/.config/podx/age-keys.txt`

### "Permission denied" during update

```bash
# Linux/macOS: Run with sudo
sudo podx update

# Windows: Run PowerShell as Administrator
```

### Update failed, need to rollback

```bash
podx rollback
```

---

## Development

### Requirements

- Go 1.22+
- Git

### Build

```bash
git clone https://github.com/dwirx/podx
cd podx
go build -o podx .
```

### Test

```bash
go test ./...
```

### Release

Releases are automated via GitHub Actions:

```bash
# Create release tag
git tag v1.0.0
git push origin v1.0.0
```

Beta releases are triggered by pushing to the `testing` branch.

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open a Pull Request

### Code Style

- Run `gofmt` before committing
- Run `go vet ./...` to check for issues
- Add tests for new features

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [age](https://github.com/FiloSottile/age) — Modern encryption tool
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions

---

<p align="center">
  Made with <strong>Go</strong> by <a href="https://github.com/dwirx">dwirx</a>
</p>
