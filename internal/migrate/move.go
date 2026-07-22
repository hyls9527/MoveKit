package migrate

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
)

// cancelled is an atomic flag used to signal cancellation of an in-progress
// file move operation.
var cancelled atomic.Int32

// CancelMigration sets the cancellation flag, causing any in-progress
// cross-volume file move to stop after the current file completes.
func CancelMigration() {
	cancelled.Store(1)
}

// IsCancelled returns true if CancelMigration has been called.
func IsCancelled() bool {
	return cancelled.Load() != 0
}

// ResetCancellation clears the cancellation flag. Call this before starting
// a new migration.
func ResetCancellation() {
	cancelled.Store(0)
}

// moveFilesSameVolume moves a directory on the same volume using os.Rename.
// This is an atomic operation on NTFS.
func moveFilesSameVolume(source, target string, onProgress func(ProgressInfo)) error {
	if onProgress != nil {
		onProgress(ProgressInfo{
			Phase: "moving",
		})
	}

	// Verify source is accessible before attempting rename.
	if _, err := os.Stat(source); err != nil {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "Cannot access source folder before move",
			Detail:  err.Error(),
			Action:  "Verify the source folder exists and is accessible",
		}
	}

	if err := os.Rename(source, target); err != nil {
		// NTFS same-volume rename is atomic — verify source state.
		if _, statErr := os.Stat(source); os.IsNotExist(statErr) {
			// Source is gone but rename reported failure — inconsistent.
			// This should never happen on NTFS.
			return &MoveKitError{
				Code:    "INTERNAL_ERROR",
				Message: "文件移动状态不一致（源文件夹消失但移动报告失败）",
				Detail:  fmt.Sprintf("rename err: %v; source no longer exists", err),
				Action:  "请检查目标路径是否有部分文件，可能需要手动恢复",
			}
		}
		return &MoveKitError{
			Code:    ErrCodeFileLocked,
			Message: "Failed to move folder (files may be locked)",
			Detail:  err.Error(),
			Action:  "Close any programs using files in this folder and retry",
		}
	}

	if onProgress != nil {
		onProgress(ProgressInfo{
			Phase: "complete",
		})
	}
	return nil
}

// moveFilesCrossVolume moves files across different drive volumes by copying
// each file individually and then deleting the source. Supports cancellation
// via the atomic cancelled flag and reports progress through the callback.
func moveFilesCrossVolume(src, dst string, onProgress ProgressCallback, folder string, conflictMode ConflictMode) error {
	ResetCancellation()

	// Count total files for progress reporting.
	var totalFiles int
	var totalBytes int64
	filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // intentional: skip files we cannot access during counting
		}
		// Skip symlinks
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if !d.IsDir() {
			totalFiles++
			if info, err := d.Info(); err == nil {
				totalBytes += info.Size()
			} // d.Info() may fail; skip byte tracking but don't fail counting
		}
		return nil
	})

	var filesMoved int
	var bytesMoved int64

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip symlinks
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		// Check cancellation.
		if IsCancelled() {
			return fmt.Errorf("migration cancelled")
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Handle file conflicts based on conflict mode.
		if _, statErr := os.Stat(dstPath); statErr == nil {
			switch conflictMode {
			case ConflictSkip:
				// Skip this file — do not copy, do not remove source.
				return nil
			case ConflictOverwrite:
				// Remove existing target file before copying.
				if err := os.Remove(dstPath); err != nil {
					return fmt.Errorf("remove existing target %s: %w", relPath, err)
				}
			}
		}

		// Copy the file using Windows CopyFileExW with ACL preservation.
		if err := Sys.File.CopyFileWithACL(path, dstPath, nil); err != nil {
			return fmt.Errorf("copy %s: %w", relPath, err)
		}

		// Remove source file after successful copy.
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove source %s: %w", relPath, err)
		}

		filesMoved++
		fi, _ := d.Info() // d.Info() may fail; skip byte tracking but don't fail the move
		if fi != nil {
			bytesMoved += fi.Size()
		}

		// Report progress.
		if onProgress != nil {
			onProgress(folder, "moving", filesMoved, totalFiles, bytesMoved, totalBytes)
		}

		return nil
	})

	if err != nil && IsCancelled() {
		return &MoveKitError{
			Code:    ErrCodeFileLocked,
			Message: "Migration was cancelled",
			Detail:  err.Error(),
			Action:  "The migration was interrupted. You can retry or roll back.",
		}
	}

	// Clean up empty source directory tree after all files are moved.
	if err == nil {
		os.RemoveAll(src) // non-critical: empty dirs may remain on error
	}

	return err
}
