# Testing Patterns

**Analysis Date:** 2026-06-21

## Overview

**Testing exists only on the Go side.** There are zero frontend tests (no `*.test.tsx` or `*.spec.tsx` files). The Go test suite has 14 test functions across 8 `*_test.go` files using nothing but Go's standard `testing` package. There is no CI/CD configuration (no `.github/` directory, no CI config files).

---

## Test Framework

**Runner:**
- Go standard `testing` package — **no external test framework** (no testify, no ginkgo, no goconvey)
- Go 1.23.5 (from `go.mod`)

**Run Commands:**
```bash
go test ./... -v              # Run all tests with verbose output
go test ./internal/migrate/   # Run migrate package tests only
go test ./internal/syscall/   # Run syscall package tests only
go test -run TestCreateJunction ./internal/syscall/  # Run a single test
go test -cover ./...          # Coverage (no configured threshold)
```

**Config File:** None. Go tests use the default `testing` package with no additional configuration.

---

## Test File Organization

**Location:**
- Co-located: `*_test.go` files live in the same package directory as the source code
- Package naming: tests use the same package name as source (e.g., `package migrate`, `package winsyscall`) — white-box testing

**Test file distribution:**
```
internal/migrate/
  backup_test.go       — 3 test functions (backup save/load/delete)
  check_test.go        — 2 test functions (PreCheck smoke, maskPath)
  conflict_test.go     — 3 test functions (conflict detection)
  move_test.go         — 2 test functions (cancellation)
  onedrive_test.go     — 1 test function (OneDrive KFM detection)
internal/syscall/
  fileops_test.go      — 2 test functions (ToLongPath table, IsCrossVolume table)
  fileops_acl_test.go  — 1 test function (CopyFileWithACL)
  junction_test.go     — 1 test function (CreateJunction/IsJunction/DeleteJunction roundtrip)
  junction_debug_test.go — 3 test functions (mklink compatibility, reparse buffer debug)
  power_test.go        — 2 test functions (power status, battery check)
  registry_test.go     — 1 test function (User Shell Folders read)
  volume_test.go       — 3 test functions (volume info, free space, drive type)
```

**Naming:**
- Test files: `{source_file}_test.go` (e.g., `junction_test.go` maps to `junction.go`)
- Some tests have dedicated files: `junction_debug_test.go`, `fileops_acl_test.go`
- Test functions: `Test{PascalCase}` (`TestCreateJunction`, `TestDetectConflicts_NoConflict`)

---

## Test Structure

### Standard Test Pattern

Tests use a flat structure with no subtests (`t.Run` is not used):

```go
func TestCreateJunction(t *testing.T) {
    tmpDir := t.TempDir()

    // 1. Setup — create test directories and files
    target := filepath.Join(tmpDir, "target")
    if err := os.Mkdir(target, 0o755); err != nil {
        t.Fatalf("Mkdir target: %v", err)
    }

    // 2. Exercise — call the function under test
    if err := CreateJunction(link, target); err != nil {
        t.Fatalf("CreateJunction: %v", err)
    }

    // 3. Verify — check results
    if !IsJunction(link) {
        t.Fatal("IsJunction returned false after CreateJunction")
    }

    // 4. Teardown (cleanup is automatic via t.TempDir)
}
```

### Table-Driven Test Pattern

Used for pure-logic functions with multiple input/output pairs (`fileops_test.go`):

```go
func TestIsCrossVolume(t *testing.T) {
    tests := []struct {
        src, tgt string
        want     bool
    }{
        {`C:\src`, `C:\dst`, false},
        {`C:\src`, `D:\dst`, true},
        {`C:\src`, `c:\dst`, false},  // case-insensitive
        {`D:\src`, `E:\dst`, true},
        {`relative`, `C:\dst`, false},
    }
    for _, tt := range tests {
        got := IsCrossVolume(tt.src, tt.tgt)
        if got != tt.want {
            t.Errorf("IsCrossVolume(%q, %q) = %v, want %v", tt.src, tt.tgt, got, tt.want)
        }
    }
}
```

### Smoke Test Pattern (Environment-Dependent)

For tests that depend on system state (registry, drives, OneDrive installation), the convention is to gracefully handle missing environments:

```go
func TestGetVolumeInfo(t *testing.T) {
    info, err := GetVolumeInfo("C")
    if err != nil {
        t.Logf("GetVolumeInfo(C) returned error (may not have C: drive): %v", err)
        return  // Graceful skip — not a test failure
    }
    // Only assert on available environments
    if info.FileSystem == "" {
        t.Error("FileSystem should not be empty")
    }
}
```

This is **not** `t.Skip()` — tests use `t.Logf` + `return` instead.

### No-Panic Smoke Test Pattern

For functions that interact with the OS and may fail in unpredictable ways:

```go
func TestPreCheck_NonNTFS(t *testing.T) {
    tmpDir := t.TempDir()
    _ = PreCheck(tmpDir, "Z")  // Deliberately ignore return value
    // The goal is to verify no panics.
    t.Log("PreCheck completed without panic")
}
```

---

## Mocking and Isolation

**No mocking framework is used.** Tests achieve isolation through:

1. **Temp directories:** `t.TempDir()` for file system isolation (used in every test that touches disk)
2. **Environment variable manipulation:** `t.Setenv("APPDATA", tmpDir)` to redirect persistent storage (`backup_test.go`)
3. **Non-existent paths:** Passing invalid drive letters (e.g., `"Z"`) to exercise error paths without side effects

**Windows API calls are NOT mocked.** Tests that call `CreateJunction`, `CopyFileWithACL`, or registry functions execute against the real Windows API. This means:
- Junction tests require NTFS (only runs on Windows)
- Registry tests gracefully handle missing keys
- Volume tests gracefully handle missing drives

**Global state management in tests:**
- `move_test.go` resets the `cancelled` atomic flag between tests via `ResetCancellation()`
- No other shared mutable state across tests

---

## Assertion Style

**No assertion library.** Tests use only standard `testing.T` methods:

| Method | Usage |
|--------|-------|
| `t.Fatalf(...)` | When test cannot continue (unrecoverable setup error) |
| `t.Fatal(...)` | Same, without formatting |
| `t.Errorf(...)` | When assertion fails but test should continue to check more conditions |
| `t.Error(...)` | Same, without formatting |
| `t.Logf(...)` | Informational output (used freely — tests are verbose by default) |
| `t.Log(...)` | Same, without formatting |

**Error message convention:**
```go
// Format: "function(param) = got, want expected" or "description: value"
t.Errorf("NewPath = %q, want %q", loaded.NewPath, `D:\Desktop`)
t.Errorf("expected 2 conflicts, got %d", len(conflicts))
t.Fatalf("ReadFile through junction: %v", err)
```

---

## Test Types

### Unit Tests (pure logic)

- `TestToLongPath` — table-driven, no external dependencies
- `TestIsCrossVolume` — table-driven, string comparison only
- `TestCancelMigration` — in-process flag manipulation
- `TestCancelMigration_Concurrent` — goroutine with channel synchronization

### Integration-Lite Tests (real FS + Win32 API)

These test against real Windows filesystem and API calls within `t.TempDir()`:
- `TestCreateJunction` — creates actual NTFS junction, verifies with `IsJunction`, reads through junction, deletes it
- `TestCopyFileWithACL` — copies a real file with ACL preservation
- `TestDetectConflicts` / `TestDetectConflicts_NoConflict` / `TestDetectConflicts_AllConflict` — create directory trees and run conflict detection
- `TestSaveLoadDelete` — full CRUD cycle on backup JSON files
- `TestJunctionWithMklink` — creates junction via `cmd /c mklink /J` then verifies and deletes programmatically

### Environment-Dependent Smoke Tests

These may pass or gracefully degrade depending on what's installed on the test machine:
- `TestGetVolumeInfo`, `TestGetDiskFreeSpace`, `TestGetDriveType` — require a `C:` drive
- `TestReadUserShellFolders` — requires standard Windows user registry keys
- `TestGetPowerStatus`, `TestIsOnBattery` — depends on laptop/desktop power state
- `TestDetectOneDriveKFM` — depends on OneDrive installation
- `TestPreCheck_NonNTFS` — exercises error paths with a non-existent `Z:` drive

### Debug/Exploratory Tests

`junction_debug_test.go` contains 3 tests that were used to debug junction buffer construction. These are verbose, log hex dumps, and call into unexported package internals (`openReparse`, `buildReparseBuffer`). They serve as documentation for the reparse point binary format.

### Runtime Verification Tool

`cmd/verify/main.go` is a standalone program (not a test) that performs 15 runtime checks: junction roundtrip, registry reads, disk enumeration, OneDrive detection, cloud sync detection, backup load, logger init, cross-volume logic, path conversion, battery status. This is used for manual verification on target Windows machines.

---

## What is NOT Tested

| Gap | Impact |
|-----|--------|
| **Frontend (React/TypeScript)** | Zero tests. All UI behavior — state machine, progress simulation, dialog flows, error handling — is untested. |
| **app.go Wails binding methods** | No test for `GetFolders()`, `GetDisks()`, `StartMigration()`, `StartupIntegrityCheck()`, `Rollback()`. These are the bridge between frontend and backend. |
| **Full migration flow** | No end-to-end test that runs `Migrate()` against a real folder on a real target drive with junction creation. |
| **Rollback flow** | `Rollback()` in `internal/migrate/rollback.go` has no test. |
| **Registry write path** | `WriteExpandString` and `WriteString` are not tested (test reads registry but never writes). |
| **Concurrent cancellation** | Only tested via flag manipulation, not with an actual in-progress file move being cancelled. |
| **Error paths for most MoveKitError codes** | Only a few error paths are exercised (drive not found, target not found). Most error codes (`ErrCodeOneDriveKFM`, `ErrCodeFileLocked`, `ErrCodeControlledFolder`, etc.) have no test coverage. |
| **Cross-platform** | All tests assume Windows + NTFS. No Linux/macOS compatibility tests. |
| **CI pipeline** | No CI/CD configuration exists. Tests run only manually on developer machines. |

---

## Fixtures and Test Data

**No fixture files.** All test data is created inline within test functions using `os.WriteFile` and `os.MkdirAll`. This keeps tests self-contained but means no reuse across test files.

**Example setup from `conflict_test.go`:**
```go
func TestDetectConflicts(t *testing.T) {
    tmpDir := t.TempDir()
    srcDir := filepath.Join(tmpDir, "src")
    dstDir := filepath.Join(tmpDir, "dst")

    os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
    os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("hello"), 0644)
    os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("world"), 0644)
    os.WriteFile(filepath.Join(srcDir, "unique.txt"), []byte("only in src"), 0644)

    // Same for dstDir with overlapping files...
}
```

---

## Coverage

**Requirements:** No coverage threshold enforced.

**To view coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

There is no coverage configuration in `go.mod`, no `-covermode` settings, and no CI gate.

---

*Testing analysis: 2026-06-21*
