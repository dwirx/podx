package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	GitHubRepo    = "dwirx/podx"
	GitHubAPIBase = "https://api.github.com/repos"
	CheckInterval = 24 * time.Hour // Check for updates once per day
)

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Assets      []Asset   `json:"assets"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo contains information about available update
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseNotes   string
	PublishedAt    time.Time
	DownloadSize   int64
	IsBeta         bool
}

// GetLatestRelease fetches the latest release from GitHub
func GetLatestRelease() (*Release, error) {
	return getReleaseByTag("latest")
}

// GetBetaRelease fetches the beta release from GitHub
func GetBetaRelease() (*Release, error) {
	return getReleaseByTag("beta")
}

// getReleaseByTag fetches a specific release
func getReleaseByTag(tag string) (*Release, error) {
	var url string
	if tag == "latest" {
		url = fmt.Sprintf("%s/%s/releases/latest", GitHubAPIBase, GitHubRepo)
	} else {
		url = fmt.Sprintf("%s/%s/releases/tags/%s", GitHubAPIBase, GitHubRepo, tag)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found")
	}

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
	case goos == "linux" && goarch == "arm":
		return "podx-linux-arm"
	case goos == "darwin" && goarch == "amd64":
		return "podx-darwin-amd64"
	case goos == "darwin" && goarch == "arm64":
		return "podx-darwin-arm64"
	case goos == "windows" && goarch == "amd64":
		return "podx-windows-amd64.exe"
	case goos == "windows" && goarch == "arm64":
		return "podx-windows-arm64.exe"
	default:
		return fmt.Sprintf("podx-%s-%s", goos, goarch)
	}
}

// CheckUpdate checks if update is available
func CheckUpdate(currentVersion string) *UpdateInfo {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		Available:      false,
	}

	release, err := GetLatestRelease()
	if err != nil {
		return info
	}

	if compareVersions(currentVersion, release.TagName) < 0 {
		info.Available = true
		info.LatestVersion = release.TagName
		info.ReleaseNotes = release.Body
		info.PublishedAt = release.PublishedAt
		info.IsBeta = release.Prerelease

		// Get download size
		assetName := GetAssetName()
		for _, asset := range release.Assets {
			if asset.Name == assetName {
				info.DownloadSize = asset.Size
				break
			}
		}
	}

	return info
}

// CheckUpdateAsync checks for updates in background and calls callback
func CheckUpdateAsync(currentVersion string, callback func(*UpdateInfo)) {
	go func() {
		info := CheckUpdate(currentVersion)
		if callback != nil {
			callback(info)
		}
	}()
}

// compareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Handle beta/dev versions
	v1Parts := strings.Split(v1, "-")
	v2Parts := strings.Split(v2, "-")

	// Compare main version parts
	v1Main := strings.Split(v1Parts[0], ".")
	v2Main := strings.Split(v2Parts[0], ".")

	maxLen := len(v1Main)
	if len(v2Main) > maxLen {
		maxLen = len(v2Main)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Main) {
			n1, _ = strconv.Atoi(v1Main[i])
		}
		if i < len(v2Main) {
			n2, _ = strconv.Atoi(v2Main[i])
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// If main versions are equal, check pre-release
	// A version without pre-release is greater than one with pre-release
	if len(v1Parts) > 1 && len(v2Parts) == 1 {
		return -1 // v1 is pre-release, v2 is not
	}
	if len(v1Parts) == 1 && len(v2Parts) > 1 {
		return 1 // v1 is not pre-release, v2 is
	}

	return 0
}

// DownloadAsset downloads the binary with progress reporting
func DownloadAsset(release *Release, progressFn func(downloaded, total int64)) (string, error) {
	assetName := GetAssetName()

	var downloadURL string
	var totalSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			totalSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download to temp file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
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

	// Copy with progress
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := tmpFile.Write(buf[:n])
			if writeErr != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return "", fmt.Errorf("failed to write file: %w", writeErr)
			}
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, totalSize)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to download: %w", err)
		}
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// Update performs the self-update with better error handling
func Update(currentVersion string, beta bool) error {
	fmt.Println("🔍 Checking for updates...")

	var release *Release
	var err error

	if beta {
		release, err = GetBetaRelease()
		if err != nil {
			return fmt.Errorf("failed to fetch beta release: %w", err)
		}
		fmt.Println("📦 Using beta channel")
	} else {
		release, err = GetLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
	}

	cmp := compareVersions(currentVersion, release.TagName)
	if cmp >= 0 && !beta {
		fmt.Printf("✅ Already on latest version (%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("📦 New version available: %s → %s\n", currentVersion, release.TagName)
	if release.Prerelease {
		fmt.Println("   ⚠️  This is a pre-release version")
	}

	// Get asset info
	assetName := GetAssetName()
	var downloadSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadSize = asset.Size
			break
		}
	}

	if downloadSize > 0 {
		fmt.Printf("   Size: %.2f MB\n", float64(downloadSize)/1024/1024)
	}

	fmt.Println("\n⬇️  Downloading...")

	// Download with progress
	lastPercent := -1
	tmpPath, err := DownloadAsset(release, func(downloaded, total int64) {
		if total > 0 {
			percent := int(downloaded * 100 / total)
			if percent != lastPercent && percent%10 == 0 {
				fmt.Printf("   %d%%\n", percent)
				lastPercent = percent
			}
		}
	})
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpPath)

	fmt.Println("   100%")

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Make temp file executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Verify the downloaded binary works
	fmt.Println("\n🔍 Verifying download...")
	if err := verifyBinary(tmpPath); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	fmt.Println("   ✓ Binary verified")

	// Check if we have write permission to the binary directory
	if !canWrite(execPath) {
		return runWithElevatedPrivileges(beta)
	}

	fmt.Println("\n📦 Installing...")

	// Backup current binary
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		if os.IsPermission(err) {
			return runWithElevatedPrivileges(beta)
		}
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary
	if err := copyFile(tmpPath, execPath); err != nil {
		// Restore backup on failure
		restoreErr := os.Rename(backupPath, execPath)
		if restoreErr != nil {
			fmt.Printf("⚠️  Failed to restore backup: %v\n", restoreErr)
			fmt.Printf("   Backup location: %s\n", backupPath)
		}
		if os.IsPermission(err) {
			return runWithElevatedPrivileges(beta)
		}
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Set permissions on new binary
	if err := os.Chmod(execPath, 0755); err != nil {
		fmt.Printf("⚠️  Failed to set permissions: %v\n", err)
	}

	// Remove backup
	os.Remove(backupPath)

	fmt.Println("")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✅ Updated to %s\n", release.TagName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
	fmt.Println("Run 'podx version' to verify the update.")

	return nil
}

// verifyBinary checks if the downloaded binary is valid
func verifyBinary(path string) error {
	// Try to run --version on the binary
	cmd := exec.Command(path, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary execution failed: %w", err)
	}

	// Check that output contains expected content
	if !strings.Contains(strings.ToLower(string(output)), "podx") {
		return fmt.Errorf("unexpected output from binary")
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// canWrite checks if we have write permission to the file
func canWrite(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// runWithElevatedPrivileges re-runs the update command with sudo
func runWithElevatedPrivileges(beta bool) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if runtime.GOOS == "windows" {
		fmt.Println("")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⚠️  Administrator privileges required")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("")
		fmt.Println("Please run as Administrator:")
		fmt.Println("  1. Right-click Command Prompt or PowerShell")
		fmt.Println("  2. Select 'Run as Administrator'")
		fmt.Println("  3. Run: podx update")
		return fmt.Errorf("administrator privileges required")
	}

	// Unix systems - use sudo
	fmt.Println("")
	fmt.Println("🔐 Permission denied. Requesting elevated privileges...")
	fmt.Println("")

	args := []string{execPath, "update"}
	if beta {
		args = append(args, "--beta")
	}

	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("sudo failed: %w", err)
	}

	os.Exit(0)
	return nil
}

// Rollback restores the previous version if backup exists
func Rollback() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	backupPath := execPath + ".bak"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	fmt.Println("🔄 Rolling back to previous version...")

	// Remove current
	if err := os.Remove(execPath); err != nil {
		return fmt.Errorf("failed to remove current binary: %w", err)
	}

	// Restore backup
	if err := os.Rename(backupPath, execPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	fmt.Println("✅ Rollback complete")
	fmt.Println("   Run 'podx version' to verify")

	return nil
}

// FormatSize formats bytes to human-readable format
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
