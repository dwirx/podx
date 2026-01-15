package docs

import (
	"os"
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

	// Define required sections and keywords
	requiredItems := []string{
		"# PODX",
		"## Features",
		"## Installation",
		"## Quick Start",
		"## Encryption Methods",
		"### Symmetric Encryption",
		"### Asymmetric Encryption",
		"#### Age X25519",
		"#### GPG/PGP",
		"## TUI (Terminal User Interface)",
		"## Performance",
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

	for _, item := range requiredItems {
		if !strings.Contains(text, item) {
			t.Errorf("README.md missing required section or keyword: %q", item)
		}
	}
}
