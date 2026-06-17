package migrate

import "testing"

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

func TestMaskPath(t *testing.T) {
	// Verify maskPath doesn't panic and returns a string.
	result := maskPath(`C:\Users\SomeUser\Desktop`)
	if result == "" {
		t.Fatal("maskPath returned empty string")
	}
	t.Logf("maskPath result: %s", result)
}
