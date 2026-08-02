package migrate

import (
	"os"
	"path/filepath"
	"testing"

	winsyscall "github.com/hyls9527/MoveKit/internal/syscall"
)

type stubRegistry struct {
	values map[string]string
}

func (s *stubRegistry) ReadExpandString(_, valueName string) (string, error) {
	if v, ok := s.values[valueName]; ok {
		return v, nil
	}
	return "", os.ErrNotExist
}

func (s *stubRegistry) WriteExpandString(_, _, _ string) error  { return nil }
func (s *stubRegistry) WriteString(_, _, _ string) error        { return nil }
func (s *stubRegistry) ReadInteger(_, _ string) (uint64, error) { return 0, os.ErrNotExist }
func (s *stubRegistry) ReadIntegerLM(_, _ string) (uint64, error) {
	return 0, os.ErrNotExist
}

type stubJunction struct {
	deleted []string
}

func (s *stubJunction) IsJunction(string) bool           { return false }
func (s *stubJunction) CreateJunction(_, _ string) error { return nil }
func (s *stubJunction) DeleteJunction(path string) error {
	s.deleted = append(s.deleted, path)
	return nil
}
func (s *stubJunction) ResolveTarget(string) (string, error) { return "", os.ErrNotExist }

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		b    int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.b); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.b, got, tt.want)
		}
	}
}

func TestCleanupTempJunctions(t *testing.T) {
	tmp := t.TempDir()
	sourcePath := filepath.Join(tmp, "Desktop")
	tempPath := sourcePath + ".movekit_tmp"
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		t.Fatalf("mkdir temp junction: %v", err)
	}

	reg := &stubRegistry{values: map[string]string{"Desktop": sourcePath}}
	jn := &stubJunction{}

	orig := Sys
	Sys = winsyscall.SyscallLayer{Registry: reg, Junction: jn}
	defer func() { Sys = orig }()

	CleanupTempJunctions()

	if len(jn.deleted) != 1 || jn.deleted[0] != tempPath {
		t.Errorf("expected DeleteJunction(%q), got %v", tempPath, jn.deleted)
	}
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Errorf("temp junction should be removed, stat err = %v", err)
	}
}

func TestRollbackCrossVolume(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("more"), 0644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	if err := rollbackCrossVolume(src, dst, nil); err != nil {
		t.Fatalf("rollbackCrossVolume: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Errorf("a.txt not copied to dst: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "b.txt")); err != nil {
		t.Errorf("sub/b.txt not copied to dst: %v", err)
	}
	if _, err := os.Stat(filepath.Join(src, "a.txt")); !os.IsNotExist(err) {
		t.Errorf("source a.txt should be removed, stat err = %v", err)
	}
}
