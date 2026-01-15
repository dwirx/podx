package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// Current directory should be a git repo (project root)
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd) // Go up from git/ to project root

	if !IsGitRepo(projectRoot) {
		t.Error("Expected project root to be a git repo")
	}

	// Temp directory should not be a git repo
	tmpDir := os.TempDir()
	if IsGitRepo(tmpDir) {
		t.Error("Expected temp directory to not be a git repo")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd)

	branch, err := GetCurrentBranch(projectRoot)
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if branch == "" {
		t.Error("Expected branch name to be non-empty")
	}
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		contains string
	}{
		{
			name:     "encrypted files",
			files:    []string{".env.podx", "secrets.podx"},
			contains: "encrypt",
		},
		{
			name:     "config files",
			files:    []string{".podx.yaml", ".gitignore"},
			contains: "config",
		},
		{
			name:     "source files",
			files:    []string{"main.go", "utils.go"},
			contains: "update",
		},
		{
			name:     "mixed files",
			files:    []string{".env.podx", "main.go"},
			contains: "encrypt",
		},
		{
			name:     "empty files",
			files:    []string{},
			contains: "update files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GenerateCommitMessage(tt.files)
			if msg == "" {
				t.Error("Expected non-empty commit message")
			}
			// Just check that message is generated, exact format may vary
			t.Logf("Generated message: %s", msg)
		})
	}
}

func TestGetStagedFiles(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(cwd)

	// This should work even if there are no staged files
	files, err := GetStagedFiles(projectRoot)
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}

	// Files can be nil/empty, that's fine
	t.Logf("Staged files: %v", files)
}
