package winsyscall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRealLayer_AllFieldsNonNil(t *testing.T) {
	layer := NewRealLayer()

	if layer.Registry == nil {
		t.Error("Registry is nil")
	}
	if layer.Process == nil {
		t.Error("Process is nil")
	}
	if layer.Volume == nil {
		t.Error("Volume is nil")
	}
	if layer.Junction == nil {
		t.Error("Junction is nil")
	}
	if layer.Shell == nil {
		t.Error("Shell is nil")
	}
	if layer.File == nil {
		t.Error("File is nil")
	}
	if layer.Power == nil {
		t.Error("Power is nil")
	}
	t.Logf("NewRealLayer: all 7 fields non-nil")
}

func TestRealImpl_RegistryAccessor(t *testing.T) {
	layer := NewRealLayer()

	val, err := layer.Registry.ReadExpandString(UserShellFoldersKey, "Desktop")
	if err != nil {
		t.Logf("ReadExpandString returned error (key may not exist on CI): %v", err)
	} else {
		t.Logf("ReadExpandString returned: %s", val)
	}
}

func TestRealImpl_VolumeChecker(t *testing.T) {
	layer := NewRealLayer()

	info, err := layer.Volume.GetVolumeInfo("C")
	if err != nil {
		t.Logf("GetVolumeInfo(C:) failed (expected if drive missing): %v", err)
		return
	}
	if info.FileSystem == "" {
		t.Error("FileSystem is empty for existing C: drive")
	}
	t.Logf("C: drive filesystem: %s", info.FileSystem)
}

func TestRealImpl_JunctionOps(t *testing.T) {
	layer := NewRealLayer()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "plain_dir")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	if layer.Junction.IsJunction(target) {
		t.Error("IsJunction returned true for a plain directory")
	}
	t.Logf("IsJunction correctly returns false for plain directory")
}
