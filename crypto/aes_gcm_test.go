package crypto

import (
	"bytes"
	"testing"
)

func TestAESGCMEncryptDecrypt(t *testing.T) {
	key := make([]byte, AESKeySize)
	for i := range key {
		key[i] = byte(i)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := AESGCMEncrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) <= AESNonceSize {
				t.Fatal("ciphertext too short")
			}

			decrypted, err := AESGCMDecrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAESGCMInvalidKeySize(t *testing.T) {
	shortKey := make([]byte, 16)
	plaintext := []byte("test")

	_, err := AESGCMEncrypt(plaintext, shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestAESGCMDecryptTamperedData(t *testing.T) {
	key := make([]byte, AESKeySize)
	plaintext := []byte("secret data")

	ciphertext, err := AESGCMEncrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	_, err = AESGCMDecrypt(ciphertext, key)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestAESGCMDecryptTooShort(t *testing.T) {
	key := make([]byte, AESKeySize)
	shortData := make([]byte, AESNonceSize-1)

	_, err := AESGCMDecrypt(shortData, key)
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}

// Additional security tests
func TestAESGCMNonceUniqueness(t *testing.T) {
	key := make([]byte, AESKeySize)
	plaintext := []byte("test nonce uniqueness")

	// Encrypt multiple times
	ciphertexts := make([][]byte, 10)
	for i := range ciphertexts {
		ct, err := AESGCMEncrypt(plaintext, key)
		if err != nil {
			t.Fatal(err)
		}
		ciphertexts[i] = ct
	}

	// All ciphertexts should be different (due to random nonces)
	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if bytes.Equal(ciphertexts[i], ciphertexts[j]) {
				t.Error("duplicate nonce detected")
			}
		}
	}
}

func TestAESGCMLargeData(t *testing.T) {
	key := make([]byte, AESKeySize)
	// Test with 1MB of data
	plaintext := bytes.Repeat([]byte("Large AES-GCM test. "), 50*1024)

	ciphertext, err := AESGCMEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt large data failed: %v", err)
	}

	decrypted, err := AESGCMDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("large data mismatch")
	}
}

func TestAESGCMDifferentKeys(t *testing.T) {
	key1 := make([]byte, AESKeySize)
	key2 := make([]byte, AESKeySize)
	for i := range key2 {
		key2[i] = byte(i + 1)
	}

	plaintext := []byte("test different keys")

	// Encrypt with key1
	ciphertext, err := AESGCMEncrypt(plaintext, key1)
	if err != nil {
		t.Fatal(err)
	}

	// Try to decrypt with key2 (should fail)
	_, err = AESGCMDecrypt(ciphertext, key2)
	if err == nil {
		t.Error("decryption with wrong key should fail")
	}

	// Decrypt with correct key (should succeed)
	decrypted, err := AESGCMDecrypt(ciphertext, key1)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decryption with correct key failed")
	}
}

// Benchmarks
func BenchmarkAESGCMEncrypt(b *testing.B) {
	key := make([]byte, AESKeySize)
	plaintext := []byte("benchmark test data for AES-GCM")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AESGCMEncrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESGCMDecrypt(b *testing.B) {
	key := make([]byte, AESKeySize)
	plaintext := []byte("benchmark test data for AES-GCM")

	ciphertext, err := AESGCMEncrypt(plaintext, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AESGCMDecrypt(ciphertext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESGCMEncryptLarge(b *testing.B) {
	key := make([]byte, AESKeySize)
	plaintext := bytes.Repeat([]byte("test "), 10*1024) // 50KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := AESGCMEncrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

