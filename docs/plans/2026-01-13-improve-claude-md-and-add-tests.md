# Improve CLAUDE.md and Add Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enhance CLAUDE.md with testing documentation and add comprehensive unit tests for the crypto, parser, and KDF packages.

**Architecture:** Add Go unit tests following table-driven test patterns. Tests will cover: AES-GCM encryption/decryption, ChaCha20 encryption/decryption, Argon2id key derivation, and .env file parsing. Update CLAUDE.md to document test commands.

**Tech Stack:** Go testing package, testify not required (use standard library assertions)

---

### Task 1: Add AES-GCM Encryption Tests

**Files:**
- Create: `crypto/aes_gcm_test.go`

**Step 1: Write the failing test file**

Create `crypto/aes_gcm_test.go`:

```go
package crypto

import (
	"bytes"
	"testing"
)

func TestAESGCMEncryptDecrypt(t *testing.T) {
	key := make([]byte, AESKeySize)
	for i := range key {
		key[i] = byte(i)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := AESGCMEncrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) <= AESNonceSize {
				t.Fatal("ciphertext too short")
			}

			decrypted, err := AESGCMDecrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAESGCMInvalidKeySize(t *testing.T) {
	shortKey := make([]byte, 16)
	plaintext := []byte("test")

	_, err := AESGCMEncrypt(plaintext, shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestAESGCMDecryptTamperedData(t *testing.T) {
	key := make([]byte, AESKeySize)
	plaintext := []byte("secret data")

	ciphertext, err := AESGCMEncrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	_, err = AESGCMDecrypt(ciphertext, key)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestAESGCMDecryptTooShort(t *testing.T) {
	key := make([]byte, AESKeySize)
	shortData := make([]byte, AESNonceSize-1)

	_, err := AESGCMDecrypt(shortData, key)
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./crypto -run TestAESGCM -v`
Expected: PASS (all tests should pass since implementation exists)

**Step 3: Commit**

```bash
git add crypto/aes_gcm_test.go
git commit -m "test: add AES-GCM encryption unit tests"
```

---

### Task 2: Add ChaCha20-Poly1305 Tests

**Files:**
- Create: `crypto/chacha_test.go`

**Step 1: Write the test file**

Create `crypto/chacha_test.go`:

```go
package crypto

import (
	"bytes"
	"testing"
)

func TestChaCha20EncryptDecrypt(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	for i := range key {
		key[i] = byte(i)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"medium", []byte("the quick brown fox jumps over the lazy dog")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := ChaCha20Encrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(ciphertext) <= ChaChaNonceSize {
				t.Fatal("ciphertext too short")
			}

			decrypted, err := ChaCha20Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestChaCha20InvalidKeySize(t *testing.T) {
	shortKey := make([]byte, 16)
	plaintext := []byte("test")

	_, err := ChaCha20Encrypt(plaintext, shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestChaCha20DecryptTamperedData(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	plaintext := []byte("secret data")

	ciphertext, err := ChaCha20Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xff

	_, err = ChaCha20Decrypt(ciphertext, key)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestChaCha20DecryptTooShort(t *testing.T) {
	key := make([]byte, ChaChaKeySize)
	shortData := make([]byte, ChaChaNonceSize-1)

	_, err := ChaCha20Decrypt(shortData, key)
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./crypto -run TestChaCha20 -v`
Expected: PASS

**Step 3: Commit**

```bash
git add crypto/chacha_test.go
git commit -m "test: add ChaCha20-Poly1305 encryption unit tests"
```

---

### Task 3: Add KDF (Argon2id) Tests

**Files:**
- Create: `crypto/kdf_test.go`

**Step 1: Write the test file**

Create `crypto/kdf_test.go`:

```go
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
```

**Step 2: Run test to verify it passes**

Run: `go test ./crypto -run TestDeriveKey -v`
Expected: PASS

**Step 3: Commit**

```bash
git add crypto/kdf_test.go
git commit -m "test: add Argon2id key derivation unit tests"
```

---

### Task 4: Add Encryptor Interface Tests

**Files:**
- Create: `crypto/crypto_test.go`

**Step 1: Write the test file**

Create `crypto/crypto_test.go`:

```go
package crypto

import (
	"bytes"
	"testing"
)

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		algo    Algorithm
		wantErr bool
	}{
		{AlgoAESGCM, false},
		{AlgoChaCha20, false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.algo), func(t *testing.T) {
			enc, err := NewEncryptor(tt.algo)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if enc == nil {
				t.Error("encryptor should not be nil")
			}
		})
	}
}

func TestEncryptorRoundtrip(t *testing.T) {
	algos := []Algorithm{AlgoAESGCM, AlgoChaCha20}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("test message for encryptor interface")

	for _, algo := range algos {
		t.Run(string(algo), func(t *testing.T) {
			enc, err := NewEncryptor(algo)
			if err != nil {
				t.Fatal(err)
			}

			ciphertext, err := enc.Encrypt(plaintext, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			decrypted, err := enc.Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("got %s, want %s", decrypted, plaintext)
			}
		})
	}
}

func TestEncryptorName(t *testing.T) {
	tests := []struct {
		algo Algorithm
		want string
	}{
		{AlgoAESGCM, "aes-gcm"},
		{AlgoChaCha20, "chacha20"},
	}

	for _, tt := range tests {
		enc, _ := NewEncryptor(tt.algo)
		if got := enc.Name(); got != tt.want {
			t.Errorf("Name() = %s, want %s", got, tt.want)
		}
	}
}

func TestEncryptToBase64(t *testing.T) {
	enc, _ := NewEncryptor(AlgoAESGCM)
	key := make([]byte, 32)
	plaintext := []byte("test")

	b64, err := EncryptToBase64(enc, plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	if len(b64) == 0 {
		t.Error("base64 output should not be empty")
	}

	decrypted, err := DecryptFromBase64(enc, b64, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("got %s, want %s", decrypted, plaintext)
	}
}

func TestDecryptFromBase64InvalidInput(t *testing.T) {
	enc, _ := NewEncryptor(AlgoAESGCM)
	key := make([]byte, 32)

	_, err := DecryptFromBase64(enc, "not-valid-base64!!!", key)
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./crypto -run "TestNewEncryptor|TestEncryptor" -v`
Expected: PASS

**Step 3: Commit**

```bash
git add crypto/crypto_test.go
git commit -m "test: add Encryptor interface unit tests"
```

---

### Task 5: Add .env Parser Tests

**Files:**
- Create: `parser/env_test.go`

**Step 1: Write the test file**

Create `parser/env_test.go`:

```go
package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantEnc   bool
		wantAlgo  string
		wantCmt   bool
	}{
		{"empty", "", "", "", false, "", true},
		{"comment", "# this is a comment", "", "", false, "", true},
		{"simple", "KEY=value", "KEY", "value", false, "", false},
		{"with_spaces", "  KEY = value  ", "KEY", " value  ", false, "", false},
		{"empty_value", "KEY=", "KEY", "", false, "", false},
		{"encrypted", "KEY=ENC[aes-gcm:YWJjZA==]", "KEY", "YWJjZA==", true, "aes-gcm", false},
		{"encrypted_age", "KEY=ENC[age:YWJjZA==]", "KEY", "YWJjZA==", true, "age", false},
		{"no_equals", "invalid line", "", "", false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := parseLine(tt.line)

			if entry.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", entry.Key, tt.wantKey)
			}
			if entry.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", entry.Value, tt.wantValue)
			}
			if entry.Encrypted != tt.wantEnc {
				t.Errorf("Encrypted = %v, want %v", entry.Encrypted, tt.wantEnc)
			}
			if entry.Algorithm != tt.wantAlgo {
				t.Errorf("Algorithm = %q, want %q", entry.Algorithm, tt.wantAlgo)
			}
			if entry.IsComment != tt.wantCmt {
				t.Errorf("IsComment = %v, want %v", entry.IsComment, tt.wantCmt)
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	content := `# Database config
DB_HOST=localhost
DB_PORT=5432
DB_PASS=ENC[aes-gcm:c2VjcmV0]

# Empty line above
API_KEY=test123
`
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := ParseEnvFile(envFile)
	if err != nil {
		t.Fatalf("ParseEnvFile error: %v", err)
	}

	if len(entries) != 7 {
		t.Errorf("got %d entries, want 7", len(entries))
	}

	// Check specific entries
	if entries[1].Key != "DB_HOST" || entries[1].Value != "localhost" {
		t.Error("DB_HOST not parsed correctly")
	}

	if !entries[3].Encrypted || entries[3].Algorithm != "aes-gcm" {
		t.Error("DB_PASS encryption not parsed correctly")
	}
}

func TestParseEnvFileNotFound(t *testing.T) {
	_, err := ParseEnvFile("/nonexistent/file.env")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestWriteEnvFile(t *testing.T) {
	entries := []EnvEntry{
		{IsComment: true, Comment: "# Config"},
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "encrypted", Encrypted: true, Algorithm: "aes-gcm"},
		{IsComment: true, Comment: ""},
	}

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, ".env.out")

	if err := WriteEnvFile(outFile, entries); err != nil {
		t.Fatalf("WriteEnvFile error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if content != "# Config\nKEY1=value1\nKEY2=ENC[aes-gcm:encrypted]\n\n" {
		t.Errorf("unexpected output:\n%s", content)
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./parser -v`
Expected: PASS

**Step 3: Commit**

```bash
git add parser/env_test.go
git commit -m "test: add .env parser unit tests"
```

---

### Task 6: Add Parser Encrypt/Decrypt Tests

**Files:**
- Modify: `parser/env_test.go` (append to existing)

**Step 1: Add encryption tests to existing file**

Append to `parser/env_test.go`:

```go

func TestEncryptDecryptEnvValues(t *testing.T) {
	// Create test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	entries := []EnvEntry{
		{IsComment: true, Comment: "# Header"},
		{Key: "SECRET1", Value: "password123"},
		{Key: "SECRET2", Value: "api-key-456"},
		{Key: "ALREADY_ENC", Value: "base64data", Encrypted: true, Algorithm: "aes-gcm"},
	}

	// Encrypt
	if err := EncryptEnvValues(entries, key, "aes-gcm"); err != nil {
		t.Fatalf("EncryptEnvValues error: %v", err)
	}

	// Verify encryption
	if !entries[1].Encrypted {
		t.Error("SECRET1 should be encrypted")
	}
	if entries[1].Algorithm != "aes-gcm" {
		t.Error("SECRET1 should use aes-gcm")
	}
	if entries[1].Value == "password123" {
		t.Error("SECRET1 value should be encrypted")
	}

	// Comment should be unchanged
	if !entries[0].IsComment {
		t.Error("comment should remain a comment")
	}

	// Already encrypted should be unchanged
	if entries[3].Value != "base64data" {
		t.Error("already encrypted value should not change")
	}

	// Decrypt
	if err := DecryptEnvValues(entries, key); err != nil {
		t.Fatalf("DecryptEnvValues error: %v", err)
	}

	// Verify decryption
	if entries[1].Encrypted {
		t.Error("SECRET1 should be decrypted")
	}
	if entries[1].Value != "password123" {
		t.Errorf("SECRET1 = %q, want %q", entries[1].Value, "password123")
	}
	if entries[2].Value != "api-key-456" {
		t.Errorf("SECRET2 = %q, want %q", entries[2].Value, "api-key-456")
	}
}

func TestEncryptEnvValuesInvalidAlgo(t *testing.T) {
	key := make([]byte, 32)
	entries := []EnvEntry{{Key: "K", Value: "V"}}

	err := EncryptEnvValues(entries, key, "invalid-algo")
	if err == nil {
		t.Error("expected error for invalid algorithm")
	}
}

func TestDecryptEnvValuesInvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	entries := []EnvEntry{
		{Key: "K", Value: "not-valid-base64!!!", Encrypted: true, Algorithm: "aes-gcm"},
	}

	err := DecryptEnvValues(entries, key)
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./parser -v`
Expected: PASS

**Step 3: Commit**

```bash
git add parser/env_test.go
git commit -m "test: add .env encrypt/decrypt tests"
```

---

### Task 7: Update CLAUDE.md with Test Documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update CLAUDE.md**

Add after "## Build Commands" section:

```markdown
## Testing

```bash
go test ./...                    # Run all tests
go test ./crypto -v              # Run crypto tests with verbose output
go test ./parser -v              # Run parser tests
go test ./... -cover             # Run with coverage
go test ./crypto -run TestAESGCM # Run specific test
```

Test files follow Go conventions:
- `crypto/aes_gcm_test.go` - AES-256-GCM encryption tests
- `crypto/chacha_test.go` - ChaCha20-Poly1305 tests
- `crypto/kdf_test.go` - Argon2id key derivation tests
- `crypto/crypto_test.go` - Encryptor interface tests
- `parser/env_test.go` - .env file parsing and encryption tests
```

**Step 2: Verify CLAUDE.md is valid**

Run: `cat CLAUDE.md | head -50`
Expected: See updated content with Testing section

**Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add testing documentation to CLAUDE.md"
```

---

### Task 8: Run Full Test Suite and Verify

**Files:**
- None (verification only)

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 2: Run with coverage**

Run: `go test ./... -cover`
Expected: Coverage report for crypto and parser packages

**Step 3: Verify build still works**

Run: `make build`
Expected: Binary builds successfully

**Step 4: Final commit with all tests**

Run: `git status`

If any uncommitted files, commit them:
```bash
git add -A
git commit -m "test: complete test suite for crypto and parser packages"
```
