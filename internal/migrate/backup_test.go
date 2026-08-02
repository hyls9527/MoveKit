package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadDelete(t *testing.T) {
	// Override the backups directory to use a temp dir.
	origAppData := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)
	defer t.Setenv("APPDATA", origAppData)

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "Desktop",
		OriginalPath:  `C:\Users\test\Desktop`,
		NewPath:       `D:\Desktop`,
		RegistryValue: "Desktop",
		Status:        "complete",
	}

	// Save.
	if err := SaveBackup(rec); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	// Verify file exists.
	backupFile := filepath.Join(tmpDir, "MoveKit", "backups", "Desktop.json")
	if _, err := os.Stat(backupFile); err != nil {
		t.Fatalf("backup file not created: %v", err)
	}

	// Load.
	loaded, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if loaded.RegValueName != "Desktop" {
		t.Errorf("RegValueName = %q, want %q", loaded.RegValueName, "Desktop")
	}
	if loaded.NewPath != `D:\Desktop` {
		t.Errorf("NewPath = %q, want %q", loaded.NewPath, `D:\Desktop`)
	}
	if loaded.Status != "complete" {
		t.Errorf("Status = %q, want %q", loaded.Status, "complete")
	}

	// Save a second backup.
	rec2 := BackupRecord{
		Version:      1,
		RegValueName: "Personal",
		OriginalPath: `C:\Users\test\Documents`,
		NewPath:      `D:\Documents`,
		Status:       "in_progress",
	}
	if err := SaveBackup(rec2); err != nil {
		t.Fatalf("SaveBackup(2): %v", err)
	}

	// LoadAllBackups.
	all, err := LoadAllBackups()
	if err != nil {
		t.Fatalf("LoadAllBackups: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("LoadAllBackups returned %d records, want 2", len(all))
	}

	// Delete.
	if err := DeleteBackup("Desktop"); err != nil {
		t.Fatalf("DeleteBackup: %v", err)
	}

	// Verify deleted.
	_, err = LoadBackup("Desktop")
	if err == nil {
		t.Error("LoadBackup should fail after delete, but got nil error")
	}

	// Clean up second backup.
	if err := DeleteBackup("Personal"); err != nil {
		t.Fatalf("DeleteBackup(2): %v", err)
	}
}

func TestLoadAllBackups_EmptyDir(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)
	defer t.Setenv("APPDATA", origAppData)

	all, err := LoadAllBackups()
	if err != nil {
		t.Fatalf("LoadAllBackups on empty dir: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 backups, got %d", len(all))
	}
}

func TestLoadBackup_NotFound(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)
	defer t.Setenv("APPDATA", origAppData)

	_, err := LoadBackup("NonExistent")
	if err == nil {
		t.Error("expected error for non-existent backup, got nil")
	}
}

func TestLoadBackup_MalformedJSON(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)
	defer t.Setenv("APPDATA", origAppData)

	backupDir := filepath.Join(tmpDir, "MoveKit", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "Desktop.json"), []byte("{broken"), 0644); err != nil {
		t.Fatalf("write malformed backup: %v", err)
	}

	if _, err := LoadBackup("Desktop"); err == nil {
		t.Error("malformed backup JSON should return an error")
	}
}
