package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckJunctionHijack_NonexistentPath(t *testing.T) {
	err := CheckJunctionHijack(`C:\ThisPathDoesNotExist_Hijack_Test`)
	if err != nil {
		t.Errorf("CheckJunctionHijack should return nil for nonexistent path, got: %v", err)
	}
}

func TestCheckJunctionHijack_PlainDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "plain_dir")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	err := CheckJunctionHijack(target)
	if err != nil {
		t.Errorf("CheckJunctionHijack should return nil for plain directory, got: %v", err)
	}
}
