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
