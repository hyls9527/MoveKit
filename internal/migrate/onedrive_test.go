package migrate

import "testing"

func TestDetectOneDriveKFM(t *testing.T) {
	// This test verifies DetectOneDriveKFM doesn't panic.
	// On a system without OneDrive, it should return IsActive=false without error.
	status, err := DetectOneDriveKFM()
	if err != nil {
		t.Fatalf("DetectOneDriveKFM returned error: %v", err)
	}
	if status == nil {
		t.Fatal("DetectOneDriveKFM returned nil status")
	}
	t.Logf("OneDrive KFM: IsActive=%v, Folders=%v, HasGPOBlock=%v",
		status.IsActive, status.Folders, status.HasGPOBlock)
}
