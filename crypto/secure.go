package crypto

import (
	"runtime"
	"sync"
)

// SecureBuffer is a buffer that can be securely wiped after use
type SecureBuffer struct {
	data []byte
	mu   sync.Mutex
}

// NewSecureBuffer creates a new secure buffer with the given size
func NewSecureBuffer(size int) *SecureBuffer {
	return &SecureBuffer{
		data: make([]byte, size),
	}
}

// NewSecureBufferFromBytes creates a secure buffer from existing bytes (copies data)
func NewSecureBufferFromBytes(src []byte) *SecureBuffer {
	sb := &SecureBuffer{
		data: make([]byte, len(src)),
	}
	copy(sb.data, src)
	return sb
}

// Bytes returns the underlying byte slice
func (s *SecureBuffer) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

// Len returns the length of the buffer
func (s *SecureBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

// Write writes data to the buffer at the beginning
func (s *SecureBuffer) Write(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy(s.data, data)
}

// Wipe securely zeros out the buffer contents
func (s *SecureBuffer) Wipe() {
	s.mu.Lock()
	defer s.mu.Unlock()
	SecureWipe(s.data)
}

// SecureWipe zeros out a byte slice to prevent sensitive data from remaining in memory
// Uses runtime.KeepAlive to prevent compiler optimizations from removing the wipe
func SecureWipe(b []byte) {
	if b == nil {
		return
	}
	for i := range b {
		b[i] = 0
	}
	// Prevent compiler from optimizing away the zeroing
	runtime.KeepAlive(b)
}

// SecureWipeMultiple wipes multiple byte slices
func SecureWipeMultiple(slices ...[]byte) {
	for _, s := range slices {
		SecureWipe(s)
	}
}

// SecureCompare performs constant-time comparison of two byte slices
// Returns true if they are equal
func SecureCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// SecureString is a string wrapper that can be wiped
type SecureString struct {
	data []byte
}

// NewSecureString creates a secure string from a regular string
func NewSecureString(s string) *SecureString {
	return &SecureString{
		data: []byte(s),
	}
}

// String returns the string value
func (s *SecureString) String() string {
	return string(s.data)
}

// Bytes returns the underlying bytes
func (s *SecureString) Bytes() []byte {
	return s.data
}

// Wipe securely zeros out the string data
func (s *SecureString) Wipe() {
	SecureWipe(s.data)
}

// WipeString attempts to wipe a string by converting to bytes
// Note: This is less effective than SecureString due to string immutability
// but provides some protection
func WipeString(s *string) {
	if s == nil || *s == "" {
		return
	}
	b := []byte(*s)
	SecureWipe(b)
	*s = ""
}

// WithSecureBuffer executes a function with a secure buffer and automatically wipes it
func WithSecureBuffer(size int, fn func(buf *SecureBuffer) error) error {
	buf := NewSecureBuffer(size)
	defer buf.Wipe()
	return fn(buf)
}

// WithSecureBytes executes a function with secure bytes and automatically wipes them
func WithSecureBytes(data []byte, fn func([]byte) error) error {
	// Make a copy so we don't wipe the original
	secure := make([]byte, len(data))
	copy(secure, data)
	defer SecureWipe(secure)
	return fn(secure)
}

// DerivedKeysSecure wraps DerivedKeys with secure wiping capability
type DerivedKeysSecure struct {
	*DerivedKeys
}

// Wipe securely wipes all derived keys
func (d *DerivedKeysSecure) Wipe() {
	if d.DerivedKeys == nil {
		return
	}
	SecureWipe(d.HeaderHMAC)
	SecureWipe(d.PayloadMAC)
	SecureWipe(d.Cipher)
	SecureWipe(d.Serpent)
}

// WrapDerivedKeys wraps DerivedKeys for secure wiping
func WrapDerivedKeys(keys *DerivedKeys) *DerivedKeysSecure {
	return &DerivedKeysSecure{DerivedKeys: keys}
}
