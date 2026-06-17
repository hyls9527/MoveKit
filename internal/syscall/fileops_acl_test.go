package winsyscall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFileWithACL(t *testing.T) {
	// Create a temporary source file.
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("hello world")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	// Copy with ACL to a destination in the same temp dir.
	dstFile := filepath.Join(tmpDir, "dest.txt")
	if err := CopyFileWithACL(srcFile, dstFile, nil); err != nil {
		t.Fatalf("CopyFileWithACL: %v", err)
	}

	// Verify destination file exists and content matches.
	got, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}
