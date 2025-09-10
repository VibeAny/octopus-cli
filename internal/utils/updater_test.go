package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpdateManager(t *testing.T) {
	repo := "VibeAny/octopus-cli"
	version := "v1.0.0"

	um := NewUpdateManager(repo, version)

	assert.NotNil(t, um)
	assert.Equal(t, repo, um.GitHubRepo)
	assert.Equal(t, version, um.CurrentVersion)
	assert.NotNil(t, um.HTTPClient)
	assert.Contains(t, um.TempDir, "octopus-update")
}

func TestGetCurrentPlatform(t *testing.T) {
	platform := GetCurrentPlatform()

	assert.NotEmpty(t, platform.OS)
	assert.NotEmpty(t, platform.Arch)

	// OS should be normalized for platform consistency
	expectedOS := runtime.GOOS
	if expectedOS == "darwin" {
		expectedOS = "macos" // Darwin is mapped to macOS for artifact naming consistency
	}
	assert.Equal(t, expectedOS, platform.OS)

	// Architecture should be normalized
	expectedArch := runtime.GOARCH
	if expectedArch == "amd64" || expectedArch == "arm64" || expectedArch == "386" {
		assert.Equal(t, expectedArch, platform.Arch)
	} else {
		// Fallback should be amd64
		assert.Equal(t, "amd64", platform.Arch)
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	um := NewUpdateManager("test/repo", "v1.0.0")

	// Mock release with assets
	release := &GitHubRelease{
		Assets: []GitHubReleaseAsset{
			{Name: "octopus-v1.0.0-windows-amd64-20240101.12345678.exe", Size: 1000},
			{Name: "octopus-v1.0.0-linux-amd64-20240101.12345678", Size: 1000},
			{Name: "octopus-v1.0.0-macos-arm64-20240101.12345678", Size: 1000},
			{Name: "octopus-v1.0.0-linux-amd64-20240101.12345678.sha256", Size: 64}, // Should be ignored
		},
	}

	testCases := []struct {
		platform PlatformInfo
		expected string
		hasError bool
	}{
		{PlatformInfo{"windows", "amd64"}, "octopus-v1.0.0-windows-amd64-20240101.12345678.exe", false},
		{PlatformInfo{"linux", "amd64"}, "octopus-v1.0.0-linux-amd64-20240101.12345678", false},
		{PlatformInfo{"macos", "arm64"}, "octopus-v1.0.0-macos-arm64-20240101.12345678", false},
		{PlatformInfo{"freebsd", "amd64"}, "", true}, // No asset for this platform
	}

	for _, tc := range testCases {
		asset, err := um.FindAssetForPlatform(release, tc.platform)

		if tc.hasError {
			assert.Error(t, err)
			assert.Nil(t, asset)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, asset)
			assert.Equal(t, tc.expected, asset.Name)
		}
	}
}

func TestVerifyDownload(t *testing.T) {
	um := NewUpdateManager("test/repo", "v1.0.0")

	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")

	testData := []byte("test data for verification")
	err := os.WriteFile(testFile, testData, 0644)
	assert.NoError(t, err)

	// Test correct size verification
	err = um.VerifyDownload(testFile, int64(len(testData)))
	assert.NoError(t, err)

	// Test incorrect size verification
	err = um.VerifyDownload(testFile, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file size mismatch")

	// Test non-existent file
	err = um.VerifyDownload("/nonexistent/file", 100)
	assert.Error(t, err)
}

func TestCalculateChecksum(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")

	testData := []byte("test data for checksum")
	err := os.WriteFile(testFile, testData, 0644)
	assert.NoError(t, err)

	checksum, err := CalculateChecksum(testFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, checksum)
	assert.Len(t, checksum, 64) // SHA256 produces 64 character hex string

	// Test same data produces same checksum
	checksum2, err := CalculateChecksum(testFile)
	assert.NoError(t, err)
	assert.Equal(t, checksum, checksum2)

	// Test non-existent file
	_, err = CalculateChecksum("/nonexistent/file")
	assert.Error(t, err)
}

func TestCleanup(t *testing.T) {
	um := NewUpdateManager("test/repo", "v1.0.0")

	// Create the temp directory
	err := os.MkdirAll(um.TempDir, 0755)
	assert.NoError(t, err)

	// Verify it exists
	_, err = os.Stat(um.TempDir)
	assert.NoError(t, err)

	// Cleanup
	err = um.Cleanup()
	assert.NoError(t, err)

	// Verify it's gone
	_, err = os.Stat(um.TempDir)
	assert.True(t, os.IsNotExist(err))
}

func TestProgressReader(t *testing.T) {
	testData := []byte("this is test data for progress reader testing")
	totalSize := int64(len(testData))

	var progressUpdates []DownloadProgress
	callback := func(progress DownloadProgress) {
		progressUpdates = append(progressUpdates, progress)
	}

	// Create progress reader
	reader := &ProgressReader{
		Reader:     strings.NewReader(string(testData)),
		Total:      totalSize,
		OnProgress: callback,
	}

	// Read all data
	buf := make([]byte, 1024)
	totalRead := 0
	for {
		n, err := reader.Read(buf)
		totalRead += n
		if err != nil {
			break
		}
	}

	assert.Equal(t, len(testData), totalRead)
	assert.NotEmpty(t, progressUpdates)

	// Last progress should be 100%
	lastProgress := progressUpdates[len(progressUpdates)-1]
	assert.Equal(t, totalSize, lastProgress.Total)
	assert.Equal(t, totalSize, lastProgress.Downloaded)
	assert.Equal(t, float64(100), lastProgress.Percentage)
}

func TestFormatBytes(t *testing.T) {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := formatBytes(tc.bytes)
		assert.Equal(t, tc.expected, result)
	}
}
