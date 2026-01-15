package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestAgeEncryptDecrypt(t *testing.T) {
	// Generate a test key pair
	privateKey, publicKey, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("failed to generate age key: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"long", bytes.Repeat([]byte("test data "), 1000)},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界 🌍")},
		{"newlines", []byte("line1\nline2\nline3")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt with public key
			ciphertext, err := AgeEncrypt(tt.plaintext, publicKey)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			// Verify ciphertext is not empty and different from plaintext
			if len(ciphertext) == 0 {
				t.Fatal("ciphertext is empty")
			}
			if len(tt.plaintext) > 0 && bytes.Equal(ciphertext, tt.plaintext) {
				t.Error("ciphertext equals plaintext")
			}

			// Decrypt with private key
			decrypted, err := AgeDecrypt(ciphertext, privateKey)
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

func TestAgeMultipleRecipients(t *testing.T) {
	// Generate two key pairs
	privateKey1, publicKey1, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("failed to generate first key: %v", err)
	}

	privateKey2, publicKey2, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("failed to generate second key: %v", err)
	}

	plaintext := []byte("secret message for multiple recipients")

	// Encrypt for both recipients
	ciphertext, err := AgeEncrypt(plaintext, publicKey1, publicKey2)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	// Both recipients should be able to decrypt
	decrypted1, err := AgeDecrypt(ciphertext, privateKey1)
	if err != nil {
		t.Fatalf("decrypt with key1 error: %v", err)
	}
	if !bytes.Equal(decrypted1, plaintext) {
		t.Error("decryption with key1 failed")
	}

	decrypted2, err := AgeDecrypt(ciphertext, privateKey2)
	if err != nil {
		t.Fatalf("decrypt with key2 error: %v", err)
	}
	if !bytes.Equal(decrypted2, plaintext) {
		t.Error("decryption with key2 failed")
	}
}

func TestAgeInvalidRecipient(t *testing.T) {
	plaintext := []byte("test")

	tests := []struct {
		name      string
		recipient string
	}{
		{"empty", ""},
		{"invalid_format", "invalid-recipient"},
		{"wrong_prefix", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGfB"},
		{"malformed", "age1xyz123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AgeEncrypt(plaintext, tt.recipient)
			if err == nil {
				t.Error("expected error for invalid recipient")
			}
		})
	}
}

func TestAgeInvalidPrivateKey(t *testing.T) {
	// Create valid ciphertext first
	privateKey, publicKey, err := GenerateAgeKey()
	if err != nil {
		t.Fatal(err)
	}

	ciphertext, err := AgeEncrypt([]byte("test"), publicKey)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		privateKey string
	}{
		{"empty", ""},
		{"invalid_format", "invalid-key"},
		{"wrong_key", "AGE-SECRET-KEY-1AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AgeDecrypt(ciphertext, tt.privateKey)
			if err == nil {
				t.Error("expected error for invalid private key")
			}
		})
	}

	// Test with wrong but valid key
	t.Run("wrong_valid_key", func(t *testing.T) {
		wrongPrivateKey, _, err := GenerateAgeKey()
		if err != nil {
			t.Fatal(err)
		}

		_, err = AgeDecrypt(ciphertext, wrongPrivateKey)
		if err == nil {
			t.Error("expected error when decrypting with wrong key")
		}
	})

	// Ensure correct key still works
	t.Run("correct_key", func(t *testing.T) {
		decrypted, err := AgeDecrypt(ciphertext, privateKey)
		if err != nil {
			t.Fatalf("decrypt with correct key failed: %v", err)
		}
		if !bytes.Equal(decrypted, []byte("test")) {
			t.Error("decryption with correct key produced wrong result")
		}
	})
}

func TestAgeDecryptTamperedData(t *testing.T) {
	privateKey, publicKey, err := GenerateAgeKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("secret data")
	ciphertext, err := AgeEncrypt(plaintext, publicKey)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with different parts of the ciphertext
	tests := []struct {
		name   string
		tamper func([]byte) []byte
	}{
		{"flip_first_byte", func(ct []byte) []byte {
			ct[0] ^= 0xff
			return ct
		}},
		{"flip_last_byte", func(ct []byte) []byte {
			ct[len(ct)-1] ^= 0xff
			return ct
		}},
		{"flip_middle_byte", func(ct []byte) []byte {
			ct[len(ct)/2] ^= 0xff
			return ct
		}},
		{"truncate", func(ct []byte) []byte {
			return ct[:len(ct)-10]
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tampered := make([]byte, len(ciphertext))
			copy(tampered, ciphertext)
			tampered = tt.tamper(tampered)

			_, err := AgeDecrypt(tampered, privateKey)
			if err == nil {
				t.Error("expected error for tampered ciphertext")
			}
		})
	}
}

func TestGenerateAgeKey(t *testing.T) {
	privateKey, publicKey, err := GenerateAgeKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Verify key formats
	if !strings.HasPrefix(privateKey, "AGE-SECRET-KEY-") {
		t.Errorf("private key has wrong prefix: %s", privateKey)
	}

	if !strings.HasPrefix(publicKey, "age1") {
		t.Errorf("public key has wrong prefix: %s", publicKey)
	}

	// Verify keys are not empty
	if len(privateKey) == 0 || len(publicKey) == 0 {
		t.Error("generated empty key")
	}

	// Verify keys can be used for encryption/decryption
	plaintext := []byte("test message")
	ciphertext, err := AgeEncrypt(plaintext, publicKey)
	if err != nil {
		t.Fatalf("encrypt with generated key failed: %v", err)
	}

	decrypted, err := AgeDecrypt(ciphertext, privateKey)
	if err != nil {
		t.Fatalf("decrypt with generated key failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("encrypt/decrypt with generated keys failed")
	}
}

func TestDeriveAgePublicKey(t *testing.T) {
	privateKey, expectedPublic, err := GenerateAgeKey()
	if err != nil {
		t.Fatal(err)
	}

	derivedPublic, err := DeriveAgePublicKey(privateKey)
	if err != nil {
		t.Fatalf("failed to derive public key: %v", err)
	}

	if derivedPublic != expectedPublic {
		t.Errorf("derived public key %s doesn't match expected %s", derivedPublic, expectedPublic)
	}
}

func TestDeriveAgePublicKeyInvalid(t *testing.T) {
	tests := []struct {
		name       string
		privateKey string
	}{
		{"empty", ""},
		{"invalid", "invalid-key"},
		{"public_key", "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeriveAgePublicKey(tt.privateKey)
			if err == nil {
				t.Error("expected error for invalid private key")
			}
		})
	}
}

func TestAgeEncryptDecryptLargeData(t *testing.T) {
	privateKey, publicKey, err := GenerateAgeKey()
	if err != nil {
		t.Fatal(err)
	}

	// Test with 1MB of data
	plaintext := bytes.Repeat([]byte("Large data test. "), 64*1024)

	ciphertext, err := AgeEncrypt(plaintext, publicKey)
	if err != nil {
		t.Fatalf("encrypt large data failed: %v", err)
	}

	decrypted, err := AgeDecrypt(ciphertext, privateKey)
	if err != nil {
		t.Fatalf("decrypt large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("large data encryption/decryption mismatch")
	}
}

func TestAgeKeyUniqueness(t *testing.T) {
	// Generate multiple keys and ensure they're unique
	keys := make(map[string]bool)
	iterations := 10

	for i := 0; i < iterations; i++ {
		privateKey, publicKey, err := GenerateAgeKey()
		if err != nil {
			t.Fatalf("iteration %d: failed to generate key: %v", i, err)
		}

		if keys[privateKey] {
			t.Errorf("duplicate private key generated: %s", privateKey)
		}
		if keys[publicKey] {
			t.Errorf("duplicate public key generated: %s", publicKey)
		}

		keys[privateKey] = true
		keys[publicKey] = true
	}

	if len(keys) != iterations*2 {
		t.Errorf("expected %d unique keys, got %d", iterations*2, len(keys))
	}
}

// Benchmark Age encryption
func BenchmarkAgeEncrypt(b *testing.B) {
	_, publicKey, err := GenerateAgeKey()
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AgeEncrypt(plaintext, publicKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgeDecrypt(b *testing.B) {
	privateKey, publicKey, err := GenerateAgeKey()
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("benchmark test data")
	ciphertext, err := AgeEncrypt(plaintext, publicKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AgeDecrypt(ciphertext, privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgeKeyGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := GenerateAgeKey()
		if err != nil {
			b.Fatal(err)
		}
	}
}
