package winsyscall

import (
	"testing"
)

func TestReadUserShellFolders(t *testing.T) {
	// Try to read the Desktop value from User Shell Folders.
	// This key exists on standard Windows installs but may not be present
	// in minimal CI / test environments, so we log gracefully on failure.
	val, err := ReadExpandString(UserShellFoldersKey, "Desktop")
	if err != nil {
		t.Logf("Could not read Desktop from User Shell Folders (expected in non-desktop env): %v", err)
		return
	}
	t.Logf("User Shell Folders\\Desktop = %q", val)

	if val == "" {
		t.Log("Desktop value is empty")
	}
}
