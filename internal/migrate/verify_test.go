package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyIntegrity_FileCountMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "verify_target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Expect 100 files, but only 1 exists — should return error.
	err := VerifyIntegrity(100, 9999, target)
	if err == nil {
		t.Error("VerifyIntegrity should return error when file count mismatches")
	}
	t.Logf("VerifyIntegrity mismatch error: %v", err)
}

func TestVerifyIntegrity_Match(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "verify_match")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "a.txt"), []byte("hi"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := VerifyIntegrity(1, 2, target)
	if err != nil {
		t.Errorf("VerifyIntegrity should pass with matching count: %v", err)
	}
}

func TestVerifyIntegrity_NonexistentTarget(t *testing.T) {
	err := VerifyIntegrity(0, 0, `C:\ThisPathDoesNotExist_12345_verify`)
	if err == nil {
		t.Error("VerifyIntegrity should return error for nonexistent path")
	}
}
