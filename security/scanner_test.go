package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file with secrets
	testFile := filepath.Join(tmpDir, "config.js")
	content := `const config = {
  host: "localhost",
  password: "secret123",
  port: 8080
};`
	os.WriteFile(testFile, []byte(content), 0644)

	matches, err := ScanFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 1 {
		t.Errorf("ScanFile() found %d matches, want 1", len(matches))
	}

	if matches[0].Line != 3 {
		t.Errorf("match on line %d, want 3", matches[0].Line)
	}
}

func TestScanFileBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create binary file
	binFile := filepath.Join(tmpDir, "binary.dat")
	os.WriteFile(binFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	matches, err := ScanFile(binFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 0 {
		t.Error("binary files should be skipped")
	}
}

func TestShouldSkipPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"node_modules/pkg/index.js", true},
		{".git/objects/abc", true},
		{"vendor/lib/file.go", true},
		{".env.podx", true},
		{"secrets.podx", true},
		{"src/main.go", false},
		{"config/database.yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := ShouldSkipPath(tt.path); got != tt.want {
				t.Errorf("ShouldSkipPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestScanDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with a secret
	os.WriteFile(filepath.Join(tmpDir, "config.js"), []byte(`password = "secret123"`), 0644)

	// Create a safe file
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("Hello world"), 0644)

	// Create a file in node_modules (should be skipped)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "pkg", "index.js"), []byte(`password = "npm_secret"`), 0644)

	// Create a .podx file (should be skipped)
	os.WriteFile(filepath.Join(tmpDir, ".env.podx"), []byte(`password = "encrypted"`), 0644)

	results, err := ScanDirectory(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should only find the config.js file
	if len(results) != 1 {
		t.Errorf("ScanDirectory() found %d files with secrets, want 1", len(results))
	}

	if len(results) > 0 && results[0].Path != "config.js" {
		t.Errorf("ScanDirectory() found %s, want config.js", results[0].Path)
	}
}
