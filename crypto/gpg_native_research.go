// +build ignore

// This file tests the gopenpgp v2 API to understand how to implement native GPG support.
//
// Key findings from successful test run:
// 1. Use crypto.GenerateKey() to create RSA/EdDSA keys (not crypto.PGP().GenerateKey())
// 2. Key.Armor() returns the armored private key (~6.8KB for 4096-bit RSA)
// 3. Key.GetArmoredPublicKey() returns the armored public key (~3.3KB for 4096-bit RSA)
// 4. helper.EncryptMessageArmored() encrypts with a public key string
// 5. helper.DecryptMessageArmored() decrypts with private key and passphrase
// 6. Both private and public keys are in ASCII-armored format (PEM-like)
// 7. Passphrase must be nil for unencrypted keys (not empty []byte(""))
// 8. The API is very clean - no need for manual OpenPGP packet manipulation
// 9. Encrypted messages are ASCII-armored by default (~919 bytes for 16-byte plaintext)
// 10. Key generation for 4096-bit RSA takes a few seconds
//
// Implementation notes for crypto/gpg.go:
// - Use helper.GenerateKey() for key generation (wrapper around crypto.GenerateKey)
// - Store keys in ~/.config/podx/gpg-keys.txt (private) and gpg-recipients/ (public)
// - For encryption: helper.EncryptMessageArmored(publicKey, plaintext)
// - For decryption: helper.DecryptMessageArmored(privateKey, passphrase, ciphertext)
// - Support both locked (with passphrase) and unlocked (nil passphrase) keys

package main

import (
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
)

func main() {
	// Test 1: Key generation
	fmt.Println("=== Testing Key Generation ===")
	keyName := "Test User"
	keyEmail := "test@podx.local"

	rsaBits := 4096
	key, err := crypto.GenerateKey(keyName, keyEmail, "rsa", rsaBits)
	if err != nil {
		fmt.Printf("Key generation failed: %v\n", err)
		return
	}

	armoredPrivateKey, err := key.Armor()
	if err != nil {
		fmt.Printf("Armor failed: %v\n", err)
		return
	}
	fmt.Printf("Generated private key (armored): %d bytes\n", len(armoredPrivateKey))

	publicKey, err := key.GetArmoredPublicKey()
	if err != nil {
		fmt.Printf("Get public key failed: %v\n", err)
		return
	}
	fmt.Printf("Public key: %d bytes\n", len(publicKey))

	// Test 2: Encryption
	fmt.Println("\n=== Testing Encryption ===")
	plaintext := "Hello, GopenPGP!"

	ciphertext, err := helper.EncryptMessageArmored(publicKey, plaintext)
	if err != nil {
		fmt.Printf("Encryption failed: %v\n", err)
		return
	}
	fmt.Printf("Encrypted message: %d bytes\n", len(ciphertext))

	// Test 3: Decryption
	fmt.Println("\n=== Testing Decryption ===")

	decrypted, err := helper.DecryptMessageArmored(armoredPrivateKey, nil, ciphertext)
	if err != nil {
		fmt.Printf("Decryption failed: %v\n", err)
		return
	}
	fmt.Printf("Decrypted: %s\n", decrypted)

	if decrypted != plaintext {
		fmt.Println("ERROR: Decrypted != Plaintext")
		return
	}

	fmt.Println("\n=== All tests passed! ===")
}
