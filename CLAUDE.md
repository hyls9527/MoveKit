<!-- GSD:project-start source:PROJECT.md -->
## Project

**MoveKit — C盘瘦身工具**

MoveKit 是一个 Windows 桌面应用，帮助用户一键迁移用户文件夹（桌面/下载/文档/图片/音乐）到其他盘。通过 NTFS Junction 技术实现100%兼容迁移，所有程序照常使用，无需改变使用习惯。单文件便携 exe，双击即用。

**Core Value:** **安全、一键完成 C盘用户文件夹瘦身，迁移后所有软件正常工作，可随时回滚。**

### Constraints

- **平台**: Windows 10/11 only（Win32 API 深度绑定）
- **目标盘**: NTFS only（Junction 依赖 NTFS 重解析点）
- **权限**: 用户级 HKCU 注册表写入，不要求管理员权限
- **兼容性**: 保持单 exe 便携分发，无安装器
- **语言**: Go 后端 + TypeScript 前端，不改技术栈
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.23.5 - Backend: Wails application runtime, Win32 syscall wrappers, migration engine. Entry points at `main.go`, `app.go`, `cmd/verify/main.go`.
- TypeScript 6.0.3 - Frontend: React UI components, type definitions, Wails bridge adapter. Source in `frontend/src/`.
- CSS (via Tailwind CSS v4) - Styling in `frontend/src/index.css` using `@theme` and `@import "tailwindcss"`.
- Python 3 (tools) - Debug/analysis utilities: `tools/frida_trace.py`, `tools/procmon_analyze.py`.
## Runtime
- Go 1.23.5 (go toolchain)
- Node.js 18+ (frontend build/dev dependency)
- Wails v2.12.0 (desktop framework — see `go.mod`)
- Windows only (Win32 API bindings throughout; no cross-platform abstractions)
- Go modules (`go.mod` / `go.sum`) — module path: `github.com/movekit/movekit`
- npm (npm install) — `frontend/package.json` with lockfile not present in repo
- Lockfile: `go.sum` present; no `package-lock.json` or `yarn.lock` detected
## Frameworks
- Wails v2.12.0 — Desktop application framework (Go backend + webview frontend). Config: `wails.json`. Generates JS bindings at `frontend/wailsjs/go/main/App.js` and `App.d.ts` from Go struct exports in `app.go`.
- React 19.2.7 — UI library. JSX mode: `react-jsx` (TypeScript compiler option). Root file: `frontend/src/App.tsx`.
- React DOM 19.2.7 — rendering via `ReactDOM.createRoot` in `frontend/src/main.tsx`.
- Go standard `testing` package — test files for `internal/syscall/` and `internal/migrate/` packages.
- No frontend test framework detected (no Jest, Vitest, or Testing Library dependencies in `frontend/package.json`).
- Vite 8.0.16 — Frontend bundler. Config: `frontend/vite.config.ts`. Plugins: `@vitejs/plugin-react`, `@tailwindcss/vite`.
- TypeScript compiler (`tsc`) — runs before Vite build as `tsc && vite build` per `frontend/package.json` scripts.
- Wails CLI (`wails dev`, `wails build`) — wraps Go build + Vite build. `go build -o MoveKit.exe .` for standalone binary.
- Tailwind CSS v4.3.1 — Utility-first CSS framework. Configured via `@theme` block in `frontend/src/index.css` and `@tailwindcss/vite` Vite plugin. No separate `tailwind.config.js` (v4 convention).
- PostCSS 8.5.15 / Autoprefixer 10.5.0 — CSS processing pipeline (standard for Vite + Tailwind).
## Key Dependencies
- `github.com/wailsapp/wails/v2` v2.12.0 — The Wails desktop framework. Provides the webview window, Go-to-JS bridge, event system, and asset server. All `app.go` bound methods are auto-generated into `frontend/wailsjs/go/main/`.
- `golang.org/x/sys` v0.30.0 — Extended Go syscall library. Used for Windows registry access (`golang.org/x/sys/windows/registry`) across `internal/syscall/registry.go` and `internal/migrate/check.go` (BitLocker, Controlled Folder Access, Windows Update detection).
- `github.com/wailsapp/go-webview2` v1.0.22 — Microsoft Edge WebView2 runtime integration (Windows webview).
- `github.com/labstack/echo/v4` v4.13.3 — HTTP framework (used internally by Wails for IPC/dev server).
- `github.com/gorilla/websocket` v1.5.3 — WebSocket transport (Wails IPC between Go and frontend).
- `github.com/go-ole/go-ole` v1.3.0 — Windows OLE/COM bindings (used by Wails for system tray, notifications, etc.).
- `github.com/godbus/dbus/v5` v5.1.0 — D-Bus bindings (Wails notification support; Linux path, present but unused on Windows).
- `github.com/jchv/go-winloader` — Windows executable loader (Wails internal).
- `github.com/samber/lo` v1.49.1 — Lodash-style Go generics library (Wails internal utility).
- `git.sr.ht/~jackmordaunt/go-toast/v2` v2.0.3 — Windows toast notification library (indirect via Wails).
## Configuration
- Go uses standard `go.mod` module system; no `.env` file detected.
- Frontend uses TypeScript `tsconfig.json` targeting ES2020 with strict mode, `module: ESNext`, bundler module resolution.
- Application config is file-based: `%APPDATA%\MoveKit\config.json` (managed by `app.go` `GetFirstRunFlag`/`SetFirstRunFlag`). Contains first-run flag only.
- Migration state tracked in `%APPDATA%\MoveKit\state.json` (`internal/migrate/state.go`).
- Backup records stored in `%APPDATA%\MoveKit\backups\{folder}.json` (`internal/migrate/backup.go`).
- Logs written to `%APPDATA%\MoveKit\logs\movekit-{date}.log` (`internal/migrate/log.go`).
- Single-instance enforcement via Windows named mutex "MoveKit_SingleInstance" (`internal/syscall/process.go`, `main.go:20`).
- Application version hardcoded as `0.1.0` in both `main.go:16` (`currentVersion`) and `internal/migrate/version.go:5` (`const Version`).
- `wails.json` — Wails project config (name: "MoveKit", output filename: "MoveKit", frontend dir: "frontend").
- `frontend/vite.config.ts` — Vite config with React and Tailwind plugins, output to `frontend/dist`.
- `frontend/tsconfig.json` — TypeScript config (strict, ES2020 target, react-jsx).
- `build/windows/info.json` — Wails build info template for Windows executable metadata.
- `build/windows/wails.exe.manifest` — Windows executable manifest.
- `build/appicon.png` / `build/windows/icon.ico` — Application icons.
- Go build: `go build -o MoveKit.exe .` (produces `MoveKit.exe` at repo root). Assets embedded via `//go:embed all:frontend/dist` in `main.go:14`.
## Platform Requirements
- Go 1.22+ (go.mod specifies 1.23.5; README says 1.22+)
- Node.js 18+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Windows OS (Win32 API calls throughout; cannot develop/test on macOS/Linux)
- Microsoft Edge WebView2 Runtime (typically pre-installed on Windows 10/11)
- Windows 10/11 (64-bit)
- NTFS target drive (junction points require NTFS — enforced in `internal/migrate/check.go:91-99`)
- Microsoft Edge WebView2 Runtime (Wails requirement for webview)
- Single portable executable: `MoveKit.exe` (~5.3 MB as built; embeds frontend assets)
- No installer needed — README confirms "便携版，单个 exe，双击即用"
- No admin rights required for read operations; registry writes may need user-level permissions (HKCU)
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Language Split
## Naming Patterns
### Go
- `main` for the root entry package (`main.go`, `app.go`, `app_types.go`)
- `winsyscall` for Win32 API wrappers (`internal/syscall/`) — not `syscall` to avoid shadowing Go's stdlib
- `migrate` for core migration logic (`internal/migrate/`)
- `folders` for read-only data definitions (`internal/folders/`)
- All `internal/` sub-packages use single-word lowercase names
- Exported functions: PascalCase (`Migrate`, `PreCheck`, `CreateJunction`, `IsJunction`)
- Unexported functions: camelCase (`backupFilePath`, `dirSize`, `checkBitLocker`, `resolveJunctionTarget`)
- Wails-bound methods on `App`: PascalCase with noun-verb pattern (`GetFolders`, `StartMigration`, `CheckOneDriveKFM`)
- Test functions: PascalCase with `Test` prefix (`TestCreateJunction`, `TestDetectConflicts`)
- Package-level: camelCase (`activeCopyCallback`, `modkernel32`, `AllFolders`)
- Constants: `ALL_CAPS` with underscores for error codes (`ErrCodeOneDriveKFM`), mixed for other constants (`IO_REPARSE_TAG_MOUNT_POINT`, `UserShellFoldersKey`)
- Local: short camelCase (`srcPath`, `driveLetter`, `filesMoved`)
- Struct names: PascalCase (`MoveKitError`, `BackupRecord`, `ProgressInfo`, `VolumeInfo`)
- Struct field JSON tags: camelCase (`json:"sourcePath"`, `json:"isSSD"`, `json:"hasUpdate"`)
- Interface: Not used for abstraction (only `interface{}` for generic Wails binding parameter)
### TypeScript / React
- Component files: PascalCase matching the component name (`FolderList.tsx`, `ErrorBoundary.tsx`, `DiskSelector.tsx`)
- Utility files: lowercase (`wails.ts`)
- Type files: lowercase (`types.ts`)
- Config files: conventional names (`vite.config.ts`, `tsconfig.json`)
- Components: `export default function PascalCase` (`FolderList`, `ProgressBar`, `ResultPage`)
- Hooks: camelCase with `use` prefix (`useHeartbeat`)
- Helper functions: camelCase (`formatBytes`, `simulateProgress`, `loadData`)
- Event handlers: camelCase with `handle` prefix (`handleStart`, `handleConfirm`, `handleDone`, `handleRollback`)
- camelCase everywhere (`showOneDriveWarning`, `selectedDisk`, `totalSize`, `fileCount`)
- State setters: `set` + PascalCase (`setState`, `setProgress`, `setShowDefenderInfo`)
- Interfaces: PascalCase (`FolderInfo`, `DiskInfo`, `ProgressInfo`, `ErrorBoundaryProps`)
- Type aliases: PascalCase (`AppState`, `ConflictFile`)
- Props: suffixed `Props` (`FolderListProps`, `DiskSelectorProps`, `ProgressBarProps`)
## Code Style
### Go
- Standard `go fmt` formatting (no formatter config files detected)
- Consistent spacing around operators and between logical blocks
- Line length: moderate, typically under 120 chars
- Go doc comments (`//` style) on every exported function, type, and constant
- Multi-line doc comments use consecutive `//` lines, not `/* */`
- Comments are in Chinese for user-facing messages, English for technical descriptions
- Section dividers: `// ----` with descriptive headers for grouping methods in `app.go`:
- Inline comments: used sparingly for non-obvious logic or flag explanations
### TypeScript / React
- No linter or formatter configured (no `.eslintrc`, `.prettierrc`, or `biome.json`)
- Consistent 2-space indentation used manually
- Trailing commas: not used
- Semicolons: not used
- JSX: double quotes for props, consistent indentation
- Section comments with emoji + dash: `// ── Heartbeat Hook ──`
- Inline comments in Chinese for UI text descriptions
- No JSDoc/TSDoc usage
- Component code is self-documenting; comments are rare outside section dividers
- All dialogs/overlays use this structure: `<fixed overlay div>` > `<card container>` > Header / Body / Footer sections
- Conditional rendering: `{condition && (<JSX>)}` pattern (never ternary for show/hide)
- State-based rendering: `{state === 'migrating' && (...)}` pattern in `App.tsx`
- Click handlers: inline arrow functions `onClick={() => handler()}` or direct reference `onClick={handler}` where no params needed
- Tailwind classes: grouped logically (layout, spacing, colors, typography, transitions)
## Import Organization
### Go
### TypeScript / React
## Error Handling
### Go
- `ONEDRIVE_KFM`, `SPACE_INSUFFICIENT`, `FILE_LOCKED`, `PERMISSION_DENIED`, `JUNCTION_FAILED`, `REGISTRY_FAILED`, `TARGET_EXISTS`, `FS_NOT_NTFS`, `NETWORK_DRIVE`, `REMOVABLE_DRIVE`, `BITLOCKER_LOCKED`, `CONTROLLED_FOLDER`, `WINDOWS_UPDATE`, `HYPERV_LOCKED`, `ONEDRIVE_SYNC`, `GPO_RESTRICTED`, `TARGET_IN_SOURCE`
### TypeScript / React
## Logging
### Go
- File-based: writes to `%APPDATA%\MoveKit\logs\movekit-{date}.log`
- Levels: `Info`, `Warn`, `Error` (no Debug/Fatal)
- Thread-safe: mutex-protected writes
- Timestamp format: `2006-01-02 15:04:05.000`
- Line format: `{timestamp} [LEVEL] {message}\n`
- **Privacy:** `maskPath()` replaces `USERPROFILE` prefix with `C:\Users\***`
## Function Design
### Go
- Maximum 4-5 parameters; progress callbacks passed as function types
- Named types for callbacks: `type ProgressCallback func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)`
- Wails-bound methods take primitives or simple structs (JSON-serialized over the bridge)
- Standard: `(result, error)` pattern
- PreCheck returns `*MoveKitError` (nil = success) rather than generic `error`
- Boolean returns for checks: `IsJunction(path string) bool`, `IsCancelled() bool`
### TypeScript / React
## Module Design
### Go
### TypeScript
## Section Separator Standards
### Go (`app.go`)
### TypeScript (`ErrorBoundary.tsx`)
## Mock Data Patterns (Frontend)
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## System Overview
```text
|                    React Frontend (TypeScript)                    |
|  App.tsx (state machine) + 11 components + lib/wails.ts          |
|  `frontend/src/`                                                  |
| Wails Runtime (v2.12.0)         | Wails-Generated Bindings       |
| Events + Bound Method Calls     | `frontend/wailsjs/go/main/App.js` |
|                    Go Backend (package main)                      |
|  main.go / app.go / app_types.go                                  |
|  App struct — 17 exported methods bound to frontend               |
| internal/migrate                | internal/syscall                |
| Core migration logic            | Win32 API wrappers              |
| - Migrate() pipeline            | - junction.go (DeviceIoControl) |
| - PreCheck() validation         | - registry.go (HKCU read/write)|
| - Backup/Rollback/Verify        | - fileops.go (CopyFileExW+ACL) |
| - OneDrive KFM detection        | - volume.go (GetVolumeInfo)    |
| - Cloud sync detection          | - process.go (toolhelp32)      |
| - Conflict detection            | - power.go (battery status)    |
| - Cancellation (atomic flag)    | - shell.go (SHChangeNotify)    |
| internal/folders                |                                |
| KnownFolder definitions + GUIDs |                                |
|                 Windows OS (NTFS / Registry / Shell)             |
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
### Primary Request Path: Migration Execution
### Cancellation Flow
- Frontend: React `useState` with explicit `AppState` enum (`'idle' | 'checking' | 'confirming' | 'migrating' | 'completed' | 'failed'`)
- Backend: `App.mu sync.Mutex` for migration lock, `App.migrating bool` for in-progress flag, `migrate.cancelled atomic.Int32` for cancellation
- Persistent state: `%APPDATA%\MoveKit\config.json` (first-run flag), `%APPDATA%\MoveKit\backups\{folder}.json` (backup records), `%APPDATA%\MoveKit\state.json` (migration state)
- Frontend-first-run: `localStorage.getItem('movekit-first-run-seen')` for guide overlay
## Key Abstractions
- Purpose: Structured error with machine-readable code, human-readable message (Chinese), technical detail, and user-facing action
- Examples: defined in `internal/migrate/errors.go:23` with error code constants at lines 3-21; duplicated in `app_types.go:88` (separate `package main` version to avoid circular imports)
- Pattern: `&MoveKitError{Code: "ONEDRIVE_KFM", Message: "OneDrive 已接管此文件夹", Detail: "...", Action: "请先关闭..."}`
- Purpose: Persistent JSON record of each migration, enabling rollback and integrity verification
- Examples: defined in `internal/migrate/migrate.go:218`; saved to `%APPDATA%\MoveKit\backups\{regValueName}.json`
- Pattern: `{Version, RegValueName, OriginalPath, NewPath, RegistryValue, CreatedAt, Status}` — starts `"in_progress"`, set to `"complete"` on success, `"rolled_back"` after rollback
- Purpose: Cross-layer callback type for progress reporting — defined in `internal/migrate/migrate.go:14`, wrapped in `app.go:314` to translate `migrate.ProgressInfo` to `main.ProgressInfo` and emit Wails events
- Pattern: `func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)`
- Purpose: `frontend/src/lib/wails.ts` dynamically imports `../../wailsjs/go/main/App` at runtime; if import fails (dev mode without Wails), falls back to mock data
- Enables frontend development with `vite` without a running Go backend
## Entry Points
- Location: `main.go` line 28 — `wails.Run(&options.App{...})`
- Triggers: User double-clicks `MoveKit.exe`
- Responsibilities: single-instance enforcement, embed frontend dist, register App bindings, start WebView2
- Command: `wails dev`
- Dev server URL: `frontend:dev:serverUrl: "auto"` (configured in `wails.json:9`)
- Responsibilities: hot-reload frontend, launch Go backend with debug support
- Location: `cmd/verify/main.go`
- Triggers: `go run ./cmd/verify`
- Responsibilities: Exercise all backend subsystems independently of the GUI — junction roundtrip, registry read, OneDrive detection, PreCheck, logging, cloud detection, backup loading
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
### Simulated Progress in Frontend
### Conservative SSD Detection
### Stub/Placeholder Methods
## Error Handling
- **PreCheck/validation errors** return `*MoveKitError` with code, message, detail, and suggested action (e.g., `internal/migrate/check.go:139-146`)
- **Panic recovery** in `Migrate()` wraps panics as `MoveKitError{Code: "INTERNAL_ERROR", ...}` (`internal/migrate/migrate.go:23-32`)
- **Registry errors** return `*MoveKitError{Code: ErrCodeRegistryFailed, ...}` with action advising admin privileges
- **Frontend error display:** `App.tsx` line 238-250 shows a danger-styled banner with close button; error state transition to `'failed'` shows `ResultPage`
- **Heartbeat monitoring:** `useHeartbeat` hook in `frontend/src/components/ErrorBoundary.tsx:92` polls `HealthCheck()` every 5 seconds during migration; after 3 consecutive failures, fires `onDead` callback transitioning to failed state
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

- **impeccable** (v3.8.0): Design language for AI coding agents. 23 commands for UI design, critique, audit, polish, and live iteration. Run `/impeccable <command>` to use. Setup: `PRODUCT.md` + `DESIGN.md` at project root.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
