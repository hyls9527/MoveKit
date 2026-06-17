package migrate

import (
	"os"
	"path/filepath"
)

// DetectConflicts walks the source directory and checks if any files with the
// same relative path already exist at the target. Returns a slice of ConflictInfo
// for each conflicting file.
func DetectConflicts(source, target string) ([]ConflictInfo, error) {
	var conflicts []ConflictInfo

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		targetFile := filepath.Join(target, relPath)
		dstInfo, err := os.Stat(targetFile)
		if os.IsNotExist(err) {
			return nil // no conflict
		}
		if err != nil {
			return err
		}

		conflicts = append(conflicts, ConflictInfo{
			RelativePath: relPath,
			SrcSize:      info.Size(),
			DstSize:      dstInfo.Size(),
			SrcModTime:   info.ModTime().Format("2006-01-02T15:04:05Z"),
			DstModTime:   dstInfo.ModTime().Format("2006-01-02T15:04:05Z"),
		})

		return nil
	})

	return conflicts, err
}
