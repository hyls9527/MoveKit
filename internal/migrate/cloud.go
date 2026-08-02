package migrate

// DetectCloudSync checks for running cloud sync processes (OneDrive, Google
// Drive, iCloud, Dropbox) and returns a list of detected providers.
func DetectCloudSync() []CloudProvider {
	var providers []CloudProvider

	checks := []struct {
		name      string
		processes []string
	}{
		{name: "OneDrive", processes: []string{"OneDrive.exe"}},
		{name: "Google Drive", processes: []string{"GoogleDriveFS.exe"}},
		{name: "iCloud", processes: []string{"iCloudDrive.exe"}},
		{name: "Dropbox", processes: []string{"Dropbox.exe"}},
	}
	for _, c := range checks {
		for _, p := range c.processes {
			if Sys.Process.IsProcessRunning(p) {
				providers = append(providers, CloudProvider{Name: c.name})
				break
			}
		}
	}
	return providers
}
