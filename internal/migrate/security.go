package migrate

import (
	"os"
	"path/filepath"
	"strings"
)

// CheckJunctionHijack checks if the target path is already a junction or
// symlink, which could indicate a junction hijack attack or a conflicting
// migration. Returns nil if the target is safe to use.
func CheckJunctionHijack(targetPath string) *MoveKitError {
	// A junction/symlink at the target is always unsafe, regardless of what it
	// resolves to.
	if Sys.Junction.IsJunction(targetPath) {
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "目标路径是 NTFS Junction 点",
			Detail:  targetPath,
			Action:  "Remove the existing junction or choose a different target",
		}
	}

	// Resolve the target path to its real location.
	resolved, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		// Path doesn't exist or can't be resolved — that's fine.
		if os.IsNotExist(err) {
			return nil
		}
		return &MoveKitError{
			Code:    ErrCodePermissionDenied,
			Message: "无法评估目标路径",
			Detail:  err.Error(),
			Action:  "Verify the target path is accessible",
		}
	}

	// If the resolved path differs from the target, it's a symlink/junction.
	if !strings.EqualFold(resolved, targetPath) {
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "目标路径是 Junction 或符号链接",
			Detail:  targetPath + " resolves to " + resolved,
			Action:  "Choose a different target path that is not a junction",
		}
	}

	return nil
}
