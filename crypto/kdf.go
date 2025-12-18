package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters (OWASP recommended)
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32 // 256-bit key

	// SaltSize ukuran salt (16 bytes recommended)
	SaltSize = 16
)

// DeriveKey menghasilkan 256-bit key dari password menggunakan Argon2id.
// Returns: key (32 bytes), salt (16 bytes), error
func DeriveKey(password []byte, salt []byte) ([]byte, []byte, error) {
	// Generate salt jika tidak diberikan
	if salt == nil {
		salt = make([]byte, SaltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	if len(salt) != SaltSize {
		return nil, nil, fmt.Errorf("invalid salt size: expected %d bytes, got %d", SaltSize, len(salt))
	}

	// Derive key menggunakan Argon2id
	key := argon2.IDKey(password, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	return key, salt, nil
}

// DeriveKeyWithSalt menghasilkan key dari password dengan salt yang sudah ada.
func DeriveKeyWithSalt(password, salt []byte) ([]byte, error) {
	if len(salt) != SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d bytes, got %d", SaltSize, len(salt))
	}

	key := argon2.IDKey(password, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return key, nil
}
