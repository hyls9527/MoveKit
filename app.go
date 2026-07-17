package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/movekit/movekit/internal/folders"
	"github.com/movekit/movekit/internal/migrate"
	sys "github.com/movekit/movekit/internal/syscall"
)

// App is the main Wails application struct. All exported methods are
// automatically bound to the frontend via Wails bindings.
type App struct {
	ctx       context.Context
	mu        sync.Mutex
	migrating bool
	sizeCache sync.Map // path -> int64, caches dirSize results
}

// cachedDirSize returns directory size with caching to avoid repeated walks.
func (a *App) cachedDirSize(path string) (int64, error) {
	if cached, ok := a.sizeCache.Load(path); ok {
		return cached.(int64), nil
	}
	size, err := migrate.DirSize(path)
	if err != nil {
		return 0, err
	}
	a.sizeCache.Store(path, size)
	return size, nil
}

// clearSizeCache clears the size cache (call after migration completes).
func (a *App) clearSizeCache() {
	a.sizeCache = sync.Map{}
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called by Wails when the application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called by Wails when the application shuts down.
func (a *App) shutdown(ctx context.Context) {}

// ---------------------------------------------------------------------------
// Health & Version
// ---------------------------------------------------------------------------

// HealthCheck returns true if the application is running correctly.
func (a *App) HealthCheck() bool {
	return true
}

// GetVersion returns the current application version string.
func (a *App) GetVersion() string {
	return currentVersion
}

// ---------------------------------------------------------------------------
// Folder State
// ---------------------------------------------------------------------------

// GetFolders returns the current state of all 5 known folders by reading
// registry values and checking junction status.
func (a *App) GetFolders() []FolderInfo {
	result := make([]FolderInfo, 0, len(folders.AllFolders))
	for _, f := range folders.AllFolders {
		fi := FolderInfo{
			Name: f.Name,
			GUID: f.GUID,
		}

		// Read current path from registry.
		rawPath, err := sys.ReadExpandString(sys.UserShellFoldersKey, f.RegValueName)
		if err != nil {
			fi.Status = "unknown"
			result = append(result, fi)
			continue
		}
		expanded := os.ExpandEnv(rawPath)
		fi.SourcePath = expanded
		fi.TargetPath = expanded

		// Calculate folder size (lazy — may take a moment for large folders).
				if size, err := a.cachedDirSize(expanded); err == nil {
			fi.SizeBytes = size
		} else {
			fi.SizeBytes = -1 // sentinel: directory size could not be calculated
		}

		// Check if it is a junction (i.e. already migrated).
		if sys.IsJunction(expanded) {
			fi.Status = "migrated"
			// Resolve the junction target as the real location.
			resolved, err := filepath.EvalSymlinks(expanded)
			if err == nil {
				fi.TargetPath = resolved
			}
		} else {
			fi.Status = "default"
			// If already on a non-C drive, no need to migrate
			if len(expanded) >= 2 && expanded[1] == ':' && !strings.EqualFold(expanded[:1], "C") {
				fi.Status = "not_on_c"
			}
		}

		result = append(result, fi)
	}
	return result
}

// ---------------------------------------------------------------------------
// Disk Information
// ---------------------------------------------------------------------------

// GetDisks returns information about available fixed drives (C/D/E/F).
func (a *App) GetDisks() []DiskInfo {
	drives := []string{"C", "D", "E", "F"}
	var result []DiskInfo
	for _, drive := range drives {
		volInfo, err := sys.GetVolumeInfo(drive)
		if err != nil {
			continue // drive doesn't exist or isn't accessible
		}
		freeBytes, totalBytes, _, err := sys.GetDiskFreeSpace(drive)
		if err != nil {
			continue
		}
		result = append(result, DiskInfo{
			Drive:       drive + ":",
			Label:       volInfo.FileSystem, // Label comes from volume name; FileSystem as proxy
			TotalBytes:  totalBytes,
			FreeBytes:   freeBytes,
			IsFixedDrive:       volInfo.IsFixedDrive,
			IsRemovable: volInfo.IsRemovable,
			IsNetwork:   volInfo.IsNetwork,
			FileSystem:  volInfo.FileSystem,
		})
	}
	return result
}

// ---------------------------------------------------------------------------
// Duration Estimation
// ---------------------------------------------------------------------------

// EstimateDuration returns an estimated migration time based on total folder
// sizes and whether the operation is cross-volume.
func (a *App) EstimateDuration(targetDrive string) DurationEstimate {
	// Calculate total bytes across all known folders.
	var totalBytes int64
	for _, f := range folders.AllFolders {
		rawPath, err := sys.ReadExpandString(sys.UserShellFoldersKey, f.RegValueName)
		if err != nil {
			continue
		}
		expanded := os.ExpandEnv(rawPath)
		size, err := a.cachedDirSize(expanded)
				if err != nil {
					// intentional: skip folders we cannot measure; estimate will be conservative
					continue
				}
				totalBytes += size
	}

	// Determine if cross-volume by checking the first folder's source drive.
	isCrossVolume := false
	if len(folders.AllFolders) > 0 {
		rawPath, err := sys.ReadExpandString(sys.UserShellFoldersKey, folders.AllFolders[0].RegValueName)
		if err == nil {
			expanded := os.ExpandEnv(rawPath)
			srcDrive := filepath.VolumeName(expanded)
			tgtDrive := targetDrive
			if len(tgtDrive) >= 2 && tgtDrive[1] == ':' {
				tgtDrive = tgtDrive[:1]
			}
			isCrossVolume = len(srcDrive) > 0 && srcDrive[:1] != tgtDrive
		}
	}

	// Check if target is SSD.
	targetIsFixedDrive := false
	tgtLetter := targetDrive
	if len(tgtLetter) >= 2 && tgtLetter[1] == ':' {
		tgtLetter = tgtLetter[:1]
	}
	volInfo, err := sys.GetVolumeInfo(tgtLetter)
	if err == nil {
		targetIsFixedDrive = volInfo.IsFixedDrive
	}

	// Heuristic estimation:
	// - Same volume rename: ~2s
	// - Cross-volume: ~50 MB/s (conservative; SSD/HDD distinction not implemented)
	estimatedSec := 2
	if isCrossVolume {
		speedMBps := 50 // Conservative: SSD/HDD distinction not implemented
		totalMB := totalBytes / (1024 * 1024)
		if totalMB > 0 {
			estimatedSec = int(totalMB/int64(speedMBps)) + 1
		}
	}

	estimatedText := formatDuration(estimatedSec)
	return DurationEstimate{
		IsCrossVolume: isCrossVolume,
		TargetIsFixedDrive:   targetIsFixedDrive,
		EstimatedSec:  estimatedSec,
		EstimatedText: estimatedText,
	}
}


func formatDuration(sec int) string {
	if sec < 60 {
		return fmt.Sprintf("%d秒", sec)
	}
	minutes := sec / 60
	seconds := sec % 60
	if minutes < 60 {
		if seconds > 0 {
			return fmt.Sprintf("%d分%d秒", minutes, seconds)
		}
		return fmt.Sprintf("%d分钟", minutes)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%d小时%d分钟", hours, minutes)
}

// ---------------------------------------------------------------------------
// OneDrive KFM Detection
// ---------------------------------------------------------------------------

// CheckOneDriveKFM checks whether OneDrive Known Folder Move is active.
func (a *App) CheckOneDriveKFM() OneDriveStatus {
	status, err := migrate.DetectOneDriveKFM()
	if err != nil {
		return OneDriveStatus{
			IsActive:    false,
			Folders:     nil,
			HasGPOBlock: false,
		}
	}
	return OneDriveStatus{
		IsActive:    status.IsActive,
		Folders:     status.Folders,
		HasGPOBlock: status.HasGPOBlock,
	}
}

// ---------------------------------------------------------------------------
// Migration
// ---------------------------------------------------------------------------

// StartMigration begins migrating all known folders to the target drive.
// Runs in a goroutine and emits "migration:progress" and "migration:done"
// events via Wails runtime.
func (a *App) StartMigration(targetDrive string) {
	a.mu.Lock()
	if a.migrating {
		a.mu.Unlock()
		return
	}
	a.migrating = true
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.migrating = false
			a.mu.Unlock()
		}()

		migrate.ResetCancellation()

		total := len(folders.AllFolders)
		for i, f := range folders.AllFolders {
			if migrate.IsCancelled() {
				runtime.EventsEmit(a.ctx, "migration:done", map[string]interface{}{
					"success": false,
					"error":   "Migration was cancelled",
				})
				return
			}

			// Emit folder-level progress event.
			runtime.EventsEmit(a.ctx, "migration:progress", ProgressInfo{
				Folder:     f.Name,
				Phase:      "starting",
				FilesMoved: 0,
				TotalFiles: 0,
				BytesMoved: 0,
				TotalBytes: 0,
				Percent:    (i * 100) / total,
			})

			// Create a progress callback that maps migrate.ProgressInfo → main.ProgressInfo
			// and emits events.
			folderName := f.Name
			folderIdx := i
			onProgress := func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64) {
				percent := 0
				if totalBytes > 0 {
					percent = int(bytesMoved * 100 / totalBytes)
				}
				// Scale percent to overall progress.
				overallPercent := ((folderIdx * 100) + percent) / total
				runtime.EventsEmit(a.ctx, "migration:progress", ProgressInfo{
					Folder:     folderName,
					Phase:      phase,
					FilesMoved: filesMoved,
					TotalFiles: totalFiles,
					BytesMoved: bytesMoved,
					TotalBytes: totalBytes,
					Percent:    overallPercent,
				})
			}

			err := migrate.Migrate(f.RegValueName, targetDrive, onProgress)
			if err != nil {
				errMsg := err.Error()
				var code, detail, action, warning string
				if mkErr, ok := err.(*migrate.MoveKitError); ok {
					code = mkErr.Code
					detail = mkErr.Detail
					action = mkErr.Action
					warning = mkErr.Warning
				}
				runtime.EventsEmit(a.ctx, "migration:done", map[string]interface{}{
					"success": false,
					"folder":  f.Name,
					"error":   errMsg,
					"code":    code,
					"detail":  detail,
					"action":  action,
					"warning": warning,
				})
				return
			}
		}

		a.clearSizeCache()
				runtime.EventsEmit(a.ctx, "migration:done", map[string]interface{}{
					"success": true,
				})
	}()
}

// CancelMigration cancels an in-progress migration.
func (a *App) CancelMigration() {
	migrate.CancelMigration()
}

// ---------------------------------------------------------------------------
// Rollback
// ---------------------------------------------------------------------------

// Rollback reverses the migration for a specific folder.
func (a *App) Rollback(regValueName string) error {
	return migrate.Rollback(regValueName)
}

// ---------------------------------------------------------------------------
// Integrity Check
// ---------------------------------------------------------------------------

// StartupIntegrityCheck checks all backed-up folders for junction integrity.
// Returns a list of issues found (empty slice if everything is OK).
func (a *App) StartupIntegrityCheck() []IntegrityIssue {
	var issues []IntegrityIssue

	backups, err := migrate.LoadAllBackups()
	if err != nil {
		issues = append(issues, IntegrityIssue{
			Type:    "error",
			Folder:  "",
			Message: "Cannot load backup records: " + err.Error(),
		})
		return issues
	}

	for _, rec := range backups {
		if rec.Status != "complete" {
			continue
		}

		// Check that the original path is a valid junction pointing to the new path.
		if !sys.IsJunction(rec.OriginalPath) {
			issues = append(issues, IntegrityIssue{
				Type:    "junction_broken",
				Folder:  rec.RegValueName,
				Message: fmt.Sprintf("Junction missing at %s", rec.OriginalPath),
			})
			continue
		}

		// Verify the junction target is accessible.
		resolved, err := filepath.EvalSymlinks(rec.OriginalPath)
		if err != nil {
			issues = append(issues, IntegrityIssue{
				Type:    "junction_broken",
				Folder:  rec.RegValueName,
				Message: fmt.Sprintf("Junction at %s cannot be resolved: %v", rec.OriginalPath, err),
			})
			continue
		}

		// Verify the resolved path matches the expected new path.
		// Use case-insensitive comparison for Windows paths.
		if !pathsEqual(resolved, rec.NewPath) {
			issues = append(issues, IntegrityIssue{
				Type:    "junction_mismatch",
				Folder:  rec.RegValueName,
				Message: fmt.Sprintf("Junction at %s points to %s, expected %s", rec.OriginalPath, resolved, rec.NewPath),
			})
		}
	}

	return issues
}

// pathsEqual does a path comparison for Windows (case-sensitive after Clean).
func pathsEqual(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

// ---------------------------------------------------------------------------
// First Run Flag / Config
// ---------------------------------------------------------------------------

// configDir returns %APPDATA%\MoveKit.
func configDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	return filepath.Join(appData, "MoveKit")
}

// configFilePath returns the path to the MoveKit config.json.
func configFilePath() string {
	return filepath.Join(configDir(), "config.json")
}

// configData is the structure of config.json.
type configData struct {
	FirstRun bool `json:"firstRun"`
}

// GetFirstRunFlag reads the first-run flag from config.json.
// Returns true if this is the first run (or if the config doesn't exist yet).
func (a *App) GetFirstRunFlag() bool {
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		return true // config doesn't exist → first run
	}
	var cfg configData
	if err := json.Unmarshal(data, &cfg); err != nil {
		return true // malformed config → treat as first run
	}
	return cfg.FirstRun
}

// SetFirstRunFlag writes the first-run flag to config.json.
func (a *App) SetFirstRunFlag(firstRun bool) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	cfg := configData{FirstRun: firstRun}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(configFilePath(), data, 0644)
}

// ---------------------------------------------------------------------------
// Conflict Detection
// ---------------------------------------------------------------------------

// DetectConflicts checks for file conflicts between source and target paths.
func (a *App) DetectConflicts(source, target string) ([]ConflictInfo, error) {
	migrateConflicts, err := migrate.DetectConflicts(source, target)
	if err != nil {
		return nil, err
	}
	result := make([]ConflictInfo, 0, len(migrateConflicts))
	for _, mc := range migrateConflicts {
		result = append(result, ConflictInfo{
			RelativePath: mc.RelativePath,
			SrcSize:      mc.SrcSize,
			DstSize:      mc.DstSize,
			SrcModTime:   mc.SrcModTime,
			DstModTime:   mc.DstModTime,
		})
	}
	return result, nil
}

// ResolveConflict resolves a file conflict. action can be "skip" or "overwrite".
// This is a placeholder — conflict resolution is not yet implemented.
func (a *App) ResolveConflict(relativePath, action string) error {
	return &migrate.MoveKitError{
		Code:    "NOT_IMPLEMENTED",
		Message: "冲突解决功能尚未实现",
		Detail:  fmt.Sprintf("ResolveConflict(%q, %q) called but conflict resolution is not yet implemented", relativePath, action),
		Action:  "当前版本不支持交互式冲突解决。如遇文件冲突，迁移将自动跳过冲突文件。",
	}
}

// ---------------------------------------------------------------------------
// Cloud Sync Detection
// ---------------------------------------------------------------------------

// DetectCloudSync detects running cloud sync providers.
func (a *App) DetectCloudSync() []CloudProvider {
	providers := migrate.DetectCloudSync()
	result := make([]CloudProvider, 0, len(providers))
	for _, p := range providers {
		result = append(result, CloudProvider{
			Name:   p.Name,
			Folder: p.Folder,
		})
	}
	return result
}

// ---------------------------------------------------------------------------
// Update Check
// ---------------------------------------------------------------------------

// CheckUpdate checks for available updates.
func (a *App) CheckUpdate() UpdateInfo {
	info, err := migrate.CheckUpdate()
	if err != nil {
		return UpdateInfo{
			HasUpdate:   false,
			NewVersion:  currentVersion,
			DownloadURL: "",
		}
	}
	return UpdateInfo{
		HasUpdate:   info.HasUpdate,
		NewVersion:  info.NewVersion,
		DownloadURL: info.DownloadURL,
	}
}

// ---------------------------------------------------------------------------
// System Notification
// ---------------------------------------------------------------------------

// NotifyCompletion shows a system notification when migration completes.
// This is a placeholder that uses Wails dialog as a fallback.
func (a *App) NotifyCompletion(title, message string) {
	// Use Wails runtime dialog as a system notification fallback.
	// A production implementation could use go-toast or similar.
	go func() {
		runtime.EventsEmit(a.ctx, "notification", map[string]string{
			"title":   title,
			"message": message,
			"time":    time.Now().Format(time.RFC3339),
		})
	}()
}
