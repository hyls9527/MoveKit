package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// stateFilePath returns the path to %APPDATA%\MoveKit\state.json.
func stateFilePath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	return filepath.Join(appData, "MoveKit", "state.json")
}

// SaveState persists the current MigrationState to disk.
func SaveState(state MigrationState) error {
	dir := filepath.Dir(stateFilePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFilePath(), data, 0644)
}

// LoadState reads the MigrationState from disk.
// Returns nil (not an error) if no state file exists.
func LoadState() (*MigrationState, error) {
	data, err := os.ReadFile(stateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var state MigrationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// ClearState removes the state file from disk.
func ClearState() error {
	err := os.Remove(stateFilePath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
