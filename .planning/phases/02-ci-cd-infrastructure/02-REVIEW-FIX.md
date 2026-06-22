---
phase: 02-ci-cd-infrastructure
fixed_at: 2026-06-22T00:00:00Z
review_path: .planning/phases/02-ci-cd-infrastructure/02-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 02: Code Review Fix Report

**Fixed at:** 2026-06-22
**Source review:** .planning/phases/02-ci-cd-infrastructure/02-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 6
- Fixed: 6
- Skipped: 0

## Fixed Issues

### CR-01: ldflags `-X` targets a `const` -- version injection silently fails

**Files modified:** `main.go`
**Commit:** 606bb8b
**Applied fix:** Changed `const currentVersion = "0.1.0"` to `var currentVersion = "0.1.0"`. Go linker's `-X` flag only works on `var` declarations; `const` values are compile-time constants that cannot be overridden at link time.

### CR-02: `IsSSD` incorrectly reports all fixed drives as SSD

**Files modified:** `internal/syscall/volume.go`, `app_types.go`, `app.go`, `cmd/verify/main.go`, `frontend/src/types.ts`, `frontend/src/lib/wails.ts`, `frontend/src/components/DiskSelector.tsx`, `frontend/wailsjs/go/models.ts`
**Commit:** c21b133
**Applied fix:** Renamed `IsSSD` to `IsFixedDrive` across all Go structs, JSON tags, TypeScript interfaces, Wailsjs bindings, and mock data. Removed the misleading SSD/HDD badge from `DiskSelector.tsx` and the HDD speed warning. Simplified speed estimation to use a single conservative 50 MB/s rate for all cross-volume moves (SSD vs HDD distinction not yet implemented). The `DurationEstimate` struct now carries `TargetIsFixedDrive` instead of `TargetIsSSD`. **Status: requires human verification** -- the speed estimation change removes the SSD doubling logic; confirm this is acceptable for the MVP.

### WR-01: `filepath.Abs` error silently ignored

**Files modified:** `internal/migrate/migrate.go`
**Commit:** 8bd1f79
**Applied fix:** Changed `absTarget, _ := filepath.Abs(targetPath)` to `absTarget, err := filepath.Abs(targetPath)` with an `if err != nil` block that falls back to `targetPath`. Prevents writing an empty string to the Shell Folders registry key when `filepath.Abs` fails.

### WR-02: `REG_EXPAND_SZ` used for fully-expanded path

**Files modified:** `internal/migrate/migrate.go`
**Commit:** 661eb02
**Applied fix:** Changed `sys.WriteExpandString(sys.UserShellFoldersKey, ...)` to `sys.WriteString(sys.UserShellFoldersKey, ...)` (REG_SZ). The registry value (`D:\Desktop`) contains no `%ENVVAR%` references, so REG_SZ is the semantically correct type. This matches the Shell Folders write at line 198 which already uses `WriteString`.

### WR-03: `Process32FirstW`/`Process32NextW` errors silently swallowed

**Files modified:** `internal/syscall/process.go`
**Commit:** 90a6456
**Applied fix:** Added explicit `syscall.GetLastError()` calls after `Process32FirstW` and `Process32NextW` return 0, with comments documenting that `ERROR_NO_MORE_FILES` (18) is the expected end-of-list signal while other errors indicate enumeration failure. Makes the error-handling intent explicit. **Status: requires human verification** -- the function still returns `false` on any error (same behavior); consider whether callers should be notified of enumeration failures for more accurate security software detection.

### WR-04: `MkdirAll` uses Unix permission `0755` on Windows

**Files modified:** `internal/migrate/migrate.go`
**Commit:** 08915e6
**Applied fix:** Changed `os.MkdirAll(sourcePath, 0755)` and `os.MkdirAll(backupDir, 0755)` to use `0777` (full Windows default ACL). Note: `0755` is also used in `app.go:478`, `internal/migrate/backup.go:27`, `internal/migrate/log.go:27`, `internal/migrate/state.go:21`, `internal/migrate/rollback.go:45`, and test files -- those were not flagged by the review and were left unchanged per surgical fix policy.

---

_Fixed: 2026-06-22T00:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
