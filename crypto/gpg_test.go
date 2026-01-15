package crypto

import (
	"bytes"
	"testing"
)

func TestCheckGPGInstalled(t *testing.T) {
	installed := CheckGPGInstalled()
	if !installed {
		t.Skip("GPG not installed, skipping GPG tests")
	}
}

func TestGPGEncryptDecrypt(t *testing.T) {
	if !CheckGPGInstalled() {
		t.Skip("GPG not installed")
	}

	// Generate test key
	email := "test@podx.local"
	recipient, err := GenerateGPGKey("Test User", email, "")
	if err != nil {
		t.Skipf("Could not generate GPG key (may require interactive GPG): %v", err)
	}

	defer func() {
		// Cleanup: delete test key (best effort)
		// exec.Command("gpg", "--batch", "--yes", "--delete-secret-and-public-key", recipient).Run()
	}()

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello gpg")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog with GPG")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界 🌍 with GPG")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := GPGEncrypt(tt.plaintext, recipient)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			// Verify ciphertext is not empty
			if len(ciphertext) == 0 {
				t.Fatal("ciphertext is empty")
			}

			// Verify it's in ASCII armor format
			if !bytes.Contains(ciphertext, []byte("-----BEGIN PGP MESSAGE-----")) {
				t.Error("ciphertext not in ASCII armor format")
			}

			// Decrypt
			decrypted, err := GPGDecrypt(ciphertext)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			// Verify decrypted matches original
			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestGPGEncryptInvalidRecipient(t *testing.T) {
	if !CheckGPGInstalled() {
		t.Skip("GPG not installed")
	}

	plaintext := []byte("test")

	tests := []struct {
		name      string
		recipient string
	}{
		{"nonexistent", "nonexistent@example.com"},
		{"invalid", "invalid-recipient-format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GPGEncrypt(plaintext, tt.recipient)
			if err == nil {
				t.Error("expected error for invalid recipient")
			}
		})
	}
}

func TestGPGDecryptInvalidData(t *testing.T) {
	if !CheckGPGInstalled() {
		t.Skip("GPG not installed")
	}

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"empty", []byte{}},
		{"invalid", []byte("not a valid pgp message")},
		{"malformed_armor", []byte("-----BEGIN PGP MESSAGE-----\ninvalid\n-----END PGP MESSAGE-----")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GPGDecrypt(tt.ciphertext)
			if err == nil {
				t.Error("expected error for invalid ciphertext")
			}
		})
	}
}

func TestGPGDecryptTamperedData(t *testing.T) {
	if !CheckGPGInstalled() {
		t.Skip("GPG not installed")
	}

	// Generate test key
	email := "tamper-test@podx.local"
	recipient, err := GenerateGPGKey("Tamper Test", email, "")
	if err != nil {
		t.Skipf("Could not generate GPG key: %v", err)
	}

	plaintext := []byte("secret data")
	ciphertext, err := GPGEncrypt(plaintext, recipient)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Tamper with the ASCII armor
	tampered := bytes.Replace(ciphertext, []byte("A"), []byte("B"), 1)

	_, err = GPGDecrypt(tampered)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestGenerateGPGKey(t *testing.T) {
	if !CheckGPGInstalled() {
		t.Skip("GPG not installed")
	}

	tests := []struct {
		name  string
		uname string
		email string
	}{
		{"basic", "GPG Test User", "gpgtest@podx.local"},
		{"with_numbers", "User123", "user123@test.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyID, err := GenerateGPGKey(tt.uname, tt.email, "")
			if err != nil {
				t.Skipf("GPG key generation failed (may require interactive setup): %v", err)
			}

			if keyID == "" {
				t.Error("key ID is empty")
			}

			// Verify the key can be used
			plaintext := []byte("test with new key")
			ciphertext, err := GPGEncrypt(plaintext, keyID)
			if err != nil {
				t.Errorf("encrypt with new key failed: %v", err)
			}

			decrypted, err := GPGDecrypt(ciphertext)
			if err != nil {
				t.Errorf("decrypt with new key failed: %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Error("encrypt/decrypt with new key failed")
			}
		})
	}
}

func TestGPGLargeData(t *testing.T) {
	if !CheckGPGInstalled() {
		t.Skip("GPG not installed")
	}

	// Generate test key
	email := "large-data@podx.local"
	recipient, err := GenerateGPGKey("Large Data Test", email, "")
	if err != nil {
		t.Skipf("Could not generate GPG key: %v", err)
	}

	// Test with 1MB of data
	plaintext := bytes.Repeat([]byte("Large GPG data test. "), 50*1024)

	ciphertext, err := GPGEncrypt(plaintext, recipient)
	if err != nil {
		t.Fatalf("encrypt large data failed: %v", err)
	}

	decrypted, err := GPGDecrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("large data encryption/decryption mismatch")
	}
}

// Benchmark GPG operations (only if GPG is installed)
func BenchmarkGPGEncrypt(b *testing.B) {
	if !CheckGPGInstalled() {
		b.Skip("GPG not installed")
	}

	// Setup: generate key
	email := "bench@podx.local"
	recipient, err := GenerateGPGKey("Benchmark User", email, "")
	if err != nil {
		b.Skipf("Could not generate GPG key: %v", err)
	}

	plaintext := []byte("benchmark test data for GPG")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GPGEncrypt(plaintext, recipient)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGPGDecrypt(b *testing.B) {
	if !CheckGPGInstalled() {
		b.Skip("GPG not installed")
	}

	// Setup: generate key and encrypt data
	email := "bench-decrypt@podx.local"
	recipient, err := GenerateGPGKey("Benchmark Decrypt", email, "")
	if err != nil {
		b.Skipf("Could not generate GPG key: %v", err)
	}

	plaintext := []byte("benchmark test data for GPG")
	ciphertext, err := GPGEncrypt(plaintext, recipient)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GPGDecrypt(ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGenerateGPGKeyNative(t *testing.T) {
	// Test native key generation without external GPG
	keyID, privateKey, publicKey, err := GenerateGPGKeyNative("Test User", "test@podx.local", "")
	if err != nil {
		t.Fatalf("native key generation failed: %v", err)
	}

	if keyID == "" {
		t.Error("key ID is empty")
	}

	if privateKey == "" {
		t.Error("private key is empty")
	}

	if publicKey == "" {
		t.Error("public key is empty")
	}

	// Verify key format
	if !bytes.Contains([]byte(privateKey), []byte("-----BEGIN PGP PRIVATE KEY BLOCK-----")) {
		t.Error("private key not in armored format")
	}

	if !bytes.Contains([]byte(publicKey), []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----")) {
		t.Error("public key not in armored format")
	}
}

func TestGPGEncryptNative(t *testing.T) {
	// Generate test key
	_, _, publicKey, err := GenerateGPGKeyNative("Test User", "test@podx.local", "")
	if err != nil {
		t.Fatalf("key generation failed: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello native gpg")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界 🌍")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := GPGEncryptNative(tt.plaintext, publicKey)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) == 0 {
				t.Fatal("ciphertext is empty")
			}

			// Verify ASCII armor format
			if !bytes.Contains(ciphertext, []byte("-----BEGIN PGP MESSAGE-----")) {
				t.Error("ciphertext not in ASCII armor format")
			}
		})
	}
}

func TestGPGEncryptDecryptNative(t *testing.T) {
	// Generate test key without passphrase
	_, privateKey, publicKey, err := GenerateGPGKeyNative("Test User", "test@podx.local", "")
	if err != nil {
		t.Fatalf("key generation failed: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello native gpg")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界 🌍")},
		{"large", bytes.Repeat([]byte("GPG native encryption test. "), 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := GPGEncryptNative(tt.plaintext, publicKey)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) == 0 {
				t.Fatal("ciphertext is empty")
			}

			// Verify ASCII armor format
			if !bytes.Contains(ciphertext, []byte("-----BEGIN PGP MESSAGE-----")) {
				t.Error("ciphertext not in ASCII armor format")
			}

			// Decrypt
			decrypted, err := GPGDecryptNative(ciphertext, privateKey, "")
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			// Verify decrypted matches original
			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("decrypted data mismatch: got %d bytes, want %d bytes", len(decrypted), len(tt.plaintext))
			}
		})
	}
}

func TestGPGEncryptMultipleRecipientsNative(t *testing.T) {
	// Generate two test keys
	_, privateKey1, publicKey1, err := GenerateGPGKeyNative("User 1", "user1@podx.local", "")
	if err != nil {
		t.Fatalf("key 1 generation failed: %v", err)
	}

	_, privateKey2, publicKey2, err := GenerateGPGKeyNative("User 2", "user2@podx.local", "")
	if err != nil {
		t.Fatalf("key 2 generation failed: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"short", []byte("multi-recipient test")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界 🌍")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt for both recipients
			publicKeys := []string{publicKey1, publicKey2}
			ciphertext, err := GPGEncryptMultipleNative(tt.plaintext, publicKeys)
			if err != nil {
				t.Fatalf("multi-recipient encrypt error: %v", err)
			}

			if len(ciphertext) == 0 {
				t.Fatal("ciphertext is empty")
			}

			// Verify ASCII armor format
			if !bytes.Contains(ciphertext, []byte("-----BEGIN PGP MESSAGE-----")) {
				t.Error("ciphertext not in ASCII armor format")
			}

			// Both recipients should be able to decrypt
			decrypted1, err := GPGDecryptNative(ciphertext, privateKey1, "")
			if err != nil {
				t.Fatalf("recipient 1 decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted1, tt.plaintext) {
				t.Error("recipient 1: decrypted data mismatch")
			}

			decrypted2, err := GPGDecryptNative(ciphertext, privateKey2, "")
			if err != nil {
				t.Fatalf("recipient 2 decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted2, tt.plaintext) {
				t.Error("recipient 2: decrypted data mismatch")
			}
		})
	}
}
