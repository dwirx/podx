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
