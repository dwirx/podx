package crypto

import (
	"bytes"
	"testing"
)

func TestChaCha20EncryptDecrypt(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
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
			ciphertext, err := ChaCha20Encrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) <= ChaChaNonceSize {
				t.Fatal("ciphertext too short")
			}

			decrypted, err := ChaCha20Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestChaCha20InvalidKeySize(t *testing.T) {
	shortKey := make([]byte, 16)
	plaintext := []byte("test")

	_, err := ChaCha20Encrypt(plaintext, shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestChaCha20DecryptTamperedData(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("secret data")

	ciphertext, err := ChaCha20Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	_, err = ChaCha20Decrypt(ciphertext, key)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestChaCha20DecryptTooShort(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	shortData := make([]byte, ChaChaNonceSize-1)

	_, err := ChaCha20Decrypt(shortData, key)
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}

// XChaCha20-Poly1305 Tests
func TestXChaCha20EncryptDecrypt(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
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
		{"long", bytes.Repeat([]byte("test data "), 1000)},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
		{"unicode", []byte("Hello 世界 🌍")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := XChaCha20Encrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) <= XChaChaNonceSize {
				t.Fatal("ciphertext too short")
			}

			decrypted, err := XChaCha20Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestXChaCha20InvalidKeySize(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"too_short", 16},
		{"too_long", 64},
		{"zero", 0},
	}

	plaintext := []byte("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := XChaCha20Encrypt(plaintext, key)
			if err == nil {
				t.Error("expected error for invalid key size")
			}
		})
	}
}

func TestXChaCha20DecryptTamperedData(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("secret data for XChaCha20")

	ciphertext, err := XChaCha20Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	_, err = XChaCha20Decrypt(ciphertext, key)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestXChaCha20DecryptTooShort(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	shortData := make([]byte, XChaChaNonceSize-1)

	_, err := XChaCha20Decrypt(shortData, key)
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}

func TestXChaCha20EncryptDecryptWithNonce(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	for i := range key {
		key[i] = byte(i)
	}

	nonce := make([]byte, XChaChaNonceSize)
	for i := range nonce {
		nonce[i] = byte(i * 2)
	}

	plaintext := []byte("test with custom nonce")

	// Encrypt with nonce
	ciphertext, err := XChaCha20EncryptWithNonce(plaintext, key, nonce)
	if err != nil {
		t.Fatalf("encrypt with nonce error: %v", err)
	}

	// Decrypt with nonce
	decrypted, err := XChaCha20DecryptWithNonce(ciphertext, key, nonce)
	if err != nil {
		t.Fatalf("decrypt with nonce error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("got %v, want %v", decrypted, plaintext)
	}
}

func TestXChaCha20WithNonceInvalidNonceSize(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("test")

	tests := []struct {
		name      string
		nonceSize int
	}{
		{"too_short", 12},
		{"too_long", 32},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce := make([]byte, tt.nonceSize)
			_, err := XChaCha20EncryptWithNonce(plaintext, key, nonce)
			if err == nil {
				t.Error("expected error for invalid nonce size")
			}
		})
	}
}

func TestXChaCha20DeterministicWithSameNonce(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	for i := range key {
		key[i] = byte(i)
	}

	nonce := make([]byte, XChaChaNonceSize)
	for i := range nonce {
		nonce[i] = byte(i * 3)
	}

	plaintext := []byte("deterministic test")

	// Encrypt twice with same nonce
	ciphertext1, err := XChaCha20EncryptWithNonce(plaintext, key, nonce)
	if err != nil {
		t.Fatal(err)
	}

	ciphertext2, err := XChaCha20EncryptWithNonce(plaintext, key, nonce)
	if err != nil {
		t.Fatal(err)
	}

	// Should produce identical ciphertext
	if !bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("same nonce should produce identical ciphertext")
	}
}

func TestXChaCha20RandomNonceDifferent(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("test randomness")

	// Encrypt multiple times (should have different nonces)
	ciphertexts := make([][]byte, 5)
	for i := range ciphertexts {
		ct, err := XChaCha20Encrypt(plaintext, key)
		if err != nil {
			t.Fatal(err)
		}
		ciphertexts[i] = ct
	}

	// All ciphertexts should be different (due to random nonces)
	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if bytes.Equal(ciphertexts[i], ciphertexts[j]) {
				t.Error("different encryptions should produce different ciphertexts")
			}
		}
	}
}

// Benchmark XChaCha20
func BenchmarkXChaCha20Encrypt(b *testing.B) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("benchmark test data for XChaCha20")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := XChaCha20Encrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkXChaCha20Decrypt(b *testing.B) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("benchmark test data for XChaCha20")

	ciphertext, err := XChaCha20Encrypt(plaintext, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := XChaCha20Decrypt(ciphertext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark ChaCha20
func BenchmarkChaCha20Encrypt(b *testing.B) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("benchmark test data for ChaCha20")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ChaCha20Encrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChaCha20Decrypt(b *testing.B) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("benchmark test data for ChaCha20")

	ciphertext, err := ChaCha20Encrypt(plaintext, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ChaCha20Decrypt(ciphertext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChaCha20EncryptLarge(b *testing.B) {
	key := make([]byte, ChaChaKeySize)
	plaintext := bytes.Repeat([]byte("test "), 10*1024) // 50KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ChaCha20Encrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

