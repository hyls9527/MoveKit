package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"

	sys "github.com/movekit/movekit/internal/syscall"
)

// dirSize returns the total size in bytes of all files under dir.
func dirSize(dir string) (int64, error) {
	var size int64
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// PreCheck validates that a migration can proceed from sourcePath to the target drive.
// Returns nil if all checks pass, or a MoveKitError describing the failure.
func PreCheck(sourcePath, targetDrive string) *MoveKitError {
	// 0. BitLocker check on target drive.
	if err := checkBitLocker(targetDrive); err != nil {
		return err
	}

	// 0b. Controlled Folder Access (ransomware protection).
	if err := checkControlledFolderAccess(); err != nil {
		return err
	}

	// 0c. Windows Update pending reboot.
	if err := checkWindowsUpdate(); err != nil {
		return err
	}

	// 0d. Battery check (warn if on battery).
	if err := checkBattery(); err != nil {
		return err
	}

	// 0e. Security software detection (informational — logged, not blocking).
	if avName := detectSecuritySoftware(); avName != "" {
		// Log detected AV but don't block the migration.
		// The caller can use this information to warn the user.
		_ = avName // future: pass to caller via context or log
	}

	// 1. Source must exist.
	fi, err := os.Stat(sourcePath)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "Source path does not exist or is not accessible",
			Detail:  err.Error(),
			Action:  "Verify the source folder exists and you have read access",
		}
	}
	if !fi.IsDir() {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "Source path is not a directory",
			Detail:  sourcePath,
			Action:  "Select a valid directory as the source",
		}
	}

	// 2. Filesystem must be NTFS.
	driveLetter := targetDrive
	if len(driveLetter) >= 2 && driveLetter[1] == ':' {
		driveLetter = driveLetter[:1]
	}
	volInfo, err := sys.GetVolumeInfo(driveLetter)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeFileSystemNotNTFS,
			Message: "Cannot read volume information for target drive",
			Detail:  err.Error(),
			Action:  "Verify the target drive is connected and accessible",
		}
	}
	if volInfo.FileSystem != "NTFS" {
		return &MoveKitError{
			Code:    ErrCodeFileSystemNotNTFS,
			Message: "Target drive is not NTFS (required for junction points)",
			Detail:  "Filesystem: " + volInfo.FileSystem,
			Action:  "MoveKit requires an NTFS volume. Reformat or choose a different drive.",
		}
	}

	// 3. Drive type must not be REMOVABLE or NETWORK.
	if volInfo.IsRemovable {
		return &MoveKitError{
			Code:    ErrCodeRemovableDrive,
			Message: "Target drive is removable",
			Detail:  "Removable drives are not supported for migration",
			Action:  "Use an internal fixed drive as the target",
		}
	}
	if volInfo.IsNetwork {
		return &MoveKitError{
			Code:    ErrCodeNetworkDrive,
			Message: "Target drive is a network drive",
			Detail:  "Network drives are not supported for migration",
			Action:  "Use a local drive as the target",
		}
	}

	// 4. Space must be sufficient.
	srcSize, err := dirSize(sourcePath)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "Cannot calculate source directory size",
			Detail:  err.Error(),
			Action:  "Verify you have read access to all files in the source",
		}
	}

	freeBytes, _, _, err := sys.GetDiskFreeSpace(driveLetter)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeSpaceInsufficient,
			Message: "Cannot read free space on target drive",
			Detail:  err.Error(),
			Action:  "Verify the target drive is accessible",
		}
	}
	if freeBytes < srcSize {
		return &MoveKitError{
			Code:    ErrCodeSpaceInsufficient,
			Message: "Insufficient space on target drive",
			Detail:  "Need " + formatBytes(srcSize) + " but only " + formatBytes(freeBytes) + " free",
			Action:  "Free up space on the target drive or choose a larger one",
		}
	}

	// 5. Target path must not already exist.
	targetPath := filepath.Join(targetDrive+":\\", filepath.Base(sourcePath))
	if _, err := os.Stat(targetPath); err == nil {
		return &MoveKitError{
			Code:    ErrCodeTargetExists,
			Message: "Target path already exists",
			Detail:  targetPath,
			Action:  "Remove or rename the existing folder at the target location",
		}
	}

	return nil
}

// checkBitLocker checks whether BitLocker encryption is active on the given drive.
// Uses registry detection. Returns nil if BitLocker is not configured or not locked.
func checkBitLocker(drive string) *MoveKitError {
	// Normalize drive letter.
	dl := drive
	if len(dl) >= 2 && dl[1] == ':' {
		dl = dl[:1]
	}
	// Check the standard BitLocker registry key under HKLM.
	keyPath := `SOFTWARE\Microsoft\Windows\CurrentVersion\BitLocker`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ)
	if err != nil {
		return nil // key doesn't exist, BitLocker not configured
	}
	defer k.Close()

	// Try to read a subkey for the specific drive. If BitLocker is configured
	// but we can't determine the protection status, assume OK (not locked).
	// A full implementation would use Win32_EncryptableVolume WMI or
	// DeviceIoControl with IOCTL_VOLUME_GET_VOLUME_DISK_EXTENTS.
	// TODO: implement full BitLocker protection status check via WMI.
	_ = dl
	return nil
}

// checkControlledFolderAccess checks whether Windows Controlled Folder Access
// (ransomware protection) is enabled, which may block file operations.
func checkControlledFolderAccess() *MoveKitError {
	val, err := sys.ReadIntegerLM(
		`SOFTWARE\Policies\Microsoft\Windows Defender\Real-Time Protection`,
		"EnableControlledFolderAccess",
	)
	if err != nil {
		return nil // not configured
	}
	if val == 1 {
		return &MoveKitError{
			Code:    ErrCodeControlledFolder,
			Message: "Windows 受控文件夹访问已启用",
			Detail:  "勒索软件防护功能可能会阻止 MoveKit 迁移文件",
			Action:  "请在 Windows 安全中心 → 病毒和威胁防护 → 勒索软件防护 → 允许应用中添加 MoveKit",
		}
	}
	return nil
}

// checkWindowsUpdate checks whether a Windows Update reboot is pending,
// which could interfere with file operations.
func checkWindowsUpdate() *MoveKitError {
	// Check the standard pending-reboot registry key.
	keyPath := `SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update`
	_, err := sys.ReadIntegerLM(keyPath, "RebootRequired")
	if err == nil {
		return &MoveKitError{
			Code:    ErrCodeWindowsUpdate,
			Message: "Windows 更新可能正在安装",
			Detail:  "系统有待完成的更新",
			Action:  "请等待 Windows 更新完成后再迁移",
		}
	}
	return nil
}

// checkBattery warns if the system is running on battery power, since
// cross-volume migration may take a long time.
func checkBattery() *MoveKitError {
	if sys.IsOnBattery() {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "笔记本未接电源",
			Detail:  "跨卷迁移可能耗时较长，建议连接电源适配器",
			Action:  "请连接电源适配器后再迁移（也可忽略此警告继续）",
		}
	}
	return nil
}

// detectSecuritySoftware scans running processes for known antivirus executables.
// Returns the display name of the detected AV, or empty string if none found.
func detectSecuritySoftware() string {
	knownAVs := map[string]string{
		"360sd.exe":   "360 杀毒",
		"360tray.exe": "360 安全卫士",
		"HuoRong.exe": "火绒安全",
		"avp.exe":     "卡巴斯基",
		"MsMpEng.exe": "Windows Defender",
		"avgui.exe":   "AVG",
		"avguard.exe": "Avira",
		"bdagent.exe": "Bitdefender",
	}
	for proc, name := range knownAVs {
		if sys.IsProcessRunning(proc) {
			return name
		}
	}
	return ""
}

// formatBytes returns a human-readable string for the given byte count.
func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
