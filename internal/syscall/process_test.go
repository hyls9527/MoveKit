package winsyscall

import (
	"io"
	"strings"
	"syscall"
	"testing"
)

func TestEnsureSingleInstance_AlreadyExists(t *testing.T) {
	// Call EnsureSingleInstance twice with the same mutex name.
	// The second call must return an error (another instance exists).
	suffix := strings.ReplaceAll(t.TempDir()[len(t.TempDir())-8:], "\\", "_")
	name := "MoveKit_TestInstance_" + suffix
	h1, err := EnsureSingleInstance(name)
	if err != nil {
		t.Fatalf("first EnsureSingleInstance failed: %v", err)
	}
	if h1 == syscall.InvalidHandle {
		t.Fatal("first call returned InvalidHandle")
	}
	defer CloseHandle(h1)

	// Second call with the same mutex name MUST return an error.
	h2, err := EnsureSingleInstance(name)
	if err == nil {
		CloseHandle(h2)
		t.Fatal("second EnsureSingleInstance should have returned an error, but didn't")
	}
	if h2 != syscall.InvalidHandle {
		t.Errorf("second call returned a valid handle on error: %v", h2)
	}
	_ = io.Discard
}

func TestCloseHandle(t *testing.T) {
	// Ensure CloseHandle does not panic with a valid handle.
	suffix := strings.ReplaceAll(t.TempDir()[len(t.TempDir())-8:], "\\", "_")
	name := "MoveKit_TestClose_" + suffix
	h, err := EnsureSingleInstance(name)
	if err != nil {
		t.Fatalf("EnsureSingleInstance: %v", err)
	}
	// CloseHandle should not panic and should succeed on a valid handle.
	CloseHandle(h)
}
