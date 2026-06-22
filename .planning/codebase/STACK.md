# Technology Stack

**Analysis Date:** 2026-06-21

## Languages

**Primary:**
- Go 1.23.5 - Backend: Wails application runtime, Win32 syscall wrappers, migration engine. Entry points at `main.go`, `app.go`, `cmd/verify/main.go`.
- TypeScript 6.0.3 - Frontend: React UI components, type definitions, Wails bridge adapter. Source in `frontend/src/`.

**Secondary:**
- CSS (via Tailwind CSS v4) - Styling in `frontend/src/index.css` using `@theme` and `@import "tailwindcss"`.
- Python 3 (tools) - Debug/analysis utilities: `tools/frida_trace.py`, `tools/procmon_analyze.py`.

## Runtime

**Environment:**
- Go 1.23.5 (go toolchain)
- Node.js 18+ (frontend build/dev dependency)
- Wails v2.12.0 (desktop framework — see `go.mod`)
- Windows only (Win32 API bindings throughout; no cross-platform abstractions)

**Package Manager:**
- Go modules (`go.mod` / `go.sum`) — module path: `github.com/movekit/movekit`
- npm (npm install) — `frontend/package.json` with lockfile not present in repo
- Lockfile: `go.sum` present; no `package-lock.json` or `yarn.lock` detected

## Frameworks

**Core:**
- Wails v2.12.0 — Desktop application framework (Go backend + webview frontend). Config: `wails.json`. Generates JS bindings at `frontend/wailsjs/go/main/App.js` and `App.d.ts` from Go struct exports in `app.go`.
- React 19.2.7 — UI library. JSX mode: `react-jsx` (TypeScript compiler option). Root file: `frontend/src/App.tsx`.
- React DOM 19.2.7 — rendering via `ReactDOM.createRoot` in `frontend/src/main.tsx`.

**Testing:**
- Go standard `testing` package — test files for `internal/syscall/` and `internal/migrate/` packages.
- No frontend test framework detected (no Jest, Vitest, or Testing Library dependencies in `frontend/package.json`).

**Build/Dev:**
- Vite 8.0.16 — Frontend bundler. Config: `frontend/vite.config.ts`. Plugins: `@vitejs/plugin-react`, `@tailwindcss/vite`.
- TypeScript compiler (`tsc`) — runs before Vite build as `tsc && vite build` per `frontend/package.json` scripts.
- Wails CLI (`wails dev`, `wails build`) — wraps Go build + Vite build. `go build -o MoveKit.exe .` for standalone binary.
- Tailwind CSS v4.3.1 — Utility-first CSS framework. Configured via `@theme` block in `frontend/src/index.css` and `@tailwindcss/vite` Vite plugin. No separate `tailwind.config.js` (v4 convention).
- PostCSS 8.5.15 / Autoprefixer 10.5.0 — CSS processing pipeline (standard for Vite + Tailwind).

## Key Dependencies

**Critical:**
- `github.com/wailsapp/wails/v2` v2.12.0 — The Wails desktop framework. Provides the webview window, Go-to-JS bridge, event system, and asset server. All `app.go` bound methods are auto-generated into `frontend/wailsjs/go/main/`.
- `golang.org/x/sys` v0.30.0 — Extended Go syscall library. Used for Windows registry access (`golang.org/x/sys/windows/registry`) across `internal/syscall/registry.go` and `internal/migrate/check.go` (BitLocker, Controlled Folder Access, Windows Update detection).

**Infrastructure (all indirect via Wails v2):**
- `github.com/wailsapp/go-webview2` v1.0.22 — Microsoft Edge WebView2 runtime integration (Windows webview).
- `github.com/labstack/echo/v4` v4.13.3 — HTTP framework (used internally by Wails for IPC/dev server).
- `github.com/gorilla/websocket` v1.5.3 — WebSocket transport (Wails IPC between Go and frontend).
- `github.com/go-ole/go-ole` v1.3.0 — Windows OLE/COM bindings (used by Wails for system tray, notifications, etc.).
- `github.com/godbus/dbus/v5` v5.1.0 — D-Bus bindings (Wails notification support; Linux path, present but unused on Windows).
- `github.com/jchv/go-winloader` — Windows executable loader (Wails internal).
- `github.com/samber/lo` v1.49.1 — Lodash-style Go generics library (Wails internal utility).
- `git.sr.ht/~jackmordaunt/go-toast/v2` v2.0.3 — Windows toast notification library (indirect via Wails).

## Configuration

**Environment:**
- Go uses standard `go.mod` module system; no `.env` file detected.
- Frontend uses TypeScript `tsconfig.json` targeting ES2020 with strict mode, `module: ESNext`, bundler module resolution.
- Application config is file-based: `%APPDATA%\MoveKit\config.json` (managed by `app.go` `GetFirstRunFlag`/`SetFirstRunFlag`). Contains first-run flag only.
- Migration state tracked in `%APPDATA%\MoveKit\state.json` (`internal/migrate/state.go`).
- Backup records stored in `%APPDATA%\MoveKit\backups\{folder}.json` (`internal/migrate/backup.go`).
- Logs written to `%APPDATA%\MoveKit\logs\movekit-{date}.log` (`internal/migrate/log.go`).
- Single-instance enforcement via Windows named mutex "MoveKit_SingleInstance" (`internal/syscall/process.go`, `main.go:20`).
- Application version hardcoded as `0.1.0` in both `main.go:16` (`currentVersion`) and `internal/migrate/version.go:5` (`const Version`).

**Build:**
- `wails.json` — Wails project config (name: "MoveKit", output filename: "MoveKit", frontend dir: "frontend").
- `frontend/vite.config.ts` — Vite config with React and Tailwind plugins, output to `frontend/dist`.
- `frontend/tsconfig.json` — TypeScript config (strict, ES2020 target, react-jsx).
- `build/windows/info.json` — Wails build info template for Windows executable metadata.
- `build/windows/wails.exe.manifest` — Windows executable manifest.
- `build/appicon.png` / `build/windows/icon.ico` — Application icons.
- Go build: `go build -o MoveKit.exe .` (produces `MoveKit.exe` at repo root). Assets embedded via `//go:embed all:frontend/dist` in `main.go:14`.

## Platform Requirements

**Development:**
- Go 1.22+ (go.mod specifies 1.23.5; README says 1.22+)
- Node.js 18+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Windows OS (Win32 API calls throughout; cannot develop/test on macOS/Linux)
- Microsoft Edge WebView2 Runtime (typically pre-installed on Windows 10/11)

**Production:**
- Windows 10/11 (64-bit)
- NTFS target drive (junction points require NTFS — enforced in `internal/migrate/check.go:91-99`)
- Microsoft Edge WebView2 Runtime (Wails requirement for webview)
- Single portable executable: `MoveKit.exe` (~5.3 MB as built; embeds frontend assets)
- No installer needed — README confirms "便携版，单个 exe，双击即用"
- No admin rights required for read operations; registry writes may need user-level permissions (HKCU)

---

*Stack analysis: 2026-06-21*
