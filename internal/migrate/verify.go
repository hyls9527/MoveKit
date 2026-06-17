package migrate

import (
	"fmt"
	"os"
	"path/filepath"
)

// VerifyIntegrity checks that the target directory contains the expected
// number of files (sourceCount) by walking the target path. Returns an
// error if the counts don't match.
func VerifyIntegrity(sourceCount int, sourceSize int64, targetPath string) error {
	var targetCount int
	var targetSize int64

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			targetCount++
			targetSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "Cannot walk target directory for integrity check",
			Detail:  err.Error(),
			Action:  "Verify the target directory is accessible",
		}
	}

	if targetCount != sourceCount {
		return &MoveKitError{
			Code:    ErrCodeFileLocked,
			Message: fmt.Sprintf("File count mismatch: expected %d, got %d", sourceCount, targetCount),
			Detail:  fmt.Sprintf("source had %d files (%d bytes), target has %d files (%d bytes)", sourceCount, sourceSize, targetCount, targetSize),
			Action:  "Some files may not have been copied. Retry the migration.",
		}
	}

	return nil
}
