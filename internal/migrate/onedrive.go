package migrate

import (
	"strings"

	sys "github.com/movekit/movekit/internal/syscall"
)

// OneDriveStatus holds the result of OneDrive Known Folder Move detection.
type OneDriveStatus struct {
	IsActive    bool     `json:"isActive"`
	Folders     []string `json:"folders"`
	HasGPOBlock bool     `json:"hasGPOBlock"`
}

// DetectOneDriveKFM checks whether OneDrive Known Folder Move (KFM) is active
// by inspecting three registry sources:
//
//  1. User Shell Folders — any value whose path contains "OneDrive"
//  2. OneDrive's own tracking key SOFTWARE\Microsoft\OneDrive\Accounts\Personal → KfmFoldersMovedAll
//  3. GPO key SOFTWARE\Policies\Microsoft\OneDrive → KFMBlockOptOut
func DetectOneDriveKFM() (*OneDriveStatus, error) {
	status := &OneDriveStatus{}

	// 1. Check User Shell Folders for OneDrive in the path.
	shellPaths, err := scanShellFoldersForOneDrive()
	if err == nil && len(shellPaths) > 0 {
		status.IsActive = true
		status.Folders = shellPaths
	}

	// 2. Check OneDrive's own KFM tracking key.
	kfmAll, err := Sys.Registry.ReadInteger(
		`SOFTWARE\Microsoft\OneDrive\Accounts\Personal`,
		"KfmFoldersMovedAll",
	)
	if err == nil && kfmAll != 0 {
		status.IsActive = true
	}

	// 3. Check GPO block on KFM opt-out.
	gpoVal, err := Sys.Registry.ReadInteger(
		`SOFTWARE\Policies\Microsoft\OneDrive`,
		"KFMBlockOptOut",
	)
	if err == nil && gpoVal != 0 {
		status.HasGPOBlock = true
		status.IsActive = true
	}

	return status, nil
}

// scanShellFoldersForOneDrive reads all values from User Shell Folders and
// returns the value names whose data contains "OneDrive" (case-insensitive).
func scanShellFoldersForOneDrive() ([]string, error) {
	// We can't enumerate all values with the current registry helpers,
	// so we check the known folder registry value names that OneDrive
	// commonly redirects.
	knownValues := []string{
		"Desktop",
		"{374DE290-123F-4565-9164-39C4925E467B}", // Downloads
		"Personal",                               // Documents
		"My Pictures",                            // Pictures
		"My Music",                               // Music
	}

	var onedriveFolders []string
	for _, valName := range knownValues {
		path, err := Sys.Registry.ReadExpandString(sys.UserShellFoldersKey, valName)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(path), "onedrive") {
			onedriveFolders = append(onedriveFolders, valName)
		}
	}
	return onedriveFolders, nil
}
