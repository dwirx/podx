# PODX

🔐 **Encryption CLI** with **project workspace** support, **Age**, **GPG**, **AES-GCM**, and **ChaCha20**.

## Quick Start

```bash
# Build
cd /home/hades/fun/ironvault
go build -o podx .

# Generate key
./podx keygen -t age

# Initialize project
./podx init

# Create .env
echo "API_KEY=secret123" > .env

# Encrypt all secrets
./podx encrypt-all

# Decrypt all secrets
./podx decrypt-all
```

## Project Workflow

```bash
# 1. Initialize project (creates .podx.yaml)
podx init

# 2. Add team members
podx add-recipient -n "Team Member" -k age1xxx...

# 3. Encrypt secrets before committing
podx encrypt-all

# 4. Commit encrypted files
git add .podx.yaml .env.podx
git commit -m "Add encrypted secrets"

# 5. Team member decrypts after clone
podx decrypt-all
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize PODX project |
| `add-recipient` | Add team member |
| `encrypt-all` | Encrypt all secrets |
| `decrypt-all` | Decrypt all secrets |
| `status` | Show project status |
| `keygen` | Generate Age/GPG keys |
| `encrypt` | Encrypt single file |
| `decrypt` | Decrypt single file |
| `env encrypt` | Encrypt .env (format-preserving) |

## Config (.podx.yaml)

```yaml
version: 1
backend: age
recipients:
  - name: Owner
    key: age1xc2ttxdm60507q6wqqmsk695arqxn4x3zpq43dkstwxhaxkccaxspz47na
  - name: Team Member
    key: age1abc...
secrets:
  - .env
  - config/secrets.yaml
```

## File Structure

```
project/
├── .podx.yaml        # Config (commit to git)
├── .env.podx         # Encrypted (commit to git)
├── .env              # Decrypted (gitignored)
└── .gitignore        # Auto-updated
```

## Keygen

```bash
# Generate Age key pair
podx keygen -t age

# Output:
# ╔══════════════════════════════════════════════════════════════════════╗
# ║                 🔑 PODX Key Generated Successfully                 ║
# ╠══════════════════════════════════════════════════════════════════════╣
# ║ Backend: age                                                         ║
# ║ Key file: ~/.config/podx/age-keys.txt                                ║
# ╠══════════════════════════════════════════════════════════════════════╣
# ║ Public Key:                                                          ║
# ║   age1xc2ttxdm60507q6wqqmsk695arqxn4x3zpq43dkstwxhaxkccaxspz47na     ║
# ║ Private Key:                                                         ║
# ║   AGE-SECRET-KEY-1R5JXEELUTZSVS5TJHJ66X57ZJJ3P9AJYJ...               ║
# ╚══════════════════════════════════════════════════════════════════════╝
```

## License

MIT
