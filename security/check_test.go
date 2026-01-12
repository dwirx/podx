package security

import (
	"os"
	"path/filepath"
	"strings"
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

func TestCheckProjectBothExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .podx.yaml
	podxYaml := `version: 1
backend: age
secrets:
  - .env
`
	os.WriteFile(filepath.Join(tmpDir, ".podx.yaml"), []byte(podxYaml), 0644)

	// Create both plain AND encrypted (should fail - plain should be deleted)
	os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("SECRET=value"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".env.podx"), []byte("SECRET=ENC[age:xxx]"), 0644)

	result := CheckProject(tmpDir, false)

	if result.Passed {
		t.Error("CheckProject should fail when both .env and .env.podx exist")
	}

	// Should report that .env should be deleted
	found := false
	for _, issue := range result.EncryptionIssues {
		if strings.Contains(issue, "deleted") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should report .env needs to be deleted after encryption")
	}
}
