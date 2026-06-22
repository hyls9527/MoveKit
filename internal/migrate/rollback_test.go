package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRollback_RecordNotFound(t *testing.T) {
	err := Rollback("nonexistent_folder")
	if err == nil {
		t.Fatal("expected error for nonexistent backup record, got nil")
	}
	mkErr, ok := err.(*MoveKitError)
	if !ok {
		t.Fatalf("expected *MoveKitError, got %T: %v", err, err)
	}
	if mkErr.Code != ErrCodeRegistryFailed {
		t.Errorf("expected code %s, got %s", ErrCodeRegistryFailed, mkErr.Code)
	}
}

func TestRollback_LoadBackupError(t *testing.T) {
	// Simulate rollback with a folder name that can't produce a valid backup path.
	// The LoadBackup call should fail gracefully.
	origSys := Sys
	defer func() { Sys = origSys }()

	err := Rollback("\x00invalid")
	if err == nil {
		t.Fatal("expected error for invalid backup folder name")
	}
}

func TestSaveBackup_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	appDataOrig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", appDataOrig)

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "TestDesktop",
		OriginalPath:  filepath.Join(tmpDir, "original"),
		NewPath:       "D:\\TestDesktop",
		RegistryValue: "TestDesktop",
		Status:        "in_progress",
	}

	err := SaveBackup(rec)
	if err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	loaded, err := LoadBackup("TestDesktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if loaded.RegValueName != rec.RegValueName {
		t.Errorf("RegValueName mismatch: got %s, want %s", loaded.RegValueName, rec.RegValueName)
	}
	if loaded.Status != rec.Status {
		t.Errorf("Status mismatch: got %s, want %s", loaded.Status, rec.Status)
	}

	_ = DeleteBackup("TestDesktop")
}

func TestDeleteBackup_Nonexistent(t *testing.T) {
	err := DeleteBackup("nonexistent_folder_xyz")
	if err == nil {
		t.Log("DeleteBackup succeeded for nonexistent (file may not have existed)")
	}
}

func TestCleanupExpiredBackups_NoBackups(t *testing.T) {
	count, err := CleanupExpiredBackups()
	if err != nil && !os.IsNotExist(err) {
		t.Logf("CleanupExpiredBackups error (expected if no backups dir): %v", err)
	}
	t.Logf("Cleaned up %d expired backups", count)
}
