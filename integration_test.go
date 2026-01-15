package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hades/podx/crypto"
	"github.com/hades/podx/keygen"
)

// TestGPGIntegration tests the complete GPG workflow end-to-end
// This test verifies:
// 1. GPG key generation using native implementation
// 2. File encryption with GPG
// 3. File decryption with GPG
// 4. Data integrity (decrypted matches original)
func TestGPGIntegration(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// 1. Generate GPG key using keygen package
	t.Log("Generating GPG key pair...")
	result, err := keygen.GenerateGPG("Test User", "test@podx.local")
	if err != nil {
		t.Fatalf("keygen.GenerateGPG failed: %v", err)
	}

	// Verify key generation result
	if result.Backend != "gpg" {
		t.Errorf("expected backend 'gpg', got '%s'", result.Backend)
	}
	if result.PublicKey == "" {
		t.Error("public key is empty")
	}
	if result.PrivateKey == "" {
		t.Error("private key is empty")
	}
	if result.Email != "test@podx.local" {
		t.Errorf("expected email 'test@podx.local', got '%s'", result.Email)
	}
	if result.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", result.Name)
	}

	// 2. Create test file with plaintext
	t.Log("Creating test file...")
	plaintext := []byte("Secret data for GPG integration test\nLine 2: 世界\nLine 3: Binary data follows: \x00\x01\xff\xfe")
	testFile := filepath.Join(tmpDir, "test-secret.txt")
	if err := os.WriteFile(testFile, plaintext, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 3. Encrypt file with GPG using native implementation
	t.Log("Encrypting file with GPG...")
	ciphertext, err := crypto.GPGEncryptNative(plaintext, result.PublicKey)
	if err != nil {
		t.Fatalf("GPGEncryptNative failed: %v", err)
	}

	// Verify ciphertext properties
	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext equals plaintext (no encryption occurred)")
	}
	if !bytes.Contains(ciphertext, []byte("-----BEGIN PGP MESSAGE-----")) {
		t.Error("ciphertext is not in PGP armor format")
	}

	// 4. Write encrypted file
	encFile := testFile + ".podx"
	if err := os.WriteFile(encFile, ciphertext, 0644); err != nil {
		t.Fatalf("failed to write encrypted file: %v", err)
	}

	// Verify encrypted file exists and is readable
	if _, err := os.Stat(encFile); os.IsNotExist(err) {
		t.Fatal("encrypted file was not created")
	}

	// 5. Decrypt file with GPG using native implementation
	t.Log("Decrypting file with GPG...")
	decrypted, err := crypto.GPGDecryptNative(ciphertext, result.PrivateKey, "")
	if err != nil {
		t.Fatalf("GPGDecryptNative failed: %v", err)
	}

	// 6. Verify decrypted data matches original plaintext
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted data mismatch:\ngot:  %q\nwant: %q", decrypted, plaintext)
	}

	t.Log("GPG integration test PASSED")
}

// TestAgeIntegration tests the complete Age workflow end-to-end
// This test verifies:
// 1. Age key generation
// 2. File encryption with Age
// 3. File decryption with Age
// 4. Data integrity (decrypted matches original)
func TestAgeIntegration(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// 1. Generate Age key using keygen package
	t.Log("Generating Age key pair...")
	result, err := keygen.GenerateAge()
	if err != nil {
		t.Fatalf("keygen.GenerateAge failed: %v", err)
	}

	// Verify key generation result
	if result.Backend != "age" {
		t.Errorf("expected backend 'age', got '%s'", result.Backend)
	}
	if result.PublicKey == "" {
		t.Error("public key is empty")
	}
	if result.PrivateKey == "" {
		t.Error("private key is empty")
	}

	// 2. Create test file with plaintext
	t.Log("Creating test file...")
	plaintext := []byte("Secret data for Age integration test\nLine 2: 世界\nLine 3: Binary data follows: \x00\x01\xff\xfe")
	testFile := filepath.Join(tmpDir, "test-secret.txt")
	if err := os.WriteFile(testFile, plaintext, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 3. Encrypt file with Age
	t.Log("Encrypting file with Age...")
	ciphertext, err := crypto.AgeEncrypt(plaintext, result.PublicKey)
	if err != nil {
		t.Fatalf("AgeEncrypt failed: %v", err)
	}

	// Verify ciphertext properties
	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext equals plaintext (no encryption occurred)")
	}

	// 4. Write encrypted file
	encFile := testFile + ".podx"
	if err := os.WriteFile(encFile, ciphertext, 0644); err != nil {
		t.Fatalf("failed to write encrypted file: %v", err)
	}

	// Verify encrypted file exists and is readable
	if _, err := os.Stat(encFile); os.IsNotExist(err) {
		t.Fatal("encrypted file was not created")
	}

	// 5. Decrypt file with Age
	t.Log("Decrypting file with Age...")
	decrypted, err := crypto.AgeDecrypt(ciphertext, result.PrivateKey)
	if err != nil {
		t.Fatalf("AgeDecrypt failed: %v", err)
	}

	// 6. Verify decrypted data matches original plaintext
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted data mismatch:\ngot:  %q\nwant: %q", decrypted, plaintext)
	}

	t.Log("Age integration test PASSED")
}

// TestGPGMultiRecipientIntegration tests GPG encryption with multiple recipients
func TestGPGMultiRecipientIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Generate two GPG keys
	t.Log("Generating GPG key pair 1...")
	result1, err := keygen.GenerateGPG("User 1", "user1@podx.local")
	if err != nil {
		t.Fatalf("keygen.GenerateGPG failed for user 1: %v", err)
	}

	t.Log("Generating GPG key pair 2...")
	result2, err := keygen.GenerateGPG("User 2", "user2@podx.local")
	if err != nil {
		t.Fatalf("keygen.GenerateGPG failed for user 2: %v", err)
	}

	// Create test data
	plaintext := []byte("Multi-recipient secret message")

	// Encrypt for both recipients
	t.Log("Encrypting for multiple recipients...")
	publicKeys := []string{result1.PublicKey, result2.PublicKey}
	ciphertext, err := crypto.GPGEncryptMultipleNative(plaintext, publicKeys)
	if err != nil {
		t.Fatalf("GPGEncryptMultipleNative failed: %v", err)
	}

	// Both recipients should be able to decrypt
	t.Log("Decrypting with recipient 1's key...")
	decrypted1, err := crypto.GPGDecryptNative(ciphertext, result1.PrivateKey, "")
	if err != nil {
		t.Fatalf("decryption with key 1 failed: %v", err)
	}
	if !bytes.Equal(decrypted1, plaintext) {
		t.Error("decryption with key 1: data mismatch")
	}

	t.Log("Decrypting with recipient 2's key...")
	decrypted2, err := crypto.GPGDecryptNative(ciphertext, result2.PrivateKey, "")
	if err != nil {
		t.Fatalf("decryption with key 2 failed: %v", err)
	}
	if !bytes.Equal(decrypted2, plaintext) {
		t.Error("decryption with key 2: data mismatch")
	}

	t.Log("GPG multi-recipient integration test PASSED")
}

// TestAgeMultiRecipientIntegration tests Age encryption with multiple recipients
func TestAgeMultiRecipientIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Generate two Age keys
	t.Log("Generating Age key pair 1...")
	result1, err := keygen.GenerateAge()
	if err != nil {
		t.Fatalf("keygen.GenerateAge failed for key 1: %v", err)
	}

	t.Log("Generating Age key pair 2...")
	result2, err := keygen.GenerateAge()
	if err != nil {
		t.Fatalf("keygen.GenerateAge failed for key 2: %v", err)
	}

	// Create test data
	plaintext := []byte("Multi-recipient Age secret message")

	// Encrypt for both recipients
	t.Log("Encrypting for multiple recipients...")
	ciphertext, err := crypto.AgeEncrypt(plaintext, result1.PublicKey, result2.PublicKey)
	if err != nil {
		t.Fatalf("AgeEncrypt with multiple recipients failed: %v", err)
	}

	// Both recipients should be able to decrypt
	t.Log("Decrypting with recipient 1's key...")
	decrypted1, err := crypto.AgeDecrypt(ciphertext, result1.PrivateKey)
	if err != nil {
		t.Fatalf("decryption with key 1 failed: %v", err)
	}
	if !bytes.Equal(decrypted1, plaintext) {
		t.Error("decryption with key 1: data mismatch")
	}

	t.Log("Decrypting with recipient 2's key...")
	decrypted2, err := crypto.AgeDecrypt(ciphertext, result2.PrivateKey)
	if err != nil {
		t.Fatalf("decryption with key 2 failed: %v", err)
	}
	if !bytes.Equal(decrypted2, plaintext) {
		t.Error("decryption with key 2: data mismatch")
	}

	t.Log("Age multi-recipient integration test PASSED")
}

// TestGPGPasswordProtectedIntegration tests GPG with password-protected keys
func TestGPGPasswordProtectedIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	passphrase := "super-secret-password"

	// Generate password-protected GPG key
	t.Log("Generating password-protected GPG key...")
	keyID, privateKey, publicKey, err := crypto.GenerateGPGKeyNative("Protected User", "protected@podx.local", passphrase)
	if err != nil {
		t.Fatalf("GenerateGPGKeyNative with passphrase failed: %v", err)
	}

	if keyID == "" || privateKey == "" || publicKey == "" {
		t.Fatal("key generation returned empty keys")
	}

	// Create test data
	plaintext := []byte("Password-protected secret")

	// Encrypt
	t.Log("Encrypting with password-protected key...")
	ciphertext, err := crypto.GPGEncryptNative(plaintext, publicKey)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt with correct passphrase
	t.Log("Decrypting with correct passphrase...")
	decrypted, err := crypto.GPGDecryptNative(ciphertext, privateKey, passphrase)
	if err != nil {
		t.Fatalf("decryption with correct passphrase failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decrypted data mismatch")
	}

	// Decrypt with wrong passphrase should fail
	t.Log("Verifying wrong passphrase fails...")
	_, err = crypto.GPGDecryptNative(ciphertext, privateKey, "wrong-password")
	if err == nil {
		t.Error("expected error when decrypting with wrong passphrase, got nil")
	}

	// Decrypt with empty passphrase should fail
	t.Log("Verifying empty passphrase fails...")
	_, err = crypto.GPGDecryptNative(ciphertext, privateKey, "")
	if err == nil {
		t.Error("expected error when decrypting locked key with empty passphrase, got nil")
	}

	t.Log("GPG password-protected integration test PASSED")
}

// TestLargeFileIntegration tests encryption/decryption of large files
func TestLargeFileIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Generate keys
	t.Log("Generating keys for large file test...")
	gpgResult, err := keygen.GenerateGPG("Large File User", "largefile@podx.local")
	if err != nil {
		t.Fatalf("GPG key generation failed: %v", err)
	}

	ageResult, err := keygen.GenerateAge()
	if err != nil {
		t.Fatalf("Age key generation failed: %v", err)
	}

	// Create large test data (1MB)
	t.Log("Creating large test file (1MB)...")
	plaintext := bytes.Repeat([]byte("This is a test of large file encryption. 世界 🌍 "), 20*1024) // ~1MB

	// Test GPG with large file
	t.Run("GPG_large_file", func(t *testing.T) {
		t.Log("Encrypting large file with GPG...")
		ciphertext, err := crypto.GPGEncryptNative(plaintext, gpgResult.PublicKey)
		if err != nil {
			t.Fatalf("GPG encryption of large file failed: %v", err)
		}

		t.Log("Decrypting large file with GPG...")
		decrypted, err := crypto.GPGDecryptNative(ciphertext, gpgResult.PrivateKey, "")
		if err != nil {
			t.Fatalf("GPG decryption of large file failed: %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Error("GPG large file: decrypted data mismatch")
		}
		t.Log("GPG large file test PASSED")
	})

	// Test Age with large file
	t.Run("Age_large_file", func(t *testing.T) {
		t.Log("Encrypting large file with Age...")
		ciphertext, err := crypto.AgeEncrypt(plaintext, ageResult.PublicKey)
		if err != nil {
			t.Fatalf("Age encryption of large file failed: %v", err)
		}

		t.Log("Decrypting large file with Age...")
		decrypted, err := crypto.AgeDecrypt(ciphertext, ageResult.PrivateKey)
		if err != nil {
			t.Fatalf("Age decryption of large file failed: %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Error("Age large file: decrypted data mismatch")
		}
		t.Log("Age large file test PASSED")
	})
}

// TestCrossCompatibilityGPG tests that GPG auto-detects and uses native implementation
func TestCrossCompatibilityGPG(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Generate GPG key
	t.Log("Generating GPG key for cross-compatibility test...")
	_, privateKey, publicKey, err := crypto.GenerateGPGKeyNative("Cross Test", "cross@podx.local", "")
	if err != nil {
		t.Fatalf("key generation failed: %v", err)
	}

	plaintext := []byte("Cross-compatibility test")

	// Test GPGEncrypt auto-detection (should use native when given armored key)
	t.Log("Testing GPGEncrypt auto-detection...")
	ciphertext, err := crypto.GPGEncrypt(plaintext, publicKey)
	if err != nil {
		t.Fatalf("GPGEncrypt failed: %v", err)
	}

	// Decrypt with GPGDecryptWithKey (should use native when privateKey provided)
	t.Log("Testing GPGDecryptWithKey auto-detection...")
	decrypted, err := crypto.GPGDecryptWithKey(ciphertext, privateKey, "")
	if err != nil {
		t.Fatalf("GPGDecryptWithKey failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("cross-compatibility: decrypted data mismatch")
	}

	t.Log("GPG cross-compatibility test PASSED")
}

// TestErrorHandlingIntegration tests error cases in integration scenarios
func TestErrorHandlingIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Test decryption with wrong key
	t.Run("GPG_wrong_key", func(t *testing.T) {
		_, _, publicKey1, _ := crypto.GenerateGPGKeyNative("User 1", "user1@podx.local", "")
		_, privateKey2, _, _ := crypto.GenerateGPGKeyNative("User 2", "user2@podx.local", "")

		plaintext := []byte("secret")
		ciphertext, _ := crypto.GPGEncryptNative(plaintext, publicKey1)

		_, err := crypto.GPGDecryptNative(ciphertext, privateKey2, "")
		if err == nil {
			t.Error("expected error when decrypting with wrong GPG key")
		}
	})

	t.Run("Age_wrong_key", func(t *testing.T) {
		_, publicKey1, _ := crypto.GenerateAgeKey()
		privateKey2, _, _ := crypto.GenerateAgeKey()

		plaintext := []byte("secret")
		ciphertext, _ := crypto.AgeEncrypt(plaintext, publicKey1)

		_, err := crypto.AgeDecrypt(ciphertext, privateKey2)
		if err == nil {
			t.Error("expected error when decrypting with wrong Age key")
		}
	})

	// Test tampered ciphertext
	t.Run("GPG_tampered", func(t *testing.T) {
		_, privateKey, publicKey, _ := crypto.GenerateGPGKeyNative("Test", "test@podx.local", "")
		plaintext := []byte("secret")
		ciphertext, _ := crypto.GPGEncryptNative(plaintext, publicKey)

		// Tamper with ciphertext
		tampered := make([]byte, len(ciphertext))
		copy(tampered, ciphertext)
		tampered[len(tampered)/2] ^= 0xFF

		_, err := crypto.GPGDecryptNative(tampered, privateKey, "")
		if err == nil {
			t.Error("expected error when decrypting tampered GPG ciphertext")
		}
	})

	t.Run("Age_tampered", func(t *testing.T) {
		privateKey, publicKey, _ := crypto.GenerateAgeKey()
		plaintext := []byte("secret")
		ciphertext, _ := crypto.AgeEncrypt(plaintext, publicKey)

		// Tamper with ciphertext
		tampered := make([]byte, len(ciphertext))
		copy(tampered, ciphertext)
		tampered[len(tampered)/2] ^= 0xFF

		_, err := crypto.AgeDecrypt(tampered, privateKey)
		if err == nil {
			t.Error("expected error when decrypting tampered Age ciphertext")
		}
	})

	t.Log("Error handling integration tests PASSED")
}
