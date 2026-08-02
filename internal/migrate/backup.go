package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// backupsDir returns the path to %APPDATA%\MoveKit\backups and ensures it exists.
func backupsDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	return filepath.Join(appData, "MoveKit", "backups")
}

// backupPathFor returns the full path for a backup JSON file given a folder name.
func backupPathFor(folder string) string {
	return filepath.Join(backupsDir(), folder+".json")
}

// SaveBackup persists a BackupRecord as JSON to %APPDATA%\MoveKit\backups\{folder}.json.
func SaveBackup(b BackupRecord) error {
	dir := backupsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	path := backupPathFor(b.RegValueName)
	return os.WriteFile(path, data, 0600) // owner-only for security
}

// LoadBackup reads a BackupRecord from disk for the given folder name.
func LoadBackup(folder string) (*BackupRecord, error) {
	path := backupPathFor(folder)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec BackupRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// LoadAllBackups reads every backup JSON file from the backups directory.
func LoadAllBackups() ([]BackupRecord, error) {
	dir := backupsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var records []BackupRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var rec BackupRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

// DeleteBackup removes the backup JSON file for the given folder name.
func DeleteBackup(folder string) error {
	path := backupPathFor(folder)
	return os.Remove(path)
}
