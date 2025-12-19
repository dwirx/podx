package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hades/podx/keygen"
)

func TestEncryptFileDeletesOriginal(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "plain.txt")
	output := input + ".enc"

	if err := os.WriteFile(input, []byte("hello\n"), 0600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := encryptFile(input, output, "aes-gcm", "pass123"); err != nil {
		t.Fatalf("encryptFile: %v", err)
	}

	if _, err := os.Stat(input); !os.IsNotExist(err) {
		t.Fatalf("expected original to be removed, got err: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output to exist: %v", err)
	}
}

func TestEnvEncryptDeletesOriginal(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, ".env")
	output := input + ".podx"

	if err := os.WriteFile(input, []byte("API_KEY=secret\n"), 0600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := envEncryptFile(input, output, "aes-gcm", "pass123"); err != nil {
		t.Fatalf("envEncryptFile: %v", err)
	}

	if _, err := os.Stat(input); !os.IsNotExist(err) {
		t.Fatalf("expected original to be removed, got err: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output to exist: %v", err)
	}
}

func TestAgeEncryptDeletesOriginal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", filepath.Join(dir, "home"))

	if _, err := keygen.GenerateAge(); err != nil {
		t.Fatalf("GenerateAge: %v", err)
	}

	input := filepath.Join(dir, "age.txt")
	output := input + ".age"

	if err := os.WriteFile(input, []byte("hello age\n"), 0600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := encryptWithAge(input, output); err != nil {
		t.Fatalf("encryptWithAge: %v", err)
	}

	if _, err := os.Stat(input); !os.IsNotExist(err) {
		t.Fatalf("expected original to be removed, got err: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output to exist: %v", err)
	}
}
