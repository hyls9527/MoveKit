package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveBackup_ReadOnlyDir(t *testing.T) {
	// Redirect APPDATA to a temp directory where backups dir will fail to create.
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	// Place a file where the "backups" directory should be, causing MkdirAll to fail.
	backupDir := filepath.Join(tmpDir, "MoveKit", "backups")
	os.MkdirAll(filepath.Join(tmpDir, "MoveKit"), 0755)
	os.WriteFile(backupDir, []byte("not-a-dir"), 0444)

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "Desktop",
		OriginalPath:  `C:\Users\test\Desktop`,
		NewPath:       `D:\Desktop`,
		RegistryValue: "Desktop",
		Status:        "in_progress",
	}

	err := saveBackup(filepath.Join(backupDir, "Desktop.json"), rec)
	if err == nil {
		t.Error("saveBackup should return error when backups dir is a file, got nil")
	}
}

func TestSaveBackup_ValidDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "Personal",
		OriginalPath:  `C:\Users\test\Documents`,
		NewPath:       `D:\Documents`,
		RegistryValue: "Personal",
		Status:        "in_progress",
	}

	err := saveBackup(backupFilePath("Personal"), rec)
	if err != nil {
		t.Fatalf("saveBackup failed: %v", err)
	}

	// Verify file was created.
	backupFile := filepath.Join(tmpDir, "MoveKit", "backups", "Personal.json")
	if _, err := os.Stat(backupFile); err != nil {
		t.Errorf("backup file not created: %v", err)
	}
}
