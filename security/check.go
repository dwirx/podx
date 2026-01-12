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
