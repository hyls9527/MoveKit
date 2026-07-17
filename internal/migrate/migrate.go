package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sys "github.com/movekit/movekit/internal/syscall"
)

// ProgressCallback is called during migration to report progress.
type ProgressCallback func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)

// CleanupTempJunctions removes any leftover .movekit_tmp junctions from previous failed migrations.
// Call this on application startup.
func CleanupTempJunctions() {
	for _, f := range []string{"Desktop", "Personal", "My Pictures", "My Music", "{374DE290-123F-4565-9164-39C4925E467B}"} {
		rawPath, err := Sys.Registry.ReadExpandString(sys.UserShellFoldersKey, f)
		if err != nil {
			continue
		}
		sourcePath := os.ExpandEnv(rawPath)
		tempPath := sourcePath + ".movekit_tmp"
		if _, err := os.Stat(tempPath); err == nil {
			// Temp junction exists, try to remove it
			Sys.Junction.DeleteJunction(tempPath)
			os.Remove(tempPath)
		}
	}
}

// Migrate moves a known folder to a new drive and creates a junction at the original location.
//
// Parameters:
//   - regValueName: the Shell Folders registry value name (e.g. "Desktop", "Personal")
//   - targetDrive: the target drive letter (e.g. "D" or "D:")
//   - onProgress: optional progress callback (may be nil)
func Migrate(regValueName, targetDrive string, onProgress ProgressCallback) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = &MoveKitError{
				Code:    "INTERNAL_ERROR",
				Message: "迁移过程发生内部错误",
				Detail:  fmt.Sprintf("%v", r),
				Action:  "请重试，如问题持续请联系开发者",
			}
		}
	}()
	log, err := NewLogger()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Close()

	log.Info("Starting migration of %s to %s", regValueName, targetDrive)

	// 1. Get the actual source path from registry (not hardcoded).
	sourcePath, err := Sys.Registry.ReadExpandString(sys.UserShellFoldersKey, regValueName)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "无法从注册表读取当前文件夹路径",
			Detail:  err.Error(),
			Action:  "Verify the folder is a known Windows folder",
		}
	}

	// Expand environment variables in the source path.
	sourcePath = os.ExpandEnv(sourcePath)
	log.Info("Source path: %s", maskPath(sourcePath))

	// 2. Chain migration: if source is already a junction, resolve to the real target.
	actualSource := sourcePath
	if Sys.Junction.IsJunction(sourcePath) {
		log.Info("Source is already a junction, resolving real target")
		junctionTarget, err := resolveJunctionTarget(sourcePath)
		if err != nil {
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "无法解析现有 Junction 目标",
				Detail:  err.Error(),
				Action:  "Check that the junction is not broken",
			}
		}
		actualSource = junctionTarget
		log.Info("Junction resolves to: %s", maskPath(actualSource))
	}

	// Validate drive letter.
	driveLetter := targetDrive
	if len(driveLetter) >= 2 && driveLetter[1] == ':' {
		driveLetter = driveLetter[:1]
	}
	if len(driveLetter) != 1 || driveLetter[0] < 'A' || driveLetter[0] > 'Z' {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "无效的目标盘符",
			Detail:  fmt.Sprintf("收到: %q", targetDrive),
			Action:  "请选择 A-Z 之间的盘符",
		}
	}
	folderName := filepath.Base(actualSource)
	targetPath := filepath.Join(driveLetter+":\\", folderName)
	log.Info("Target path: %s", maskPath(targetPath))

	// Check for junction hijack at target.
	if hijackErr := CheckJunctionHijack(targetPath); hijackErr != nil {
		return hijackErr
	}

	// 3. PreCheck.
	if preErr := PreCheck(actualSource, driveLetter); preErr != nil {
		if preErr.Code == ErrCodeSecuritySoftware {
			log.Warn("检测到安全软件: %s，继续迁移", preErr.Warning)
			// Non-blocking: proceed with migration despite AV detection
		} else {
			log.Error("PreCheck failed: %s — %s", preErr.Code, preErr.Detail)
			return preErr
		}
	}

	// Check for file conflicts before moving.
	// Skip if source is already a junction (re-migration): files are already at target.
	if !Sys.Junction.IsJunction(sourcePath) {
		conflicts, err := DetectConflicts(actualSource, targetPath)
		if err != nil {
			log.Warn("Conflict detection failed: %v", err)
		}
		if len(conflicts) > 0 {
			log.Error("Found %d file conflicts, blocking migration", len(conflicts))
			return &MoveKitError{
				Code:    ErrCodeTargetExists,
				Message: fmt.Sprintf("目标路径存在 %d 个同名文件", len(conflicts)),
				Detail:  fmt.Sprintf("冲突文件数: %d", len(conflicts)),
				Action:  "请先手动删除目标路径中的同名文件，或选择其他盘符",
			}
		}
	} else {
		log.Info("Source is junction (re-migration), skipping conflict detection")
	}

	// 4. Save backup record (status=in_progress).
	backupPath := backupFilePath(regValueName)
	backup := BackupRecord{
		Version:       1,
		RegValueName:  regValueName,
		OriginalPath:  sourcePath,
		NewPath:       targetPath,
		RegistryValue: regValueName,
		CreatedAt:     time.Now(),
		Status:        "in_progress",
	}
	if err := saveBackup(backupPath, backup); err != nil {
		log.Error("Failed to save initial backup record: %v", err)
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "无法保存迁移备份记录",
			Detail:  err.Error(),
			Action:  "请检查磁盘空间和 %APPDATA%\\MoveKit 目录权限后重试",
		}
	}

	// 5. Move files.
	if onProgress != nil {
		onProgress(regValueName, "moving", 0, 0, 0, 0)
	}

	log.Info("Moving files from %s to %s", maskPath(actualSource), maskPath(targetPath))
	if Sys.Volume.IsCrossVolume(actualSource, targetPath) {
		// Cross-volume: copy + delete.
		if err := moveFilesCrossVolume(actualSource, targetPath, onProgress, regValueName); err != nil {
			log.Error("Cross-volume move failed: %v", err)
			return err
		}
	} else {
		// Same-volume: use moveFilesSameVolume for better error handling.
		var svProgress func(ProgressInfo)
		if onProgress != nil {
			svProgress = func(pi ProgressInfo) {
				onProgress(regValueName, pi.Phase, 0, 0, 0, 0)
			}
		}
		if err := moveFilesSameVolume(actualSource, targetPath, svProgress); err != nil {
			log.Error("Same-volume move failed: %v", err)
			return err
		}
	}
	log.Info("Files moved successfully")

	// 6. Create junction at original location (atomic-safe strategy).
	junctionCreated := false

	// Step A: Check if existing junction already points to our target.
	if Sys.Junction.IsJunction(sourcePath) {
		existingTarget, resolveErr := resolveJunctionTarget(sourcePath)
		if resolveErr == nil && strings.EqualFold(existingTarget, targetPath) {
			log.Info("Existing junction already points to target, skipping recreation")
			junctionCreated = true
		} else {
			// Junction points elsewhere or is broken; delete it
			log.Info("Deleting old junction (points to %s)", maskPath(existingTarget))
			if delErr := Sys.Junction.DeleteJunction(sourcePath); delErr != nil {
				log.Warn("Failed to delete old junction: %v", delErr)
			}
		}
	}

	// Step B: Create junction if not already correct.
	if !junctionCreated {
		tempJunction := sourcePath + ".movekit_tmp"
		// Clean up any leftover temp junction
		os.Remove(tempJunction)

		// Create temp directory for junction
		if mdErr := os.MkdirAll(tempJunction, 0700); mdErr != nil {
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "无法创建临时目录",
				Detail:  mdErr.Error(),
				Action:  "检查源路径的写入权限",
			}
		}

		// Create junction at temp location
		if err := Sys.Junction.CreateJunction(tempJunction, targetPath); err != nil {
			os.Remove(tempJunction)
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "无法创建临时 junction",
				Detail:  err.Error(),
				Action:  "检查目标路径是否存在",
			}
		}

		// Verify temp junction points correctly
		resolved, resolveErr := resolveJunctionTarget(tempJunction)
		if resolveErr != nil || !strings.EqualFold(resolved, targetPath) {
			Sys.Junction.DeleteJunction(tempJunction)
			os.Remove(tempJunction)
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "Junction 验证失败",
				Detail:  fmt.Sprintf("期望 %s，实际 %s", targetPath, resolved),
				Action:  "请检查 NTFS 文件系统是否正常",
			}
		}

		// Ensure source directory doesn't exist (remove if it does)
		if _, statErr := os.Stat(sourcePath); statErr == nil {
			// Source exists but is not a junction (shouldn't happen after move)
			os.Remove(sourcePath)
		}

		// Atomic swap: rename temp junction to source path
		if renameErr := os.Rename(tempJunction, sourcePath); renameErr != nil {
			Sys.Junction.DeleteJunction(tempJunction)
			os.Remove(tempJunction)
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "Junction 替换失败",
				Detail:  renameErr.Error(),
				Action:  "请手动检查源路径状态",
			}
		}

		log.Info("Junction created atomically: %s -> %s", maskPath(sourcePath), maskPath(targetPath))
	}

	// 7. Update registry (with rollback on partial failure).
	// Backup old values first.
	oldUserShell, _ := Sys.Registry.ReadExpandString(sys.UserShellFoldersKey, regValueName)

	// User Shell Folders: REG_SZ (fully expanded path, no env vars)
	regValue := fmt.Sprintf("%s\\%s", driveLetter+":", folderName)
	if regErr := Sys.Registry.WriteString(sys.UserShellFoldersKey, regValueName, regValue); regErr != nil {
		log.Error("Failed to update User Shell Folders: %v", regErr)
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "注册表 User Shell Folders 写入失败",
			Detail:  regErr.Error(),
			Action:  "文件已移动但注册表未更新，请以管理员身份运行",
		}
	}

	// Shell Folders: REG_SZ (fully expanded).
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath // fallback if Abs fails
	}
	if regErr := Sys.Registry.WriteString(sys.ShellFoldersKey, regValueName, absTarget); regErr != nil {
		log.Error("Failed to update Shell Folders: %v — rolling back User Shell Folders", regErr)
		// Rollback the first write
		if oldUserShell != "" {
			Sys.Registry.WriteString(sys.UserShellFoldersKey, regValueName, oldUserShell)
		}
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "注册表 Shell Folders 写入失败（已回滚 User Shell Folders）",
			Detail:  regErr.Error(),
			Action:  "文件已移动但注册表未更新，请以管理员身份运行",
		}
	}
	log.Info("Registry updated for %s", regValueName)

	// 8. Notify Shell.
	Sys.Shell.NotifyFolderRenamed(sourcePath, targetPath)
	Sys.Shell.NotifyDirUpdated(filepath.Dir(sourcePath))
	log.Info("Shell notifications sent")

	// 9. Mark backup complete.
	backup.Status = "complete"
	backup.NewPath = targetPath
	if err := saveBackup(backupPath, backup); err != nil {
		log.Warn("Failed to save backup completion record: %v — rollback may be affected", err)
		// intentional: migration succeeded; backup update is non-critical
	}

	if onProgress != nil {
		onProgress(regValueName, "complete", 0, 0, 0, 0)
	}

	log.Info("Migration of %s completed successfully", regValueName)
	return nil
}

// BackupRecord mirrors the structure from app_types.go but lives in the migrate package
// to avoid circular imports.
type BackupRecord struct {
	Version       int       `json:"version"`
	RegValueName  string    `json:"regValueName"`
	OriginalPath  string    `json:"originalPath"`
	NewPath       string    `json:"newPath"`
	RegistryValue string    `json:"registryValue"`
	CreatedAt     time.Time `json:"createdAt"`
	Status        string    `json:"status"`
}

// backupFilePath returns the path to the backup JSON file for a given folder.
func backupFilePath(regValueName string) string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	backupDir := filepath.Join(appData, "MoveKit", "backups")
	_ = os.MkdirAll(backupDir, 0700) //nolint:gosec // backupDir derived from trusted env vars
	return filepath.Join(backupDir, regValueName+".json")
}

// saveBackup marshals the backup record to JSON and writes it to disk.
func saveBackup(path string, record BackupRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// resolveJunctionTarget reads the reparse point to find where a junction points to.
// Falls back to checking if the directory is accessible (non-junction case).
func resolveJunctionTarget(junctionPath string) (string, error) {
	resolved, err := filepath.EvalSymlinks(junctionPath)
	if err != nil {
		return "", fmt.Errorf("resolve junction %s: %w", junctionPath, err)
	}
	return resolved, nil
}
