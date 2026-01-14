package crypto

import (
	"bytes"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	// Test data with good compressibility
	original := []byte("Hello, this is a test message that should compress well. " +
		"Hello, this is a test message that should compress well. " +
		"Hello, this is a test message that should compress well. " +
		"Hello, this is a test message that should compress well. ")

	compressed, err := Compress(original, CompressionDefault)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	// Should be smaller
	if len(compressed) >= len(original) {
		t.Logf("Warning: compressed size %d >= original size %d", len(compressed), len(original))
	}

	// Verify header
	if !IsCompressed(compressed) {
		t.Error("compressed data should have compression header")
	}

	// Decompress
	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(decompressed, original) {
		t.Error("decompressed data does not match original")
	}
}

func TestCompressLevels(t *testing.T) {
	// Create compressible data
	original := bytes.Repeat([]byte("test data for compression level testing "), 100)

	levels := []CompressionLevel{
		CompressionFast,
		CompressionDefault,
		CompressionBetter,
		CompressionBest,
	}

	for _, level := range levels {
		compressed, err := Compress(original, level)
		if err != nil {
			t.Errorf("Compress with level %d failed: %v", level, err)
			continue
		}

		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Errorf("Decompress with level %d failed: %v", level, err)
			continue
		}

		if !bytes.Equal(decompressed, original) {
			t.Errorf("level %d: decompressed does not match original", level)
		}
	}
}

func TestCompressEmpty(t *testing.T) {
	empty := []byte{}
	result, err := Compress(empty, CompressionDefault)
	if err != nil {
		t.Fatalf("Compress empty failed: %v", err)
	}

	if len(result) != 0 {
		t.Error("compressing empty should return empty")
	}
}

func TestIsCompressed(t *testing.T) {
	// Not compressed
	if IsCompressed([]byte("hello")) {
		t.Error("regular data should not be detected as compressed")
	}

	// Compressed
	compressed, _ := Compress([]byte("test data to compress here!!"), CompressionDefault)
	if !IsCompressed(compressed) {
		t.Error("compressed data should be detected as compressed")
	}

	// Too short
	if IsCompressed([]byte("hi")) {
		t.Error("short data should not be detected as compressed")
	}
}

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "short data",
			data:     []byte("hi"),
			expected: false,
		},
		{
			name:     "normal text",
			data:     bytes.Repeat([]byte("hello world "), 20),
			expected: true,
		},
		{
			name:     "PNG header",
			data:     append([]byte{0x89, 0x50, 0x4E, 0x47}, bytes.Repeat([]byte{0}, 100)...),
			expected: false,
		},
		{
			name:     "JPEG header",
			data:     append([]byte{0xFF, 0xD8, 0xFF}, bytes.Repeat([]byte{0}, 100)...),
			expected: false,
		},
		{
			name:     "ZIP header",
			data:     append([]byte{0x50, 0x4B, 0x03, 0x04}, bytes.Repeat([]byte{0}, 100)...),
			expected: false,
		},
		{
			name:     "GZIP header",
			data:     append([]byte{0x1F, 0x8B}, bytes.Repeat([]byte{0}, 100)...),
			expected: false,
		},
		{
			name:     "already ZSTD",
			data:     append([]byte("ZSTD"), bytes.Repeat([]byte{0}, 100)...),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldCompress(tc.data)
			if result != tc.expected {
				t.Errorf("ShouldCompress() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestCompressIfBeneficial(t *testing.T) {
	// Compressible data
	compressible := bytes.Repeat([]byte("aaaaaaaaaa"), 100)
	result, wasCompressed, err := CompressIfBeneficial(compressible, CompressionDefault)
	if err != nil {
		t.Fatalf("CompressIfBeneficial failed: %v", err)
	}
	if !wasCompressed {
		t.Error("highly compressible data should be compressed")
	}
	if len(result) >= len(compressible) {
		t.Error("compressed size should be smaller")
	}

	// Already compressed (JPEG-like)
	jpeg := append([]byte{0xFF, 0xD8, 0xFF}, bytes.Repeat([]byte{0x42}, 100)...)
	result2, wasCompressed2, err := CompressIfBeneficial(jpeg, CompressionDefault)
	if err != nil {
		t.Fatalf("CompressIfBeneficial for JPEG failed: %v", err)
	}
	if wasCompressed2 {
		t.Error("JPEG should not be compressed")
	}
	if !bytes.Equal(result2, jpeg) {
		t.Error("non-compressed data should be returned unchanged")
	}
}

func TestGetCompressionStats(t *testing.T) {
	original := bytes.Repeat([]byte("test data "), 100)
	compressed, _ := Compress(original, CompressionBetter)

	stats := GetCompressionStats(original, compressed)

	if stats.OriginalSize != int64(len(original)) {
		t.Errorf("OriginalSize: expected %d, got %d", len(original), stats.OriginalSize)
	}
	if stats.CompressedSize != int64(len(compressed)) {
		t.Errorf("CompressedSize: expected %d, got %d", len(compressed), stats.CompressedSize)
	}
	if stats.Ratio <= 0 || stats.Ratio >= 1 {
		t.Errorf("Ratio should be between 0 and 1 for compressible data, got %f", stats.Ratio)
	}
	if stats.Level != CompressionBetter {
		t.Errorf("Level: expected %d, got %d", CompressionBetter, stats.Level)
	}
}

func TestCompressWriter(t *testing.T) {
	var buf bytes.Buffer

	writer, err := NewCompressWriter(&buf, CompressionDefault)
	if err != nil {
		t.Fatalf("NewCompressWriter failed: %v", err)
	}

	data := []byte("test data for streaming compression")
	_, err = writer.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify we got compressed output
	if buf.Len() == 0 {
		t.Error("buffer should not be empty after compression")
	}
}

func TestDecompressReader(t *testing.T) {
	// Create compressed data
	original := []byte("test data for streaming decompression testing here")
	compressed, _ := Compress(original, CompressionDefault)

	// Skip our header and read just the zstd data
	reader, err := NewDecompressReader(bytes.NewReader(compressed[14:]))
	if err != nil {
		t.Fatalf("NewDecompressReader failed: %v", err)
	}
	defer reader.Close()

	var result bytes.Buffer
	_, err = result.ReadFrom(reader)
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}

	if !bytes.Equal(result.Bytes(), original) {
		t.Error("decompressed data does not match original")
	}
}

func BenchmarkCompress(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark test data for compression "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Compress(data, CompressionDefault)
	}
}

func BenchmarkDecompress(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark test data for decompression "), 1000)
	compressed, _ := Compress(data, CompressionDefault)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decompress(compressed)
	}
}
