---
phase: 02-ci-cd-infrastructure
reviewed: 2026-06-22T00:00:00Z
depth: standard
files_reviewed: 12
files_reviewed_list:
  - .github/dependabot.yml
  - .github/workflows/ci.yml
  - .golangci.yml
  - .npmrc
  - frontend/eslint.config.js
  - frontend/package.json
  - internal/migrate/migrate.go
  - internal/syscall/junction_test.go
  - internal/syscall/process.go
  - internal/syscall/process_test.go
  - internal/syscall/volume.go
  - README.md
findings:
  critical: 2
  warning: 4
  info: 5
  total: 11
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-06-22
**Depth:** standard
**Files Reviewed:** 12
**Status:** issues_found

## Summary

Reviewed 12 files spanning CI/CD infrastructure (workflows, linter config, dependabot), frontend tooling (ESLint, npm), backend migration/syscall code, and documentation. Two critical issues were found: a silent ldflags failure in the CI release job (Go `const` cannot be overridden), and an `IsSSD` field that reports all fixed drives as SSD, misleading the UI. Four warnings cover swallowed errors, incorrect registry value types, and permission conventions on Windows. Five informational items cover dead code, overly broad linter exclusions, and documentation inaccuracies.

## Critical Issues

### CR-01: ldflags `-X` targets a `const` -- version injection silently fails

**File:** `.github/workflows/ci.yml:165` and `main.go:16`
**Issue:** The release job injects the git tag version via `-ldflags "-X main.currentVersion=$version"`, but `currentVersion` is declared as `const currentVersion = "0.1.0"` in `main.go:16`. Go's `-X` linker flag only works on `var` declarations, not `const`. The build succeeds but the binary is always stamped with `"0.1.0"` regardless of the git tag. This means every release has the wrong version number.

**Fix:**
```go
// main.go:16 -- change const to var
var currentVersion = "0.1.0"
```

### CR-02: `IsSSD` incorrectly reports all fixed drives as SSD

**File:** `internal/syscall/volume.go:73`
**Issue:** The `GetVolumeInfo` function sets `IsSSD: driveType == DRIVE_FIXED`, which classifies ALL fixed drives (including spinning HDDs) as SSD. This field is consumed in `app.go:185` to determine migration speed estimates (`speedMBps := 100` for "SSD" vs `50` for HDD) and shown in the UI disk selector. Users running on an HDD target will see "SSD" labels and receive optimistic (doubled) speed estimates. The `IsSSDDrive` function at line 130 exists as a separate attempt but always returns `false`, creating two inconsistent approaches.

**Fix:** Either implement proper SSD detection via `DeviceIoControl` with `IOCTL_STORAGE_QUERY_PROPERTY` (specifically `StorageDeviceSeekPenaltyProperty` or `StorageDeviceTrimProperty`), or rename the field to `IsFixedDrive` to avoid misleading callers. The comment at line 71-72 already acknowledges the heuristic is inadequate.

## Warnings

### WR-01: `filepath.Abs` error silently ignored -- empty string may be written to registry

**File:** `internal/migrate/migrate.go:197`
**Issue:** `absTarget, _ := filepath.Abs(targetPath)` silently ignores the error from `filepath.Abs`. If `targetPath` is somehow invalid, `absTarget` could be an empty string `""`, which is then written to `ShellFoldersKey` as `REG_SZ` via `sys.WriteString`. This would corrupt the Shell Folders registry value, breaking folder redirection for the user. The comment says "fallback to targetPath" but no fallback logic is present.

**Fix:**
```go
absTarget, err := filepath.Abs(targetPath)
if err != nil {
    absTarget = targetPath // fallback as documented
}
```

### WR-02: `REG_EXPAND_SZ` used for fully-expanded path without environment variables

**File:** `internal/migrate/migrate.go:185-186`
**Issue:** `sys.WriteExpandString` writes `REG_EXPAND_SZ` but the value (`D:\Desktop`) contains no `%ENVVAR%` references. Windows convention reserves `REG_EXPAND_SZ` for paths that need environment variable expansion (e.g., `%USERPROFILE%\Desktop`). For fully-expanded paths, `REG_SZ` (via `WriteString`) is the correct type. While functionally Windows may tolerate this, it is semantically wrong and inconsistent with how line 198 correctly uses `WriteString` for the same fully-expanded path in the Shell Folders key. Some Windows components or third-party tools may misbehave when finding a non-expandable string in an `REG_EXPAND_SZ` value.

**Fix:** Use `sys.WriteString(sys.UserShellFoldersKey, regValueName, regValue)` (REG_SZ) to match the Shell Folders write at line 198, or construct a path with `%USERPROFILE%` replacement if `REG_EXPAND_SZ` is truly needed.

### WR-03: `Process32FirstW`/`Process32NextW` errors silently swallowed in security software detection

**File:** `internal/syscall/process.go:85-87, 95-98`
**Issue:** When `Process32FirstW` or `Process32NextW` returns 0, the code returns `false` ("process not running") without checking `GetLastError`. If the error is something other than `ERROR_NO_MORE_FILES` (e.g., `ERROR_ACCESS_DENIED` or `ERROR_INVALID_HANDLE`), the failure is silently interpreted as "target process not found". This means security software detection (`IsProcessRunning`) can produce false negatives when toolhelp32 enumeration fails for permission or handle-related reasons.

**Fix:** Check the last error after `r1 == 0` and at minimum distinguish `ERROR_NO_MORE_FILES` (expected end of list) from other errors. Unexpected errors should at least be logged.

### WR-04: `MkdirAll` uses Unix permission `0755` on Windows

**File:** `internal/migrate/migrate.go:164, 249`
**Issue:** `os.MkdirAll(sourcePath, 0755)` and `os.MkdirAll(backupDir, 0755)` use Unix-style octal permissions. Go on Windows maps these to Windows security descriptors, but `0755` may not produce the expected default ACL (full control for the creating user). Using `0777` (`os.ModePerm`) is the conventional choice for Windows directories where no specific permission restriction is intended.

**Fix:** Replace `0755` with `0777` (or `os.ModePerm`) in both `MkdirAll` calls at lines 164 and 249.

## Info

### IN-01: `IsSSDDrive` always returns `false` -- dead code

**File:** `internal/syscall/volume.go:130-155`
**Issue:** The `IsSSDDrive` function opens a file handle to the drive, acquires `CreateFile` handle, and then unconditionally returns `false` with the comment "Conservative default - can't reliably detect without admin". The function is unused by any caller (confirmed by grep across the codebase). This is dead code that takes 25 lines including a file handle open/close. Already documented as a known anti-pattern in `CONCERNS.md` and `ARCHITECTURE.md`.

**Fix:** Either implement the `IOCTL_STORAGE_QUERY_PROPERTY` call as the TODO comment suggests (it does not require admin), or remove the function entirely.

### IN-02: `DetectConflicts` return error intentionally swallowed

**File:** `internal/migrate/migrate.go:102`
**Issue:** `conflicts, _ := DetectConflicts(actualSource, targetPath)` intentionally ignores the error with an inline comment explaining it's MVP behavior. While intentional, this means that if `DetectConflicts` fails due to a real error (permission issue, invalid path, etc.), the migration proceeds silently without conflict detection. The difference between "0 conflicts found" and "detection failed" is lost.

**Fix:** Log the error when it occurs so that post-mortem analysis can distinguish "no conflicts" from "could not check for conflicts".

### IN-03: Dead import workaround in test file

**File:** `internal/syscall/process_test.go:4, 33`
**Issue:** The file imports `"io"` and uses `_ = io.Discard` at line 33 to avoid an "unused import" error. However, `.golangci.yml` lines 41-42 already exclude test files from the `unused` linter, making this workaround unnecessary. This is dead code that masks the linter exclusion's overly broad scope.

**Fix:** Remove the `"io"` import and the `_ = io.Discard` line. Consider narrowing the `.golangci.yml` test-file exclusion to only skip genuinely unused test helpers rather than all `unused` checks.

### IN-04: `.golangci.yml` excludes entire `internal/syscall/` from formatters

**File:** `.golangci.yml:61-63`
**Issue:** The `formatters.exclusions.paths` block excludes the entire `internal/syscall/` package (and `internal/folders/`) from `gofmt` and `goimports` checks. This is overly broad -- formatting issues in non-test syscall files (e.g., `volume.go` uses 4-space indentation at line 131 while the rest of the codebase uses tabs) go undetected. The only justifiable exclusion would be the `//go:build windows` constraint files if they cause the tool to fail, but that is not the case here.

**Fix:** Remove `internal/syscall/.*` from formatter exclusions, or narrow the exclusion to specific files that need it.

### IN-05: README.md component/method counts may be inaccurate

**File:** `README.md:67, 75`
**Issue:** Line 67 states "16 个 Wails 绑定方法" but the codebase and `CLAUDE.md` reference 17 exported methods on the `App` struct. Line 75 states "12 个 UI 组件" but the CLAUDE.md references 11 components. These counts drift over time and become misleading.

**Fix:** Update the counts to reflect the current codebase, or replace hardcoded numbers with descriptions (e.g., "all Wails-bound methods" instead of "17 个 Wails 绑定方法").

---

_Reviewed: 2026-06-22T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
