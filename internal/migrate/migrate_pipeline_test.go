package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sys "github.com/hyls9527/MoveKit/internal/syscall"
)

// ---------------------------------------------------------------------------
// Fake syscall layer
// ---------------------------------------------------------------------------

type fakeRegistry struct {
	srcPath     string
	writeErr    bool
	userWrites  []string
	shellWrites []string
	lmValues    map[string]uint64
}

func (f *fakeRegistry) ReadExpandString(keyPath, valueName string) (string, error) {
	if keyPath == sys.UserShellFoldersKey && valueName == "Desktop" {
		return f.srcPath, nil
	}
	return "", fmt.Errorf("fake: value %q not found", valueName)
}

func (f *fakeRegistry) WriteExpandString(keyPath, valueName, value string) error {
	if f.writeErr {
		return fmt.Errorf("fake: write failed")
	}
	f.userWrites = append(f.userWrites, value)
	return nil
}

func (f *fakeRegistry) WriteString(keyPath, valueName, value string) error {
	if f.writeErr {
		return fmt.Errorf("fake: write failed")
	}
	if keyPath == sys.ShellFoldersKey {
		f.shellWrites = append(f.shellWrites, value)
	} else {
		f.userWrites = append(f.userWrites, value)
	}
	return nil
}

func (f *fakeRegistry) ReadInteger(keyPath, valueName string) (uint64, error) {
	if v, ok := f.lmValues[keyPath+"\\"+valueName]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("fake: value not found")
}

func (f *fakeRegistry) ReadIntegerLM(keyPath, valueName string) (uint64, error) {
	if v, ok := f.lmValues[keyPath+"\\"+valueName]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("fake: value not found")
}

type fakeVolume struct{}

func (fakeVolume) GetVolumeInfo(drive string) (sys.VolumeInfo, error) {
	return sys.VolumeInfo{FileSystem: "NTFS", IsFixedDrive: true}, nil
}

func (fakeVolume) GetDiskFreeSpace(drive string) (int64, int64, int64, error) {
	return 1 << 40, 2 << 40, 1 << 40, nil
}

func (fakeVolume) IsCrossVolume(src, dst string) bool { return true }

type fakePower struct{ onBattery bool }

func (f fakePower) IsOnBattery() bool { return f.onBattery }

type fakeProcess struct{}

func (fakeProcess) IsProcessRunning(name string) bool { return false }

// fakeFile simulates CopyFileWithACL. It can fail on a specific basename and
// can trigger cancellation after N copies.
type fakeFile struct {
	failOn      string
	cancelAfter int
	copies      int
}

func (f *fakeFile) CopyFileWithACL(src, dst string, callback sys.CopyProgressCallback) error {
	f.copies++
	base := filepath.Base(src)
	if f.failOn != "" && base == f.failOn {
		return fmt.Errorf("simulated copy failure for %s", base)
	}
	if f.cancelAfter > 0 && f.copies >= f.cancelAfter {
		CancelMigration()
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := out.ReadFrom(in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// pipelineHarness sets up a real source directory, a real target root, a fake
// registry/volume/power layer, and restores the package-level Sys on cleanup.
type pipelineHarness struct {
	srcDir     string
	targetRoot string
	reg        *fakeRegistry
}

func newPipelineHarness(t *testing.T, files ...string) *pipelineHarness {
	t.Helper()
	t.Setenv("APPDATA", t.TempDir())

	h := &pipelineHarness{
		srcDir:     filepath.Join(t.TempDir(), "Desktop"),
		targetRoot: t.TempDir(),
		reg:        &fakeRegistry{lmValues: map[string]uint64{}},
	}
	if err := os.MkdirAll(h.srcDir, 0755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(h.srcDir, name), []byte("data-"+name), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	origSys := Sys
	origRoot := testTargetRoot
	t.Cleanup(func() {
		Sys = origSys
		testTargetRoot = origRoot
	})

	real := &sys.RealImpl{}
	h.reg.srcPath = h.srcDir
	Sys = sys.SyscallLayer{
		Registry: h.reg,
		Volume:   fakeVolume{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{},
		Process:  fakeProcess{},
	}
	testTargetRoot = h.targetRoot
	return h
}

func (h *pipelineHarness) targetDir() string {
	return filepath.Join(h.targetRoot, "Desktop")
}

func (h *pipelineHarness) assertFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s at %s: %v", name, dir, err)
		}
	}
}

func TestMigrate_Success(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt", "b.txt")

	err := Migrate("Desktop", "D", nil, ConflictSkip)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	h.assertFiles(t, h.targetDir(), "a.txt", "b.txt")
	if !Sys.Junction.IsJunction(h.srcDir) {
		t.Error("source path should be a junction after successful migration")
	}
	if len(h.reg.userWrites) == 0 || h.reg.userWrites[0] != `D:\Desktop` {
		t.Errorf("User Shell Folders write missing or wrong: %v", h.reg.userWrites)
	}
	if len(h.reg.shellWrites) != 1 {
		t.Errorf("Shell Folders write missing: %v", h.reg.shellWrites)
	}

	rec, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if rec.Status != "complete" {
		t.Errorf("backup status = %q, want complete", rec.Status)
	}
}

func TestMigrate_RegistryFailureRestoresState(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt", "b.txt")
	h.reg.writeErr = true

	err := Migrate("Desktop", "D", nil, ConflictSkip)
	if err == nil {
		t.Fatal("expected error when registry write fails")
	}

	h.assertFiles(t, h.srcDir, "a.txt", "b.txt")
	if Sys.Junction.IsJunction(h.srcDir) {
		t.Error("junction should have been removed during restore")
	}
	if _, err := os.Stat(h.targetDir()); !os.IsNotExist(err) {
		t.Errorf("target dir should be gone after restore, stat err: %v", err)
	}

	rec, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if rec.Status != "rolled_back" {
		t.Errorf("backup status = %q, want rolled_back", rec.Status)
	}
}

func TestMigrate_PartialCopyFailureRestoresState(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt", "b.txt")
	fileOps := &fakeFile{failOn: "b.txt"}
	real := &sys.RealImpl{}
	Sys.File = fileOps
	Sys.Junction = real
	Sys.Shell = real

	err := Migrate("Desktop", "D", nil, ConflictSkip)
	if err == nil {
		t.Fatal("expected error when copy fails midway")
	}

	h.assertFiles(t, h.srcDir, "a.txt", "b.txt")
	if _, err := os.Stat(h.targetDir()); !os.IsNotExist(err) {
		t.Errorf("target dir should be gone after restore, stat err: %v", err)
	}
	if Sys.Junction.IsJunction(h.srcDir) {
		t.Error("source should not be a junction after failed copy")
	}

	rec, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if rec.Status != "rolled_back" {
		t.Errorf("backup status = %q, want rolled_back", rec.Status)
	}
}

func TestMigrate_CancelRestoresState(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt", "b.txt")
	fileOps := &fakeFile{cancelAfter: 1}
	real := &sys.RealImpl{}
	Sys.File = fileOps
	Sys.Junction = real
	Sys.Shell = real

	err := Migrate("Desktop", "D", nil, ConflictSkip)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	mkErr, ok := err.(*MoveKitError)
	if !ok || mkErr.Code != ErrCodeCancelled {
		t.Fatalf("expected CANCELLED MoveKitError, got %#v", err)
	}

	h.assertFiles(t, h.srcDir, "a.txt", "b.txt")
	if _, err := os.Stat(h.targetDir()); !os.IsNotExist(err) {
		t.Errorf("target dir should be gone after cancel restore, stat err: %v", err)
	}
	if Sys.Junction.IsJunction(h.srcDir) {
		t.Error("source should not be a junction after cancelled migration")
	}

	rec, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if rec.Status != "rolled_back" {
		t.Errorf("backup status = %q, want rolled_back", rec.Status)
	}
	ResetCancellation()
}

func TestMigrate_ContinuesOnBatteryWarning(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt")
	Sys.Power = fakePower{onBattery: true}

	err := Migrate("Desktop", "D", nil, ConflictSkip)
	if err != nil {
		t.Fatalf("Migrate should continue on battery warning, got: %v", err)
	}
	h.assertFiles(t, h.targetDir(), "a.txt")
}

func TestMigrate_InvalidDriveLetter(t *testing.T) {
	_ = newPipelineHarness(t, "a.txt")
	err := Migrate("Desktop", "中", nil, ConflictSkip)
	if err == nil {
		t.Fatal("expected error for invalid drive letter")
	}
	mkErr, ok := err.(*MoveKitError)
	if !ok {
		t.Fatalf("expected *MoveKitError, got %T", err)
	}
	if mkErr.Code != ErrCodePermissionDenied {
		t.Errorf("code = %q, want PERMISSION_DENIED", mkErr.Code)
	}
}

func TestMigrate_AcceptsLowercaseDriveLetter(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt")

	err := Migrate("Desktop", "d", nil, ConflictSkip)
	if err != nil {
		t.Fatalf("Migrate with lowercase drive letter: %v", err)
	}
	h.assertFiles(t, h.targetDir(), "a.txt")
	if len(h.reg.userWrites) == 0 || h.reg.userWrites[0] != `D:\Desktop` {
		t.Errorf("registry value should be normalized to uppercase D:, got %v", h.reg.userWrites)
	}
}

func TestMigrate_NoOpPreservesExistingBackup(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt")
	// Make the computed target equal the source directory so the no-op branch
	// is hit, then point the target root back to the harness value.
	h.targetRoot = filepath.Dir(h.srcDir)
	testTargetRoot = h.targetRoot

	original := BackupRecord{
		Version:       1,
		RegValueName:  "Desktop",
		OriginalPath:  `C:\Users\Test\Desktop`,
		NewPath:       `D:\Desktop`,
		RegistryValue: "Desktop",
		Status:        "complete",
	}
	if err := SaveBackup(original); err != nil {
		t.Fatalf("SaveBackup: %v", err)
	}

	if err := Migrate("Desktop", "D", nil, ConflictSkip); err != nil {
		t.Fatalf("no-op migration should succeed, got: %v", err)
	}

	rec, err := LoadBackup("Desktop")
	if err != nil {
		t.Fatalf("LoadBackup: %v", err)
	}
	if rec.OriginalPath != original.OriginalPath || rec.NewPath != original.NewPath {
		t.Errorf("no-op must not overwrite the rollback record: got %+v", rec)
	}
}

func TestMigrate_NoOpWithoutBackupCreatesNone(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t, "a.txt")
	h.targetRoot = filepath.Dir(h.srcDir)
	testTargetRoot = h.targetRoot

	if err := Migrate("Desktop", "D", nil, ConflictSkip); err != nil {
		t.Fatalf("no-op migration should succeed, got: %v", err)
	}
	if _, err := LoadBackup("Desktop"); err == nil {
		t.Error("no-op must not create a self-referencing backup record")
	}
}

func TestRestoreAfterFailedMove_MergeFailurePreservesTarget(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t)

	target := h.targetDir()
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "a.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	// Replace the source directory with a file so mergeDirBack cannot
	// recreate it and must fail partway.
	if err := os.Remove(h.srcDir); err != nil {
		t.Fatalf("remove source dir: %v", err)
	}
	if err := os.WriteFile(h.srcDir, []byte("not-a-dir"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	err := restoreAfterFailedMove(h.srcDir, h.srcDir, target)
	if err == nil {
		t.Fatal("expected restore to fail when the source path is a file")
	}
	if _, err := os.Stat(filepath.Join(target, "a.txt")); err != nil {
		t.Errorf("target files must be preserved when restore fails: %v", err)
	}
}

func BenchmarkMigratePipelineCrossVolume(b *testing.B) {
	b.ReportAllocs()
	b.Setenv("APPDATA", b.TempDir())

	srcDir := filepath.Join(b.TempDir(), "Desktop")
	targetRoot := b.TempDir()
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		if err := os.WriteFile(filepath.Join(srcDir, fmt.Sprintf("f%04d.bin", i)), make([]byte, 4096), 0644); err != nil {
			b.Fatal(err)
		}
	}

	origSys := Sys
	origRoot := testTargetRoot
	defer func() {
		Sys = origSys
		testTargetRoot = origRoot
	}()
	reg := &fakeRegistry{srcPath: srcDir, lmValues: map[string]uint64{}}
	real := &sys.RealImpl{}
	Sys = sys.SyscallLayer{
		Registry: reg,
		Volume:   fakeVolume{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{},
		Process:  fakeProcess{},
	}
	testTargetRoot = targetRoot

	ResetCancellation()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Migrate("Desktop", "D", nil, ConflictSkip); err != nil {
			b.Fatal(err)
		}
	}
}

func TestMigrate_AlreadyMigratedSameTargetIsNoOp(t *testing.T) {
	ResetCancellation()
	h := newPipelineHarness(t) // source dir exists and is empty

	// Set up the "already migrated" state: the real folder lives at the
	// target, and the source path is a junction pointing at it.
	if err := os.MkdirAll(h.targetDir(), 0755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(h.targetDir(), "a.txt"), []byte("data-a"), 0644); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	if err := Sys.Junction.CreateJunction(h.srcDir, h.targetDir()); err != nil {
		t.Fatalf("create junction: %v", err)
	}

	err := Migrate("Desktop", "D", nil, ConflictSkip)
	if err != nil {
		if mkErr, ok := err.(*MoveKitError); ok {
			t.Logf("MoveKitError detail: %s", mkErr.Detail)
		}
		t.Fatalf("Migrate should be a no-op for an already-migrated folder, got: %v", err)
	}
	if !Sys.Junction.IsJunction(h.srcDir) {
		t.Error("existing junction should be preserved")
	}
	h.assertFiles(t, h.targetDir(), "a.txt")
	if len(h.reg.userWrites) != 0 || len(h.reg.shellWrites) != 0 {
		t.Errorf("no registry writes expected for no-op migration, got user=%v shell=%v", h.reg.userWrites, h.reg.shellWrites)
	}

	if _, err := LoadBackup("Desktop"); err == nil {
		t.Error("no-op must not create a self-referencing backup record")
	}
}
