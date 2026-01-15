package keygen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Sample armored GPG keys for testing
const testGPGPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQGNBGcI1TsBDAC8Q5kxvMJZLQ7wYvM1zPrjx4aA9yI5FjZhLqL0lR5vKNd5tZb5
wLv9xZc8yH3dS2nR7vL1aQ4kT6xW8pJ2yH5dN3cL9vR2sK4pL1zT8vX3wJ5cN6h3
=abcd
-----END PGP PUBLIC KEY BLOCK-----`

const testGPGPrivateKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lQcYBGcI1TsBDAC8Q5kxvMJZLQ7wYvM1zPrjx4aA9yI5FjZhLqL0lR5vKNd5tZb5
wLv9xZc8yH3dS2nR7vL1aQ4kT6xW8pJ2yH5dN3cL9vR2sK4pL1zT8vX3wJ5cN6h3
=wxyz
-----END PGP PRIVATE KEY BLOCK-----`

func TestSaveGPGKeyPair(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "podx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override config dir for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test data
	keyID := "ABCD1234EFGH5678"
	name := "Test User"
	email := "test@example.com"

	// Save the key pair
	err = SaveGPGKeyPair(keyID, testGPGPublicKey, testGPGPrivateKey, name, email)
	if err != nil {
		t.Fatalf("SaveGPGKeyPair failed: %v", err)
	}

	// Verify directory was created
	configDir := filepath.Join(tempDir, configDir)
	gpgDir := filepath.Join(configDir, gpgKeysDir)
	if _, err := os.Stat(gpgDir); os.IsNotExist(err) {
		t.Errorf("GPG keys directory was not created: %s", gpgDir)
	}

	// Verify public key file exists
	pubKeyPath := filepath.Join(gpgDir, keyID+"_public.asc")
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		t.Errorf("Public key file was not created: %s", pubKeyPath)
	}

	// Verify public key content
	pubKeyContent, err := os.ReadFile(pubKeyPath)
	if err != nil {
		t.Fatalf("Failed to read public key: %v", err)
	}
	if string(pubKeyContent) != testGPGPublicKey {
		t.Errorf("Public key content mismatch.\nExpected:\n%s\nGot:\n%s", testGPGPublicKey, string(pubKeyContent))
	}

	// Verify public key file permissions (should be 0644)
	pubInfo, err := os.Stat(pubKeyPath)
	if err != nil {
		t.Fatalf("Failed to stat public key: %v", err)
	}
	if pubInfo.Mode().Perm() != 0644 {
		t.Errorf("Public key has wrong permissions: got %o, want 0644", pubInfo.Mode().Perm())
	}

	// Verify private key file exists
	privKeyPath := filepath.Join(gpgDir, keyID+"_private.asc")
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		t.Errorf("Private key file was not created: %s", privKeyPath)
	}

	// Verify private key content
	privKeyContent, err := os.ReadFile(privKeyPath)
	if err != nil {
		t.Fatalf("Failed to read private key: %v", err)
	}
	if string(privKeyContent) != testGPGPrivateKey {
		t.Errorf("Private key content mismatch.\nExpected:\n%s\nGot:\n%s", testGPGPrivateKey, string(privKeyContent))
	}

	// Verify private key file permissions (should be 0600)
	privInfo, err := os.Stat(privKeyPath)
	if err != nil {
		t.Fatalf("Failed to stat private key: %v", err)
	}
	if privInfo.Mode().Perm() != 0600 {
		t.Errorf("Private key has wrong permissions: got %o, want 0600", privInfo.Mode().Perm())
	}

	// Verify metadata file exists
	metaPath := filepath.Join(gpgDir, keyID+"_meta.txt")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("Metadata file was not created: %s", metaPath)
	}

	// Verify metadata content
	metaContent, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}
	metaStr := string(metaContent)
	if !strings.Contains(metaStr, "name: "+name) {
		t.Errorf("Metadata does not contain name. Content:\n%s", metaStr)
	}
	if !strings.Contains(metaStr, "email: "+email) {
		t.Errorf("Metadata does not contain email. Content:\n%s", metaStr)
	}
}

func TestListGPGKeys(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "podx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override config dir for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test 1: Empty directory should return empty list
	keys, err := ListGPGKeys()
	if err != nil {
		t.Fatalf("ListGPGKeys failed on empty dir: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}

	// Test 2: Add multiple keys and list them
	key1ID := "AAAA1111BBBB2222"
	key2ID := "CCCC3333DDDD4444"

	err = SaveGPGKeyPair(key1ID, testGPGPublicKey, testGPGPrivateKey, "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Failed to save key 1: %v", err)
	}

	err = SaveGPGKeyPair(key2ID, testGPGPublicKey, testGPGPrivateKey, "Bob", "bob@example.com")
	if err != nil {
		t.Fatalf("Failed to save key 2: %v", err)
	}

	// List keys
	keys, err = ListGPGKeys()
	if err != nil {
		t.Fatalf("ListGPGKeys failed: %v", err)
	}

	// Verify count
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Verify keys contain correct data
	foundAlice := false
	foundBob := false
	for _, key := range keys {
		if key.KeyID == key1ID && key.Name == "Alice" && key.Email == "alice@example.com" {
			foundAlice = true
			if !strings.HasPrefix(key.PublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
				t.Errorf("Alice's public key format is incorrect")
			}
			if !strings.HasPrefix(key.PrivateKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----") {
				t.Errorf("Alice's private key format is incorrect")
			}
			if key.FilePath == "" {
				t.Errorf("Alice's file path is empty")
			}
		}
		if key.KeyID == key2ID && key.Name == "Bob" && key.Email == "bob@example.com" {
			foundBob = true
			if !strings.HasPrefix(key.PublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
				t.Errorf("Bob's public key format is incorrect")
			}
			if !strings.HasPrefix(key.PrivateKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----") {
				t.Errorf("Bob's private key format is incorrect")
			}
			if key.FilePath == "" {
				t.Errorf("Bob's file path is empty")
			}
		}
	}

	if !foundAlice {
		t.Errorf("Alice's key not found in list")
	}
	if !foundBob {
		t.Errorf("Bob's key not found in list")
	}
}
