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
  <a href="#encryption-methods">Encryption</a> •
  <a href="#tui-terminal-user-interface">TUI</a> •
  <a href="#performance">Performance</a>
</p>

---

## Features

### Encryption
- **Age Encryption** — Modern X25519 asymmetric encryption (default)
- **Native Go GPG** — Pure Go OpenPGP implementation (no external GPG binary required)
- **AES-256-GCM** — Hardware-accelerated symmetric encryption
- **ChaCha20-Poly1305** — High-performance symmetric encryption (fastest)
- **XChaCha20-Poly1305** — Extended nonce for safer random nonces
- **Paranoid Mode** — XChaCha20 + Serpent cascade encryption
- **Format-Preserving** — `.env` files stay readable (`KEY=ENC[...]`)

### Team Collaboration
- **Multi-Recipient** — Encrypt to multiple team members
- **Project Workspaces** — Per-project `.podx.yaml` configuration
- **Git-Friendly** — Commit encrypted files safely
- **Shamir Secret Sharing** — Split secrets into shares (k-of-n)
- **QR Code Export** — Share secrets via QR codes

### Security
- **Pre-commit Hook** — Block commits with exposed secrets
- **Secret Scanning** — Detect API keys, passwords, tokens
- **Memory Protection** — Secure wiping of sensitive data
- **Compression** — Automatic zstd compression before encryption

### User Experience
- **Interactive TUI** — Beautiful terminal interface with file browser
- **Cross-Platform** — Linux, macOS, Windows (AMD64 & ARM64)
- **Self-Update** — Built-in update with rollback capability

---

## Installation

### Linux / macOS

```bash
# Install latest stable version
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
```

### Windows (PowerShell)

```powershell
# Install latest stable version
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex
```

### Build from Source

```bash
git clone https://github.com/dwirx/podx
cd podx
go build -o podx .
```

---

## Quick Start

### 1. Generate Key
```bash
# Generate Age key (recommended)
podx keygen -t age

# Or generate GPG key (native implementation)
podx keygen -t gpg -n "Alice" -e "alice@example.com"
```

### 2. Initialize Project
```bash
cd your-project
podx init
```

### 3. Create & Encrypt Secrets
```bash
echo "API_KEY=secret123" > .env
podx encrypt-all
```

### 4. Commit to Git
```bash
git add .podx.yaml .env.podx
git commit -m "Add encrypted secrets"
```

---

## Encryption Methods

PODX supports 5 different encryption methods, optimized for different use cases.

### Symmetric Encryption
Used for file encryption with passwords.

1. **ChaCha20-Poly1305** (Recommended)
   - **Performance:** ~1,000,000 ops/sec
   - **Memory:** 64 B/op
   - **Best for:** Most use cases, systems without AES hardware acceleration
   - **Security:** 256-bit key, 96-bit nonce, Poly1305 MAC

2. **XChaCha20-Poly1305**
   - **Performance:** ~936,000 ops/sec
   - **Memory:** 80 B/op
   - **Best for:** Streaming large files, high-volume encryption
   - **Security:** Extended 192-bit nonce eliminates random nonce collision risk

3. **AES-256-GCM**
   - **Performance:** ~683,000 ops/sec
   - **Memory:** 1,312 B/op
   - **Best for:** Systems with AES-NI instructions, compliance requirements
   - **Security:** NIST standard, authenticated encryption

### Asymmetric Encryption
Used for sharing secrets with team members.

#### Age X25519
- **Status:** Default & Recommended
- **Algorithm:** X25519 (Curve25519) + ChaCha20-Poly1305
- **Performance:** ~3,384 ops/sec
- **Key Format:** `age1...` (public), `AGE-SECRET-KEY-1...` (private)
- **Pros:** Modern, simple, small keys, fast key generation

#### GPG/PGP
- **Implementation:** **Native Go GPG** (via `gopenpgp`)
- **Algorithm:** RSA 4096-bit (default)
- **Performance:** ~614 ops/sec
- **Pros:** Wide compatibility, password-protected keys, enterprise standard
- **Features:**
  - **No external GPG binary required** - Uses pure Go implementation
  - **Cross-platform** - Works identically on all OSs
  - **OpenPGP Compatible** - RFC 4880 compliant
  - **Auto-fallback** - Gracefully handles shell GPG if needed for specific operations

---

## Performance

Benchmark results from `docs/encryption-test-report.md` (2026-01-15):

### Encryption Speed (ops/sec)
| Algorithm | Speed | Type | Note |
|-----------|-------|------|------|
| **ChaCha20-Poly1305** | 1,000,000 | Symmetric | ⚡ FASTEST |
| **XChaCha20-Poly1305** | 936,196 | Symmetric | Extended Nonce |
| **AES-256-GCM** | 683,055 | Symmetric | Hardware Accel |
| **Age X25519** | 3,384 | Asymmetric | Modern |
| **GPG Native** | 614 | Asymmetric | RSA 4096-bit |

### Memory Usage (bytes/op)
| Algorithm | Memory | Efficiency |
|-----------|--------|------------|
| **ChaCha20-Poly1305** | 64 B | ⭐ MOST EFFICIENT |
| **XChaCha20-Poly1305** | 80 B | Excellent |
| **AES-256-GCM** | 1,312 B | Good |
| **Age X25519** | 152 KB | Standard |
| **GPG Native** | 354 KB | RSA Overhead |

---

## TUI (Terminal User Interface)

Launch the interactive interface with `podx`.

### Tabs
- **Dashboard** `[*]` - Project overview, quick actions, key management
- **Commands** `[>]` - Run any CLI command interactively
- **Security** `[#]` - Run security checks (patterns, gitignore)
- **Files** `[@]` - Browse, encrypt, and decrypt files

### Encryption Selection
When encrypting a file (`e` key), you can now select:
1. **Algorithm:** AES-GCM, ChaCha20, XChaCha20, or Paranoid
2. **Method:** Password (Symmetric) or Key-based (Asymmetric)

### Key Manager
Access via Dashboard or `[G]` in Commands tab:
- Generate Age or GPG keys
- Import existing keys
- Set default keys
- Manage key names for easier identification

---

## Commands

### Project & Files
```bash
# Encrypt/Decrypt project
podx encrypt-all
podx decrypt-all

# Encrypt single file (XChaCha20)
podx encrypt -c xchacha20 -i secret.txt -o secret.enc

# Encrypt .env (Format Preserving)
podx env encrypt -i .env -o .env.podx
```

### Key Management
```bash
# Generate Age key
podx keygen -t age

# Generate GPG key (Native)
podx keygen -t gpg -n "User" -e "user@example.com"

# List keys (TUI)
podx keys
```

### Security & Shamir
```bash
# Check for secrets
podx check

# Split secret (3-of-5)
podx shamir split -i key.txt -t 3 -n 5 -qr

# Combine shares
podx shamir combine -d ./shares -o key.txt
```

---

## Security

### Key Storage
Keys are stored in `~/.config/podx/`:
- `age-keys.txt`: Age private keys
- `gpg-keys/`: GPG keyrings (native implementation)
- `age-recipients/`: Public keys

### Security Features
- **Authenticated Encryption:** All methods use AEAD (GCM, Poly1305)
- **Argon2id KDF:** Password hashing with 64MB memory cost
- **Secure Memory:** Sensitive data is wiped from memory after use
- **Pre-commit Hook:** Prevents committing plaintext secrets

### Best Practices
1. **Never commit .env files** - Use `.env.podx`
2. **Use Age for new projects** - It's faster and simpler
3. **Use GPG for compatibility** - If you need to interface with existing PGP systems
4. **Backup your keys** - If you lose your private key, data is unrecoverable

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

---

## Contributing

Contributions are welcome!
1. Fork the repo
2. Create feature branch
3. Commit changes
4. Push and create PR

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [age](https://github.com/FiloSottile/age)
- [gopenpgp](https://github.com/ProtonMail/gopenpgp)
- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [cobra](https://github.com/spf13/cobra)

<p align="center">
  Made with <strong>Go</strong> by <a href="https://github.com/dwirx">dwirx</a>
</p>
