# PODX Security Suite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `podx check` command and pre-commit hook to prevent accidental commit of unencrypted secrets.

**Architecture:** New `security/` package with pattern scanner, gitignore validator, and hook manager. Single `podx check` command integrates all checks. Pre-commit hook calls `podx check --pre-commit`.

**Tech Stack:** Go standard library (regexp, os/exec, filepath), git hooks

---

### Task 1: Create Security Package with Pattern Definitions

**Files:**
- Create: `security/patterns.go`
- Create: `security/patterns_test.go`

**Step 1: Write the test file**

Create `security/patterns_test.go`:

```go
package security

import "testing"

func TestPatternMatching(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"aws_access_key", "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE", true},
		{"aws_secret", "aws_secret_access_key = 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'", true},
		{"private_key", "-----BEGIN RSA PRIVATE KEY-----", true},
		{"password_assignment", `password = "secret123"`, true},
		{"password_single_quote", `password='secret123'`, true},
		{"api_key", `api_key = "sk-1234567890"`, true},
		{"generic_secret", `secret = "mysecret"`, true},
		{"connection_string", "mongodb://user:pass@localhost:27017", true},
		{"postgres_url", "postgres://admin:password@db.example.com/mydb", true},
		{"safe_code", "const username = 'admin'", false},
		{"safe_comment", "// TODO: add password validation", false},
		{"encrypted_value", "PASSWORD=ENC[age:YWdlLWVuY3J5cH...]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsSecret(tt.content); got != tt.want {
				t.Errorf("ContainsSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindSecrets(t *testing.T) {
	content := `DB_HOST=localhost
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
password = "admin123"
DEBUG=true`

	matches := FindSecrets(content)
	if len(matches) != 2 {
		t.Errorf("FindSecrets() found %d matches, want 2", len(matches))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./security -run TestPattern -v`
Expected: FAIL (package doesn't exist)

**Step 3: Write the implementation**

Create `security/patterns.go`:

```go
package security

import (
	"regexp"
	"strings"
)

// SecretPattern defines a pattern to detect secrets
type SecretPattern struct {
	Name    string
	Pattern *regexp.Regexp
}

// Compiled patterns for secret detection
var secretPatterns = []SecretPattern{
	{"AWS Access Key ID", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"AWS Secret Key", regexp.MustCompile(`(?i)aws_secret_access_key\s*[=:]\s*['"]?[A-Za-z0-9/+=]{40}`)},
	{"Private Key", regexp.MustCompile(`-----BEGIN\s+(?:RSA|DSA|EC|OPENSSH|PGP)?\s*PRIVATE KEY-----`)},
	{"Password Assignment", regexp.MustCompile(`(?i)password\s*[=:]\s*['"][^'"]+['"]`)},
	{"API Key Assignment", regexp.MustCompile(`(?i)api[_-]?key\s*[=:]\s*['"][^'"]+['"]`)},
	{"Secret Assignment", regexp.MustCompile(`(?i)secret\s*[=:]\s*['"][^'"]+['"]`)},
	{"MongoDB Connection", regexp.MustCompile(`mongodb(?:\+srv)?://[^:]+:[^@]+@`)},
	{"Postgres Connection", regexp.MustCompile(`postgres(?:ql)?://[^:]+:[^@]+@`)},
}

// encryptedPattern matches already-encrypted values (should be ignored)
var encryptedPattern = regexp.MustCompile(`ENC\[[a-zA-Z0-9-]+:[^\]]+\]`)

// ContainsSecret returns true if the content contains any secret pattern
func ContainsSecret(content string) bool {
	// Skip if it's an encrypted value
	if encryptedPattern.MatchString(content) {
		return false
	}

	for _, p := range secretPatterns {
		if p.Pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// SecretMatch represents a found secret
type SecretMatch struct {
	Pattern string
	Line    int
	Content string
}

// FindSecrets returns all secret matches in content
func FindSecrets(content string) []SecretMatch {
	var matches []SecretMatch
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// Skip encrypted values
		if encryptedPattern.MatchString(line) {
			continue
		}

		for _, p := range secretPatterns {
			if p.Pattern.MatchString(line) {
				matches = append(matches, SecretMatch{
					Pattern: p.Name,
					Line:    lineNum + 1,
					Content: truncate(line, 60),
				})
				break // One match per line is enough
			}
		}
	}
	return matches
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./security -run TestPattern -v`
Expected: PASS

**Step 5: Commit**

```bash
git add security/patterns.go security/patterns_test.go
git commit -m "feat(security): add secret pattern detection"
```

---

### Task 2: Create File Scanner

**Files:**
- Create: `security/scanner.go`
- Create: `security/scanner_test.go`

**Step 1: Write the test file**

Create `security/scanner_test.go`:

```go
package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file with secrets
	testFile := filepath.Join(tmpDir, "config.js")
	content := `const config = {
  host: "localhost",
  password: "secret123",
  apiKey: "test"
};`
	os.WriteFile(testFile, []byte(content), 0644)

	matches, err := ScanFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 1 {
		t.Errorf("ScanFile() found %d matches, want 1", len(matches))
	}

	if matches[0].Line != 3 {
		t.Errorf("match on line %d, want 3", matches[0].Line)
	}
}

func TestScanFileBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create binary file
	binFile := filepath.Join(tmpDir, "binary.dat")
	os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	matches, err := ScanFile(binFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 0 {
		t.Error("binary files should be skipped")
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"node_modules/pkg/index.js", true},
		{".git/objects/abc", true},
		{"vendor/lib/file.go", true},
		{".env.podx", true},
		{"secrets.podx", true},
		{"src/main.go", false},
		{"config/database.yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := ShouldSkipPath(tt.path); got != tt.want {
				t.Errorf("ShouldSkipPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./security -run "TestScan|TestShouldSkip" -v`
Expected: FAIL

**Step 3: Write the implementation**

Create `security/scanner.go`:

```go
package security

import (
	"os"
	"path/filepath"
	"strings"
)

// Directories to skip during scanning
var skipDirs = []string{
	"node_modules",
	".git",
	"vendor",
	".venv",
	"venv",
	"__pycache__",
	".idea",
	".vscode",
}

// ShouldSkipPath returns true if path should be skipped
func ShouldSkipPath(path string) bool {
	// Skip .podx files (already encrypted)
	if strings.HasSuffix(path, ".podx") {
		return true
	}

	// Skip known directories
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		for _, skip := range skipDirs {
			if part == skip {
				return true
			}
		}
	}

	return false
}

// ScanFile scans a single file for secrets
func ScanFile(path string) ([]SecretMatch, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Skip binary files (check for null bytes in first 512 bytes)
	checkLen := len(data)
	if checkLen > 512 {
		checkLen = 512
	}
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return nil, nil // Binary file, skip
		}
	}

	return FindSecrets(string(data)), nil
}

// ScanResult represents scan results for a file
type ScanResult struct {
	Path    string
	Matches []SecretMatch
}

// ScanDirectory scans all files in directory for secrets
func ScanDirectory(root string) ([]ScanResult, error) {
	var results []ScanResult

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		relPath, _ := filepath.Rel(root, path)

		// Skip directories we don't care about
		if d.IsDir() {
			if ShouldSkipPath(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files we don't care about
		if ShouldSkipPath(relPath) {
			return nil
		}

		matches, err := ScanFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		if len(matches) > 0 {
			results = append(results, ScanResult{
				Path:    relPath,
				Matches: matches,
			})
		}

		return nil
	})

	return results, err
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./security -run "TestScan|TestShouldSkip" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add security/scanner.go security/scanner_test.go
git commit -m "feat(security): add file scanner for secrets"
```

---

### Task 3: Create Gitignore Validator

**Files:**
- Create: `security/gitignore.go`
- Create: `security/gitignore_test.go`

**Step 1: Write the test file**

Create `security/gitignore_test.go`:

```go
package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore with only .env
	gitignore := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignore, []byte(".env\n"), 0644)

	// Required patterns
	required := []string{".env", ".env.local", "secrets.yaml"}

	missing := CheckGitignore(tmpDir, required)

	if len(missing) != 2 {
		t.Errorf("CheckGitignore() found %d missing, want 2", len(missing))
	}

	// .env should not be missing
	for _, m := range missing {
		if m == ".env" {
			t.Error(".env should not be reported as missing")
		}
	}
}

func TestFixGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	gitignore := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignore, []byte("# Existing\nnode_modules/\n"), 0644)

	missing := []string{".env", ".env.local"}
	err := FixGitignore(tmpDir, missing)
	if err != nil {
		t.Fatal(err)
	}

	// Read back
	data, _ := os.ReadFile(gitignore)
	content := string(data)

	if !strings.Contains(content, ".env") {
		t.Error(".env not added to gitignore")
	}
	if !strings.Contains(content, ".env.local") {
		t.Error(".env.local not added to gitignore")
	}
	if !strings.Contains(content, "# PODX") {
		t.Error("PODX comment not added")
	}
}

func TestCheckGitignoreNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	required := []string{".env"}

	missing := CheckGitignore(tmpDir, required)

	if len(missing) != 1 {
		t.Errorf("should report all as missing when no .gitignore")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./security -run TestCheckGitignore -v`
Expected: FAIL

**Step 3: Write the implementation**

Create `security/gitignore.go`:

```go
package security

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// CheckGitignore checks if required patterns are in .gitignore
// Returns list of missing patterns
func CheckGitignore(dir string, required []string) []string {
	gitignorePath := filepath.Join(dir, ".gitignore")

	// Read existing patterns
	existing := make(map[string]bool)
	if file, err := os.Open(gitignorePath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				existing[line] = true
			}
		}
	}

	// Find missing
	var missing []string
	for _, pattern := range required {
		if !existing[pattern] {
			missing = append(missing, pattern)
		}
	}

	return missing
}

// FixGitignore adds missing patterns to .gitignore
func FixGitignore(dir string, patterns []string) error {
	if len(patterns) == 0 {
		return nil
	}

	gitignorePath := filepath.Join(dir, ".gitignore")

	// Open for append (create if not exists)
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add header
	f.WriteString("\n# PODX - Decrypted secrets (DO NOT COMMIT)\n")

	// Add patterns
	for _, pattern := range patterns {
		f.WriteString(pattern + "\n")
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./security -run TestCheckGitignore -v`
Expected: PASS

**Step 5: Commit**

```bash
git add security/gitignore.go security/gitignore_test.go
git commit -m "feat(security): add gitignore validator"
```

---

### Task 4: Create Check Command Logic

**Files:**
- Create: `security/check.go`
- Create: `security/check_test.go`

**Step 1: Write the test file**

Create `security/check_test.go`:

```go
package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .podx.yaml
	podxYaml := `version: 1
backend: age
recipients:
  - name: Test
    key: age1test
secrets:
  - .env
`
	os.WriteFile(filepath.Join(tmpDir, ".podx.yaml"), []byte(podxYaml), 0644)

	// Create unencrypted .env (should fail)
	os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("SECRET=value"), 0644)

	result := CheckProject(tmpDir, false)

	if result.Passed {
		t.Error("CheckProject should fail with unencrypted .env")
	}

	if len(result.EncryptionIssues) != 1 {
		t.Errorf("expected 1 encryption issue, got %d", len(result.EncryptionIssues))
	}
}

func TestCheckProjectPassed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .podx.yaml
	podxYaml := `version: 1
backend: age
secrets:
  - .env
`
	os.WriteFile(filepath.Join(tmpDir, ".podx.yaml"), []byte(podxYaml), 0644)

	// Create encrypted .env.podx only (no plain .env)
	os.WriteFile(filepath.Join(tmpDir, ".env.podx"), []byte("SECRET=ENC[age:xxx]"), 0644)

	// Create .gitignore
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(".env\n"), 0644)

	result := CheckProject(tmpDir, false)

	if !result.Passed {
		t.Errorf("CheckProject should pass: %+v", result)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./security -run TestCheckProject -v`
Expected: FAIL

**Step 3: Write the implementation**

Create `security/check.go`:

```go
package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckResult holds the result of a security check
type CheckResult struct {
	Passed           bool
	EncryptionIssues []string
	GitignoreIssues  []string
	SecretFindings   []ScanResult
}

// PodxConfig represents .podx.yaml
type PodxConfig struct {
	Version int      `yaml:"version"`
	Secrets []string `yaml:"secrets"`
}

// CheckProject performs all security checks on a project
func CheckProject(dir string, fix bool) CheckResult {
	result := CheckResult{Passed: true}

	// Load .podx.yaml
	config, err := loadPodxConfig(dir)
	if err != nil {
		// No .podx.yaml, nothing to check
		return result
	}

	// Check 1: Encryption status
	for _, pattern := range config.Secrets {
		plainPath := filepath.Join(dir, pattern)
		encPath := plainPath + ".podx"

		plainExists := fileExists(plainPath)
		encExists := fileExists(encPath)

		if plainExists && !encExists {
			result.EncryptionIssues = append(result.EncryptionIssues,
				fmt.Sprintf("%s exists but %s.podx is missing", pattern, pattern))
			result.Passed = false
		} else if plainExists && encExists {
			result.EncryptionIssues = append(result.EncryptionIssues,
				fmt.Sprintf("%s should be deleted after encryption", pattern))
			result.Passed = false
		}
	}

	// Check 2: Gitignore
	missing := CheckGitignore(dir, config.Secrets)
	if len(missing) > 0 {
		result.GitignoreIssues = missing
		result.Passed = false

		if fix {
			FixGitignore(dir, missing)
			result.GitignoreIssues = nil // Fixed
		}
	}

	// Check 3: Secret patterns
	scanResults, _ := ScanDirectory(dir)
	if len(scanResults) > 0 {
		result.SecretFindings = scanResults
		result.Passed = false
	}

	return result
}

func loadPodxConfig(dir string) (*PodxConfig, error) {
	data, err := os.ReadFile(filepath.Join(dir, ".podx.yaml"))
	if err != nil {
		return nil, err
	}

	var config PodxConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FormatResult formats check result for display
func FormatResult(result CheckResult, preCommit bool) string {
	if preCommit {
		if result.Passed {
			return ""
		}
		return "PODX check failed. Run 'podx check' for details."
	}

	var sb strings.Builder
	sb.WriteString("🔍 PODX Security Check\n\n")

	// Encryption status
	if len(result.EncryptionIssues) == 0 {
		sb.WriteString("✅ Encryption Status    All secrets encrypted\n")
	} else {
		sb.WriteString("❌ Encryption Status\n")
		for _, issue := range result.EncryptionIssues {
			sb.WriteString(fmt.Sprintf("   %s\n", issue))
		}
	}

	// Gitignore
	if len(result.GitignoreIssues) == 0 {
		sb.WriteString("✅ Gitignore            Properly configured\n")
	} else {
		sb.WriteString("⚠️  Gitignore Issues\n")
		for _, pattern := range result.GitignoreIssues {
			sb.WriteString(fmt.Sprintf("   Missing: %s\n", pattern))
		}
	}

	// Secret patterns
	if len(result.SecretFindings) == 0 {
		sb.WriteString("✅ Pattern Scan         No secrets detected\n")
	} else {
		sb.WriteString("❌ Sensitive Patterns Found\n")
		for _, file := range result.SecretFindings {
			for _, match := range file.Matches {
				sb.WriteString(fmt.Sprintf("   %s:%d  %s\n", file.Path, match.Line, match.Content))
			}
		}
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	if result.Passed {
		sb.WriteString("Result: PASSED\n")
	} else {
		sb.WriteString("Result: FAILED\n")
	}

	return sb.String()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./security -run TestCheckProject -v`
Expected: PASS

**Step 5: Commit**

```bash
git add security/check.go security/check_test.go
git commit -m "feat(security): add check command logic"
```

---

### Task 5: Create Hook Manager

**Files:**
- Create: `security/hook.go`
- Create: `security/hook_test.go`

**Step 1: Write the test file**

Create `security/hook_test.go`:

```go
package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git/hooks directory
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	err := InstallHook(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Check hook exists
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("pre-commit hook not created")
	}

	// Check it's executable
	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("pre-commit hook should be executable")
	}
}

func TestUninstallHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git/hooks with our hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)
	hookPath := filepath.Join(hooksDir, "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\n# PODX pre-commit hook\npodx check"), 0755)

	err := UninstallHook(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("pre-commit hook should be removed")
	}
}

func TestHookStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// No hook installed
	if IsHookInstalled(tmpDir) {
		t.Error("should report not installed")
	}

	// Create .git/hooks with our hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)
	hookPath := filepath.Join(hooksDir, "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\n# PODX pre-commit hook\npodx check"), 0755)

	if !IsHookInstalled(tmpDir) {
		t.Error("should report installed")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./security -run TestHook -v`
Expected: FAIL

**Step 3: Write the implementation**

Create `security/hook.go`:

```go
package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookMarker = "# PODX pre-commit hook"

const hookScript = `#!/bin/sh
# PODX pre-commit hook
# Installed by: podx hook install

if ! command -v podx &> /dev/null; then
    echo "Error: podx not found in PATH"
    exit 1
fi

podx check --pre-commit
exit $?
`

// InstallHook installs the pre-commit hook
func InstallHook(dir string) error {
	hooksDir := filepath.Join(dir, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "pre-commit")

	// Check if .git exists
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository")
	}

	// Create hooks directory if needed
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	// Check if hook already exists and isn't ours
	if data, err := os.ReadFile(hookPath); err == nil {
		if !strings.Contains(string(data), hookMarker) {
			return fmt.Errorf("pre-commit hook already exists (not PODX)")
		}
	}

	// Write hook
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return err
	}

	return nil
}

// UninstallHook removes the pre-commit hook
func UninstallHook(dir string) error {
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")

	// Check if it's our hook
	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to uninstall
		}
		return err
	}

	if !strings.Contains(string(data), hookMarker) {
		return fmt.Errorf("pre-commit hook is not PODX hook")
	}

	return os.Remove(hookPath)
}

// IsHookInstalled returns true if PODX hook is installed
func IsHookInstalled(dir string) bool {
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")

	data, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(data), hookMarker)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./security -run TestHook -v`
Expected: PASS

**Step 5: Commit**

```bash
git add security/hook.go security/hook_test.go
git commit -m "feat(security): add pre-commit hook manager"
```

---

### Task 6: Integrate Commands into main.go

**Files:**
- Modify: `main.go`

**Step 1: Add check command handler**

Add to main.go after the "help" case:

```go
case "check":
	handleCheck(os.Args[2:])
case "hook":
	if len(os.Args) < 3 {
		fmt.Println("Usage: podx hook <install|uninstall|status>")
		os.Exit(1)
	}
	handleHook(os.Args[2], os.Args[3:])
```

**Step 2: Add handler functions**

Add to main.go:

```go
func handleCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	preCommit := fs.Bool("pre-commit", false, "Pre-commit mode (exit code only)")
	fix := fs.Bool("fix", false, "Auto-fix gitignore issues")
	fs.Parse(args)

	cwd, _ := os.Getwd()
	result := security.CheckProject(cwd, *fix)

	output := security.FormatResult(result, *preCommit)
	if output != "" {
		fmt.Print(output)
	}

	if !result.Passed {
		os.Exit(1)
	}
}

func handleHook(subcmd string, args []string) {
	cwd, _ := os.Getwd()

	switch subcmd {
	case "install":
		if err := security.InstallHook(cwd); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Pre-commit hook installed")
		fmt.Println("\nThe hook will run 'podx check' before each commit.")
		fmt.Println("If unencrypted secrets are found, the commit will be blocked.")
		fmt.Println("\nTo uninstall: podx hook uninstall")

	case "uninstall":
		if err := security.UninstallHook(cwd); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Pre-commit hook removed")

	case "status":
		if security.IsHookInstalled(cwd) {
			fmt.Println("✅ PODX pre-commit hook is installed")
		} else {
			fmt.Println("❌ PODX pre-commit hook is not installed")
			fmt.Println("Run 'podx hook install' to enable")
		}

	default:
		fmt.Printf("Unknown hook command: %s\n", subcmd)
		fmt.Println("Usage: podx hook <install|uninstall|status>")
		os.Exit(1)
	}
}
```

**Step 3: Add import**

Add to imports in main.go:

```go
"github.com/hades/podx/security"
```

**Step 4: Update printUsage**

Add to printUsage() in the OTHER section:

```go
  check      Check for unencrypted secrets
  hook       Manage pre-commit hook
```

**Step 5: Test commands work**

Run: `go run . check --help`
Run: `go run . hook status`
Expected: Commands work without error

**Step 6: Commit**

```bash
git add main.go
git commit -m "feat: add check and hook commands to CLI"
```

---

### Task 7: Update CLAUDE.md and Documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add security section to CLAUDE.md**

Add after Testing section:

```markdown
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
- `patterns.go` - Secret pattern detection (AWS keys, passwords, API keys)
- `scanner.go` - File scanning with binary detection
- `gitignore.go` - Gitignore validation and fixing
- `hook.go` - Pre-commit hook management
- `check.go` - Main check command logic
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add security features to CLAUDE.md"
```

---

### Task 8: Run Full Test Suite

**Files:**
- None (verification only)

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 2: Run with coverage**

Run: `go test ./... -cover`
Expected: security package should have good coverage

**Step 3: Test commands manually**

```bash
go run . check
go run . hook status
go run . hook install
go run . check
go run . hook uninstall
```

**Step 4: Build and verify**

Run: `make build`
Expected: Binary builds successfully

**Step 5: Final commit if needed**

Run: `git status`
If any uncommitted changes, commit them.
