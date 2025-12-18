package project

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hades/podx/crypto"
	"github.com/hades/podx/keygen"
)

const (
	ConfigFileName = ".podx.yaml"
	EncryptedExt   = ".podx"
)

// Recipient represents a team member who can decrypt secrets
type Recipient struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

// Config represents .podx.yaml project configuration
type Config struct {
	Version    int         `yaml:"version"`
	Backend    string      `yaml:"backend"`
	Recipients []Recipient `yaml:"recipients"`
	Secrets    []string    `yaml:"secrets"`
}

// Project represents a PODX-enabled project
type Project struct {
	RootDir string
	Config  *Config
}

// Init initializes a new PODX project in the current directory
func Init(dir string) (*Project, error) {
	configPath := filepath.Join(dir, ConfigFileName)

	// Check if already initialized
	if _, err := os.Stat(configPath); err == nil {
		return nil, fmt.Errorf("project already initialized (.podx.yaml exists)")
	}

	// Try to load user's default Age public key
	var defaultRecipient *Recipient
	if pubKey, err := keygen.LoadAgeRecipient(); err == nil {
		defaultRecipient = &Recipient{
			Name: "Owner",
			Key:  pubKey,
		}
	}

	// Create default config
	config := &Config{
		Version: 1,
		Backend: "age",
		Secrets: []string{".env"},
	}

	if defaultRecipient != nil {
		config.Recipients = []Recipient{*defaultRecipient}
	}

	// Save config
	project := &Project{
		RootDir: dir,
		Config:  config,
	}

	if err := project.Save(); err != nil {
		return nil, err
	}

	// Update .gitignore
	if err := project.UpdateGitignore(); err != nil {
		// Non-fatal, just warn
		fmt.Printf("⚠️  Could not update .gitignore: %v\n", err)
	}

	return project, nil
}

// Load loads an existing PODX project from directory
func Load(dir string) (*Project, error) {
	configPath := filepath.Join(dir, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("no .podx.yaml found. Run 'podx init' first")
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid .podx.yaml: %w", err)
	}

	return &Project{
		RootDir: dir,
		Config:  &config,
	}, nil
}

// Save writes the config to .podx.yaml
func (p *Project) Save() error {
	data, err := yaml.Marshal(p.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := "# PODX Project Configuration\n# https://github.com/hades/podx\n\n"
	content := header + string(data)

	configPath := filepath.Join(p.RootDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddRecipient adds a new recipient to the project
func (p *Project) AddRecipient(name, key string) error {
	// Validate age key format
	if !strings.HasPrefix(key, "age1") {
		return fmt.Errorf("invalid Age public key (should start with 'age1')")
	}

	// Check for duplicate
	for _, r := range p.Config.Recipients {
		if r.Key == key {
			return fmt.Errorf("recipient with this key already exists")
		}
	}

	p.Config.Recipients = append(p.Config.Recipients, Recipient{
		Name: name,
		Key:  key,
	})

	return p.Save()
}

// AddSecret adds a new secret file pattern to the project
func (p *Project) AddSecret(pattern string) error {
	// Check for duplicate
	for _, s := range p.Config.Secrets {
		if s == pattern {
			return fmt.Errorf("secret pattern already exists")
		}
	}

	p.Config.Secrets = append(p.Config.Secrets, pattern)
	return p.Save()
}

// EncryptAll encrypts all secret files in the project
func (p *Project) EncryptAll() (int, error) {
	if len(p.Config.Recipients) == 0 {
		return 0, fmt.Errorf("no recipients configured. Add with 'podx add-recipient'")
	}

	// Get all recipient public keys
	var recipientKeys []string
	for _, r := range p.Config.Recipients {
		recipientKeys = append(recipientKeys, r.Key)
	}

	count := 0
	for _, pattern := range p.Config.Secrets {
		// Expand glob pattern
		matches, err := filepath.Glob(filepath.Join(p.RootDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Skip already encrypted files
			if strings.HasSuffix(match, EncryptedExt) {
				continue
			}

			relPath, _ := filepath.Rel(p.RootDir, match)
			baseName := filepath.Base(match)

			// Check if it's a .env file - use format-preserving encryption
			if strings.HasPrefix(baseName, ".env") || strings.HasSuffix(baseName, ".env") {
				if err := p.encryptEnvFile(match, recipientKeys); err != nil {
					return count, fmt.Errorf("failed to encrypt %s: %w", relPath, err)
				}
				fmt.Printf("✓ Encrypted: %s → %s%s (format-preserving)\n", relPath, relPath, EncryptedExt)
			} else {
				// Regular file encryption
				if err := p.encryptRegularFile(match, recipientKeys); err != nil {
					return count, fmt.Errorf("failed to encrypt %s: %w", relPath, err)
				}
				fmt.Printf("✓ Encrypted: %s → %s%s\n", relPath, relPath, EncryptedExt)
			}

			// Delete original file after successful encryption
			if err := os.Remove(match); err != nil {
				fmt.Printf("⚠️  Could not delete original: %s\n", err)
			} else {
				fmt.Printf("🗑️  Deleted original: %s\n", relPath)
			}

			count++
		}
	}

	return count, nil
}

// encryptEnvFile encrypts a .env file preserving its format (KEY=ENC[...] format)
func (p *Project) encryptEnvFile(filePath string, recipientKeys []string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			result = append(result, line)
			continue
		}

		// Parse KEY=VALUE
		idx := strings.Index(line, "=")
		if idx == -1 {
			result = append(result, line)
			continue
		}

		key := line[:idx]
		value := line[idx+1:]

		// Skip already encrypted values
		if strings.HasPrefix(value, "ENC[") {
			result = append(result, line)
			continue
		}

		// Encrypt value with Age
		ciphertext, err := crypto.AgeEncrypt([]byte(value), recipientKeys...)
		if err != nil {
			return fmt.Errorf("failed to encrypt key '%s': %w", key, err)
		}

		// Format as ENC[age:base64]
		encValue := fmt.Sprintf("ENC[age:%s]", base64.StdEncoding.EncodeToString(ciphertext))
		result = append(result, fmt.Sprintf("%s=%s", key, encValue))
	}

	// Write encrypted .env file
	encPath := filePath + EncryptedExt
	return os.WriteFile(encPath, []byte(strings.Join(result, "\n")), 0644)
}

// encryptRegularFile encrypts a regular file (binary encryption)
func (p *Project) encryptRegularFile(filePath string, recipientKeys []string) error {
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	ciphertext, err := crypto.AgeEncrypt(plaintext, recipientKeys...)
	if err != nil {
		return err
	}

	encPath := filePath + EncryptedExt
	return os.WriteFile(encPath, ciphertext, 0644)
}

// DecryptAll decrypts all encrypted secret files
func (p *Project) DecryptAll() (int, error) {
	// Load user's identity
	identity, err := keygen.LoadAgeIdentity()
	if err != nil {
		return 0, fmt.Errorf("no Age identity found. Generate with 'podx keygen -t age'")
	}

	count := 0
	for _, pattern := range p.Config.Secrets {
		// Look for encrypted versions
		encPattern := filepath.Join(p.RootDir, pattern+EncryptedExt)
		matches, err := filepath.Glob(encPattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			decPath := strings.TrimSuffix(match, EncryptedExt)
			relPath, _ := filepath.Rel(p.RootDir, decPath)
			baseName := filepath.Base(decPath)

			// Check if it's a .env file
			if strings.HasPrefix(baseName, ".env") || strings.HasSuffix(baseName, ".env") {
				if err := p.decryptEnvFile(match, decPath, identity); err != nil {
					return count, fmt.Errorf("failed to decrypt %s: %w", relPath, err)
				}
				fmt.Printf("✓ Decrypted: %s%s → %s (format-preserving)\n", relPath, EncryptedExt, relPath)
			} else {
				if err := p.decryptRegularFile(match, decPath, identity); err != nil {
					return count, fmt.Errorf("failed to decrypt %s: %w", relPath, err)
				}
				fmt.Printf("✓ Decrypted: %s%s → %s\n", relPath, EncryptedExt, relPath)
			}

			count++
		}
	}

	return count, nil
}

// decryptEnvFile decrypts a format-preserving .env file
func (p *Project) decryptEnvFile(encPath, decPath, identity string) error {
	data, err := os.ReadFile(encPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			result = append(result, line)
			continue
		}

		// Parse KEY=VALUE
		idx := strings.Index(line, "=")
		if idx == -1 {
			result = append(result, line)
			continue
		}

		key := line[:idx]
		value := line[idx+1:]

		// Check if value is encrypted (ENC[age:base64])
		if strings.HasPrefix(value, "ENC[age:") && strings.HasSuffix(value, "]") {
			// Extract base64 ciphertext
			encData := strings.TrimPrefix(value, "ENC[age:")
			encData = strings.TrimSuffix(encData, "]")

			ciphertext, err := base64.StdEncoding.DecodeString(encData)
			if err != nil {
				return fmt.Errorf("invalid base64 for key '%s': %w", key, err)
			}

			plaintext, err := crypto.AgeDecrypt(ciphertext, identity)
			if err != nil {
				return fmt.Errorf("failed to decrypt key '%s': %w", key, err)
			}

			result = append(result, fmt.Sprintf("%s=%s", key, string(plaintext)))
		} else {
			// Value is not encrypted, keep as-is
			result = append(result, line)
		}
	}

	return os.WriteFile(decPath, []byte(strings.Join(result, "\n")), 0600)
}

// decryptRegularFile decrypts a binary encrypted file
func (p *Project) decryptRegularFile(encPath, decPath, identity string) error {
	ciphertext, err := os.ReadFile(encPath)
	if err != nil {
		return err
	}

	plaintext, err := crypto.AgeDecrypt(ciphertext, identity)
	if err != nil {
		return err
	}

	return os.WriteFile(decPath, plaintext, 0600)
}

// UpdateGitignore adds decrypted secret patterns to .gitignore
func (p *Project) UpdateGitignore() error {
	gitignorePath := filepath.Join(p.RootDir, ".gitignore")

	var existing []string
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existing = strings.Split(string(data), "\n")
	}

	// Patterns to add
	toAdd := []string{
		"",
		"# PODX - Decrypted secrets (DO NOT COMMIT)",
	}
	for _, secret := range p.Config.Secrets {
		toAdd = append(toAdd, secret)
	}

	// Check which patterns are already present
	var newPatterns []string
	for _, pattern := range toAdd {
		found := false
		for _, line := range existing {
			if strings.TrimSpace(line) == strings.TrimSpace(pattern) {
				found = true
				break
			}
		}
		if !found {
			newPatterns = append(newPatterns, pattern)
		}
	}

	if len(newPatterns) == 0 {
		return nil // Already up to date
	}

	// Append to .gitignore
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, pattern := range newPatterns {
		if _, err := f.WriteString(pattern + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// Status returns project status information
func (p *Project) Status() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📁 Project: %s\n", p.RootDir))
	sb.WriteString(fmt.Sprintf("🔐 Backend: %s\n", p.Config.Backend))
	sb.WriteString(fmt.Sprintf("👥 Recipients: %d\n", len(p.Config.Recipients)))

	for _, r := range p.Config.Recipients {
		sb.WriteString(fmt.Sprintf("   - %s (%s...)\n", r.Name, r.Key[:20]))
	}

	sb.WriteString(fmt.Sprintf("📄 Secrets: %d patterns\n", len(p.Config.Secrets)))
	for _, s := range p.Config.Secrets {
		sb.WriteString(fmt.Sprintf("   - %s\n", s))
	}

	return sb.String()
}

// PrintInitSuccess displays success message after init
func PrintInitSuccess(p *Project) {
	width := 70

	fmt.Printf("╔%s╗\n", strings.Repeat("═", width))
	fmt.Printf("║%s║\n", centerText("🎉 PODX Project Initialized", width))
	fmt.Printf("╠%s╣\n", strings.Repeat("═", width))

	// Config file
	fmt.Printf("║ %-68s ║\n", "Config: .podx.yaml")

	// Recipients
	if len(p.Config.Recipients) > 0 {
		fmt.Printf("║ %-68s ║\n", fmt.Sprintf("Recipients: %d", len(p.Config.Recipients)))
		for _, r := range p.Config.Recipients {
			keyShort := r.Key
			if len(keyShort) > 40 {
				keyShort = keyShort[:40] + "..."
			}
			fmt.Printf("║   %-66s ║\n", fmt.Sprintf("• %s: %s", r.Name, keyShort))
		}
	} else {
		fmt.Printf("║ %-68s ║\n", "⚠️  No recipients. Add with: podx add-recipient -n NAME -k KEY")
	}

	// Secrets
	fmt.Printf("║ %-68s ║\n", fmt.Sprintf("Secrets: %s", strings.Join(p.Config.Secrets, ", ")))

	fmt.Printf("╠%s╣\n", strings.Repeat("═", width))
	fmt.Printf("║ %-68s ║\n", "Next steps:")
	fmt.Printf("║   %-66s ║\n", "1. Add recipients: podx add-recipient -n 'Name' -k age1...")
	fmt.Printf("║   %-66s ║\n", "2. Encrypt secrets: podx encrypt-all")
	fmt.Printf("║   %-66s ║\n", "3. Commit .podx.yaml and *.podx files")
	fmt.Printf("╚%s╝\n", strings.Repeat("═", width))
}

func centerText(text string, width int) string {
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", padding), text, strings.Repeat(" ", width-padding-len(text)))
}
