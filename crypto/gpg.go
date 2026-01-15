package crypto

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// GPGEncrypt encrypts plaintext with GPG using recipient ID (email or key ID)
// or armored public key. Automatically detects armored keys and uses native
// implementation, otherwise falls back to shell GPG.
func GPGEncrypt(plaintext []byte, recipient string) ([]byte, error) {
	// Detect if recipient is an armored public key
	if bytes.Contains([]byte(recipient), []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----")) {
		// Use native implementation for armored public keys
		return GPGEncryptNative(plaintext, recipient)
	}

	// Fall back to shell GPG for recipient IDs (email/key ID)
	return gpgEncryptShell(plaintext, recipient)
}

// gpgEncryptShell encrypts using external GPG binary (legacy implementation)
func gpgEncryptShell(plaintext []byte, recipient string) ([]byte, error) {
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

// GPGDecrypt decrypts ciphertext with GPG using local keyring (shell GPG).
// For native decryption with a provided private key, use GPGDecryptWithKey.
func GPGDecrypt(ciphertext []byte) ([]byte, error) {
	return gpgDecryptShell(ciphertext)
}

// GPGDecryptWithKey decrypts ciphertext using either native implementation
// (if privateKey provided) or shell GPG (if privateKey is empty).
func GPGDecryptWithKey(ciphertext []byte, privateKey string, passphrase string) ([]byte, error) {
	if privateKey != "" {
		// Use native implementation when private key is provided
		return GPGDecryptNative(ciphertext, privateKey, passphrase)
	}
	// Fall back to shell GPG
	return gpgDecryptShell(ciphertext)
}

// gpgDecryptShell decrypts using external GPG binary (legacy implementation)
func gpgDecryptShell(ciphertext []byte) ([]byte, error) {
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

// GenerateGPGKey generates a new GPG key pair using native implementation.
// Returns the key ID (email). For full key material (private/public keys),
// use GenerateGPGKeyNative directly.
func GenerateGPGKey(name, email, passphrase string) (string, error) {
	// Use native implementation
	keyID, _, _, err := GenerateGPGKeyNative(name, email, passphrase)
	if err != nil {
		return "", err
	}
	return keyID, nil
}

// generateGPGKeyShell generates a GPG key using external GPG binary (legacy implementation)
func generateGPGKeyShell(name, email, passphrase string) (string, error) {
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

// GenerateGPGKeyNative generates a new GPG key pair using native Go implementation
// Returns: keyID (email), privateKey (armored), publicKey (armored), error
func GenerateGPGKeyNative(name, email, passphrase string) (string, string, string, error) {
	rsaBits := 4096

	// Generate key using gopenpgp - USE CORRECT API from research!
	key, err := crypto.GenerateKey(name, email, "rsa", rsaBits)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate PGP key: %w", err)
	}

	// Get armored private key
	privateKeyArmored, err := key.Armor()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to armor private key: %w", err)
	}

	// Get armored public key
	publicKeyArmored, err := key.GetArmoredPublicKey()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get public key: %w", err)
	}

	// If passphrase provided, lock the key
	if passphrase != "" {
		locked, err := key.Lock([]byte(passphrase))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to lock key with passphrase: %w", err)
		}
		privateKeyArmored, err = locked.Armor()
		if err != nil {
			return "", "", "", fmt.Errorf("failed to armor locked key: %w", err)
		}
	}

	// Use email as key ID for consistency
	return email, privateKeyArmored, publicKeyArmored, nil
}

// GPGEncryptNative encrypts plaintext using native Go PGP implementation
// publicKeyArmored should be an armored PGP public key
func GPGEncryptNative(plaintext []byte, publicKeyArmored string) ([]byte, error) {
	if len(publicKeyArmored) == 0 {
		return nil, fmt.Errorf("public key is empty")
	}

	// Parse public key
	publicKeyObj, err := crypto.NewKeyFromArmored(publicKeyArmored)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Create a keyring with the public key
	keyRing, err := crypto.NewKeyRing(publicKeyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring: %w", err)
	}

	// Create a plaintext message from bytes
	message := crypto.NewPlainMessage(plaintext)

	// Encrypt the message
	pgpMessage, err := keyRing.Encrypt(message, nil)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Get armored ciphertext
	armoredMessage, err := pgpMessage.GetArmored()
	if err != nil {
		return nil, fmt.Errorf("failed to armor message: %w", err)
	}

	return []byte(armoredMessage), nil
}

// GPGDecryptNative decrypts ciphertext using native Go PGP implementation
// privateKeyArmored should be an armored PGP private key
// passphrase should be empty string if key is not password-protected
func GPGDecryptNative(ciphertext []byte, privateKeyArmored string, passphrase string) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext is empty")
	}

	if len(privateKeyArmored) == 0 {
		return nil, fmt.Errorf("private key is empty")
	}

	// Parse private key
	privateKeyObj, err := crypto.NewKeyFromArmored(privateKeyArmored)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Unlock key if passphrase is provided
	if passphrase != "" {
		privateKeyObj, err = privateKeyObj.Unlock([]byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to unlock key: %w", err)
		}
		defer privateKeyObj.ClearPrivateParams()
	}

	// Create keyring
	keyRing, err := crypto.NewKeyRing(privateKeyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring: %w", err)
	}

	// Parse the encrypted message
	pgpMessage, err := crypto.NewPGPMessageFromArmored(string(ciphertext))
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted message: %w", err)
	}

	// Decrypt the message
	plainMessage, err := keyRing.Decrypt(pgpMessage, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plainMessage.GetBinary(), nil
}

// GPGEncryptMultipleNative encrypts plaintext for multiple recipients using native Go PGP implementation
// publicKeysArmored should be a slice of armored PGP public keys
func GPGEncryptMultipleNative(plaintext []byte, publicKeysArmored []string) ([]byte, error) {
	if len(publicKeysArmored) == 0 {
		return nil, fmt.Errorf("no public keys provided")
	}

	// Parse all public keys and add to keyring
	var keyRing *crypto.KeyRing
	for i, pubKeyArmored := range publicKeysArmored {
		publicKeyObj, err := crypto.NewKeyFromArmored(pubKeyArmored)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key %d: %w", i, err)
		}

		if keyRing == nil {
			// Create keyring with first key
			keyRing, err = crypto.NewKeyRing(publicKeyObj)
			if err != nil {
				return nil, fmt.Errorf("failed to create keyring: %w", err)
			}
		} else {
			// Add subsequent keys to the keyring
			err = keyRing.AddKey(publicKeyObj)
			if err != nil {
				return nil, fmt.Errorf("failed to add public key %d to keyring: %w", i, err)
			}
		}
	}

	// Create a plaintext message from bytes
	message := crypto.NewPlainMessage(plaintext)

	// Encrypt the message for all recipients
	pgpMessage, err := keyRing.Encrypt(message, nil)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Get armored ciphertext
	armoredMessage, err := pgpMessage.GetArmored()
	if err != nil {
		return nil, fmt.Errorf("failed to armor message: %w", err)
	}

	return []byte(armoredMessage), nil
}
