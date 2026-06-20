package winsyscall

import (
	"path/filepath"
	"testing"
)

func TestNotifyFolderRenamed_NoPanic(t *testing.T) {
	// Smoke test: NotifyFolderRenamed with valid ASCII paths must not panic.
	tmpDir := t.TempDir()
	oldPath := filepath.Join(tmpDir, "old")
	newPath := filepath.Join(tmpDir, "new")
	// The shell notification on test machines may or may not take effect,
	// but the function must not crash.
	NotifyFolderRenamed(oldPath, newPath)
	t.Log("NotifyFolderRenamed completed without panic")
}

func TestNotifyDirUpdated_NoPanic(t *testing.T) {
	// Smoke test: NotifyDirUpdated with a valid path must not panic.
	tmpDir := t.TempDir()
	NotifyDirUpdated(tmpDir)
	t.Log("NotifyDirUpdated completed without panic")
}

func TestNotifyFolderRenamed_InvalidUTF16(t *testing.T) {
	// NotifyFolderRenamed with an invalid UTF-16 path must not panic.
	// It should log a warning and return gracefully.
	invalidPath := string([]rune{0xD800}) // isolated surrogate = invalid UTF-16
	NotifyFolderRenamed(invalidPath, invalidPath)
	t.Log("NotifyFolderRenamed with invalid UTF-16 completed without panic")
}
