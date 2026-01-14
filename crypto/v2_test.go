package crypto

import (
	"bytes"
	"testing"
)

func TestHKDFSubkeyDerivation(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = byte(i + 100)
	}

	// Test deriving all subkeys for normal mode
	keys, err := DeriveAllSubkeys(masterKey, salt, false)
	if err != nil {
		t.Fatalf("DeriveAllSubkeys failed: %v", err)
	}

	if len(keys.HeaderHMAC) != 64 {
		t.Errorf("HeaderHMAC length: expected 64, got %d", len(keys.HeaderHMAC))
	}
	if len(keys.PayloadMAC) != 32 {
		t.Errorf("PayloadMAC length: expected 32, got %d", len(keys.PayloadMAC))
	}
	if len(keys.Cipher) != 32 {
		t.Errorf("Cipher length: expected 32, got %d", len(keys.Cipher))
	}
	if keys.Serpent != nil {
		t.Error("Serpent key should be nil in normal mode")
	}

	// Test paranoid mode
	paranoKeys, err := DeriveAllSubkeys(masterKey, salt, true)
	if err != nil {
		t.Fatalf("DeriveAllSubkeys (paranoid) failed: %v", err)
	}

	if len(paranoKeys.Serpent) != 32 {
		t.Errorf("Serpent length: expected 32, got %d", len(paranoKeys.Serpent))
	}

	// Keys should be different
	if bytes.Equal(paranoKeys.Cipher, paranoKeys.Serpent) {
		t.Error("Cipher and Serpent keys should be different")
	}
}

func TestXChaCha20Encryption(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("Hello, XChaCha20-Poly1305!")

	ciphertext, err := XChaCha20Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("XChaCha20Encrypt failed: %v", err)
	}

	// Ciphertext should be longer than plaintext (nonce + tag)
	if len(ciphertext) <= len(plaintext) {
		t.Error("Ciphertext should be longer than plaintext")
	}

	// Decrypt
	decrypted, err := XChaCha20Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("XChaCha20Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, expected %s", decrypted, plaintext)
	}
}

func TestSerpentCTREncryption(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("Hello, Serpent-256-CTR!")

	ciphertext, err := SerpentCTREncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("SerpentCTREncrypt failed: %v", err)
	}

	// Ciphertext should include nonce
	if len(ciphertext) != len(plaintext)+SerpentNonceSize {
		t.Errorf("Ciphertext length: expected %d, got %d", len(plaintext)+SerpentNonceSize, len(ciphertext))
	}

	// Decrypt
	decrypted, err := SerpentCTRDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("SerpentCTRDecrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, expected %s", decrypted, plaintext)
	}
}

func TestCascadeEncryption(t *testing.T) {
	xchachaKey := make([]byte, 32)
	serpentKey := make([]byte, 32)
	for i := range xchachaKey {
		xchachaKey[i] = byte(i)
		serpentKey[i] = byte(i + 32)
	}

	plaintext := []byte("Hello, XChaCha20 + Serpent cascade!")

	ciphertext, err := CascadeEncrypt(plaintext, xchachaKey, serpentKey)
	if err != nil {
		t.Fatalf("CascadeEncrypt failed: %v", err)
	}

	// Decrypt
	decrypted, err := CascadeDecrypt(ciphertext, xchachaKey, serpentKey)
	if err != nil {
		t.Fatalf("CascadeDecrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, expected %s", decrypted, plaintext)
	}
}

func TestBlake2bMAC(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("Test data for BLAKE2b MAC")

	mac, err := ComputeBlake2bMAC(data, key)
	if err != nil {
		t.Fatalf("ComputeBlake2bMAC failed: %v", err)
	}

	if len(mac) != BLAKE2bOutputSize {
		t.Errorf("MAC length: expected %d, got %d", BLAKE2bOutputSize, len(mac))
	}

	// Verify
	valid, err := VerifyBlake2bMAC(data, key, mac)
	if err != nil {
		t.Fatalf("VerifyBlake2bMAC failed: %v", err)
	}
	if !valid {
		t.Error("MAC verification failed")
	}

	// Verify with wrong data should fail
	valid, _ = VerifyBlake2bMAC([]byte("wrong data"), key, mac)
	if valid {
		t.Error("MAC verification should fail for wrong data")
	}
}

func TestHMACSHA3(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("Test data for HMAC-SHA3-512")

	mac, err := ComputeHMACSHA3(data, key)
	if err != nil {
		t.Fatalf("ComputeHMACSHA3 failed: %v", err)
	}

	if len(mac) != HMACSHA3OutputSize {
		t.Errorf("MAC length: expected %d, got %d", HMACSHA3OutputSize, len(mac))
	}

	// Verify
	valid, err := VerifyHMACSHA3(data, key, mac)
	if err != nil {
		t.Fatalf("VerifyHMACSHA3 failed: %v", err)
	}
	if !valid {
		t.Error("MAC verification failed")
	}
}

func TestEncryptV2NormalMode(t *testing.T) {
	password := []byte("test-password-123")
	plaintext := []byte("Secret message for v2 format testing")

	opts := &EncryptOptions{
		Mode:   ModeNormal,
		Cipher: CipherAESGCM,
	}

	ciphertext, err := EncryptV2(plaintext, password, opts)
	if err != nil {
		t.Fatalf("EncryptV2 failed: %v", err)
	}

	// Check v2 format
	if !IsV2Format(ciphertext) {
		t.Error("Output should be in v2 format")
	}

	// Decrypt
	decrypted, err := DecryptV2(ciphertext, password)
	if err != nil {
		t.Fatalf("DecryptV2 failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch")
	}
}

func TestEncryptV2ParanoidMode(t *testing.T) {
	password := []byte("paranoid-password-456")
	plaintext := []byte("Ultra secret message for paranoid mode")

	opts := &EncryptOptions{
		Mode:   ModeParanoid,
		Cipher: CipherCascade,
	}

	ciphertext, err := EncryptV2(plaintext, password, opts)
	if err != nil {
		t.Fatalf("EncryptV2 (paranoid) failed: %v", err)
	}

	// Check v2 format
	if !IsV2Format(ciphertext) {
		t.Error("Output should be in v2 format")
	}

	// Decrypt
	decrypted, err := DecryptV2(ciphertext, password)
	if err != nil {
		t.Fatalf("DecryptV2 (paranoid) failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch")
	}
}

func TestEncryptV2XChaCha20(t *testing.T) {
	password := []byte("xchacha-password")
	plaintext := []byte("XChaCha20 test message")

	opts := &EncryptOptions{
		Mode:   ModeNormal,
		Cipher: CipherXChaCha20,
	}

	ciphertext, err := EncryptV2(plaintext, password, opts)
	if err != nil {
		t.Fatalf("EncryptV2 (xchacha20) failed: %v", err)
	}

	decrypted, err := DecryptV2(ciphertext, password)
	if err != nil {
		t.Fatalf("DecryptV2 (xchacha20) failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text mismatch")
	}
}

func TestDetectAndDecrypt(t *testing.T) {
	password := []byte("detect-test-password")
	plaintext := []byte("Auto-detect format test")

	// Test v2 format
	opts := &EncryptOptions{
		Mode:   ModeNormal,
		Cipher: CipherAESGCM,
	}

	v2Ciphertext, err := EncryptV2(plaintext, password, opts)
	if err != nil {
		t.Fatalf("EncryptV2 failed: %v", err)
	}

	decrypted, err := DetectAndDecrypt(v2Ciphertext, password)
	if err != nil {
		t.Fatalf("DetectAndDecrypt (v2) failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted text mismatch for v2")
	}
}

func TestAdaptiveKDFParams(t *testing.T) {
	// Test normal mode
	params := AdaptiveKDFParams(KDFModeNormal, 0)
	if params.Time != 4 {
		t.Errorf("Normal mode time: expected 4, got %d", params.Time)
	}

	// Test paranoid mode
	params = AdaptiveKDFParams(KDFModeParanoid, 0)
	if params.Time != 8 {
		t.Errorf("Paranoid mode time: expected 8, got %d", params.Time)
	}

	// Test custom memory
	params = AdaptiveKDFParams(KDFModeNormal, 128)
	if params.Memory != 128*1024 {
		t.Errorf("Custom memory: expected %d, got %d", 128*1024, params.Memory)
	}
}

func TestDeriveNonce(t *testing.T) {
	masterKey := make([]byte, 32)
	salt := make([]byte, 16)

	nonce1, err := DeriveNonce(masterKey, salt, 0, XChaChaNonceSize)
	if err != nil {
		t.Fatalf("DeriveNonce failed: %v", err)
	}

	if len(nonce1) != XChaChaNonceSize {
		t.Errorf("Nonce length: expected %d, got %d", XChaChaNonceSize, len(nonce1))
	}

	// Different chunk index should give different nonce
	nonce2, err := DeriveNonce(masterKey, salt, 1, XChaChaNonceSize)
	if err != nil {
		t.Fatalf("DeriveNonce (chunk 1) failed: %v", err)
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Error("Nonces for different chunks should be different")
	}
}

func TestParseEncryptMode(t *testing.T) {
	tests := []struct {
		input    string
		expected EncryptionMode
		hasError bool
	}{
		{"normal", ModeNormal, false},
		{"", ModeNormal, false},
		{"paranoid", ModeParanoid, false},
		{"invalid", ModeNormal, true},
	}

	for _, tc := range tests {
		mode, err := ParseEncryptMode(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("ParseEncryptMode(%q) should return error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseEncryptMode(%q) unexpected error: %v", tc.input, err)
			}
			if mode != tc.expected {
				t.Errorf("ParseEncryptMode(%q) = %v, expected %v", tc.input, mode, tc.expected)
			}
		}
	}
}

func TestParseCipher(t *testing.T) {
	tests := []struct {
		input    string
		expected CipherType
		hasError bool
	}{
		{"aes-gcm", CipherAESGCM, false},
		{"aes", CipherAESGCM, false},
		{"", CipherAESGCM, false},
		{"xchacha20", CipherXChaCha20, false},
		{"chacha", CipherXChaCha20, false},
		{"cascade", CipherCascade, false},
		{"invalid", CipherAESGCM, true},
	}

	for _, tc := range tests {
		cipher, err := ParseCipher(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("ParseCipher(%q) should return error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseCipher(%q) unexpected error: %v", tc.input, err)
			}
			if cipher != tc.expected {
				t.Errorf("ParseCipher(%q) = %v, expected %v", tc.input, cipher, tc.expected)
			}
		}
	}
}
