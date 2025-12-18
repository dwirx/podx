package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// ChaChaKeySize adalah ukuran key untuk ChaCha20-Poly1305 (32 bytes)
	ChaChaKeySize = 32
	// ChaChaNonceSize adalah ukuran nonce (12 bytes untuk standard, 24 untuk XChaCha20)
	ChaChaNonceSize = 12
)

// ChaCha20Encrypt mengenkripsi plaintext menggunakan ChaCha20-Poly1305.
// Nonce di-prepend ke ciphertext.
// Format output: [nonce (12 bytes)][ciphertext+tag]
func ChaCha20Encrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != ChaChaKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ChaChaKeySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt dan prepend nonce
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// ChaCha20Decrypt mendekripsi ciphertext yang dienkripsi dengan ChaCha20-Poly1305.
// Expectation: ciphertext format [nonce (12 bytes)][ciphertext+tag]
func ChaCha20Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != ChaChaKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ChaChaKeySize, len(key))
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce dan ciphertext
	nonce, ciphertextData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := aead.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
