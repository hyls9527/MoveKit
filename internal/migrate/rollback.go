package migrate

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	sys "github.com/movekit/movekit/internal/syscall"
)

// Rollback reverses a migration for the given folder name:
// 1. Loads the backup record
// 2. Deletes the junction at the original path (if present)
// 3. Moves files from the new location back to the original
// 4. Restores the original registry value
// 5. Marks the backup status as "rolled_back"
func Rollback(folderName string) error {
	// 1. Load backup record.
	rec, err := LoadBackup(folderName)
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Cannot load backup record for rollback",
			Detail:  err.Error(),
			Action:  "Verify that a backup exists for this folder",
		}
	}

	// 2. Delete junction at original path if it is one.
	if Sys.Junction.IsJunction(rec.OriginalPath) {
		if err := Sys.Junction.DeleteJunction(rec.OriginalPath); err != nil {
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "Failed to remove junction during rollback",
				Detail:  err.Error(),
				Action:  "Run as administrator and retry",
			}
		}
		// Remove the now-empty directory that hosted the junction.
		os.Remove(rec.OriginalPath)
	}

	// 3. Move files back from new location to original.
		// Ensure the parent of the original path exists.
		parentDir := filepath.Dir(rec.OriginalPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return &MoveKitError{
				Code:    ErrCodePermissionDenied,
				Message: "无法创建回滚目标目录",
				Detail:  err.Error(),
				Action:  "检查原盘的写入权限",
			}
		}

		// Detect cross-volume: if source and target are on different drives, use copy+delete
		if Sys.Volume.IsCrossVolume(rec.NewPath, rec.OriginalPath) {
			// Cross-volume rollback: copy each file then delete source
			if err := rollbackCrossVolume(rec.NewPath, rec.OriginalPath); err != nil {
				return err
			}
		} else {
			// Same-volume: atomic rename
			if err := os.Rename(rec.NewPath, rec.OriginalPath); err != nil {
				return &MoveKitError{
					Code:    ErrCodeFileLocked,
					Message: "回滚移动文件失败",
					Detail:  err.Error(),
					Action:  "关闭正在使用目标文件夹的程序后重试",
				}
			}
		}

	// 4. Restore registry values.
	// Write the original path back to User Shell Folders.
	if err := Sys.Registry.WriteExpandString(sys.UserShellFoldersKey, rec.RegValueName, rec.OriginalPath); err != nil {
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to restore User Shell Folders registry during rollback",
			Detail:  err.Error(),
			Action:  "Run as administrator and manually restore the registry value",
		}
	}

	// Write the original path back to Shell Folders.
	absOriginal, _ := filepath.Abs(rec.OriginalPath)
	if err := Sys.Registry.WriteString(sys.ShellFoldersKey, rec.RegValueName, absOriginal); err != nil {
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to restore Shell Folders registry during rollback",
			Detail:  err.Error(),
			Action:  "Run as administrator and manually restore the registry value",
		}
	}

	// 5. Update backup status.
		rec.Status = "rolled_back"
		if err := SaveBackup(*rec); err != nil {
			// Non-fatal: the rollback succeeded even if we can't update the record.
			return nil
		}

		return nil
	}

	// rollbackCrossVolume moves files back across volumes using copy+delete.
	func rollbackCrossVolume(src, dst string) error {
		// Walk source, copy each file to destination, then delete source
		err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			dstPath := filepath.Join(dst, relPath)

			if d.IsDir() {
				return os.MkdirAll(dstPath, 0755)
			}

			// Copy file
			if err := Sys.File.CopyFileWithACL(path, dstPath, nil); err != nil {
				return fmt.Errorf("copy %s: %w", relPath, err)
			}

			// Remove source
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("remove source %s: %w", relPath, err)
			}

			return nil
		})

		if err != nil {
			return &MoveKitError{
				Code:    ErrCodeFileLocked,
				Message: "跨卷回滚失败",
				Detail:  err.Error(),
				Action:  "关闭正在使用文件夹的程序后重试",
			}
		}

		// Clean up empty source directory
		os.RemoveAll(src)
		return nil
	}
