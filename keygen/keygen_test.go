package keygen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateGPGNative(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Override config dir for testing by setting HOME env var
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Call GenerateGPG
	result, err := GenerateGPG("Test User", "test@podx.local")
	if err != nil {
		t.Fatalf("GenerateGPG failed: %v", err)
	}

	// Verify result.Backend == "gpg"
	if result.Backend != "gpg" {
		t.Errorf("Expected backend 'gpg', got '%s'", result.Backend)
	}

	// Verify result.Email == "test@podx.local"
	if result.Email != "test@podx.local" {
		t.Errorf("Expected email 'test@podx.local', got '%s'", result.Email)
	}

	// Verify result.PublicKey contains "-----BEGIN PGP PUBLIC KEY BLOCK-----"
	if !strings.Contains(result.PublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
		t.Errorf("Public key does not contain expected header")
	}

	// Verify result.PrivateKey contains "-----BEGIN PGP PRIVATE KEY BLOCK-----"
	if !strings.Contains(result.PrivateKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----") {
		t.Errorf("Private key does not contain expected header")
	}

	// Verify result.Name is set
	if result.Name == "" {
		t.Errorf("Expected name to be set, got empty string")
	}
	if result.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", result.Name)
	}

	// Verify keys are saved in config directory
	configDir := filepath.Join(tempDir, configDir)
	gpgDir := filepath.Join(configDir, gpgKeysDir)

	// Check if GPG directory was created
	if _, err := os.Stat(gpgDir); os.IsNotExist(err) {
		t.Errorf("GPG directory was not created: %s", gpgDir)
	}

	// List saved GPG keys
	keys, err := ListGPGKeys()
	if err != nil {
		t.Fatalf("ListGPGKeys failed: %v", err)
	}

	// Verify at least one key was saved
	if len(keys) == 0 {
		t.Errorf("No GPG keys were saved")
	}

	// Verify the saved key has correct email
	foundKey := false
	for _, key := range keys {
		if key.Email == "test@podx.local" && key.Name == "Test User" {
			foundKey = true
			// Verify public and private keys are saved
			if !strings.Contains(key.PublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
				t.Errorf("Saved public key format is incorrect")
			}
			if !strings.Contains(key.PrivateKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----") {
				t.Errorf("Saved private key format is incorrect")
			}
			break
		}
	}

	if !foundKey {
		t.Errorf("Generated key not found in saved keys")
	}
}
