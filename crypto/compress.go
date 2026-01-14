package crypto

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// Compression levels
const (
	CompressionNone    CompressionLevel = 0
	CompressionFast    CompressionLevel = 1
	CompressionDefault CompressionLevel = 3
	CompressionBetter  CompressionLevel = 7
	CompressionBest    CompressionLevel = 11
)

// CompressionLevel represents zstd compression level
type CompressionLevel int

// CompressionHeader is prepended to compressed data
const (
	CompressionMagic   = "ZSTD"
	CompressionVersion = byte(1)
)

// CompressedHeader represents the compression header
type CompressedHeader struct {
	Magic          [4]byte // "ZSTD"
	Version        byte    // 1
	Level          byte    // compression level used
	OriginalSize   uint64  // original uncompressed size
}

// Compress compresses data using zstd with the specified level
func Compress(data []byte, level CompressionLevel) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	// Convert level to zstd encoder level
	var encoderLevel zstd.EncoderLevel
	switch {
	case level <= CompressionFast:
		encoderLevel = zstd.SpeedFastest
	case level <= CompressionDefault:
		encoderLevel = zstd.SpeedDefault
	case level <= CompressionBetter:
		encoderLevel = zstd.SpeedBetterCompression
	default:
		encoderLevel = zstd.SpeedBestCompression
	}

	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(encoderLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	compressed := encoder.EncodeAll(data, nil)

	// Build output with header
	// Header: magic(4) + version(1) + level(1) + originalSize(8) = 14 bytes
	output := make([]byte, 14+len(compressed))
	copy(output[0:4], CompressionMagic)
	output[4] = CompressionVersion
	output[5] = byte(level)

	// Store original size (big endian)
	originalSize := uint64(len(data))
	output[6] = byte(originalSize >> 56)
	output[7] = byte(originalSize >> 48)
	output[8] = byte(originalSize >> 40)
	output[9] = byte(originalSize >> 32)
	output[10] = byte(originalSize >> 24)
	output[11] = byte(originalSize >> 16)
	output[12] = byte(originalSize >> 8)
	output[13] = byte(originalSize)

	copy(output[14:], compressed)

	return output, nil
}

// Decompress decompresses zstd-compressed data
func Decompress(data []byte) ([]byte, error) {
	if len(data) < 14 {
		return nil, fmt.Errorf("data too short for compression header")
	}

	// Verify magic
	if string(data[0:4]) != CompressionMagic {
		return nil, fmt.Errorf("invalid compression magic")
	}

	// Check version
	if data[4] != CompressionVersion {
		return nil, fmt.Errorf("unsupported compression version: %d", data[4])
	}

	// Read original size
	originalSize := uint64(data[6])<<56 |
		uint64(data[7])<<48 |
		uint64(data[8])<<40 |
		uint64(data[9])<<32 |
		uint64(data[10])<<24 |
		uint64(data[11])<<16 |
		uint64(data[12])<<8 |
		uint64(data[13])

	// Decompress
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	decompressed, err := decoder.DecodeAll(data[14:], nil)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	// Verify size
	if uint64(len(decompressed)) != originalSize {
		return nil, fmt.Errorf("size mismatch: expected %d, got %d", originalSize, len(decompressed))
	}

	return decompressed, nil
}

// IsCompressed checks if data has compression header
func IsCompressed(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return string(data[0:4]) == CompressionMagic
}

// CompressReader wraps a reader with zstd compression
type CompressReader struct {
	reader  io.Reader
	encoder *zstd.Encoder
	buf     bytes.Buffer
}

// NewCompressReader creates a compressing reader
func NewCompressReader(r io.Reader, level CompressionLevel) (*CompressReader, error) {
	var encoderLevel zstd.EncoderLevel
	switch {
	case level <= CompressionFast:
		encoderLevel = zstd.SpeedFastest
	case level <= CompressionDefault:
		encoderLevel = zstd.SpeedDefault
	case level <= CompressionBetter:
		encoderLevel = zstd.SpeedBetterCompression
	default:
		encoderLevel = zstd.SpeedBestCompression
	}

	cr := &CompressReader{reader: r}
	encoder, err := zstd.NewWriter(&cr.buf, zstd.WithEncoderLevel(encoderLevel))
	if err != nil {
		return nil, err
	}
	cr.encoder = encoder
	return cr, nil
}

// CompressWriter wraps a writer with zstd compression
type CompressWriter struct {
	writer  io.Writer
	encoder *zstd.Encoder
}

// NewCompressWriter creates a compressing writer
func NewCompressWriter(w io.Writer, level CompressionLevel) (*CompressWriter, error) {
	var encoderLevel zstd.EncoderLevel
	switch {
	case level <= CompressionFast:
		encoderLevel = zstd.SpeedFastest
	case level <= CompressionDefault:
		encoderLevel = zstd.SpeedDefault
	case level <= CompressionBetter:
		encoderLevel = zstd.SpeedBetterCompression
	default:
		encoderLevel = zstd.SpeedBestCompression
	}

	encoder, err := zstd.NewWriter(w, zstd.WithEncoderLevel(encoderLevel))
	if err != nil {
		return nil, err
	}

	return &CompressWriter{writer: w, encoder: encoder}, nil
}

// Write compresses and writes data
func (cw *CompressWriter) Write(p []byte) (int, error) {
	return cw.encoder.Write(p)
}

// Close flushes and closes the encoder
func (cw *CompressWriter) Close() error {
	return cw.encoder.Close()
}

// DecompressReader wraps a reader with zstd decompression
type DecompressReader struct {
	decoder *zstd.Decoder
}

// NewDecompressReader creates a decompressing reader
func NewDecompressReader(r io.Reader) (*DecompressReader, error) {
	decoder, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &DecompressReader{decoder: decoder}, nil
}

// Read decompresses and reads data
func (dr *DecompressReader) Read(p []byte) (int, error) {
	return dr.decoder.Read(p)
}

// Close closes the decoder
func (dr *DecompressReader) Close() error {
	dr.decoder.Close()
	return nil
}

// CompressionStats holds compression statistics
type CompressionStats struct {
	OriginalSize   int64
	CompressedSize int64
	Ratio          float64
	Level          CompressionLevel
}

// GetCompressionStats returns statistics about compressed data
func GetCompressionStats(original, compressed []byte) *CompressionStats {
	stats := &CompressionStats{
		OriginalSize:   int64(len(original)),
		CompressedSize: int64(len(compressed)),
	}

	if stats.OriginalSize > 0 {
		stats.Ratio = float64(stats.CompressedSize) / float64(stats.OriginalSize)
	}

	// Extract level from header if present
	if len(compressed) >= 6 && string(compressed[0:4]) == CompressionMagic {
		stats.Level = CompressionLevel(compressed[5])
	}

	return stats
}

// ShouldCompress determines if data should be compressed based on type/size
// Returns false for already compressed formats (images, videos, archives)
func ShouldCompress(data []byte) bool {
	if len(data) < 100 {
		// Too small to benefit from compression
		return false
	}

	// Check for already compressed formats by magic bytes
	if len(data) >= 4 {
		// PNG
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return false
		}
		// JPEG
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return false
		}
		// GIF
		if string(data[0:3]) == "GIF" {
			return false
		}
		// ZIP/JAR
		if data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
			return false
		}
		// GZIP
		if data[0] == 0x1F && data[1] == 0x8B {
			return false
		}
		// 7z
		if data[0] == 0x37 && data[1] == 0x7A && data[2] == 0xBC && data[3] == 0xAF {
			return false
		}
		// RAR
		if string(data[0:4]) == "Rar!" {
			return false
		}
		// ZSTD (already compressed)
		if string(data[0:4]) == CompressionMagic {
			return false
		}
		// MP4/MOV
		if len(data) >= 8 && string(data[4:8]) == "ftyp" {
			return false
		}
		// WebP
		if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
			return false
		}
	}

	return true
}

// CompressIfBeneficial compresses data only if it results in smaller output
func CompressIfBeneficial(data []byte, level CompressionLevel) ([]byte, bool, error) {
	if !ShouldCompress(data) {
		return data, false, nil
	}

	compressed, err := Compress(data, level)
	if err != nil {
		return nil, false, err
	}

	// Only use compressed if it's smaller
	if len(compressed) < len(data) {
		return compressed, true, nil
	}

	return data, false, nil
}
