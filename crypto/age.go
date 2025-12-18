package crypto

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
)

// AgeEncrypt mengenkripsi plaintext dengan Age encryption menggunakan public key recipient
func AgeEncrypt(plaintext []byte, recipients ...string) ([]byte, error) {
	var ageRecipients []age.Recipient
	for _, r := range recipients {
		recipient, err := age.ParseX25519Recipient(r)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient '%s': %w", r, err)
		}
		ageRecipients = append(ageRecipients, recipient)
	}

	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, ageRecipients...)
	if err != nil {
		return nil, fmt.Errorf("failed to create age encryptor: %w", err)
	}

	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("failed to write plaintext: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encryptor: %w", err)
	}

	return buf.Bytes(), nil
}

// AgeDecrypt mendekripsi ciphertext dengan Age private key identity
func AgeDecrypt(ciphertext []byte, identityData string) ([]byte, error) {
	identity, err := age.ParseX25519Identity(strings.TrimSpace(identityData))
	if err != nil {
		return nil, fmt.Errorf("invalid identity: %w", err)
	}

	r, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read plaintext: %w", err)
	}

	return plaintext, nil
}

// GenerateAgeKey generates a new Age X25519 key pair
// Returns: privateKey, publicKey, error
func GenerateAgeKey() (string, string, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate age key: %w", err)
	}

	return identity.String(), identity.Recipient().String(), nil
}
