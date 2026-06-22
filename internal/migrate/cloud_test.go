package migrate

import (
	"testing"
)

func TestDetectCloudSync_ReturnsSlice(t *testing.T) {
	providers := DetectCloudSync()
	// nil is a valid empty slice in Go — DetectCloudSync may return nil
	// when no cloud providers are installed.
	for _, p := range providers {
		if p.Name == "" {
			t.Error("CloudProvider has empty Name")
		}
		if p.Folder == "" {
			t.Error("CloudProvider has empty Folder")
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
