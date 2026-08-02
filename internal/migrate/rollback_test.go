package migrate

import (
	"os"
	"path/filepath"
	"testing"

	sys "github.com/hyls9527/MoveKit/internal/syscall"
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

func TestRollback_SameVolumeSuccess(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)

	originalPath := filepath.Join(tmp, "original")
	newPath := filepath.Join(tmp, "new")
	if err := os.MkdirAll(newPath, 0755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newPath, "a.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "Desktop",
		OriginalPath:  originalPath,
		NewPath:       newPath,
		RegistryValue: "Desktop",
		Status:        "complete",
	}
	if err := SaveBackup(rec); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	origSys := Sys
	defer func() { Sys = origSys }()
	real := &sys.RealImpl{}
	reg := &fakeRegistry{lmValues: map[string]uint64{}}
	Sys = sys.SyscallLayer{
		Registry: reg,
		Volume:   sameVolumeChecker{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{},
		Process:  fakeProcess{},
	}

	err := Rollback("Desktop")
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if _, err := os.Stat(filepath.Join(originalPath, "a.txt")); err != nil {
		t.Errorf("file not restored to original path: %v", err)
	}
	if len(reg.userWrites) != 1 || reg.userWrites[0] != originalPath {
		t.Errorf("User Shell Folders not restored: %v", reg.userWrites)
	}
	loaded, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if loaded.Status != "rolled_back" {
		t.Errorf("status = %q, want rolled_back", loaded.Status)
	}
}

func TestRollback_RegistryFailureKeepsFilesAtOriginal(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)

	originalPath := filepath.Join(tmp, "original")
	newPath := filepath.Join(tmp, "new")
	if err := os.MkdirAll(newPath, 0755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newPath, "a.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "Desktop",
		OriginalPath:  originalPath,
		NewPath:       newPath,
		RegistryValue: "Desktop",
		Status:        "complete",
	}
	if err := SaveBackup(rec); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	origSys := Sys
	defer func() { Sys = origSys }()
	real := &sys.RealImpl{}
	reg := &fakeRegistry{lmValues: map[string]uint64{}, writeErr: true}
	Sys = sys.SyscallLayer{
		Registry: reg,
		Volume:   sameVolumeChecker{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{},
		Process:  fakeProcess{},
	}

	err := Rollback("Desktop")
	if err == nil {
		t.Fatal("expected registry failure error")
	}
	// Files are already back at the original path; only the registry write
	// failed. The backup record must remain untouched so a retry can finish.
	if _, err := os.Stat(filepath.Join(originalPath, "a.txt")); err != nil {
		t.Errorf("file should be at original path after failed registry restore: %v", err)
	}
	loaded, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if loaded.Status != "complete" {
		t.Errorf("status = %q, want complete (retryable)", loaded.Status)
	}
}

func TestRollback_IdempotentRetryWhenFilesAlreadyRestored(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)

	originalPath := filepath.Join(tmp, "original")
	missingNewPath := filepath.Join(tmp, "new-missing")
	if err := os.MkdirAll(originalPath, 0755); err != nil {
		t.Fatalf("mkdir original: %v", err)
	}
	if err := os.WriteFile(filepath.Join(originalPath, "a.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	rec := BackupRecord{
		Version:       1,
		RegValueName:  "Desktop",
		OriginalPath:  originalPath,
		NewPath:       missingNewPath,
		RegistryValue: "Desktop",
		Status:        "complete",
	}
	if err := SaveBackup(rec); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	origSys := Sys
	defer func() { Sys = origSys }()
	real := &sys.RealImpl{}
	reg := &fakeRegistry{lmValues: map[string]uint64{}}
	Sys = sys.SyscallLayer{
		Registry: reg,
		Volume:   sameVolumeChecker{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{},
		Process:  fakeProcess{},
	}

	err := Rollback("Desktop")
	if err != nil {
		t.Fatalf("Rollback retry should succeed when files are already back: %v", err)
	}
	if _, err := os.Stat(filepath.Join(originalPath, "a.txt")); err != nil {
		t.Errorf("original file missing: %v", err)
	}
	if len(reg.userWrites) != 1 || reg.userWrites[0] != originalPath {
		t.Errorf("registry not restored: %v", reg.userWrites)
	}
	loaded, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if loaded.Status != "rolled_back" {
		t.Errorf("status = %q, want rolled_back", loaded.Status)
	}
}

// sameVolumeChecker makes Rollback use os.Rename instead of copy+delete.
type sameVolumeChecker struct{}

func (sameVolumeChecker) GetVolumeInfo(drive string) (sys.VolumeInfo, error) {
	return sys.VolumeInfo{FileSystem: "NTFS", IsFixedDrive: true}, nil
}

func (sameVolumeChecker) GetDiskFreeSpace(drive string) (int64, int64, int64, error) {
	return 1 << 40, 2 << 40, 1 << 40, nil
}

func (sameVolumeChecker) IsCrossVolume(src, dst string) bool { return false }
