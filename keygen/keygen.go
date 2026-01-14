package keygen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hades/podx/crypto"
)

const (
	configDir        = ".config/podx"
	ageKeysFile      = "age-keys.txt"
	ageRecipientsDir = "age-recipients"
)

// KeygenResult contains the result of key generation
type KeygenResult struct {
	Backend    string
	KeyFile    string
	PublicKey  string
	PrivateKey string
	Email      string
	Name       string // Key name/label
}

// GetConfigDir returns the podx config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

// EnsureConfigDir creates the config directory if not exists
func EnsureConfigDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create age-recipients subdirectory
	recipientsDir := filepath.Join(dir, ageRecipientsDir)
	if err := os.MkdirAll(recipientsDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create recipients directory: %w", err)
	}

	return dir, nil
}

// GenerateAge generates a new Age key pair and saves it
func GenerateAge() (*KeygenResult, error) {
	return GenerateAgeWithName("")
}

// GenerateAgeWithName generates a new Age key pair with a name/label and saves it
func GenerateAgeWithName(name string) (*KeygenResult, error) {
	privateKey, publicKey, err := crypto.GenerateAgeKey()
	if err != nil {
		return nil, err
	}

	configDir, err := EnsureConfigDir()
	if err != nil {
		return nil, err
	}

	keyFile := filepath.Join(configDir, ageKeysFile)

	// Set default name if empty
	if name == "" {
		// Count existing keys to generate default name
		existingKeys, _ := ListAgeKeys()
		name = fmt.Sprintf("Key #%d", len(existingKeys)+1)
	}

	// Append to keys file with standard age format:
	// # name: <name>
	// # created: <timestamp>
	// # public key: <public_key>
	// AGE-SECRET-KEY-...
	content := fmt.Sprintf("# name: %s\n# created: %s\n# public key: %s\n%s\n",
		name, time.Now().Format(time.RFC3339), publicKey, privateKey)

	f, err := os.OpenFile(keyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return nil, fmt.Errorf("failed to write key: %w", err)
	}

	// Save public key to recipients directory
	pubKeyFile := filepath.Join(configDir, ageRecipientsDir, "default.txt")
	if err := os.WriteFile(pubKeyFile, []byte(publicKey+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to save public key: %w", err)
	}

	return &KeygenResult{
		Backend:    "age",
		KeyFile:    keyFile,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Name:       name,
	}, nil
}

// GenerateGPG generates a new GPG key pair
func GenerateGPG(name, email string) (*KeygenResult, error) {
	if !crypto.CheckGPGInstalled() {
		return nil, fmt.Errorf("gpg is not installed. Please install GPG first")
	}

	keyID, err := crypto.GenerateGPGKey(name, email, "")
	if err != nil {
		return nil, err
	}

	return &KeygenResult{
		Backend:   "gpg",
		PublicKey: keyID,
		Email:     email,
	}, nil
}

// LoadAgeIdentity loads the Age private key from config
func LoadAgeIdentity() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	keyFile := filepath.Join(configDir, ageKeysFile)
	identity, err := parseAgeIdentityFromFile(keyFile)
	if err == nil {
		return identity, nil
	}

	return "", fmt.Errorf("no age identity found in %s. Generate with 'podx keygen -t age'", keyFile)
}

// parseAgeIdentityFromFile parses an Age identity from a key file
func parseAgeIdentityFromFile(keyFile string) (string, error) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return "", err
	}

	// Parse the key file - find the last valid identity
	lines := strings.Split(string(data), "\n")
	var identity string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			identity = line
		}
	}

	if identity == "" {
		return "", fmt.Errorf("no age identity found in %s", keyFile)
	}

	return identity, nil
}

// LoadAgeRecipient loads the Age public key from config
// It first tries the dedicated recipients file, then falls back to parsing age-keys.txt
func LoadAgeRecipient() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	// Try the dedicated public key file first
	pubKeyFile := filepath.Join(configDir, ageRecipientsDir, "default.txt")
	if data, err := os.ReadFile(pubKeyFile); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Fallback: parse public key from age-keys.txt
	keyFile := filepath.Join(configDir, ageKeysFile)
	if data, err := os.ReadFile(keyFile); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Look for "# public key: age1..."
			if strings.HasPrefix(line, "# public key:") {
				pubKey := strings.TrimSpace(strings.TrimPrefix(line, "# public key:"))
				if strings.HasPrefix(pubKey, "age1") {
					return pubKey, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no age public key found. Generate with 'podx keygen -t age'")
}

// PrintKeygenResult displays the key generation result in a beautiful box
func PrintKeygenResult(result *KeygenResult) {
	width := 70

	// Top border
	fmt.Printf("╔%s╗\n", strings.Repeat("═", width))

	// Title
	title := "🔑 PODX Key Generated Successfully"
	padding := (width - len(title)) / 2
	fmt.Printf("║%s%s%s║\n", strings.Repeat(" ", padding), title, strings.Repeat(" ", width-padding-len(title)))

	// Separator
	fmt.Printf("╠%s╣\n", strings.Repeat("═", width))

	// Backend
	printRow("Backend:", result.Backend, width)

	// Key file (for Age)
	if result.KeyFile != "" {
		// Shorten path
		shortPath := result.KeyFile
		if home, err := os.UserHomeDir(); err == nil {
			shortPath = strings.Replace(result.KeyFile, home, "~", 1)
		}
		printRow("Key file:", shortPath, width)
	}

	// Email (for GPG)
	if result.Email != "" {
		printRow("Email:", result.Email, width)
	}

	// Separator
	fmt.Printf("╠%s╣\n", strings.Repeat("═", width))

	// Keys section for Age
	if result.Backend == "age" {
		printRow("Public Key:", "", width)
		printRow("  "+result.PublicKey, "", width)
		fmt.Printf("║%s║\n", strings.Repeat(" ", width))
		printRow("Private Key:", "", width)
		printRow("  "+result.PrivateKey, "", width)
	} else {
		printRow("Key ID:", result.PublicKey, width)
	}

	// Bottom border
	fmt.Printf("╚%s╝\n", strings.Repeat("═", width))

	// Additional info
	if result.Backend == "age" {
		fmt.Println()
		fmt.Println("📋 Public key copied to: ~/.config/podx/age-recipients/default.txt")
		fmt.Println("🔐 Private key saved to: ~/.config/podx/age-keys.txt")
	}
}

func printRow(label, value string, width int) {
	content := label
	if value != "" {
		content = label + " " + value
	}

	// Truncate if too long
	if len(content) > width-2 {
		content = content[:width-5] + "..."
	}

	fmt.Printf("║ %s%s║\n", content, strings.Repeat(" ", width-len(content)-1))
}

// GetKeyInfo returns information about the current Age key configuration
type KeyInfo struct {
	PublicKey   string
	KeyFilePath string
	HasKey      bool
}

// GetAgeKeyInfo returns information about the current Age key configuration
func GetAgeKeyInfo() KeyInfo {
	info := KeyInfo{}

	// Try to load public key
	pubKey, err := LoadAgeRecipient()
	if err == nil {
		info.PublicKey = pubKey
		info.HasKey = true
	}

	// Get key file path
	configDir, err := GetConfigDir()
	if err == nil {
		info.KeyFilePath = filepath.Join(configDir, ageKeysFile)
	}

	return info
}

// AgeKeyEntry represents a single Age key entry with metadata
type AgeKeyEntry struct {
	Name       string // Key name/label (e.g., team member name)
	CreatedAt  string // Creation timestamp
	PublicKey  string // Public key (age1...)
	PrivateKey string // Private key (AGE-SECRET-KEY-...)
	Index      int    // Index in the key file
}

// ListAgeKeys returns all Age keys from the config file
func ListAgeKeys() ([]AgeKeyEntry, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	keyFile := filepath.Join(configDir, ageKeysFile)
	data, err := os.ReadFile(keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []AgeKeyEntry{}, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var keys []AgeKeyEntry
	var currentEntry AgeKeyEntry
	keyIndex := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse name comment
		if strings.HasPrefix(line, "# name:") {
			currentEntry.Name = strings.TrimSpace(strings.TrimPrefix(line, "# name:"))
			continue
		}

		// Parse created timestamp
		if strings.HasPrefix(line, "# created:") {
			currentEntry.CreatedAt = strings.TrimSpace(strings.TrimPrefix(line, "# created:"))
			continue
		}

		// Parse public key comment
		if strings.HasPrefix(line, "# public key:") {
			currentEntry.PublicKey = strings.TrimSpace(strings.TrimPrefix(line, "# public key:"))
			continue
		}

		// Parse private key
		if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			currentEntry.PrivateKey = line
			currentEntry.Index = keyIndex
			// Set default name if empty
			if currentEntry.Name == "" {
				currentEntry.Name = fmt.Sprintf("Key #%d", keyIndex+1)
			}
			keys = append(keys, currentEntry)
			currentEntry = AgeKeyEntry{}
			keyIndex++
		}
	}

	return keys, nil
}

// ImportAgeKey imports a private key from a string and saves it
func ImportAgeKey(privateKey string) (*KeygenResult, error) {
	return ImportAgeKeyWithName(privateKey, "")
}

// ImportAgeKeyWithName imports a private key with a name/label and saves it
func ImportAgeKeyWithName(privateKey, name string) (*KeygenResult, error) {
	privateKey = strings.TrimSpace(privateKey)

	// Validate private key format
	if !strings.HasPrefix(privateKey, "AGE-SECRET-KEY-") {
		return nil, fmt.Errorf("invalid Age private key (must start with 'AGE-SECRET-KEY-')")
	}

	// Derive public key from private key using age
	publicKey, err := crypto.DeriveAgePublicKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	configDir, err := EnsureConfigDir()
	if err != nil {
		return nil, err
	}

	keyFile := filepath.Join(configDir, ageKeysFile)

	// Check if key already exists
	existingKeys, _ := ListAgeKeys()
	for _, k := range existingKeys {
		if k.PrivateKey == privateKey {
			return nil, fmt.Errorf("this key already exists in your key store")
		}
	}

	// Set default name if empty
	if name == "" {
		name = fmt.Sprintf("Imported Key #%d", len(existingKeys)+1)
	}

	// Append to keys file with standard age format
	content := fmt.Sprintf("# name: %s\n# created: %s\n# public key: %s\n%s\n",
		name, time.Now().Format(time.RFC3339), publicKey, privateKey)

	f, err := os.OpenFile(keyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return nil, fmt.Errorf("failed to write key: %w", err)
	}

	// Update default public key file
	pubKeyFile := filepath.Join(configDir, ageRecipientsDir, "default.txt")
	if err := os.WriteFile(pubKeyFile, []byte(publicKey+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to save public key: %w", err)
	}

	return &KeygenResult{
		Backend:    "age",
		KeyFile:    keyFile,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Name:       name,
	}, nil
}

// SetDefaultKey sets a specific key as the default for encryption/decryption
func SetDefaultKey(index int) error {
	keys, err := ListAgeKeys()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(keys) {
		return fmt.Errorf("invalid key index: %d", index)
	}

	selectedKey := keys[index]

	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Update default public key file
	pubKeyFile := filepath.Join(configDir, ageRecipientsDir, "default.txt")
	if err := os.WriteFile(pubKeyFile, []byte(selectedKey.PublicKey+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to set default key: %w", err)
	}

	return nil
}

// DeleteKey removes a key from the key store by index
func DeleteKey(index int) error {
	keys, err := ListAgeKeys()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(keys) {
		return fmt.Errorf("invalid key index: %d", index)
	}

	if len(keys) == 1 {
		return fmt.Errorf("cannot delete the only key - generate a new one first")
	}

	// Rebuild the key file without the deleted key
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	keyFile := filepath.Join(configDir, ageKeysFile)

	var content strings.Builder
	for i, k := range keys {
		if i != index {
			content.WriteString(fmt.Sprintf("# name: %s\n# created: %s\n# public key: %s\n%s\n",
				k.Name, k.CreatedAt, k.PublicKey, k.PrivateKey))
		}
	}

	if err := os.WriteFile(keyFile, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to update key file: %w", err)
	}

	// If we deleted the default key, set the first remaining key as default
	currentDefault, _ := LoadAgeRecipient()
	deletedKey := keys[index]
	if currentDefault == deletedKey.PublicKey && len(keys) > 1 {
		newDefaultIndex := 0
		if index == 0 {
			newDefaultIndex = 1
		}
		_ = SetDefaultKey(newDefaultIndex)
	}

	return nil
}

// RenameKey renames a key in the key store
func RenameKey(index int, newName string) error {
	keys, err := ListAgeKeys()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(keys) {
		return fmt.Errorf("invalid key index: %d", index)
	}

	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Update the name
	keys[index].Name = newName

	// Rebuild the key file with updated name
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	keyFile := filepath.Join(configDir, ageKeysFile)

	var content strings.Builder
	for _, k := range keys {
		content.WriteString(fmt.Sprintf("# name: %s\n# created: %s\n# public key: %s\n%s\n",
			k.Name, k.CreatedAt, k.PublicKey, k.PrivateKey))
	}

	if err := os.WriteFile(keyFile, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to update key file: %w", err)
	}

	return nil
}
