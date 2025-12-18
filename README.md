# PODX

🔐 **Encryption CLI** with project workspace, Age, GPG, AES-GCM, ChaCha20.

## Installation

### Linux / macOS (One-liner)
```bash
curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
```

### Windows (PowerShell as Admin)
```powershell
iwr -useb https://raw.githubusercontent.com/dwirx/podx/main/install.ps1 | iex
```

### Manual Download
Download binary from [Releases](https://github.com/dwirx/podx/releases):
- `podx-linux-amd64` - Linux Intel/AMD
- `podx-linux-arm64` - Linux ARM
- `podx-darwin-amd64` - macOS Intel
- `podx-darwin-arm64` - macOS Apple Silicon
- `podx-windows-amd64.exe` - Windows

### Build from Source
```bash
git clone https://github.com/dwirx/podx
cd podx
go build -o podx .
```

## Quick Start

```bash
# 1. Generate key
podx keygen -t age

# 2. Init project
podx init

# 3. Create secrets
echo "API_KEY=secret123" > .env

# 4. Encrypt & commit
podx encrypt-all
git add .podx.yaml .env.podx

# 5. Decrypt after clone
podx decrypt-all
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize project |
| `add-recipient` | Add team member |
| `encrypt-all` | Encrypt all secrets |
| `decrypt-all` | Decrypt all secrets |
| `status` | Show project status |
| `keygen` | Generate Age/GPG keys |
| `encrypt` | Encrypt single file |
| `decrypt` | Decrypt single file |
| `update` | Self-update |
| `version` | Show version info |

## Self-Update

```bash
podx update    # Update to latest version
podx version   # Check current version
```

## Config (.podx.yaml)

```yaml
version: 1
backend: age
recipients:
  - name: Owner
    key: age1xc2ttxdm60507...
secrets:
  - .env
```

## Releasing (for maintainers)

```bash
# Create and push tag
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions auto-builds all platforms
```

## License

MIT
