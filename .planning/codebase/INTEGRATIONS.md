# External Integrations

**Analysis Date:** 2026-06-21

## APIs & External Services

**No external HTTP/cloud APIs:** MoveKit is a fully offline desktop application. It does not call any remote APIs, SDKs, or cloud services.

**Update Check:**
- Placeholder implementation in `internal/migrate/version.go:11-17`. The `CheckUpdate()` function always returns `HasUpdate: false` with no network call. Intended to query GitHub Releases or a remote API in the future, but not implemented.

**System Notifications:**
- Wails notification system available via `frontend/wailsjs/runtime/runtime.d.ts` (e.g. `SendNotification`, `InitializeNotifications`). The `go-toast` library (`git.sr.ht/~jackmordaunt/go-toast/v2`) is an indirect Wails dependency for Windows toast notifications.
- Current notification implementation in `app.go:564-574` (`NotifyCompletion`) uses Wails runtime `EventsEmit` instead of native toast notifications. Placeholder pattern.

## Windows System APIs (Win32)

MoveKit directly calls the following Windows DLLs through Go's `syscall.NewLazyDLL` and `golang.org/x/sys/windows/registry`:

### kernel32.dll (via `internal/syscall/`)
| Function | File | Purpose |
|----------|------|---------|
| `CopyFileExW` | `internal/syscall/fileops.go:36` | Copy files with progress callback for cross-volume migration |
| `MoveFileExW` | `internal/syscall/fileops.go:37` | Move/rename files and directories |
| `CreateFileW` | `internal/syscall/junction.go:32` | Open directory handles for reparse point operations |
| `CloseHandle` | `internal/syscall/junction.go:33` | Close Windows file and mutex handles |
| `DeviceIoControl` | `internal/syscall/junction.go:31` | Issue FSCTL_SET_REPARSE_POINT, FSCTL_GET_REPARSE_POINT, FSCTL_DELETE_REPARSE_POINT for NTFS junction manipulation |
| `GetVolumeInformationW` | `internal/syscall/volume.go:20` | Query filesystem type (NTFS check), volume serial number, filesystem flags |
| `GetDriveTypeW` | `internal/syscall/volume.go:21` | Determine if a drive is fixed, removable, network, CD-ROM, or RAM disk |
| `GetDiskFreeSpaceExW` | `internal/syscall/volume.go:22` | Query available, total, and free disk space for capacity checks |
| `CreateMutexW` | `internal/syscall/process.go:12` | Create named mutex for single-instance enforcement |
| `CreateToolhelp32Snapshot` | `internal/syscall/process.go:13` | Enumerate running processes (security software detection) |
| `Process32FirstW` | `internal/syscall/process.go:14` | Get first process in snapshot |
| `Process32NextW` | `internal/syscall/process.go:15` | Iterate process entries |
| `GetSystemPowerStatus` | `internal/syscall/power.go:7` | Check AC power / battery status |
| `LocalFree` | `internal/syscall/fileops.go:38` | Free memory allocated by GetNamedSecurityInfoW |

### advapi32.dll (via `internal/syscall/`)
| Function | File | Purpose |
|----------|------|---------|
| `GetNamedSecurityInfoW` | `internal/syscall/fileops.go:37` | Read source file ACL (DACL) |
| `SetNamedSecurityInfoW` | `internal/syscall/fileops.go:38` | Apply ACL to destination file after copy |

### shell32.dll (via `internal/syscall/`)
| Function | File | Purpose |
|----------|------|---------|
| `SHChangeNotify` | `internal/syscall/shell.go:21` | Notify Windows Explorer of folder renames (`SHCNE_RENAMEFOLDER`) and directory updates (`SHCNE_UPDATEDIR`) after migration |

### Registry Access (via `golang.org/x/sys/windows/registry`)
| Key Path | Hive | File | Purpose |
|----------|------|------|---------|
| `Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders` | HKCU | `internal/syscall/registry.go:13` | Read/write REG_EXPAND_SZ values for known folder paths (Desktop, Documents, etc.) |
| `Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Folders` | HKCU | `internal/syscall/registry.go:17` | Read/write REG_SZ values for fully-expanded known folder paths |
| `Software\Microsoft\OneDrive\Accounts\Personal` | HKCU | `internal/migrate/onedrive.go:33-36` | Read `KfmFoldersMovedAll` DWORD to detect OneDrive Known Folder Move |
| `Software\Policies\Microsoft\OneDrive` | HKCU | `internal/migrate/onedrive.go:42-45` | Read `KFMBlockOptOut` DWORD to detect GPO-enforced OneDrive KFM |
| `SOFTWARE\Microsoft\Windows\CurrentVersion\BitLocker` | HKLM | `internal/migrate/check.go:170` | Check BitLocker encryption status (partial implementation) |
| `SOFTWARE\Policies\Microsoft\Windows Defender\Real-Time Protection` | HKLM | `internal/migrate/check.go:189-191` | Read `EnableControlledFolderAccess` DWORD for ransomware protection detection |
| `SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update` | HKLM | `internal/migrate/check.go:211-212` | Read `RebootRequired` DWORD for pending update reboot detection |

### NTFS Reparse Points (via kernel32 DeviceIoControl)

IOCTL codes used for junction creation/deletion:

| Control Code | Value | File | Purpose |
|-------------|-------|------|---------|
| `FSCTL_SET_REPARSE_POINT` | `0x000900A4` | `internal/syscall/junction.go:17` | Create an NTFS junction/mount point |
| `FSCTL_GET_REPARSE_POINT` | `0x000900A8` | `internal/syscall/junction.go:18` | Read reparse point data (junction detection) |
| `FSCTL_DELETE_REPARSE_POINT` | `0x000900AC` | `internal/syscall/junction.go:19` | Remove reparse point attribute (junction deletion) |

Reparse tag: `IO_REPARSE_TAG_MOUNT_POINT` = `0xA0000003` (`internal/syscall/junction.go:14`)

## Data Storage

**Databases:**
- None. No SQL, NoSQL, or embedded database. All persistence is file-based JSON.

**File Storage (local filesystem only):**
- `%APPDATA%\MoveKit\config.json` — Application configuration (`app.go:442-486`)
- `%APPDATA%\MoveKit\state.json` — In-progress migration state (`internal/migrate/state.go`)
- `%APPDATA%\MoveKit\backups\{RegValueName}.json` — Per-folder backup records (`internal/migrate/backup.go`)
- `%APPDATA%\MoveKit\logs\movekit-{YYYY-MM-DD}.log` — Dated migration logs (`internal/migrate/log.go`)
- `frontend/dist/` — Built frontend assets (committed at build time, embedded in Go binary via `//go:embed`)
- `%LOCALAPPDATA%\Microsoft\OneDrive` — Checked for existence only (OneDrive cloud sync detection in `internal/migrate/cloud.go:22-28`)
- `%LOCALAPPDATA%\Google\DriveFS` — Checked for existence only (Google Drive detection)
- `%APPDATA%\Apple Computer\MobileSync` — Checked for existence only (iCloud detection)

**Caching:**
- None. No in-memory or on-disk cache. Every folder size and disk info is computed fresh on each call.

## Authentication & Identity

**Auth Provider:**
- None. No user authentication. The application runs as the current Windows user and operates on HKCU registry hive and user-accessible files.

**Permission Model:**
- Runs with whatever privileges the current Windows user has. HKCU registry writes work without elevation. HKLM registry reads work without elevation. File operations require standard user permissions on the target drive. Junction creation requires write access to the source directory.

## Monitoring & Observability

**Error Tracking:**
- None (no Sentry, LogRocket, or similar error tracking service).

**Logs:**
- File-based structured logging to `%APPDATA%\MoveKit\logs\movekit-{date}.log` (`internal/migrate/log.go`).
- Log levels: INFO, WARN, ERROR.
- Log format: `{timestamp} [{LEVEL}] {message}`.
- User-sensitive paths are masked via `maskPath()` replacing `USERPROFILE` with `C:\Users\***`.
- Thread-safe via `sync.Mutex` on the Logger struct.
- Console output: `println()` in `main.go` for startup errors; `fmt.Printf` in `cmd/verify/main.go` for verification tool.

**Heartbeat:**
- Frontend-to-backend health check via `HealthCheck()` bound method (returns `true`). Frontend `ErrorBoundary` component (`frontend/src/components/ErrorBoundary.tsx`) monitors backend responsiveness with a heartbeat hook during migration.

## CI/CD & Deployment

**Hosting:**
- GitHub Releases (distribution channel, per README.md: "从 Releases 页面下载最新版 MoveKit.exe").

**CI Pipeline:**
- None detected. No GitHub Actions workflows, CI config files, or automated build scripts in the repository.

**Build Process (manual):**
1. `cd MoveKit/frontend && npx vite build` — Build frontend assets to `frontend/dist/`
2. `cd .. && go build -o MoveKit.exe .` — Compile Go binary with embedded frontend
3. Or use `wails build` for a full Wails production build with Windows metadata

**Binary Distribution:**
- Single portable `.exe` file (`MoveKit.exe`, ~5.3 MB). No installer, no MSI, no dependencies to install. Windows Defender / SmartScreen warning noted in README for first-time users.

## Environment Configuration

**Required env vars (read, not set, by the application):**
- `APPDATA` — Used to locate config/log/backup directories (`%APPDATA%\MoveKit\`)
- `LOCALAPPDATA` — Used for cloud sync detection paths
- `USERPROFILE` — Used as fallback for `APPDATA` and `LOCALAPPDATA`, and for path masking in logs

**No env file or secrets:** All configuration is file-based or registry-based. No `.env` file or secret management.

**Secrets location:**
- Not applicable (no API keys, credentials, or secrets used).

## Webhooks & Callbacks

**Incoming:**
- None. No HTTP server, no webhook endpoints.

**Outgoing:**
- None. No outbound HTTP calls.

**Wails IPC (internal bridge):**
- Frontend-to-backend: Go methods bound via Wails `Bind` in `main.go:38-40`. The `App` struct exports 17 methods auto-generated to `frontend/wailsjs/go/main/App.js`.
- Backend-to-frontend: Wails `runtime.EventsEmit()` for events: `"migration:progress"`, `"migration:done"`, `"notification"` (`app.go:292-356, 568`).
- WebSocket transport: Wails uses `github.com/gorilla/websocket` internally for IPC communication between Go and the WebView2 frontend.

## Cloud Sync Provider Detection

MoveKit detects running cloud sync providers via filesystem heuristics (not API calls):

| Provider | Detection Method | File |
|----------|-----------------|------|
| OneDrive | `%LOCALAPPDATA%\Microsoft\OneDrive` directory existence + OneDrive subdirectories in `USERPROFILE` | `internal/migrate/cloud.go:22-81` |
| Google Drive | `%LOCALAPPDATA%\Google\DriveFS` directory existence | `internal/migrate/cloud.go:32-38` |
| iCloud | `%APPDATA%\Apple Computer\MobileSync` directory existence | `internal/migrate/cloud.go:47-53` |

Additional OneDrive Known Folder Move detection (`internal/migrate/onedrive.go:22-52`):
- Checks User Shell Folders registry values for "OneDrive" in path
- Reads `KfmFoldersMovedAll` DWORD under OneDrive's HKCU registry key
- Checks `KFMBlockOptOut` GPO policy

## Security Software Detection

Detected by process enumeration (`internal/migrate/check.go:240-257`, `internal/syscall/process.go:74-101`):

| Executable | Display Name |
|-----------|-------------|
| `360sd.exe` | 360 杀毒 |
| `360tray.exe` | 360 安全卫士 |
| `HuoRong.exe` | 火绒安全 |
| `avp.exe` | 卡巴斯基 |
| `MsMpEng.exe` | Windows Defender |
| `avgui.exe` | AVG |
| `avguard.exe` | Avira |
| `bdagent.exe` | Bitdefender |

Detection is informational only (logged, not blocking).

---

*Integration audit: 2026-06-21*
