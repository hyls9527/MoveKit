# Codebase Structure

**Analysis Date:** 2026-06-21

## Directory Layout

```
MoveKit/
├── .gitignore                          # Build artifact exclusion (*.exe, build/bin/)
├── .planning/                          # GSD planning artifacts (auto-generated)
│   └── codebase/                       # Codebase analysis documents
├── README.md                           # Project readme (Chinese)
├── go.mod                              # Go module definition (github.com/movekit/movekit)
├── go.sum                              # Go dependency checksums
├── wails.json                          # Wails v2 project configuration
├── main.go                             # Application entry point, Wails Run()
├── app.go                              # App struct — 17 Wails-bound methods
├── app_types.go                        # Shared Go types with JSON tags (Wails-facing)
├── MoveKit.exe                         # Compiled binary (ignored in .gitignore but present)
│
├── cmd/                                # CLI tools (separate from Wails GUI)
│   └── verify/
│       └── main.go                     # Backend runtime verification CLI (15 subsystem checks)
│
├── internal/                           # Internal Go packages (no external import)
│   ├── folders/
│   │   └── known.go                    # KnownFolder struct, AllFolders slice, GUIDToRegValue map
│   │
│   ├── migrate/                        # Core migration business logic
│   │   ├── types.go                    # Internal types (ProgressInfo, ConflictInfo, CloudProvider, etc.)
│   │   ├── errors.go                   # MoveKitError type + 20 error code constants
│   │   ├── migrate.go                  # Migrate() — the main migration pipeline (backup→move→junction→registry→shell)
│   │   ├── check.go                    # PreCheck() — 6 validation checks before migration
│   │   ├── backup.go                   # SaveBackup() / LoadBackup() / LoadAllBackups() / DeleteBackup() / CleanupExpiredBackups()
│   │   ├── rollback.go                 # Rollback() — reverse migration (junction delete → move back → registry restore)
│   │   ├── move.go                     # moveFilesCrossVolume() with cancellation + moveFilesSameVolume()
│   │   ├── conflict.go                 # DetectConflicts() — filepath.Walk comparing src vs target
│   │   ├── onedrive.go                 # DetectOneDriveKFM() — OneDrive Known Folder Move detection
│   │   ├── cloud.go                    # DetectCloudSync() — OneDrive/iCloud/Google Drive process detection
│   │   ├── security.go                 # CheckJunctionHijack() — junction hijack detection
│   │   ├── verify.go                   # VerifyIntegrity() — file count/size verification after move
│   │   ├── state.go                    # SaveState() / LoadState() / ClearState() — migration state JSON
│   │   ├── log.go                      # Logger — file-based structured logging with path masking
│   │   ├── version.go                  # Version constant + CheckUpdate() placeholder
│   │   ├── check_test.go               # PreCheck tests (no-panic verification, maskPath)
│   │   ├── backup_test.go              # Backup CRUD tests
│   │   ├── conflict_test.go            # Conflict detection tests
│   │   ├── move_test.go                # Move/cancellation tests
│   │   └── onedrive_test.go            # OneDrive KFM detection tests
│   │
│   └── syscall/                        # Win32 API wrappers (package winsyscall)
│       ├── junction.go                 # CreateJunction() / IsJunction() / DeleteJunction() via DeviceIoControl
│       ├── registry.go                 # ReadExpandString() / WriteExpandString() / WriteString() / ReadInteger() / ReadIntegerLM()
│       ├── fileops.go                  # CopyFileEx() / CopyFileWithACL() / MoveFileEx() / IsCrossVolume() / ToLongPath()
│       ├── volume.go                   # GetVolumeInfo() / GetDriveType() / GetDiskFreeSpace() / IsSSDDrive()
│       ├── process.go                  # EnsureSingleInstance() / CloseHandle() / IsProcessRunning()
│       ├── power.go                    # GetPowerStatus() / IsOnBattery()
│       ├── shell.go                    # SHChangeNotify wrapper (NotifyFolderRenamed / NotifyDirUpdated)
│       ├── junction_test.go            # Junction create/is/delete roundtrip tests
│       ├── junction_debug_test.go      # Junction debugging diagnostics
│       ├── fileops_test.go             # File operation tests (CopyFileEx, IsCrossVolume, ToLongPath)
│       ├── fileops_acl_test.go         # ACL preservation tests
│       ├── registry_test.go            # Registry read/write tests
│       ├── volume_test.go              # Volume/disk info tests
│       └── power_test.go               # Power/battery status tests
│
├── frontend/                           # React + TypeScript + Tailwind CSS v4 (Vite)
│   ├── .gitignore                      # node_modules/, .vite/, *.local
│   ├── index.html                      # Root HTML — loads /src/main.tsx
│   ├── package.json                    # NPM config (react 19, tailwindcss 4, vite 8, typescript 6)
│   ├── package.json.md5                # Wails-generated MD5 of package.json
│   ├── package-lock.json               # NPM lockfile
│   ├── tsconfig.json                   # TypeScript config (ES2020, React JSX, strict mode)
│   ├── vite.config.ts                  # Vite config (React + Tailwind CSS v4 plugins)
│   │
│   ├── src/
│   │   ├── main.tsx                    # React entry — ReactDOM.createRoot renders <App/>
│   │   ├── App.tsx                     # Root component — state machine (AppState enum), all dialog overlays, heartbeat monitoring
│   │   ├── types.ts                    # TypeScript type interfaces (FolderInfo, DiskInfo, ProgressInfo, AppState)
│   │   ├── index.css                   # Tailwind CSS v4 imports + custom @theme tokens (colors, radii, shadows, fonts)
│   │   ├── vite-env.d.ts               # Vite type declarations
│   │   │
│   │   ├── lib/
│   │   │   └── wails.ts                # Wails binding proxy — dynamic import of real bindings or mock fallback
│   │   │
│   │   └── components/
│   │       ├── DiskSelector.tsx         # Target disk radio button list with recommendations
│   │       ├── FolderList.tsx           # Folder status cards (default/migrated/not_on_c/unknown/junction_broken)
│   │       ├── ProgressBar.tsx          # Migration progress — byte bar + file count bar + cancel
│   │       ├── ResultPage.tsx           # Migration result — success/failure, freed space, per-folder rollback
│   │       ├── ErrorBoundary.tsx        # React error boundary class + useHeartbeat hook
│   │       ├── OneDriveWarning.tsx      # Dialog — OneDrive KFM detected, disable instructions
│   │       ├── CloudWarning.tsx         # Dialog — cloud sync provider warnings (OneDrive/iCloud/GDrive)
│   │       ├── ConflictDialog.tsx       # Dialog — file conflict table with skip/overwrite/cancel
│   │       ├── DefenderInfo.tsx         # Dialog — Windows Defender/SmartScreen whitelist instructions
│   │       ├── FirstRunGuide.tsx        # Dialog — 3-step onboarding carousel
│   │       ├── HelpPanel.tsx            # Slide-out panel — FAQ accordion + GitHub links
│   │       └── UpdateNotice.tsx         # Dialog — version comparison, download link
│   │
│   ├── wailsjs/                        # Auto-generated by Wails (DO NOT EDIT manually)
│   │   ├── go/
│   │   │   └── main/
│   │   │       ├── App.js              # Generated JS stubs for all Go App methods
│   │   │       └── App.d.ts            # Generated TS declarations for all Go App methods
│   │   ├── go/
│   │   │   └── models.ts              # Generated TS class wrappers for Go structs
│   │   └── runtime/
│   │       ├── package.json            # Wails runtime npm package
│   │       ├── runtime.js              # Wails runtime library
│   │       └── runtime.d.ts            # Wails runtime type declarations
│   │
│   └── dist/                           # Built frontend output (gitignored)
│
├── tools/                              # Python analysis/utility scripts
│   ├── frida_trace.py                  # Frida dynamic instrumentation trace script
│   └── procmon_analyze.py              # Process Monitor log analysis script
│
└── build/                              # Build output directory (gitignored)
```

## Directory Purposes

**Root (`./`):**
- Purpose: Go module root, Wails project configuration, main entry points, readme
- Contains: `go.mod`, `wails.json`, `main.go`, `app.go`, `app_types.go`, `README.md`
- Key files: `main.go` (Wails bootstrap), `app.go` (frontend-bound methods), `app_types.go` (Wails-facing types)

**`cmd/verify/`:**
- Purpose: CLI verification tool — exercises all backend subsystems independently of the Wails GUI
- Contains: `main.go` — standalone Go binary with 15 subsystem checks (GetFolders simulation, GetDisks, OneDrive KFM, PreCheck, IsCrossVolume, ToLongPath, DetectConflicts, DetectCloudSync, LoadAllBackups, CheckUpdate, Logger, IsOnBattery, Junction roundtrip, Registry reads)
- Key files: `cmd/verify/main.go`

**`internal/folders/`:**
- Purpose: Static data — the 5 known Windows user folder definitions
- Contains: `known.go` — `KnownFolder` struct, `AllFolders` slice, `GUIDToRegValue` map
- Key files: `known.go` (24 lines, pure data)

**`internal/migrate/`:**
- Purpose: Core migration business logic — everything between the Wails App layer and the syscall layer
- Contains: 17 Go source files + 5 test files. Pipeline orchestration, pre-flight validation, file move, backup/rollback, OneDrive detection, cloud sync detection, conflict detection, cancellation, structured logging, state persistence, versioning
- Key files: `migrate.go` (264 lines, main pipeline), `check.go` (277 lines, 6 pre-flight checks), `move.go` (159 lines, cross-volume move with cancellation), `backup.go` (108 lines, CRUD), `rollback.go` (93 lines)

**`internal/syscall/`:**
- Purpose: Low-level Windows API wrappers — the only package that makes direct Win32 calls
- Contains: 7 Go source files + 6 test files. Junction creation/deletion/checking (DeviceIoControl), registry read/write (both HKCU and HKLM), file copy with ACL preservation (CopyFileExW + Get/SetNamedSecurityInfoW), volume/disk info, process enumeration (toolhelp32), power status, shell change notifications
- Key files: `junction.go` (246 lines), `fileops.go` (189 lines), `registry.go` (100 lines), `volume.go` (152 lines)

**`frontend/`:**
- Purpose: React frontend — the UI layer rendered in WebView2
- Contains: `package.json`, `tsconfig.json`, `vite.config.ts`, `index.html`, `src/` (React app), `wailsjs/` (auto-generated bindings), `dist/` (build output)
- Key files: `src/App.tsx` (411 lines, state machine + all dialogs), `src/lib/wails.ts` (99 lines, binding proxy), `src/types.ts` (33 lines)

**`frontend/src/components/`:**
- Purpose: Reusable UI components — one component per distinct UI state or dialog
- Contains: 11 React components (`.tsx` files) + `ErrorBoundary` (class component + hook)
- Component count: 12 including `ErrorBoundary`

**`frontend/wailsjs/`:**
- Purpose: Wails auto-generated binding code — bridges Go backend methods to JavaScript
- Contains: `go/main/App.js` (JS stubs), `go/main/App.d.ts` (TS declarations), `go/models.ts` (TS class wrappers), `runtime/` (Wails runtime library)
- Generated: Yes — by `wails build` or `wails generate module`
- Committed: Yes — required for type checking during frontend dev

**`tools/`:**
- Purpose: Utility scripts for analysis and debugging
- Contains: `frida_trace.py` (Frida dynamic instrumentation), `procmon_analyze.py` (Process Monitor log analysis)

**`build/`:**
- Purpose: Wails build output
- Contains: Compiled binaries and build artifacts
- Generated: Yes
- Committed: No (gitignored)

**`frontend/dist/`:**
- Purpose: Vite build output — the compiled frontend assets embedded into the Go binary
- Contains: HTML, JS, CSS bundles
- Generated: Yes
- Committed: No (gitignored, embedded via Go's `//go:embed` in main.go)

## Key File Locations

**Entry Points:**
- `main.go`: Wails application entry — single instance mutex, embed FS, `wails.Run()`
- `frontend/src/main.tsx`: React root — `ReactDOM.createRoot`, renders `<App/>`
- `cmd/verify/main.go`: Standalone CLI verification — independent of Wails GUI

**Configuration:**
- `wails.json`: Wails v2 project config — window title/size, frontend build commands, dev server
- `go.mod`: Go module definition — `github.com/movekit/movekit`, Go 1.23.5, two direct dependencies
- `frontend/package.json`: NPM config — React 19, Tailwind CSS v4, Vite 8, TypeScript 6
- `frontend/tsconfig.json`: TypeScript config — ES2020, strict mode, React JSX preserve
- `frontend/vite.config.ts`: Vite config — React plugin, Tailwind CSS v4 plugin
- `%APPDATA%/MoveKit/config.json`: Runtime config — first-run flag
- `%APPDATA%/MoveKit/backups/{folder}.json`: Migration backup records
- `%APPDATA%/MoveKit/state.json`: Migration state persistence
- `%APPDATA%/MoveKit/logs/movekit-{date}.log`: Structured migration logs

**Core Logic:**
- `app.go`: `App` struct — 17 exported methods (GetFolders, GetDisks, StartMigration, Rollback, etc.)
- `app_types.go`: All Wails-facing Go types with JSON tags
- `internal/migrate/migrate.go`: `Migrate()` — the main 9-step migration pipeline
- `internal/migrate/check.go`: `PreCheck()` — 6 pre-flight validation checks
- `internal/migrate/move.go`: `moveFilesCrossVolume()` — cross-volume file copy with cancellation
- `internal/syscall/junction.go`: `CreateJunction()` / `IsJunction()` / `DeleteJunction()` — DeviceIoControl wrappers
- `internal/syscall/registry.go`: Registry read/write for User Shell Folders and Shell Folders

**Testing:**
- Go: All test files co-located with source files in `internal/migrate/` and `internal/syscall/`
- Test run command: `go test ./... -v`
- No frontend test framework configured — no Jest, Vitest, or React Testing Library in `package.json`

## Naming Conventions

**Files:**
- Go: lowercase, `snake_case.go` for packages (e.g., `app_types.go`, `known.go`). All Go file names in `internal/` use lowercase.
- TypeScript/React: PascalCase for components (`DiskSelector.tsx`, `FolderList.tsx`), lowercase for utilities (`wails.ts`, `types.ts`), `vite-env.d.ts` for type declarations
- Test files: `*_test.go` suffix, co-located with source (Go convention)
- Config files: standard ecosystem names (`package.json`, `tsconfig.json`, `vite.config.ts`, `wails.json`, `.gitignore`)

**Directories:**
- Go: lowercase, no hyphens (`internal/migrate`, `internal/syscall`, `internal/folders`, `cmd/verify`)
- Frontend: lowercase, no hyphens (`frontend/src`, `frontend/src/components`, `frontend/src/lib`)

**Go packages:**
- `main` — entry point + App
- `folders` — known folder data
- `migrate` — core migration logic (compound verb)
- `winsyscall` — Win32 API wrappers (package name differs from directory `syscall` to avoid shadowing stdlib `syscall`)

**TypeScript exports:**
- React components: `export default function ComponentName`
- Utility/binding functions: `export async function FunctionName()` — PascalCase, async
- Types/interfaces: exported PascalCase interfaces (`FolderInfo`, `DiskInfo`, `ProgressInfo`)
- `AppState`: TypeScript type alias (not enum) — `'idle' | 'checking' | 'confirming' | 'migrating' | 'completed' | 'failed'`

**Go exported symbols:**
- Types: PascalCase (`FolderInfo`, `BackupRecord`, `MoveKitError`)
- Functions: PascalCase (`Migrate`, `PreCheck`, `CreateJunction`, `IsOnBattery`)
- Constants: either PascalCase (`UserShellFoldersKey`) or SCREAMING_SNAKE_CASE for OS-level values (`IO_REPARSE_TAG_MOUNT_POINT`, `SHCNE_RENAMEFOLDER`)
- Error codes: `ErrCode` prefix, SCREAMING_SNAKE_CASE (`ErrCodeOneDriveKFM`, `ErrCodeFileLocked`)

## Where to Add New Code

**New Feature (Go backend API):**
- Primary code: New method on `App` struct in `app.go` — every exported method is automatically bound to the frontend
- Types for the new method: Add to `app_types.go` with JSON tags
- Business logic: New file or function in `internal/migrate/`
- OS-level calls: New file or function in `internal/syscall/`
- Frontend binding wrapper: Add export in `frontend/src/lib/wails.ts` (both real binding call and mock fallback)
- Frontend types: Add TypeScript interface in `frontend/src/types.ts`

**New Component (React frontend):**
- Implementation: `frontend/src/components/NewComponent.tsx` — export default, PascalCase filename
- Import in `App.tsx` and add rendering logic within the state machine flow
- Component props: Define interface inline in the component file (no separate props directory)
- Styling: Tailwind CSS v4 utility classes + custom `@theme` tokens in `index.css` if new design tokens needed

**New CLI Tool:**
- New directory under `cmd/{toolname}/` with its own `main.go`
- Import `internal/migrate` and `internal/syscall` — no dependence on `main` package

**Utilities:**
- Shared Go helpers: `internal/migrate/` (for migration-adjacent logic) or `internal/syscall/` (for new Win32 wrappers)
- Shared frontend helpers: `frontend/src/lib/` (alongside `wails.ts`)

## Special Directories

**`frontend/wailsjs/`:**
- Purpose: Auto-generated Wails bindings — bridges between Go backend and JS frontend
- Generated: Yes — by `wails build` or `wails generate module`
- Committed: Yes — needed for TypeScript type checking during frontend development with `vite`
- Do not manually edit: The header in `App.d.ts` reads "This file is automatically generated. DO NOT EDIT"

**`frontend/dist/`:**
- Purpose: Vite build output — compiled HTML/JS/CSS that gets embedded into the Go binary via `//go:embed`
- Generated: Yes — by `npx vite build` (triggered by `wails build` or `frontend:build` in `wails.json`)
- Committed: No — gitignored; re-generated each build

**`build/`:**
- Purpose: Wails build artifacts — compiled `.exe` and supporting files
- Generated: Yes — by `wails build`
- Committed: No — gitignored (but `MoveKit.exe` in root is also gitignored)

**`tools/`:**
- Purpose: Python analysis scripts — not part of the build
- Generated: No
- Committed: Yes

---

*Structure analysis: 2026-06-21*
