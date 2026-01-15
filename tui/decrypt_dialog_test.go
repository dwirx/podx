package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hades/podx/project"
)

func TestDetectGPGEncryption(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantGPG   bool
	}{
		{
			name: "GPG armored message",
			content: `-----BEGIN PGP MESSAGE-----

hQEMA1234567890ABC
...encrypted data...
-----END PGP MESSAGE-----`,
			wantGPG: true,
		},
		{
			name:    "Non-GPG encrypted file",
			content: "some random binary data\x00\x01\x02",
			wantGPG: false,
		},
		{
			name:    "Plain text file",
			content: "This is just a plain text file",
			wantGPG: false,
		},
		{
			name:    "Empty file",
			content: "",
			wantGPG: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.podx")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0600)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test detection
			got := detectGPGEncryption(tmpFile)
			if got != tt.wantGPG {
				t.Errorf("detectGPGEncryption() = %v, want %v", got, tt.wantGPG)
			}
		})
	}
}

func TestDetectGPGEncryption_NonExistentFile(t *testing.T) {
	got := detectGPGEncryption("/path/that/does/not/exist.podx")
	if got != false {
		t.Errorf("detectGPGEncryption() for non-existent file = %v, want false", got)
	}
}

func TestShowDecrypt_GPGAutoDetection(t *testing.T) {
	// Create temp GPG-encrypted file
	tmpDir := t.TempDir()
	gpgFile := filepath.Join(tmpDir, "secret.txt.podx")
	gpgContent := []byte("-----BEGIN PGP MESSAGE-----\nhQEMA1234567890ABC\n-----END PGP MESSAGE-----")
	err := os.WriteFile(gpgFile, gpgContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create GPG test file: %v", err)
	}

	// Create encrypt dialog model
	model := NewEncryptDialogModel()

	// Create file info for the GPG file
	files := []FileInfo{
		{
			Name:        "secret.txt.podx",
			Path:        gpgFile,
			IsEncrypted: true,
			IsDir:       false,
		},
	}

	// Call ShowDecrypt
	cmd := model.ShowDecrypt(files, &project.Project{}, 80, 24)

	// Verify that:
	// 1. Method was set to GPG
	if model.method != MethodGPG {
		t.Errorf("ShowDecrypt() method = %v, want %v", model.method, MethodGPG)
	}

	// 2. State was set to Processing
	if model.state != StateProcessing {
		t.Errorf("ShowDecrypt() state = %v, want %v", model.state, StateProcessing)
	}

	// 3. A command was returned (doDecryptGPG)
	if cmd == nil {
		t.Error("ShowDecrypt() returned nil command, want doDecryptGPG command")
	}
}

func TestShowDecrypt_NonGPGFile(t *testing.T) {
	// Create temp non-GPG file
	tmpDir := t.TempDir()
	nonGPGFile := filepath.Join(tmpDir, "secret.txt.podx")
	nonGPGContent := []byte("some password-encrypted content")
	err := os.WriteFile(nonGPGFile, nonGPGContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create non-GPG test file: %v", err)
	}

	// Create encrypt dialog model
	model := NewEncryptDialogModel()

	// Create file info
	files := []FileInfo{
		{
			Name:        "secret.txt.podx",
			Path:        nonGPGFile,
			IsEncrypted: true,
			IsDir:       false,
		},
	}

	// Call ShowDecrypt
	cmd := model.ShowDecrypt(files, &project.Project{}, 80, 24)

	// Verify that:
	// 1. State is still at method selection (not auto-routed to GPG)
	if model.state != StateSelectMethod {
		t.Errorf("ShowDecrypt() state = %v, want %v", model.state, StateSelectMethod)
	}

	// 2. No command was returned
	if cmd != nil {
		t.Error("ShowDecrypt() returned command, want nil for non-GPG files")
	}
}

