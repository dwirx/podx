# PODX Security Suite Design

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add pre-commit hook and security checking features to prevent accidental commit of unencrypted secrets.

**Architecture:** New `security/` package with check, scanner, gitignore validation, and hook management. Single `podx check` command that combines all validations.

**Tech Stack:** Go standard library, regexp for pattern matching, os/exec for git operations.

---

## Features Overview

### 1. `podx check` Command

Unified security verification command.

```bash
podx check                    # Full check with detailed output
podx check --pre-commit       # Silent mode, exit code only (for hooks)
podx check --fix              # Auto-fix what can be fixed (gitignore)
```

**Checks performed:**
1. Encryption status - are all secret files in .podx.yaml encrypted?
2. Gitignore validation - are decrypted files in .gitignore?
3. Secret pattern scan - any hardcoded secrets in tracked files?

**Exit codes:**
- `0` = all OK
- `1` = issues found (blocks commit)

### 2. `podx hook install/uninstall` Commands

```bash
podx hook install      # Install pre-commit hook
podx hook uninstall    # Remove pre-commit hook
podx hook status       # Check if hook is installed
```

**Hook behavior:**
- Runs `podx check --pre-commit` before each commit
- Blocks commit if any issues found
- Shows clear error message with remediation steps

### 3. Secret Pattern Scanner

Detects common secret patterns in tracked files:

| Pattern | Example |
|---------|---------|
| AWS Access Key | `AKIA[0-9A-Z]{16}` |
| AWS Secret Key | `aws_secret_access_key` assignments |
| Private Keys | `-----BEGIN.*PRIVATE KEY-----` |
| Password assignments | `password = "value"` |
| API key assignments | `api_key = "value"` |
| Generic secrets | `secret = "value"` |
| Connection strings | `mongodb://user:pass@`, `postgres://` |

**Exclusions:**
- Files matching `.podx` extension (already encrypted)
- `.git/` directory
- `node_modules/`, `vendor/`, `.venv/`
- Binary files
- Files in `.gitignore`

### 4. Gitignore Validator

Checks that all patterns in `.podx.yaml` secrets list are in `.gitignore`.

With `--fix` flag: automatically adds missing patterns.

---

## File Structure

```
security/
├── check.go           # Main check logic, combines all validators
├── scanner.go         # Secret pattern detection
├── gitignore.go       # Gitignore validation and fixing
├── hook.go            # Pre-commit hook install/uninstall
└── patterns.go        # Regex patterns for secret detection

security/check_test.go
security/scanner_test.go
security/gitignore_test.go
security/hook_test.go
```

**main.go additions:**
- `case "check":` → `handleCheck()`
- `case "hook":` → `handleHook()`

---

## Command Output Examples

### `podx check` (issues found)

```
🔍 PODX Security Check

❌ Encryption Status
   .env exists but .env.podx is missing
   Run: podx encrypt-all

⚠️  Gitignore Issues
   Missing from .gitignore: .env
   Run: podx check --fix

❌ Sensitive Patterns Found
   config/database.yml:23  AWS_SECRET_ACCESS_KEY=AKIA...
   src/config.js:15        password = "admin123"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Result: FAILED (3 issues found)

To fix:
  1. Run 'podx encrypt-all' to encrypt secrets
  2. Run 'podx check --fix' to update .gitignore
  3. Remove hardcoded secrets from source files
```

### `podx check` (all OK)

```
🔍 PODX Security Check

✅ Encryption Status    All secrets encrypted
✅ Gitignore            Properly configured
✅ Pattern Scan         No secrets detected

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Result: PASSED
```

### `podx hook install`

```
✅ Pre-commit hook installed

The hook will run 'podx check' before each commit.
If unencrypted secrets are found, the commit will be blocked.

To uninstall: podx hook uninstall
```

---

## Pre-commit Hook Script

```bash
#!/bin/sh
# PODX pre-commit hook
# Installed by: podx hook install

if ! command -v podx &> /dev/null; then
    echo "Error: podx not found in PATH"
    exit 1
fi

podx check --pre-commit
exit $?
```

---

## Implementation Notes

1. **Scanner performance:** Use `filepath.WalkDir` with early exit on excluded directories
2. **Binary detection:** Check first 512 bytes for null bytes
3. **Git integration:** Use `git ls-files` to only scan tracked files
4. **Pattern matching:** Compile regexes once at init, reuse for all files
5. **Gitignore parsing:** Simple line-by-line, handle comments and negation

---

## Test Plan

1. **check_test.go:**
   - Test with encrypted project → PASS
   - Test with unencrypted .env → FAIL
   - Test --pre-commit mode exits correctly

2. **scanner_test.go:**
   - Test each pattern type is detected
   - Test exclusions work (node_modules, .podx files)
   - Test binary files are skipped

3. **gitignore_test.go:**
   - Test missing patterns detected
   - Test --fix adds patterns correctly
   - Test existing patterns not duplicated

4. **hook_test.go:**
   - Test install creates hook file
   - Test uninstall removes hook
   - Test status reports correctly
