# Codebase Concerns

**Analysis Date:** 2026-06-21

## Tech Debt

### BitLocker Detection Stub

- Issue: The `checkBitLocker()` function in `internal/migrate/check.go:163-183` opens the registry key but the actual protection status check is a TODO stub. The function always returns nil (no error), regardless of whether BitLocker is actually engaged on the target drive. If BitLocker locks the volume mid-migration, the copy will fail with cryptic Windows errors rather than a clear pre-check message.
- Files: `internal/migrate/check.go`
- Impact: Silent migration failures if BitLocker is active on the target drive. Users on managed enterprise machines (very common target audience) are at highest risk.
- Fix approach: Implement WMI query for `Win32_EncryptableVolume` or use `DeviceIoControl` with `IOCTL_VOLUME_GET_VOLUME_DISK_EXTENTS` to detect the encryption status. Alternatively, invoke `manage-bde -status` via `os/exec` as a pragmatic Windows-only solution.

### SSD Detection Always Conservative-False

- Issue: `IsSSDDrive()` in `internal/syscall/volume.go:130-151` always returns `false` with the comment "Conservative default - can't reliably detect without admin". The `GetVolumeInfo()` function in the same file (line 73) uses a heuristic: all fixed drives are treated as SSD. Both approaches are inaccurate and disagree with each other.
- Files: `internal/syscall/volume.go`
- Impact: The duration estimate and HDD/SSD warnings in the UI are misleading. An HDD target may show as "SSD" in the disk selector, while `IsSSDDrive` (unused currently) would mark everything as HDD.
- Fix approach: Use `DeviceIoControl` with `IOCTL_STORAGE_QUERY_PROPERTY` / `StorageDeviceSeekPenaltyProperty` which works without admin rights on Windows 8+. Consolidate SSD detection into a single function used everywhere.

### Update Check Placeholder

- Issue: `CheckUpdate()` in `internal/migrate/version.go:11-18` always returns `HasUpdate: false` -- it never queries a remote source.
- Files: `internal/migrate/version.go`
- Impact: Users never get notified about new versions. If a critical bug is fixed in a later release, users on the old version will not know.
- Fix approach: Query a GitHub releases API endpoint (`https://api.github.com/repos/{owner}/{repo}/releases/latest`) with proper timeout and error handling.

### Conflict Resolution MVP-Only

- Issue: `DetectConflicts` is called in `internal/migrate/migrate.go:97-102`, but conflicts are only logged, never acted upon. The comment says "For MVP: log conflicts but proceed (user already confirmed)". Files in conflict will be silently overwritten during the cross-volume copy.
- Files: `internal/migrate/migrate.go:97-102`, `internal/migrate/conflict.go`, `frontend/src/App.tsx` (ConflictDialog exists but is never shown because conflicts are not emitted to frontend)
- Impact: User data loss if a file exists at both source and target with different content. The conflict dialog UI component exists in the frontend but is never wired up.
- Fix approach: Emit detected conflicts to the frontend via Wails events before the copy phase. Block migration until the user resolves each conflict. The `ConflictDialog` component code already handles the UI -- it just needs wiring.

### System Notification Placeholder

- Issue: `NotifyCompletion()` in `app.go:564-573` does not call Windows native toast notifications. It emits a Wails runtime event instead.
- Files: `app.go:564-573`
- Impact: If the user switches away from the MoveKit window, they will not receive a system-level notification when migration completes. On a long migration (minutes to hours), this is a poor user experience.
- Fix approach: The project already includes `git.sr.ht/~jackmordaunt/go-toast/v2` as an indirect dependency (via Wails). Use this directly to fire a Windows toast notification.

### ResolveConflict Stub

- Issue: `ResolveConflict()` in `app.go:513-518` is a no-op stub: "conflict resolution is handled during migration". It silently returns nil.
- Files: `app.go:513-518`
- Impact: The frontend `ConflictDialog` calls this function, but it does nothing. The dialog appears to work but conflict resolution is never applied.
- Fix approach: Implement actual conflict resolution logic: maintain a resolution map (`relativePath -> "skip" | "overwrite"`) that the cross-volume mover checks before each file copy.

### Frontend simulates progress with setInterval

- Issue: `simulateProgress()` in `frontend/src/App.tsx:139-181` uses JavaScript `setInterval` to fake progress bar advancement. It does not consume real progress events from the Wails `migration:progress` event stream that the Go backend emits.
- Files: `frontend/src/App.tsx:129-181`
- Impact: The progress bar shows fabricated data. If the migration takes much longer or shorter than the "20 steps" heuristic, the progress bar will either hang at 100% or jump to completion prematurely. This destroys user trust.
- Fix approach: Remove `simulateProgress()`. Use `useEffect` to subscribe to Wails `runtime.EventsOn("migration:progress", ...)` events. The backend already emits real progress data via `migration:progress` events in `app.go:300-330`.

### Cancel button does not cancel the migration

- Issue: `handleCancel()` in `frontend/src/App.tsx:183-185` sets React state to `'idle'` but never calls `CancelMigration()` from the Wails bindings. The Go backend's migration goroutine continues running.
- Files: `frontend/src/App.tsx:183-185`, `frontend/src/lib/wails.ts:53-55`
- Impact: User believes they cancelled the migration. But the backend continues copying files. Resource consumption and potential data inconsistency.
- Fix approach: In `handleCancel()`, call `CancelMigration()` (which sets `cancelled.Store(1)` in the Go backend). Also ensure the real Wails `migration:done` event handler updates the UI state.

### Security Software Detection Result Discarded

- Issue: `detectSecuritySoftware()` is called in `internal/migrate/check.go:52-56`, but the result is assigned to `_`. The name of any detected AV is never used for warnings or logging.
- Files: `internal/migrate/check.go:52-56`
- Impact: Users with aggressive AV (360, HuoRong, Kaspersky) may have their real-time protection interfere with junction creation or cross-volume file operations without any warning.
- Fix approach: Log the detected AV name. Emit a warning event to the frontend suggesting the user temporarily pause the AV during migration. Consider adding a dedicated pre-migration AV warning dialog.

### percent variable discarded

- Issue: In `internal/migrate/move.go:137`, the `percent` variable is computed but discarded with `_ = percent`. The progress callback is already called before this line with the full data, so this is a dead computation.
- Files: `internal/migrate/move.go:137`
- Impact: Dead code that suggests the progress reporting was restructured mid-development and not cleaned up. Low severity but indicates incomplete refactoring.
- Fix approach: Remove the unused variable assignment and its associated comment.

## Known Bugs

### Error type assertion without comma-ok check

- Bug: In `internal/syscall/process.go:43`, the code does `e.(syscall.Errno)` without the comma-ok pattern. If `e` is nil (some Windows API calls set LastError to 0 on success) or not of type `syscall.Errno`, this will panic.
- Files: `internal/syscall/process.go:43`
- Trigger: Calling `EnsureSingleInstance()` on a Windows system where `CreateMutexW` succeeds but returns a non-standard error type in the third return value.
- Workaround: None. This is a latent panic.
- Fix:
```go
if errno, ok := e.(syscall.Errno); ok && errno == ERROR_ALREADY_EXISTS {
    closeHandle(syscall.Handle(h))
    return syscall.InvalidHandle, fmt.Errorf("another instance is already running (mutex %q)", name)
}
```

### Junction debug tests show buffer mismatch

- Bug: `TestDebugReparseBuffer` in `internal/syscall/junction_debug_test.go:46-163` compares the buffer produced by `buildReparseBuffer()` against a buffer produced by `mklink /J`. The test logs differences in `SubLen` and `PrintLen` and runs `FSCTL_SET_REPARSE_POINT` but the SET call may fail (line 162: "SET failed"). This suggests the reparse buffer construction may be incorrect for some paths.
- Files: `internal/syscall/junction_debug_test.go`, `internal/syscall/junction.go:78-120`
- Trigger: Paths with non-ASCII characters or long paths may produce incorrect reparse buffers. The `buildReparseBuffer` function's `PrintNameOffset` calculation at line 106 (`uint16(subLen+2)`) assumes the substitute name's NUL terminator is exactly 2 bytes after its length, which may not be correct for all substitution name configurations.
- Workaround: The debug test only logs the issue. The production code at `CreateJunction` in `junction.go:125-174` does not validate the buffer structure is consistent with `mklink` behavior.
- Fix approach: Verify the buffer layout against Microsoft's REPARSE_DATA_BUFFER specification. Cross-reference with `mklink /J` output on various path types (long, Unicode, spaces). Consider using `CreateSymbolicLinkW` or the `mklink /J` command as a fallback if `DeviceIoControl` SET fails.

### Backup save errors silently swallowed

- Bug: In `internal/migrate/migrate.go:115-117` and line 206, `saveBackup()` errors are explicitly discarded with `_`. If the backup file cannot be written (disk full, permission denied, AppData missing), the migration proceeds without a backup record. Rollback becomes impossible.
- Files: `internal/migrate/migrate.go:115-117`, `internal/migrate/migrate.go:206`
- Trigger: Running on a system with full C drive, or a system where `%APPDATA%\MoveKit\backups` cannot be created.
- Workaround: None. The migration completes but there is no backup to roll back from.
- Fix approach: Abort migration if the initial backup save fails (line 115). If only the completion save fails (line 206), log a warning but note the backup is incomplete.

### Same-volume move has no fallback

- Bug: In `internal/migrate/migrate.go:132-141`, if `os.Rename()` fails on the same volume (which Should be atomic on NTFS), the error is returned but no cleanup is done. If the rename partially fails (extremely rare on NTFS but possible with junction/DFS interactions), files could be in an inconsistent state.
- Files: `internal/migrate/migrate.go:132-141`
- Trigger: Rename failure due to file locks, path length limits, or NTFS metadata corruption.
- Workaround: The error is returned with a user-friendly message to "close programs and retry".
- Fix approach: On NTFS rename failure, verify the source and target states. If the rename is truly atomic on NTFS (which it is for same-volume moves), the source either exists (on failure) or doesn't (on success). Add an assertion that confirms this before reporting to the user.

### dirSize silently fails in GetFolders

- Bug: In `app.go:81-83`, `dirSize(expanded)` errors are silently ignored. The `fi.SizeBytes` field remains 0. All five known folders could show 0 bytes in the UI if the directory has permission issues.
- Files: `app.go:81-83`
- Trigger: User folders with restricted NTFS ACLs, junction points that point to inaccessible locations, or deeply nested paths exceeding MAX_PATH.
- Workaround: None. UI shows 0 bytes for that folder with no indication of error.
- Fix approach: Set `fi.SizeBytes` to -1 on error to indicate unmeasurable, and display "无法计算大小" in the UI instead of "0 B".

### Shell notify UTF16 errors ignored

- Bug: In `internal/syscall/shell.go:43-44` and line 56, `syscall.UTF16PtrFromString()` errors are discarded with `_`. If the path contains characters that cannot be encoded as UTF-16 (extremely rare), the resulting Ptr may be garbage.
- Files: `internal/syscall/shell.go:40-63`
- Trigger: Paths with invalid Unicode surrogates or extremely long strings.
- Workaround: None. Explorer may not refresh, leaving the user confused.
- Fix approach: At minimum, log the error. Better: fall back to a broad `SHCNE_UPDATEDIR` with the parent directory if individual path encoding fails.

## Security Considerations

### Global mutable state without synchronization

- Risk: `activeCopyCallback` in `internal/syscall/fileops.go:49` is a package-level variable modified without any synchronization primitive. Only one cross-volume copy can be active at a time, but the comment acknowledges this ("only one copy can be active at a time") without enforcing it.
- Files: `internal/syscall/fileops.go:49-57`
- Current mitigation: The `App` struct uses `sync.Mutex` to prevent concurrent migrations, so the race is not triggerable under normal use. However, if `CopyFileEx` or `CopyFileWithACL` is ever called from a different goroutine (e.g., a future parallel copy feature), this will cause data corruption or crashes.
- Recommendations: Add a `sync.Mutex` to guard `activeCopyCallback`. Enforce single-copy semantics at the `CopyFileEx` level rather than relying on the caller.

### e.(syscall.Errno) type assertion without comma-ok

- Risk: As described in Known Bugs, this can panic if the Windows error value is nil or an unexpected type. In a Go desktop application, a panic would crash the entire process without cleanup.
- Files: `internal/syscall/process.go:43`
- Current mitigation: On standard Windows 10/11 with kernel32.dll, `CreateMutexW` always returns `syscall.Errno` for errors. This works in practice.
- Recommendations: Use the comma-ok pattern as shown in the Known Bugs fix above.

### Junction hijack check uses simple string comparison

- Risk: `CheckJunctionHijack()` in `internal/migrate/security.go:29-37` uses `resolved != targetPath` for comparison. But `filepath.EvalSymlinks` normalizes paths (e.g., removes trailing slashes, resolves mount points) while `targetPath` is constructed via `filepath.Join(driveLetter+":\", folderName)`. Different normalization of the same path could cause false positives.
- Files: `internal/migrate/security.go:13-50`
- Current mitigation: False positives are safe -- they block migration rather than allowing a dangerous one. False negatives (missing a hijack) are the real risk.
- Recommendations: Use case-insensitive path comparison with `strings.EqualFold(filepath.Clean(a), filepath.Clean(b))`. Also check for `..` components in the `folderName` to prevent directory traversal.

### Registry writes without privilege detection

- Risk: Registry writes (`WriteExpandString`, `WriteString`) to `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders` require write access. While HKCU typically does not need administrator rights, some managed environments use GPO to restrict registry modifications. The error messages say "Run as administrator" which may be misleading (the real fix might involve GPO changes).
- Files: `internal/migrate/migrate.go:175-195`, `internal/migrate/rollback.go:64-83`
- Current mitigation: Errors are caught and returned with actionable messages.
- Recommendations: Pre-check registry write access during `PreCheck()`. Attempt a test write/delete of a dummy value under the target key to detect permission issues early, before any files are moved.

### No input validation on registry value names

- Risk: `regValueName` parameters are passed from the frontend to backend without validation. While they come from `folders.AllFolders` (controlled by the app), if the frontend is ever compromised or a future API allows arbitrary values, registry key injection is theoretically possible.
- Files: `app.go:361` (`CancelMigration`), `app.go:370` (`Rollback`), `app.go:493` (`DetectConflicts`)
- Current mitigation: All call sites in the app use values from the hardcoded `folders.AllFolders` list.
- Recommendations: Add a validation function that checks `regValueName` against known allowed values. Fail fast with a clear error message for unexpected values.

### Frontend cancel bypass

- Risk: `handleCancel()` only sets React state; the goroutine in Go continues. If the user starts another migration quickly (state reverts to 'idle'), the `mu.Lock()` in `StartMigration` prevents actual concurrent migration, but the old goroutine is still running in the background consuming resources and potentially modifying the filesystem.
- Files: `frontend/src/App.tsx:183-185`
- Current mitigation: The `sync.Mutex` in `App` prevents a new `StartMigration` from starting while the old goroutine runs.
- Recommendations: Call Wails `CancelMigration()` in `handleCancel()`. Subscribe to `migration:done` Wails events to update state from the backend truth, not just the frontend UI state.

### No path traversal protection for user input

- Risk: The `targetDrive` parameter from the frontend is used as a drive letter prefix. While the frontend `DiskSelector` restricts choices to discovered drives, a compromised or modified frontend could pass an arbitrary string like `C:\Windows\System32\malicious` as the `targetDrive`.
- Files: `app.go:271` (`StartMigration(targetDrive string)`), `internal/migrate/migrate.go:77-82`
- Current mitigation: The code normalizes by extracting only the first character as a drive letter. `os.Rename` and `filepath.Join` further limit the scope.
- Recommendations: Validate that `targetDrive` is exactly one letter (A-Z) optionally followed by a colon. Reject anything else. Also validate `folderName` from `filepath.Base(actualSource)` in `migrate.go:81` to prevent special directory names.

## Performance Bottlenecks

### dirSize tree walk on every GetFolders call

- Problem: `GetFolders()` in `app.go:61-104` calls `dirSize()` for each of the 5 known folders. `dirSize()` walks the entire directory tree synchronously. For large folders (e.g., 50GB Documents), this can take seconds on HDD or network-mapped drives.
- Files: `app.go:81-83`
- Cause: No caching. No incremental size tracking. Every `GetFolders()` call recalculates from scratch.
- Improvement path: Cache folder sizes with a TTL (e.g., 30 seconds). Show a "calculating..." placeholder while sizes load in the background. Consider using `FindFirstFileEx` with `FindExInfoBasic` to get sizes without enumerating all files, or maintain a size cache file.

### Double tree walk for cross-volume file counting

- Problem: `moveFilesCrossVolume()` in `internal/migrate/move.go:63-83` performs a full `filepath.WalkDir` to count files and total bytes, then performs a second full `filepath.WalkDir` at line 89 to actually copy files. Each walk may touch thousands of files.
- Files: `internal/migrate/move.go:63-141`
- Cause: The first pass counts; the second pass copies. No way to skip the count.
- Improvement path: Remove the counting pass. Report progress with only the files-copied count (no percentage). OR: Collect all file paths during the first walk into a slice, then iterate the slice for copying. The memory cost for 100k file paths (~5MB) is acceptable on modern systems.

### dirSize also called in EstimateDuration

- Problem: `EstimateDuration()` in `app.go:143-208` also calls `dirSize()` for each of the 5 known folders, duplicating the work already done (or about to be done) by `GetFolders()`.
- Files: `app.go:146-157`
- Cause: No shared size computation between `GetFolders` and `EstimateDuration`.
- Improvement path: Pass pre-computed folder sizes from `GetFolders` to the frontend. Use those sizes for the estimate calculation. Cache sizes at the backend.

### DetectConflicts walks entire source tree

- Problem: `DetectConflicts()` in `internal/migrate/conflict.go:11-48` does a full `filepath.Walk` of the source to check for conflicts at the target. For a 50GB Documents folder with 100k files, this is expensive.
- Files: `internal/migrate/conflict.go:14-45`
- Cause: No way to limit scope or cache results.
- Improvement path: Run conflict detection in a goroutine with progress reporting. Show a "正在检查文件冲突..." state in the UI. Consider sampling (check first 1000 files) for a quick estimate, then run full check in background.

## Fragile Areas

### Cross-volume copy-then-delete leaves orphans on failure

- Files: `internal/migrate/move.go:114-122`
- Why fragile: For each file, `CopyFileExW` succeeds, then `os.Remove()` fails (e.g., file lock appears between copy and delete). The file now exists in both source and target. On cancellation, some files are copied, some are not, and some are deleted.
- Safe modification: The code already handles cancellation (checks `IsCancelled()` before each file). But partial failure during copy is unrecoverable -- you cannot know which files exist at the target.
- Test coverage: Only cancellation tests exist in `internal/migrate/move_test.go`. No tests for `CopyFileExW` failure followed by `os.Remove` failure in a realistic NTFS scenario.

### Migration can fail between registry update and backup save

- Files: `internal/migrate/migrate.go:172-210`
- Why fragile: Steps are: 1) Create junction, 2) Update User Shell Folders, 3) Update Shell Folders, 4) Send shell notifications, 5) Update backup. If the process or system crashes between step 2 and step 5, the junction exists, the registry is updated, but the backup record still says `"in_progress"`. A subsequent rollback attempt would not find a complete record.
- Safe modification: Save the backup record with `status: "complete"` immediately after the junction is created and before registry modifications begin. If registry writes fail, the backup record has `newPath` recorded so manual recovery is possible.
- Test coverage: None. No integration tests for the full `Migrate()` function -- all tests are unit tests of individual functions.

### Junction creation after MkdirAll fails leaves orphaned target files

- Files: `internal/migrate/migrate.go:152-169`
- Why fragile: Files are moved to the target path. Then `os.MkdirAll(sourcePath, 0755)` at the original location fails (permission error). The junction is never created. The files are at the target but the source path is inaccessible (not a junction, not the original files).
- Safe modification: Verify that the source directory can be created before moving files. Pre-check write permissions on the parent of the source path.
- Test coverage: None for this specific failure mode.

### Rollback uses os.Rename (same-volume only)

- Files: `internal/migrate/rollback.go:54`
- Why fragile: `os.Rename()` is used unconditionally for rollback. If the original location was on C drive and files were moved to D drive, the rollback rename would fail (cross-volume). But in the current design, all known folders start on C drive, so rollback should be a cross-volume move back to C. This suggests rollback may fail if files were moved cross-volume.
- Safe modification: Check if rollback is cross-volume using `sys.IsCrossVolume(rec.NewPath, rec.OriginalPath)`. If cross-volume, use `moveFilesCrossVolume()` for rollback. Add the target direction to the backup record.
- Test coverage: `Rollback()` has no unit test at all. `move_test.go` only tests cancellation.

### App struct shutdown is a no-op

- Files: `app.go:39`
- Why fragile: `shutdown(ctx context.Context)` does nothing. If the user closes the app during migration, the migration goroutine is orphaned. Files may be partially copied. The backup record remains at `"in_progress"`.
- Safe modification: In `shutdown()`, set the cancellation flag via `migrate.CancelMigration()`. Wait briefly for the goroutine to exit (or at minimum log the abandoned state). At startup, detect `"in_progress"` backup records and offer recovery options.
- Test coverage: None.

## Scaling Limits

### Maximum 4 drive letters checked

- Current capacity: `GetDisks()` in `app.go:111-135` checks only drives C, D, E, F.
- Limit: Systems with more than 4 drives (G:, H:, etc.) cannot be selected as targets. Multi-drive workstations (common in video editing, development, or server use) are affected.
- Scaling path: Iterate drive letters from C to Z. Check each with `GetDriveType()`. Skip non-fixed drives.

### Single-threaded file copy

- Current capacity: `CopyFileExW` copies one file at a time sequentially in the main migration goroutine. No parallelism.
- Limit: For folders with many small files (>100k), the per-file copy overhead dominates and migration is slow.
- Scaling path: Use a worker pool (4-8 goroutines) each calling `CopyFileExW` with chunked file lists. Ensure cancellation propagation.

### 30-day hardcoded backup expiry

- Current capacity: Backup records older than 30 days are cleaned up in `CleanupExpiredBackups()` (`internal/migrate/backup.go:92-108`). 30 days is hardcoded.
- Limit: Users who migrate and forget, then need to rollback after 31 days, lose their backup records.
- Scaling path: Make the retention period configurable. Show a warning in the UI before cleanup runs. Store backups in a compressed format to reduce disk usage for longer retention.

### No log rotation or size limit

- Current capacity: `Logger` appends to a daily log file indefinitely. No size-based rotation or cleanup of old logs.
- Limit: Over months of use, `%APPDATA%\MoveKit\logs\` could grow very large.
- Scaling path: Implement automatic log cleanup for files older than N days. Compress old logs. Or cap at a total log directory size.

## Dependencies at Risk

### golang.org/x/sys v0.30.0

- Risk: This is a compatibility library that wraps Windows API functions. Version 0.30.0 interacts with `syscall.NewLazyDLL` and raw syscalls throughout the codebase. API changes in newer Go releases (especially Go 1.24+ syscall deprecations) could break these calls.
- Impact: Many of the `unsafe.Pointer` syscalls would need to be rewritten to use `golang.org/x/sys/windows` package functions instead of raw `syscall.NewLazyDLL.NewProc().Call()` patterns.
- Migration plan: Migrate raw syscalls (`procCreateFileW.Call(...)`) to use `golang.org/x/sys/windows.CreateFile(...)` wrapper functions. These are more stable across Go versions and provide proper error handling.

### Wails v2.12.0

- Risk: Wails v2 is stable but v3 is in development. The `wails.json` schema references v2 explicitly. Bindings are generated by `wails build` into `frontend/wailsjs/`.
- Impact: The auto-generated bindings in `frontend/wailsjs/` are committed to the repository. If Wails update the codegen, the bindings could drift.
- Migration plan: Regenerate bindings on each Wails version bump. Add a CI check that detects binding drift.

### react ^19.2.7

- Risk: React 19 is relatively new. The `StrictMode` in `main.tsx:7` causes double-rendering in development mode, which could expose race conditions in the Wails binding calls during dev testing.
- Impact: Minor. Production builds (`vite build`) do not double-render.
- Migration plan: Be aware of React 19 concurrent features (useTransition, useDeferredValue) that could interact with synchronous Wails Go calls. Test thoroughly under `StrictMode` in dev.

## Missing Critical Features

### No power-loss protection during migration

- Problem: If the system loses power mid-migration (desktop power outage, laptop battery dies), files exist in an inconsistent state. The copy-then-delete approach means some files are at the target and some are at the source, with no transaction log to determine which are which.
- Blocks: Safe migration on unreliable hardware or during storms.
- Priority: High for a tool that modifies the user's profile folders.

### No pre-migration system restore point creation

- Problem: Windows System Restore can undo registry changes but not file moves. The tool does not create a restore point before migration.
- Blocks: Easy recovery for non-technical users who encounter issues.
- Priority: Medium. A restore point costs ~30 seconds and provides a safety net for registry-only reversions.

### No integrity verification after copy

- Problem: `VerifyIntegrity()` exists (`internal/migrate/verify.go`) but is never called from the `Migrate()` function or the `moveFilesCrossVolume()` function. Files are copied without checksum verification.
- Blocks: Confidence that the copy was bit-perfect.
- Priority: Medium. NTFS file copy via `CopyFileExW` is reliable, but for a tool that deletes source files, verification is important for user trust.

### No disk health check before migration

- Problem: No SMART status check or disk health verification before writing to the target drive.
- Blocks: Avoiding migration to a failing drive.
- Priority: Low for MVP, but valuable for user confidence.

## Test Coverage Gaps

### Untested Go packages

| Package / File | Tests? | Risk |
|---|---|---|
| `app.go` (App layer) | None | HIGH - All Wails bindings, migration orchestration |
| `app_types.go` | None | LOW - Type definitions only |
| `main.go` | None | MEDIUM - Entry point, single-instance logic |
| `internal/migrate/cloud.go` | None | LOW - Simple path existence checks |
| `internal/migrate/log.go` | None | MEDIUM - File I/O, thread safety |
| `internal/migrate/migrate.go` (`Migrate()`) | None | HIGH - Core migration orchestration |
| `internal/migrate/move.go` (`moveFilesCrossVolume`) | Only cancellation tests | HIGH - File copy logic |
| `internal/migrate/rollback.go` | None | HIGH - Rollback logic (files moved back, registry restored) |
| `internal/migrate/security.go` | None | MEDIUM - Junction hijack detection |
| `internal/migrate/state.go` | None | LOW - JSON file state persistence |
| `internal/migrate/verify.go` | None | MEDIUM - Integrity verification (also unused) |
| `internal/migrate/version.go` | None | LOW - Update check (placeholder) |
| `internal/syscall/process.go` | None | MEDIUM - Mutex, process enumeration, error type assertion bug |
| `internal/syscall/shell.go` | None | LOW - Shell notification wrappers |
| `cmd/verify/main.go` | None | LOW - Manual verification CLI tool |

### Prioritized test gaps

**Highest priority:**
- `Migrate()` integration test -- mock registry, simulate junction create/delete, verify folder movement. This is the core function with the most failure modes (lines of code: ~200, 8 sequential steps, any of which can fail).
- `moveFilesCrossVolume()` real test -- the current `move_test.go` only tests cancellation flags. Needs actual file copy with verification in a temp directory.

**High priority:**
- `Rollback()` integration test -- verify restore of registry, file movement back, junction cleanup.
- `CheckJunctionHijack()` security test -- verify edge cases with symlinks, junctions, path normalization tricks.

**Medium priority:**
- `App.StartMigration()` test with Wails event emissions -- verify progress events and done events fire correctly.
- `GetFolders()` test -- verify all 5 known folder states (default, migrated, not_on_c, junction_broken, unknown).

### No frontend tests at all

- What's not tested: Any React component, hook, or Wails binding wrapper.
- Files: All `frontend/src/**/*.tsx` and `frontend/src/lib/wails.ts`
- Risk: The `simulateProgress()` bug (fake progress vs real Wails events) would have been caught by even a basic integration test. The cancel button not calling `CancelMigration()` is trivially testable.
- Priority: Medium for now (the frontend has relatively little logic -- most logic is in Go). High before adding more UI features.
- Approach: Add Vitest (already in node_modules via Vite dependency). Test at minimum: `wails.ts` mock/real binding delegation, `ProgressBar` rendering with edge case props, `FolderList` status badge rendering.

---

*Concerns audit: 2026-06-21*
