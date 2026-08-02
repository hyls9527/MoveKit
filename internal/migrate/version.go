package migrate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Version is the current version of the MoveKit migration engine.
const Version = "0.2.1"

// githubReleasesLatestURL is the GitHub API endpoint for the latest release.
// It is a variable so tests can point it at a local test server.
var githubReleasesLatestURL = "https://api.github.com/repos/hyls9527/MoveKit/releases/latest"

// updateCheckClient is overridable in tests.
var updateCheckClient = &http.Client{Timeout: 5 * time.Second}

// CheckUpdate queries the GitHub Releases API for the latest published
// version and reports whether a newer version is available.
func CheckUpdate() (*UpdateInfo, error) {
	req, err := http.NewRequest(http.MethodGet, githubReleasesLatestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("check update: %w", err)
	}
	req.Header.Set("User-Agent", "MoveKit/"+Version)

	resp, err := updateCheckClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("check update: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check update: unexpected status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("check update: parse response: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	return &UpdateInfo{
		HasUpdate:   compareVersions(latest, Version) > 0,
		NewVersion:  latest,
		DownloadURL: release.HTMLURL,
	}, nil
}

// compareVersions compares two dotted numeric version strings (e.g. "1.2.3").
// It returns -1, 0, or 1 depending on whether a is older, equal, or newer.
func compareVersions(a, b string) int {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}

func parseVersion(s string) []int {
	parts := strings.Split(s, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			break
		}
		out = append(out, n)
	}
	return out
}

// VersionString returns a human-readable version string.
func VersionString() string {
	return fmt.Sprintf("MoveKit v%s", Version)
}
