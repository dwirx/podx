package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/vault/shamir"
)

// ShamirPreset represents a preset threshold configuration
type ShamirPreset struct {
	Name        string
	Threshold   int // Minimum shares needed (N)
	TotalShares int // Total shares to generate (M)
	Description string
}

// Available presets
var ShamirPresets = []ShamirPreset{
	{Name: "2-of-3", Threshold: 2, TotalShares: 3, Description: "2 of 3 shares required (small team)"},
	{Name: "3-of-5", Threshold: 3, TotalShares: 5, Description: "3 of 5 shares required (medium team)"},
	{Name: "4-of-7", Threshold: 4, TotalShares: 7, Description: "4 of 7 shares required (large team)"},
	{Name: "5-of-9", Threshold: 5, TotalShares: 9, Description: "5 of 9 shares required (enterprise)"},
}

// Share represents a single share of a split secret
type Share struct {
	Index     byte   `json:"index"`     // Share index (1-based)
	Data      []byte `json:"data"`      // Share data
	Threshold int    `json:"threshold"` // Minimum shares needed to recover
	Total     int    `json:"total"`     // Total shares created
	ID        string `json:"id"`        // Unique ID for this split operation
	Created   int64  `json:"created"`   // Unix timestamp
}

// ShareExport is the format for exporting shares
type ShareExport struct {
	Version  int    `json:"version"`
	Share    *Share `json:"share"`
	Checksum string `json:"checksum"` // Base64 encoded BLAKE2b hash of share data
}

// ShamirSplit splits a secret into n shares with threshold k
// At least k shares are needed to recover the secret
func ShamirSplit(secret []byte, threshold, total int) ([]*Share, error) {
	if threshold < 2 {
		return nil, fmt.Errorf("threshold must be at least 2")
	}
	if total < threshold {
		return nil, fmt.Errorf("total shares must be >= threshold")
	}
	if total > 255 {
		return nil, fmt.Errorf("maximum 255 shares supported")
	}
	if len(secret) == 0 {
		return nil, fmt.Errorf("secret cannot be empty")
	}

	// Generate unique ID for this split
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}
	id := base64.RawURLEncoding.EncodeToString(idBytes)

	// Use hashicorp/vault's shamir implementation
	rawShares, err := shamir.Split(secret, total, threshold)
	if err != nil {
		return nil, fmt.Errorf("shamir split failed: %w", err)
	}

	// Convert to our Share format
	shares := make([]*Share, total)
	for i, rawShare := range rawShares {
		shares[i] = &Share{
			Index:     byte(i + 1), // 1-based indexing
			Data:      rawShare,
			Threshold: threshold,
			Total:     total,
			ID:        id,
			Created:   time.Now().Unix(),
		}
	}

	return shares, nil
}

// ShamirCombine recovers the secret from shares
func ShamirCombine(shares []*Share) ([]byte, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("no shares provided")
	}

	// Check all shares are from same split
	id := shares[0].ID
	threshold := shares[0].Threshold

	for i, s := range shares {
		if s.ID != id {
			return nil, fmt.Errorf("share %d has different ID", i)
		}
	}

	if len(shares) < threshold {
		return nil, fmt.Errorf("need at least %d shares, got %d", threshold, len(shares))
	}

	// Convert to raw format for hashicorp/vault's shamir
	rawShares := make([][]byte, len(shares))
	for i, s := range shares {
		rawShares[i] = s.Data
	}

	// Combine using hashicorp/vault's implementation
	secret, err := shamir.Combine(rawShares)
	if err != nil {
		return nil, fmt.Errorf("shamir combine failed: %w", err)
	}

	return secret, nil
}

// ExportShare exports a share to JSON format
func ExportShare(share *Share) ([]byte, error) {
	// Compute checksum using BLAKE2b with 32-byte key
	checksumKey := make([]byte, 32)
	copy(checksumKey, []byte("shamir-share-checksum-key-podx!!"))
	checksum, err := ComputeBlake2bMAC(share.Data, checksumKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute checksum: %w", err)
	}

	export := &ShareExport{
		Version:  1,
		Share:    share,
		Checksum: base64.StdEncoding.EncodeToString(checksum),
	}

	return json.MarshalIndent(export, "", "  ")
}

// ImportShare imports a share from JSON format
func ImportShare(data []byte) (*Share, error) {
	var export ShareExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("invalid share format: %w", err)
	}

	if export.Version != 1 {
		return nil, fmt.Errorf("unsupported share version: %d", export.Version)
	}

	// Verify checksum
	expectedChecksum, err := base64.StdEncoding.DecodeString(export.Checksum)
	if err != nil {
		return nil, fmt.Errorf("invalid checksum encoding: %w", err)
	}

	checksumKey := make([]byte, 32)
	copy(checksumKey, []byte("shamir-share-checksum-key-podx!!"))
	valid, err := VerifyBlake2bMAC(export.Share.Data, checksumKey, expectedChecksum)
	if err != nil {
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("share checksum mismatch - data may be corrupted")
	}

	return export.Share, nil
}

// SaveShare saves a share to a file
func SaveShare(share *Share, dir string) (string, error) {
	data, err := ExportShare(share)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("share-%d-of-%d_%s.json", share.Index, share.Total, share.ID)
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write share: %w", err)
	}

	return path, nil
}

// LoadShare loads a share from a file
func LoadShare(path string) (*Share, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read share: %w", err)
	}

	return ImportShare(data)
}

// SaveAllShares saves all shares to a directory
func SaveAllShares(shares []*Share, dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	paths := make([]string, len(shares))
	for i, share := range shares {
		path, err := SaveShare(share, dir)
		if err != nil {
			return nil, err
		}
		paths[i] = path
	}

	return paths, nil
}

// LoadSharesFromDir loads all shares from a directory
func LoadSharesFromDir(dir string) ([]*Share, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var shares []*Share
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "share-") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		share, err := LoadShare(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", entry.Name(), err)
		}
		shares = append(shares, share)
	}

	return shares, nil
}

// GetPreset returns a preset by name
func GetPreset(name string) (*ShamirPreset, error) {
	for _, p := range ShamirPresets {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("unknown preset: %s", name)
}

// ValidateShares checks if shares can be combined
func ValidateShares(shares []*Share) error {
	if len(shares) == 0 {
		return fmt.Errorf("no shares provided")
	}

	id := shares[0].ID
	threshold := shares[0].Threshold

	indices := make(map[byte]bool)

	for i, s := range shares {
		if s.ID != id {
			return fmt.Errorf("share %d has different ID (from different split)", i+1)
		}
		if indices[s.Index] {
			return fmt.Errorf("duplicate share index: %d", s.Index)
		}
		indices[s.Index] = true
	}

	if len(shares) < threshold {
		return fmt.Errorf("insufficient shares: need %d, have %d", threshold, len(shares))
	}

	return nil
}

// ShareInfo returns human-readable information about a share
func ShareInfo(share *Share) string {
	return fmt.Sprintf("Share %d of %d (threshold: %d, ID: %s, created: %s)",
		share.Index,
		share.Total,
		share.Threshold,
		share.ID,
		time.Unix(share.Created, 0).Format("2006-01-02 15:04:05"))
}
