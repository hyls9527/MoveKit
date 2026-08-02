package migrate

import (
	"encoding/json"
	"fmt"
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
	if err := os.MkdirAll(dir, 0750); err != nil {
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
	// #nosec G304 -- path is built from %APPDATA% and a known folder name
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
// Corrupted or unreadable files are skipped and reported in the returned
// warnings slice so callers can surface them instead of failing silently.
func LoadAllBackups() ([]BackupRecord, []string, error) {
	dir := backupsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var records []BackupRecord
	var warnings []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		// #nosec G304 -- path is built from %APPDATA% and directory entries
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("备份文件 %s 无法读取: %v", entry.Name(), err))
			continue
		}
		var rec BackupRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			warnings = append(warnings, fmt.Sprintf("备份文件 %s 已损坏，已跳过: %v", entry.Name(), err))
			continue
		}
		records = append(records, rec)
	}
	return records, warnings, nil
}

// DeleteBackup removes the backup JSON file for the given folder name.
func DeleteBackup(folder string) error {
	path := backupPathFor(folder)
	return os.Remove(path)
}
