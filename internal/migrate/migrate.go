package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sys "github.com/movekit/movekit/internal/syscall"
)

// ProgressCallback is called during migration to report progress.
type ProgressCallback func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)

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
	sourcePath, err := sys.ReadExpandString(sys.UserShellFoldersKey, regValueName)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Cannot read current folder path from registry",
			Detail:  err.Error(),
			Action:  "Verify the folder is a known Windows folder",
		}
	}

	// Expand environment variables in the source path.
	sourcePath = os.ExpandEnv(sourcePath)
	log.Info("Source path: %s", maskPath(sourcePath))

	// 2. Chain migration: if source is already a junction, resolve to the real target.
	actualSource := sourcePath
	if sys.IsJunction(sourcePath) {
		log.Info("Source is already a junction, resolving real target")
		// Resolve the junction target by reading the reparse point.
		// For now we use the directory contents — the junction target is where
		// the files actually live. We'll read the reparse data to get the target.
		junctionTarget, err := resolveJunctionTarget(sourcePath)
		if err != nil {
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "Cannot resolve existing junction target",
				Detail:  err.Error(),
				Action:  "Check that the junction is not broken",
			}
		}
		actualSource = junctionTarget
		log.Info("Junction resolves to: %s", maskPath(actualSource))
	}

	// Determine target path.
	driveLetter := targetDrive
	if len(driveLetter) >= 2 && driveLetter[1] == ':' {
		driveLetter = driveLetter[:1]
	}
	folderName := filepath.Base(actualSource)
	targetPath := filepath.Join(driveLetter+":\\", folderName)
	log.Info("Target path: %s", maskPath(targetPath))

	// Check for junction hijack at target.
	if hijackErr := CheckJunctionHijack(targetPath); hijackErr != nil {
		return hijackErr
	}

	// 3. PreCheck.
	if err := PreCheck(actualSource, driveLetter); err != nil {
		if mkErr, ok := err.(*MoveKitError); ok && mkErr.Code == "SECURITY_SOFTWARE_DETECTED" {
			log.Warn("检测到安全软件: %s，继续迁移", mkErr.Warning)
			// Non-blocking: proceed with migration despite AV detection
		} else {
			log.Error("PreCheck failed: %s — %s", err.Code, err.Detail)
			return err
		}
	}

	// Check for file conflicts before moving.
	conflicts, _ := DetectConflicts(actualSource, targetPath) // intentional: conflicts are logged but do not block migration (per MVP behavior)
	if len(conflicts) > 0 {
		log.Warn("Found %d file conflicts", len(conflicts))
		// For MVP: log conflicts but proceed (user already confirmed)
		// Future: emit conflict event and wait for user decision
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
	if sys.IsCrossVolume(actualSource, targetPath) {
		// Cross-volume: copy + delete.
		if err := moveFilesCrossVolume(actualSource, targetPath, onProgress, regValueName); err != nil {
			log.Error("Cross-volume move failed: %v", err)
			return err
		}
	} else {
		// Same-volume: atomic rename.
		if err := os.Rename(actualSource, targetPath); err != nil {
			return &MoveKitError{
				Code:    ErrCodeFileLocked,
				Message: "Failed to move folder (files may be locked)",
				Detail:  err.Error(),
				Action:  "Close any programs using files in this folder and retry",
			}
		}
	}
	log.Info("Files moved successfully")

	// 6. Create junction at original location.
	// If source was already a junction, delete it first so we can create a new one.
	if sys.IsJunction(sourcePath) {
		if err := sys.DeleteJunction(sourcePath); err != nil {
			log.Warn("Failed to delete old junction: %v", err)
		}
	}

	// Ensure the source directory exists for the junction.
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Cannot create directory for junction",
			Detail:  err.Error(),
			Action:  "Check permissions on the source location",
		}
	}

	if err := sys.CreateJunction(sourcePath, targetPath); err != nil {
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Failed to create junction point",
			Detail:  err.Error(),
			Action:  "Ensure the source directory exists and you have write permission",
		}
	}
	log.Info("Junction created: %s -> %s", maskPath(sourcePath), maskPath(targetPath))

	// 7. Update registry.
	// User Shell Folders: REG_EXPAND_SZ (with %USERPROFILE% etc.)
	regValue := fmt.Sprintf("%s\\%s", driveLetter+":", folderName)
	if err := sys.WriteExpandString(sys.UserShellFoldersKey, regValueName, regValue); err != nil {
		log.Error("Failed to update User Shell Folders: %v", err)
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to update User Shell Folders registry",
			Detail:  err.Error(),
			Action:  "The files were moved but the registry was not updated. Run as administrator.",
		}
	}

	// Shell Folders: REG_SZ (fully expanded).
	absTarget, _ := filepath.Abs(targetPath) // filepath.Abs on an already-absolute path; fallback to targetPath
	if err := sys.WriteString(sys.ShellFoldersKey, regValueName, absTarget); err != nil {
		log.Error("Failed to update Shell Folders: %v", err)
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to update Shell Folders registry",
			Detail:  err.Error(),
			Action:  "The files were moved but the registry was not updated. Run as administrator.",
		}
	}
	log.Info("Registry updated for %s", regValueName)

	// 8. Notify Shell.
	sys.NotifyFolderRenamed(sourcePath, targetPath)
	sys.NotifyDirUpdated(filepath.Dir(sourcePath))
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
	os.MkdirAll(backupDir, 0755)
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
	// Walk into the junction to discover files — if it's a valid junction,
	// files inside are actually at the target location.
	// We use a heuristic: read the first file's real path to extract the volume.
	// A proper implementation would use DeviceIoControl/FSCTL_GET_REPARSE_POINT.

	// Simple approach: use filepath.EvalSymlinks which resolves junctions on Windows.
	resolved, err := filepath.EvalSymlinks(junctionPath)
	if err != nil {
		return "", fmt.Errorf("resolve junction %s: %w", junctionPath, err)
	}
	return resolved, nil
}
