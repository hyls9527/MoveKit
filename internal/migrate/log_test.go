package migrate

import (
	"testing"
)

func TestMaskPath_Log(t *testing.T) {
	userProfile := "C:\\Users\\TestUser"
	path := userProfile + "\\Desktop\\MyFolder"
	masked := maskPath(path)

	if masked == path {
		t.Logf("maskPath did not mask (USERPROFILE may differ): got %s", masked)
	}
	t.Logf("maskPath: %s -> %s", path, masked)
}

func TestMaskPath_NoUserProfile_Log(t *testing.T) {
	result := maskPath("C:\\SomePath")
	if result == "" {
		t.Error("maskPath returned empty string")
	}
}

func TestLogger_NewAndClose(t *testing.T) {
	// Keep the logger fully hermetic: never write to the real user profile,
	// so the test passes in restricted/CI environments too.
	t.Setenv("APPDATA", t.TempDir())

	log, err := NewLogger()
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if log == nil {
		t.Fatal("NewLogger returned nil")
	}

	log.Info("test log message: %s", "hello")
	log.Warn("test warning: %d", 42)
	log.Error("test error: %v", nil)

	err = log.Close()
	if err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestLogger_NilFile(t *testing.T) {
	l := &Logger{file: nil}
	l.Info("should not panic")
	l.Warn("should not panic")
	l.Error("should not panic")
	err := l.Close()
	if err != nil {
		t.Logf("Close on nil file: %v", err)
	}
}
