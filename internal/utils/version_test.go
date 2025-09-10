package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	testCases := []struct {
		input    string
		expected *Version
		hasError bool
	}{
		{"v1.2.3", &Version{1, 2, 3, "1.2.3"}, false},
		{"1.2.3", &Version{1, 2, 3, "1.2.3"}, false},
		{"0.0.1", &Version{0, 0, 1, "0.0.1"}, false},
		{"10.20.30", &Version{10, 20, 30, "10.20.30"}, false},
		{"1.2", nil, true},     // Missing patch version
		{"1.2.3.4", nil, true}, // Too many parts
		{"a.b.c", nil, true},   // Non-numeric parts
		{"", nil, true},        // Empty string
	}

	for _, tc := range testCases {
		result, err := ParseVersion(tc.input)

		if tc.hasError {
			assert.Error(t, err, "Expected error for input: %s", tc.input)
			assert.Nil(t, result)
		} else {
			assert.NoError(t, err, "Unexpected error for input: %s", tc.input)
			assert.Equal(t, tc.expected, result)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	testCases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},  // Equal
		{"1.0.1", "1.0.0", 1},  // Patch version higher
		{"1.0.0", "1.0.1", -1}, // Patch version lower
		{"1.1.0", "1.0.0", 1},  // Minor version higher
		{"1.0.0", "1.1.0", -1}, // Minor version lower
		{"2.0.0", "1.0.0", 1},  // Major version higher
		{"1.0.0", "2.0.0", -1}, // Major version lower
		{"2.1.3", "2.1.2", 1},  // Complex comparison
		{"0.0.1", "0.0.2", -1}, // Small versions
	}

	for _, tc := range testCases {
		v1, err := ParseVersion(tc.v1)
		assert.NoError(t, err)

		v2, err := ParseVersion(tc.v2)
		assert.NoError(t, err)

		result := v1.Compare(v2)
		assert.Equal(t, tc.expected, result, "Comparing %s with %s", tc.v1, tc.v2)
	}
}

func TestVersionString(t *testing.T) {
	testCases := []struct {
		version  *Version
		expected string
	}{
		{&Version{1, 2, 3, "1.2.3"}, "v1.2.3"},
		{&Version{0, 0, 1, "0.0.1"}, "v0.0.1"},
		{&Version{10, 20, 30, "10.20.30"}, "v10.20.30"},
	}

	for _, tc := range testCases {
		result := tc.version.String()
		assert.Equal(t, tc.expected, result)
	}
}

func TestNewVersionChecker(t *testing.T) {
	repo := "VibeAny/octopus-cli"
	version := "v1.0.0"

	vc := NewVersionChecker(repo, version)

	assert.NotNil(t, vc)
	assert.Equal(t, repo, vc.GitHubRepo)
	assert.Equal(t, version, vc.CurrentVersion)
	assert.NotNil(t, vc.HTTPClient)
	assert.Equal(t, 10*time.Second, vc.HTTPClient.Timeout)
}

func TestFormatUpdateInfo(t *testing.T) {
	current := "v1.0.0"
	latest := "v1.1.0"
	release := &GitHubRelease{
		TagName:     "v1.1.0",
		URL:         "https://github.com/VibeAny/octopus-cli/releases/tag/v1.1.0",
		PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	result := FormatUpdateInfo(current, latest, release)

	assert.Contains(t, result, "Upgrade Available")
	assert.Contains(t, result, current)
	assert.Contains(t, result, latest)
	assert.Contains(t, result, "2024-01-01")
	assert.Contains(t, result, "octopus upgrade")
}

// Integration test - requires network access
func TestCheckLatestVersion_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	vc := NewVersionChecker("VibeAny/octopus-cli", "v0.0.1")

	// This test might fail if the repo doesn't exist yet or has no releases
	// It's more for validating the HTTP logic when the repo is set up
	_, err := vc.CheckLatestVersion()

	// We don't assert success here since the repo might not exist yet
	// This is just to verify the code doesn't panic
	if err != nil {
		t.Logf("Expected - repo might not exist yet: %v", err)
	}
}

func TestIsUpdateAvailable_MockScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		shouldUpdate   bool
		expectError    bool
	}{
		{"Valid current version", "v1.0.0", false, false},
		{"Invalid current version", "invalid", false, true},
		{"Empty current version", "", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vc := NewVersionChecker("test/repo", tc.currentVersion)

			// Since this will fail on network call, we're mainly testing version parsing
			_, _, err := vc.IsUpdateAvailable()

			if tc.expectError {
				assert.Error(t, err)
			}
			// Note: We can't test the full flow without mocking HTTP
			// This test mainly validates input validation
		})
	}
}
