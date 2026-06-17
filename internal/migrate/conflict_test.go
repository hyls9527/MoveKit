package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectConflicts(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Create source files.
	os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("world"), 0644)
	os.WriteFile(filepath.Join(srcDir, "unique.txt"), []byte("only in src"), 0644)

	// Create target with some overlapping files.
	os.MkdirAll(filepath.Join(dstDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(dstDir, "file1.txt"), []byte("different content"), 0644)
	os.WriteFile(filepath.Join(dstDir, "subdir", "file2.txt"), []byte("also different"), 0644)
	// unique.txt does NOT exist in dst — no conflict.

	conflicts, err := DetectConflicts(srcDir, dstDir)
	if err != nil {
		t.Fatalf("DetectConflicts: %v", err)
	}

	if len(conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d", len(conflicts))
	}

	// Check that unique.txt is not in the conflicts.
	for _, c := range conflicts {
		if c.RelativePath == "unique.txt" {
			t.Error("unique.txt should not be a conflict")
		}
	}
}

func TestDetectConflicts_NoConflict(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Source has files, target is empty.
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("data"), 0644)
	os.MkdirAll(dstDir, 0755)

	conflicts, err := DetectConflicts(srcDir, dstDir)
	if err != nil {
		t.Fatalf("DetectConflicts: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestDetectConflicts_AllConflict(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Both source and target have the same files.
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("src-a"), 0644)
	os.WriteFile(filepath.Join(dstDir, "a.txt"), []byte("dst-a"), 0644)
	os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("src-b"), 0644)
	os.WriteFile(filepath.Join(dstDir, "b.txt"), []byte("dst-b"), 0644)

	conflicts, err := DetectConflicts(srcDir, dstDir)
	if err != nil {
		t.Fatalf("DetectConflicts: %v", err)
	}

	if len(conflicts) != 2 {
		t.Errorf("expected 2 conflicts, got %d", len(conflicts))
	}
}
