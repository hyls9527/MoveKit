package winsyscall

import (
	"io"
	"syscall"
	"testing"
)

// isErrorAlreadyExists is a helper that uses comma-ok to safely check
// if an error is syscall.Errno with value ERROR_ALREADY_EXISTS (183).
func isErrorAlreadyExists(e error) bool {
	if errno, ok := e.(syscall.Errno); ok && errno == 183 {
		return true
	}
	return false
}

func TestEnsureSingleInstance_NoPanicOnNilError(t *testing.T) {
	// Table-driven test: verify isErrorAlreadyExists uses comma-ok correctly.
	// The real EnsureSingleInstance passes e from procCreateMutexW.Call()
	// which may be nil (on success), syscall.Errno, or an unexpected type.
	// The comma-ok pattern prevents panic in all cases.
	tests := []struct {
		name  string
		input error
		want  bool
	}{
		{name: "nil error", input: nil, want: false},
		{name: "io.EOF (wrong type)", input: io.EOF, want: false},
		{name: "syscall.Errno(183) ALREADY_EXISTS", input: syscall.Errno(183), want: true},
		{name: "syscall.Errno(0) success", input: syscall.Errno(0), want: false},
		{name: "syscall.Errno(5) access denied", input: syscall.Errno(5), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isErrorAlreadyExists(tt.input)
			if got != tt.want {
				t.Errorf("isErrorAlreadyExists(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureSingleInstance_AlreadyExists(t *testing.T) {
	// Call EnsureSingleInstance twice with the same mutex name.
	// The second call must return an error (another instance exists).
	name := "MoveKit_TestInstance_" + t.TempDir()[len(t.TempDir())-8:]
	h1, err := EnsureSingleInstance(name)
	if err != nil {
		t.Fatalf("first EnsureSingleInstance failed: %v", err)
	}
	if h1 == syscall.InvalidHandle {
		t.Fatal("first call returned InvalidHandle")
	}
	defer CloseHandle(h1)

	// Second call with same name must detect already exists.
	h2, err2 := EnsureSingleInstance(name)
	if err2 == nil {
		CloseHandle(h2)
		t.Fatal("second EnsureSingleInstance should have returned error for existing mutex, got nil")
	}
	if h2 != syscall.InvalidHandle {
		CloseHandle(h2)
		t.Error("second call should return InvalidHandle on error")
	}
}

func TestCloseHandle(t *testing.T) {
	// Ensure CloseHandle does not panic with a valid handle.
	name := "MoveKit_TestClose_" + t.TempDir()[len(t.TempDir())-8:]
	h, err := EnsureSingleInstance(name)
	if err != nil {
		t.Fatalf("EnsureSingleInstance: %v", err)
	}
	// This must not panic.
	CloseHandle(h)
	t.Log("CloseHandle completed without panic")
}
