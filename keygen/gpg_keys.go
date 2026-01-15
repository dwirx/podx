package keygen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const gpgKeysDir = "gpg-keys"

// GPGKeyEntry represents a single GPG key entry with metadata
type GPGKeyEntry struct {
	KeyID      string // GPG Key ID (e.g., ABCD1234EFGH5678)
	Name       string // User name
	Email      string // User email
	PublicKey  string // Armored public key
	PrivateKey string // Armored private key
	FilePath   string // Path to key files
}

// SaveGPGKeyPair saves a GPG key pair to the config directory
func SaveGPGKeyPair(keyID, publicKey, privateKey, name, email string) error {
	// Get config dir
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	// Create gpg-keys subdirectory
	gpgDir := filepath.Join(configDir, gpgKeysDir)
	if err := os.MkdirAll(gpgDir, 0700); err != nil {
		return fmt.Errorf("failed to create GPG keys directory: %w", err)
	}

	// Save public key to {keyID}_public.asc (mode 0644)
	pubKeyPath := filepath.Join(gpgDir, keyID+"_public.asc")
	if err := os.WriteFile(pubKeyPath, []byte(publicKey), 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	// Save private key to {keyID}_private.asc (mode 0600)
	privKeyPath := filepath.Join(gpgDir, keyID+"_private.asc")
	if err := os.WriteFile(privKeyPath, []byte(privateKey), 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save metadata to {keyID}_meta.txt
	metaPath := filepath.Join(gpgDir, keyID+"_meta.txt")
	metaContent := fmt.Sprintf("name: %s\nemail: %s\n", name, email)
	if err := os.WriteFile(metaPath, []byte(metaContent), 0644); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// ListGPGKeys returns all GPG keys from the config directory
func ListGPGKeys() ([]GPGKeyEntry, error) {
	// Get config dir
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	// Read gpg-keys directory
	gpgDir := filepath.Join(configDir, gpgKeysDir)
	entries, err := os.ReadDir(gpgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []GPGKeyEntry{}, nil
		}
		return nil, err
	}

	// Parse key files and metadata
	// Group files by key ID
	keyMap := make(map[string]*GPGKeyEntry)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Parse public key files
		if strings.HasSuffix(name, "_public.asc") {
			keyID := strings.TrimSuffix(name, "_public.asc")
			if keyMap[keyID] == nil {
				keyMap[keyID] = &GPGKeyEntry{KeyID: keyID}
			}
			pubKeyPath := filepath.Join(gpgDir, name)
			pubKeyData, err := os.ReadFile(pubKeyPath)
			if err == nil {
				keyMap[keyID].PublicKey = string(pubKeyData)
				keyMap[keyID].FilePath = pubKeyPath
			}
		}

		// Parse private key files
		if strings.HasSuffix(name, "_private.asc") {
			keyID := strings.TrimSuffix(name, "_private.asc")
			if keyMap[keyID] == nil {
				keyMap[keyID] = &GPGKeyEntry{KeyID: keyID}
			}
			privKeyPath := filepath.Join(gpgDir, name)
			privKeyData, err := os.ReadFile(privKeyPath)
			if err == nil {
				keyMap[keyID].PrivateKey = string(privKeyData)
			}
		}

		// Parse metadata files
		if strings.HasSuffix(name, "_meta.txt") {
			keyID := strings.TrimSuffix(name, "_meta.txt")
			if keyMap[keyID] == nil {
				keyMap[keyID] = &GPGKeyEntry{KeyID: keyID}
			}
			metaPath := filepath.Join(gpgDir, name)
			metaData, err := os.ReadFile(metaPath)
			if err == nil {
				// Parse simple format: "name: ...\nemail: ...\n"
				lines := strings.Split(string(metaData), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "name: ") {
						keyMap[keyID].Name = strings.TrimSpace(strings.TrimPrefix(line, "name: "))
					}
					if strings.HasPrefix(line, "email: ") {
						keyMap[keyID].Email = strings.TrimSpace(strings.TrimPrefix(line, "email: "))
					}
				}
			}
		}
	}

	// Convert map to array
	var keys []GPGKeyEntry
	for _, key := range keyMap {
		// Only include entries with both public and private keys
		if key.PublicKey != "" && key.PrivateKey != "" {
			keys = append(keys, *key)
		}
	}

	return keys, nil
}
