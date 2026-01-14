package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
	"runtime"

	"github.com/pbnjay/memory"
	"golang.org/x/crypto/argon2"
)

const (
	// Legacy Argon2 parameters (v1 compatibility)
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32 // 256-bit key

	// SaltSize ukuran salt (16 bytes recommended)
	SaltSize = 16

	// Minimum memory for Argon2id (64 MB)
	MinArgon2Memory = 64 * 1024

	// Default memory for normal mode (256 MB)
	DefaultNormalMemory = 256 * 1024

	// Default memory for paranoid mode (512 MB)
	DefaultParanoidMemory = 512 * 1024
)

// KDFMode represents the KDF operation mode
type KDFMode int

const (
	KDFModeNormal   KDFMode = iota // Normal mode: 4 passes, 256MB, 4 threads
	KDFModeParanoid                // Paranoid mode: 8 passes, 512MB, 8 threads
	KDFModeLegacy                  // Legacy mode: 3 passes, 64MB, 4 threads (v1 compat)
)

// KDFParams holds Argon2id parameters
type KDFParams struct {
	Time    uint32 // Number of passes
	Memory  uint32 // Memory in KB
	Threads uint8  // Number of threads
	KeyLen  uint32 // Output key length
}

// DefaultKDFParams returns default parameters for the given mode
func DefaultKDFParams(mode KDFMode) *KDFParams {
	switch mode {
	case KDFModeNormal:
		return &KDFParams{
			Time:    4,
			Memory:  DefaultNormalMemory,
			Threads: 4,
			KeyLen:  argon2KeyLen,
		}
	case KDFModeParanoid:
		return &KDFParams{
			Time:    8,
			Memory:  DefaultParanoidMemory,
			Threads: 8,
			KeyLen:  argon2KeyLen,
		}
	case KDFModeLegacy:
		return &KDFParams{
			Time:    argon2Time,
			Memory:  argon2Memory,
			Threads: argon2Threads,
			KeyLen:  argon2KeyLen,
		}
	default:
		return DefaultKDFParams(KDFModeNormal)
	}
}

// AdaptiveKDFParams returns parameters adapted to available system memory
func AdaptiveKDFParams(mode KDFMode, configuredMemoryMB uint32) *KDFParams {
	params := DefaultKDFParams(mode)

	// Override with configured memory if specified
	if configuredMemoryMB > 0 {
		params.Memory = configuredMemoryMB * 1024 // Convert MB to KB
	}

	// Get available system memory
	availableMemory := getAvailableMemory()

	// Require at least 2x the memory for safe operation
	requiredMemory := uint64(params.Memory) * 1024 * 2 // in bytes

	if availableMemory < requiredMemory {
		// Reduce memory if not enough available
		reducedMemory := uint32(availableMemory / 1024 / 2) // Use half of available, in KB

		// Ensure minimum memory
		if reducedMemory < MinArgon2Memory {
			reducedMemory = MinArgon2Memory
		}

		params.Memory = reducedMemory
		fmt.Printf("Warning: Reduced Argon2id memory to %d MB due to low available memory\n", reducedMemory/1024)
	}

	// Adapt threads to available CPUs
	numCPU := runtime.NumCPU()
	if int(params.Threads) > numCPU {
		params.Threads = uint8(numCPU)
	}

	return params
}

// getAvailableMemory returns available system memory in bytes
func getAvailableMemory() uint64 {
	// Try to get free memory
	free := memory.FreeMemory()
	if free > 0 {
		return free
	}

	// Fallback to total memory / 4
	total := memory.TotalMemory()
	if total > 0 {
		return total / 4
	}

	// Default fallback: assume 1GB available
	return 1024 * 1024 * 1024
}

// DeriveKey menghasilkan 256-bit key dari password menggunakan Argon2id.
// Uses legacy parameters for backward compatibility.
// Returns: key (32 bytes), salt (16 bytes), error
func DeriveKey(password []byte, salt []byte) ([]byte, []byte, error) {
	return DeriveKeyWithParams(password, salt, DefaultKDFParams(KDFModeLegacy))
}

// DeriveKeyWithSalt menghasilkan key dari password dengan salt yang sudah ada.
// Uses legacy parameters for backward compatibility.
func DeriveKeyWithSalt(password, salt []byte) ([]byte, error) {
	key, _, err := DeriveKeyWithParams(password, salt, DefaultKDFParams(KDFModeLegacy))
	return key, err
}

// DeriveKeyWithMode derives a key using parameters for the specified mode
func DeriveKeyWithMode(password, salt []byte, mode KDFMode) ([]byte, []byte, error) {
	params := AdaptiveKDFParams(mode, 0)
	return DeriveKeyWithParams(password, salt, params)
}

// DeriveKeyWithParams derives a key using custom parameters
func DeriveKeyWithParams(password, salt []byte, params *KDFParams) ([]byte, []byte, error) {
	// Generate salt if not provided
	if salt == nil {
		salt = make([]byte, SaltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	if len(salt) != SaltSize {
		return nil, nil, fmt.Errorf("invalid salt size: expected %d bytes, got %d", SaltSize, len(salt))
	}

	// Derive key using Argon2id
	key := argon2.IDKey(password, salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	return key, salt, nil
}

// DeriveKeyV2 derives a key for v2 format with mode support
// Returns the master key which should be used with DeriveAllSubkeys for actual encryption keys
func DeriveKeyV2(password, salt []byte, mode KDFMode, configuredMemoryMB uint32) ([]byte, []byte, *KDFParams, error) {
	params := AdaptiveKDFParams(mode, configuredMemoryMB)

	key, salt, err := DeriveKeyWithParams(password, salt, params)
	if err != nil {
		return nil, nil, nil, err
	}

	return key, salt, params, nil
}

// KDFParamsFromFlags converts mode and memory flags to KDF params
func KDFParamsFromFlags(paranoid bool, memoryMB uint32) *KDFParams {
	mode := KDFModeNormal
	if paranoid {
		mode = KDFModeParanoid
	}
	return AdaptiveKDFParams(mode, memoryMB)
}
