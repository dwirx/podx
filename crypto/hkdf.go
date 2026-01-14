package crypto

import (
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

// SubkeyType identifies the purpose of a derived subkey
type SubkeyType int

const (
	SubkeyHeaderHMAC SubkeyType = iota // 64 bytes - for header authentication
	SubkeyPayloadMAC                   // 32 bytes - for payload MAC (BLAKE2b or HMAC-SHA3)
	SubkeyCipher                       // 32 bytes - for AES/XChaCha20
	SubkeySerpent                      // 32 bytes - for Serpent (paranoid mode)
)

// SubkeyInfo maps subkey types to their HKDF info strings
var SubkeyInfo = map[SubkeyType]string{
	SubkeyHeaderHMAC: "podx-header-hmac-v2",
	SubkeyPayloadMAC: "podx-payload-mac-v2",
	SubkeyCipher:     "podx-cipher-key-v2",
	SubkeySerpent:    "podx-serpent-key-v2",
}

// SubkeySizes maps subkey types to their sizes in bytes
var SubkeySizes = map[SubkeyType]int{
	SubkeyHeaderHMAC: 64,
	SubkeyPayloadMAC: 32,
	SubkeyCipher:     32,
	SubkeySerpent:    32,
}

// DerivedKeys holds all subkeys derived from a master key
type DerivedKeys struct {
	HeaderHMAC []byte // 64 bytes
	PayloadMAC []byte // 32 bytes
	Cipher     []byte // 32 bytes
	Serpent    []byte // 32 bytes (only used in paranoid mode)
}

// DeriveSubkey derives a single subkey from the master key using HKDF-SHA3-256
func DeriveSubkey(masterKey, salt []byte, keyType SubkeyType) ([]byte, error) {
	info, ok := SubkeyInfo[keyType]
	if !ok {
		return nil, fmt.Errorf("unknown subkey type: %d", keyType)
	}

	size, ok := SubkeySizes[keyType]
	if !ok {
		return nil, fmt.Errorf("unknown subkey size for type: %d", keyType)
	}

	// HKDF using SHA3-256
	hkdfReader := hkdf.New(sha3.New256, masterKey, salt, []byte(info))

	subkey := make([]byte, size)
	if _, err := io.ReadFull(hkdfReader, subkey); err != nil {
		return nil, fmt.Errorf("failed to derive subkey: %w", err)
	}

	return subkey, nil
}

// DeriveAllSubkeys derives all subkeys needed for encryption
// For normal mode, Serpent key is nil
// For paranoid mode, all keys including Serpent are derived
func DeriveAllSubkeys(masterKey, salt []byte, paranoidMode bool) (*DerivedKeys, error) {
	keys := &DerivedKeys{}
	var err error

	// Derive header HMAC key
	keys.HeaderHMAC, err = DeriveSubkey(masterKey, salt, SubkeyHeaderHMAC)
	if err != nil {
		return nil, fmt.Errorf("failed to derive header HMAC key: %w", err)
	}

	// Derive payload MAC key
	keys.PayloadMAC, err = DeriveSubkey(masterKey, salt, SubkeyPayloadMAC)
	if err != nil {
		return nil, fmt.Errorf("failed to derive payload MAC key: %w", err)
	}

	// Derive cipher key
	keys.Cipher, err = DeriveSubkey(masterKey, salt, SubkeyCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to derive cipher key: %w", err)
	}

	// Derive Serpent key only for paranoid mode
	if paranoidMode {
		keys.Serpent, err = DeriveSubkey(masterKey, salt, SubkeySerpent)
		if err != nil {
			return nil, fmt.Errorf("failed to derive Serpent key: %w", err)
		}
	}

	return keys, nil
}

// DeriveNonce derives a nonce for a specific chunk index (for streaming encryption)
func DeriveNonce(masterKey, salt []byte, chunkIndex uint64, nonceSize int) ([]byte, error) {
	info := fmt.Sprintf("podx-nonce-v2-chunk-%d", chunkIndex)

	hkdfReader := hkdf.New(sha3.New256, masterKey, salt, []byte(info))

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(hkdfReader, nonce); err != nil {
		return nil, fmt.Errorf("failed to derive nonce: %w", err)
	}

	return nonce, nil
}

// DeriveRekeyMaterial derives new key material for rekeying (paranoid mode)
// This is used for rekeying every 1GB in streaming encryption
func DeriveRekeyMaterial(masterKey, salt []byte, rekeyIndex uint64) (*DerivedKeys, error) {
	// Create a new salt based on rekey index
	info := fmt.Sprintf("podx-rekey-v2-%d", rekeyIndex)

	hkdfReader := hkdf.New(sha3.New256, masterKey, salt, []byte(info))

	// Derive a new "master key" for this rekey segment
	newMasterKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, newMasterKey); err != nil {
		return nil, fmt.Errorf("failed to derive rekey material: %w", err)
	}

	// Derive subkeys from the new master key
	return DeriveAllSubkeys(newMasterKey, salt, true)
}
