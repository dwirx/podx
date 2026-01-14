package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/aead/serpent"
)

const (
	// SerpentKeySize adalah ukuran key untuk Serpent-256 (32 bytes)
	SerpentKeySize = 32
	// SerpentBlockSize adalah ukuran block untuk Serpent (16 bytes)
	SerpentBlockSize = 16
	// SerpentNonceSize adalah ukuran nonce/IV untuk CTR mode (16 bytes)
	SerpentNonceSize = 16
)

// SerpentCTREncrypt mengenkripsi plaintext menggunakan Serpent-256 dalam CTR mode.
// Format output: [nonce (16 bytes)][ciphertext]
// Catatan: CTR mode tidak menyediakan authentication, harus dikombinasikan dengan MAC
func SerpentCTREncrypt(plaintext, key []byte) ([]byte, error) {
	if len(key) != SerpentKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", SerpentKeySize, len(key))
	}

	block, err := serpent.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create Serpent cipher: %w", err)
	}

	// Generate random nonce/IV
	nonce := make([]byte, SerpentNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create CTR stream
	stream := cipher.NewCTR(block, nonce)

	// Encrypt
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// Prepend nonce
	result := make([]byte, SerpentNonceSize+len(ciphertext))
	copy(result[:SerpentNonceSize], nonce)
	copy(result[SerpentNonceSize:], ciphertext)

	return result, nil
}

// SerpentCTRDecrypt mendekripsi ciphertext yang dienkripsi dengan Serpent-256 CTR.
// Expectation: ciphertext format [nonce (16 bytes)][ciphertext]
func SerpentCTRDecrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != SerpentKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", SerpentKeySize, len(key))
	}

	if len(ciphertext) < SerpentNonceSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := serpent.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create Serpent cipher: %w", err)
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:SerpentNonceSize]
	data := ciphertext[SerpentNonceSize:]

	// Create CTR stream
	stream := cipher.NewCTR(block, nonce)

	// Decrypt
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)

	return plaintext, nil
}

// SerpentCTREncryptWithNonce mengenkripsi dengan nonce yang sudah ditentukan
// Digunakan untuk streaming encryption dengan derived nonces
func SerpentCTREncryptWithNonce(plaintext, key, nonce []byte) ([]byte, error) {
	if len(key) != SerpentKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", SerpentKeySize, len(key))
	}

	if len(nonce) != SerpentNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", SerpentNonceSize, len(nonce))
	}

	block, err := serpent.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create Serpent cipher: %w", err)
	}

	stream := cipher.NewCTR(block, nonce)

	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// SerpentCTRDecryptWithNonce mendekripsi dengan nonce yang sudah ditentukan
func SerpentCTRDecryptWithNonce(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(key) != SerpentKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", SerpentKeySize, len(key))
	}

	if len(nonce) != SerpentNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", SerpentNonceSize, len(nonce))
	}

	block, err := serpent.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create Serpent cipher: %w", err)
	}

	stream := cipher.NewCTR(block, nonce)

	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// CascadeEncrypt mengenkripsi dengan cascade: XChaCha20-Poly1305 → Serpent-CTR
// Ini adalah enkripsi "paranoid mode" yang menggunakan dua cipher berbeda
// Format output: [XChaCha20 nonce (24)][Serpent nonce (16)][encrypted data]
func CascadeEncrypt(plaintext, xchachaKey, serpentKey []byte) ([]byte, error) {
	// First layer: XChaCha20-Poly1305 (authenticated)
	intermediate, err := XChaCha20Encrypt(plaintext, xchachaKey)
	if err != nil {
		return nil, fmt.Errorf("xchacha20 encryption failed: %w", err)
	}

	// Second layer: Serpent-CTR
	ciphertext, err := SerpentCTREncrypt(intermediate, serpentKey)
	if err != nil {
		return nil, fmt.Errorf("serpent encryption failed: %w", err)
	}

	return ciphertext, nil
}

// CascadeDecrypt mendekripsi cascade: Serpent-CTR → XChaCha20-Poly1305
func CascadeDecrypt(ciphertext, xchachaKey, serpentKey []byte) ([]byte, error) {
	// First layer: decrypt Serpent-CTR
	intermediate, err := SerpentCTRDecrypt(ciphertext, serpentKey)
	if err != nil {
		return nil, fmt.Errorf("serpent decryption failed: %w", err)
	}

	// Second layer: decrypt XChaCha20-Poly1305
	plaintext, err := XChaCha20Decrypt(intermediate, xchachaKey)
	if err != nil {
		return nil, fmt.Errorf("xchacha20 decryption failed: %w", err)
	}

	return plaintext, nil
}

// CascadeEncryptWithNonces mengenkripsi dengan nonces yang sudah ditentukan
// Digunakan untuk streaming encryption
func CascadeEncryptWithNonces(plaintext, xchachaKey, serpentKey, xchaChaNonce, serpentNonce []byte) ([]byte, error) {
	// First layer: XChaCha20-Poly1305
	intermediate, err := XChaCha20EncryptWithNonce(plaintext, xchachaKey, xchaChaNonce)
	if err != nil {
		return nil, fmt.Errorf("xchacha20 encryption failed: %w", err)
	}

	// Second layer: Serpent-CTR
	ciphertext, err := SerpentCTREncryptWithNonce(intermediate, serpentKey, serpentNonce)
	if err != nil {
		return nil, fmt.Errorf("serpent encryption failed: %w", err)
	}

	return ciphertext, nil
}

// CascadeDecryptWithNonces mendekripsi dengan nonces yang sudah ditentukan
func CascadeDecryptWithNonces(ciphertext, xchachaKey, serpentKey, xchaChaNonce, serpentNonce []byte) ([]byte, error) {
	// First layer: decrypt Serpent-CTR
	intermediate, err := SerpentCTRDecryptWithNonce(ciphertext, serpentKey, serpentNonce)
	if err != nil {
		return nil, fmt.Errorf("serpent decryption failed: %w", err)
	}

	// Second layer: decrypt XChaCha20-Poly1305
	plaintext, err := XChaCha20DecryptWithNonce(intermediate, xchachaKey, xchaChaNonce)
	if err != nil {
		return nil, fmt.Errorf("xchacha20 decryption failed: %w", err)
	}

	return plaintext, nil
}

// DeriveSerpentNonce derives a Serpent nonce from chunk index (for streaming)
func DeriveSerpentNonce(chunkIndex uint64) []byte {
	nonce := make([]byte, SerpentNonceSize)
	binary.BigEndian.PutUint64(nonce[8:], chunkIndex)
	return nonce
}
