package crypto

import (
	"encoding/base64"
	"fmt"
)

// Algorithm type untuk memilih algoritma enkripsi
type Algorithm string

const (
	AlgoAESGCM   Algorithm = "aes-gcm"
	AlgoChaCha20 Algorithm = "chacha20"
)

// Encryptor adalah interface untuk enkripsi/dekripsi
type Encryptor interface {
	Encrypt(plaintext, key []byte) ([]byte, error)
	Decrypt(ciphertext, key []byte) ([]byte, error)
	Name() string
}

// aesGCMEncryptor implements Encryptor untuk AES-256-GCM
type aesGCMEncryptor struct{}

func (a *aesGCMEncryptor) Encrypt(plaintext, key []byte) ([]byte, error) {
	return AESGCMEncrypt(plaintext, key)
}

func (a *aesGCMEncryptor) Decrypt(ciphertext, key []byte) ([]byte, error) {
	return AESGCMDecrypt(ciphertext, key)
}

func (a *aesGCMEncryptor) Name() string {
	return string(AlgoAESGCM)
}

// chaCha20Encryptor implements Encryptor untuk ChaCha20-Poly1305
type chaCha20Encryptor struct{}

func (c *chaCha20Encryptor) Encrypt(plaintext, key []byte) ([]byte, error) {
	return ChaCha20Encrypt(plaintext, key)
}

func (c *chaCha20Encryptor) Decrypt(ciphertext, key []byte) ([]byte, error) {
	return ChaCha20Decrypt(ciphertext, key)
}

func (c *chaCha20Encryptor) Name() string {
	return string(AlgoChaCha20)
}

// NewEncryptor membuat Encryptor berdasarkan algoritma yang dipilih
func NewEncryptor(algo Algorithm) (Encryptor, error) {
	switch algo {
	case AlgoAESGCM:
		return &aesGCMEncryptor{}, nil
	case AlgoChaCha20:
		return &chaCha20Encryptor{}, nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %s (supported: aes-gcm, chacha20)", algo)
	}
}

// EncryptToBase64 mengenkripsi dan mengembalikan hasil dalam format base64
func EncryptToBase64(enc Encryptor, plaintext, key []byte) (string, error) {
	ciphertext, err := enc.Encrypt(plaintext, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptFromBase64 mendekripsi dari format base64
func DecryptFromBase64(enc Encryptor, ciphertextB64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return enc.Decrypt(ciphertext, key)
}
