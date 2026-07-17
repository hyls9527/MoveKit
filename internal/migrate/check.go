package migrate

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirSize returns the total size in bytes of all files under dir.
func DirSize(dir string) (int64, error) {
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
	var detectedAV string
	if avName := detectSecuritySoftware(); avName != "" {
		detectedAV = avName
	}

	// 1. Source must exist.
	fi, err := os.Stat(sourcePath)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "源路径不存在或无法访问",
			Detail:  err.Error(),
			Action:  "Verify the source folder exists and you have read access",
		}
	}
	if !fi.IsDir() {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "源路径不是目录",
			Detail:  sourcePath,
			Action:  "Select a valid directory as the source",
		}
	}

	// 2. Filesystem must be NTFS.
	driveLetter := targetDrive
	if len(driveLetter) >= 2 && driveLetter[1] == ':' {
		driveLetter = driveLetter[:1]
	}
	volInfo, err := Sys.Volume.GetVolumeInfo(driveLetter)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeFileSystemNotNTFS,
			Message: "无法读取目标盘的卷信息",
			Detail:  err.Error(),
			Action:  "Verify the target drive is connected and accessible",
		}
	}
	if volInfo.FileSystem != "NTFS" {
		return &MoveKitError{
			Code:    ErrCodeFileSystemNotNTFS,
			Message: "目标盘不是 NTFS 文件系统（Junction 依赖 NTFS）",
			Detail:  "Filesystem: " + volInfo.FileSystem,
			Action:  "MoveKit requires an NTFS volume. Reformat or choose a different drive.",
		}
	}

	// 3. Drive type must not be REMOVABLE or NETWORK.
	if volInfo.IsRemovable {
		return &MoveKitError{
			Code:    ErrCodeRemovableDrive,
			Message: "目标盘是可移动驱动器",
			Detail:  "Removable drives are not supported for migration",
			Action:  "Use an internal fixed drive as the target",
		}
	}
	if volInfo.IsNetwork {
		return &MoveKitError{
			Code:    ErrCodeNetworkDrive,
			Message: "目标盘是网络驱动器",
			Detail:  "Network drives are not supported for migration",
			Action:  "Use a local drive as the target",
		}
	}

	// 4. Space must be sufficient.
	srcSize, err := DirSize(sourcePath)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "无法计算源目录大小",
			Detail:  err.Error(),
			Action:  "Verify you have read access to all files in the source",
		}
	}

	freeBytes, _, _, err := Sys.Volume.GetDiskFreeSpace(driveLetter)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeSpaceInsufficient,
			Message: "无法读取目标盘剩余空间",
			Detail:  err.Error(),
			Action:  "Verify the target drive is accessible",
		}
	}
	if freeBytes < srcSize {
		return &MoveKitError{
			Code:    ErrCodeSpaceInsufficient,
			Message: "目标盘空间不足",
			Detail:  "Need " + formatBytes(srcSize) + " but only " + formatBytes(freeBytes) + " free",
			Action:  "Free up space on the target drive or choose a larger one",
		}
	}

	// 5. Target path must not already exist.
	targetPath := filepath.Join(targetDrive+":\\", filepath.Base(sourcePath))
	if _, err := os.Stat(targetPath); err == nil {
		return &MoveKitError{
			Code:    ErrCodeTargetExists,
			Message: "目标路径已存在",
			Detail:  targetPath,
			Action:  "Remove or rename the existing folder at the target location",
		}
	}

	// 6. Warn about detected security software (informational, non-blocking).
	if detectedAV != "" {
		return &MoveKitError{
			Code:    ErrCodeSecuritySoftware,
			Message: fmt.Sprintf("检测到安全软件: %s", detectedAV),
			Detail:  "安全软件可能会干扰 junction 创建或文件操作",
			Action:  "建议暂时关闭实时防护后再迁移",
			Warning: detectedAV,
		}
	}

	return nil
}

// checkBitLocker checks whether BitLocker encryption is active on the given drive.
//
// STUB: Always returns nil. Real BitLocker protection status detection requires
// WMI (Win32_EncryptableVolume) or DeviceIoControl(IOCTL_VOLUME_GET_VOLUME_DISK_EXTENTS).
// Neither is implemented yet — see .planning/codebase/CONCERNS.md.
func checkBitLocker(drive string) *MoveKitError {
	_ = drive
	return nil
}

// checkControlledFolderAccess checks whether Windows Controlled Folder Access
// (ransomware protection) is enabled, which may block file operations.
func checkControlledFolderAccess() *MoveKitError {
	val, err := Sys.Registry.ReadIntegerLM(
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
	_, err := Sys.Registry.ReadIntegerLM(keyPath, "RebootRequired")
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
	if Sys.Power.IsOnBattery() {
		return &MoveKitError{
			Code:    ErrCodeBatteryWarning,
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
		if Sys.Process.IsProcessRunning(proc) {
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
