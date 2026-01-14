package crypto

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateShareQRCode(t *testing.T) {
	secret := []byte("test secret for QR code")
	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	pngData, err := GenerateShareQRCode(shares[0], nil)
	if err != nil {
		t.Fatalf("GenerateShareQRCode failed: %v", err)
	}

	// Verify PNG header
	if len(pngData) < 8 {
		t.Fatal("PNG data too short")
	}

	// PNG magic bytes
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngMagic {
		if pngData[i] != b {
			t.Errorf("Invalid PNG header at byte %d: expected %x, got %x", i, b, pngData[i])
		}
	}
}

func TestSaveShareQRCode(t *testing.T) {
	secret := []byte("test secret")
	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	tmpDir := t.TempDir()

	path, err := SaveShareQRCode(shares[0], tmpDir, nil)
	if err != nil {
		t.Fatalf("SaveShareQRCode failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("QR code file was not created")
	}

	// Verify filename
	if !strings.HasSuffix(path, ".png") {
		t.Error("QR code file should have .png extension")
	}
}

func TestSaveAllShareQRCodes(t *testing.T) {
	secret := []byte("test secret")
	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	tmpDir := t.TempDir()

	paths, err := SaveAllShareQRCodes(shares, tmpDir, nil)
	if err != nil {
		t.Fatalf("SaveAllShareQRCodes failed: %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("expected 3 paths, got %d", len(paths))
	}

	// Verify all files exist
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("QR code file not found: %s", path)
		}
	}
}

func TestParseShareFromQRData(t *testing.T) {
	secret := []byte("test secret for parsing")
	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	// Export and create QR data format
	data, err := ExportShare(shares[0])
	if err != nil {
		t.Fatalf("ExportShare failed: %v", err)
	}

	qrData := "PODX-SHARE:" + string(mustBase64Encode(data))

	// Parse back
	parsed, err := ParseShareFromQRData(qrData)
	if err != nil {
		t.Fatalf("ParseShareFromQRData failed: %v", err)
	}

	if parsed.Index != shares[0].Index {
		t.Errorf("index mismatch: expected %d, got %d", shares[0].Index, parsed.Index)
	}
	if parsed.ID != shares[0].ID {
		t.Errorf("ID mismatch")
	}
}

func TestParseShareFromQRDataInvalid(t *testing.T) {
	// Missing prefix
	_, err := ParseShareFromQRData("invalid data")
	if err == nil {
		t.Error("should fail for missing prefix")
	}

	// Invalid base64
	_, err = ParseShareFromQRData("PODX-SHARE:!!!invalid!!!")
	if err == nil {
		t.Error("should fail for invalid base64")
	}

	// Invalid JSON
	_, err = ParseShareFromQRData("PODX-SHARE:" + string(mustBase64Encode([]byte("not json"))))
	if err == nil {
		t.Error("should fail for invalid JSON")
	}
}

func TestGenerateShareQRCodeASCII(t *testing.T) {
	secret := []byte("test")
	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	ascii, err := GenerateShareQRCodeASCII(shares[0])
	if err != nil {
		t.Fatalf("GenerateShareQRCodeASCII failed: %v", err)
	}

	if ascii == "" {
		t.Error("ASCII QR code should not be empty")
	}

	// Should contain block characters
	if !strings.Contains(ascii, "█") && !strings.Contains(ascii, "▀") && !strings.Contains(ascii, "▄") {
		t.Log("ASCII QR code may not use expected block characters")
	}
}

func TestQRCodeOptions(t *testing.T) {
	secret := []byte("test")
	shares, _ := ShamirSplit(secret, 2, 3)

	opts := &QRCodeOptions{
		Size:          512,
		RecoveryLevel: 3, // Highest
		DisableBorder: true,
	}

	pngData, err := GenerateShareQRCode(shares[0], opts)
	if err != nil {
		t.Fatalf("GenerateShareQRCode with options failed: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("PNG data should not be empty")
	}
}

func TestCanFitInQRCode(t *testing.T) {
	// Small secret should fit
	smallSecret := []byte("small")
	smallShares, _ := ShamirSplit(smallSecret, 2, 3)

	if !CanFitInQRCode(smallShares[0], 2) {
		t.Error("small share should fit in QR code")
	}

	// Very large secret may not fit
	largeSecret := make([]byte, 2000) // 2KB secret
	for i := range largeSecret {
		largeSecret[i] = byte(i % 256)
	}
	largeShares, _ := ShamirSplit(largeSecret, 2, 3)

	// With highest error correction, may not fit
	if CanFitInQRCode(largeShares[0], 3) {
		t.Log("large share fits in QR code with level 3 - this is unexpected but not necessarily wrong")
	}
}

func TestQRCodeCapacity(t *testing.T) {
	levels := []int{0, 1, 2, 3}
	prevCapacity := 5000 // Higher than any actual capacity

	for _, level := range levels {
		capacity := QRCodeCapacity(level)
		if capacity <= 0 {
			t.Errorf("capacity for level %d should be positive", level)
		}
		if capacity >= prevCapacity {
			t.Errorf("capacity should decrease as error correction increases: level %d has %d >= %d",
				level, capacity, prevCapacity)
		}
		prevCapacity = capacity
	}
}

func TestGenerateCombinedQRSheet(t *testing.T) {
	secret := []byte("test secret")
	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	pngData, err := GenerateCombinedQRSheet(shares, nil)
	if err != nil {
		t.Fatalf("GenerateCombinedQRSheet failed: %v", err)
	}

	// Verify PNG header
	if len(pngData) < 8 {
		t.Fatal("PNG data too short")
	}

	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47}
	for i, b := range pngMagic {
		if pngData[i] != b {
			t.Errorf("Invalid PNG header")
		}
	}
}

func TestGenerateCombinedQRSheetEmpty(t *testing.T) {
	_, err := GenerateCombinedQRSheet([]*Share{}, nil)
	if err == nil {
		t.Error("should fail for empty shares")
	}
}

func TestGeneratePrintableSheet(t *testing.T) {
	secret := []byte("test")
	shares, _ := ShamirSplit(secret, 2, 3)

	pngData, err := GeneratePrintableSheet(shares[0], nil)
	if err != nil {
		t.Fatalf("GeneratePrintableSheet failed: %v", err)
	}

	if len(pngData) == 0 {
		t.Error("printable sheet should not be empty")
	}
}

// Helper function
func mustBase64Encode(data []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(data))
}

func TestQRCodeWithDifferentSecretSizes(t *testing.T) {
	sizes := []int{16, 32, 64, 128, 256}

	for _, size := range sizes {
		t.Run(filepath.Base(t.Name()), func(t *testing.T) {
			secret := make([]byte, size)
			for i := range secret {
				secret[i] = byte(i)
			}

			shares, err := ShamirSplit(secret, 2, 3)
			if err != nil {
				t.Fatalf("ShamirSplit failed for size %d: %v", size, err)
			}

			pngData, err := GenerateShareQRCode(shares[0], nil)
			if err != nil {
				t.Fatalf("GenerateShareQRCode failed for size %d: %v", size, err)
			}

			if len(pngData) == 0 {
				t.Errorf("PNG data empty for size %d", size)
			}
		})
	}
}
