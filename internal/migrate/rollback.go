package migrate

import (
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
	if sys.IsJunction(rec.OriginalPath) {
		if err := sys.DeleteJunction(rec.OriginalPath); err != nil {
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
			Message: "Cannot create parent directory for rollback",
			Detail:  err.Error(),
			Action:  "Check write permissions on the original drive",
		}
	}

	if err := os.Rename(rec.NewPath, rec.OriginalPath); err != nil {
		return &MoveKitError{
			Code:    ErrCodeFileLocked,
			Message: "Failed to move files back during rollback",
			Detail:  err.Error(),
			Action:  "Close any programs using files in the target folder and retry",
		}
	}

	// 4. Restore registry values.
	// Write the original path back to User Shell Folders.
	if err := sys.WriteExpandString(sys.UserShellFoldersKey, rec.RegValueName, rec.OriginalPath); err != nil {
		return &MoveKitError{
			Code:    ErrCodeRegistryFailed,
			Message: "Failed to restore User Shell Folders registry during rollback",
			Detail:  err.Error(),
			Action:  "Run as administrator and manually restore the registry value",
		}
	}

	// Write the original path back to Shell Folders.
	absOriginal, _ := filepath.Abs(rec.OriginalPath)
	if err := sys.WriteString(sys.ShellFoldersKey, rec.RegValueName, absOriginal); err != nil {
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
