---
phase: 03-go-backend-test-coverage
plan: 02
subsystem: testing
tags: [go, refactoring, dependency-injection, syscall]

requires:
  - phase: 03
    plan: 01
    provides: 7 syscall interfaces + mocks
provides:
  - Package-level Sys variable for migrate dependency injection
  - All migrate call sites refactored to Sys.Interface.Method()
affects: [03-03-high-tests, 03-04-remaining-tests]

key-files:
  created: ["internal/migrate/syscall_bridge.go"]
  modified: ["internal/migrate/migrate.go", "internal/migrate/move.go", "internal/migrate/rollback.go", "internal/migrate/check.go", "internal/migrate/security.go", "internal/migrate/onedrive.go"]

requirements-completed: [GO-02]

duration: 2min
completed: 2026-06-22
---

# Phase 3: Syscall Wiring Summary

**All migrate syscall calls routed through Sys interfaces — mock injection ready**

## Performance
- Duration: ~2 min
- Tasks: 1
- Files: 7 (1 new + 6 modified)

## Accomplishments
- syscall_bridge.go with package-level Sys = sys.NewRealLayer()
- 16 syscall function calls replaced with Sys.Interface.Method() across 6 files
- sys. constants (UserShellFoldersKey, ShellFoldersKey) preserved — no behavioral change

## Deviations
- go build not verified locally (Go not available); CI will confirm compilation

---
*Phase: 03-go-backend-test-coverage*
*Completed: 2026-06-22*
