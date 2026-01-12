package crypto

import (
	"bytes"
	"testing"
)

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		algo    Algorithm
		wantErr bool
	}{
		{AlgoAESGCM, false},
		{AlgoChaCha20, false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.algo), func(t *testing.T) {
			enc, err := NewEncryptor(tt.algo)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if enc == nil {
				t.Error("encryptor should not be nil")
			}
		})
	}
}

func TestEncryptorRoundtrip(t *testing.T) {
	algos := []Algorithm{AlgoAESGCM, AlgoChaCha20}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("test message for encryptor interface")

	for _, algo := range algos {
		t.Run(string(algo), func(t *testing.T) {
			enc, err := NewEncryptor(algo)
			if err != nil {
				t.Fatal(err)
			}

			ciphertext, err := enc.Encrypt(plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			decrypted, err := enc.Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("got %s, want %s", decrypted, plaintext)
			}
		})
	}
}

func TestEncryptorName(t *testing.T) {
	tests := []struct {
		algo Algorithm
		want string
	}{
		{AlgoAESGCM, "aes-gcm"},
		{AlgoChaCha20, "chacha20"},
	}

	for _, tt := range tests {
		enc, _ := NewEncryptor(tt.algo)
		if got := enc.Name(); got != tt.want {
			t.Errorf("Name() = %s, want %s", got, tt.want)
		}
	}
}

func TestEncryptToBase64(t *testing.T) {
	enc, _ := NewEncryptor(AlgoAESGCM)
	key := make([]byte, 32)
	plaintext := []byte("test")

	b64, err := EncryptToBase64(enc, plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	if len(b64) == 0 {
		t.Error("base64 output should not be empty")
	}

	decrypted, err := DecryptFromBase64(enc, b64, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("got %s, want %s", decrypted, plaintext)
	}
}

func TestDecryptFromBase64InvalidInput(t *testing.T) {
	enc, _ := NewEncryptor(AlgoAESGCM)
	key := make([]byte, 32)

	_, err := DecryptFromBase64(enc, "not-valid-base64!!!", key)
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}
