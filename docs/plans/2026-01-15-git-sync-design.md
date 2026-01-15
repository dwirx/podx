# PODX Git Sync Feature Design

## Overview

`podx sync` adalah single command yang menggabungkan encrypt → security check → git add → git commit → git push dengan safety layer built-in.

## Command Interface

```bash
# Basic usage - interactive commit message
podx sync

# Dengan commit message
podx sync -m "feat: add authentication"

# Override remote/branch
podx sync --remote upstream --branch develop

# Kombinasi
podx sync -m "fix: login bug" --remote origin --branch hotfix
```

## Workflow

```
podx sync
    │
    ├─→ 1. Auto-encrypt (file di secrets pattern dari .podx.yaml)
    ├─→ 2. Security check (scan semua staged files)
    │       ├─→ Auto-fix: gitignore, encrypt remaining
    │       └─→ Block/warn berdasarkan severity
    ├─→ 3. Git add (files yang sudah aman)
    ├─→ 4. Git commit (dengan message prompt/flag)
    └─→ 5. Git push (ke current branch + origin, atau override)
```

## Security Severity System

| Level  | Contoh                                    | Default Behavior              |
|--------|-------------------------------------------|-------------------------------|
| HIGH   | AWS keys, passwords in code, private keys | Block total, harus fix manual |
| MEDIUM | Potential API keys, suspicious patterns   | Tanya user, bisa skip         |
| LOW    | Generic secrets pattern, uncertain        | Warning, lanjut               |

### Config Override

```yaml
# .podx.yaml
version: 1
backend: age
sync:
  security_mode: default  # default | strict | relaxed
  # strict: semua severity block
  # relaxed: semua jadi warning only
  # default: severity-based
```

## Commit Message Flow

```
$ podx sync

Encrypting files...
✓ .env → .env.podx

Security check...
✓ No issues found

Staged files:
  M .env.podx
  M config/settings.go

Suggested commit message:
┌────────────────────────────────────────┐
│ chore: encrypt .env, update settings   │
└────────────────────────────────────────┘
Edit message (Enter to accept, e to edit):
```

Jika user tekan Enter, commit dengan suggested message.
Jika user tekan 'e', buka editor untuk edit message.
Jika ada flag `-m`, skip prompt dan langsung pakai message tersebut.

## Push Behavior

- Default: push ke current branch + remote `origin`
- Override dengan flags: `--remote <name>` dan `--branch <name>`
- Jika branch belum ada di remote, otomatis `--set-upstream`

## File Structure

```
├── git/
│   ├── git.go          # Core git operations (add, commit, push)
│   ├── sync.go         # Main sync command logic
│   ├── message.go      # Auto-generate commit message
│   └── git_test.go     # Tests
├── main.go             # Add sync command routing
├── tui/
│   ├── commands.go     # Add Sync menu item
│   └── dashboard.go    # Add Sync quick action
```

## Implementation Steps

1. Create `git/` package with core git operations
2. Implement auto-encrypt step (reuse project.EncryptAll)
3. Implement security check with severity levels (extend security package)
4. Implement commit message generation and prompt
5. Implement push with smart defaults
6. Add `sync` command to main.go
7. Add TUI integration (Commands tab + Dashboard)
8. Add tests

## TUI Integration

### Commands Tab
- Add "Sync to Git" menu item
- Shows same flow as CLI but with TUI prompts

### Dashboard
- Add "Sync" quick action button
- One-click sync with interactive prompts
