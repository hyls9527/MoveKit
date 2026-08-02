package migrate

import (
	"testing"

	sys "github.com/hyls9527/MoveKit/internal/syscall"
)

func TestDetectCloudSync_ReturnsSlice(t *testing.T) {
	providers := DetectCloudSync()
	for _, p := range providers {
		if p.Name == "" {
			t.Error("CloudProvider has empty Name")
		}
	}
	t.Logf("Detected %d cloud sync providers", len(providers))
}

func TestCloudProviderFields(t *testing.T) {
	cp := CloudProvider{Name: "OneDrive", Folder: "C:\\Test"}
	if cp.Name != "OneDrive" {
		t.Errorf("Name = %s, want OneDrive", cp.Name)
	}
	if cp.Folder != "C:\\Test" {
		t.Errorf("Folder = %s, want C:\\Test", cp.Folder)
	}
}

type fakeProcessChecker struct {
	running map[string]bool
}

func (f fakeProcessChecker) IsProcessRunning(name string) bool {
	return f.running[name]
}

func TestDetectCloudSync_UsesRunningProcesses(t *testing.T) {
	origSys := Sys
	defer func() { Sys = origSys }()

	Sys = sys.SyscallLayer{
		Process: fakeProcessChecker{running: map[string]bool{"OneDrive.exe": true}},
	}
	providers := DetectCloudSync()
	if len(providers) != 1 || providers[0].Name != "OneDrive" {
		t.Fatalf("expected only OneDrive, got %+v", providers)
	}

	Sys = sys.SyscallLayer{
		Process: fakeProcessChecker{running: map[string]bool{}},
	}
	providers = DetectCloudSync()
	if len(providers) != 0 {
		t.Fatalf("expected no providers when nothing is running, got %+v", providers)
	}
}
