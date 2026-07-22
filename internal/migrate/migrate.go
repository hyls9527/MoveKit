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

// Migrate moves a known folder to a new drive and creates a junction at the original location.
//
// Parameters:
//   - regValueName: the Shell Folders registry value name (e.g. "Desktop", "Personal")
//   - targetDrive: the target drive letter (e.g. "D" or "D:")
//   - onProgress: optional progress callback (may be nil)
// ConflictMode determines how file conflicts are handled during migration.
type ConflictMode string

const (
	ConflictSkip      ConflictMode = "skip"
	ConflictOverwrite ConflictMode = "overwrite"
)

func Migrate(regValueName, targetDrive string, onProgress ProgressCallback, conflictMode ConflictMode) (retErr error) {
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
	if Sys.Junction.IsJunction(sourcePath) {
		log.Info("Source is already a junction, resolving real target")
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
	conflicts, _ := DetectConflicts(actualSource, targetPath) // intentional: conflicts are logged but do not block migration (per MVP behavior)
	if len(conflicts) > 0 {
		log.Warn("Found %d file conflicts", len(conflicts))
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
	// If the file move succeeds but a later step (junction/registry) fails,
	// we must move the files back to avoid leaving them orphaned at the target.
	var filesMovedOk bool
	defer func() {
		if filesMovedOk && retErr != nil {
			log.Error("Post-move step failed (%v), rolling back file move from %s to %s",
				retErr, maskPath(targetPath), maskPath(actualSource))
			if rbErr := rollbackMoveFiles(targetPath, actualSource); rbErr != nil {
				log.Error("CRITICAL: rollback of file move also failed: %v", rbErr)
			} else {
				log.Info("File move rollback succeeded")
			}
		}
	}()

	if onProgress != nil {
		onProgress(regValueName, "moving", 0, 0, 0, 0)
	}

	log.Info("Moving files from %s to %s", maskPath(actualSource), maskPath(targetPath))
	if Sys.Volume.IsCrossVolume(actualSource, targetPath) {
		// Cross-volume: copy + delete.
		if err := moveFilesCrossVolume(actualSource, targetPath, onProgress, regValueName, conflictMode); err != nil {
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
	filesMovedOk = true
	log.Info("Files moved successfully")

	// 6. Create junction at original location.
	// Strategy: create at temp path first, verify, then swap. This minimizes
	// the window where the junction is missing (crash-safe).
	tmpJunction := sourcePath + ".movekit_tmp"

	// Clean up any leftover temp junction from a previous crash.
	if _, statErr := os.Stat(tmpJunction); statErr == nil {
		os.RemoveAll(tmpJunction)
	}

	if mdErr := os.MkdirAll(tmpJunction, 0700); mdErr != nil {
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Cannot create temp directory for junction",
			Detail:  mdErr.Error(),
			Action:  "Check permissions on the source location",
		}
	}

	if err := Sys.Junction.CreateJunction(tmpJunction, targetPath); err != nil {
		os.RemoveAll(tmpJunction)
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Failed to create junction point",
			Detail:  err.Error(),
			Action:  "Ensure the source directory exists and you have write permission",
		}
	}

	// Verify the temp junction actually points to our target.
	if resolved, resErr := resolveJunctionTarget(tmpJunction); resErr != nil || !pathsEqual(resolved, targetPath) {
		Sys.Junction.DeleteJunction(tmpJunction)
		os.RemoveAll(tmpJunction)
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Junction verification failed",
			Detail:  fmt.Sprintf("expected %s, got %s (err: %v)", targetPath, resolved, resErr),
			Action:  "Retry the migration",
		}
	}

	// Now swap: delete old junction, rename temp to final.
	if Sys.Junction.IsJunction(sourcePath) {
		if err := Sys.Junction.DeleteJunction(sourcePath); err != nil {
			Sys.Junction.DeleteJunction(tmpJunction)
			os.RemoveAll(tmpJunction)
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "Failed to remove old junction during swap",
				Detail:  err.Error(),
				Action:  "Run as administrator and retry",
			}
		}
		os.Remove(sourcePath) // remove empty dir left by junction deletion
	}

	if err := os.Rename(tmpJunction, sourcePath); err != nil {
		// Fallback: try mklink /J via cmd.
		Sys.Junction.DeleteJunction(tmpJunction)
		os.RemoveAll(tmpJunction)
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Failed to finalize junction",
			Detail:  err.Error(),
			Action:  "Run as administrator and retry",
		}
	}
	log.Info("Junction created: %s -> %s", maskPath(sourcePath), maskPath(targetPath))

	// 7. Update registry — dual-key write with rollback.
	// User Shell Folders and Shell Folders are separate keys.
	// Both must succeed or we roll back the first write.
	regValue := fmt.Sprintf("%s\\%s", driveLetter+":", folderName)

	// Backup old values for rollback.
	oldUserShell, _ := Sys.Registry.ReadExpandString(sys.UserShellFoldersKey, regValueName)
	oldShell, _ := Sys.Registry.ReadExpandString(sys.ShellFoldersKey, regValueName)

	if regErr := Sys.Registry.WriteString(sys.UserShellFoldersKey, regValueName, regValue); regErr != nil {
		log.Error("Failed to update User Shell Folders: %v", regErr)
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to update User Shell Folders registry",
			Detail:  regErr.Error(),
			Action:  "The files were moved but the registry was not updated. Run as administrator.",
		}
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath // fallback if Abs fails
	}
	if err := Sys.Registry.WriteString(sys.ShellFoldersKey, regValueName, absTarget); err != nil {
		log.Error("Failed to update Shell Folders: %v — rolling back User Shell Folders", err)
		// Roll back the first write.
		if rbErr := Sys.Registry.WriteString(sys.UserShellFoldersKey, regValueName, oldUserShell); rbErr != nil {
			log.Error("CRITICAL: Rollback of User Shell Folders also failed: %v", rbErr)
		} else if oldShell != "" {
			Sys.Registry.WriteString(sys.ShellFoldersKey, regValueName, oldShell)
		}
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to update Shell Folders registry",
			Detail:  err.Error(),
			Action:  "The files were moved but the registry was not updated. Run as administrator.",
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

// pathsEqual does a path comparison for Windows (case-insensitive after Clean).
func pathsEqual(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// rollbackMoveFiles moves a directory tree from src back to dst.
// Used when the file move succeeds but a subsequent step (junction/registry)
// fails — this prevents files from being orphaned at the target.
func rollbackMoveFiles(src, dst string) error {
	// Ensure dst parent exists.
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("rollback: create parent dir %s: %w", filepath.Dir(dst), err)
	}

	if Sys.Volume.IsCrossVolume(src, dst) {
		return rollbackCrossVolume(src, dst)
	}
	// Same volume: os.Rename is atomic on NTFS.
	return os.Rename(src, dst)
}
