package crypto

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
)

// QRCodeOptions configures QR code generation
type QRCodeOptions struct {
	Size            int  // Image size in pixels (default: 256)
	RecoveryLevel   int  // Error correction level: 0=L, 1=M, 2=Q, 3=H (default: 2)
	DisableBorder   bool // Remove quiet zone border
}

// DefaultQRCodeOptions returns default QR code options
func DefaultQRCodeOptions() *QRCodeOptions {
	return &QRCodeOptions{
		Size:          256,
		RecoveryLevel: 2, // Q level - 25% recovery
	}
}

// GenerateShareQRCode generates a QR code image for a share
func GenerateShareQRCode(share *Share, opts *QRCodeOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultQRCodeOptions()
	}

	// Export share to JSON
	data, err := ExportShare(share)
	if err != nil {
		return nil, fmt.Errorf("failed to export share: %w", err)
	}

	// Encode as base64 for QR (more efficient than raw JSON)
	encoded := base64.StdEncoding.EncodeToString(data)

	// Add prefix for identification
	qrData := "PODX-SHARE:" + encoded

	// Check size limit (QR codes have max capacity ~4296 alphanumeric chars)
	if len(qrData) > 4000 {
		return nil, fmt.Errorf("share data too large for QR code (%d bytes)", len(qrData))
	}

	// Convert recovery level
	var level qrcode.RecoveryLevel
	switch opts.RecoveryLevel {
	case 0:
		level = qrcode.Low
	case 1:
		level = qrcode.Medium
	case 2:
		level = qrcode.High
	case 3:
		level = qrcode.Highest
	default:
		level = qrcode.High
	}

	// Generate QR code
	qr, err := qrcode.New(qrData, level)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.DisableBorder = opts.DisableBorder

	// Generate PNG
	pngData, err := qr.PNG(opts.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PNG: %w", err)
	}

	return pngData, nil
}

// SaveShareQRCode saves a share as a QR code image
func SaveShareQRCode(share *Share, dir string, opts *QRCodeOptions) (string, error) {
	pngData, err := GenerateShareQRCode(share, opts)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("share-%d-of-%d_%s.png", share.Index, share.Total, share.ID)
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, pngData, 0600); err != nil {
		return "", fmt.Errorf("failed to write QR code: %w", err)
	}

	return path, nil
}

// SaveAllShareQRCodes saves all shares as QR code images
func SaveAllShareQRCodes(shares []*Share, dir string, opts *QRCodeOptions) ([]string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	paths := make([]string, len(shares))
	for i, share := range shares {
		path, err := SaveShareQRCode(share, dir, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to save QR code for share %d: %w", i+1, err)
		}
		paths[i] = path
	}

	return paths, nil
}

// ParseShareFromQRData parses a share from QR code data string
func ParseShareFromQRData(qrData string) (*Share, error) {
	// Check prefix
	prefix := "PODX-SHARE:"
	if len(qrData) < len(prefix) || qrData[:len(prefix)] != prefix {
		return nil, fmt.Errorf("invalid QR code format: missing PODX-SHARE prefix")
	}

	// Decode base64
	encoded := qrData[len(prefix):]
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}

	// Import share
	return ImportShare(data)
}

// GenerateShareQRCodeASCII generates an ASCII art QR code for terminal display
func GenerateShareQRCodeASCII(share *Share) (string, error) {
	// Export share to JSON
	data, err := ExportShare(share)
	if err != nil {
		return "", fmt.Errorf("failed to export share: %w", err)
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(data)
	qrData := "PODX-SHARE:" + encoded

	// Check size limit
	if len(qrData) > 4000 {
		return "", fmt.Errorf("share data too large for QR code")
	}

	// Generate QR code
	qr, err := qrcode.New(qrData, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	return qr.ToSmallString(false), nil
}

// GeneratePrintableSheet generates a printable sheet with share info and QR code
func GeneratePrintableSheet(share *Share, opts *QRCodeOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultQRCodeOptions()
		opts.Size = 400 // Larger for printing
	}

	// Generate QR code
	qrPNG, err := GenerateShareQRCode(share, opts)
	if err != nil {
		return nil, err
	}

	// For now, just return the QR code PNG
	// In a full implementation, this would combine text info with the QR code
	// using an image library to create a complete printable sheet

	return qrPNG, nil
}

// GenerateCombinedQRSheet creates a single image with multiple QR codes
func GenerateCombinedQRSheet(shares []*Share, opts *QRCodeOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultQRCodeOptions()
	}

	if len(shares) == 0 {
		return nil, fmt.Errorf("no shares provided")
	}

	// Calculate grid layout
	cols := 2
	if len(shares) <= 2 {
		cols = len(shares)
	}
	rows := (len(shares) + cols - 1) / cols

	qrSize := opts.Size
	padding := 20 // Padding between QR codes
	labelHeight := 30 // Height for label text

	// Create combined image
	width := cols*qrSize + (cols+1)*padding
	height := rows*(qrSize+labelHeight) + (rows+1)*padding

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Generate and place each QR code
	for i, share := range shares {
		row := i / cols
		col := i % cols

		// Generate individual QR code
		qr, err := qrcode.New(fmt.Sprintf("PODX-SHARE:%s",
			base64.StdEncoding.EncodeToString(mustExportShare(share))), qrcode.High)
		if err != nil {
			return nil, fmt.Errorf("failed to create QR for share %d: %w", i+1, err)
		}

		qrImg := qr.Image(qrSize)

		// Calculate position
		x := col*qrSize + (col+1)*padding
		y := row*(qrSize+labelHeight) + (row+1)*padding + labelHeight

		// Draw QR code onto combined image
		for dy := 0; dy < qrSize; dy++ {
			for dx := 0; dx < qrSize; dx++ {
				if dx < qrImg.Bounds().Dx() && dy < qrImg.Bounds().Dy() {
					img.Set(x+dx, y+dy, qrImg.At(dx, dy))
				}
			}
		}

		// Draw label (simplified - just a line of dark pixels as placeholder)
		labelY := row*(qrSize+labelHeight) + (row+1)*padding
		for lx := x; lx < x+qrSize; lx++ {
			img.Set(lx, labelY+5, color.Black)
		}
	}

	// Encode to PNG
	var buf []byte
	f, err := os.CreateTemp("", "qr-sheet-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	f.Seek(0, 0)
	buf, err = os.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// mustExportShare exports a share and panics on error (for internal use)
func mustExportShare(share *Share) []byte {
	data, err := ExportShare(share)
	if err != nil {
		panic(err)
	}
	return data
}

// QRCodeCapacity returns the maximum data size for a QR code at given error correction level
func QRCodeCapacity(level int) int {
	// Approximate capacities for alphanumeric mode
	switch level {
	case 0: // L - 7% recovery
		return 4296
	case 1: // M - 15% recovery
		return 3391
	case 2: // Q - 25% recovery
		return 2420
	case 3: // H - 30% recovery
		return 1852
	default:
		return 2420
	}
}

// CanFitInQRCode checks if a share can fit in a QR code
func CanFitInQRCode(share *Share, level int) bool {
	data, err := ExportShare(share)
	if err != nil {
		return false
	}

	// Base64 encoding increases size by ~33%
	encoded := base64.StdEncoding.EncodeToString(data)
	totalSize := len("PODX-SHARE:") + len(encoded)

	return totalSize <= QRCodeCapacity(level)
}
