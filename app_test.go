package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hyls9527/MoveKit/internal/migrate"
)

func TestDirSize_NonexistentPath(t *testing.T) {
	_, err := migrate.DirSize(`C:\ThisPathDoesNotExist_12345`)
	if err == nil {
		t.Error("dirSize(nonexistent) should return error, got nil")
	}
}

func TestDirSize_ValidPath(t *testing.T) {
	size, err := migrate.DirSize(t.TempDir())
	if err != nil {
		t.Fatalf("dirSize(TempDir) unexpected error: %v", err)
	}
	if size != 0 {
		t.Errorf("empty TempDir should have size 0, got %d", size)
	}
}

func TestGetFolders_DirSizeError(t *testing.T) {
	_, err := migrate.DirSize(`X:\Missing\Path`)
	if err == nil {
		t.Skip("unexpected: path existed (can't test error path)")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		sec  int
		want string
	}{
		{0, "0秒"},
		{30, "30秒"},
		{60, "1分钟"},
		{90, "1分30秒"},
		{3600, "1小时0分钟"},
		{3661, "1小时1分钟"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.sec)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.sec, got, tt.want)
		}
	}
}

func TestPathsEqual(t *testing.T) {
	if !pathsEqual(`C:\Users\Test`, `C:\Users\Test`) {
		t.Error("pathsEqual should be true for identical paths")
	}
	if !pathsEqual(`C:\Users\Test`, `c:\users\test`) {
		t.Error("pathsEqual should be case-insensitive on Windows")
	}
	if pathsEqual(`C:\Users\Test`, `C:\Users\Test2`) {
		t.Error("pathsEqual should be false for different paths")
	}
}

func TestConfigDir(t *testing.T) {
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", `C:\Users\Test\AppData\Roaming`)
	defer os.Setenv("APPDATA", orig)

	dir := configDir()
	expected := filepath.Join(`C:\Users\Test\AppData\Roaming`, "MoveKit")
	if dir != expected {
		t.Errorf("configDir() = %s, want %s", dir, expected)
	}
}

func TestConfigDir_FallbackToUserProfile(t *testing.T) {
	origAppData := os.Getenv("APPDATA")
	origUserProfile := os.Getenv("USERPROFILE")
	os.Unsetenv("APPDATA")
	os.Setenv("USERPROFILE", `C:\Users\Fallback`)
	defer func() {
		os.Setenv("APPDATA", origAppData)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	dir := configDir()
	expected := filepath.Join(`C:\Users\Fallback`, "AppData", "Roaming", "MoveKit")
	if dir != expected {
		t.Errorf("configDir() fallback = %s, want %s", dir, expected)
	}
}

func TestConfigFilePath(t *testing.T) {
	orig := os.Getenv("APPDATA")
	os.Setenv("APPDATA", `C:\Users\Test\AppData\Roaming`)
	defer os.Setenv("APPDATA", orig)

	path := configFilePath()
	expected := filepath.Join(`C:\Users\Test\AppData\Roaming`, "MoveKit", "config.json")
	if path != expected {
		t.Errorf("configFilePath() = %s, want %s", path, expected)
	}
}

func TestGetFirstRunFlag_NoConfig(t *testing.T) {
	orig := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", orig)

	app := NewApp()
	if !app.GetFirstRunFlag() {
		t.Error("GetFirstRunFlag should return true when no config exists")
	}
}

func TestGetFirstRunFlag_MalformedConfig(t *testing.T) {
	orig := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", orig)

	dir := filepath.Join(tmpDir, "MoveKit")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not-json"), 0644); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	app := NewApp()
	if !app.GetFirstRunFlag() {
		t.Error("malformed config should be treated as first run")
	}
}

func TestSetFirstRunFlag_RoundTrip(t *testing.T) {
	orig := os.Getenv("APPDATA")
	tmpDir := t.TempDir()
	os.Setenv("APPDATA", tmpDir)
	defer os.Setenv("APPDATA", orig)

	app := NewApp()
	err := app.SetFirstRunFlag(false)
	if err != nil {
		t.Fatalf("SetFirstRunFlag(false): %v", err)
	}

	if app.GetFirstRunFlag() {
		t.Error("GetFirstRunFlag should return false after SetFirstRunFlag(false)")
	}

	err = app.SetFirstRunFlag(true)
	if err != nil {
		t.Fatalf("SetFirstRunFlag(true): %v", err)
	}

	if !app.GetFirstRunFlag() {
		t.Error("GetFirstRunFlag should return true after SetFirstRunFlag(true)")
	}
}

func TestGetVersion(t *testing.T) {
	app := NewApp()
	v := app.GetVersion()
	if v == "" {
		t.Error("GetVersion returned empty string")
	}
}

func TestHealthCheck(t *testing.T) {
	app := NewApp()
	if !app.HealthCheck() {
		t.Error("HealthCheck should return true")
	}
}

func TestProgressThrottle(t *testing.T) {
	base := time.Now()
	th := &progressThrottle{interval: 150 * time.Millisecond}

	if !th.allow("moving", base) {
		t.Error("first event should be allowed")
	}
	if th.allow("moving", base.Add(50*time.Millisecond)) {
		t.Error("event inside the interval should be dropped")
	}
	if !th.allow("moving", base.Add(200*time.Millisecond)) {
		t.Error("event after the interval should be allowed")
	}
	if !th.allow("complete", base.Add(1*time.Millisecond)) {
		t.Error("phase transitions must always pass through")
	}
}

func TestResolveConflict_NotImplemented(t *testing.T) {
	app := NewApp()
	err := app.ResolveConflict("test/file.txt", "skip")
	if err == nil {
		t.Error("ResolveConflict should return NOT_IMPLEMENTED error")
	}
}
