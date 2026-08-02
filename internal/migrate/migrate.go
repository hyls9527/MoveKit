package migrate

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/movekit/movekit/internal/folders"
	sys "github.com/movekit/movekit/internal/syscall"
)

// ProgressCallback is called during migration to report progress.
type ProgressCallback func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)

// CleanupTempJunctions removes any leftover .movekit_tmp junctions from previous failed migrations.
// Call this on application startup.
func CleanupTempJunctions() {
	for _, f := range folders.AllFolders {
		rawPath, err := Sys.Registry.ReadExpandString(sys.UserShellFoldersKey, f.RegValueName)
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

// testTargetRoot redirects the computed target path during unit tests so the
// full Migrate pipeline can run against temp directories. Production code
// never sets this variable.
var testTargetRoot string

// ConflictMode determines how file conflicts are handled during migration.
type ConflictMode string

const (
	ConflictSkip      ConflictMode = "skip"
	ConflictOverwrite ConflictMode = "overwrite"
)

// Migrate moves a known folder to a new drive and creates a junction at the original location.
//
// Parameters:
//   - regValueName: the Shell Folders registry value name (e.g. "Desktop", "Personal")
//   - targetDrive: the target drive letter (e.g. "D" or "D:")
//   - onProgress: optional progress callback (may be nil)
//   - conflictMode: how to handle file conflicts at the target ("skip" or "overwrite")
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
	if len(driveLetter) != 1 || !isASCIIDriveLetter(driveLetter[0]) {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "无效的目标盘符",
			Detail:  fmt.Sprintf("收到: %q", targetDrive),
			Action:  "请选择 A-Z 之间的盘符",
		}
	}
	driveLetter = strings.ToUpper(driveLetter)
	folderName := filepath.Base(actualSource)
	targetPath := filepath.Join(driveLetter+":\\", folderName)
	if testTargetRoot != "" {
		targetPath = filepath.Join(testTargetRoot, folderName)
	}
	log.Info("Target path: %s", maskPath(targetPath))

	// 2b. Already-migrated no-op: when the real folder location is exactly the
	// target, there is nothing to move (the source path is a junction pointing
	// here, or the folder already lives on the target drive). Returning early
	// keeps batch re-runs and retries after partial failures safe. An existing
	// backup record is left untouched so rollback still knows the original
	// location; a self-referencing record would make rollback impossible.
	if strings.EqualFold(filepath.Clean(targetPath), filepath.Clean(actualSource)) {
		log.Info("Folder already resides at %s, migration is a no-op", maskPath(targetPath))
		if _, loadErr := LoadBackup(regValueName); loadErr != nil {
			// No backup record exists (e.g. the folder was moved by other
			// means); there is nothing meaningful to record for rollback.
			log.Info("No backup record for %s; nothing to do", regValueName)
		}
		return nil
	}

	// Check for junction hijack at target.
	if hijackErr := CheckJunctionHijack(targetPath); hijackErr != nil {
		return hijackErr
	}

	// 3. PreCheck.
	if preErr := PreCheck(actualSource, driveLetter); preErr != nil {
		if preErr.Code == ErrCodeSecuritySoftware || preErr.Code == ErrCodeBatteryWarning {
			log.Warn("非阻塞警告 %s: %s，继续迁移", preErr.Code, preErr.Warning)
			// Non-blocking: proceed with migration despite AV/battery warning
		} else {
			log.Error("PreCheck failed: %s — %s", preErr.Code, preErr.Detail)
			return preErr
		}
	}

	// Check for file conflicts before moving.
	// Skip if source is already a junction (re-migration): files are already at target.
	if !Sys.Junction.IsJunction(sourcePath) {
		if _, statErr := os.Stat(targetPath); statErr == nil {
			// PreCheck already rejects an existing target directory, so this
			// only runs in the TOCTOU window (target created meanwhile).
			conflicts, conflictErr := DetectConflicts(actualSource, targetPath)
			if conflictErr != nil {
				log.Warn("Conflict detection failed: %v", conflictErr)
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
	// If any step fails after the move begins, restore the original state:
	// remove a newly-created junction first, then merge any files that were
	// already copied back into the source directory.
	moveStarted := false
	defer func() {
		if moveStarted && retErr != nil {
			log.Error("Migration step failed after files were moved (%v), restoring original state", retErr)
			if rbErr := restoreAfterFailedMove(sourcePath, actualSource, targetPath); rbErr != nil {
				log.Error("CRITICAL: restore after failed move also failed: %v", rbErr)
			} else {
				log.Info("Files restored after failed migration")
				backup.Status = "rolled_back"
				if saveErr := saveBackup(backupPath, backup); saveErr != nil {
					log.Warn("Failed to mark backup as rolled_back: %v", saveErr)
				}
			}
		}
	}()

	if onProgress != nil {
		onProgress(regValueName, "moving", 0, 0, 0, 0)
	}
	moveStarted = true

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

	// User Shell Folders: REG_EXPAND_SZ to preserve the value type this key
	// conventionally uses (the written path may contain no env vars, which is
	// fine for EXPAND_SZ).
	regValue := fmt.Sprintf("%s\\%s", driveLetter+":", folderName)
	if regErr := Sys.Registry.WriteExpandString(sys.UserShellFoldersKey, regValueName, regValue); regErr != nil {
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
	return os.WriteFile(path, data, 0600)
}

// resolveJunctionTarget reads the reparse point to find where a junction
// points to.
func resolveJunctionTarget(junctionPath string) (string, error) {
	return Sys.Junction.ResolveTarget(junctionPath)
}

// restoreAfterFailedMove returns the filesystem to its pre-move state after a
// migration step failed. It first removes a junction that was created at the
// original path (only when it points to our target), then merges any files
// left at the target back into the source directory. Source files that never
// left are authoritative: their partial/duplicate copies at the target are
// discarded instead of overwriting the original.
func restoreAfterFailedMove(sourcePath, actualSource, targetPath string) error {
	if strings.EqualFold(filepath.Clean(targetPath), filepath.Clean(actualSource)) {
		return nil
	}

	// 1. Remove a junction created at the original path, but only if it points
	// to our target. Never delete a pre-existing junction pointing elsewhere.
	if Sys.Junction.IsJunction(sourcePath) {
		resolved, err := resolveJunctionTarget(sourcePath)
		if err != nil {
			return fmt.Errorf("resolve junction at %s before restore: %w", sourcePath, err)
		}
		if !strings.EqualFold(resolved, targetPath) {
			return fmt.Errorf("junction at %s points to %s, expected %s; refusing to delete", sourcePath, resolved, targetPath)
		}
		if err := Sys.Junction.DeleteJunction(sourcePath); err != nil {
			return fmt.Errorf("delete junction at %s before restore: %w", sourcePath, err)
		}
		_ = os.Remove(sourcePath) // remove the empty directory that hosted the junction
	}

	// 2. Nothing was moved if the target never appeared (e.g. atomic
	// same-volume rename failed before creating it).
	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// 3. Merge the target contents back into the source directory.
	if err := mergeDirBack(targetPath, actualSource); err != nil {
		return err
	}
	return os.RemoveAll(targetPath)
}

// mergeDirBack moves every file under src into dst. If the same relative file
// already exists at dst, the destination copy wins and the src copy is
// discarded (the source file never left, so it is authoritative).
func mergeDirBack(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0700)
		}

		if _, err := os.Stat(dstPath); err == nil {
			// The source file still exists, so the copy at the target is a
			// duplicate (or a partial copy). Source wins.
			return os.Remove(path) //nolint:gosec // G122: restore walk; symlinks skipped, path derives from WalkDir
		} else if !os.IsNotExist(err) {
			return err
		}

		if err := copyFilePlain(path, dstPath); err != nil {
			return fmt.Errorf("restore copy %s: %w", relPath, err)
		}
		if err := os.Remove(path); err != nil { //nolint:gosec // G122: restore walk; symlinks skipped
			return fmt.Errorf("restore remove %s: %w", relPath, err)
		}
		return nil
	})
}

// copyFilePlain copies a single file without ACL preservation. It is used
// only during restore-after-failure, where the priority is returning data to
// its original location even if the ACL-aware copy path is itself broken.
func copyFilePlain(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // G304: restore-only helper; src is a WalkDir path under the migration target
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}
	out, err := os.Create(dst) //nolint:gosec // G304: restore-only helper; dst is under the original source path
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
