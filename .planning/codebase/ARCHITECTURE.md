<!-- refreshed: 2026-06-21 -->
# Architecture

**Analysis Date:** 2026-06-21

## System Overview

```text
+-----------------------------------------------------------------+
|                    React Frontend (TypeScript)                    |
|  App.tsx (state machine) + 11 components + lib/wails.ts          |
|  `frontend/src/`                                                  |
+---------------------------------+-------------------------------+
| Wails Runtime (v2.12.0)         | Wails-Generated Bindings       |
| Events + Bound Method Calls     | `frontend/wailsjs/go/main/App.js` |
+---------------------------------+-------------------------------+
|                    Go Backend (package main)                      |
|  main.go / app.go / app_types.go                                  |
|  App struct — 17 exported methods bound to frontend               |
+----------------+-----------------+-------------------------------+
| internal/migrate                | internal/syscall                |
| Core migration logic            | Win32 API wrappers              |
| - Migrate() pipeline            | - junction.go (DeviceIoControl) |
| - PreCheck() validation         | - registry.go (HKCU read/write)|
| - Backup/Rollback/Verify        | - fileops.go (CopyFileExW+ACL) |
| - OneDrive KFM detection        | - volume.go (GetVolumeInfo)    |
| - Cloud sync detection          | - process.go (toolhelp32)      |
| - Conflict detection            | - power.go (battery status)    |
| - Cancellation (atomic flag)    | - shell.go (SHChangeNotify)    |
+---------------------------------+-------------------------------+
| internal/folders                |                                |
| KnownFolder definitions + GUIDs |                                |
+---------------------------------+-------------------------------+
                        |
                        v
+-----------------------------------------------------------------+
|                 Windows OS (NTFS / Registry / Shell)             |
+-----------------------------------------------------------------+
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| `main.go` | Wails entry point, single-instance mutex, embed FS for frontend dist | `main.go` |
| `App` struct | 17 exported methods auto-bound to frontend via Wails; state machine orchestration | `app.go` |
| `app_types.go` | All shared TypeScript-facing Go types with JSON tags (FolderInfo, DiskInfo, ProgressInfo, etc.) | `app_types.go` |
| **internal/syscall** | Low-level Win32 API wrappers: junction ops, registry, file ops with ACL, volume info, process enumeration, power status, shell notifications | `internal/syscall/*.go` |
| **internal/migrate** | Core migration pipeline, pre-flight checks, backup/rollback, OneDrive/cloud detection, conflict detection, cancellation, structured logging | `internal/migrate/*.go` |
| **internal/folders** | Known folder definitions (5 folders: desktop, downloads, music, documents, pictures) with GUIDs and registry value names | `internal/folders/known.go` |
| `cmd/verify/main.go` | CLI verification tool that exercises all backend subsystems without the Wails GUI | `cmd/verify/main.go` |
| `frontend/src/App.tsx` | Frontend state machine (idle/checking/confirming/migrating/completed/failed), orchestrates all UI components and dialogs | `frontend/src/App.tsx` |
| `frontend/src/lib/wails.ts` | Wails binding proxy — imports real bindings at runtime or falls back to mock data for dev mode | `frontend/src/lib/wails.ts` |
| `frontend/src/types.ts` | Shared TypeScript type definitions mirroring Go structs | `frontend/src/types.ts` |
| 11 React components | Each a discrete UI state or dialog overlay | `frontend/src/components/*.tsx` |
| **frontend/wailsjs/** | Auto-generated Wails binding code (Go method stubs in JS + TypeScript type classes) | `frontend/wailsjs/go/main/` |

## Pattern Overview

**Overall:** Wails v2 Binding Pattern with State Machine Frontend

The Go backend application struct (`App` in `app.go`) exposes exported methods. Wails automatically generates JavaScript binding stubs in `frontend/wailsjs/go/main/App.js` and TypeScript type declarations in `frontend/wailsjs/go/models.ts`. The frontend calls Go methods as async functions, and the Go backend emits events (`migration:progress`, `migration:done`, `notification`) back to the frontend via `runtime.EventsEmit`.

**Key Characteristics:**
- Dual-mode frontend: real Wails bindings at runtime, mock data fallback when running in dev (vite)
- CancellationToken pattern: atomic flag (`int32`) checked during file iteration
- Error code classification: every migration error has a machine-readable code (`ErrCode*`), human-readable message (Chinese), detail, and suggested action
- Path masking: log output masks user profile directory for privacy
- Single-instance enforcement: named mutex at startup

## Layers

### Frontend Layer
- Purpose: UI rendering, state management, user interaction
- Location: `frontend/src/`
- Contains: React components, TypeScript types, Wails binding proxy, Tailwind CSS v4 config
- Depends on: Wails runtime bindings (`frontend/wailsjs/`) for Go method calls; emits no direct OS calls
- Used by: No internal consumers (leaf layer)

### Go Main Package (`package main`)
- Purpose: Wails application entry point, App struct with all frontend-bound methods, type definitions for JSON serialization
- Location: `main.go`, `app.go`, `app_types.go`
- Contains: `main()` entry, `App` struct (17 exported methods), config I/O (first-run flag, `%APPDATA%\MoveKit\config.json`)
- Depends on: `internal/migrate`, `internal/syscall`, `internal/folders`, Wails runtime
- Used by: Wails framework (binds App methods to frontend)

### Migration Logic Layer (`internal/migrate`)
- Purpose: Core business logic — migration pipeline, pre-flight checks, backup/rollback, OneDrive/cloud detection, conflict detection, structured file logging, versioning, cancellation
- Location: `internal/migrate/` (17 Go files + 5 test files)
- Contains: `Migrate()`, `PreCheck()`, `SaveBackup()`/`LoadBackup()`, `Rollback()`, `DetectOneDriveKFM()`, `DetectCloudSync()`, `DetectConflicts()`, `CancelMigration()`/`IsCancelled()`, `Logger`, `MoveKitError` type
- Depends on: `internal/syscall`, `golang.org/x/sys/windows/registry`
- Used by: `main` package (App methods)

### System Call Layer (`internal/syscall`)
- Purpose: Direct Windows API access via `syscall` package — all OS interaction lives here and nowhere else
- Location: `internal/syscall/` (7 Go files + 6 test files)
- Contains: `CreateJunction()`/`IsJunction()`/`DeleteJunction()`, `ReadExpandString()`/`WriteExpandString()`/`WriteString()`, `CopyFileEx()`/`CopyFileWithACL()`/`MoveFileEx()`, `GetVolumeInfo()`/`GetDiskFreeSpace()`/`IsCrossVolume()`, `EnsureSingleInstance()`/`IsProcessRunning()`, `IsOnBattery()`, `NotifyFolderRenamed()`/`NotifyDirUpdated()`
- Depends on: `golang.org/x/sys/windows/registry`, Windows `kernel32.dll`/`advapi32.dll`/`shell32.dll`
- Used by: `internal/migrate`, `main` package (App methods)

### Known Folders Layer (`internal/folders`)
- Purpose: Static data — defines the 5 known Windows user folders with their GUIDs and registry value names
- Location: `internal/folders/known.go`
- Contains: `KnownFolder` struct, `AllFolders` slice, `GUIDToRegValue` map
- Depends on: Nothing (pure data)
- Used by: `main` package (App methods), `internal/migrate`

## Data Flow

### Primary Request Path: Folder State Query

1. **Frontend calls `GetFolders()`** via Wails binding (`frontend/src/lib/wails.ts:33`)
2. **Wails routes to `App.GetFolders()`** (`app.go:61`)
3. **Iterates `folders.AllFolders`** (`internal/folders/known.go:9`) — reads registry for each folder path (`sys.ReadExpandString` at `internal/syscall/registry.go:22`)
4. **Checks junction status** via `sys.IsJunction()` (`internal/syscall/junction.go:178`) — uses `DeviceIoControl` + `FSCTL_GET_REPARSE_POINT`
5. **Calculates folder size** via `dirSize()` (`app.go:211`) — `filepath.Walk` on source
6. **Returns `[]FolderInfo`** (serialized as JSON by Wails) → TypeScript `FolderInfo[]`

### Primary Request Path: Migration Execution

1. **Frontend calls `StartMigration(targetDrive)`** via Wails binding (`frontend/src/lib/wails.ts:48`)
2. **`App.StartMigration()`** (`app.go:271`) acquires mutex, resets cancellation, spawns goroutine
3. **For each folder in `folders.AllFolders`:**
   - Emits `migration:progress` event via `runtime.EventsEmit`
   - Calls `migrate.Migrate(regValueName, targetDrive, onProgress)` (`internal/migrate/migrate.go:22`)
4. **`Migrate()` pipeline:**
   a. Read source path from registry (`internal/migrate/migrate.go:42`)
   b. Resolve if source is already a junction (chain migration) (`line 57-74`)
   c. `CheckJunctionHijack()` on target (`internal/migrate/security.go:13`)
   d. `PreCheck()` validation (`internal/migrate/check.go:30`) — NTFS, space, BitLocker, Defender, Windows Update, battery, drive type
   e. Save backup record with `status: "in_progress"` (`line 106-117`)
   f. Move files — cross-volume (copy+delete via `moveFilesCrossVolume()` at `internal/migrate/move.go:63`) or same-volume (atomic `os.Rename` at `line 133`)
   g. Create junction at original location via `sys.CreateJunction()` (`internal/syscall/junction.go:125`) — `DeviceIoControl` + `FSCTL_SET_REPARSE_POINT`
   h. Update registry (both `User Shell Folders` as `REG_EXPAND_SZ` and `Shell Folders` as `REG_SZ`) via `sys.WriteExpandString()`/`sys.WriteString()` (`internal/syscall/registry.go`)
   i. Notify shell via `sys.NotifyFolderRenamed()` and `sys.NotifyDirUpdated()` (`internal/syscall/shell.go`)
   j. Mark backup `status: "complete"` (`line 204-206`)
5. **Goroutine emits** `migration:done` event with success/failure

### Cancellation Flow

1. Frontend calls `CancelMigration()` → `migrate.CancelMigration()` (`internal/migrate/move.go:19`)
2. Sets atomic flag `cancelled.Store(1)`
3. `moveFilesCrossVolume()` checks `IsCancelled()` at each file iteration (`internal/migrate/move.go:100`)
4. Returns immediately after current file completes; goroutine exits with cancellation error
5. `App.StartMigration()` emits `migration:done` with `success: false`

**State Management:**
- Frontend: React `useState` with explicit `AppState` enum (`'idle' | 'checking' | 'confirming' | 'migrating' | 'completed' | 'failed'`)
- Backend: `App.mu sync.Mutex` for migration lock, `App.migrating bool` for in-progress flag, `migrate.cancelled atomic.Int32` for cancellation
- Persistent state: `%APPDATA%\MoveKit\config.json` (first-run flag), `%APPDATA%\MoveKit\backups\{folder}.json` (backup records), `%APPDATA%\MoveKit\state.json` (migration state)
- Frontend-first-run: `localStorage.getItem('movekit-first-run-seen')` for guide overlay

## Key Abstractions

**MoveKitError:**
- Purpose: Structured error with machine-readable code, human-readable message (Chinese), technical detail, and user-facing action
- Examples: defined in `internal/migrate/errors.go:23` with error code constants at lines 3-21; duplicated in `app_types.go:88` (separate `package main` version to avoid circular imports)
- Pattern: `&MoveKitError{Code: "ONEDRIVE_KFM", Message: "OneDrive 已接管此文件夹", Detail: "...", Action: "请先关闭..."}`

**BackupRecord:**
- Purpose: Persistent JSON record of each migration, enabling rollback and integrity verification
- Examples: defined in `internal/migrate/migrate.go:218`; saved to `%APPDATA%\MoveKit\backups\{regValueName}.json`
- Pattern: `{Version, RegValueName, OriginalPath, NewPath, RegistryValue, CreatedAt, Status}` — starts `"in_progress"`, set to `"complete"` on success, `"rolled_back"` after rollback

**ProgressCallback:**
- Purpose: Cross-layer callback type for progress reporting — defined in `internal/migrate/migrate.go:14`, wrapped in `app.go:314` to translate `migrate.ProgressInfo` to `main.ProgressInfo` and emit Wails events
- Pattern: `func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)`

**Wails Binding Proxy (dual-mode):**
- Purpose: `frontend/src/lib/wails.ts` dynamically imports `../../wailsjs/go/main/App` at runtime; if import fails (dev mode without Wails), falls back to mock data
- Enables frontend development with `vite` without a running Go backend

## Entry Points

**Wails GUI Entry:**
- Location: `main.go` line 28 — `wails.Run(&options.App{...})`
- Triggers: User double-clicks `MoveKit.exe`
- Responsibilities: single-instance enforcement, embed frontend dist, register App bindings, start WebView2

**Wails Dev Entry:**
- Command: `wails dev`
- Dev server URL: `frontend:dev:serverUrl: "auto"` (configured in `wails.json:9`)
- Responsibilities: hot-reload frontend, launch Go backend with debug support

**CLI Verification Entry:**
- Location: `cmd/verify/main.go`
- Triggers: `go run ./cmd/verify`
- Responsibilities: Exercise all backend subsystems independently of the GUI — junction roundtrip, registry read, OneDrive detection, PreCheck, logging, cloud detection, backup loading

**Frontend Entry:**
- Location: `frontend/src/main.tsx:6` — `ReactDOM.createRoot(document.getElementById('root')!)`
- Triggers: WebView2 loads `frontend/dist/index.html`

## Architectural Constraints

- **Threading:** Go model (goroutines concurrency). Wails GUI runs on main goroutine. Migration runs in a separate goroutine (`app.go:280`). Single-instance mutex (`EnsureSingleInstance`) at startup. Migration lock via `sync.Mutex` in App struct.
- **Global state:** `migrate.cancelled` atomic flag (`internal/migrate/move.go:15`) — singleton for cancellation; `winsyscall.activeCopyCallback` global callback pointer (`internal/syscall/fileops.go:49`) — single active copy at a time; `winsyscall.modkernel32` lazy DLL handle (`internal/syscall/junction.go:31`)
- **Circular imports:** None detected. Package `main` imports `internal/*` packages. `internal/migrate` imports `internal/syscall`. `internal/syscall` imports only stdlib and `golang.org/x/sys`. Clean dependency DAG.
- **Type duplication:** Several types exist in both `internal/migrate` and `app_types.go` (package `main`) — e.g., `BackupRecord`, `MoveKitError`, `ProgressInfo`, `ConflictInfo`, `CloudProvider`, `MigrationState`, `UpdateInfo`. This avoids Go's circular import restriction (package `main` cannot be imported). The `main` types add JSON tags for Wails serialization; the `migrate` types are internal-only.
- **Windows-only:** The application is Windows-only (NTFS junctions, Windows registry, Win32 API). No cross-platform abstractions exist.

## Anti-Patterns

### Type Duplication Between Packages

**What happens:** Several types (`BackupRecord`, `MoveKitError`, `ProgressInfo`, `ConflictInfo`, `CloudProvider`, `MigrationState`, `UpdateInfo`) are defined in both `app_types.go` (package `main`) and `internal/migrate/types.go` + `internal/migrate/migrate.go` + `internal/migrate/errors.go`.
**Why it's wrong:** Changes to one definition require manual synchronization of the other. The `main` versions add JSON tags (for Wails serialization) that the `migrate` versions lack. The two `BackupRecord` structs differ slightly (`Folder`/`GUID` fields in `app_types.go` vs `RegValueName` in `migrate` — actually `migrate.BackupRecord` has `RegValueName` but `app_types.BackupRecord` has `Folder`/`GUID`/`RegistryValue`).
**Do this instead:** Define types in a shared `types` package that both `main` and `migrate` can import, or use the `migrate` types directly in `app_types.go` via type aliases that add JSON tags.

### Simulated Progress in Frontend

**What happens:** `App.tsx` line 139-181 has a `simulateProgress()` function that uses `setInterval` to mock progress updates. It runs after `StartMigration()` resolves, regardless of whether real Wails events are firing.
**Why it's wrong:** In a real Wails build, the Go backend emits `migration:progress` events that should drive the UI. The `simulateProgress` code races with real events and produces incorrect progress data. The comment at line 131 says "in real Wails this would come from events" but no Wails event listener is implemented for `migration:progress`.
**Do this instead:** Remove `simulateProgress`. In `App.tsx`, add `useEffect()` that listens for `runtime.EventsOn("migration:progress", ...)` and updates state from event data. Or, for a simpler approach, have the `StartMigration` goroutine emit all progress events and the event listener in the frontend drives the UI.

### Conservative SSD Detection

**What happens:** `GetVolumeInfo()` in `internal/syscall/volume.go:73` defaults `IsSSD` to `true` when `driveType == DRIVE_FIXED` — which is incorrect for spinning HDDs. `IsSSDDrive()` at line 130 exists as a separate function but always returns `false` ("Conservative default - can't reliably detect without admin").
**Why it's wrong:** All fixed drives are reported as SSD, inflating speed estimates and giving users false confidence about HDD migration performance. The separate `IsSSDDrive` function is unused.
**Do this instead:** Use `IOCTL_STORAGE_QUERY_PROPERTY` with `StorageDeviceSeekPenaltyProperty` or `StorageDeviceTrimProperty` to accurately detect SSD vs HDD without admin rights.

### Stub/Placeholder Methods

**What happens:** `App.ResolveConflict()` (`app.go:513`) always returns `nil` — a no-op placeholder. `migrate.CheckUpdate()` (`internal/migrate/version.go:11`) always returns `{HasUpdate: false, NewVersion: Version}`. `App.NotifyCompletion()` (`app.go:564`) emits a Wails event rather than using a system notification API (even though `go-toast` is in `go.mod` as an indirect dependency).
**Why it's wrong:** Conflict resolution is silently ignored during real migration. Update checking is non-functional. System notifications don't actually appear.
**Do this instead:** Implement conflict resolution during migration by pausing and emitting events to the frontend. Implement update check by querying GitHub Releases API. Enable `go-toast` for Windows toast notifications.

## Error Handling

**Strategy:** Structured errors with machine-readable codes throughout the migration layer.

**Patterns:**
- **PreCheck/validation errors** return `*MoveKitError` with code, message, detail, and suggested action (e.g., `internal/migrate/check.go:139-146`)
- **Panic recovery** in `Migrate()` wraps panics as `MoveKitError{Code: "INTERNAL_ERROR", ...}` (`internal/migrate/migrate.go:23-32`)
- **Registry errors** return `*MoveKitError{Code: ErrCodeRegistryFailed, ...}` with action advising admin privileges
- **Frontend error display:** `App.tsx` line 238-250 shows a danger-styled banner with close button; error state transition to `'failed'` shows `ResultPage`
- **Heartbeat monitoring:** `useHeartbeat` hook in `frontend/src/components/ErrorBoundary.tsx:92` polls `HealthCheck()` every 5 seconds during migration; after 3 consecutive failures, fires `onDead` callback transitioning to failed state

## Cross-Cutting Concerns

**Logging:** File-based structured logging to `%APPDATA%\MoveKit\logs\movekit-{date}.log`. `migrate.Logger` type (`internal/migrate/log.go:13`) with `Info`/`Warn`/`Error` levels. All log paths are masked via `maskPath()` (`log.go:83`) replacing `USERPROFILE` with `C:\Users\***`. Logger is per-migration (created in `Migrate()`, closed on return).

**Validation:** `migrate.PreCheck()` (`internal/migrate/check.go:30`) runs 6 sub-checks before migration: BitLocker status, Controlled Folder Access (ransomware protection), Windows Update pending reboot, battery power check, source existence, NTFS filesystem requirement, drive type validation (not removable/network), disk space sufficiency, target path availability. Security software detection runs but is informational only (not blocking).

**Authentication:** No authentication layer. The application requires administrator-equivalent privileges to write to User Shell Folders / Shell Folders registry keys. The app does not explicitly request elevation — registry write failures are surfaced as `MoveKitError` with action guidance.

**Single Instance:** Named mutex `"MoveKit_SingleInstance"` via `EnsureSingleInstance()` in `main.go:20` — if another instance holds the mutex, the process exits immediately.

---

*Architecture analysis: 2026-06-21*
