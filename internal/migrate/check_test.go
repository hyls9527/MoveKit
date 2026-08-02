package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sys "github.com/hyls9527/MoveKit/internal/syscall"
)

func TestPreCheck_NonNTFS(t *testing.T) {
	// This test verifies PreCheck doesn't panic when called with
	// a source that exists. The test environment may or may not have
	// an NTFS volume, so we just ensure no panics occur.
	//
	// On a typical Windows dev machine this will either pass (NTFS) or
	// return a meaningful error (non-NTFS or missing drive).
	tmpDir := t.TempDir()

	// Test with a non-existent target drive to exercise the error paths.
	_ = PreCheck(tmpDir, "Z")
	// We don't assert specific errors because the test environment varies.
	// The goal is to verify no panics.

	t.Log("PreCheck completed without panic")
}

func TestDetectSecuritySoftware_NoPanic(t *testing.T) {
	// Call detectSecuritySoftware. It should not panic.
	// May return empty string if no known AV is running.
	avName := detectSecuritySoftware()
	// Verify return type is string (no panic means this succeeded).
	if avName != "" {
		t.Logf("detected security software: %s", avName)
	} else {
		t.Log("no known security software detected")
	}
}

func TestPreCheck_SecuritySoftware(t *testing.T) {
	// Call PreCheck with a TempDir source and valid drive.
	// Verify the function completes without panic.
	tmpDir := t.TempDir()
	err := PreCheck(tmpDir, "Z")
	// On non-existent Z: drive, PreCheck will return an error.
	// The key assertion: detectSecuritySoftware was called and its result
	// may surface via the Warning field in the returned MoveKitError.
	if err != nil {
		t.Logf("PreCheck returned error (expected on missing drive): code=%s", err.Code)
		if err.Warning != "" {
			t.Logf("security software warning: %s", err.Warning)
		}
	}
	t.Log("PreCheck completed (with or without security software detection)")
}

func TestMaskPath(t *testing.T) {
	// Verify maskPath doesn't panic and returns a string.
	result := maskPath(`C:\Users\SomeUser\Desktop`)
	if result == "" {
		t.Fatal("maskPath returned empty string")
	}
	t.Logf("maskPath result: %s", result)
}

func TestPreCheck_RebootRequiredValue(t *testing.T) {
	tmpDir := t.TempDir()
	origSys := Sys
	defer func() { Sys = origSys }()

	reg := &fakeRegistry{lmValues: map[string]uint64{}}
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

	// RebootRequired = 0 must not block migration.
	reg.lmValues[`SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired`] = 0
	if err := PreCheck(tmpDir, "D"); err != nil {
		t.Errorf("PreCheck should pass with RebootRequired=0, got: %v", err)
	}

	// RebootRequired = 1 must block migration.
	reg.lmValues[`SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired`] = 1
	err := PreCheck(tmpDir, "D")
	if err == nil {
		t.Fatal("PreCheck should fail with RebootRequired=1")
	}
	if err.Code != ErrCodeWindowsUpdate {
		t.Errorf("code = %q, want WINDOWS_UPDATE", err.Code)
	}
}

func TestPreCheck_BatteryWarning(t *testing.T) {
	tmpDir := t.TempDir()
	origSys := Sys
	defer func() { Sys = origSys }()

	real := &sys.RealImpl{}
	Sys = sys.SyscallLayer{
		Registry: &fakeRegistry{lmValues: map[string]uint64{}},
		Volume:   fakeVolume{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{onBattery: true},
		Process:  fakeProcess{},
	}

	err := PreCheck(tmpDir, "D")
	if err == nil {
		t.Fatal("PreCheck should report battery warning")
	}
	if err.Code != ErrCodeBatteryWarning {
		t.Errorf("code = %q, want BATTERY_WARNING", err.Code)
	}
}

func TestPreCheck_RejectsNonASCIIDriveLetter(t *testing.T) {
	tmpDir := t.TempDir()
	origSys := Sys
	defer func() { Sys = origSys }()

	real := &sys.RealImpl{}
	Sys = sys.SyscallLayer{
		Registry: &fakeRegistry{lmValues: map[string]uint64{}},
		Volume:   fakeVolume{},
		Junction: real,
		Shell:    real,
		File:     real,
		Power:    fakePower{},
		Process:  fakeProcess{},
	}

	err := PreCheck(tmpDir, "中")
	if err == nil {
		t.Fatal("expected error for non-ASCII drive letter")
	}
	if err.Code != ErrCodeFileSystemNotNTFS {
		t.Errorf("code = %q, want FS_NOT_NTFS", err.Code)
	}
}

func TestDirStats(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("12345"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	sub := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("123"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	size, files, err := DirStats(tmpDir)
	if err != nil {
		t.Fatalf("DirStats: %v", err)
	}
	if files != 2 {
		t.Errorf("files = %d, want 2", files)
	}
	if size != 8 {
		t.Errorf("size = %d, want 8", size)
	}
}

func BenchmarkDirStats(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 500; i++ {
		sub := filepath.Join(dir, fmt.Sprintf("sub-%03d", i%20))
		_ = os.MkdirAll(sub, 0755)
		_ = os.WriteFile(filepath.Join(sub, fmt.Sprintf("f%04d.bin", i)), make([]byte, 512), 0644)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := DirStats(dir)
		if err != nil {
			b.Fatal(err)
		}
	}
}
