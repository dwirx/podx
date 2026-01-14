package crypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestShamirSplitCombine(t *testing.T) {
	secret := []byte("this is a super secret key that must be protected")

	tests := []struct {
		name      string
		threshold int
		total     int
	}{
		{"2-of-3", 2, 3},
		{"3-of-5", 3, 5},
		{"4-of-7", 4, 7},
		{"5-of-5", 5, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shares, err := ShamirSplit(secret, tc.threshold, tc.total)
			if err != nil {
				t.Fatalf("ShamirSplit failed: %v", err)
			}

			if len(shares) != tc.total {
				t.Errorf("expected %d shares, got %d", tc.total, len(shares))
			}

			// Combine using exactly threshold shares
			recovered, err := ShamirCombine(shares[:tc.threshold])
			if err != nil {
				t.Fatalf("ShamirCombine failed: %v", err)
			}

			if !bytes.Equal(recovered, secret) {
				t.Errorf("recovered secret does not match original")
			}

			// Try with more shares
			if tc.total > tc.threshold {
				recovered2, err := ShamirCombine(shares)
				if err != nil {
					t.Fatalf("ShamirCombine with all shares failed: %v", err)
				}
				if !bytes.Equal(recovered2, secret) {
					t.Errorf("recovered secret (all shares) does not match original")
				}
			}
		})
	}
}

func TestShamirDifferentShareCombinations(t *testing.T) {
	secret := []byte("test secret for combination testing")

	shares, err := ShamirSplit(secret, 3, 5)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	// Try different combinations of 3 shares
	combinations := [][]int{
		{0, 1, 2},
		{0, 1, 3},
		{0, 1, 4},
		{0, 2, 3},
		{0, 2, 4},
		{0, 3, 4},
		{1, 2, 3},
		{1, 2, 4},
		{1, 3, 4},
		{2, 3, 4},
	}

	for _, combo := range combinations {
		subset := []*Share{shares[combo[0]], shares[combo[1]], shares[combo[2]]}
		recovered, err := ShamirCombine(subset)
		if err != nil {
			t.Fatalf("ShamirCombine failed for combo %v: %v", combo, err)
		}
		if !bytes.Equal(recovered, secret) {
			t.Errorf("combo %v: recovered secret does not match", combo)
		}
	}
}

func TestShamirInsufficientShares(t *testing.T) {
	secret := []byte("secret")

	shares, err := ShamirSplit(secret, 3, 5)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	// Try with only 2 shares (need 3)
	_, err = ShamirCombine(shares[:2])
	if err == nil {
		t.Error("ShamirCombine should fail with insufficient shares")
	}
}

func TestShamirInvalidParams(t *testing.T) {
	secret := []byte("secret")

	// Threshold < 2
	_, err := ShamirSplit(secret, 1, 3)
	if err == nil {
		t.Error("should fail with threshold < 2")
	}

	// Total < threshold
	_, err = ShamirSplit(secret, 3, 2)
	if err == nil {
		t.Error("should fail with total < threshold")
	}

	// Empty secret
	_, err = ShamirSplit([]byte{}, 2, 3)
	if err == nil {
		t.Error("should fail with empty secret")
	}

	// Too many shares
	_, err = ShamirSplit(secret, 2, 256)
	if err == nil {
		t.Error("should fail with > 255 shares")
	}
}

func TestShareExportImport(t *testing.T) {
	secret := []byte("export test secret")

	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	for _, share := range shares {
		// Export
		data, err := ExportShare(share)
		if err != nil {
			t.Fatalf("ExportShare failed: %v", err)
		}

		// Import
		imported, err := ImportShare(data)
		if err != nil {
			t.Fatalf("ImportShare failed: %v", err)
		}

		// Verify
		if imported.Index != share.Index {
			t.Errorf("index mismatch: expected %d, got %d", share.Index, imported.Index)
		}
		if !bytes.Equal(imported.Data, share.Data) {
			t.Error("data mismatch")
		}
		if imported.Threshold != share.Threshold {
			t.Errorf("threshold mismatch")
		}
		if imported.Total != share.Total {
			t.Errorf("total mismatch")
		}
		if imported.ID != share.ID {
			t.Errorf("ID mismatch")
		}
	}
}

func TestShareSaveLoad(t *testing.T) {
	secret := []byte("save/load test secret")

	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	// Create temp directory
	tmpDir := t.TempDir()

	// Save all shares
	paths, err := SaveAllShares(shares, tmpDir)
	if err != nil {
		t.Fatalf("SaveAllShares failed: %v", err)
	}

	if len(paths) != 3 {
		t.Errorf("expected 3 paths, got %d", len(paths))
	}

	// Load shares back
	loaded, err := LoadSharesFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadSharesFromDir failed: %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("expected 3 loaded shares, got %d", len(loaded))
	}

	// Combine and verify
	recovered, err := ShamirCombine(loaded)
	if err != nil {
		t.Fatalf("ShamirCombine failed: %v", err)
	}

	if !bytes.Equal(recovered, secret) {
		t.Error("recovered secret does not match original")
	}
}

func TestValidateShares(t *testing.T) {
	secret := []byte("validation test")

	shares, err := ShamirSplit(secret, 2, 3)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	// Valid shares
	if err := ValidateShares(shares); err != nil {
		t.Errorf("ValidateShares should succeed: %v", err)
	}

	// Empty
	if err := ValidateShares([]*Share{}); err == nil {
		t.Error("should fail for empty shares")
	}

	// Duplicate index
	dupShares := []*Share{shares[0], shares[0]}
	if err := ValidateShares(dupShares); err == nil {
		t.Error("should fail for duplicate indices")
	}

	// Different ID
	secret2 := []byte("another secret")
	shares2, _ := ShamirSplit(secret2, 2, 3)
	mixedShares := []*Share{shares[0], shares2[1]}
	if err := ValidateShares(mixedShares); err == nil {
		t.Error("should fail for mixed IDs")
	}
}

func TestGetPreset(t *testing.T) {
	preset, err := GetPreset("2-of-3")
	if err != nil {
		t.Fatalf("GetPreset failed: %v", err)
	}
	if preset.Threshold != 2 || preset.TotalShares != 3 {
		t.Error("preset values incorrect")
	}

	_, err = GetPreset("invalid")
	if err == nil {
		t.Error("should fail for invalid preset")
	}
}

func TestShareInfo(t *testing.T) {
	shares, _ := ShamirSplit([]byte("test"), 2, 3)
	info := ShareInfo(shares[0])

	if info == "" {
		t.Error("ShareInfo should return non-empty string")
	}
}

func TestShamirLargeSecret(t *testing.T) {
	// Test with larger secret (encryption key size)
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i * 7)
	}

	shares, err := ShamirSplit(secret, 3, 5)
	if err != nil {
		t.Fatalf("ShamirSplit failed: %v", err)
	}

	recovered, err := ShamirCombine(shares[:3])
	if err != nil {
		t.Fatalf("ShamirCombine failed: %v", err)
	}

	if !bytes.Equal(recovered, secret) {
		t.Error("recovered secret does not match")
	}
}

func TestImportCorruptedShare(t *testing.T) {
	secret := []byte("test")
	shares, _ := ShamirSplit(secret, 2, 3)

	data, _ := ExportShare(shares[0])

	// Corrupt the data
	data[len(data)-10] = ^data[len(data)-10]

	_, err := ImportShare(data)
	if err == nil {
		t.Error("should fail for corrupted share")
	}
}

func BenchmarkShamirSplit(b *testing.B) {
	secret := make([]byte, 32) // 256-bit key

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ShamirSplit(secret, 3, 5)
	}
}

func BenchmarkShamirCombine(b *testing.B) {
	secret := make([]byte, 32)
	shares, _ := ShamirSplit(secret, 3, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ShamirCombine(shares[:3])
	}
}

func TestSaveShareCreatesFile(t *testing.T) {
	secret := []byte("test secret")
	shares, _ := ShamirSplit(secret, 2, 3)

	tmpDir := t.TempDir()
	path, err := SaveShare(shares[0], tmpDir)
	if err != nil {
		t.Fatalf("SaveShare failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("share file was not created")
	}

	// Verify filename format
	expectedPrefix := "share-1-of-3_"
	if filepath.Base(path)[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("unexpected filename: %s", filepath.Base(path))
	}
}
