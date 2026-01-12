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
