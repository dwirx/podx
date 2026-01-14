package crypto

import (
	"crypto/subtle"
	"fmt"
	"hash"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

const (
	// BLAKE2b constants
	BLAKE2bKeySize    = 32 // 256-bit key
	BLAKE2bOutputSize = 32 // 256-bit output

	// HMAC-SHA3 constants
	HMACSHA3KeySize    = 32 // 256-bit key
	HMACSHA3OutputSize = 64 // 512-bit output (SHA3-512)
)

// MACType represents the type of MAC algorithm
type MACType int

const (
	MACTypeBlake2b MACType = iota // Normal mode
	MACTypeSHA3                   // Paranoid mode
)

// MAC is the interface for Message Authentication Code implementations
type MAC interface {
	// Write adds data to the MAC computation
	Write(data []byte) (int, error)
	// Sum computes the MAC and appends it to b
	Sum(b []byte) []byte
	// Reset resets the MAC to initial state
	Reset()
	// Size returns the size of the MAC output
	Size() int
	// Verify checks if the provided tag matches the computed MAC
	Verify(tag []byte) bool
}

// blake2bMAC implements MAC using Keyed-BLAKE2b
type blake2bMAC struct {
	hash hash.Hash
	key  []byte
}

// NewBlake2bMAC creates a new Keyed-BLAKE2b MAC
func NewBlake2bMAC(key []byte) (MAC, error) {
	if len(key) != BLAKE2bKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", BLAKE2bKeySize, len(key))
	}

	h, err := blake2b.New256(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create BLAKE2b: %w", err)
	}

	return &blake2bMAC{
		hash: h,
		key:  key,
	}, nil
}

func (m *blake2bMAC) Write(data []byte) (int, error) {
	return m.hash.Write(data)
}

func (m *blake2bMAC) Sum(b []byte) []byte {
	return m.hash.Sum(b)
}

func (m *blake2bMAC) Reset() {
	// Re-create the hash with the key
	h, _ := blake2b.New256(m.key)
	m.hash = h
}

func (m *blake2bMAC) Size() int {
	return BLAKE2bOutputSize
}

func (m *blake2bMAC) Verify(tag []byte) bool {
	computed := m.hash.Sum(nil)
	return subtle.ConstantTimeCompare(computed, tag) == 1
}

// hmacSHA3 implements MAC using HMAC-SHA3-512
type hmacSHA3 struct {
	key    []byte
	buffer []byte
}

// NewHMACSHA3 creates a new HMAC-SHA3-512 MAC
func NewHMACSHA3(key []byte) (MAC, error) {
	if len(key) != HMACSHA3KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", HMACSHA3KeySize, len(key))
	}

	return &hmacSHA3{
		key:    key,
		buffer: nil,
	}, nil
}

func (m *hmacSHA3) Write(data []byte) (int, error) {
	m.buffer = append(m.buffer, data...)
	return len(data), nil
}

func (m *hmacSHA3) Sum(b []byte) []byte {
	// HMAC computation using SHA3-512
	// SHA3 is resistant to length-extension attacks, but we still use HMAC structure
	// for consistency with standard practices

	blockSize := 136 // SHA3-512 block size

	// Pad key if needed
	key := m.key
	if len(key) > blockSize {
		h := sha3.New512()
		h.Write(key)
		key = h.Sum(nil)
	}
	if len(key) < blockSize {
		key = append(key, make([]byte, blockSize-len(key))...)
	}

	// ipad and opad
	ipad := make([]byte, blockSize)
	opad := make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		ipad[i] = key[i] ^ 0x36
		opad[i] = key[i] ^ 0x5c
	}

	// Inner hash: SHA3-512(ipad || message)
	innerHash := sha3.New512()
	innerHash.Write(ipad)
	innerHash.Write(m.buffer)
	innerResult := innerHash.Sum(nil)

	// Outer hash: SHA3-512(opad || inner_hash)
	outerHash := sha3.New512()
	outerHash.Write(opad)
	outerHash.Write(innerResult)

	return outerHash.Sum(b)
}

func (m *hmacSHA3) Reset() {
	m.buffer = nil
}

func (m *hmacSHA3) Size() int {
	return HMACSHA3OutputSize
}

func (m *hmacSHA3) Verify(tag []byte) bool {
	computed := m.Sum(nil)
	return subtle.ConstantTimeCompare(computed, tag) == 1
}

// NewMAC creates a new MAC based on the type
func NewMAC(macType MACType, key []byte) (MAC, error) {
	switch macType {
	case MACTypeBlake2b:
		return NewBlake2bMAC(key)
	case MACTypeSHA3:
		return NewHMACSHA3(key)
	default:
		return nil, fmt.Errorf("unknown MAC type: %d", macType)
	}
}

// ComputeBlake2bMAC computes a BLAKE2b MAC for the given data
func ComputeBlake2bMAC(data, key []byte) ([]byte, error) {
	mac, err := NewBlake2bMAC(key)
	if err != nil {
		return nil, err
	}
	mac.Write(data)
	return mac.Sum(nil), nil
}

// ComputeHMACSHA3 computes an HMAC-SHA3-512 MAC for the given data
func ComputeHMACSHA3(data, key []byte) ([]byte, error) {
	mac, err := NewHMACSHA3(key)
	if err != nil {
		return nil, err
	}
	mac.Write(data)
	return mac.Sum(nil), nil
}

// VerifyBlake2bMAC verifies a BLAKE2b MAC
func VerifyBlake2bMAC(data, key, tag []byte) (bool, error) {
	computed, err := ComputeBlake2bMAC(data, key)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(computed, tag) == 1, nil
}

// VerifyHMACSHA3 verifies an HMAC-SHA3-512 MAC
func VerifyHMACSHA3(data, key, tag []byte) (bool, error) {
	computed, err := ComputeHMACSHA3(data, key)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(computed, tag) == 1, nil
}
