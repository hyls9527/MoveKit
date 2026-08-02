package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyls9527/MoveKit/internal/folders"
	"github.com/hyls9527/MoveKit/internal/migrate"
	sys "github.com/hyls9527/MoveKit/internal/syscall"
)

func TestGetFolders_Structure(t *testing.T) {
	app := NewApp()
	got := app.GetFolders()
	if len(got) != len(folders.AllFolders) {
		t.Errorf("GetFolders returned %d entries, want %d", len(got), len(folders.AllFolders))
	}
	for _, f := range got {
		if f.Name == "" || f.GUID == "" {
			t.Errorf("folder entry missing name/guid: %+v", f)
		}
		if f.Status == "" {
			t.Errorf("folder %s missing status", f.Name)
		}
	}
}

func TestGetDisks_ContainsDrive(t *testing.T) {
	app := NewApp()
	disks := app.GetDisks()
	if len(disks) == 0 {
		t.Fatal("GetDisks returned no drives (CI runner should have C:)")
	}
	for _, d := range disks {
		if d.Drive == "" {
			t.Error("disk entry missing drive letter")
		}
	}
}

func TestEstimateDuration_NoPanic(t *testing.T) {
	app := NewApp()
	est := app.EstimateDuration("D")
	if est.EstimatedText == "" {
		t.Error("EstimateDuration returned empty EstimatedText")
	}
	if est.EstimatedSec < 0 {
		t.Errorf("EstimateDuration returned negative seconds: %d", est.EstimatedSec)
	}
}

func TestCheckOneDriveKFM_NoPanic(t *testing.T) {
	app := NewApp()
	status := app.CheckOneDriveKFM()
	if status.Folders == nil {
		// nil slice is acceptable; just ensure no panic and defaults are sane
		if status.IsActive {
			t.Error("IsActive=true with nil Folders")
		}
	}
}

func TestStartupIntegrityCheck_EmptyBackups(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	t.Setenv("APPDATA", t.TempDir())
	defer os.Setenv("APPDATA", origAppData)

	app := NewApp()
	issues := app.StartupIntegrityCheck()
	if len(issues) != 0 {
		t.Errorf("expected 0 issues with no backups, got %d: %v", len(issues), issues)
	}
}

func TestStartupIntegrityCheck_CorruptBackup(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppData)

	backupDir := filepath.Join(tmp, "MoveKit", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "Broken.json"), []byte("{broken"), 0644); err != nil {
		t.Fatalf("write broken backup: %v", err)
	}

	app := NewApp()
	issues := app.StartupIntegrityCheck()
	found := false
	for _, issue := range issues {
		if issue.Type == "backup_corrupt" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a backup_corrupt issue, got: %v", issues)
	}
}

func TestStartupIntegrityCheck_CompleteBackupJunctionMissing(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppData)

	backupDir := filepath.Join(tmp, "MoveKit", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := migrate.BackupRecord{
		Version:      1,
		RegValueName: "Desktop",
		OriginalPath: filepath.Join(tmp, "no-such-junction"),
		NewPath:      filepath.Join(tmp, "elsewhere"),
		Status:       "complete",
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "Desktop.json"), data, 0644); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	app := NewApp()
	issues := app.StartupIntegrityCheck()
	found := false
	for _, issue := range issues {
		if issue.Type == "junction_broken" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a junction_broken issue, got: %v", issues)
	}
}

func TestStartupIntegrityCheck_JunctionOKAndMismatch(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppData)

	backupDir := filepath.Join(tmp, "MoveKit", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Real junction: link -> target.
	target := filepath.Join(tmp, "real-target")
	link := filepath.Join(tmp, "real-junction")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.MkdirAll(link, 0755); err != nil {
		t.Fatalf("mkdir link: %v", err)
	}
	if err := sys.CreateJunction(link, target); err != nil {
		t.Skipf("cannot create junction in this environment: %v", err)
	}
	defer sys.DeleteJunction(link)

	writeRecord := func(name, orig, newPath string) {
		rec := migrate.BackupRecord{
			Version:      1,
			RegValueName: name,
			OriginalPath: orig,
			NewPath:      newPath,
			Status:       "complete",
		}
		data, err := json.Marshal(rec)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if err := os.WriteFile(filepath.Join(backupDir, name+".json"), data, 0644); err != nil {
			t.Fatalf("write backup: %v", err)
		}
	}

	// Matching junction: no issue.
	writeRecord("Desktop", link, target)
	// Mismatched junction: junction_mismatch issue.
	writeRecord("Documents", link, filepath.Join(tmp, "wrong-target"))

	app := NewApp()
	issues := app.StartupIntegrityCheck()
	var mismatch, broken int
	for _, issue := range issues {
		switch issue.Type {
		case "junction_mismatch":
			mismatch++
		case "junction_broken":
			broken++
		}
	}
	if mismatch != 1 {
		t.Errorf("expected 1 junction_mismatch, got %d (issues: %v)", mismatch, issues)
	}
	if broken != 0 {
		t.Errorf("expected 0 junction_broken, got %d", broken)
	}
}

func TestMoveKitError_ErrorMethod(t *testing.T) {
	err := &migrate.MoveKitError{Code: "X", Message: "hello", Detail: "d", Action: "a"}
	if err.Error() != "hello" {
		t.Errorf("Error() = %q, want %q", err.Error(), "hello")
	}
}

func TestRollback_RejectsUnknownFolder(t *testing.T) {
	app := NewApp()
	for _, name := range []string{"..\\..\\evil", "Desktop.json", "", "{GUID}"} {
		err := app.Rollback(name)
		if err == nil {
			t.Errorf("Rollback(%q) should be rejected", name)
			continue
		}
		mkErr, ok := err.(*migrate.MoveKitError)
		if !ok {
			t.Errorf("Rollback(%q) error type = %T, want *migrate.MoveKitError", name, err)
			continue
		}
		if mkErr.Code != "INVALID_FOLDER" {
			t.Errorf("Rollback(%q) code = %q, want INVALID_FOLDER", name, mkErr.Code)
		}
	}
}

func TestSetFirstRunFlag_ConfigDirBlocked(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	defer os.Setenv("APPDATA", origAppData)

	// Make %APPDATA%\MoveKit a regular file so MkdirAll fails.
	if err := os.WriteFile(filepath.Join(tmp, "MoveKit"), []byte("block"), 0644); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}

	app := NewApp()
	if err := app.SetFirstRunFlag(false); err == nil {
		t.Error("SetFirstRunFlag should fail when the config directory is blocked")
	}
}

func TestDetectConflicts_EmptyDirs(t *testing.T) {
	app := NewApp()
	src := t.TempDir()
	dst := t.TempDir()
	conflicts, err := app.DetectConflicts(src, dst)
	if err != nil {
		t.Fatalf("DetectConflicts on empty dirs: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts on empty dirs, got %d", len(conflicts))
	}
}

func TestDetectCloudSync_NoPanic(t *testing.T) {
	app := NewApp()
	providers := app.DetectCloudSync()
	if providers == nil {
		t.Error("DetectCloudSync returned nil slice")
	}
}

func TestCheckUpdate_NoUpdate(t *testing.T) {
	app := NewApp()
	info := app.CheckUpdate()
	if info.HasUpdate {
		t.Errorf("CheckUpdate should report no update, got %+v", info)
	}
	if info.NewVersion == "" {
		t.Error("CheckUpdate returned empty NewVersion")
	}
}

func TestCancelMigration_FlagRoundTrip(t *testing.T) {
	migrate.ResetCancellation()
	defer migrate.ResetCancellation()

	app := NewApp()
	app.CancelMigration()
	if !migrate.IsCancelled() {
		t.Error("CancelMigration should set the cancellation flag")
	}
	migrate.ResetCancellation()
	if migrate.IsCancelled() {
		t.Error("ResetCancellation should clear the flag")
	}
}

func TestCachedDirStats_RoundTrip(t *testing.T) {
	app := NewApp()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	ds, err := app.cachedDirStats(dir)
	if err != nil {
		t.Fatalf("cachedDirStats: %v", err)
	}
	if ds.size != 5 {
		t.Errorf("size = %d, want 5", ds.size)
	}
	if ds.files != 1 {
		t.Errorf("files = %d, want 1", ds.files)
	}

	// Second call should hit the cache and return identical values.
	ds2, err := app.cachedDirStats(dir)
	if err != nil {
		t.Fatalf("cachedDirStats (cached): %v", err)
	}
	if ds2 != ds {
		t.Errorf("cached result mismatch: %+v vs %+v", ds, ds2)
	}

	app.clearSizeCache()
	if _, ok := app.statsCache.Load(dir); ok {
		t.Error("clearSizeCache should drop cached entries")
	}
}
