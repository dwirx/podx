# PODX 🔐

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
  <a href="#-features">Features</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-encryption-methods">Encryption</a> •
  <a href="#-tui-terminal-user-interface">TUI</a> •
  <a href="#-architecture">Architecture</a> •
  <a href="#-performance">Performance</a>
</p>

---

## ✨ Features

### 🔐 Encryption Methods
- **Age X25519** — Modern asymmetric encryption (recommended for teams)
- **Native Go GPG** — Pure Go OpenPGP (no external GPG binary required!)
- **ChaCha20-Poly1305** — Fastest symmetric encryption (1M+ ops/sec)
- **XChaCha20-Poly1305** — Extended nonce for high-volume encryption
- **AES-256-GCM** — Hardware-accelerated FIPS-compliant encryption
- **Format-Preserving** — `.env` files stay readable (`KEY=ENC[...]`)

### 👥 Team Collaboration
- **Multi-Recipient** — Encrypt to multiple team members in one operation
- **Project Workspaces** — Per-project `.podx.yaml` configuration
- **Git-Friendly** — Commit encrypted files safely with auto-gitignore
- **Recipient Management** — Add/remove team members easily

### 🛡️ Security Features
- **Pre-commit Hook** — Block commits with exposed secrets automatically
- **Secret Scanning** — Detect AWS keys, API tokens, passwords in code
- **Memory Protection** — Secure wiping of sensitive data from memory
- **Compression** — Automatic zstd compression before encryption

### 🎨 User Experience
- **Interactive TUI** — Beautiful terminal interface with:
  - 📊 Dashboard with project status
  - 📁 File browser with encrypt/decrypt actions
  - 🔑 Key manager with import/export
  - 🛡️ Live security checks
  - ⚡ Quick actions menu
- **Cross-Platform** — Linux, macOS, Windows (AMD64 & ARM64)
- **Self-Update** — Built-in update mechanism with rollback

---

## 📦 Installation

### Linux / macOS

```bash
# Install latest stable version
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash

# Verify installation
podx version
```

### Windows (PowerShell)

```powershell
# Install latest stable version
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex

# Verify installation
podx version
```

### Package Managers

```bash
# Homebrew (macOS/Linux)
brew install dwirx/tap/podx

# Scoop (Windows)
scoop bucket add dwirx https://github.com/dwirx/scoop-bucket
scoop install podx

# Nix
nix profile install github:dwirx/podx
```

### Build from Source

```bash
# Requires Go 1.21+
git clone https://github.com/dwirx/podx
cd podx
make build
sudo make install

# Or use go install
go install github.com/dwirx/podx@latest
```

---

## 🚀 Quick Start

### 1️⃣ Generate Key

```bash
# Generate Age key (recommended for teams)
podx keygen -t age
# Public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
# Private key saved to: ~/.config/podx/age-keys.txt

# Or generate GPG key (native implementation, no GPG binary needed)
podx keygen -t gpg -n "Alice" -e "alice@example.com"
# Key ID: A1B2C3D4E5F6G7H8
# Keys saved to: ~/.config/podx/gpg-keys/
```

### 2️⃣ Initialize Project

```bash
cd your-project
podx init
# Created .podx.yaml with your public key
# Added .gitignore entries for secret patterns
```

### 3️⃣ Encrypt Secrets

```bash
# Create .env file
cat > .env << EOF
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk_live_abc123xyz789
AWS_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
EOF

# Encrypt with Age (team sharing)
podx encrypt .env
# ✓ Encrypted: .env → .env.podx
# ✓ Deleted original: .env

# Or encrypt with password (ChaCha20, fastest)
podx encrypt secrets.txt -m chacha20
# Password: ••••••••
# ✓ Encrypted: secrets.txt → secrets.txt.podx

# Or encrypt with GPG (OpenPGP compatible)
podx encrypt config.yaml -m gpg -r alice@example.com
# ✓ Encrypted for: alice@example.com
```

### 4️⃣ Share with Team

```bash
# Add team member to project
podx add-recipient Bob age1yhm4gctwfmrpz87tdslm550wrx6m79y9f2hdzt0lne5nxx6gdtxqm0u3dv

# Encrypt all secrets for the team
podx encrypt-all
# ✓ Encrypted 3 files
# - .env → .env.podx
# - config/api-keys.json → config/api-keys.json.podx
# - secrets/db.conf → secrets/db.conf.podx

# Commit encrypted files to Git
git add .
git commit -m "feat: add encrypted configuration"
git push
```

### 5️⃣ Team Member Decrypts

```bash
# Clone repository
git clone https://github.com/your-org/your-project
cd your-project

# Decrypt all secrets (uses local private key automatically)
podx decrypt-all
# ✓ Decrypted 3 files
# - .env.podx → .env
# - config/api-keys.json.podx → config/api-keys.json
# - secrets/db.conf.podx → secrets/db.conf
```

---

## 🔐 Encryption Methods

PODX supports **5 encryption algorithms** with comprehensive test coverage (314+ tests, all passing):

### Symmetric Encryption (Password-Based)

#### 1. ChaCha20-Poly1305 ⚡ FASTEST & RECOMMENDED

```bash
podx encrypt file.txt -m chacha20
```

**Performance:**
- Encryption: **1,000,000 ops/sec** (1.15 μs/op)
- Decryption: **4,064,662 ops/sec** (0.30 μs/op)
- Memory: **64 B/op** (most efficient)

**When to Use:**
- Default choice for password-based encryption
- Best performance on all platforms (no hardware dependency)
- Excellent for mobile/ARM devices
- High-volume encryption scenarios

**Security:**
- AEAD cipher (authenticated encryption)
- 12-byte random nonce
- Poly1305 MAC for tamper detection
- IETF RFC 8439 standard

---

#### 2. XChaCha20-Poly1305 (Extended Nonce)

```bash
podx encrypt file.txt -m xchacha20
```

**Performance:**
- Encryption: **936,196 ops/sec** (1.35 μs/op)
- Decryption: **2,420,415 ops/sec** (0.49 μs/op)
- Memory: **80 B/op**

**When to Use:**
- High-volume encryption (millions of files)
- Streaming data encryption
- When you need safer random nonce generation
- Long-running processes that encrypt frequently

**Security:**
- Extended **24-byte nonce** (vs 12-byte in ChaCha20)
- Better nonce collision resistance
- Supports deterministic encryption with custom nonce
- Same authentication as ChaCha20

---

#### 3. AES-256-GCM (FIPS Compliant)

```bash
podx encrypt file.txt -m aes-gcm
```

**Performance:**
- Encryption: **683,055 ops/sec** (1.49 μs/op)
- Decryption: **1,564,368 ops/sec** (0.83 μs/op)
- Memory: **1,312 B/op**

**When to Use:**
- Compliance requirements (FIPS 140-2)
- Hardware AES-NI acceleration available (Intel/AMD CPUs)
- Enterprise environments with strict standards
- Integration with legacy systems

**Security:**
- Industry-standard AEAD cipher
- Hardware-accelerated on modern CPUs
- NIST approved algorithm
- 12-byte random nonce

---

### Asymmetric Encryption (Public-Key)

#### 4. Age X25519 🚀 RECOMMENDED FOR TEAMS

```bash
# Generate key
podx keygen -t age

# Encrypt for recipient
podx encrypt file.txt -r age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

# Encrypt for multiple recipients
podx encrypt file.txt -r alice.txt -r bob.txt -r charlie.txt
```

**Performance:**
- Encryption: **3,384 ops/sec** (344 μs/op)
- Decryption: **3,450 ops/sec** (346 μs/op)
- Memory: **152 KB/op**
- Key Generation: **8,391 ops/sec** (149 μs/op)

**When to Use:**
- Team collaboration (multiple recipients)
- Modern applications (designed for 21st century)
- Simple key management (no complex GPG/PGP)
- Cross-platform key sharing

**Key Features:**
- ✨ Simple key format (`age1...` public, `AGE-SECRET-KEY-...` private)
- 👥 Multiple recipient support (encrypt once for many)
- 🔐 X25519 Curve25519 ECDH + ChaCha20-Poly1305 AEAD
- 📱 Small keys (easy to share via QR code)
- 🚀 Fast performance for asymmetric crypto

**Security:**
- Modern cryptography by Filippo Valsorda (Google)
- Curve25519 key exchange (ECDH)
- ChaCha20-Poly1305 for data encryption
- No complex trust model (simple recipient list)

---

#### 5. GPG/PGP 🔑 ENTERPRISE STANDARD (Native Go)

```bash
# Generate GPG key (NO external GPG binary required!)
podx keygen -t gpg -n "Alice" -e "alice@example.com"

# Encrypt with armored public key (native)
podx encrypt file.txt --gpg-key "$(cat alice_public.asc)"

# Encrypt for email/Key ID (uses GPG keyring if available)
podx encrypt file.txt -r alice@example.com

# Encrypt for multiple recipients (native)
podx encrypt file.txt --gpg-key alice.asc --gpg-key bob.asc

# Decrypt (auto-detects private key)
podx decrypt file.txt.podx
```

**Performance:**
- Encryption: **614 ops/sec** (1.99 ms/op)
- Decryption: **30 ops/sec** (33.4 ms/op)
- Memory: **354 KB/op**
- Key Generation: **0.45 ops/sec** (2.2 seconds/op)

**When to Use:**
- Enterprise environments with existing PGP infrastructure
- OpenPGP compatibility required (Thunderbird, ProtonMail, etc.)
- Interoperability with GPG/PGP ecosystem
- Password-protected private keys needed
- Long-term archival with industry standard format

**🌟 Native Go Implementation Benefits:**

✨ **No External Dependencies**
- Pure Go implementation via `gopenpgp/v2`
- Works without GPG binary installed
- Cross-platform (Linux, macOS, Windows)
- No C dependencies, no CGO

🔐 **Full Feature Support**
- RSA 4096-bit key generation
- Multiple recipient encryption
- Password-protected private keys
- Armored ASCII format (copy-paste friendly)
- OpenPGP RFC 4880 compliant

🚀 **Better Developer Experience**
- Consistent API across platforms
- No shell command execution
- Better error messages
- Memory-safe (no C code)

**Key Features:**
- 🔗 OpenPGP compatible (works with GPG, Thunderbird, ProtonMail)
- 🔐 RSA 4096-bit keys (industry standard)
- 👥 Multiple recipient support
- 🔒 Password-protected keys (native support)
- 📦 Pure Go (no external binaries)
- 🌍 Cross-platform (Linux, macOS, Windows)

**Security:**
- RSA 4096-bit encryption
- OpenPGP RFC 4880 compliant
- Password protection with bcrypt
- Armored output (ASCII-safe)
- Memory cleanup after operations

---

### 📊 Performance Comparison

#### Encryption Speed Ranking

| Rank | Algorithm | Speed (ops/sec) | Use Case |
|------|-----------|-----------------|----------|
| 🥇 | **ChaCha20-Poly1305** | 1,000,000 | ⚡ Default symmetric encryption |
| 🥈 | XChaCha20-Poly1305 | 936,196 | High-volume/streaming |
| 🥉 | AES-256-GCM | 683,055 | FIPS compliance |
| 4 | Age X25519 | 3,384 | 🚀 Team sharing |
| 5 | GPG Native | 614 | Enterprise/OpenPGP |

#### Decryption Speed Ranking

| Rank | Algorithm | Speed (ops/sec) | Memory (B/op) |
|------|-----------|-----------------|---------------|
| 🥇 | **ChaCha20-Poly1305** | 4,064,662 | 64 ⭐ |
| 🥈 | XChaCha20-Poly1305 | 2,420,415 | 80 |
| 🥉 | AES-256-GCM | 1,564,368 | 1,312 |
| 4 | Age X25519 | 3,450 | 152,071 |
| 5 | GPG Native | 30 | 354,874 |

#### Memory Efficiency

| Algorithm | Encryption | Decryption | Note |
|-----------|------------|------------|------|
| ChaCha20-Poly1305 | 112 B | **64 B** | ⭐ Most efficient |
| XChaCha20-Poly1305 | 136 B | 80 B | Excellent |
| AES-256-GCM | 1,360 B | 1,312 B | Good |
| Age X25519 | 82,960 B | 152,071 B | Asymmetric overhead |
| GPG Native | 213,785 B | 354,874 B | RSA overhead |

---

### 🎯 Recommendations by Use Case

| Use Case | Recommended Method | Why |
|----------|-------------------|-----|
| **Single user, password-based** | ChaCha20-Poly1305 | Fastest, most efficient |
| **Team collaboration** | Age X25519 | Simple, modern, multi-recipient |
| **Enterprise/OpenPGP** | GPG Native | Standard, compatible, no binary needed |
| **FIPS compliance** | AES-256-GCM | NIST approved, hardware accelerated |
| **High-volume streaming** | XChaCha20-Poly1305 | Extended nonce, safer for many operations |
| **Mobile/ARM devices** | ChaCha20-Poly1305 | No hardware dependency, fast |
| **Legacy PGP integration** | GPG Native | Works with existing GPG infrastructure |

---

### 🔐 GPG vs Age Comparison

| Feature | Age X25519 | GPG Native |
|---------|------------|------------|
| **Speed (encrypt)** | 3,384 ops/sec | 614 ops/sec |
| **Speed (decrypt)** | 3,450 ops/sec | 30 ops/sec |
| **Key Generation** | ~150ms | ~2.2s |
| **External Dependency** | ❌ None | ❌ None (Pure Go!) |
| **Key Format** | Simple text | Armored ASCII |
| **Multiple Recipients** | ✅ Yes | ✅ Yes |
| **Password Protection** | ❌ No | ✅ Yes |
| **OpenPGP Compatible** | ❌ No | ✅ Yes (RFC 4880) |
| **Enterprise Support** | Limited | Excellent |
| **Learning Curve** | Very easy | Moderate |
| **Best For** | Modern apps, teams | Legacy systems, enterprises |

**💡 Recommendation:**
- **New projects:** Use Age (faster, simpler)
- **Existing PGP infrastructure:** Use GPG Native (compatible, no binary needed)
- **Need both:** PODX supports both! Switch anytime.

---

## 🎨 TUI (Terminal User Interface)

Launch the beautiful interactive TUI:

```bash
podx
```

### 📊 Features

#### 1. Dashboard Tab
- 📈 Project status and encryption statistics
- 🔑 Active recipients and key information
- ⚡ Quick actions (encrypt-all, decrypt-all, sync)
- 🛡️ Live security check status
- 🔄 Update notifications

#### 2. Commands Tab
- 📋 Interactive menu for all podx commands
- 📝 Form dialogs for command parameters
- 🔍 Search/filter commands
- ✅ Command output with syntax highlighting

#### 3. Security Tab
- 🔍 Real-time secret scanning
- 📁 Gitignore validation
- 🪝 Pre-commit hook status
- ⚠️ Detailed security findings
- 🔧 Auto-fix recommendations

#### 4. Files Tab
- 📂 File browser with tree view
- 🔐 Encrypt/decrypt actions (press `e` or `d`)
- 📊 Encryption status indicators
- 🔍 Search and filter files
- 👁️ Preview panel for encrypted files
- ✅ Multi-select (Space to toggle)

#### 5. Logs Tab
- 📜 Command execution history
- 🔍 Filter by level (info, warning, error)
- 🎨 Syntax-highlighted output

### ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` / `1,2,3,4,5` | Switch between tabs |
| `↑↓` / `j,k` | Navigate up/down |
| `←→` / `h,l` | Back/Select |
| `Enter` | Confirm action |
| `Space` | Toggle selection (Files tab) |
| `e` | Encrypt selected files |
| `d` | Decrypt selected files |
| `a` | Select all / Deselect all |
| `g` | Go to path |
| `/` | Filter/search |
| `p` | Toggle preview panel |
| `r` | Refresh data |
| `?` | Help overlay |
| `q` / `Ctrl+C` | Quit |

### 🎬 TUI Demo

```bash
# Launch TUI
podx

# Navigate to Files tab (press 2)
# Select files (Space to toggle)
# Press 'e' to encrypt
# Choose encryption method:
#   → Password (ChaCha20, AES-GCM, XChaCha20)
#   → Age Key (asymmetric, multi-recipient)
#   → GPG Key (OpenPGP compatible)
# Enter password or select recipients
# ✓ Files encrypted!
```

---

## 🏗️ Architecture

### System Overview

```mermaid
graph TB
    subgraph "🎨 User Interface"
        CLI["⌨️ CLI Commands"]
        TUI["🖥️ TUI (Bubble Tea)"]
    end

    subgraph "⚙️ Core Components"
        PROJECT["📦 Project Manager"]
        KEYGEN["🔑 Key Generator"]
        SECURITY["🛡️ Security Scanner"]
        UPDATER["🔄 Self-Updater"]
    end

    subgraph "🔐 Crypto Engine"
        FACTORY["🏭 Encryptor Factory"]

        subgraph "Symmetric Algorithms"
            CHACHA["ChaCha20-Poly1305"]
            XCHACHA["XChaCha20-Poly1305"]
            AES["AES-256-GCM"]
        end

        subgraph "Asymmetric Algorithms"
            AGE["Age X25519"]
            GPG["GPG Native (gopenpgp)"]
        end

        KDF["🔐 Argon2id KDF"]
    end

    subgraph "🗄️ Storage"
        KEYSTORE["🔑 Key Store"]
        CONFIG["⚙️ .podx.yaml"]
        GITIGNORE["📄 .gitignore"]
    end

    CLI --> PROJECT
    TUI --> PROJECT
    TUI --> KEYGEN
    TUI --> SECURITY

    PROJECT --> FACTORY
    KEYGEN --> KEYSTORE

    FACTORY --> CHACHA
    FACTORY --> XCHACHA
    FACTORY --> AES
    FACTORY --> AGE
    FACTORY --> GPG

    CHACHA --> KDF
    XCHACHA --> KDF
    AES --> KDF

    PROJECT --> CONFIG
    SECURITY --> GITIGNORE

    KEYSTORE -.->|"~/.config/podx/age-keys.txt"| AGE
    KEYSTORE -.->|"~/.config/podx/gpg-keys/"| GPG

    classDef ui fill:#A0C981,stroke:#303446,stroke-width:2px,color:#303446
    classDef core fill:#D48AEA,stroke:#303446,stroke-width:2px,color:#303446
    classDef crypto fill:#6CB6EB,stroke:#303446,stroke-width:2px,color:#303446
    classDef storage fill:#EF9F76,stroke:#303446,stroke-width:2px,color:#303446

    class CLI,TUI ui
    class PROJECT,KEYGEN,SECURITY,UPDATER core
    class FACTORY,CHACHA,XCHACHA,AES,AGE,GPG,KDF crypto
    class KEYSTORE,CONFIG,GITIGNORE storage
```

### 📁 Project Structure

```
podx/
├── 🎨 UI Layer
│   ├── main.go                 # CLI entry point
│   └── tui/                    # Terminal UI (Bubble Tea)
│       ├── tui.go              # Main TUI model
│       ├── dashboard.go        # Dashboard tab
│       ├── commands.go         # Commands tab
│       ├── security.go         # Security tab
│       ├── files.go            # Files tab (browser)
│       ├── encrypt_dialog.go   # Encryption method selection
│       ├── key_manager.go      # Key management UI
│       └── styles.go           # Dracula-inspired theme
│
├── 🔐 Crypto Layer
│   └── crypto/
│       ├── crypto.go           # Encryptor interface & factory
│       ├── aes_gcm.go          # AES-256-GCM implementation
│       ├── chacha.go           # ChaCha20 & XChaCha20
│       ├── age.go              # Age X25519 encryption
│       ├── gpg.go              # GPG Native (gopenpgp)
│       └── kdf.go              # Argon2id key derivation
│
├── 📦 Core Components
│   ├── project/                # Project workspace management
│   │   └── project.go          # .podx.yaml config, encrypt-all/decrypt-all
│   ├── keygen/                 # Key generation
│   │   ├── keygen.go           # Age & GPG key generation
│   │   └── gpg_keys.go         # GPG key storage & listing
│   ├── security/               # Security scanning
│   │   ├── scanner.go          # Secret pattern detection
│   │   ├── gitignore.go        # Gitignore management
│   │   └── hook.go             # Pre-commit hook
│   ├── parser/                 # Format-preserving .env parser
│   │   └── env.go              # Parse & encrypt .env files
│   └── updater/                # Self-update mechanism
│       └── updater.go          # GitHub releases download
│
└── 📄 Configuration
    ├── .podx.yaml              # Project config (recipients, secrets)
    ├── go.mod                  # Go dependencies
    └── Makefile                # Build automation
```

### 🔄 Encryption Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Factory
    participant Encryptor
    participant KDF
    participant Storage

    User->>CLI: podx encrypt file.txt -m chacha20
    CLI->>CLI: Read file.txt
    CLI->>Factory: NewEncryptor("chacha20")
    Factory->>Encryptor: Create ChaCha20 instance

    CLI->>User: Prompt for password
    User->>CLI: Enter password

    CLI->>KDF: DeriveKey(password, salt)
    Note over KDF: Argon2id with<br/>memory=64MB, time=3, parallelism=4
    KDF->>CLI: Return 32-byte key

    CLI->>Encryptor: Encrypt(plaintext, key)
    Note over Encryptor: Generate random nonce<br/>ChaCha20-Poly1305 AEAD
    Encryptor->>CLI: Return [salt][algo][nonce][ciphertext+tag]

    CLI->>Storage: Write file.txt.podx
    Storage->>CLI: Success
    CLI->>User: ✓ Encrypted: file.txt → file.txt.podx
```

### 🔓 Decryption Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Factory
    participant Encryptor
    participant KDF
    participant Storage

    User->>CLI: podx decrypt file.txt.podx
    CLI->>Storage: Read file.txt.podx
    Storage->>CLI: Return ciphertext

    CLI->>CLI: Parse header [salt][algo]
    CLI->>Factory: NewEncryptor(algo)
    Factory->>Encryptor: Create instance

    CLI->>User: Prompt for password
    User->>CLI: Enter password

    CLI->>KDF: DeriveKey(password, salt)
    KDF->>CLI: Return 32-byte key

    CLI->>Encryptor: Decrypt(ciphertext, key)
    Note over Encryptor: Verify MAC tag<br/>Decrypt with nonce

    alt Authentication Success
        Encryptor->>CLI: Return plaintext
        CLI->>Storage: Write file.txt
        Storage->>CLI: Success
        CLI->>User: ✓ Decrypted: file.txt.podx → file.txt
    else Authentication Failed
        Encryptor->>CLI: Error: tampered or wrong password
        CLI->>User: ✗ Decryption failed
    end
```

### 🔑 Key Management

```mermaid
graph LR
    subgraph "Age Keys"
        AGEGEN["Generate Age Key"]
        AGEPRIV["Private Key<br/>~/.config/podx/age-keys.txt"]
        AGEPUB["Public Key<br/>~/.config/podx/age-recipients/default.txt"]
    end

    subgraph "GPG Keys (Native)"
        GPGGEN["Generate GPG Key<br/>(Pure Go - no GPG binary)"]
        GPGPRIV["Private Key<br/>~/.config/podx/gpg-keys/{ID}_private.asc"]
        GPGPUB["Public Key<br/>~/.config/podx/gpg-keys/{ID}_public.asc"]
        GPGMETA["Metadata<br/>~/.config/podx/gpg-keys/{ID}_meta.txt"]
    end

    AGEGEN --> AGEPRIV
    AGEGEN --> AGEPUB

    GPGGEN --> GPGPRIV
    GPGGEN --> GPGPUB
    GPGGEN --> GPGMETA

    AGEPRIV -.->|"mode 0600"| DECRYPT["Decrypt"]
    AGEPUB -.->|"mode 0644"| ENCRYPT["Encrypt"]

    GPGPRIV -.->|"mode 0600<br/>armored"| DECRYPT
    GPGPUB -.->|"mode 0644<br/>armored"| ENCRYPT

    classDef keyNode fill:#A0C981,stroke:#303446,stroke-width:2px,color:#303446
    classDef fileNode fill:#EF9F76,stroke:#303446,stroke-width:2px,color:#303446
    classDef actionNode fill:#6CB6EB,stroke:#303446,stroke-width:2px,color:#303446

    class AGEGEN,GPGGEN keyNode
    class AGEPRIV,AGEPUB,GPGPRIV,GPGPUB,GPGMETA fileNode
    class ENCRYPT,DECRYPT actionNode
```

---

## 📊 Performance

### Benchmarks (Apple M1)

```
Algorithm                Speed (ops/sec)    Time/op     Memory/op
─────────────────────────────────────────────────────────────────
ChaCha20 Encrypt         1,000,000         1.15 μs     112 B
ChaCha20 Decrypt         4,064,662         0.30 μs      64 B  ⭐

XChaCha20 Encrypt          936,196         1.35 μs     136 B
XChaCha20 Decrypt        2,420,415         0.49 μs      80 B

AES-GCM Encrypt            683,055         1.49 μs   1,360 B
AES-GCM Decrypt          1,564,368         0.83 μs   1,312 B

Age Encrypt                  3,384       344.39 μs  82,960 B
Age Decrypt                  3,450       346.36 μs 152,071 B

GPG Encrypt (Native)           614         1.99 ms 213,785 B
GPG Decrypt (Native)            30        33.42 ms 354,874 B

Age KeyGen                   8,391       148.86 μs   1,520 B
GPG KeyGen (Native)           0.45         2.21 s  4.92 MB
```

### Large File Performance (1MB)

```
Algorithm                Time        Throughput
─────────────────────────────────────────────────
ChaCha20                55.27 ms    18.1 MB/s
XChaCha20               60.12 ms    16.6 MB/s
AES-GCM                 40.32 ms    24.8 MB/s
Age                    380.45 ms     2.6 MB/s
GPG Native            2150.30 ms     0.5 MB/s
```

**Full benchmark report:** [`docs/encryption-test-report.md`](docs/encryption-test-report.md)

---

## 🔒 Security

### Secret Detection Patterns

PODX automatically scans for:

- **AWS Credentials** — Access keys, secret keys, session tokens
- **API Keys** — Stripe, GitHub, Google, SendGrid, Twilio, etc.
- **Database URLs** — PostgreSQL, MySQL, MongoDB connection strings
- **Private Keys** — RSA, SSH, PGP private keys
- **JWT Tokens** — JSON Web Tokens
- **Generic Secrets** — Any string matching `password=`, `secret=`, `api_key=`

### Pre-Commit Hook

```bash
# Install hook
podx hook install

# What the hook does:
# 1. Check for unencrypted secret files (.env, config/*.key)
# 2. Scan staged files for exposed secrets (AWS keys, passwords)
# 3. Validate .gitignore patterns are present
# 4. BLOCK commit if issues found
# 5. Provide fix suggestions

# Example blocked commit:
$ git commit -m "add config"
✗ Security check failed!

Found exposed secrets:
  - config/database.yml:5   AWS_ACCESS_KEY=AKIA...
  - .env:12                 DATABASE_PASSWORD=secret123

Fix suggestions:
  1. Encrypt files: podx encrypt-all
  2. Add to .gitignore: echo ".env" >> .gitignore
  3. Re-run: git add . && git commit
```

### Gitignore Management

```bash
# Auto-fix gitignore
podx check --fix

# Patterns automatically added:
.env
.env.local
.env.*.local
*.key
*.pem
secrets/
config/*.secret
*.podx.decrypted
```

### Key Storage Locations

**Age Keys:**
- Private: `~/.config/podx/age-keys.txt` (mode `0600`)
- Public: `~/.config/podx/age-recipients/default.txt` (mode `0644`)

**GPG Keys (Native):**
- Private: `~/.config/podx/gpg-keys/{ID}_private.asc` (mode `0600`, armored)
- Public: `~/.config/podx/gpg-keys/{ID}_public.asc` (mode `0644`, armored)
- Metadata: `~/.config/podx/gpg-keys/{ID}_meta.txt` (mode `0644`)

### Memory Protection

- Sensitive data (keys, passwords) wiped from memory after use
- No plaintext secrets in logs
- Secure random number generation (crypto/rand)

---

## 📖 Commands Reference

### Project Management

```bash
# Initialize project
podx init
# Creates .podx.yaml with your public key
# Adds .gitignore entries

# Show project status
podx status
# Shows encryption status, recipients, secret files

# Add team member
podx add-recipient Alice age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

# Encrypt all secrets
podx encrypt-all
# Encrypts files matching .podx.yaml secret patterns
# Deletes originals after successful encryption

# Decrypt all secrets
podx decrypt-all
# Decrypts all .podx files in the project
```

### File Encryption

```bash
# Encrypt with password (ChaCha20, default)
podx encrypt secrets.json

# Encrypt with specific algorithm
podx encrypt file.txt -m aes-gcm
podx encrypt file.txt -m xchacha20

# Encrypt with Age (asymmetric)
podx encrypt file.txt -r age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

# Encrypt for multiple Age recipients
podx encrypt file.txt -r alice.txt -r bob.txt -r charlie.txt

# Encrypt with GPG (native, no binary needed)
podx encrypt file.txt -m gpg -r alice@example.com
podx encrypt file.txt --gpg-key "$(cat alice_public.asc)"

# Encrypt .env file (format-preserving)
podx env encrypt .env
# KEY=VALUE becomes KEY=ENC[base64...]
```

### File Decryption

```bash
# Decrypt (auto-detects algorithm)
podx decrypt file.txt.podx

# Decrypt .env file
podx env decrypt .env.podx

# Decrypt with specific key
podx decrypt file.txt.podx --key-file /path/to/key
```

### Key Management

```bash
# Generate Age key
podx keygen -t age
# Output: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p

# Generate GPG key (native, no GPG binary needed)
podx keygen -t gpg -n "Alice" -e "alice@example.com"
# Output: Key ID A1B2C3D4E5F6G7H8

# Show your Age public key
podx key-info
```

### Security

```bash
# Run security checks
podx check
# - Scans for exposed secrets
# - Validates .gitignore
# - Checks encryption status

# Auto-fix gitignore issues
podx check --fix

# Install pre-commit hook
podx hook install

# Check hook status
podx hook status
```

### Update

```bash
# Update to latest version
podx update

# Update to specific version
podx update --version v1.2.3

# Rollback to previous version
podx rollback
```

---

## 🤝 Team Workflow

### Setup (Once per team member)

```bash
# 1. Install PODX
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash

# 2. Generate your key
podx keygen -t age
# Share your PUBLIC key with the team: age1ql3z...

# 3. Clone the project
git clone https://github.com/your-org/your-project
cd your-project

# 4. Decrypt secrets
podx decrypt-all
```

### Daily Workflow

```bash
# Pull latest changes
git pull

# Decrypt updated secrets
podx decrypt-all

# Work on your code...

# Encrypt secrets before committing
podx encrypt-all

# Commit and push
git add .
git commit -m "feat: add new feature"
git push
```

### Adding New Team Member

```bash
# Team admin adds new member's public key
podx add-recipient Bob age1yhm4gctwfmrpz87tdslm550wrx6m79y9f2hdzt0lne5nxx6gdtxqm0u3dv

# Re-encrypt all secrets for the new team
podx encrypt-all

# Commit and push
git add .podx.yaml **/*.podx
git commit -m "chore: add Bob to recipients"
git push

# Bob can now decrypt
# (on Bob's machine)
git pull
podx decrypt-all
```

---

## 🔧 Configuration

### `.podx.yaml` Example

```yaml
version: "1"
backend: age  # or "gpg"

# Team members (recipients)
recipients:
  - name: Alice
    key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
  - name: Bob
    key: age1yhm4gctwfmrpz87tdslm550wrx6m79y9f2hdzt0lne5nxx6gdtxqm0u3dv
  - name: Charlie (GPG)
    key: "-----BEGIN PGP PUBLIC KEY BLOCK-----\n..."

# Secret file patterns (glob)
secrets:
  - "*.env"
  - "*.env.local"
  - "secrets/**"
  - "config/*.key"
  - "config/*.secret"
  - ".aws/credentials"
```

### Environment Variables

```bash
# Custom config directory
export PODX_CONFIG_DIR="$HOME/.podx"

# Custom key file (Age)
export PODX_AGE_KEY_FILE="$HOME/.ssh/age-key.txt"

# Custom recipient file
export PODX_AGE_RECIPIENT_FILE="$HOME/.ssh/age-recipients.txt"

# Skip security checks (not recommended)
export PODX_SKIP_SECURITY_CHECK=1

# Debug mode
export PODX_DEBUG=1
```

---

## 🚀 CI/CD Integration

### GitHub Actions

```yaml
name: Decrypt Secrets

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install PODX
        run: |
          curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Decrypt secrets
        env:
          PODX_AGE_KEY: ${{ secrets.PODX_AGE_PRIVATE_KEY }}
        run: |
          echo "$PODX_AGE_KEY" > /tmp/age-key.txt
          chmod 600 /tmp/age-key.txt
          export PODX_AGE_KEY_FILE=/tmp/age-key.txt
          podx decrypt-all

      - name: Run tests
        run: npm test
```

### GitLab CI

```yaml
decrypt_secrets:
  stage: prepare
  script:
    - curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
    - echo "$PODX_AGE_KEY" > /tmp/age-key.txt
    - chmod 600 /tmp/age-key.txt
    - export PODX_AGE_KEY_FILE=/tmp/age-key.txt
    - podx decrypt-all
  artifacts:
    paths:
      - .env
      - config/
    expire_in: 1 hour
```

---

## 🐛 Troubleshooting

### "No age key found"

```bash
# Generate a key first
podx keygen -t age

# Or specify key file
export PODX_AGE_KEY_FILE=/path/to/your/key.txt
podx decrypt file.txt.podx
```

### "Decryption failed: wrong key"

```bash
# Make sure you're using the same key that was used for encryption
podx key-info  # Show your current public key

# The file was encrypted for different recipients
# Ask the file owner to re-encrypt for your key
```

### "Permission denied: .podx.yaml"

```bash
# Fix file permissions
chmod 644 .podx.yaml

# If it's a Git issue
git update-index --chmod=+x .podx.yaml
```

### Pre-commit hook not working

```bash
# Re-install hook
podx hook uninstall
podx hook install

# Check hook status
podx hook status

# Manual check
.git/hooks/pre-commit
```

### GPG "No private key found"

```bash
# List your GPG keys
podx keygen -t gpg --list

# Generate new GPG key (native, no GPG binary needed)
podx keygen -t gpg -n "Your Name" -e "you@example.com"

# Keys are stored in: ~/.config/podx/gpg-keys/
```

---

## 🤝 Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

```bash
# Fork and clone
git clone https://github.com/your-username/podx
cd podx

# Create branch
git checkout -b feature/amazing-feature

# Make changes and test
make test
make build

# Commit (conventional commits)
git commit -m "feat: add amazing feature"

# Push and create PR
git push origin feature/amazing-feature
```

### Running Tests

```bash
# All tests
go test ./... -v

# With coverage
go test ./... -cover

# Specific package
go test ./crypto -v

# Benchmarks
go test ./crypto -bench=. -benchmem

# Integration tests
go test -run TestGPGIntegration -v
go test -run TestAgeIntegration -v
```

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Write tests for new features
- Document public APIs

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [Age encryption](https://age-encryption.org/) by Filippo Valsorda (Google)
- [ProtonMail gopenpgp](https://github.com/ProtonMail/gopenpgp) for native GPG implementation
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the beautiful TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for TUI styling
- [Go Cryptography](https://pkg.go.dev/golang.org/x/crypto) libraries

---

## 📚 Documentation

- [Encryption Test Report](docs/encryption-test-report.md) — Performance benchmarks and test coverage
- [GPG Native Migration](docs/gpg-native-migration.md) — Migration guide from shell GPG to native
- [Implementation Plans](docs/plans/) — Detailed implementation plans and design docs

---

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=dwirx/podx&type=Date)](https://star-history.com/#dwirx/podx&Date)

---

<p align="center">
  Made with ❤️ for secure secret management
</p>

<p align="center">
  <a href="https://github.com/dwirx/podx">🌟 Star on GitHub</a> •
  <a href="https://github.com/dwirx/podx/issues">🐛 Report Bug</a> •
  <a href="https://github.com/dwirx/podx/issues">💡 Request Feature</a>
</p>
