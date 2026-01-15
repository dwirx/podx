package git

import (
	"path/filepath"
	"strings"
)

// GenerateCommitMessage generates a commit message based on changed files
func GenerateCommitMessage(files []string) string {
	if len(files) == 0 {
		return "chore: update files"
	}

	// Categorize files
	var encrypted, config, source, docs, tests, other []string

	for _, f := range files {
		base := filepath.Base(f)
		ext := filepath.Ext(f)

		switch {
		case strings.HasSuffix(f, ".podx"):
			encrypted = append(encrypted, base)
		case base == ".podx.yaml" || base == ".gitignore":
			config = append(config, base)
		case ext == ".go" || ext == ".js" || ext == ".ts" || ext == ".py":
			source = append(source, base)
		case ext == ".md" || ext == ".txt" || ext == ".rst":
			docs = append(docs, base)
		case strings.Contains(f, "_test.go") || strings.Contains(f, ".test."):
			tests = append(tests, base)
		default:
			other = append(other, base)
		}
	}

	// Build message based on what changed
	var parts []string

	if len(encrypted) > 0 {
		if len(encrypted) == 1 {
			parts = append(parts, "encrypt "+encrypted[0])
		} else {
			parts = append(parts, "encrypt secrets")
		}
	}

	if len(config) > 0 {
		parts = append(parts, "update config")
	}

	if len(source) > 0 {
		if len(source) == 1 {
			parts = append(parts, "update "+source[0])
		} else {
			parts = append(parts, "update source files")
		}
	}

	if len(docs) > 0 {
		parts = append(parts, "update docs")
	}

	if len(tests) > 0 {
		parts = append(parts, "update tests")
	}

	if len(other) > 0 && len(parts) == 0 {
		if len(other) == 1 {
			parts = append(parts, "update "+other[0])
		} else {
			parts = append(parts, "update files")
		}
	}

	if len(parts) == 0 {
		return "chore: update files"
	}

	// Determine prefix
	prefix := "chore"
	if len(encrypted) > 0 && len(source) == 0 && len(docs) == 0 {
		prefix = "chore"
	} else if len(source) > 0 {
		prefix = "feat" // Could be refined based on diff analysis
	} else if len(docs) > 0 {
		prefix = "docs"
	} else if len(tests) > 0 {
		prefix = "test"
	}

	return prefix + ": " + strings.Join(parts, ", ")
}

// SuggestMessageFromDiff could analyze git diff to suggest better messages
// For now, we use file-based heuristics
func SuggestMessageFromDiff(dir string) (string, error) {
	// Get all files that will be committed
	staged, err := GetStagedFiles(dir)
	if err != nil {
		return "", err
	}

	modified, err := GetModifiedFiles(dir)
	if err != nil {
		return "", err
	}

	untracked, err := GetUntrackedFiles(dir)
	if err != nil {
		return "", err
	}

	// Combine all files
	allFiles := append(staged, modified...)
	allFiles = append(allFiles, untracked...)

	return GenerateCommitMessage(allFiles), nil
}
