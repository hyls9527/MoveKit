package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCancelMigration(t *testing.T) {
	// Ensure clean state.
	ResetCancellation()

	if IsCancelled() {
		t.Fatal("expected IsCancelled() == false before cancel")
	}

	CancelMigration()

	if !IsCancelled() {
		t.Fatal("expected IsCancelled() == true after CancelMigration()")
	}

	// Reset and verify.
	ResetCancellation()

	if IsCancelled() {
		t.Fatal("expected IsCancelled() == false after ResetCancellation()")
	}
}

func TestCancelMigration_Concurrent(t *testing.T) {
	ResetCancellation()

	done := make(chan struct{})
	go func() {
		// Simulate a goroutine checking cancellation.
		for !IsCancelled() {
			// busy wait
		}
		close(done)
	}()

	CancelMigration()

	<-done // should return quickly since cancel was set
}

func TestMoveFilesSameVolume_Success(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("world"), 0644)

	phases := make([]string, 0)
	onProgress := func(pi ProgressInfo) {
		phases = append(phases, pi.Phase)
	}

	err := moveFilesSameVolume(srcDir, dstDir, onProgress)
	if err != nil {
		t.Fatalf("moveFilesSameVolume failed: %v", err)
	}

	// Source should no longer exist (moved).
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Errorf("source %s still exists after same-volume move", srcDir)
	}
	// Target should exist.
	if _, err := os.Stat(dstDir); err != nil {
		t.Errorf("target %s does not exist after same-volume move: %v", dstDir, err)
	}
	// Files should be at target.
	for _, fn := range []string{"a.txt", "b.txt"} {
		if _, err := os.Stat(filepath.Join(dstDir, fn)); err != nil {
			t.Errorf("file %s missing from target: %v", fn, err)
		}
	}
	// Verify progress callback was invoked.
	if len(phases) < 2 {
		t.Errorf("expected at least 2 progress phases, got %d: %v", len(phases), phases)
	}
}

func TestMoveFilesSameVolume_TargetConflict(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)
	// Put a file in dst so os.Rename fails (non-empty target).
	os.WriteFile(filepath.Join(dstDir, "locked.txt"), []byte("locked"), 0644)

	err := moveFilesSameVolume(srcDir, dstDir, nil)
	if err == nil {
		t.Fatal("expected error when target is non-empty, got nil")
	}
	mkErr, ok := err.(*MoveKitError)
	if !ok {
		t.Fatalf("expected *MoveKitError, got %T: %v", err, err)
	}
	if mkErr.Code != ErrCodeFileLocked {
		t.Errorf("expected code %q, got %q", ErrCodeFileLocked, mkErr.Code)
	}
	// Source should still exist (rename is atomic).
	if _, err := os.Stat(srcDir); err != nil {
		t.Errorf("source should still exist after failed rename: %v", err)
	}
}
