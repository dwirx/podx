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

func TestFixGitignoreNoDuplicateHeader(t *testing.T) {
	tmpDir := t.TempDir()
	gitignore := filepath.Join(tmpDir, ".gitignore")

	// First call
	FixGitignore(tmpDir, []string{".env"})
	// Second call
	FixGitignore(tmpDir, []string{".env.local"})

	// Read back
	data, _ := os.ReadFile(gitignore)
	content := string(data)

	// Count occurrences of PODX header
	count := strings.Count(content, "# PODX")
	if count != 1 {
		t.Errorf("PODX header appears %d times, want 1", count)
	}
}
