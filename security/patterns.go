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
