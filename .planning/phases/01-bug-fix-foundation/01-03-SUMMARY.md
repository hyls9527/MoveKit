---
phase: 01-bug-fix-foundation
plan: 03
subsystem: app
tags: [bug-fix, app, dirSize, resolve-conflict, security-software]
requires: [BUG-05, BUG-08, IMP-04, IMP-01]
provides: [BUG-05-fix, BUG-08-fix, IMP-04-fix, IMP-01-app-audit]
affects: [app.go, app_types.go, internal/migrate/check.go, internal/migrate/errors.go, internal/migrate/migrate.go]
tech-stack:
  added: []
  patterns: [sentinel-value, non-blocking-warning, structured-error-extension]
key-files:
  created:
    - app_test.go
  modified:
    - app.go
    - app_types.go
    - internal/migrate/check.go
    - internal/migrate/errors.go
    - internal/migrate/migrate.go
    - internal/migrate/check_test.go
requirements-completed: [BUG-05, BUG-08, IMP-04, IMP-01]
key-decisions:
  - "BUG-05: dirSize error sets SizeBytes = -1 sentinel instead of silently 0"
  - "IMP-04: ResolveConflict returns NOT_IMPLEMENTED MoveKitError instead of nil"
  - "BUG-08: Security software detection flows PreCheck -> MoveKitError.Warning -> migration:done event"
  - "BUG-08: SECURITY_SOFTWARE_DETECTED is non-blocking — Migrate() logs and continues"
  - "IMP-01 PreCheck audit: zero silent error ignores across all 11 error checks"
duration: 0 min
completed: 2026-06-20
---

# Phase 01 Plan 03: App-Layer Fixes Summary

Fixed three app-layer issues: dirSize errors now propagate as SizeBytes=-1 sentinel, ResolveConflict returns explicit error, and security software detection surfaces through PreCheck Warning field. Error audit completed on GetFolders, EstimateDuration, and PreCheck.

**Substantive changes:**
- GetFolders: dirSize error sets SizeBytes = -1 instead of silently defaulting to 0
- ResolveConflict: returns MoveKitError NOT_IMPLEMENTED instead of nil
- MoveKitError gains Warning field for non-blocking informational warnings
- PreCheck: security software detection captured and returned as SECURITY_SOFTWARE_DETECTED
- Migrate(): SECURITY_SOFTWARE_DETECTED treated as non-blocking (logs and continues)
- migration:done event includes warning field from MoveKitError.Warning

## Tasks Completed

| Task | Name | Commits | Files |
|------|------|---------|-------|
| 1 | Fix BUG-05 + IMP-04 — dirSize error + ResolveConflict | d13358c, 33c3c94 | app.go, app_types.go, app_test.go |
| 2 | Fix BUG-08 — security software detection | b697cc6, c19e775 | check.go, errors.go, migrate.go, app.go, app_types.go, check_test.go |

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written.

### Verification Notes

`go test` and `go build` could not be executed in this environment (Go not found in PATH). All code changes follow the plan specifications exactly:
- app.go: SizeBytes = -1 on dirSize error
- app.go: ResolveConflict returns *migrate.MoveKitError with Code "NOT_IMPLEMENTED"
- errors.go: Warning field added to MoveKitError struct
- check.go: detectedAV captured, SECURITY_SOFTWARE_DETECTED returned
- migrate.go: SECURITY_SOFTWARE_DETECTED treated as non-blocking
- PreCheck IMP-01 audit: all 11 error paths verified as checked or documented

## Known Stubs

None — all changes are functional fixes, not stubs.

## Next

Ready for Plan 01-04 (Frontend Progress + Cancel Wiring).
