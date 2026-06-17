package migrate

import (
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
