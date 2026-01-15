package tui

import (
	"testing"
)

// TestCommandsHasGPGKeygen verifies that the keygen-gpg command exists
// and is properly configured
func TestCommandsHasGPGKeygen(t *testing.T) {
	commands := GetAllCommands()

	var found bool
	var gpgKeygenCmd CommandItem

	for _, item := range commands {
		if cmd, ok := item.(CommandItem); ok {
			if cmd.id == "keygen-gpg" {
				found = true
				gpgKeygenCmd = cmd
				break
			}
		}
	}

	if !found {
		t.Fatal("keygen-gpg command not found in command list")
	}

	// Verify the command is properly configured
	if gpgKeygenCmd.name != "keygen (GPG)" {
		t.Errorf("Expected name 'keygen (GPG)', got '%s'", gpgKeygenCmd.name)
	}

	if gpgKeygenCmd.category != "Keys" {
		t.Errorf("Expected category 'Keys', got '%s'", gpgKeygenCmd.category)
	}

	if !gpgKeygenCmd.needsForm {
		t.Error("Expected needsForm to be true")
	}

	if gpgKeygenCmd.needsConfirm {
		t.Error("Expected needsConfirm to be false")
	}

	// Verify command args
	expectedArgs := []string{"keygen", "-t", "gpg"}
	if len(gpgKeygenCmd.args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(gpgKeygenCmd.args))
	}

	for i, arg := range expectedArgs {
		if i >= len(gpgKeygenCmd.args) || gpgKeygenCmd.args[i] != arg {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, arg, gpgKeygenCmd.args[i])
		}
	}
}

// TestCommandNeedsForm verifies the CommandNeedsForm helper function
func TestCommandNeedsForm(t *testing.T) {
	testCases := []struct {
		commandID string
		expected  bool
	}{
		{"keygen-gpg", true},
		{"keygen-age", false},
		{"add-recipient", true},
		{"encrypt", true},
		{"decrypt", true},
		{"env-encrypt", true},
		{"env-decrypt", true},
		{"sync", true},
		{"init", false},
		{"status", false},
	}

	for _, tc := range testCases {
		t.Run(tc.commandID, func(t *testing.T) {
			result := CommandNeedsForm(tc.commandID)
			if result != tc.expected {
				t.Errorf("CommandNeedsForm(%s) = %v, expected %v", tc.commandID, result, tc.expected)
			}
		})
	}
}
