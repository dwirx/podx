package docs

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestREADMEStructure(t *testing.T) {
	// Read README.md from the root directory
	content, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}
	text := string(content)

	// Define required sections and keywords (with regex for emoji tolerance)
	requiredPatterns := map[string]string{
		"# PODX":                          `# PODX`,
		"## Features":                     `## .*Features`,
		"## Installation":                 `## .*Installation`,
		"## Quick Start":                  `## .*Quick Start`,
		"## Encryption Methods":           `## .*Encryption Methods`,
		"### Symmetric Encryption":        `### .*Symmetric Encryption`,
		"### Asymmetric Encryption":       `### .*Asymmetric Encryption`,
		"#### Age X25519":                 `#### .*Age X25519`,
		"#### GPG/PGP":                    `#### .*GPG/PGP`,
		"## TUI (Terminal User Interface)": `## .*TUI \(Terminal User Interface\)`,
		"## Performance":                  `## .*Performance`,
	}

	// Check patterns
	for desc, pattern := range requiredPatterns {
		matched, err := regexp.MatchString(pattern, text)
		if err != nil {
			t.Errorf("Invalid regex pattern for %q: %v", desc, err)
		}
		if !matched {
			t.Errorf("README.md missing required section: %q (pattern: %s)", desc, pattern)
		}
	}

	// Define required keywords (exact string match)
	requiredKeywords := []string{
		// GPG specific keywords
		"Native Go GPG",
		"gopenpgp",
		"No external GPG binary required",
		"RSA 4096-bit",
		// Encryption algorithms
		"AES-256-GCM",
		"ChaCha20-Poly1305",
		"XChaCha20-Poly1305",
	}

	for _, keyword := range requiredKeywords {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
			t.Errorf("README.md missing required keyword: %q", keyword)
		}
	}
}
