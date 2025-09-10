package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Body        string               `json:"body"`
	URL         string               `json:"html_url"`
	PublishedAt time.Time            `json:"published_at"`
	Assets      []GitHubReleaseAsset `json:"assets"`
	Prerelease  bool                 `json:"prerelease"`
	Draft       bool                 `json:"draft"`
}

// GitHubReleaseAsset represents a GitHub release asset
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

// VersionChecker handles version checking and comparison
type VersionChecker struct {
	GitHubRepo     string
	CurrentVersion string
	HTTPClient     *http.Client
}

// NewVersionChecker creates a new version checker
func NewVersionChecker(repo, currentVersion string) *VersionChecker {
	return &VersionChecker{
		GitHubRepo:     repo,
		CurrentVersion: currentVersion,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ParseVersion parses a semantic version string
func ParseVersion(versionStr string) (*Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Regular expression for semantic versioning
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Raw:   versionStr,
	}, nil
}

// Compare compares two versions
// Returns:
//
//	-1 if v < other
//	 0 if v == other
//	 1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	return 0
}

// String returns the string representation of the version
func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// CheckLatestVersion checks for the latest version on GitHub
func (vc *VersionChecker) CheckLatestVersion() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", vc.GitHubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", "Octopus-CLI/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := vc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Skip drafts and prereleases
	if release.Draft || release.Prerelease {
		return nil, fmt.Errorf("latest release is draft or prerelease")
	}

	return &release, nil
}

// IsUpdateAvailable checks if an update is available
func (vc *VersionChecker) IsUpdateAvailable() (bool, *GitHubRelease, error) {
	latestRelease, err := vc.CheckLatestVersion()
	if err != nil {
		return false, nil, err
	}

	currentVersion, err := ParseVersion(vc.CurrentVersion)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse current version: %w", err)
	}

	latestVersion, err := ParseVersion(latestRelease.TagName)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse latest version: %w", err)
	}

	isNewer := latestVersion.Compare(currentVersion) > 0
	return isNewer, latestRelease, nil
}

// FormatUpdateInfo formats update information for display
func FormatUpdateInfo(current, latest string, release *GitHubRelease) string {
	info := "📦 Upgrade Available!\n"
	info += fmt.Sprintf("   Current: %s\n", FormatBold(current))
	info += fmt.Sprintf("   Latest:  %s\n", FormatHighlight(latest))

	if release != nil {
		info += fmt.Sprintf("   Released: %s\n", FormatDim(release.PublishedAt.Format("2006-01-02")))
		if release.URL != "" {
			info += fmt.Sprintf("   Details: %s\n", FormatDim(release.URL))
		}
	}

	info += fmt.Sprintf("\nRun '%s' to upgrade.", FormatHighlight("octopus upgrade"))

	return info
}
