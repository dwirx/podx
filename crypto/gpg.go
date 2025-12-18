package crypto

import (
	"bytes"
	"fmt"
	"os/exec"
)

// GPGEncrypt mengenkripsi plaintext dengan GPG menggunakan recipient ID (email atau key ID)
func GPGEncrypt(plaintext []byte, recipient string) ([]byte, error) {
	cmd := exec.Command("gpg", "--encrypt", "--armor", "--recipient", recipient, "--batch", "--yes")
	cmd.Stdin = bytes.NewReader(plaintext)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gpg encrypt failed: %s: %w", stderr.String(), err)
	}

	return stdout.Bytes(), nil
}

// GPGDecrypt mendekripsi ciphertext dengan GPG (menggunakan keyring lokal)
func GPGDecrypt(ciphertext []byte) ([]byte, error) {
	cmd := exec.Command("gpg", "--decrypt", "--batch", "--yes")
	cmd.Stdin = bytes.NewReader(ciphertext)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gpg decrypt failed: %s: %w", stderr.String(), err)
	}

	return stdout.Bytes(), nil
}

// GenerateGPGKey generates a new GPG key pair
func GenerateGPGKey(name, email, passphrase string) (string, error) {
	keyParams := fmt.Sprintf(`%%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: %s
Name-Email: %s
Expire-Date: 0
%%commit
`, name, email)

	cmd := exec.Command("gpg", "--batch", "--gen-key")
	cmd.Stdin = bytes.NewReader([]byte(keyParams))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gpg key generation failed: %s: %w", stderr.String(), err)
	}

	// Get the key fingerprint
	listCmd := exec.Command("gpg", "--list-keys", "--keyid-format", "long", email)
	var listOut bytes.Buffer
	listCmd.Stdout = &listOut

	if err := listCmd.Run(); err != nil {
		return email, nil // Return email as identifier
	}

	return email, nil
}

// CheckGPGInstalled verifies GPG is available
func CheckGPGInstalled() bool {
	_, err := exec.LookPath("gpg")
	return err == nil
}
