package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	GitHubRepo    = "dwirx/podx"
	GitHubAPIBase = "https://api.github.com/repos"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
	Body    string  `json:"body"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetLatestRelease fetches the latest release from GitHub
func GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/%s/releases/latest", GitHubAPIBase, GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// GetAssetName returns the correct asset name for current platform
func GetAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	switch {
	case goos == "linux" && goarch == "amd64":
		return "podx-linux-amd64"
	case goos == "linux" && goarch == "arm64":
		return "podx-linux-arm64"
	case goos == "darwin" && goarch == "amd64":
		return "podx-darwin-amd64"
	case goos == "darwin" && goarch == "arm64":
		return "podx-darwin-arm64"
	case goos == "windows" && goarch == "amd64":
		return "podx-windows-amd64.exe"
	default:
		return fmt.Sprintf("podx-%s-%s", goos, goarch)
	}
}

// DownloadAsset downloads the binary to a temporary file
func DownloadAsset(release *Release) (string, error) {
	assetName := GetAssetName()

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download to temp file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "podx-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// Update performs the self-update
func Update(currentVersion string) error {
	fmt.Println("🔍 Checking for updates...")

	release, err := GetLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")

	if latestVersion == currentClean {
		fmt.Printf("✓ Already on latest version (%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("📦 New version available: %s → %s\n", currentVersion, release.TagName)
	fmt.Println("⬇️  Downloading...")

	tmpPath, err := DownloadAsset(release)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Make temp file executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Backup current binary
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Restore backup
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	fmt.Printf("✅ Updated to %s\n", release.TagName)
	fmt.Println("   Run 'podx version' to verify")

	return nil
}

// CheckUpdate checks if update is available (non-blocking)
func CheckUpdate(currentVersion string) (string, bool) {
	release, err := GetLatestRelease()
	if err != nil {
		return "", false
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")

	if latestVersion != currentClean {
		return release.TagName, true
	}

	return "", false
}
