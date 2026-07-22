package migrate

import "fmt"

// Version is the current version of the MoveKit migration engine.
const Version = "0.2.0"

// CheckUpdate checks for available updates. This is a placeholder that
// always returns no update available. A real implementation would query
// a remote API or GitHub releases.
func CheckUpdate() (*UpdateInfo, error) {
	// Placeholder: no update check implemented yet.
	return &UpdateInfo{
		HasUpdate:   false,
		NewVersion:  Version,
		DownloadURL: "",
	}, nil
}

// VersionString returns a human-readable version string.
func VersionString() string {
	return fmt.Sprintf("MoveKit v%s", Version)
}
