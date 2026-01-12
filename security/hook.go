package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookMarker = "# PODX pre-commit hook"

const hookScript = `#!/bin/sh
# PODX pre-commit hook
# Installed by: podx hook install

if ! command -v podx > /dev/null 2>&1; then
    echo "Error: podx not found in PATH"
    exit 1
fi

podx check --pre-commit
exit $?
`

// InstallHook installs the pre-commit hook
func InstallHook(dir string) error {
	hooksDir := filepath.Join(dir, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "pre-commit")

	// Check if .git exists
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository")
	}

	// Create hooks directory if needed
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	// Check if hook already exists and isn't ours
	if data, err := os.ReadFile(hookPath); err == nil {
		if !strings.Contains(string(data), hookMarker) {
			return fmt.Errorf("pre-commit hook already exists (not PODX)")
		}
	}

	// Write hook
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return err
	}

	return nil
}

// UninstallHook removes the pre-commit hook
func UninstallHook(dir string) error {
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")

	// Check if it's our hook
	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to uninstall
		}
		return err
	}

	if !strings.Contains(string(data), hookMarker) {
		return fmt.Errorf("pre-commit hook is not PODX hook")
	}

	return os.Remove(hookPath)
}

// IsHookInstalled returns true if PODX hook is installed
func IsHookInstalled(dir string) bool {
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")

	data, err := os.ReadFile(hookPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(data), hookMarker)
}
