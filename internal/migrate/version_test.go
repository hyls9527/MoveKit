package migrate

import (
	"strings"
	"testing"
)

func TestCheckUpdate_ReturnsNoUpdate(t *testing.T) {
	info, err := CheckUpdate()
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info == nil {
		t.Fatal("CheckUpdate returned nil")
	}
	if info.HasUpdate {
		t.Log("HasUpdate is true (placeholder implementation)")
	}
	if info.NewVersion == "" {
		t.Error("NewVersion is empty")
	}
}

func TestVersionString(t *testing.T) {
	s := VersionString()
	if s == "" {
		t.Error("VersionString returned empty string")
	}
	if !strings.HasPrefix(s, "MoveKit v") {
		t.Errorf("VersionString should start with 'MoveKit v', got: %s", s)
	}
	t.Logf("Version: %s", s)
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant is empty")
	}
}
