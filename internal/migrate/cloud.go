package migrate

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectCloudSync checks for running cloud sync processes (OneDrive, iCloud,
// Google Drive) and returns a list of detected providers. This uses a heuristic
// approach by checking for known process-related markers on disk.
func DetectCloudSync() []CloudProvider {
	var providers []CloudProvider

	// Check for OneDrive.
	// OneDrive stores its sync state in %LOCALAPPDATA%\Microsoft\OneDrive.
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}

	onedrivePath := filepath.Join(localAppData, "Microsoft", "OneDrive")
	if _, err := os.Stat(onedrivePath); err == nil {
		providers = append(providers, CloudProvider{
			Name:   "OneDrive",
			Folder: onedrivePath,
		})
	}

	// Check for Google Drive.
	// Google Drive stores its config in %LOCALAPPDATA%\Google\DriveFS.
	gdrivePath := filepath.Join(localAppData, "Google", "DriveFS")
	if _, err := os.Stat(gdrivePath); err == nil {
		providers = append(providers, CloudProvider{
			Name:   "Google Drive",
			Folder: gdrivePath,
		})
	}

	// Check for iCloud.
	// iCloud stores its data in a well-known location under AppData.
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}

	icloudPath := filepath.Join(appData, "Apple Computer", "MobileSync")
	if _, err := os.Stat(icloudPath); err == nil {
		providers = append(providers, CloudProvider{
			Name:   "iCloud",
			Folder: icloudPath,
		})
	}

	// Also check the user's home directory for OneDrive folders (common pattern).
	homeDir := os.Getenv("USERPROFILE")
	if homeDir != "" {
		entries, _ := os.ReadDir(homeDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.Contains(strings.ToLower(name), "onedrive") {
				// Avoid duplicates.
				found := false
				for _, p := range providers {
					if p.Name == "OneDrive" {
						found = true
						break
					}
				}
				if !found {
					providers = append(providers, CloudProvider{
						Name:   "OneDrive",
						Folder: filepath.Join(homeDir, name),
					})
				}
			}
		}
	}

	return providers
}
