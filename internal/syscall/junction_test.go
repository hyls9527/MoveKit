package winsyscall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateJunction(t *testing.T) {
	tmpDir := t.TempDir()

	// Junctions require NTFS. Skip if temp dir is not on an NTFS volume.
	volInfo, err := GetVolumeInfo(filepath.VolumeName(tmpDir))
	if err != nil || volInfo.FileSystem != "NTFS" {
		t.Skipf("junction test requires NTFS volume (got %s)", volInfo.FileSystem)
	}

	// Create a target directory.
	target := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir target: %v", err)
	}

	// Create a file inside target so we can verify it's reachable through the junction.
	if err := os.WriteFile(filepath.Join(target, "hello.txt"), []byte("world"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create a link directory.
	link := filepath.Join(tmpDir, "link")
	if err := os.Mkdir(link, 0o755); err != nil {
		t.Fatalf("Mkdir link: %v", err)
	}

	// Create the junction.
	if err := CreateJunction(link, target); err != nil {
		t.Fatalf("CreateJunction: %v", err)
	}

	// Verify it looks like a junction.
	if !IsJunction(link) {
		t.Fatal("IsJunction returned false after CreateJunction")
	}

	// Verify we can read through the junction.
	data, err := os.ReadFile(filepath.Join(link, "hello.txt"))
	if err != nil {
		t.Fatalf("ReadFile through junction: %v", err)
	}
	if string(data) != "world" {
		t.Fatalf("ReadFile content = %q, want %q", string(data), "world")
	}

	// Delete the junction.
	if err := DeleteJunction(link); err != nil {
		t.Fatalf("DeleteJunction: %v", err)
	}

	// Verify it no longer looks like a junction.
	if IsJunction(link) {
		t.Fatal("IsJunction returned true after DeleteJunction")
	}
}
