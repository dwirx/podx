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
	// ChaChaNonceSize adalah ukuran nonce (12 bytes untuk standard)
	ChaChaNonceSize = 12
	// XChaChaNonceSize adalah ukuran nonce untuk XChaCha20 (24 bytes)
	XChaChaNonceSize = 24
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

// XChaCha20Encrypt mengenkripsi plaintext menggunakan XChaCha20-Poly1305.
// XChaCha20 menggunakan nonce 24 bytes (lebih besar dari ChaCha20 12 bytes)
// yang aman untuk random nonce generation.
// Format output: [nonce (24 bytes)][ciphertext+tag]
func XChaCha20Encrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != ChaChaKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ChaChaKeySize, len(key))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create XChaCha20-Poly1305: %w", err)
	}

	// Generate random nonce (24 bytes for XChaCha20)
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt dan prepend nonce
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// XChaCha20Decrypt mendekripsi ciphertext yang dienkripsi dengan XChaCha20-Poly1305.
// Expectation: ciphertext format [nonce (24 bytes)][ciphertext+tag]
func XChaCha20Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != ChaChaKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ChaChaKeySize, len(key))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create XChaCha20-Poly1305: %w", err)
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

// XChaCha20EncryptWithNonce mengenkripsi dengan nonce yang sudah ditentukan
// Digunakan untuk streaming encryption dengan derived nonces
func XChaCha20EncryptWithNonce(plaintext, key, nonce []byte) ([]byte, error) {
	if len(key) != ChaChaKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ChaChaKeySize, len(key))
	}

	if len(nonce) != XChaChaNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", XChaChaNonceSize, len(nonce))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create XChaCha20-Poly1305: %w", err)
	}

	// Encrypt without prepending nonce (caller manages nonce)
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

// XChaCha20DecryptWithNonce mendekripsi dengan nonce yang sudah ditentukan
func XChaCha20DecryptWithNonce(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(key) != ChaChaKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ChaChaKeySize, len(key))
	}

	if len(nonce) != XChaChaNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", XChaChaNonceSize, len(nonce))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create XChaCha20-Poly1305: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
