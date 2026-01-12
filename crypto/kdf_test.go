package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	password := []byte("test-password")

	key1, salt1, err := DeriveKey(password, nil)
	if err != nil {
		t.Fatalf("DeriveKey error: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("key length = %d, want 32", len(key1))
	}

	if len(salt1) != SaltSize {
		t.Errorf("salt length = %d, want %d", len(salt1), SaltSize)
	}

	// Same password, same salt should produce same key
	key2, err := DeriveKeyWithSalt(password, salt1)
	if err != nil {
		t.Fatalf("DeriveKeyWithSalt error: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("same password+salt should produce same key")
	}
}

func TestDeriveKeyDifferentSalts(t *testing.T) {
	password := []byte("test-password")

	key1, salt1, err := DeriveKey(password, nil)
	if err != nil {
		t.Fatal(err)
	}

	key2, salt2, err := DeriveKey(password, nil)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("random salts should differ")
	}

	if bytes.Equal(key1, key2) {
		t.Error("different salts should produce different keys")
	}
}

func TestDeriveKeyWithProvidedSalt(t *testing.T) {
	password := []byte("test-password")
	salt := make([]byte, SaltSize)
	for i := range salt {
		salt[i] = byte(i)
	}

	key1, returnedSalt, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(returnedSalt, salt) {
		t.Error("returned salt should match provided salt")
	}

	key2, err := DeriveKeyWithSalt(password, salt)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("keys should match for same password+salt")
	}
}

func TestDeriveKeyInvalidSaltSize(t *testing.T) {
	password := []byte("test-password")
	badSalt := make([]byte, 8) // wrong size

	_, _, err := DeriveKey(password, badSalt)
	if err == nil {
		t.Error("expected error for invalid salt size")
	}

	_, err = DeriveKeyWithSalt(password, badSalt)
	if err == nil {
		t.Error("expected error for invalid salt size")
	}
}
