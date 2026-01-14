package crypto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// StreamingThreshold is the file size threshold for streaming encryption (100MB)
const StreamingThreshold = 100 * 1024 * 1024

// ProgressCallback is called during streaming operations with progress info
type ProgressCallback func(bytesProcessed, totalBytes int64)

// StreamingEncryptFile encrypts a file using streaming mode
func StreamingEncryptFile(inputPath, outputPath string, password []byte, opts *EncryptOptions, progress ProgressCallback) error {
	if opts == nil {
		opts = DefaultEncryptOptions(ModeNormal)
	}
	opts.Streaming = true

	// Open input file
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	// Get file size
	stat, err := input.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}
	totalSize := stat.Size()

	// Create output file
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	// Determine KDF mode
	kdfMode := KDFModeNormal
	if opts.Mode == ModeParanoid {
		kdfMode = KDFModeParanoid
	}

	// Derive master key
	masterKey, salt, _, err := DeriveKeyV2(password, nil, kdfMode, opts.MemoryMB)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	// Derive subkeys
	paranoid := opts.Mode == ModeParanoid
	keys, err := DeriveAllSubkeys(masterKey, salt, paranoid)
	if err != nil {
		return fmt.Errorf("subkey derivation failed: %w", err)
	}

	// Build and write header
	flags := &FileFlags{
		Mode:      opts.Mode,
		Cipher:    opts.Cipher,
		Streaming: true,
	}

	header := make([]byte, HeaderSizeV2)
	copy(header[0:4], MagicV2)
	header[4] = FormatVersionV2
	header[5] = flags.ToByte()
	copy(header[6:22], salt)

	// Compute header HMAC
	headerData := header[0:22]
	headerMAC, err := ComputeHMACSHA3(headerData, keys.HeaderHMAC[:32])
	if err != nil {
		return fmt.Errorf("header MAC failed: %w", err)
	}
	copy(header[22:86], headerMAC)

	if _, err := output.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write total chunk count placeholder (will update at end)
	chunkCountOffset := int64(HeaderSizeV2)
	if _, err := output.Write(make([]byte, 8)); err != nil {
		return fmt.Errorf("failed to write chunk count placeholder: %w", err)
	}

	// Initialize MAC for entire payload
	var payloadMAC MAC
	if paranoid {
		payloadMAC, err = NewHMACSHA3(keys.PayloadMAC)
	} else {
		payloadMAC, err = NewBlake2bMAC(keys.PayloadMAC)
	}
	if err != nil {
		return fmt.Errorf("failed to create payload MAC: %w", err)
	}

	// Encrypt in chunks
	buffer := make([]byte, StreamingChunkSize)
	var chunkIndex uint64
	var bytesProcessed int64
	var currentKeys = keys

	for {
		// Check for rekeying in paranoid mode
		if paranoid && chunkIndex > 0 && chunkIndex%RekeyInterval == 0 {
			rekeyIdx := chunkIndex / RekeyInterval
			currentKeys, err = DeriveRekeyMaterial(masterKey, salt, rekeyIdx)
			if err != nil {
				return fmt.Errorf("rekeying failed at chunk %d: %w", chunkIndex, err)
			}
		}

		// Read chunk
		n, readErr := input.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]

			// Derive nonce for this chunk
			nonce, err := DeriveNonce(masterKey, salt, chunkIndex, XChaChaNonceSize)
			if err != nil {
				return fmt.Errorf("nonce derivation failed: %w", err)
			}

			// Encrypt chunk
			var encryptedChunk []byte
			switch opts.Cipher {
			case CipherAESGCM:
				// For AES-GCM, derive a 12-byte nonce
				aesNonce, _ := DeriveNonce(masterKey, salt, chunkIndex, AESNonceSize)
				encryptedChunk, err = AESGCMEncryptWithNonce(chunk, currentKeys.Cipher, aesNonce)
			case CipherXChaCha20:
				encryptedChunk, err = XChaCha20EncryptWithNonce(chunk, currentKeys.Cipher, nonce)
			case CipherCascade:
				serpentNonce := DeriveSerpentNonce(chunkIndex)
				encryptedChunk, err = CascadeEncryptWithNonces(chunk, currentKeys.Cipher, currentKeys.Serpent, nonce, serpentNonce)
			default:
				return fmt.Errorf("unknown cipher type: %d", opts.Cipher)
			}
			if err != nil {
				return fmt.Errorf("chunk encryption failed: %w", err)
			}

			// Write chunk length (4 bytes) + encrypted chunk
			chunkLen := make([]byte, 4)
			binary.BigEndian.PutUint32(chunkLen, uint32(len(encryptedChunk)))
			if _, err := output.Write(chunkLen); err != nil {
				return fmt.Errorf("failed to write chunk length: %w", err)
			}
			if _, err := output.Write(encryptedChunk); err != nil {
				return fmt.Errorf("failed to write encrypted chunk: %w", err)
			}

			// Update payload MAC
			payloadMAC.Write(chunkLen)
			payloadMAC.Write(encryptedChunk)

			chunkIndex++
			bytesProcessed += int64(n)

			if progress != nil {
				progress(bytesProcessed, totalSize)
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read error: %w", readErr)
		}
	}

	// Write final MAC
	finalMAC := payloadMAC.Sum(nil)
	if _, err := output.Write(finalMAC); err != nil {
		return fmt.Errorf("failed to write payload MAC: %w", err)
	}

	// Update chunk count at placeholder position
	if _, err := output.Seek(chunkCountOffset, 0); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}
	chunkCount := make([]byte, 8)
	binary.BigEndian.PutUint64(chunkCount, chunkIndex)
	if _, err := output.Write(chunkCount); err != nil {
		return fmt.Errorf("failed to write chunk count: %w", err)
	}

	return nil
}

// StreamingDecryptFile decrypts a streaming-encrypted file
func StreamingDecryptFile(inputPath, outputPath string, password []byte, progress ProgressCallback) error {
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	stat, err := input.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}
	totalSize := stat.Size()

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	// Read header
	header := make([]byte, HeaderSizeV2)
	if _, err := io.ReadFull(input, header); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Verify magic and version
	if string(header[0:4]) != MagicV2 {
		return fmt.Errorf("invalid magic bytes")
	}
	if header[4] != FormatVersionV2 {
		return fmt.Errorf("unsupported version: %d", header[4])
	}

	flags := ParseFlags(header[5])
	if !flags.Streaming {
		return fmt.Errorf("file is not in streaming format")
	}

	salt := header[6:22]
	storedHeaderMAC := header[22:86]

	// Determine KDF mode
	kdfMode := KDFModeNormal
	if flags.Mode == ModeParanoid {
		kdfMode = KDFModeParanoid
	}

	// Derive master key
	masterKey, _, _, err := DeriveKeyV2(password, salt, kdfMode, 0)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	// Derive subkeys
	paranoid := flags.Mode == ModeParanoid
	keys, err := DeriveAllSubkeys(masterKey, salt, paranoid)
	if err != nil {
		return fmt.Errorf("subkey derivation failed: %w", err)
	}

	// Verify header MAC
	headerData := header[0:22]
	headerMAC, err := ComputeHMACSHA3(headerData, keys.HeaderHMAC[:32])
	if err != nil {
		return fmt.Errorf("header MAC computation failed: %w", err)
	}
	if !bytes.Equal(headerMAC, storedHeaderMAC) {
		return fmt.Errorf("header MAC verification failed")
	}

	// Read chunk count
	chunkCountBytes := make([]byte, 8)
	if _, err := io.ReadFull(input, chunkCountBytes); err != nil {
		return fmt.Errorf("failed to read chunk count: %w", err)
	}
	chunkCount := binary.BigEndian.Uint64(chunkCountBytes)

	// Initialize payload MAC
	var payloadMAC MAC
	if paranoid {
		payloadMAC, err = NewHMACSHA3(keys.PayloadMAC)
	} else {
		payloadMAC, err = NewBlake2bMAC(keys.PayloadMAC)
	}
	if err != nil {
		return fmt.Errorf("failed to create payload MAC: %w", err)
	}

	// Decrypt chunks
	var currentKeys = keys
	var bytesProcessed int64

	for chunkIndex := uint64(0); chunkIndex < chunkCount; chunkIndex++ {
		// Check for rekeying in paranoid mode
		if paranoid && chunkIndex > 0 && chunkIndex%RekeyInterval == 0 {
			rekeyIdx := chunkIndex / RekeyInterval
			currentKeys, err = DeriveRekeyMaterial(masterKey, salt, rekeyIdx)
			if err != nil {
				return fmt.Errorf("rekeying failed at chunk %d: %w", chunkIndex, err)
			}
		}

		// Read chunk length
		chunkLenBytes := make([]byte, 4)
		if _, err := io.ReadFull(input, chunkLenBytes); err != nil {
			return fmt.Errorf("failed to read chunk length: %w", err)
		}
		chunkLen := binary.BigEndian.Uint32(chunkLenBytes)

		// Read encrypted chunk
		encryptedChunk := make([]byte, chunkLen)
		if _, err := io.ReadFull(input, encryptedChunk); err != nil {
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		// Update payload MAC
		payloadMAC.Write(chunkLenBytes)
		payloadMAC.Write(encryptedChunk)

		// Derive nonce for this chunk
		nonce, err := DeriveNonce(masterKey, salt, chunkIndex, XChaChaNonceSize)
		if err != nil {
			return fmt.Errorf("nonce derivation failed: %w", err)
		}

		// Decrypt chunk
		var plaintext []byte
		switch flags.Cipher {
		case CipherAESGCM:
			aesNonce, _ := DeriveNonce(masterKey, salt, chunkIndex, AESNonceSize)
			plaintext, err = AESGCMDecryptWithNonce(encryptedChunk, currentKeys.Cipher, aesNonce)
		case CipherXChaCha20:
			plaintext, err = XChaCha20DecryptWithNonce(encryptedChunk, currentKeys.Cipher, nonce)
		case CipherCascade:
			serpentNonce := DeriveSerpentNonce(chunkIndex)
			plaintext, err = CascadeDecryptWithNonces(encryptedChunk, currentKeys.Cipher, currentKeys.Serpent, nonce, serpentNonce)
		default:
			return fmt.Errorf("unknown cipher type: %d", flags.Cipher)
		}
		if err != nil {
			return fmt.Errorf("chunk decryption failed: %w", err)
		}

		// Write plaintext
		if _, err := output.Write(plaintext); err != nil {
			return fmt.Errorf("failed to write plaintext: %w", err)
		}

		bytesProcessed += int64(len(encryptedChunk)) + 4

		if progress != nil {
			progress(bytesProcessed, totalSize)
		}
	}

	// Read and verify payload MAC
	macSize := GetMACSize(flags.Mode)
	storedPayloadMAC := make([]byte, macSize)
	if _, err := io.ReadFull(input, storedPayloadMAC); err != nil {
		return fmt.Errorf("failed to read payload MAC: %w", err)
	}

	computedMAC := payloadMAC.Sum(nil)
	if !bytes.Equal(computedMAC, storedPayloadMAC) {
		return fmt.Errorf("payload MAC verification failed (corrupted data)")
	}

	return nil
}

// AESGCMEncryptWithNonce encrypts with a specific nonce
func AESGCMEncryptWithNonce(plaintext, key, nonce []byte) ([]byte, error) {
	if len(key) != AESKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", AESKeySize, len(key))
	}

	if len(nonce) != AESNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", AESNonceSize, len(nonce))
	}

	block, err := newAESCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := newGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESGCMDecryptWithNonce decrypts with a specific nonce
func AESGCMDecryptWithNonce(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(key) != AESKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", AESKeySize, len(key))
	}

	if len(nonce) != AESNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", AESNonceSize, len(nonce))
	}

	block, err := newAESCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := newGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// ShouldUseStreaming determines if streaming should be used based on file size
func ShouldUseStreaming(filePath string) (bool, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return stat.Size() > StreamingThreshold, nil
}

// EncryptFileAuto automatically chooses between binary and streaming encryption
func EncryptFileAuto(inputPath, outputPath string, password []byte, opts *EncryptOptions, progress ProgressCallback) error {
	shouldStream, err := ShouldUseStreaming(inputPath)
	if err != nil {
		return err
	}

	if shouldStream {
		return StreamingEncryptFile(inputPath, outputPath, password, opts, progress)
	}

	// Binary mode
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	ciphertext, err := EncryptV2(plaintext, password, opts)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, ciphertext, 0600)
}

// DecryptFileAuto automatically detects format and decrypts
func DecryptFileAuto(inputPath, outputPath string, password []byte, progress ProgressCallback) error {
	// Read header to check format
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	header := make([]byte, HeaderSizeV2+8) // +8 for chunk count
	n, _ := input.Read(header)
	input.Close()

	if n >= HeaderSizeV2 && string(header[0:4]) == MagicV2 {
		flags := ParseFlags(header[5])
		if flags.Streaming {
			return StreamingDecryptFile(inputPath, outputPath, password, progress)
		}
	}

	// Binary mode
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	plaintext, err := DetectAndDecrypt(data, password)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, plaintext, 0600)
}
