# Architecture Research — Test Architecture for Multi-Layer Wails Desktop App

**Domain:** Windows desktop application testing and quality assurance (MoveKit v0.1.0)
**Researched:** 2026-06-21
**Confidence:** HIGH

## System Overview — Five-Layer Test Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                      L5: End-to-End Tests (Playwright)                     │
│  Real browser → wails dev (localhost:34115) → Go backend → Win32 API      │
│  Full user journeys: start app → select disk → migrate → verify result    │
├──────────────────────────────────────────────────────────────────────────┤
│                    L4: Integration Tests (Go test binary)                   │
│  Full Migrate() pipeline in t.TempDir() with real NTFS junctions           │
│  App.StartMigration() → migrate.Migrate() → syscall.* (no Wails runtime)  │
├──────────────────────────────────────────────────────────────────────────┤
│                    L3: React Component Tests (Vitest + RTL)                 │
│  Vitest + jsdom → React Testing Library → wails.ts mock proxy             │
│  Per-component rendering, state transitions, user interactions            │
├──────────────────────────────────────────────────────────────────────────┤
│                    L2: Go App Layer Tests (testify suite)                   │
│  App struct with injected interfaces → testify/mock → no Wails runtime    │
│  GetFolders, GetDisks, StartMigration, Rollback — all 17 bound methods    │
├──────────────────────────────────────────────────────────────────────────┤
│                    L1: Go Unit Tests (standard testing)                     │
│  Pure Go logic: syscall wrappers, migrate helpers, folders data           │
│  Table-driven, t.TempDir(), fstest.MapFS — no external deps              │
└──────────────────────────────────────────────────────────────────────────┘
```

## Component Boundaries — What Talks to What

| Layer | Component | Responsibility | Communicates With | Test Tools |
|-------|-----------|---------------|-------------------|------------|
| L5 | E2E Specs | Full user journeys: startup → scan → migrate → rollback | Wails dev server (localhost:34115) | Playwright |
| L4 | Integration Tests | Go-side pipeline: Migrate(), Rollback(), PreCheck() with real NTFS junctions | internal/migrate, internal/syscall | Go test + t.TempDir() |
| L3 | Component Tests | React UI: state machine transitions, dialog overlays, progress bar rendering | lib/wails.ts (mocked), types.ts | Vitest, React Testing Library, jsdom |
| L2 | App Layer Tests | Wails-bound methods: GetFolders, GetDisks, StartMigration, Rollback | App struct (with injected interfaces), internal/migrate, internal/syscall | testify/suite, testify/mock |
| L1 | Unit Tests | Pure Go logic: ToLongPath, IsCrossVolume, maskPath, conflict detection, error types | internal/migrate, internal/syscall, internal/folders | Go standard testing, table-driven, t.TempDir(), fstest.MapFS |

### Key Dependency Rule

```
L5 (E2E) ──depends on──▶ wails dev (requires running Go + React)
L4 (Integration) ──depends on──▶ L1 + L2 (imports internal/* packages directly)
L3 (Component) ──depends on──▶ mock wails.ts (independent of Go)
L2 (App Layer) ──depends on──▶ L1 (imports internal/* packages, mocks Wails runtime)
L1 (Unit) ──depends on──▶ nothing (pure Go logic or real Win32 in t.TempDir)
```

**Build order implication:** L1 first (no deps), then L2 (depends on L1 helpers), then L4 (depends on L2 patterns). L3 and L5 are independently buildable — L3 needs only the mock proxy, L5 needs `wails dev`.

## Recommended Project Structure

```
MoveKit/
├── main.go                      # Wails entry (unchanged)
├── app.go                       # App struct — refactored for testability
├── app_types.go                 # Wails-facing types (unchanged)
│
├── internal/
│   ├── folders/
│   │   ├── known.go
│   │   └── known_test.go        # NEW: verify AllFolders data integrity
│   │
│   ├── migrate/
│   │   ├── migrate.go
│   │   ├── migrate_test.go      # NEW: Migrate() integration test
│   │   ├── check.go
│   │   ├── check_test.go
│   │   ├── backup.go
│   │   ├── backup_test.go
│   │   ├── rollback.go
│   │   ├── rollback_test.go     # NEW: Rollback() integration test
│   │   ├── move.go
│   │   ├── move_test.go
│   │   ├── conflict.go
│   │   ├── conflict_test.go
│   │   ├── onedrive.go
│   │   ├── onedrive_test.go
│   │   ├── cloud.go
│   │   ├── cloud_test.go        # NEW: cloud sync detection tests
│   │   ├── security.go
│   │   ├── security_test.go     # NEW: junction hijack tests
│   │   ├── state.go
│   │   ├── state_test.go        # NEW: state persistence tests
│   │   ├── verify.go
│   │   ├── verify_test.go       # NEW: integrity verification tests
│   │   ├── log.go
│   │   ├── log_test.go          # NEW: logger thread-safety tests
│   │   ├── version.go
│   │   ├── version_test.go      # NEW: version check tests
│   │   ├── errors.go            # (types only, no test needed)
│   │   └── types.go             # (types only, no test needed)
│   │
│   ├── syscall/
│   │   ├── junction.go
│   │   ├── junction_test.go
│   │   ├── junction_debug_test.go
│   │   ├── registry.go
│   │   ├── registry_test.go
│   │   ├── fileops.go
│   │   ├── fileops_test.go
│   │   ├── fileops_acl_test.go
│   │   ├── volume.go
│   │   ├── volume_test.go
│   │   ├── process.go
│   │   ├── process_test.go      # NEW: process enumeration + mutex tests
│   │   ├── power.go
│   │   ├── power_test.go
│   │   ├── shell.go
│   │   └── shell_test.go        # NEW: shell notification tests
│   │
│   └── testutil/                # NEW: shared test helpers for Go tests
│       ├── fs.go                # CreateTempTree(t) — realistic folder trees
│       ├── registry.go          # StubRegistry — in-memory registry for tests
│       └── mock_wails.go        # MockWailsRuntime — fakes EventsEmit/EventsOn
│
├── integration/                 # NEW: cross-package integration tests
│   ├── migrate_full_test.go     # Full Migrate() → Rollback() roundtrip
│   ├── app_integration_test.go  # App.StartMigration + event emission
│   └── helpers_test.go          # Shared integration test fixtures
│
├── frontend/
│   ├── src/
│   │   ├── App.tsx
│   │   ├── App.test.tsx         # NEW: state machine transitions
│   │   ├── types.ts
│   │   ├── main.tsx
│   │   ├── lib/
│   │   │   ├── wails.ts
│   │   │   └── wails.test.ts    # NEW: mock/real binding delegation
│   │   ├── components/
│   │   │   ├── DiskSelector.tsx
│   │   │   ├── DiskSelector.test.tsx  # NEW
│   │   │   ├── FolderList.tsx
│   │   │   ├── FolderList.test.tsx    # NEW
│   │   │   ├── ProgressBar.tsx
│   │   │   ├── ProgressBar.test.tsx   # NEW
│   │   │   ├── ResultPage.tsx
│   │   │   ├── ResultPage.test.tsx    # NEW
│   │   │   ├── ErrorBoundary.tsx
│   │   │   ├── ErrorBoundary.test.tsx # NEW
│   │   │   ├── OneDriveWarning.tsx
│   │   │   ├── CloudWarning.tsx
│   │   │   ├── ConflictDialog.tsx
│   │   │   ├── DefenderInfo.tsx
│   │   │   ├── FirstRunGuide.tsx
│   │   │   ├── HelpPanel.tsx
│   │   │   └── UpdateNotice.tsx
│   │   └── test/
│   │       └── setup.ts         # NEW: global Vitest + jsdom setup
│   ├── vitest.config.ts         # NEW: Vitest configuration
│   └── package.json             # ADD: vitest, @testing-library/*, jsdom deps
│
└── e2e/                         # NEW: Playwright end-to-end tests
    ├── playwright.config.ts
    └── specs/
        ├── full-migration.spec.ts
        └── rollback.spec.ts
```

### Structure Rationale

- **Co-located `*_test.go`**: Go convention. Tests live next to source. White-box (`package migrate`, not `package migrate_test`) for accessing unexported internals.
- **`internal/testutil/`**: Shared test helpers that both `migrate` and `syscall` tests can import. Avoids duplication of temp-tree creation, registry stubbing, and Wails runtime mocking.
- **`integration/`**: Tests that exercise the full Go pipeline without the Wails GUI. These are plain Go tests (not a separate binary) that import `internal/*` packages. They depend on real NTFS junctions but run in `t.TempDir()`.
- **`frontend/src/test/setup.ts`**: Global Vitest configuration — jsdom environment, `@testing-library/jest-dom` matchers, mock window APIs not available in jsdom (matchMedia, ResizeObserver, IntersectionObserver).
- **`e2e/`**: Playwright specs that require `wails dev` to be running. Test full user journeys end-to-end with real Go backend.
- **No `cmd/verify/` changes**: `cmd/verify/main.go` remains as a manual verification tool, distinct from automated tests.

## Data Flow — How Test Information Moves

### L1/L2/L4: Go Test Data Flow

```
Test function (t *testing.T)
    │
    ├─ t.TempDir() ───────────▶ Isolated filesystem (auto-cleaned)
    │   ├── os.MkdirAll        Create realistic folder trees
    │   ├── os.WriteFile       Seed test files with known content
    │   └── CreateJunction()   Real NTFS junctions (Windows-only)
    │
    ├─ StubRegistry ──────────▶ In-memory registry (testutil/registry.go)
    │   └── ReadExpandString() Returns known paths instead of real HKCU
    │
    ├─ MockWailsRuntime ──────▶ Captures EventsEmit/EventsOn calls
    │   └── Assert event was emitted with correct payload
    │
    └─ Assertions ────────────▶ Verify files, registry, event state
        ├── t.Errorf / t.Fatalf          (standard)
        ├── assert.Equal / require.NoError (testify)
        └── mock.AssertExpectations(t)    (testify/mock)
```

### L3: React Component Test Data Flow

```
Vitest test runner
    │
    ├─ jsdom ─────────────────▶ Browser-like DOM environment
    │
    ├─ render(<Component />) ─▶ renders via @testing-library/react
    │   ├── Props from test ──▶ Inline mock data (no real Wails calls)
    │   └── User events ──────▶ userEvent.click / userEvent.type
    │
    ├─ Mock wails.ts ─────────▶ vi.mock('../lib/wails'), returns controlled data
    │   └── Can simulate: success, failure, loading, network error
    │
    └─ Assertions ────────────▶ screen.getByText / getByRole / waitFor
        ├── expect(...).toBeInTheDocument()
        ├── expect(...).toHaveTextContent()
        └── expect(onClick).toHaveBeenCalledTimes(1)
```

### L5: E2E Test Data Flow (Full Stack)

```
Playwright test runner
    │
    ├─ page.goto('http://localhost:34115') ──▶ Wails dev server
    │                                            │
    │   ┌────────────────────────────────────────┘
    │   ▼
    │   WebView2 rendering ──▶ React components render ──▶ User sees UI
    │   │                                                                 │
    │   │  User clicks "开始迁移"                                         │
    │   ▼                                                                 │
    │   wailsjs/go/main/App.js ──▶ WebSocket ──▶ Wails Runtime            │
    │                                                │                    │
    │                                                ▼                    │
    │                                          App.StartMigration()      │
    │                                                │                    │
    │                                                ▼                    │
    │                                          migrate.Migrate()         │
    │                                                │                    │
    │                                                ▼                    │
    │                                          syscall.CreateJunction()  │
    │                                                │                    │
    │   ◄─── runtime.EventsEmit("migration:progress")                    │
    │   ▼                                                                 │
    │   React state updates ──▶ ProgressBar redraws                      │
    │                                                                     │
    └─ Assertions ────────────▶ page.getByText / expect(page.locator(...))
        ├── Verify "迁移完成" message appears
        ├── Verify progress bar reached 100%
        └── Verify target folder is a junction (via page.evaluate + Go call)
```

## Architectural Patterns for Testability

### Pattern 1: Interface Injection on App Struct (Go L2)

**What:** The `App` struct currently has concrete dependencies on `internal/migrate` and `internal/syscall` packages. To make it testable without the Wails runtime, introduce interfaces for the subset of operations App uses from those packages. Default to real implementations, override in tests.

**When to use:** Testing any of the 17 Wails-bound methods on App. Required for `GetFolders`, `GetDisks`, `StartMigration`, `Rollback`, and event-emitting methods.

**Trade-offs:** Adds indirection, but the interface surface is small (each bound method calls 1-3 internal functions). The current code already has a clean separation between App (orchestration) and internal/* (logic) — interfaces formalize this boundary.

**Example:**
```go
// app.go — refactored for testability
type App struct {
    ctx      context.Context
    mu       sync.Mutex
    migrating bool

    // Injected dependencies (defaults to real implementations)
    folderReader  FolderReader     // reads registry + junction status
    diskChecker   DiskChecker      // enumerates drives
    migrator      Migrator         // runs Migrate() pipeline
    backupStore   BackupStore      // CRUD for backup records
    eventEmitter  EventEmitter     // wraps runtime.EventsEmit
}

type FolderReader interface {
    GetFolderInfo(regValueName string) (FolderInfo, error)
}

type Migrator interface {
    Migrate(regValueName, targetDrive string, onProgress ProgressCallback) error
    Rollback(regValueName string) error
}
```

### Pattern 2: Mock Wails Binding Proxy (React L3)

**What:** `frontend/src/lib/wails.ts` already supports dual-mode (real bindings or mock fallback). Extend this so tests can inject mock return values via Vitest's `vi.mock()`. Each exported function becomes a `vi.fn()` that returns controlled data.

**When to use:** Every React component test. Never call real Go bindings from Vitest — they don't exist in the Vitest/Node.js environment.

**Example:**
```typescript
// ComponentFoo.test.tsx
import { vi, describe, test, expect } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock the entire wails module before importing the component
vi.mock('../lib/wails', () => ({
    GetFolders: vi.fn().mockResolvedValue([
        { name: '桌面', path: 'C:\\Users\\Test\\Desktop', status: 'default', sizeBytes: 1234 }
    ]),
    GetDisks: vi.fn().mockResolvedValue([
        { drive: 'D:', label: '数据盘', freeBytes: 100_000_000_000, isSSD: false, isNTFS: true }
    ]),
}));

import DiskSelector from './DiskSelector';

describe('DiskSelector', () => {
    test('shows disk label and free space', async () => {
        render(<DiskSelector disks={[{ drive: 'D:', label: '数据盘', freeBytes: 100_000_000_000, isSSD: false, isNTFS: true }]} selected={null} onSelect={vi.fn()} />);
        expect(screen.getByText('D:')).toBeInTheDocument();
        expect(screen.getByText(/数据盘/)).toBeInTheDocument();
    });
});
```

### Pattern 3: Real NTFS Junction Integration Tests (Go L4)

**What:** Use `t.TempDir()` to create isolated filesystem sandboxes on NTFS volumes. Run the full `Migrate()` pipeline — copy files, create junctions, update a stubbed registry, verify results by reading through the junction. Cleanup is automatic via `t.Cleanup()`.

**When to use:** Testing `Migrate()`, `Rollback()`, `moveFilesCrossVolume()`, and the junction create/delete roundtrip with real `DeviceIoControl` calls.

**Trade-offs:** These tests are Windows-only (NTFS junction requirement) and slow (real filesystem I/O). Run them in CI only on Windows runners. Use build tags `//go:build windows` or skip with `t.Skip()` on non-Windows platforms.

**Example:**
```go
func TestMigrate_Roundtrip(t *testing.T) {
    // Create source folder with known content in a temp dir
    srcDir := t.TempDir()
    srcContent := filepath.Join(srcDir, "Documents")
    os.MkdirAll(srcContent, 0755)
    os.WriteFile(filepath.Join(srcContent, "report.docx"), []byte("Q4 Report"), 0644)

    // Set up stub registry pointing to srcContent
    reg := testutil.NewStubRegistry()
    reg.Set("Desktop", srcContent)

    // Run migration to a target drive (using another temp dir as "target drive")
    targetDrive := filepath.Join(t.TempDir(), "target") + string(os.PathSeparator)
    os.MkdirAll(targetDrive, 0755)

    err := migrate.Migrate(reg, "Desktop", "T:", migrate.ProgressCallback(func(...) {}))
    require.NoError(t, err)

    // Verify: source is now a junction
    assert.True(t, syscall.IsJunction(srcContent))

    // Verify: files accessible through junction
    data, err := os.ReadFile(filepath.Join(srcContent, "report.docx"))
    require.NoError(t, err)
    assert.Equal(t, "Q4 Report", string(data))

    // Rollback
    err = migrate.Rollback(reg, "Desktop")
    require.NoError(t, err)

    // Verify: junction removed, original content restored
    assert.False(t, syscall.IsJunction(srcContent))
    data, _ = os.ReadFile(filepath.Join(srcContent, "report.docx"))
    assert.Equal(t, "Q4 Report", string(data))
}
```

### Pattern 4: E2E with Playwright Against Dev Server (L5)

**What:** Run `wails dev` in one terminal. Playwright navigates to `http://localhost:34115` (or the auto-detected port). Go backend bindings are fully active. Tests exercise complete user journeys.

**When to use:** Smoke tests for the full user flow. Verify that UI interactions correctly trigger Go backend methods and that Go-to-frontend events update the UI. Not for exhaustive testing — L1-L4 handle that.

**Trade-offs:** Requires `wails dev` to be running. Can't run in standard CI without a Windows desktop environment. Mark these as `@smoke` tests that run on release validation, not every commit.

**Example:**
```typescript
// e2e/specs/full-migration.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Full Migration Flow', () => {
    test('migrates Desktop folder to D: drive', async ({ page }) => {
        await page.goto('http://localhost:34115');

        // Wait for folder list to load
        await expect(page.getByText('桌面')).toBeVisible({ timeout: 10000 });

        // Select D: drive
        await page.getByRole('radio', { name: /D:/ }).click();

        // Start migration
        await page.getByRole('button', { name: '开始迁移' }).click();

        // Wait for progress
        await expect(page.getByText(/正在迁移/)).toBeVisible();

        // Wait for completion
        await expect(page.getByText('迁移完成')).toBeVisible({ timeout: 60000 });

        // Verify result shows freed space
        await expect(page.getByText(/已释放/)).toBeVisible();
    });
});
```

### Pattern 5: StubRegistry for Test Isolation (Go L1/L2/L4)

**What:** Replace real Windows registry access with an in-memory map during tests. The `StubRegistry` implements the same interface that `internal/syscall/registry.go` exposes. Tests inject the stub instead of calling `ReadExpandString` on the real HKCU hive.

**When to use:** Every test that touches registry-dependent code: `GetFolders()`, `Migrate()`, `Rollback()`, `DetectOneDriveKFM()`.

**Why not mock at the syscall layer:** The registry functions are thin wrappers around `golang.org/x/sys/windows/registry`. Mocking at that level (DLL calls) is fragile. Instead, introduce an interface at the `internal/migrate` boundary — tests stub the registry, production uses the real API.

```go
// internal/testutil/registry.go
type StubRegistry struct {
    userShellFolders map[string]string  // {Desktop => "C:\Users\X\Desktop"}
    shellFolders     map[string]string
    oneDriveKFM      bool
}

func (r *StubRegistry) ReadExpandString(key, valueName string) (string, error) {
    if v, ok := r.userShellFolders[valueName]; ok {
        return v, nil
    }
    return "", migrate.ErrRegistryKeyNotFound
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Testing App Methods with Real Wails Runtime

**What people do:** Try to call `app.GetFolders()` in a Go test and have it hit real Wails `runtime.EventsEmit` which panics without a running WebView2.

**Why it's wrong:** Wails runtime methods (`runtime.EventsEmit`, `runtime.EventsOn`) require a running WebView2 host. They will panic or return nil contexts in a plain `go test`.

**Do this instead:** Inject an `EventEmitter` interface into App. In production, the implementation wraps `runtime.EventsEmit`. In tests, use a mock that records emitted events for assertion.

### Anti-Pattern 2: Frontend Tests That Call Real Wails Bindings

**What people do:** Import real `wailsjs/go/main/App.js` in Vitest and call `GetFolders()`. The JS binding tries to call `window.go.main.App.GetFolders()` which doesn't exist in jsdom.

**Why it's wrong:** The Wails bridge only exists inside WebView2. In Vitest's jsdom environment, `window.go` is undefined — the call fails silently or throws.

**Do this instead:** Mock `lib/wails.ts` with Vitest's `vi.mock()`. Every test controls exactly what data the mock returns. Test the mock proxy itself in `wails.test.ts` to verify it delegates correctly when real bindings are available.

### Anti-Pattern 3: Unit Testing Migrate() Without Stubbing syscall

**What people do:** Call `Migrate()` in a Go test without stubbing registry/wails, causing it to read real HKCU keys and attempt real junction creation on production directories.

**Why it's wrong:** This would modify the developer's actual user profile folders. Junction creation in real paths could break the developer's Windows environment.

**Do this instead:** Inject a `RegistryReader` interface and a `JunctionCreator` interface into the Migrate function. In tests, use `StubRegistry` and a junction creator that operates within `t.TempDir()`. Every migration test operates in an isolated temp directory.

### Anti-Pattern 4: Global Test Fixtures Without Cleanup

**What people do:** Create a `testdata/` directory with pre-built folder trees, commit it, and have tests depend on it without cleanup.

**Why it's wrong:** Committed test fixtures drift from code changes, can't be generated on the fly for edge cases, and create merge conflicts. Filesystem fixtures are particularly dangerous because NTFS-specific attributes (junctions, ACLs) don't survive Git.

**Do this instead:** Create all test data inline in `t.TempDir()` using `os.MkdirAll` and `os.WriteFile`. Use `testutil.CreateTempTree(t, spec)` for complex folder structures. Never commit binary filesystem artifacts.

### Anti-Pattern 5: Skipping Cancellation Tests Because They're Hard

**What people do:** Test only the happy path for long-running operations like `Migrate()`. Leave cancellation untested because it requires goroutine synchronization.

**Why it's wrong:** Cancellation is the most error-prone part of the system. The current codebase has a known bug where the frontend cancel button doesn't actually call `CancelMigration()`. Untested cancellation means silent data corruption on user abort.

**Do this instead:** Write concurrent tests: start `Migrate()` in a goroutine, wait for progress events (channel-based), call `CancelMigration()`, verify the goroutine exits with a cancellation error within a timeout. The `move_test.go` already demonstrates this pattern — extend it to the full pipeline.

## Integration Points

### Internal Boundaries — Test Implications

| Boundary | Current State | Test Impact | Refactoring Needed |
|----------|---------------|-------------|-------------------|
| App → migrate | Concrete imports, no interfaces | App methods untestable without Wails runtime | Introduce Migrator, FolderReader, EventEmitter interfaces on App struct |
| migrate → syscall | Concrete imports, Type duplication (BackupRecord, MoveKitError in both packages) | migrate tests must use real syscall (NTFS junction creation) | Make syscall operations injectable via interfaces; use StubRegistry |
| React → Wails bindings | `lib/wails.ts` dual-mode proxy (real/mock) | Already partially testable — just needs mocking extension | Formalize the mock protocol in `wails.test.ts` |
| Wails Runtime → Go App | Auto-generated bindings in `wailsjs/` | No test coverage for binding correctness | `wails generate module` validates; unit test the App methods directly |
| Go App → Wails Events | `runtime.EventsEmit("migration:progress", ...)` | Event payload format untested | Inject EventEmitter interface; mock in App tests; verify payload shape |
| All → Windows OS | Win32 DLL calls via `syscall.NewLazyDLL` | Tests run on real Windows API (no DLL mocking) | Keep L1/L4 Windows-only. Use `//go:build windows` tags. Accept that some tests need a Windows CI runner. |

### External Service Integration — Test Approach

MoveKit has no external HTTP APIs. All "external" integration is with Windows OS facilities:

| Win32 Facility | Test Isolation Strategy |
|----------------|------------------------|
| NTFS Junctions (DeviceIoControl) | Create in `t.TempDir()` — works on any NTFS volume. No mock needed. |
| Registry (HKCU/HKLM) | Inject StubRegistry for unit/integration tests. Keep one smoke test that reads real HKCU (graceful skip if key missing). |
| File Copy (CopyFileExW + ACL) | Exercise in `t.TempDir()` with real `CopyFileEx`. ACL tests need real files to copy ACL from — create temp files. |
| Process Enumeration (toolhelp32) | Stub the process list for unit tests. Smoke test with real process list (always returns at least the test process itself). |
| Power Status | Stub for unit tests (returns AC/battery based on test parameter). Smoke test reads real status but never fails — `t.Log` and continue. |
| Shell Notifications (SHChangeNotify) | Test that the function is called with correct parameters. The actual notification is a fire-and-forget syscall — no return value to verify. |

## Test Fixture Management

### Recommended Approach: Inline Creation + Shared Helpers

**No fixture files committed to the repository.** All test data created inline within tests.

```
Fixture Strategy Decision Tree:

Is the setup trivial (1-2 files, 1 directory)?
├── YES: Create inline in the test function with os.MkdirAll + os.WriteFile
└── NO: Does the setup pattern repeat across multiple tests?
        ├── YES: Extract to testutil.CreateTempTree(t, spec)
        └── NO: Create inline — complex setup IS the test's value

Can the test work with fstest.MapFS?
├── YES: Use fstest.MapFS (fast, no cleanup, parallel-safe)
└── NO (needs real FS behavior): Use t.TempDir()
```

### Shared Test Helpers (`internal/testutil/`)

```go
// internal/testutil/fs.go

// FolderSpec describes a directory tree to create in a temp dir.
type FolderSpec struct {
    Name     string
    Files    map[string]string  // filename → content
    Subdirs  []FolderSpec
}

// CreateTempTree creates a directory tree from a spec inside t.TempDir().
// Returns the root path.
func CreateTempTree(t *testing.T, spec FolderSpec) string {
    t.Helper()
    root := t.TempDir()
    createTree(t, root, spec)
    return root
}

// Usage in tests:
func TestDetectConflicts(t *testing.T) {
    src := CreateTempTree(t, FolderSpec{
        Name: "src",
        Files: map[string]string{
            "shared.txt":  "src version",
            "unique.txt":  "only in src",
        },
        Subdirs: []FolderSpec{{
            Name: "nested",
            Files: map[string]string{"deep.txt": "nested content"},
        }},
    })
    // ... test conflict detection between src and dst
}
```

## Suggested Build Order (Dependencies Between Components)

Phase order for implementing the test architecture:

```
Phase A: Foundation — Make code testable
├── 1. Refactor App struct: inject interfaces (FolderReader, Migrator, EventEmitter, etc.)
├── 2. Create internal/testutil/ package: StubRegistry, CreateTempTree, MockEventEmitter
├── 3. Add Vitest + React Testing Library to frontend/package.json
├── 4. Create frontend/src/test/setup.ts with jsdom configuration
└── 5. Create vitest.config.ts

Phase B: Fill gaps — Layer 1 (Go Unit Tests)
├── 6. Write _test.go for ALL currently untested files in internal/migrate/ and internal/syscall/
└── 7. Standard testing + table-driven + t.TempDir()

Phase C: Test the bridge — Layer 2 (App Layer Tests)
├── 8. Write app_test.go using testify/suite with injected mocks
└── 9. Cover all 17 Wails-bound methods (GetFolders through NotifyCompletion)

Phase D: Test the UI — Layer 3 (React Component Tests)
├── 10. Write *.test.tsx for core components: App, DiskSelector, FolderList, ProgressBar, ResultPage
└── 11. Write wails.test.ts for mock proxy verification

Phase E: Test the pipeline — Layer 4 (Integration Tests)
├── 12. Write Migrate() integration test with real NTFS junctions in t.TempDir()
├── 13. Write Rollback() integration test
└── 14. Write App.StartMigration() integration test with event assertions

Phase F: Smoke tests — Layer 5 (E2E Tests)
├── 15. Set up Playwright configuration
└── 16. Write critical-path E2E specs (launch → scan → migrate → verify)

Phase G: CI and coverage gates
├── 17. Set up GitHub Actions CI for go test + vitest
├── 18. Add coverage thresholds (80%+ Go, 70%+ frontend)
└── 19. Add Windows-specific CI runner for L4 integration tests
```

**Phase ordering rationale:**
- Phase A MUST come first — the refactoring to inject interfaces enables all downstream testing. Without it, neither App layer tests (L2) nor integration tests (L4) are practical.
- Phase B (unit tests) is parallelizable with Phase A. Standard Go tests don't need the App refactoring.
- Phase C depends on Phase A (needs the injected interfaces).
- Phase D (React tests) is fully independent of Go-side changes — can be done in parallel with B and C.
- Phase E depends on A and B (needs testutil package and working unit tests to build on).
- Phase F depends on the application being functional and testable. Needs wails dev to work.
- Phase G ties everything together — validates coverage thresholds are met.

## Scalability Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Current (1 developer, 1 platform) | Single test suite, manual runs, Windows-only `go test ./...` |
| Multiple contributors | CI pipeline on push/PR. go test + vitest matrix. Windows runner for L4. |
| Adding macOS/Linux (out of scope) | `//go:build` tags segregate platform-specific tests. Cross-platform CI matrix. |
| Large test suite (>1000 test cases) | Split slow integration tests from fast unit tests. `go test -short` skips L4/L5. Test sharding in CI. |

### Scaling Priorities

1. **First bottleneck:** Slow L4 integration tests (real NTFS junctions) block CI. **Fix:** Tag integration tests with `//go:build integration` and run them on a scheduled basis, not per-commit.
2. **Second bottleneck:** Vitest startup time with Tailwind CSS v4 processing. **Fix:** Use Vitest's `css: true` config option. Tailwind v4's CSS-based approach is faster than v3's JIT compilation for testing.
3. **Third bottleneck:** Playwright E2E tests require a Windows desktop environment. **Fix:** Tag as `@smoke`, run manually before release. Use `wails build` + Windows Sandbox for fully automated E2E.

## Sources

- Wails v2 Documentation and GitHub Discussions (#4205 on E2E testing) — HIGH confidence
- [Testify Documentation — stretchr/testify](https://github.com/stretchr/testify) — HIGH confidence
- [Vitest Documentation — vitest.dev](https://vitest.dev/) — HIGH confidence
- [React Testing Library Documentation](https://testing-library.com/docs/react-testing-library/intro) — HIGH confidence
- [Playwright Documentation](https://playwright.dev/) — HIGH confidence
- [Go testing package — t.TempDir and fstest.MapFS](https://pkg.go.dev/testing) — HIGH confidence
- [Go test fixture management patterns](https://datasea.cn/go1210351334.html) — MEDIUM confidence (community blog, verified against official Go docs)
- MoveKit codebase analysis: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/TESTING.md`, `.planning/codebase/STRUCTURE.md` — HIGH confidence (direct codebase inspection)
- [Better Stack Community — Testing in Go with Testify](https://betterstack.com/community/guides/scaling-go/golang-testify/) — MEDIUM confidence
- [OneUptime Blog — Unit Test React with Vitest and RTL](https://oneuptime.com/blog/post/2026-01-15-unit-test-react-vitest-testing-library/view) — MEDIUM confidence

---
*Architecture research for: MoveKit test architecture — greenfield-quality milestone*
*Researched: 2026-06-21*
