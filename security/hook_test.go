package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git/hooks directory
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	err := InstallHook(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Check hook exists
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("pre-commit hook not created")
	}

	// Check it's executable
	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("pre-commit hook should be executable")
	}
}

func TestUninstallHook(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git/hooks with our hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)
	hookPath := filepath.Join(hooksDir, "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\n# PODX pre-commit hook\npodx check"), 0755)

	err := UninstallHook(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("pre-commit hook should be removed")
	}
}

func TestHookStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// No hook installed
	if IsHookInstalled(tmpDir) {
		t.Error("should report not installed")
	}

	// Create .git/hooks with our hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)
	hookPath := filepath.Join(hooksDir, "pre-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\n# PODX pre-commit hook\npodx check"), 0755)

	if !IsHookInstalled(tmpDir) {
		t.Error("should report installed")
	}
}
