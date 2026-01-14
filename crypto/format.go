package crypto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// File format constants
const (
	// Magic bytes for v2 format
	MagicV2 = "PDX2"

	// Version numbers
	FormatVersionV1 = 0x01
	FormatVersionV2 = 0x02

	// Header sizes
	HeaderSizeV2 = 4 + 1 + 1 + 16 + 64 // magic(4) + version(1) + flags(1) + salt(16) + headerHMAC(64)
)

// EncryptionMode represents the encryption mode
type EncryptionMode int

const (
	ModeNormal   EncryptionMode = iota // Normal mode
	ModeParanoid                       // Paranoid mode with cascade encryption
)

// CipherType represents the cipher algorithm
type CipherType int

const (
	CipherAESGCM    CipherType = iota // AES-256-GCM
	CipherXChaCha20                   // XChaCha20-Poly1305
	CipherCascade                     // XChaCha20 + Serpent cascade
)

// FileFlags encodes encryption options into a single byte
type FileFlags struct {
	Mode      EncryptionMode // bit 0: 0=normal, 1=paranoid
	Cipher    CipherType     // bits 1-2: cipher type
	Streaming bool           // bit 3: 0=binary, 1=streaming
}

// ToByte encodes flags to a single byte
func (f *FileFlags) ToByte() byte {
	var b byte

	// bit 0: mode
	if f.Mode == ModeParanoid {
		b |= 0x01
	}

	// bits 1-2: cipher
	b |= byte(f.Cipher&0x03) << 1

	// bit 3: streaming
	if f.Streaming {
		b |= 0x08
	}

	return b
}

// ParseFlags decodes flags from a byte
func ParseFlags(b byte) *FileFlags {
	return &FileFlags{
		Mode:      EncryptionMode(b & 0x01),
		Cipher:    CipherType((b >> 1) & 0x03),
		Streaming: (b & 0x08) != 0,
	}
}

// HeaderV2 represents the v2 file header
type HeaderV2 struct {
	Magic      [4]byte  // "PDX2"
	Version    byte     // 0x02
	Flags      byte     // encoded FileFlags
	Salt       [16]byte // Argon2id salt
	HeaderHMAC [64]byte // HMAC of header fields (magic + version + flags + salt)
}

// EncryptedFileV2 represents a complete v2 encrypted file
type EncryptedFileV2 struct {
	Header     *HeaderV2
	Payload    []byte // Encrypted data
	PayloadMAC []byte // 32 bytes (BLAKE2b) or 64 bytes (HMAC-SHA3)
}

// EncryptOptions holds options for encryption
type EncryptOptions struct {
	Mode      EncryptionMode
	Cipher    CipherType
	MemoryMB  uint32 // Argon2id memory in MB (0 for default)
	Streaming bool   // Use streaming for large files
}

// DefaultEncryptOptions returns default encryption options for the given mode
func DefaultEncryptOptions(mode EncryptionMode) *EncryptOptions {
	opts := &EncryptOptions{
		Mode:      mode,
		Streaming: false,
	}

	if mode == ModeParanoid {
		opts.Cipher = CipherCascade
	} else {
		opts.Cipher = CipherAESGCM
	}

	return opts
}

// EncryptV2 encrypts data using v2 format
func EncryptV2(plaintext, password []byte, opts *EncryptOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultEncryptOptions(ModeNormal)
	}

	// Determine KDF mode
	kdfMode := KDFModeNormal
	if opts.Mode == ModeParanoid {
		kdfMode = KDFModeParanoid
	}

	// Derive master key
	masterKey, salt, _, err := DeriveKeyV2(password, nil, kdfMode, opts.MemoryMB)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// Derive subkeys
	paranoid := opts.Mode == ModeParanoid
	keys, err := DeriveAllSubkeys(masterKey, salt, paranoid)
	if err != nil {
		return nil, fmt.Errorf("subkey derivation failed: %w", err)
	}

	// Encrypt payload
	var ciphertext []byte
	switch opts.Cipher {
	case CipherAESGCM:
		ciphertext, err = AESGCMEncrypt(plaintext, keys.Cipher)
	case CipherXChaCha20:
		ciphertext, err = XChaCha20Encrypt(plaintext, keys.Cipher)
	case CipherCascade:
		ciphertext, err = CascadeEncrypt(plaintext, keys.Cipher, keys.Serpent)
	default:
		return nil, fmt.Errorf("unknown cipher type: %d", opts.Cipher)
	}
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Build header
	flags := &FileFlags{
		Mode:      opts.Mode,
		Cipher:    opts.Cipher,
		Streaming: opts.Streaming,
	}

	header := &HeaderV2{
		Version: FormatVersionV2,
		Flags:   flags.ToByte(),
	}
	copy(header.Magic[:], MagicV2)
	copy(header.Salt[:], salt)

	// Compute header HMAC (over magic + version + flags + salt)
	headerData := make([]byte, 4+1+1+16) // 22 bytes
	copy(headerData[0:4], header.Magic[:])
	headerData[4] = header.Version
	headerData[5] = header.Flags
	copy(headerData[6:22], header.Salt[:])

	headerMAC, err := ComputeHMACSHA3(headerData, keys.HeaderHMAC[:32])
	if err != nil {
		return nil, fmt.Errorf("header MAC failed: %w", err)
	}
	copy(header.HeaderHMAC[:], headerMAC)

	// Compute payload MAC
	var payloadMAC []byte
	if paranoid {
		payloadMAC, err = ComputeHMACSHA3(ciphertext, keys.PayloadMAC)
	} else {
		payloadMAC, err = ComputeBlake2bMAC(ciphertext, keys.PayloadMAC)
	}
	if err != nil {
		return nil, fmt.Errorf("payload MAC failed: %w", err)
	}

	// Build output: header + ciphertext + payloadMAC
	output := make([]byte, HeaderSizeV2+len(ciphertext)+len(payloadMAC))
	copy(output[0:4], header.Magic[:])
	output[4] = header.Version
	output[5] = header.Flags
	copy(output[6:22], header.Salt[:])
	copy(output[22:86], header.HeaderHMAC[:])
	copy(output[86:86+len(ciphertext)], ciphertext)
	copy(output[86+len(ciphertext):], payloadMAC)

	return output, nil
}

// DecryptV2 decrypts data in v2 format
func DecryptV2(data, password []byte) ([]byte, error) {
	if len(data) < HeaderSizeV2 {
		return nil, errors.New("data too short for v2 format")
	}

	// Parse header
	if string(data[0:4]) != MagicV2 {
		return nil, errors.New("invalid magic bytes")
	}

	version := data[4]
	if version != FormatVersionV2 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	flags := ParseFlags(data[5])
	salt := data[6:22]
	storedHeaderMAC := data[22:86]

	// Determine KDF mode
	kdfMode := KDFModeNormal
	if flags.Mode == ModeParanoid {
		kdfMode = KDFModeParanoid
	}

	// Derive master key
	masterKey, _, _, err := DeriveKeyV2(password, salt, kdfMode, 0)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// Derive subkeys
	paranoid := flags.Mode == ModeParanoid
	keys, err := DeriveAllSubkeys(masterKey, salt, paranoid)
	if err != nil {
		return nil, fmt.Errorf("subkey derivation failed: %w", err)
	}

	// Verify header HMAC
	headerData := data[0:22]
	headerMAC, err := ComputeHMACSHA3(headerData, keys.HeaderHMAC[:32])
	if err != nil {
		return nil, fmt.Errorf("header MAC computation failed: %w", err)
	}
	if !bytes.Equal(headerMAC, storedHeaderMAC) {
		return nil, errors.New("header MAC verification failed")
	}

	// Determine payload MAC size
	payloadMACSize := BLAKE2bOutputSize
	if paranoid {
		payloadMACSize = HMACSHA3OutputSize
	}

	if len(data) < HeaderSizeV2+payloadMACSize {
		return nil, errors.New("data too short: missing payload MAC")
	}

	// Extract ciphertext and MAC
	ciphertext := data[HeaderSizeV2 : len(data)-payloadMACSize]
	storedPayloadMAC := data[len(data)-payloadMACSize:]

	// Verify payload MAC
	var payloadMAC []byte
	if paranoid {
		payloadMAC, err = ComputeHMACSHA3(ciphertext, keys.PayloadMAC)
	} else {
		payloadMAC, err = ComputeBlake2bMAC(ciphertext, keys.PayloadMAC)
	}
	if err != nil {
		return nil, fmt.Errorf("payload MAC computation failed: %w", err)
	}
	if !bytes.Equal(payloadMAC, storedPayloadMAC) {
		return nil, errors.New("payload MAC verification failed (wrong password or corrupted data)")
	}

	// Decrypt payload
	var plaintext []byte
	switch flags.Cipher {
	case CipherAESGCM:
		plaintext, err = AESGCMDecrypt(ciphertext, keys.Cipher)
	case CipherXChaCha20:
		plaintext, err = XChaCha20Decrypt(ciphertext, keys.Cipher)
	case CipherCascade:
		plaintext, err = CascadeDecrypt(ciphertext, keys.Cipher, keys.Serpent)
	default:
		return nil, fmt.Errorf("unknown cipher type: %d", flags.Cipher)
	}
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// IsV2Format checks if data is in v2 format
func IsV2Format(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return string(data[0:4]) == MagicV2
}

// DetectAndDecrypt automatically detects format and decrypts
func DetectAndDecrypt(data, password []byte) ([]byte, error) {
	if IsV2Format(data) {
		return DecryptV2(data, password)
	}

	// Try v1 format
	return DecryptV1(data, password)
}

// DecryptV1 decrypts data in v1 format (for backward compatibility)
func DecryptV1(data, password []byte) ([]byte, error) {
	if len(data) < SaltSize+1 {
		return nil, errors.New("data too short for v1 format")
	}

	// Extract salt and algo
	salt := data[:SaltSize]
	algoB := data[SaltSize]
	ciphertext := data[SaltSize+1:]

	algo := AlgoAESGCM
	if algoB == 1 || algoB == 0x02 {
		algo = AlgoChaCha20
	}

	// Derive key using legacy parameters
	key, err := DeriveKeyWithSalt(password, salt)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// Get encryptor
	enc, err := NewEncryptor(algo)
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := enc.Decrypt(ciphertext, key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	return plaintext, nil
}

// EncryptModeString returns the string representation of encryption mode
func EncryptModeString(mode EncryptionMode) string {
	switch mode {
	case ModeNormal:
		return "normal"
	case ModeParanoid:
		return "paranoid"
	default:
		return "unknown"
	}
}

// ParseEncryptMode parses mode string to EncryptionMode
func ParseEncryptMode(s string) (EncryptionMode, error) {
	switch s {
	case "normal", "":
		return ModeNormal, nil
	case "paranoid":
		return ModeParanoid, nil
	default:
		return ModeNormal, fmt.Errorf("unknown mode: %s (use: normal, paranoid)", s)
	}
}

// CipherString returns the string representation of cipher type
func CipherString(c CipherType) string {
	switch c {
	case CipherAESGCM:
		return "aes-gcm"
	case CipherXChaCha20:
		return "xchacha20"
	case CipherCascade:
		return "cascade"
	default:
		return "unknown"
	}
}

// ParseCipher parses cipher string to CipherType
func ParseCipher(s string) (CipherType, error) {
	switch s {
	case "aes-gcm", "aes", "":
		return CipherAESGCM, nil
	case "xchacha20", "chacha", "xchacha":
		return CipherXChaCha20, nil
	case "cascade":
		return CipherCascade, nil
	default:
		return CipherAESGCM, fmt.Errorf("unknown cipher: %s (use: aes-gcm, xchacha20, cascade)", s)
	}
}

// EncryptV2Info returns info about encrypted data
func EncryptV2Info(data []byte) (string, error) {
	if !IsV2Format(data) {
		return "v1 format (legacy)", nil
	}

	if len(data) < 6 {
		return "", errors.New("data too short")
	}

	flags := ParseFlags(data[5])

	mode := EncryptModeString(flags.Mode)
	cipher := CipherString(flags.Cipher)
	streaming := "no"
	if flags.Streaming {
		streaming = "yes"
	}

	return fmt.Sprintf("v2 format: mode=%s, cipher=%s, streaming=%s", mode, cipher, streaming), nil
}

// GetMACSize returns the MAC size for the given encryption mode
func GetMACSize(mode EncryptionMode) int {
	if mode == ModeParanoid {
		return HMACSHA3OutputSize // 64 bytes
	}
	return BLAKE2bOutputSize // 32 bytes
}

// ChunkSize for streaming encryption (1MB)
const StreamingChunkSize = 1024 * 1024

// RekeyInterval for paranoid mode (1GB = 1024 chunks)
const RekeyInterval = 1024

// EncodeUint64 encodes a uint64 to bytes
func EncodeUint64(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
