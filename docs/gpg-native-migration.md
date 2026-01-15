# GPG Native Implementation Migration Guide

## Overview

PODX now uses native Go PGP implementation (gopenpgp) by default, eliminating the need for external GPG binary installation while maintaining full backward compatibility with shell-based GPG.

## Benefits

### Pure Go Implementation
- **No External Dependencies**: Works without GPG installed on the system
- **Cross-Platform**: Consistent behavior across Linux, macOS, and Windows
- **Portable**: Single binary deployment without system requirements
- **Fast**: No process spawning overhead for encryption/decryption

### Enhanced Features
- **Password-Protected Keys**: Native support for passphrase-protected private keys
- **Multi-Recipient**: Encrypt for multiple recipients in a single operation
- **Full Control**: Direct access to key material (private/public keys)
- **Better Testing**: Pure Go tests that work on any system

### Backward Compatibility
- **Auto-Detection**: Automatically selects native or shell GPG based on input
- **Existing Code**: No changes required for most existing code
- **GPG Interop**: Works with GPG-generated keys and vice versa

## Architecture

### Function Migration

| Old Function | New Behavior | Legacy Fallback |
|-------------|-------------|-----------------|
| `GPGEncrypt(plaintext, recipient)` | Auto-detects armored public key → native, otherwise → shell | `gpgEncryptShell()` |
| `GPGDecrypt(ciphertext)` | Uses shell GPG with local keyring | `gpgDecryptShell()` |
| `GenerateGPGKey(name, email, pass)` | Uses native implementation, returns keyID only | `generateGPGKeyShell()` |

### New Functions

| Function | Purpose |
|----------|---------|
| `GenerateGPGKeyNative(name, email, pass)` | Generate keys, returns (keyID, privateKey, publicKey, error) |
| `GPGEncryptNative(plaintext, publicKey)` | Encrypt with armored public key |
| `GPGEncryptMultipleNative(plaintext, []publicKeys)` | Encrypt for multiple recipients |
| `GPGDecryptNative(ciphertext, privateKey, pass)` | Decrypt with armored private key |
| `GPGDecryptWithKey(ciphertext, privateKey, pass)` | Auto-select native (if key provided) or shell |

## Migration Examples

### Example 1: Key Generation

**Before (Shell GPG):**
```go
// Required GPG installed on system
keyID, err := crypto.GenerateGPGKey("Alice", "alice@example.com", "")
if err != nil {
    return err
}
// Only returns keyID (email), keys stored in GPG keyring
```

**After (Native):**
```go
// No GPG required
keyID, privateKey, publicKey, err := crypto.GenerateGPGKeyNative("Alice", "alice@example.com", "")
if err != nil {
    return err
}
// Returns full key material for storage/distribution
// Can save privateKey securely, share publicKey with team
```

**Backward Compatible:**
```go
// Still works, uses native internally but only returns keyID
keyID, err := crypto.GenerateGPGKey("Alice", "alice@example.com", "")
```

### Example 2: Encryption

**Before (Shell GPG only):**
```go
// Required recipient's key in GPG keyring
ciphertext, err := crypto.GPGEncrypt(plaintext, "alice@example.com")
```

**After (Auto-detect):**
```go
// Option 1: Use armored public key (native, no GPG required)
publicKey := "-----BEGIN PGP PUBLIC KEY BLOCK-----\n..."
ciphertext, err := crypto.GPGEncrypt(plaintext, publicKey)

// Option 2: Use recipient ID (shell GPG, for compatibility)
ciphertext, err := crypto.GPGEncrypt(plaintext, "alice@example.com")
```

**Multi-Recipient (New):**
```go
// Encrypt for multiple recipients
publicKeys := []string{alicePublicKey, bobPublicKey, carolPublicKey}
ciphertext, err := crypto.GPGEncryptMultipleNative(plaintext, publicKeys)
// All three can decrypt with their respective private keys
```

### Example 3: Decryption

**Before (Shell GPG only):**
```go
// Required private key in GPG keyring
plaintext, err := crypto.GPGDecrypt(ciphertext)
```

**After (Auto-select):**
```go
// Option 1: Use armored private key (native, no GPG required)
plaintext, err := crypto.GPGDecryptWithKey(ciphertext, privateKey, "passphrase")

// Option 2: Use shell GPG (for compatibility with local keyring)
plaintext, err := crypto.GPGDecrypt(ciphertext)
// or
plaintext, err := crypto.GPGDecryptWithKey(ciphertext, "", "")
```

### Example 4: Password-Protected Keys

**New Feature (Native only):**
```go
// Generate password-protected key
keyID, privateKey, publicKey, err := crypto.GenerateGPGKeyNative(
    "Alice",
    "alice@example.com",
    "my-strong-passphrase",
)

// Public key can be shared (no password needed for encryption)
ciphertext, err := crypto.GPGEncryptNative(plaintext, publicKey)

// Private key requires passphrase for decryption
plaintext, err := crypto.GPGDecryptNative(ciphertext, privateKey, "my-strong-passphrase")

// Wrong passphrase returns error
_, err = crypto.GPGDecryptNative(ciphertext, privateKey, "wrong-passphrase")
// err: failed to unlock key
```

## Usage Patterns

### Pattern 1: Team Secrets Management (Recommended)

**Use native implementation with key storage:**

```go
// 1. Each team member generates a key pair
keyID, privateKey, publicKey, err := crypto.GenerateGPGKeyNative(
    "Alice Developer",
    "alice@company.com",
    "secure-passphrase",
)

// 2. Store private key securely (e.g., ~/.config/podx/pgp-keys.txt)
err = savePrivateKey(privateKey)

// 3. Share public key with team (e.g., commit to repo)
err = addToTeamRecipients(publicKey)

// 4. Encrypt secrets for all team members
teamPublicKeys := loadTeamPublicKeys() // []string
ciphertext, err := crypto.GPGEncryptMultipleNative(secretData, teamPublicKeys)

// 5. Each member can decrypt with their private key
plaintext, err := crypto.GPGDecryptNative(ciphertext, privateKey, "secure-passphrase")
```

### Pattern 2: System Integration (Backward Compatible)

**Use shell GPG for integration with existing GPG infrastructure:**

```go
// Works with system GPG keyring
ciphertext, err := crypto.GPGEncrypt(plaintext, "team@company.com")
plaintext, err := crypto.GPGDecrypt(ciphertext)
```

### Pattern 3: Hybrid Approach

**Use native for new keys, shell for existing keys:**

```go
// Generate new keys with native implementation
keyID, privateKey, publicKey, err := crypto.GenerateGPGKeyNative("New User", "new@company.com", "")

// Encrypt for both new (native) and existing (shell GPG) recipients
// Encrypt with public key (native)
ciphertext1, err := crypto.GPGEncryptNative(plaintext, publicKey)

// Encrypt with recipient ID (shell GPG)
ciphertext2, err := crypto.GPGEncrypt(plaintext, "existing@company.com")
```

## Testing

All GPG tests now work without requiring GPG installation:

```bash
# Run all GPG tests (native tests always run)
go test ./crypto -run TestGPG -v

# Shell GPG tests skip if GPG not installed
# Native tests always run and pass

# Benchmarks
go test ./crypto -bench BenchmarkGPG -benchmem
```

**Example output:**
```
=== RUN   TestGPGEncryptDecrypt
    gpg_test.go:17: GPG not installed
--- SKIP: TestGPGEncryptDecrypt (shell GPG test)

=== RUN   TestGPGEncryptDecryptNative
--- PASS: TestGPGEncryptDecryptNative (0.90s)

=== RUN   TestGPGPasswordProtectedKeyNative
--- PASS: TestGPGPasswordProtectedKeyNative (2.18s)
```

## API Reference

### GenerateGPGKeyNative

```go
func GenerateGPGKeyNative(name, email, passphrase string) (keyID, privateKey, publicKey string, err error)
```

Generates a new 4096-bit RSA PGP key pair using native Go implementation.

**Parameters:**
- `name`: User's full name (e.g., "Alice Developer")
- `email`: User's email address (e.g., "alice@company.com"), used as keyID
- `passphrase`: Optional passphrase to protect private key (empty string for no protection)

**Returns:**
- `keyID`: Email address as key identifier
- `privateKey`: Armored private key (PGP format)
- `publicKey`: Armored public key (PGP format)
- `err`: Error if generation fails

**Example:**
```go
keyID, privKey, pubKey, err := crypto.GenerateGPGKeyNative("Alice", "alice@example.com", "")
```

### GPGEncryptNative

```go
func GPGEncryptNative(plaintext []byte, publicKeyArmored string) ([]byte, error)
```

Encrypts plaintext using native Go PGP implementation with an armored public key.

**Parameters:**
- `plaintext`: Data to encrypt
- `publicKeyArmored`: Armored PGP public key (-----BEGIN PGP PUBLIC KEY BLOCK-----)

**Returns:**
- Armored ciphertext (PGP MESSAGE format)
- Error if encryption fails

**Example:**
```go
ciphertext, err := crypto.GPGEncryptNative([]byte("secret"), publicKey)
```

### GPGEncryptMultipleNative

```go
func GPGEncryptMultipleNative(plaintext []byte, publicKeysArmored []string) ([]byte, error)
```

Encrypts plaintext for multiple recipients using native implementation.

**Parameters:**
- `plaintext`: Data to encrypt
- `publicKeysArmored`: Slice of armored PGP public keys

**Returns:**
- Armored ciphertext that can be decrypted by any recipient
- Error if encryption fails

**Example:**
```go
publicKeys := []string{alicePubKey, bobPubKey}
ciphertext, err := crypto.GPGEncryptMultipleNative(secretData, publicKeys)
```

### GPGDecryptNative

```go
func GPGDecryptNative(ciphertext []byte, privateKeyArmored, passphrase string) ([]byte, error)
```

Decrypts ciphertext using native Go PGP implementation with an armored private key.

**Parameters:**
- `ciphertext`: Armored PGP message to decrypt
- `privateKeyArmored`: Armored PGP private key
- `passphrase`: Passphrase to unlock private key (empty string if not protected)

**Returns:**
- Decrypted plaintext
- Error if decryption fails or passphrase is wrong

**Example:**
```go
plaintext, err := crypto.GPGDecryptNative(ciphertext, privateKey, "my-passphrase")
```

### GPGDecryptWithKey

```go
func GPGDecryptWithKey(ciphertext []byte, privateKey, passphrase string) ([]byte, error)
```

Auto-selects native implementation (if privateKey provided) or shell GPG (if empty).

**Parameters:**
- `ciphertext`: Encrypted data to decrypt
- `privateKey`: Optional armored private key (empty for shell GPG)
- `passphrase`: Optional passphrase (empty if not needed)

**Returns:**
- Decrypted plaintext
- Error if decryption fails

**Example:**
```go
// Use native with private key
plaintext, err := crypto.GPGDecryptWithKey(ciphertext, privateKey, "pass")

// Use shell GPG
plaintext, err := crypto.GPGDecryptWithKey(ciphertext, "", "")
```

## Performance Characteristics

### Key Generation
- **Native**: ~1-2 seconds for 4096-bit RSA
- **Shell GPG**: Similar, plus process spawning overhead

### Encryption (1KB data)
- **Native**: ~5-10ms (single recipient)
- **Native**: ~15-20ms (3 recipients)
- **Shell GPG**: ~20-50ms (process spawning)

### Decryption (1KB data)
- **Native**: ~3-5ms
- **Shell GPG**: ~20-50ms (process spawning)

### Large Data (1MB)
- **Native Encryption**: ~50-100ms
- **Native Decryption**: ~30-60ms
- **Shell GPG**: Similar raw performance, but higher overhead

## Security Considerations

### Key Storage
- **Private Keys**: Store securely with filesystem permissions (0600)
- **Passphrase Protection**: Recommended for private keys
- **Public Keys**: Can be safely distributed (0644)

### Memory Safety
- Native implementation clears private key parameters after use
- Passphrase strings should be cleared after use in production code

### Cryptographic Standards
- **Algorithm**: RSA 4096-bit keys (compatible with GPG)
- **Format**: OpenPGP RFC 4880 compliant
- **Interoperability**: Works with GPG, Thunderbird, ProtonMail, etc.

## Troubleshooting

### "Failed to parse public key"
- Ensure public key is in armored format (-----BEGIN PGP PUBLIC KEY BLOCK-----)
- Check for whitespace or encoding issues

### "Failed to unlock key"
- Verify passphrase is correct
- Ensure private key was generated with passphrase protection

### "Decryption failed"
- Check if ciphertext was encrypted for the correct public key
- Verify ciphertext wasn't tampered with
- Ensure private key matches the public key used for encryption

### Shell GPG tests skip
- Expected behavior when GPG not installed
- Native tests provide coverage
- Install GPG if shell GPG functionality is needed

## Migration Checklist

- [ ] Update key generation to use `GenerateGPGKeyNative` and store full key material
- [ ] Replace recipient IDs with armored public keys where possible
- [ ] Use `GPGEncryptNative` for new encryption code
- [ ] Use `GPGDecryptNative` with stored private keys
- [ ] Add passphrase protection for sensitive private keys
- [ ] Update tests to use native implementation
- [ ] Consider multi-recipient encryption for team secrets
- [ ] Document public key distribution method for your team
- [ ] Verify backward compatibility with existing encrypted data

## See Also

- [crypto/gpg.go](../crypto/gpg.go) - Implementation and examples
- [crypto/gpg_test.go](../crypto/gpg_test.go) - Test cases and usage patterns
- [gopenpgp documentation](https://pkg.go.dev/github.com/ProtonMail/gopenpgp/v2/crypto)
- [OpenPGP RFC 4880](https://www.rfc-editor.org/rfc/rfc4880)
