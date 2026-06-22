package migrate

import (
	"testing"
)

func TestMoveKitError_ErrorMethod(t *testing.T) {
	e := &MoveKitError{
		Code:    "TEST_CODE",
		Message: "测试错误消息",
		Detail:  "详细错误信息",
		Action:  "建议操作",
		Warning: "警告信息",
	}
	if e.Error() != e.Message {
		t.Errorf("Error() = %s, want %s", e.Error(), e.Message)
	}
}

func TestErrorCodeConstants(t *testing.T) {
	codes := map[string]bool{}
	checkConsts := []string{
		ErrCodeOneDriveKFM,
		ErrCodeSpaceInsufficient,
		ErrCodeFileLocked,
		ErrCodePermissionDenied,
		ErrCodeJunctionFailed,
		ErrCodeRegistryFailed,
		ErrCodeTargetExists,
		ErrCodeFileSystemNotNTFS,
		ErrCodeNetworkDrive,
		ErrCodeRemovableDrive,
		ErrCodeBitLockerLocked,
		ErrCodeControlledFolder,
		ErrCodeWindowsUpdate,
		ErrCodeHyperVLocked,
		ErrCodeOneDriveSync,
		ErrCodeGPORestricted,
		ErrCodeTargetInSource,
	}
	for _, c := range checkConsts {
		if c == "" {
			t.Error("error code constant is empty")
		}
		if codes[c] {
			t.Errorf("duplicate error code: %s", c)
		}
		codes[c] = true
	}
	t.Logf("Verified %d unique error code constants", len(checkConsts))
}

func TestMoveKitError_WarningOmitEmpty(t *testing.T) {
	e := MoveKitError{
		Code:    "NO_WARNING",
		Message: "No warning",
	}
	e2 := MoveKitError{
		Code:    "WITH_WARNING",
		Message: "Has warning",
		Warning: "caution",
	}
	_ = e
	_ = e2
}
