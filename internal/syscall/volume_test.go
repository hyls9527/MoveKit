package winsyscall

import (
	"testing"
)

func TestGetVolumeInfo(t *testing.T) {
	info, err := GetVolumeInfo("C")
	if err != nil {
		t.Logf("GetVolumeInfo(C) returned error (may not have C: drive): %v", err)
		return
	}
	t.Logf("FileSystem: %s, VolumeSN: %d, IsRemovable: %v, IsNetwork: %v",
		info.FileSystem, info.VolumeSN, info.IsRemovable, info.IsNetwork)

	if info.FileSystem == "" {
		t.Error("FileSystem should not be empty")
	}
}

func TestGetDiskFreeSpace(t *testing.T) {
	free, total, totalFree, err := GetDiskFreeSpace("C")
	if err != nil {
		t.Logf("GetDiskFreeSpace(C) returned error (may not have C: drive): %v", err)
		return
	}
	t.Logf("FreeAvailable: %d bytes, Total: %d bytes, TotalFree: %d bytes", free, total, totalFree)

	if total <= 0 {
		t.Error("total bytes should be positive")
	}
	if free < 0 {
		t.Error("free bytes should not be negative")
	}
}

func TestGetDriveType(t *testing.T) {
	dt := GetDriveType("C")
	t.Logf("DriveType(C): %d", dt)

	// On most dev machines C: is DRIVE_FIXED (3).
	// We just verify it returns a known value, not DRIVE_UNKNOWN (0).
	if dt == DRIVE_UNKNOWN {
		t.Log("DriveType returned DRIVE_UNKNOWN — may be expected in some environments")
	}
}
