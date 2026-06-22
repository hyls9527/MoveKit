# Coding Conventions

**Analysis Date:** 2026-06-21

## Language Split

The codebase is split between Go (backend/Wails) and TypeScript + React + Tailwind CSS v4 (frontend). Conventions differ between the two. Each section below explicitly states which side it applies to.

---

## Naming Patterns

### Go

**Packages:**
- `main` for the root entry package (`main.go`, `app.go`, `app_types.go`)
- `winsyscall` for Win32 API wrappers (`internal/syscall/`) — not `syscall` to avoid shadowing Go's stdlib
- `migrate` for core migration logic (`internal/migrate/`)
- `folders` for read-only data definitions (`internal/folders/`)
- All `internal/` sub-packages use single-word lowercase names

**Functions:**
- Exported functions: PascalCase (`Migrate`, `PreCheck`, `CreateJunction`, `IsJunction`)
- Unexported functions: camelCase (`backupFilePath`, `dirSize`, `checkBitLocker`, `resolveJunctionTarget`)
- Wails-bound methods on `App`: PascalCase with noun-verb pattern (`GetFolders`, `StartMigration`, `CheckOneDriveKFM`)
- Test functions: PascalCase with `Test` prefix (`TestCreateJunction`, `TestDetectConflicts`)

**Variables:**
- Package-level: camelCase (`activeCopyCallback`, `modkernel32`, `AllFolders`)
- Constants: `ALL_CAPS` with underscores for error codes (`ErrCodeOneDriveKFM`), mixed for other constants (`IO_REPARSE_TAG_MOUNT_POINT`, `UserShellFoldersKey`)
- Local: short camelCase (`srcPath`, `driveLetter`, `filesMoved`)

**Types:**
- Struct names: PascalCase (`MoveKitError`, `BackupRecord`, `ProgressInfo`, `VolumeInfo`)
- Struct field JSON tags: camelCase (`json:"sourcePath"`, `json:"isSSD"`, `json:"hasUpdate"`)
- Interface: Not used for abstraction (only `interface{}` for generic Wails binding parameter)

**Import Aliases:**
```go
// Consistent pattern used across the codebase:
sys "github.com/movekit/movekit/internal/syscall"
```

### TypeScript / React

**Files:**
- Component files: PascalCase matching the component name (`FolderList.tsx`, `ErrorBoundary.tsx`, `DiskSelector.tsx`)
- Utility files: lowercase (`wails.ts`)
- Type files: lowercase (`types.ts`)
- Config files: conventional names (`vite.config.ts`, `tsconfig.json`)

**Functions:**
- Components: `export default function PascalCase` (`FolderList`, `ProgressBar`, `ResultPage`)
- Hooks: camelCase with `use` prefix (`useHeartbeat`)
- Helper functions: camelCase (`formatBytes`, `simulateProgress`, `loadData`)
- Event handlers: camelCase with `handle` prefix (`handleStart`, `handleConfirm`, `handleDone`, `handleRollback`)

**Variables:**
- camelCase everywhere (`showOneDriveWarning`, `selectedDisk`, `totalSize`, `fileCount`)
- State setters: `set` + PascalCase (`setState`, `setProgress`, `setShowDefenderInfo`)

**Types/Interfaces:**
- Interfaces: PascalCase (`FolderInfo`, `DiskInfo`, `ProgressInfo`, `ErrorBoundaryProps`)
- Type aliases: PascalCase (`AppState`, `ConflictFile`)
- Props: suffixed `Props` (`FolderListProps`, `DiskSelectorProps`, `ProgressBarProps`)

---

## Code Style

### Go

**Formatting:**
- Standard `go fmt` formatting (no formatter config files detected)
- Consistent spacing around operators and between logical blocks
- Line length: moderate, typically under 120 chars

**Comments:**
- Go doc comments (`//` style) on every exported function, type, and constant
- Multi-line doc comments use consecutive `//` lines, not `/* */`
- Comments are in Chinese for user-facing messages, English for technical descriptions
- Section dividers: `// ----` with descriptive headers for grouping methods in `app.go`:
```go
// ---------------------------------------------------------------------------
// Health & Version
// ---------------------------------------------------------------------------
```
- Inline comments: used sparingly for non-obvious logic or flag explanations

**Code Organization within files:**
1. Package declaration
2. Import block (stdlib first, blank line, third-party)
3. Constants
4. Variables
5. Types
6. Functions (exported first, then unexported)

### TypeScript / React

**Formatting:**
- No linter or formatter configured (no `.eslintrc`, `.prettierrc`, or `biome.json`)
- Consistent 2-space indentation used manually
- Trailing commas: not used
- Semicolons: not used
- JSX: double quotes for props, consistent indentation

**Comments:**
- Section comments with emoji + dash: `// ── Heartbeat Hook ──`
- Inline comments in Chinese for UI text descriptions
- No JSDoc/TSDoc usage
- Component code is self-documenting; comments are rare outside section dividers

**JSX Conventions:**
- All dialogs/overlays use this structure: `<fixed overlay div>` > `<card container>` > Header / Body / Footer sections
- Conditional rendering: `{condition && (<JSX>)}` pattern (never ternary for show/hide)
- State-based rendering: `{state === 'migrating' && (...)}` pattern in `App.tsx`
- Click handlers: inline arrow functions `onClick={() => handler()}` or direct reference `onClick={handler}` where no params needed
- Tailwind classes: grouped logically (layout, spacing, colors, typography, transitions)

---

## Import Organization

### Go

**Order:**
1. Standard library imports (`context`, `fmt`, `os`, `path/filepath`, `sync`, `syscall`, `time`, `unsafe`)
2. Blank line
3. Third-party imports (`github.com/wailsapp/wails/v2`)
4. Blank line
5. Internal imports (`github.com/movekit/movekit/internal/...`)

**Example from `app.go`:**
```go
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
```

### TypeScript / React

**Order:**
1. React hooks (`useState`, `useEffect`, `useCallback`, `useRef` from `react`)
2. Local types (`../types`)
3. Local lib (`./lib/wails`)
4. Component imports (`./components/...`)
5. Hook imports (if separate)

**Example from `App.tsx`:**
```typescript
import { useState, useEffect, useCallback, useRef } from 'react';
import { AppState, DiskInfo, FolderInfo, ProgressInfo } from './types';
import { GetFolders, GetDisks, HealthCheck, ... } from './lib/wails';
import DiskSelector from './components/DiskSelector';
import FolderList from './components/FolderList';
```

---

## Error Handling

### Go

**Primary pattern: Custom error type with structured fields.**
```go
type MoveKitError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail"`
    Action  string `json:"action"`
}

func (e *MoveKitError) Error() string { return e.Message }
```

**Error codes** are string constants defined in `internal/migrate/errors.go`:
- `ONEDRIVE_KFM`, `SPACE_INSUFFICIENT`, `FILE_LOCKED`, `PERMISSION_DENIED`, `JUNCTION_FAILED`, `REGISTRY_FAILED`, `TARGET_EXISTS`, `FS_NOT_NTFS`, `NETWORK_DRIVE`, `REMOVABLE_DRIVE`, `BITLOCKER_LOCKED`, `CONTROLLED_FOLDER`, `WINDOWS_UPDATE`, `HYPERV_LOCKED`, `ONEDRIVE_SYNC`, `GPO_RESTRICTED`, `TARGET_IN_SOURCE`

**Error wrapping:** `fmt.Errorf` with `%w` for standard errors, `MoveKitError` for domain errors.
```go
return fmt.Errorf("init logger: %w", err)
return &MoveKitError{Code: ErrCodeFileLocked, Message: "...", Detail: err.Error(), Action: "..."}
```

**Named return values** used for deferred error assignment in `Migrate()`:
```go
func Migrate(...) (retErr error) {
    defer func() {
        if r := recover(); r != nil {
            retErr = &MoveKitError{...}
        }
    }()
```

**No panics** in business logic. Only deferred `recover()` in the top-level `Migrate()` function converts panics to `MoveKitError`.

### TypeScript / React

**Pattern: Try-catch with Chinese error messages for users.**
```typescript
try {
  const [f, d] = await Promise.all([GetFolders(), GetDisks()]);
  setFolders(f);
} catch {
  setError('无法加载数据，请检查后端服务');
}
```

**Error boundary:** Class component `ErrorBoundary` wraps the entire app for render-time errors with a "restart" recovery option.

**Heartbeat monitoring:** `useHeartbeat` hook detects backend disconnection during migration and triggers error state.

---

## Logging

### Go

**Custom Logger** in `internal/migrate/log.go`:
- File-based: writes to `%APPDATA%\MoveKit\logs\movekit-{date}.log`
- Levels: `Info`, `Warn`, `Error` (no Debug/Fatal)
- Thread-safe: mutex-protected writes
- Timestamp format: `2006-01-02 15:04:05.000`
- Line format: `{timestamp} [LEVEL] {message}\n`
- **Privacy:** `maskPath()` replaces `USERPROFILE` prefix with `C:\Users\***`

**Usage pattern:**
```go
log, err := NewLogger()
if err != nil {
    return fmt.Errorf("init logger: %w", err)
}
defer log.Close()
log.Info("Starting migration of %s to %s", regValueName, targetDrive)
log.Error("PreCheck failed: %s — %s", err.Code, err.Detail)
```

**Frontend:** No logging framework. Uses `console.error` in `ErrorBoundary.componentDidCatch`.

---

## Function Design

### Go

**Size:** Most functions are 10-40 lines. The largest is `Migrate()` at ~200 lines (migration orchestrator with 9 distinct steps). `app.go` methods are moderate (10-60 lines), mainly delegating to internal packages.

**Parameters:**
- Maximum 4-5 parameters; progress callbacks passed as function types
- Named types for callbacks: `type ProgressCallback func(folder, phase string, filesMoved, totalFiles int, bytesMoved, totalBytes int64)`
- Wails-bound methods take primitives or simple structs (JSON-serialized over the bridge)

**Return Values:**
- Standard: `(result, error)` pattern
- PreCheck returns `*MoveKitError` (nil = success) rather than generic `error`
- Boolean returns for checks: `IsJunction(path string) bool`, `IsCancelled() bool`

### TypeScript / React

**Component Props:** All props defined as interfaces directly above the component.
```typescript
interface FolderListProps {
  folders: FolderInfo[];
}
export default function FolderList({ folders }: FolderListProps) { ... }
```

**State initialization pattern:** `useState` with explicit initial values, not lazy initializers.
```typescript
const [state, setState] = useState<AppState>('idle');
const [progress, setProgress] = useState<ProgressInfo>({ folder: '', phase: '准备中', ... });
```

---

## Module Design

### Go

**Exports:** All public functions are exported. No internal-only public APIs. Unused code is left in place (e.g., `moveFilesSameVolume` is defined but `Migrate()` uses `os.Rename` directly instead).

**Type duplication:** Types are duplicated between `app_types.go` (package main) and `internal/migrate/types.go` to avoid circular imports. Both define `ProgressInfo`, `ConflictInfo`, `CloudProvider`, `MigrationState`, `UpdateInfo` with identical JSON tags but separate struct definitions.

**Barrel files:** Not used. Each Go file is self-contained with its own imports.

### TypeScript

**Component exports:** All components use `export default function`. Named exports only for hooks and utilities (`useHeartbeat` from `ErrorBoundary.tsx`).

**Type sharing:** Shared types in `frontend/src/types.ts`. Component-specific types (like `ConflictFile`) defined in the component file.

**Barrel files:** Not used. Direct imports from source files.

**Mock fallback:** `frontend/src/lib/wails.ts` uses dynamic import with try-catch to fall back to mock data when Wails bindings are not available (dev mode without Wails runtime).

---

## Section Separator Standards

### Go (`app.go`)

```go
// ---------------------------------------------------------------------------
// Section Name
// ---------------------------------------------------------------------------
```

### TypeScript (`ErrorBoundary.tsx`)

```typescript
// ── Section Name ──
```

---

## Mock Data Patterns (Frontend)

`frontend/src/lib/wails.ts` provides mock implementations for all Wails-bound functions when running outside the Wails runtime:
```typescript
let realBindings: any = null;
try {
  realBindings = await import('../../wailsjs/go/main/App');
} catch {
  // No real bindings available — use mocks
}
// Each exported function: if (realBindings) return realBindings.X(); else return mockData;
```

Mock data is realistic (sample file sizes, drive labels, paths) and supports the full UI flow without the Go backend.

---

*Convention analysis: 2026-06-21*
