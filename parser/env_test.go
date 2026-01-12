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

	// Save encrypted value to verify already-encrypted entries are skipped
	encryptedValue := entries[1].Value

	// Try to encrypt again - already encrypted should be unchanged
	if err := EncryptEnvValues(entries, key, "aes-gcm"); err != nil {
		t.Fatalf("EncryptEnvValues (second call) error: %v", err)
	}
	if entries[1].Value != encryptedValue {
		t.Error("already encrypted value should not change on re-encryption")
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
