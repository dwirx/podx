package crypto

import (
	"bytes"
	"testing"
)

func TestSecureWipe(t *testing.T) {
	// Create buffer with data
	data := []byte("sensitive password data here")
	original := make([]byte, len(data))
	copy(original, data)

	// Verify data is there
	if !bytes.Equal(data, original) {
		t.Fatal("data should equal original before wipe")
	}

	// Wipe
	SecureWipe(data)

	// Verify data is zeroed
	for i, b := range data {
		if b != 0 {
			t.Errorf("byte %d should be 0, got %d", i, b)
		}
	}

	// Test nil slice
	SecureWipe(nil) // Should not panic
}

func TestSecureWipeMultiple(t *testing.T) {
	slice1 := []byte("secret1")
	slice2 := []byte("secret2")
	slice3 := []byte("secret3")

	SecureWipeMultiple(slice1, slice2, slice3)

	for i, s := range [][]byte{slice1, slice2, slice3} {
		for j, b := range s {
			if b != 0 {
				t.Errorf("slice %d, byte %d should be 0, got %d", i, j, b)
			}
		}
	}
}

func TestSecureBuffer(t *testing.T) {
	buf := NewSecureBuffer(32)

	// Write data
	data := []byte("test data for secure buffer!!!!!")
	buf.Write(data)

	// Verify data is there
	if !bytes.Equal(buf.Bytes()[:len(data)], data) {
		t.Error("buffer should contain written data")
	}

	// Wipe
	buf.Wipe()

	// Verify wiped
	for i, b := range buf.Bytes() {
		if b != 0 {
			t.Errorf("byte %d should be 0 after wipe, got %d", i, b)
		}
	}
}

func TestSecureBufferFromBytes(t *testing.T) {
	original := []byte("original secret data")
	buf := NewSecureBufferFromBytes(original)

	// Verify copy was made
	if !bytes.Equal(buf.Bytes(), original) {
		t.Error("buffer should contain original data")
	}

	// Wipe buffer
	buf.Wipe()

	// Original should be unchanged
	if bytes.Equal(original, make([]byte, len(original))) {
		t.Error("original should not be wiped")
	}
}

func TestSecureCompare(t *testing.T) {
	a := []byte("hello world")
	b := []byte("hello world")
	c := []byte("hello worle")
	d := []byte("short")

	if !SecureCompare(a, b) {
		t.Error("identical slices should be equal")
	}

	if SecureCompare(a, c) {
		t.Error("different slices should not be equal")
	}

	if SecureCompare(a, d) {
		t.Error("different length slices should not be equal")
	}
}

func TestSecureString(t *testing.T) {
	ss := NewSecureString("my secret password")

	if ss.String() != "my secret password" {
		t.Error("string should match original")
	}

	ss.Wipe()

	// Verify wiped
	for i, b := range ss.Bytes() {
		if b != 0 {
			t.Errorf("byte %d should be 0 after wipe, got %d", i, b)
		}
	}
}

func TestWithSecureBuffer(t *testing.T) {
	var capturedBuf *SecureBuffer

	err := WithSecureBuffer(16, func(buf *SecureBuffer) error {
		buf.Write([]byte("secret data here"))
		capturedBuf = buf
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Buffer should be wiped after function returns
	for i, b := range capturedBuf.Bytes() {
		if b != 0 {
			t.Errorf("byte %d should be 0 after WithSecureBuffer, got %d", i, b)
		}
	}
}

func TestWithSecureBytes(t *testing.T) {
	original := []byte("original data")
	var capturedCopy []byte

	err := WithSecureBytes(original, func(data []byte) error {
		capturedCopy = data
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Copy should be wiped
	for i, b := range capturedCopy {
		if b != 0 {
			t.Errorf("byte %d should be 0 after WithSecureBytes, got %d", i, b)
		}
	}

	// Original should be unchanged
	if string(original) != "original data" {
		t.Error("original should not be modified")
	}
}

func TestDerivedKeysSecure(t *testing.T) {
	keys := &DerivedKeys{
		HeaderHMAC: []byte("header-hmac-key-here-64-bytes!!!" +
			"header-hmac-key-here-64-bytes!!!"),
		PayloadMAC: []byte("payload-mac-key-here-32-bytes!!"),
		Cipher:     []byte("cipher-key-here-32-bytes!!!!!!!!"),
		Serpent:    []byte("serpent-key-here-32-bytes!!!!!!!"),
	}

	secure := WrapDerivedKeys(keys)
	secure.Wipe()

	// All keys should be zeroed
	allZero := func(b []byte) bool {
		for _, v := range b {
			if v != 0 {
				return false
			}
		}
		return true
	}

	if !allZero(keys.HeaderHMAC) {
		t.Error("HeaderHMAC should be wiped")
	}
	if !allZero(keys.PayloadMAC) {
		t.Error("PayloadMAC should be wiped")
	}
	if !allZero(keys.Cipher) {
		t.Error("Cipher should be wiped")
	}
	if !allZero(keys.Serpent) {
		t.Error("Serpent should be wiped")
	}
}
