# Pitfalls Research: MoveKit Testing

**Domain:** Windows desktop application testing (filesystem-heavy, NTFS junctions, registry, Win32 syscalls)
**Researched:** 2026-06-21
**Confidence:** HIGH

---

## Critical Pitfalls

### Pitfall 1: Junction Test Leakage via t.TempDir() + CreateJunction → RemoveAll Failure

**What goes wrong:**
Tests that create real NTFS junctions inside `t.TempDir()` cause Go's built-in cleanup (`os.RemoveAll`) to fail silently or with "access is denied" on Windows. The junction is treated as a reparse point, and `os.RemoveAll` can traverse into the junction target (another temp dir or non-existent path) or fail to delete the junction directory because the underlying handle isn't released before cleanup runs. This leaves orphaned directories in `%TEMP%` and, worse, can cause subsequent tests to fail because the junction target directory was deleted but the junction link still exists pointing at stale paths.

**Why it happens:**
- NTFS junctions differ from POSIX symlinks: `os.RemoveAll` on Windows does NOT automatically handle junction reparse points specially.
- `t.TempDir()` cleanup runs `os.RemoveAll` after the test returns. If the junction's target is inside the same temp tree, the traversal order matters — the target may be deleted before the junction, or the junction may point to an already-deleted path, causing errors.
- Go's testing package retry logic (issue #51442) handles transient sharing violations but does NOT handle junction/reparse-point-specific deletion failures.
- The existing `junction_test.go` and `junction_debug_test.go` create real junctions inside `t.TempDir()` — these currently pass only because temp directories are on the same volume and the target is still alive during cleanup in simple cases. When targets are on different volumes (e.g., C: temp dir, D: target), cleanup will reliably fail.

**How to avoid:**
1. **Always call `DeleteJunction()` before the test returns.** Add `t.Cleanup(func() { DeleteJunction(link); os.Remove(link) })` immediately after creating each junction.
2. **Keep junction targets inside the same `t.TempDir()` tree** so the cleanup order (target before link) is deterministic and both are removed by `t.TempDir()`.
3. **Do NOT create junctions pointing outside the temp directory tree.** Tests that create junctions pointing to real paths (e.g., `C:\Users\...`) are destructive and must be avoided entirely.
4. **Use a helper:** Create a `testCreateJunction(t, link, target)` wrapper that registers the cleanup automatically.

**Warning signs:**
- `go test -count=1` passes but `go test -count=3` (repeated runs) fails with "The process cannot access the file"
- Tests leave behind directories named like `TestCreateJunction*` in `%TEMP%`
- CI on Windows runners shows intermittent `TempDir RemoveAll cleanup` errors

**Phase to address:**
TEST-01 and TEST-02 (migrate and syscall package testing) — must bake junction cleanup discipline into every test that calls `CreateJunction` or `IsJunction`.

---

### Pitfall 2: Registry Pollution — Writing to Real HKCU Keys During Tests

**What goes wrong:**
The `registry.go` file has functions that write directly to real HKCU keys (`User Shell Folders`, `Shell Folders`). The existing `registry_test.go` already touches the real `HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders\Desktop` key just to read it. When tests for `WriteExpandString` and `WriteString` are added, they will write to real registry paths that control Windows Explorer's folder redirection. A test that writes `D:\Desktop` to `HKCU\...\Desktop` and crashes before cleanup will permanently break the user's Desktop folder configuration.

**Why it happens:**
- `registry.go` hardcodes registry key paths as string constants and uses `registry.OpenKey(registry.CURRENT_USER, ...)` directly — no abstraction layer.
- The existing `ReadExpandString` test reads the real Desktop key value "because it exists on most machines" — this pattern normalizes touching real registry state in tests.
- Go's `golang.org/x/sys/windows/registry` package has no built-in mocking capability.
- Chromium's `RegistryOverrideManager` pattern (overriding registry hives with temp keys) is well-known in C++ but has no Go equivalent library.

**How to avoid:**
1. **Introduce a `RegistryAccessor` interface** that abstracts `ReadExpandString`, `WriteExpandString`, `WriteString`, `ReadInteger`, `ReadIntegerLM` behind an interface. Production code uses the real implementation; tests use an in-memory fake.
2. **The in-memory fake** should store values as `map[string]map[string]interface{}` (keyPath → valueName → value) and validate types.
3. **If integration tests MUST touch the real registry**, use a dedicated test key path (e.g., `HKCU\Software\MoveKit\Test\...`) and never write to the real Explorer keys. Register `t.Cleanup` to delete the test key.
4. **Add a `//go:build integration` tag** to any test that touches real registry so it never runs in `go test ./...` by default.

**Warning signs:**
- After running tests, the user's Desktop folder appears broken in Explorer
- `regedit` shows unexpected values under `User Shell Folders` after `go test`
- CI fails because registry keys that exist on dev machines don't exist in CI runners

**Phase to address:**
Phase 1 (syscall testing) — MUST introduce the `RegistryAccessor` interface before adding write tests. This is a dependency for all migrate package tests that call `rollback.go` (which writes to registry during rollback).

---

### Pitfall 3: Mocking Win32 Syscalls at the Wrong Layer

**What goes wrong:**
MoveKit uses `syscall.NewLazyDLL` and raw `Call()` for six Win32 APIs: `CreateFileW`, `DeviceIoControl`, `CreateMutexW`, `CreateToolhelp32Snapshot`, `Process32FirstW/NextW`, `GetSystemPowerStatus`, `GetVolumeInformationW`, `GetDriveTypeW`, `GetDiskFreeSpaceExW`, `SHChangeNotify`. Attempting to mock these at the syscall layer (e.g., by swapping DLL handles or using monkey-patching libraries like `go-mpatch`) will produce tests that pass but test nothing real — the mock replaces the actual Win32 behavior with assumptions that may not hold.

Additionally, a `go-mpatch` approach (monkey-patching function pointers) is fundamentally incompatible with Go 1.23's compiler optimizations, breaks under `-race`, and doesn't survive across package boundaries reliably.

**Why it happens:**
- Syscalls are the lowest layer — mocking here means you are testing your mock, not the actual Windows API behavior.
- The junction code's correctness depends on binary-identical reparse buffer layout matching what `mklink /J` produces — syscall mocking cannot verify this.
- `DeviceIoControl` is a catch-all that behaves differently depending on control codes (SET/GET/DELETE reparse point) — a mock that returns "success" for all control codes hides real bugs.

**How to avoid:**
1. **Mock at the business-logic layer, not the syscall layer.** Test `junction.go` functions with real junctions in `t.TempDir()` (they're fast — ~0.5ms per junction creation). They've already been validated against `mklink /J` output in `junction_debug_test.go`.
2. **For functions that CANNOT use real junctions safely** (e.g., `IsJunction` failure paths, `GetVolumeInfo` errors), extract a thin interface. Example:
   ```go
   type VolumeChecker interface {
       GetVolumeInfo(drive string) (VolumeInfo, error)
       GetDiskFreeSpace(drive string) (int64, int64, int64, error)
   }
   ```
   The real implementation is a struct with no fields; the fake returns controlled errors.
3. **For `IsProcessRunning` and `EnsureSingleInstance`:** These CAN use real Windows APIs in tests — run actual test executables or use the fact that the test process itself is a real process.
4. **For `SHChangeNotify` and `GetSystemPowerStatus`:** These are fire-and-forget or query-only — test the wrapper's input/output mapping, not the syscall itself. Use interface abstraction for `PowerStatus` queries to simulate battery vs AC.

**Warning signs:**
- Tests pass but the real junction creation fails on a different Windows version
- Mock expects `DeviceIoControl` returns error code 5 but Windows 11 24H2 returns error code 87 for the same failure
- `go test -race` fails after adding syscall mocks

**Phase to address:**
Phase 1 (syscall testing) — decide which Win32 functions get interfaces vs real testing. Volume/disk functions need interfaces. Junction functions should be tested with real NTFS. Process/power functions need interfaces for error-path testing.

---

### Pitfall 4: Shared `cancelled atomic.Int32` — Test Isolation Breaks

**What goes wrong:**
`move.go` uses a package-level `var cancelled atomic.Int32` shared across all tests. If TestA sets the cancellation flag and TestB runs concurrently (or in sequence without reset), TestB gets a leftover true state. The existing `move_test.go` handles this with explicit `ResetCancellation()` calls, but future tests (especially parallel tests with `t.Parallel()`) will inherit stale state.

More critically: the cancellation flag controls the `moveFilesCrossVolume` goroutine. When a test calls `CancelMigration()` then exits, the goroutine may still be running and checking `IsCancelled()` on the shared flag. The next test that calls `ResetCancellation()` can mask the running goroutine's state.

**Why it happens:**
- `cancelled` is a package-level `atomic.Int32` (global mutable state).
- `moveFilesCrossVolume` uses `ResetCancellation()` internally at the start — so tests that call this function reset each other's cancellation flags.
- The existing `TestCancelMigration_Concurrent` test creates a goroutine that busy-waits on `IsCancelled()`. If tests run with `-parallel`, multiple busy-wait goroutines all read/write the same flag.
- Go test parallelization is per-package — all tests in `migrate` package share this flag.

**How to avoid:**
1. **Replace the package-level `atomic.Int32` with a per-operation context.** Make `moveFilesCrossVolume` accept a `context.Context` and check `ctx.Err() != nil` instead of `IsCancelled()`.
   ```go
   func moveFilesCrossVolume(ctx context.Context, src, dst string, ...) error {
       select {
       case <-ctx.Done():
           return ctx.Err()
       default:
       }
       // ... rest of logic, periodically checking ctx.Done()
   }
   ```
2. **For the `CancelMigration()` API (called from Wails frontend via `app.go`),** the `App` struct should hold its own `context.CancelFunc` that it calls:
   ```go
   type App struct {
       ctx          context.Context
       migrateCtx   context.Context
       cancelMigrate context.CancelFunc
       mu           sync.Mutex
   }
   ```
3. **Until refactoring to contexts is complete, use a per-test sentinel:** In tests, wrap the global flag with a test-scoped struct that provides isolation.

**Warning signs:**
- Tests pass individually (`go test -run TestX`) but fail when run together (`go test ./internal/migrate/...`)
- `go test -parallel 4` introduces flaky cancellations
- Race detector reports data races on the cancellation flag

**Phase to address:**
Phase 1 (migrate testing) — if testing `moveFilesCrossVolume` cancellation paths, MUST refactor to context-based cancellation first. Otherwise tests will be flaky by construction.

---

### Pitfall 5: Testing Rollback Without Registry Isolation Wrecks Developer Machines

**What goes wrong:**
`rollback.go` performs operations that are destructive when tested naively:
1. Calls `sys.IsJunction(rec.OriginalPath)` on a real path from a backup record
2. Calls `sys.DeleteJunction(rec.OriginalPath)` — deletes a real junction
3. Calls `os.Remove(rec.OriginalPath)` — deletes a real directory
4. Calls `os.Rename(rec.NewPath, rec.OriginalPath)` — moves real files
5. Calls `sys.WriteExpandString(sys.UserShellFoldersKey, ...)` — writes to real registry
6. Calls `sys.WriteString(sys.ShellFoldersKey, ...)` — writes to real registry

A test that calls `Rollback("Desktop")` will attempt to delete the developer's real Desktop junction, move files, and rewrite registry keys. This is catastrophic.

**Why it happens:**
- `Rollback` reads backup records from `%APPDATA%\MoveKit\backups\` — a real filesystem path.
- The existing `backup_test.go` already overrides `APPDATA` with `t.Setenv` — this is good but insufficient because `Rollback` also calls `sys.WriteExpandString` which writes to the real HKCU registry.
- Developers may write a test that calls `Rollback("Desktop")` with a mock backup record and doesn't realize the side effects until it's too late.

**How to avoid:**
1. **Interface-ify `Rollback`'s external dependencies.** It needs interfaces for:
   - Backup storage (`SaveBackup`/`LoadBackup`)
   - Registry writing (`WriteExpandString`/`WriteString`)
   - Filesystem operations (junction ops, rename, remove)
   - Junction verification (`IsJunction`)
2. **In tests, inject a full fake registry + fake filesystem.** The test creates a temp directory tree that looks like the original state and verifies that `Rollback` restores it correctly without touching any real paths.
3. **Add a safety guard:** In the real `Rollback` implementation, verify that `os.Getenv("MOVEKIT_TEST_MODE")` is not set before writing to registry. Tests set this env var, and the real function panics/runs in dry-run mode when it's present.

**Warning signs:**
- After running `go test`, Windows Explorer shows folders as empty or missing
- Registry keys under `User Shell Folders` have been changed to test paths
- `go test` requires elevation to run (because registry writes failed without admin)

**Phase to address:**
Phase 2 or later — `rollback_test.go` MUST NOT be written until the interface abstraction for registry and junction operations is complete. This is blocked on Pitfall 2 (RegistryAccessor) and Pitfall 6 (filesystem interface).

---

### Pitfall 6: Using afero (or any VFS Library) Without Understanding Its Windows NTFS Gaps

**What goes wrong:**
Developers reach for `spf13/afero` or similar VFS libraries to mock the filesystem in tests. afero's `MemMapFs` does not support:
- NTFS junction points / reparse points
- Windows extended attributes (ACL preservation — `CopyFileWithACL`)
- NTFS alternate data streams
- Case-insensitive path resolution (Windows filesystem is case-insensitive, `MemMapFs` is case-sensitive)
- Windows drive letters and volume mounting

Code that passes tests with afero but fails in production on real NTFS is a common pattern in Go Windows tooling projects.

**Why it happens:**
- afero was designed primarily for Linux server workloads (file processing, config management) where NTFS features don't exist.
- The documentation doesn't prominently warn about Windows-specific gaps.
- The temptation to "just mock the filesystem" is strong when testing code that does `filepath.WalkDir`, `os.Rename`, `os.RemoveAll`.

**How to avoid:**
1. **Do NOT use afero or any generic VFS library for MoveKit testing.** The app's entire value proposition depends on NTFS-specific behavior (junctions, ACL preservation via `CopyFileWithACL`). A generic VFS cannot model this.
2. **Instead, define minimal interfaces specific to MoveKit's needs:**
   ```go
   type FileMover interface {
       MoveSameVolume(src, dst string) error
       MoveCrossVolume(src, dst string, progress ProgressCallback) error
   }
   type JunctionOps interface {
       CreateJunction(link, target string) error
       IsJunction(path string) bool
       DeleteJunction(path string) error
   }
   type FileCopier interface {
       CopyFileWithACL(src, dst string, progress func(uint64)) error
   }
   ```
   These are thin interfaces that the real implementations satisfy. Tests use `t.TempDir()` + real NTFS for the filesystem backend but controlled inputs for business logic paths.
3. **For tests that exercise the actual file copy/move logic,** use `t.TempDir()` directories (which are always on NTFS on modern Windows). This gives real NTFS behavior without mocking.

**Warning signs:**
- Tests pass on Linux CI but fail on Windows
- Code uses `afero.Fs` interface parameters but the real implementation works around afero's Windows gaps
- Path comparison tests pass because test paths use lowercase but production paths are mixed-case

**Phase to address:**
Phase 1 (architecture decision) — decide on interface strategy before any new tests are written. This is a one-way door: once interfaces are chosen, all tests depend on them.

---

### Pitfall 7: Privilege Escalation Through File Permissions in Tests

**What goes wrong:**
MoveKit's `fileops.go` includes `CopyFileWithACL` which calls Win32 security APIs to preserve file permissions during copy. The existing `fileops_acl_test.go` may need to read or write files with specific permissions. Tests that create files with restrictive permissions (e.g., read-only, SYSTEM-only) inside `t.TempDir()` can cause `t.TempDir()` cleanup to fail because the test process doesn't have delete permission on the files it just created.

Additionally: the `security.go` file in migrate package detects running security software (360, Huorong, etc.) by process name. Tests that verify this detection should NOT attempt to launch these processes — they should provide a fake process list.

**Why it happens:**
- `CopyFileWithACL` may set permissions that are more restrictive than the default temp directory permissions.
- Windows permission inheritance means child items can inherit restrictive ACLs from parents created by the test.
- `os.Chmod` on Windows sets the read-only attribute but tests may check `os.FileMode` bits, which behave differently across platforms (see Go issue #21322).

**How to avoid:**
1. **After creating files with custom ACLs in tests, restore permissions before cleanup:**
   ```go
   t.Cleanup(func() {
       os.Chmod(path, 0666)
   })
   ```
2. **Test `CopyFileWithACL` correctness by comparing final file content and basic attributes, not exact ACL structures** (which vary by Windows version and filesystem).
3. **For `DetectCloudSync` / `IsProcessRunning` tests:** Inject a process-listing interface so tests can provide predefined process lists without calling real `CreateToolhelp32Snapshot`.
4. **Never run tests as Administrator.** A test that creates files owned by SYSTEM or another user becomes undeletable by the non-admin test process. If tests require admin rights, they should `t.Skip()` when not elevated.

**Warning signs:**
- Tests fail to clean up temp directories after running `CopyFileWithACL` tests
- `os.RemoveAll` returns "access is denied" on test temp dirs
- Tests pass as Admin but fail as normal user

**Phase to address:**
Phase 1 (syscall testing) — the first time `CopyFileWithACL` is tested, the cleanup restoration pattern must be established.

---

### Pitfall 8: Go Error-Wrapping Anti-Patterns Masking Test Failures

**What goes wrong:**
The codebase has several patterns where errors are silently swallowed or inadequately tested:

1. **BUG-03 (backup save errors swallowed):** `backup.go` likely calls `os.WriteFile` without checking the error. The existing test verifies save/load but if `WriteFile` silently fails, the test still passes.
2. **BUG-05 (dirSize errors ignored):** `app.go:GetFolders()` calls `dirSize(expanded)` and uses `if err == nil` as a gate — progress is silently inaccurate when dirSize errors. A test that only checks FolderInfo.Status == "migrated" won't catch this.
3. **BUG-06 (UTF16 encoding errors ignored):** `shell.go:NotifyFolderRenamed` calls `syscall.UTF16PtrFromString` and discards the error. A test verifies the shell notification was sent but cannot know if the path was garbled.
4. **BUG-08 (security software check results discarded):** Some check result is computed but never used. A test for the check function passes but the integration is broken.

When writing tests for these buggy areas, the common pitfall is to test the happy path (which happens to work) without testing the error path that the bug hides.

**Why it happens:**
- Go developers often test "does it return nil error" without testing "does it return the RIGHT error."
- When the production code swallows errors, a passing test gives false confidence — the test passes because the error was dropped, not because the operation succeeded.
- BUG-03, BUG-05, BUG-06, BUG-08 should each be FIXED before writing comprehensive tests for the affected functions, otherwise the tests validate buggy behavior.

**How to avoid:**
1. **For each BUG-XX in the Active list:** fix the bug FIRST, then write the test that proves the fix works. "Write test → it should fail → fix bug → test passes" is the correct TDD flow.
2. **For each function that takes an `onProgress` callback:** write a test that provides a callback which accumulates values and verifies the callback was called the correct number of times with correct values (catches BUG-05 where dirSize affects progress).
3. **For each function that writes to a file or registry:** the test MUST verify the written value, not just the absence of an error.
4. **Use `errors.Is` for assertions, not string comparison.** The codebase uses custom error types (`MoveKitError`) — tests should check `var mkErr *MoveKitError; errors.As(err, &mkErr)` and verify `.Code`.

**Warning signs:**
- Test assertion: `if err != nil { t.Fatal(err) }` — no check on WHAT the error is
- Functions that return `error` but the test never calls them in a way that could produce errors
- "Test passed" after fixing a bug — but the test would have passed BEFORE the fix too

**Phase to address:**
Phase 0 (bug fixes) — fix bugs before Phase 1 (test writing). Tests that validate buggy behavior become technical debt themselves.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Using `t.Setenv("APPDATA", tmpDir)` to isolate backup tests (as in `backup_test.go`) | No interface refactoring needed | Only works for backup; doesn't isolate registry writes, junction ops, or real path usage | NOW — already used, low risk when confined to backup tests only |
| Writing tests that read real HKCU keys "because they exist" (as in `registry_test.go`) | Quick test coverage | Pollutes dev registry if expanded to writes; breaks in CI without standard Windows config | NEVER for writes. OK for read-only tests with `t.Logf` (graceful skip on missing keys) |
| Skipping `t.Helper()` in test utility functions | Less typing | Stack traces point inside the helper, not the test case that failed | NEVER — always use `t.Helper()` in shared test helpers |
| Testing `MoveKitError` code paths with string matching (`err.Error() == "..."`) | Easy to write | Breaks when error messages are improved/translated | NEVER — use `errors.As` with `*MoveKitError` and check `.Code` |
| Using `atomic.Int32` package-level for cancellation instead of `context.Context` | Simpler API, no context plumbing | Test isolation impossible; concurrent tests interfere | Only for MVP with single-test execution. MUST refactor for parallel tests |
| Calling `procDeviceIoCtrl.Call(...)` directly in tests (as in `junction_debug_test.go`) | Verifies raw Win32 behavior | Tests are fragile to Windows version changes; doesn't use the public API | Acceptable for debug/regression tests marked `//go:build ignore` or as investigation-only |

---

## Integration Gotchas

Common mistakes when connecting test infrastructure to MoveKit's OS integrations.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| NTFS Junction API | Creating junctions then relying on `t.TempDir()` cleanup to delete them | Explicit `DeleteJunction()` + `os.Remove()` in `t.Cleanup` before `t.TempDir()` runs its cleanup |
| Windows Registry | Writing to real HKCU Explorer keys during tests | Interface-ify registry ops; use in-memory fake for unit tests; use dedicated `HKCU\Software\MoveKit\Test\` for integration tests |
| Win32 Process Snapshot | Calling real `CreateToolhelp32Snapshot` to test `IsProcessRunning` | Interface-ify process listing; test real API under integration build tag |
| OneDrive KFM Detection | Assuming OneDrive is or isn't installed in test environment | Interface-ify OneDrive status; use fake with controlled KFM state |
| Cloud Sync Provider Detection | Hardcoding process name list that changes with software updates | Test against a provider registry interface; the test provides process names, not the production code's hardcoded list |
| Wails Runtime Events | Calling `runtime.EventsEmit` in tests (needs Wails context) | Extract event emission behind an `EventEmitter` interface; test that the correct event payload is constructed |

---

## Performance Traps

Patterns that work at small scale but fail as test count grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Creating real NTFS junctions per test case | Test suite slows linearly; each junction is ~0.5ms, but 200 tests = 100ms overhead, and `t.TempDir()` creation adds ~1-2ms each | Group junction-dependent tests; use `t.TempDir()` once per subtest, not per assertion | ~50 junction-creating tests |
| Running `dirSize()` on real folders during test | Test reads entire directory tree; on a developer machine with large Desktop, single test takes seconds | Inject a `DirSizer` interface that tests override with fixed values | First time a test runs on a machine with >1GB Desktop folder |
| `filepath.WalkDir` in test assertions on full temp dirs | WalkDir traverses every file in the temp tree — if a test creates 10k files, verification walk is expensive | Check specific files after operations, not full tree diffs | Test creates >1000 files |
| `exec.Command("cmd", "/c", "mklink", "/J", ...)` in tests | Spawning cmd.exe per test is ~50ms; slower than direct `CreateJunction` syscall (~0.5ms) | Use `CreateJunction` for setup, reserve `mklink` only for cross-validation tests | ~20 mklink-based tests |

---

## Security Mistakes

Domain-specific security issues in testing.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Tests create files with `os.Chmod(0777)` — on Windows, this opens files to Everyone:Write | Temp files become world-writable | Use `0644` for temp files; never set 0777 without explicit need |
| Tests that call `os.MkdirAll` then `os.Chmod` leave race windows | Time-of-check-to-time-of-use in test security | Use `os.Mkdir(path, perm)` directly when perm matters |
| Hardcoding test paths like `C:\Users\test\Desktop` | Test uses a real user profile path; on some machines this could be a real user's Desktop | Use `filepath.Join(t.TempDir(), "Desktop")` for all test paths |
| Tests read `os.Getenv("APPDATA")` without `t.Setenv` | Test uses real APPDATA path, writing to actual app data directory | Always `t.Setenv("APPDATA", t.TempDir())` for any test that calls `SaveBackup` or `configDir()` |

---

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Junction creation tests:** Verify read-through works (file content accessible via junction), not just `IsJunction() == true`. BUG-02 (buffer size mismatch) shows that a "junction" can be created but be non-functional.
- [ ] **Junction deletion tests:** Verify `IsJunction() == false` AND that the directory still exists as a normal directory (not deleted).
- [ ] **Registry write tests:** Verify the exact value and type written, not just the absence of error. `WriteExpandString` writing an empty string succeeds silently.
- [ ] **Cancellation tests:** Verify goroutines actually stop — not just that `IsCancelled()` returns true. Use goroutine leak detection (e.g., `runtime.NumGoroutine()` before/after).
- [ ] **Rollback tests:** Verify all 5 steps in `rollback.go` complete. A test that checks only step 5 misses steps 1-4.
- [ ] **Cross-volume move tests:** Test on both same-volume AND different-volume drives. `moveFilesSameVolume` (os.Rename) and `moveFilesCrossVolume` (copy + delete) have completely different failure modes.
- [ ] **onProgress callback tests:** Verify the callback is called, the values are monotonically increasing, and percent never exceeds 100. BUG-05 and BUG-07 show these invariants are violated.
- [ ] **Error type wrapping:** Verify that `MoveKitError` wraps the underlying error correctly so `errors.As` and `errors.Is` work. BUG-03 suggests `fmt.Errorf("...: %w", err)` is missing.
- [ ] **Wails binding tests (app.go):** Verify that `StartMigration` sets `a.migrating` correctly, emits the right events, and respects the mutex. Currently zero test coverage.
- [ ] **Frontend tests:** Verify that React components handle `migration:progress` events correctly (percent display, folder names), error states (displaying error code/message/action), and cancellation (button state toggle).

---

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Junction test leakage (orphaned temp dirs) | LOW | `dir %TEMP%\*TestCreate*` to find and manually `rmdir` leaked dirs |
| Registry pollution (real HKCU keys overwritten) | MEDIUM | Restore original registry values from Windows backup or manually reset Known Folder paths in Explorer properties |
| Rollback test trashes real Desktop | HIGH | Immediately restore from `%APPDATA%\MoveKit\backups\Desktop.json` backup record; if backup is gone, manually recreate junction with `mklink /J` |
| `CopyFileWithACL` sets restrictive permissions on test temp dir | LOW | Take ownership of temp dir via Explorer properties; use `icacls %TEMP%\... /reset` to restore inherited permissions |
| Shared `cancelled` flag causes flaky tests | MEDIUM | Add `t.Cleanup(func() { migrate.ResetCancellation() })` to every test as temporary mitigation; refactor to context-based cancellation as permanent fix |

---

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| P1: Junction test leakage | Phase 1 (syscall testing) | All junction-creating tests include `t.Cleanup(DeleteJunction)`; zero `TempDir RemoveAll` failures in CI |
| P2: Registry pollution | Phase 1 (syscall testing) | `RegistryAccessor` interface exists; all migrate tests use fake registry; integration tests under build tag only |
| P3: Wrong mock layer | Phase 1 (arch decision before tests) | Interface strategy documented; each syscall function has a decision: "real test" or "interface + fake" |
| P4: Shared cancelled flag | Phase 1 (or pre-phase bug fix) | `moveFilesCrossVolume` accepts `context.Context`; no package-level `atomic.Int32` for cancellation |
| P5: Rollback test destruction | Phase 2 (migrate testing) | Rollback tests use fake registry + fake filesystem; zero real HKCU writes during `go test` |
| P6: Wrong VFS choice | Phase 1 (arch decision) | No import of `spf13/afero` or similar generic VFS; MoveKit-specific interfaces used instead |
| P7: ACL permission cleanup | Phase 1 (syscall testing) | `CopyFileWithACL` tests restore permissions in `t.Cleanup`; cleanup never fails |
| P8: Error swallowing | Phase 0 (bug fixes) | Each BUG-XX is fixed before its test is written; test validates both success AND specific error code |

---

## Sources

### Official Go Issues (HIGH confidence)
- golang/go#51442 — `testing: TempDir RemoveAll cleanup failures on Windows`
- golang/go#45672 — `debug/pe: "Access is denied" during Windows cleanup`
- golang/go#65545 — `x/pkgsite: TempDir RemoveAll cleanup errors on Windows`
- golang/go#21322 — `os: inconsistent cross-platform OpenFile/Write behavior`
- golang/go#36532 — `testing: reconsider adding Context method to testing.T`
- golang/go#58211 — `testing: test-scoped context proposal`

### Go Standard Library Patterns (MEDIUM confidence)
- `testing/fstest.MapFS` — read-only in-memory filesystem, not suitable for write operations
- `cmd/go/internal/robustio` — Windows retry pattern for `RemoveAll`
- `golang.org/x/sys/windows/registry` — no built-in mocking support

### Domain-Specific Patterns (MEDIUM confidence)
- Chromium's `RegistryOverrideManager` — registry hive override for test isolation (C++ pattern, no Go equivalent)
- `spf13/afero` — Go VFS library lacking Windows NTFS junction, ACL, and case-insensitive path support

### MoveKit Codebase Analysis (HIGH confidence — directly observed)
- BUG-01 through BUG-08 — 8 known bugs affecting testability (see `.planning/CONCERNS.md`)
- `junction_test.go` and `junction_debug_test.go` — existing patterns for junction testing
- `move.go` — package-level `cancelled atomic.Int32` shared mutable state
- `registry.go` — direct `golang.org/x/sys/windows/registry` calls without abstraction

---
*Pitfalls research for: MoveKit v0.1.0 testing & QA (greenfield-quality milestone)*
*Researched: 2026-06-21*
