package migrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadClearState(t *testing.T) {
	tmpDir := t.TempDir()
	appDataOrig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", appDataOrig)

	state := MigrationState{
		Status:     "in_progress",
		Folder:     "Desktop",
		Source:     filepath.Join(tmpDir, "Desktop"),
		Target:     "D:\\Desktop",
		StartTime:  time.Now(),
		FilesMoved: 42,
	}

	err := SaveState(state)
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadState returned nil for saved state")
	}
	if loaded.Folder != "Desktop" {
		t.Errorf("Folder = %s, want Desktop", loaded.Folder)
	}
	if loaded.FilesMoved != 42 {
		t.Errorf("FilesMoved = %d, want 42", loaded.FilesMoved)
	}

	err = ClearState()
	if err != nil {
		t.Fatalf("ClearState: %v", err)
	}

	loaded2, _ := LoadState()
	if loaded2 != nil {
		t.Error("LoadState after ClearState should return nil")
	}
}
