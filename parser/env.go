package parser

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hades/podx/crypto"
)

// EnvEntry merepresentasikan satu baris dalam file .env
type EnvEntry struct {
	Key       string
	Value     string
	Encrypted bool
	Algorithm string
	Comment   string   // Jika baris adalah komentar atau empty
	Raw       string   // Baris original
	IsComment bool
}

// encryptedValuePattern untuk parsing ENC[algo:base64] format
var encryptedValuePattern = regexp.MustCompile(`^ENC\[([a-zA-Z0-9-]+):(.+)\]$`)

// ParseEnvFile membaca dan mem-parse file .env
func ParseEnvFile(path string) ([]EnvEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var entries []EnvEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLine(line)
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return entries, nil
}

// parseLine mem-parse satu baris .env
func parseLine(line string) EnvEntry {
	trimmed := strings.TrimSpace(line)

	// Empty line atau comment
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return EnvEntry{
			Raw:       line,
			IsComment: true,
			Comment:   line,
		}
	}

	// Parse KEY=VALUE
	idx := strings.Index(line, "=")
	if idx == -1 {
		// Bukan format valid, treat sebagai comment
		return EnvEntry{
			Raw:       line,
			IsComment: true,
			Comment:   line,
		}
	}

	key := strings.TrimSpace(line[:idx])
	value := line[idx+1:]

	// Cek apakah value sudah dienkripsi (ENC[algo:base64])
	if matches := encryptedValuePattern.FindStringSubmatch(value); matches != nil {
		return EnvEntry{
			Key:       key,
			Value:     matches[2], // base64 ciphertext
			Encrypted: true,
			Algorithm: matches[1],
			Raw:       line,
		}
	}

	return EnvEntry{
		Key:       key,
		Value:     value,
		Encrypted: false,
		Raw:       line,
	}
}

// WriteEnvFile menulis entries ke file .env
func WriteEnvFile(path string, entries []EnvEntry) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for _, entry := range entries {
		var line string
		if entry.IsComment {
			line = entry.Comment
		} else if entry.Encrypted {
			line = fmt.Sprintf("%s=ENC[%s:%s]", entry.Key, entry.Algorithm, entry.Value)
		} else {
			line = fmt.Sprintf("%s=%s", entry.Key, entry.Value)
		}
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	return nil
}

// EncryptEnvValues mengenkripsi semua nilai dalam entries
func EncryptEnvValues(entries []EnvEntry, key []byte, algo crypto.Algorithm) error {
	enc, err := crypto.NewEncryptor(algo)
	if err != nil {
		return err
	}

	for i := range entries {
		if entries[i].IsComment || entries[i].Encrypted {
			continue
		}

		ciphertext, err := enc.Encrypt([]byte(entries[i].Value), key)
		if err != nil {
			return fmt.Errorf("failed to encrypt key '%s': %w", entries[i].Key, err)
		}

		entries[i].Value = base64.StdEncoding.EncodeToString(ciphertext)
		entries[i].Encrypted = true
		entries[i].Algorithm = string(algo)
	}

	return nil
}

// DecryptEnvValues mendekripsi semua nilai dalam entries
func DecryptEnvValues(entries []EnvEntry, key []byte) error {
	for i := range entries {
		if entries[i].IsComment || !entries[i].Encrypted {
			continue
		}

		// Pilih encryptor berdasarkan algorithm yang tersimpan
		enc, err := crypto.NewEncryptor(crypto.Algorithm(entries[i].Algorithm))
		if err != nil {
			return fmt.Errorf("unsupported algorithm '%s' for key '%s': %w", 
				entries[i].Algorithm, entries[i].Key, err)
		}

		ciphertext, err := base64.StdEncoding.DecodeString(entries[i].Value)
		if err != nil {
			return fmt.Errorf("failed to decode base64 for key '%s': %w", entries[i].Key, err)
		}

		plaintext, err := enc.Decrypt(ciphertext, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt key '%s': %w", entries[i].Key, err)
		}

		entries[i].Value = string(plaintext)
		entries[i].Encrypted = false
		entries[i].Algorithm = ""
	}

	return nil
}
