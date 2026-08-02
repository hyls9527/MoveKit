package winsyscall

import (
	"os"
	"path/filepath"
	"strings"
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

func TestResolveTarget(t *testing.T) {
	tmpDir := t.TempDir()

	volInfo, err := GetVolumeInfo(filepath.VolumeName(tmpDir))
	if err != nil || volInfo.FileSystem != "NTFS" {
		t.Skipf("junction test requires NTFS volume (got %s)", volInfo.FileSystem)
	}

	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("Mkdir target: %v", err)
	}
	if err := os.Mkdir(link, 0o755); err != nil {
		t.Fatalf("Mkdir link: %v", err)
	}
	if err := CreateJunction(link, target); err != nil {
		t.Fatalf("CreateJunction: %v", err)
	}

	resolved, err := ResolveTarget(link)
	if err != nil {
		t.Fatalf("ResolveTarget: %v", err)
	}
	absTarget, _ := filepath.Abs(target)
	if !strings.EqualFold(resolved, absTarget) {
		t.Errorf("ResolveTarget = %q, want %q", resolved, absTarget)
	}

	if _, err := ResolveTarget(filepath.Join(tmpDir, "not-a-junction")); err == nil {
		t.Error("ResolveTarget should fail for a non-junction path")
	}
}

func TestParseMountPointTarget(t *testing.T) {
	buf, err := buildReparseBuffer(`\??\C:\Users\Test\Desktop`, `C:\Users\Test\Desktop`)
	if err != nil {
		t.Fatalf("buildReparseBuffer: %v", err)
	}

	got, err := parseMountPointTarget(buf)
	if err != nil {
		t.Fatalf("parseMountPointTarget(valid): %v", err)
	}
	if !strings.EqualFold(got, `C:\Users\Test\Desktop`) {
		t.Errorf("parsed target = %q, want C:\\Users\\Test\\Desktop", got)
	}

	if _, err := parseMountPointTarget(buf[:10]); err == nil {
		t.Error("truncated buffer should fail")
	}

	badTag := append([]byte(nil), buf...)
	badTag[0] = 0xFF // wrong reparse tag
	if _, err := parseMountPointTarget(badTag); err == nil {
		t.Error("wrong reparse tag should fail")
	}

	overflow := append([]byte(nil), buf...)
	overflow[8] = 0xFF // SubstituteNameOffset out of range
	overflow[9] = 0xFF
	overflow[10] = 0xFF // SubstituteNameLength out of range
	overflow[11] = 0xFF
	if _, err := parseMountPointTarget(overflow); err == nil {
		t.Error("out-of-range substitute name should fail")
	}
}

func BenchmarkParseMountPointTarget(b *testing.B) {
	buf, err := buildReparseBuffer(`\??\C:\Users\Test\Desktop`, `C:\Users\Test\Desktop`)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseMountPointTarget(buf); err != nil {
			b.Fatal(err)
		}
	}
}
