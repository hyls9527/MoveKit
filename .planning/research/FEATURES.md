# Feature Landscape — Testing & Quality

**Domain:** Windows Desktop Application Testing & Quality Assurance (MoveKit quality-focused iteration)
**Researched:** 2026-06-21
**Confidence:** HIGH

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features a production-ready Windows desktop tool MUST have. Missing any of these means the product is not trustworthy.

| # | Feature | Why Expected | Complexity | Notes |
|---|---------|--------------|------------|-------|
| T1 | **Go Unit Test Coverage >= 80%** | Core migration logic is destructive (moves user files). Without tests, any regression = data loss risk. Users trust tools that prove correctness. | MEDIUM | Currently ~30% (11/26 files). Need tests for: `migrate.go`, `cloud.go`, `errors.go`, `log.go`, `rollback.go`, `security.go`, `state.go`, `types.go`, `verify.go`, `version.go`, `app.go`, `app_types.go`, `process.go`, `shell.go`, `folders/known.go` |
| T2 | **CI/CD Pipeline with Automated Test Runs** | Every commit must pass tests before merge. For a tool that moves user data, "it worked on my machine" is unacceptable. | MEDIUM | GitHub Actions: go test -race, golangci-lint, go vet, build verification. Currently zero CI. |
| T3 | **Regression Tests for All 8 Known Bugs** | Users hit these bugs. Fixing without a regression test guarantees they'll come back. Each bug fix MUST ship with a test that reproduces the failure. | LOW (per bug) | BUG-01 through BUG-08 from CONCERNS.md. Test-first: write the failing test, then fix. |
| T4 | **Static Analysis (golangci-lint + go vet)** | Catches nil derefs, unchecked errors, dead code, race conditions BEFORE they ship. MoveKit already has bugs from these categories (BUG-03 silent errors, BUG-07 dead code). | LOW | Add `.golangci.yml` config. Enable: errcheck, govet, staticcheck, gosec, unused, gosimple. Run in CI. |
| T5 | **Structured Error Handling Audit** | Migration failures must produce actionable Chinese-language messages. Silent error swallowing (BUG-03, BUG-05, BUG-06, BUG-08) erodes trust. | LOW | Review every `_ = err` and `if err != nil { return }` in the codebase. Add error propagation. |
| T6 | **Build Verification** | The portable exe must build cleanly. Broken builds waste developer time and block releases. | LOW | `wails build` in CI. Verify output exe exists and is non-zero. |
| T7 | **Code Formatting Standards** | Consistent formatting = readable diffs = faster code review = fewer bugs missed. Go has `gofmt`. TypeScript needs Prettier/ESLint. | LOW | Add `gofmt -l .` check in CI. Add Prettier + ESLint to frontend. |
| T8 | **Test Isolation (Temp Dirs, No Global State)** | Tests that depend on real filesystem state or real registry keys are flaky and unreliable. Must use temp dirs and mock OS boundaries. | LOW | Already partially done in existing tests. Audit remaining tests for isolation. |

### Differentiators (Competitive Advantage)

Features that make MoveKit stand out as a quality-first tool compared to FolderMove, FreeMove, and other C-drive cleanup utilities.

| # | Feature | Value Proposition | Complexity | Notes |
|---|---------|-------------------|------------|-------|
| D1 | **Frontend Component Tests (Vitest + React Testing Library)** | Competitors like FolderMove have zero frontend tests. MoveKit having test coverage for UI components demonstrates engineering maturity — users who check GitHub see a professional project. | HIGH | 12 React components + error boundary. Need Vitest + jsdom setup. Mock Wails runtime bindings. First-time setup cost is significant but one-time. |
| D2 | **Go-Frontend Binding Integration Tests** | The Wails bridge (Go methods called from JS) is a critical integration point. Testing it directly catches binding mismatches before runtime. Competitors don't do this. | HIGH | Can test via `wails dev` + Playwright pointing at `http://localhost:34115`. Or create a Go test harness that exercises the bound methods directly. |
| D3 | **Syscall Interface Abstraction for Testability** | MoveKit's heavy Win32 API usage (junction, registry, shell, volume, power, process) makes testing hard. Abstracting behind Go interfaces lets tests inject mock OS behavior — enabling thorough testing of edge cases competitors can't test. | MEDIUM | Apply "Clock interface" pattern to syscall layer. Create `Syscaller`, `RegistryAccess`, `VolumeChecker` interfaces. This directly helps fix BUG-01 (unsafe type assertion in process.go). |
| D4 | **Coverage Enforcement as CI Gate** | "80% coverage" is a target; a CI gate makes it non-negotiable. Every PR that drops coverage is rejected. This institutionalizes quality — competitors don't have this discipline. | LOW | Add coverage threshold check in GitHub Actions. `go tool cover -func` + awk to enforce 80% minimum. |
| D5 | **Migration Scenario Fixture Library** | Test factories that generate realistic migration scenarios: "user with OneDrive KFM enabled", "user with Google Drive syncing Desktop", "user on laptop with 5% battery". Reusable test data = faster test writing = more edge cases covered. | MEDIUM | Create `internal/migrate/testutil/` package with fixture functions. Uses temp dirs, fake registry data. |
| D6 | **Frontend TypeScript Strict Mode** | TypeScript strict null checks catch the frontend equivalent of nil-pointer panics. Enabling `strict: true` in tsconfig and fixing existing violations prevents a whole class of bugs. | LOW | Currently `tsconfig.json` exists but unknown strictness level. Enable `strict: true`, `noUncheckedIndexedAccess`, fix resulting errors. |
| D7 | **Smoke Test for Key User Workflow** | One automated test that runs the complete happy path: "select folders -> pick target disk -> execute migration -> verify junctions -> rollback". Runs in CI on every push. Catches integration-level regressions instantly. | MEDIUM | Runner requires Windows (NTFS + WebView2). Use GitHub Actions `windows-latest` runner. Test exercises Go methods directly (not through UI). |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem like good ideas for a quality iteration but would increase complexity without proportional value at this stage.

| # | Anti-Feature | Why Requested | Why Problematic | Alternative |
|---|-------------|---------------|-----------------|-------------|
| A1 | **E2E Testing with Playwright/Cypress** | "Test the whole app end-to-end like a real user." | Only works in `wails dev` mode on `localhost:34115`. Cannot test native behaviors (file dialogs, window management, WebView2 quirks). Production build is completely untestable this way. High maintenance burden — UI selectors break on any component change. | Integration tests at the Go binding layer (D2) give 90% of the value with 10% of the maintenance cost. Reserve true E2E for a future phase when the UI has stabilized. |
| A2 | **Visual Regression Testing (Screenshot Comparison)** | "Make sure the UI doesn't break visually." | Extremely brittle. Different screen resolutions, DPI scaling, WebView2 versions produce different pixel output. Every TailwindCSS change triggers false positives. High flake rate erodes trust in CI. | Manual visual review during PR. If visual consistency becomes a recurring problem, reconsider in v0.3+. |
| A3 | **Performance Benchmarking Infrastructure** | "Track if migration gets slower over time." | Premature optimization. Current migration is file-copy-bound (OS-level bottleneck), not code-bound. Performance is acceptable until user feedback says otherwise. Benchmarking infrastructure is maintenance overhead with no current payoff. | Ad-hoc `go test -bench` when performance questions arise. |
| A4 | **Full MSIX/Store Certification Testing** | "Be ready for Microsoft Store distribution." | MoveKit is a portable exe by design decision (see PROJECT.md Out of Scope). Store certification tests (App Capabilities, UAC, Package Compliance, Branding) are irrelevant for portable exe distribution. | Not applicable. Revisit only if distribution strategy changes. |
| A5 | **Cross-Platform CI (macOS/Linux Runners)** | "Make sure the code is clean everywhere." | MoveKit deeply depends on Win32 API (`golang.org/x/sys/windows`, WebView2). Most Go files won't even compile on non-Windows. Running CI on ubuntu/macOS would require build tags and conditional compilation that adds complexity without value. | All CI on `windows-latest` runner. Cross-platform is explicitly out of scope. |
| A6 | **Fuzz Testing (Go Fuzzing)** | "Find edge cases automatically." | Fuzzing is powerful but requires: (1) understanding what inputs matter, (2) writing fuzz targets, (3) triaging findings. Doing this before basic unit coverage exists is backwards. Fuzz testing is a multiplier — useless when the base is zero. | Establish 80% unit coverage first (T1). Then identify the 2-3 most critical parsing/serialization functions for fuzz targets in a later phase. |
| A7 | **Test Data Factories Using Real User Data** | "Tests should use real folder structures from real users." | Privacy risk (user folder names contain PII). Adds complexity to manage test data. Tests become coupled to specific data shapes that change over time. | Synthetic test fixtures using generated data (D5). Real-world edge cases can be anonymized and added as specific test cases. |
| A8 | **Automated Accessibility Testing (axe-core)** | "Make the app accessible to screen readers." | MoveKit targets Chinese-speaking Windows users. Accessibility is important but: (1) no user demand yet, (2) WebView2 accessibility support varies by Windows version, (3) adds CI complexity. This is a v1.x concern. | Manual keyboard-navigation check during PR review. Add axe-core in a future accessibility-focused phase. |

---

## Feature Dependencies

```
[T3: Regression Tests] ──requires──> [Bug fixes exist]
                                           └──requires──> [T5: Error Handling Audit] (BUG-03, BUG-05, BUG-06, BUG-08)

[T1: Go Unit Coverage] ──enabled by──> [D3: Syscall Abstraction] (makes untestable code testable)
[T1: Go Unit Coverage] ──enabled by──> [D5: Fixture Library] (speeds up test writing)

[T2: CI/CD Pipeline] ──contains──> [T4: Static Analysis]
[T2: CI/CD Pipeline] ──contains──> [T6: Build Verification]
[T2: CI/CD Pipeline] ──contains──> [T7: Formatting Standards]
[T2: CI/CD Pipeline] ──contains──> [D4: Coverage Gate]

[D1: Frontend Tests] ──requires──> [D6: TypeScript Strict Mode] (catches bugs test setup would mask)
[D1: Frontend Tests] ──independent of──> [Go tests] (can be done in parallel)

[D2: Binding Integration Tests] ──builds on──> [T1: Go Coverage] (needs backend to be reliable first)

[D7: Smoke Test] ──requires──> [T1 + D1 + D2] (integration test needs all layers tested)
```

### Dependency Notes

- **T3 requires bug fixes first:** Regression tests prove the fix works. Without fixed code, the test just documents the bug. Fix-then-test order per bug.
- **D3 enables T1:** Files like `process.go`, `shell.go`, `volume.go` are currently untestable because they call Win32 APIs directly. Syscall abstraction is prerequisite for testing these files.
- **D1 and Go tests are independent:** Frontend and backend test work can proceed in parallel. Different languages, different tools, different developers.
- **D2 is deferred until T1 stabilizes:** Integration tests are flaky if the backend has bugs. Get backend reliable first, then test the bridge.
- **D7 is the capstone:** Smoke test validates the whole system works together. Ships last in the milestone.

---

## MVP Definition (This Milestone: v0.1.0 Quality)

Per PROJECT.md, this is a greenfield-quality milestone: bug fixing + test coverage improvement targeting 80%+. No new features.

### Must Ship (P1)

- [ ] **T3: Regression Tests for 8 Known Bugs** — Each bug fix accompanied by a test that reproduces the failure. Non-negotiable. Without this, bugs recur.
- [ ] **T1: Go Unit Test Coverage >= 80%** — 15 files without tests must get coverage. Core paths (migrate, rollback, junction, registry) covered first.
- [ ] **T2: CI/CD Pipeline** — GitHub Actions on `windows-latest` runner. Runs `go test -race`, `golangci-lint`, `go vet`, `wails build` on every push.
- [ ] **T4: Static Analysis (golangci-lint)** — `.golangci.yml` with errcheck, govet, staticcheck, gosec, unused. Zero warnings in CI.
- [ ] **T5: Error Handling Audit** — Every silent error swallow (`_ = err`, unchecked returns) audited and fixed. Bugs BUG-03/05/06/08 are symptoms of this.
- [ ] **T8: Test Isolation** — All existing and new tests use temp dirs. No test touches real registry or filesystem.

### Should Ship If Time Allows (P2)

- [ ] **D3: Syscall Interface Abstraction** — Makes process.go, shell.go, volume.go testable. Enables coverage on currently untestable files.
- [ ] **D4: Coverage Enforcement Gate** — CI rejects PRs that drop coverage below 80%.
- [ ] **D5: Migration Scenario Fixtures** — Reusable test data in `internal/migrate/testutil/`.
- [ ] **D6: TypeScript Strict Mode** — `strict: true` in tsconfig, fix resulting errors.
- [ ] **D1: Frontend Component Tests** — Vitest + React Testing Library for 12 components. Mock Wails bindings.

### Defer to Future (P3)

- [ ] **D2: Binding Integration Tests** — Valuable but complex. Do after Go backend coverage is solid.
- [ ] **D7: Smoke Test** — Full migration+rollback workflow test. Capstone of the quality effort.
- [ ] **T7: Frontend Formatting (Prettier/ESLint)** — Add when frontend tests exist to validate against.

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| T3: Regression Tests for 8 Bugs | HIGH (prevents recurrence) | LOW (8 focused tests) | P1 |
| T2: CI/CD Pipeline | HIGH (catches regressions automatically) | MEDIUM (setup + windows runner) | P1 |
| T4: Static Analysis (golangci-lint) | HIGH (prevents bug classes) | LOW (config + fix existing) | P1 |
| T1: Go Unit Coverage >= 80% | HIGH (core correctness) | HIGH (15 files, many need abstraction first) | P1 |
| T5: Error Handling Audit | HIGH (fixes 4 bugs at root) | LOW (systematic review) | P1 |
| T8: Test Isolation | MEDIUM (prevents flaky tests) | LOW (audit existing) | P1 |
| D3: Syscall Abstraction | MEDIUM (enables testing) | MEDIUM (interface design + refactor) | P2 |
| D4: Coverage Gate in CI | MEDIUM (institutionalizes quality) | LOW (CI config) | P2 |
| D5: Fixture Library | MEDIUM (speeds up test writing) | MEDIUM (design + implementation) | P2 |
| D6: TypeScript Strict Mode | MEDIUM (prevents frontend bugs) | LOW (config + fix errors) | P2 |
| D1: Frontend Component Tests | MEDIUM (catches UI regressions) | HIGH (setup + 12 components) | P2 |
| D2: Binding Integration Tests | LOW (catches bridge bugs) | HIGH (requires running Wails) | P3 |
| D7: Smoke Test | MEDIUM (catches integration regressions) | MEDIUM (workflow script) | P3 |
| T7: Frontend Formatting | LOW (cosmetic) | LOW (config) | P3 |

---

## Competitor Feature Analysis (Quality Dimension)

| Quality Feature | FolderMove / FreeMove | Sifty (MIT) | Fluent Cleaner | MoveKit Target |
|-----------------|----------------------|-------------|----------------|----------------|
| Automated tests | None visible | Go tests present | Unknown (C#/WinUI) | 80%+ coverage, CI gate |
| CI/CD pipeline | None | GitHub Actions | Unknown | GitHub Actions, windows-latest |
| Error handling | Basic (C# try/catch) | Structured (Go errors) | Standard WinUI | Structured error codes + 中文 messages |
| Rollback/undo | No (manual reverse) | Yes (recycle bin + sifty undo) | No (permanent delete) | Yes (backup record + junction removal) |
| Dry-run / preview | No | Yes (default dry-run) | Shows before delete | Pre-flight checks (disk, NTFS, cloud, battery) |
| Open source | Yes (C#) | Yes (Go, MIT) | Yes (C#, MIT) | Yes (Go, MIT) |
| Test isolation | N/A | Go temp dirs | Unknown | Temp dirs + syscall mocks |
| Frontend tests | None | N/A (TUI) | Unknown | Vitest + RTL (target) |

**Key insight:** Among C-drive tools, Sifty is the quality leader (dry-run default, undo, audit logs, Go tests, CI). MoveKit's differentiation is NTFS Junction migration (not deletion), built-in rollback, and cloud conflict detection. Quality-wise, MoveKit should match Sifty's testing rigor while adding the frontend testing dimension Sifty lacks (TUI vs GUI).

---

## Phase Implications

### Recommended Phase Order for This Milestone

1. **Phase: Foundation (CI + Lint + Fix Bugs)** — T2, T4, T5, T3
   - Set up CI first so every subsequent phase gets automated verification
   - Add linting to catch errors mechanically
   - Audit error handling (root cause of 4 bugs)
   - Fix bugs with regression tests (one commit per bug)
   - **Rationale:** No point writing new tests if CI doesn't exist and bugs are unfixed

2. **Phase: Go Coverage** — D3, D5, T1, T8, D4
   - Abstract syscall interfaces (prerequisite for testing untestable files)
   - Build fixture library (accelerates test writing)
   - Write tests for all 15 uncovered Go files
   - Audit test isolation
   - Add coverage enforcement gate
   - **Rationale:** Backend is the foundation. Frontend tests are meaningless if the backend is unreliable.

3. **Phase: Frontend Quality** — D6, D1, T7
   - Enable TypeScript strict mode
   - Set up Vitest + React Testing Library
   - Write component tests for 12 React components
   - Add Prettier + ESLint
   - **Rationale:** Frontend is independent of backend testing. Can overlap Phase 2 if resources allow.

4. **Phase: Integration Verification** — D2, D7
   - Binding integration tests
   - Full smoke test (happy path migration + rollback)
   - **Rationale:** Capstone. Validates that Phases 1-3 produced a working whole.

### Research Flags for Future Phases

- **Phase 2 (Go Coverage):** `process.go` and `shell.go` testing may need deeper Windows API research — specifically how to mock `syscall.CreateToolhelp32Snapshot` and `SHChangeNotify`.
- **Phase 3 (Frontend Tests):** Wails runtime mocking (`window.go.main.App.*`) needs investigation — the auto-generated `wailsjs/` bindings may need a mock layer.
- **Phase 4 (Integration):** GitHub Actions `windows-latest` runner availability and WebView2 pre-installation status needs verification.

---

## Sources

- [Wails v2 E2E Testing Discussion #4205](https://github.com/wailsapp/wails/discussions/4205) — Playwright/Cypress feasibility, dev mode testing at localhost:34115
- [Go Testing Patterns: Table-Driven Tests + Syscall Mocking](https://datasea.cn/go0215490221.html) — Boundary coverage matrix, interface abstraction for OS calls
- [Go CI/CD Pipeline Standards](https://datasea.cn/go0223506072.html) — golangci-lint, coverage gates, GitHub Actions patterns
- [Production Desktop Application QA Guide](https://www.testriq.com/blog/post/what-tests-are-standard-in-desktop-application-testing) — Test categories for production readiness
- [Desktop Application Testing Best Practices](https://www.ironorbit.com/desktop-application-testing-guide/) — Compatibility, performance, security testing requirements
- [Go Standard Library Windows Tests](https://go.googlesource.com/sys/+/2f1e207ee39ff70f3433e49c6eb52677a515e3b5/windows/syscall_test.go) — Real syscall test patterns
- [Sifty (MIT) — C-drive maintenance tool](https://github.com/Vortrix5/sifty) — Competitor with strongest quality practices: dry-run default, undo, audit logs
- [Fluent Cleaner](https://www.XDA-Developers.com/ditched-pc-manager-for-fluent-cleaner-after-it-found-more-junk/) — WinUI 3 cleanup tool, Winapp2.ini integration
- MoveKit PROJECT.md — Scope constraints, 8 known bugs, 18/26 Go files without tests, zero frontend tests

---
*Feature research for: MoveKit testing & quality (greenfield-quality milestone)*
*Researched: 2026-06-21*
