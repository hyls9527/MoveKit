package migrate

const (
	ErrCodeOneDriveKFM       = "ONEDRIVE_KFM"
	ErrCodeSpaceInsufficient = "SPACE_INSUFFICIENT"
	ErrCodeFileLocked        = "FILE_LOCKED"
	ErrCodePermissionDenied  = "PERMISSION_DENIED"
	ErrCodeJunctionFailed    = "JUNCTION_FAILED"
	ErrCodeRegistryFailed    = "REGISTRY_FAILED"
	ErrCodeTargetExists      = "TARGET_EXISTS"
	ErrCodeFileSystemNotNTFS = "FS_NOT_NTFS"
	ErrCodeNetworkDrive      = "NETWORK_DRIVE"
	ErrCodeRemovableDrive    = "REMOVABLE_DRIVE"
	ErrCodeBitLockerLocked   = "BITLOCKER_LOCKED"
	ErrCodeControlledFolder  = "CONTROLLED_FOLDER"
	ErrCodeWindowsUpdate     = "WINDOWS_UPDATE"
	ErrCodeHyperVLocked      = "HYPERV_LOCKED"
	ErrCodeOneDriveSync      = "ONEDRIVE_SYNC"
	ErrCodeGPORestricted     = "GPO_RESTRICTED"
	ErrCodeTargetInSource    = "TARGET_IN_SOURCE"
)

type MoveKitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Action  string `json:"action"`
	Warning string `json:"warning,omitempty"` // Non-blocking informational warning
}

func (e *MoveKitError) Error() string { return e.Message }
