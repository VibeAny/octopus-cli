package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// UpdateManager handles downloading and installing updates
type UpdateManager struct {
	GitHubRepo     string
	CurrentVersion string
	HTTPClient     *http.Client
	TempDir        string
}

// PlatformInfo represents current platform information
type PlatformInfo struct {
	OS   string
	Arch string
}

// DownloadProgress represents download progress
type DownloadProgress struct {
	Total      int64
	Downloaded int64
	Percentage float64
	Speed      string
	ETA        string
}

// ProgressCallback is called during download to report progress
type ProgressCallback func(progress DownloadProgress)

// NewUpdateManager creates a new update manager
func NewUpdateManager(repo, currentVersion string) *UpdateManager {
	tempDir := filepath.Join(os.TempDir(), "octopus-update")

	return &UpdateManager{
		GitHubRepo:     repo,
		CurrentVersion: currentVersion,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		TempDir: tempDir,
	}
}

// GetCurrentPlatform detects the current platform and architecture
func GetCurrentPlatform() PlatformInfo {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Normalize OS names to match release naming
	switch osName {
	case "darwin":
		osName = "macos"
	case "windows":
		osName = "windows"
	case "linux":
		osName = "linux"
	}

	// Normalize architecture names to match release naming
	switch archName {
	case "amd64":
		archName = "amd64"
	case "arm64":
		archName = "arm64"
	case "386":
		archName = "386"
	default:
		archName = "amd64" // fallback
	}

	return PlatformInfo{
		OS:   osName,
		Arch: archName,
	}
}

// FindAssetForPlatform finds the appropriate asset for the current platform
func (um *UpdateManager) FindAssetForPlatform(release *GitHubRelease, platform PlatformInfo) (*GitHubReleaseAsset, error) {
	// Expected naming pattern: octopus-v1.0.0-platform-arch-YYYYMMDD.xxxxxxxx[.exe]
	var candidates []GitHubReleaseAsset

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)

		// Check if asset matches our platform
		if strings.Contains(name, platform.OS) && strings.Contains(name, platform.Arch) {
			// Skip if it's a checksum file
			if strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".md5") {
				continue
			}

			candidates = append(candidates, asset)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no asset found for platform %s-%s", platform.OS, platform.Arch)
	}

	// Return the first matching candidate
	return &candidates[0], nil
}

// DownloadUpdate downloads the update file
func (um *UpdateManager) DownloadUpdate(asset *GitHubReleaseAsset, progressCallback ProgressCallback) (string, error) {
	// Create temp directory
	if err := os.MkdirAll(um.TempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download file path
	downloadPath := filepath.Join(um.TempDir, asset.Name)

	// Create the download request
	req, err := http.NewRequest("GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("User-Agent", "Octopus-CLI/1.0")

	// Make the request
	resp, err := um.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create the output file
	outFile, err := os.Create(downloadPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength

	// Create progress tracking reader
	var reader io.Reader = resp.Body
	if progressCallback != nil && contentLength > 0 {
		reader = &ProgressReader{
			Reader:     resp.Body,
			Total:      contentLength,
			OnProgress: progressCallback,
		}
	}

	// Copy with progress
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return "", fmt.Errorf("failed to save downloaded file: %w", err)
	}

	return downloadPath, nil
}

// VerifyDownload verifies the downloaded file (basic size check)
func (um *UpdateManager) VerifyDownload(filePath string, expectedSize int64) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}

	if stat.Size() != expectedSize {
		return fmt.Errorf("file size mismatch: expected %d, got %d", expectedSize, stat.Size())
	}

	return nil
}

// CalculateChecksum calculates SHA256 checksum of a file
func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// BackupCurrentBinary creates a backup of the current binary
func (um *UpdateManager) BackupCurrentBinary() (string, error) {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Create backup path
	backupPath := currentPath + ".backup"

	// Copy current binary to backup
	if err := copyFile(currentPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// InstallUpdate replaces the current binary with the new one
func (um *UpdateManager) InstallUpdate(updatePath string) error {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Make the update file executable
	if err := os.Chmod(updatePath, 0755); err != nil {
		return fmt.Errorf("failed to make update executable: %w", err)
	}

	// Replace current binary
	if err := os.Rename(updatePath, currentPath); err != nil {
		return fmt.Errorf("failed to replace current binary: %w", err)
	}

	return nil
}

// Cleanup removes temporary files
func (um *UpdateManager) Cleanup() error {
	return os.RemoveAll(um.TempDir)
}

// RestoreFromBackup restores the original binary from backup
func (um *UpdateManager) RestoreFromBackup(backupPath string) error {
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	return os.Rename(backupPath, currentPath)
}

// ProgressReader wraps an io.Reader to provide progress callbacks
type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	OnProgress ProgressCallback
	read       int64
	startTime  time.Time
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	if pr.startTime.IsZero() {
		pr.startTime = time.Now()
	}

	n, err = pr.Reader.Read(p)
	pr.read += int64(n)

	if pr.OnProgress != nil {
		elapsed := time.Since(pr.startTime)
		percentage := float64(pr.read) / float64(pr.Total) * 100

		var speed, eta string
		if elapsed.Seconds() > 0 {
			bytesPerSecond := float64(pr.read) / elapsed.Seconds()
			speed = formatBytes(int64(bytesPerSecond)) + "/s"

			if bytesPerSecond > 0 {
				remaining := float64(pr.Total-pr.read) / bytesPerSecond
				eta = time.Duration(remaining * float64(time.Second)).Round(time.Second).String()
			}
		}

		pr.OnProgress(DownloadProgress{
			Total:      pr.Total,
			Downloaded: pr.read,
			Percentage: percentage,
			Speed:      speed,
			ETA:        eta,
		})
	}

	return n, err
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

// formatBytes formats bytes as human readable string
func formatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}

	if bytes < 1024 {
		return strconv.FormatInt(bytes, 10) + " B"
	}

	div, exp := int64(1024), 0
	for n := bytes / 1024; n >= 1024; n /= 1024 {
		div *= 1024
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp+1])
}
