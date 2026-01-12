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
