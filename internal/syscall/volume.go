package winsyscall

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Drive type constants returned by GetDriveTypeW.
const (
	DRIVE_UNKNOWN   = 0
	DRIVE_REMOVABLE = 2
	DRIVE_FIXED     = 3
	DRIVE_REMOTE    = 4
	DRIVE_CDROM     = 5
	DRIVE_RAMDISK   = 6
)

var (
	procGetVolumeInformationW = modkernel32.NewProc("GetVolumeInformationW")
	procGetDriveTypeW         = modkernel32.NewProc("GetDriveTypeW")
	procGetDiskFreeSpaceExW   = modkernel32.NewProc("GetDiskFreeSpaceExW")
)

// VolumeInfo holds information about a volume (drive).
type VolumeInfo struct {
	VolumeName   string
	FileSystem   string
	VolumeSN     uint32
	IsRemovable  bool
	IsNetwork    bool
	IsFixedDrive bool
}

// GetVolumeInfo returns volume information for the given drive letter.
// The drive parameter should be a drive letter such as "C" or "C:".
func GetVolumeInfo(drive string) (VolumeInfo, error) {
	root := driveRootPath(drive)

	rootPtr, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return VolumeInfo{}, fmt.Errorf("utf16 root path: %w", err)
	}

	var fsNameBuf [256]uint16
	var volNameBuf [256]uint16
	var serialNumber uint32
	var maxComponentLen uint32
	var fsFlags uint32

	r1, _, e := procGetVolumeInformationW.Call(
		uintptr(unsafe.Pointer(rootPtr)),
		uintptr(unsafe.Pointer(&volNameBuf[0])),
		uintptr(len(volNameBuf)),
		uintptr(unsafe.Pointer(&serialNumber)),
		uintptr(unsafe.Pointer(&maxComponentLen)),
		uintptr(unsafe.Pointer(&fsFlags)),
		uintptr(unsafe.Pointer(&fsNameBuf[0])),
		uintptr(len(fsNameBuf)),
	)
	if r1 == 0 {
		return VolumeInfo{}, fmt.Errorf("GetVolumeInformationW(%s): %w", root, e)
	}

	driveType := GetDriveType(drive)

	return VolumeInfo{
		VolumeName:  syscall.UTF16ToString(volNameBuf[:]),
		FileSystem:  syscall.UTF16ToString(fsNameBuf[:]),
		VolumeSN:    serialNumber,
		IsRemovable: driveType == DRIVE_REMOVABLE,
		IsNetwork:   driveType == DRIVE_REMOTE,
		// IsFixedDrive indicates a non-removable, non-network drive.
		// SSD vs HDD differentiation requires more advanced queries (e.g. DeviceIoControl
		// with IOCTL_STORAGE_QUERY_PROPERTY). Not implemented yet.
		IsFixedDrive: driveType == DRIVE_FIXED,
	}, nil
}

// GetDriveType returns the drive type for the given drive letter.
// The drive parameter should be a drive letter such as "C" or "C:".
func GetDriveType(drive string) uint32 {
	root := driveRootPath(drive)
	rootPtr, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return DRIVE_UNKNOWN
	}
	r1, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(rootPtr)))
	return uint32(r1)
}

// GetDiskFreeSpace returns disk space information for the given drive letter.
// Returns freeBytesAvailableToCaller, totalNumberOfBytes, and totalNumberOfFreeBytes.
func GetDiskFreeSpace(drive string) (freeBytes, totalBytes, totalFreeBytes int64, err error) {
	root := driveRootPath(drive)
	rootPtr, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("utf16 root path: %w", err)
	}

	var freeAvail int64
	var total int64
	var totalFree int64

	r1, _, e := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(rootPtr)),
		uintptr(unsafe.Pointer(&freeAvail)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if r1 == 0 {
		return 0, 0, 0, fmt.Errorf("GetDiskFreeSpaceExW(%s): %w", root, e)
	}

	return freeAvail, total, totalFree, nil
}

// driveRootPath normalises a drive letter to the "C:\" form required by Win32 volume APIs.
func driveRootPath(drive string) string {
	if len(drive) == 1 {
		return drive + `:\`
	}
	if len(drive) == 2 && drive[1] == ':' {
		return drive + `\`
	}
	// Already has trailing backslash or is a UNC path — return as-is.
	return drive
}
