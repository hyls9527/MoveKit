# Roadmap: MoveKit Quality Hardening

## Overview

MoveKit is entering a quality-focused milestone. No new user-facing features — instead, we fix 8 known bugs, establish CI/CD automation, build comprehensive test coverage (targeting 80%+ Go, component-level frontend), and verify end-to-end migration integrity. Each phase delivers an independently verifiable quality improvement. Phases follow a strict dependency order: bug fixes first (testing broken code is waste), then CI (every change verified), then backend coverage (foundation), then frontend quality, and finally integration verification.

## Phases

- [x] **Phase 1: Bug Fix Foundation** - Resolve all 8 known bugs with regression tests; complete error handling audit and small improvements (completed 2026-06-20)
- [ ] **Phase 2: CI/CD Infrastructure** - GitHub Actions workflow on windows-latest with lint, test, race detection, and build verification
- [ ] **Phase 3: Go Backend Test Coverage** - Syscall interface abstraction, 80%+ coverage with enforcement gate, zero real-HKCU tests
- [ ] **Phase 4: Frontend Quality** - TypeScript strict mode, Vitest + RTL infrastructure, component and state machine tests
- [ ] **Phase 5: Integration & Verification** - Full Migrate+Rollback roundtrip with real NTFS junctions in isolated environments

## Phase Details

### Phase 1: Bug Fix Foundation
**Goal**: All 8 known bugs are fixed with regression tests; critical quality improvements (error handling audit, same-drive optimization, cancel wiring, conflict resolution) are in place.
**Depends on**: Nothing (first phase)
**Requirements**: BUG-01, BUG-02, BUG-03, BUG-04, BUG-05, BUG-06, BUG-07, BUG-08, IMP-01, IMP-02, IMP-03, IMP-04
**Success Criteria** (what must be TRUE):
  1. All 8 documented bugs are fixed, each with a regression test that reproduces the failure before the fix and passes after
  2. Error handling audit complete — zero silently-ignored error return values across the entire Go codebase
  3. Same-drive folder migration completes near-instantly via `os.Rename` instead of copy-and-delete
  4. Frontend progress bar reflects real Wails backend events (not timer-based simulation), and cancel button stops an in-progress migration with observable effect
  5. `ResolveConflict` surfaces errors to the caller instead of silently returning nil
**Plans**: 4 plans in 2 waves

Plans:
- [x] 01-01-PLAN.md — Fix BUG-01 (panic type assertion), BUG-02 (junction buffer alignment), BUG-06 (UTF16 error logging) in syscall layer + regression tests
- [x] 01-02-PLAN.md — Fix BUG-03 (backup save errors propagated), BUG-04 (same-drive rename tested), BUG-07 (percent dead code removed) in migrate layer + regression tests
- [x] 01-03-PLAN.md — Fix BUG-05 (dirSize error propagation), BUG-08 (security software detection surfaced), IMP-04 (ResolveConflict error) in app layer + regression tests
- [x] 01-04-PLAN.md — Replace simulateProgress with real Wails events (IMP-02), wire cancel button to CancelMigration (IMP-03)

### Phase 2: CI/CD Infrastructure
**Goal**: Every code change is automatically verified by CI on every push — lint, test with race detection, and production build all pass.
**Depends on**: Phase 1
**Requirements**: CI-01, CI-02, CI-03, CI-04
**Success Criteria** (what must be TRUE):
  1. GitHub Actions workflow runs on every push and completes successfully on `windows-latest` runner
  2. `golangci-lint` passes with zero warnings across all Go packages
  3. `go test -race -cover` runs on every push and produces a human-readable coverage report
  4. `wails build` succeeds on the CI runner, confirming the project is always in a shippable state
**Plans**: TBD

### Phase 3: Go Backend Test Coverage
**Goal**: Go backend has comprehensive test coverage (>=80% overall, >=85% for migrate package) with clean syscall abstractions; no test touches real HKCU registry or leaks NTFS junctions.
**Depends on**: Phase 1, Phase 2
**Requirements**: GO-01, GO-02, GO-03, GO-04, GO-05
**Success Criteria** (what must be TRUE):
  1. Syscall interface abstraction layer (RegistryAccessor, ProcessChecker, VolumeChecker, and others) is in place with uber-go/mock-generated mocks
  2. Every Go source file in `internal/syscall/` and `internal/migrate/` has a corresponding `_test.go` file with passing tests
  3. All 17 Wails binding methods in `app.go` have test coverage using mocked dependencies
  4. CI rejects any pull request that drops overall coverage below 80% or migrate package coverage below 85%
  5. Zero tests touch the real HKCU registry or leave orphaned NTFS junctions — all tests use TempDir and in-memory fakes
**Plans**: TBD

### Phase 4: Frontend Quality
**Goal**: Frontend is type-safe (TypeScript strict mode) with component-level test coverage; Wails mock factory enables independent frontend testing.
**Depends on**: Phase 2 (can overlap Phase 3)
**Requirements**: FE-01, FE-02, FE-03, FE-04, FE-05
**Success Criteria** (what must be TRUE):
  1. TypeScript strict mode is enabled and the project builds with zero type errors
  2. All 12 React components render without crashing in Vitest + React Testing Library
  3. App.tsx state machine transitions (idle -> scanning -> migrating -> complete) are verified by automated tests
  4. `lib/wails.ts` mock factory can inject arbitrary backend responses, enabling any component test scenario
**Plans**: TBD

**UI hint**: yes

### Phase 5: Integration & Verification
**Goal**: End-to-end migration and rollback workflows are verified in realistic but safe environments; the complete user journey is confirmed working.
**Depends on**: Phase 3, Phase 4
**Requirements**: INT-01, INT-02, INT-03, INT-04
**Success Criteria** (what must be TRUE):
  1. Full `Migrate()` flow runs against real NTFS junctions in an isolated TempDir and completes without errors
  2. Full `Rollback()` restores original state (files, junctions, registry) after migration in an isolated test environment
  3. Wails binding layer smoke test confirms the Go-to-JS bridge works correctly end-to-end
  4. A complete user journey (migrate -> verify contents -> rollback -> verify restoration) passes as a manual or automated smoke test
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Bug Fix Foundation | 0/? | Not started | - |
| 2. CI/CD Infrastructure | 0/? | Not started | - |
| 3. Go Backend Test Coverage | 0/? | Not started | - |
| 4. Frontend Quality | 0/? | Not started | - |
| 5. Integration & Verification | 0/? | Not started | - |
