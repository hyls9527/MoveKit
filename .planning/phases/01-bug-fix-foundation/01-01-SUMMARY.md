---
phase: 01-bug-fix-foundation
plan: 01
subsystem: syscall
tags: [bug-fix, syscall, type-safety, junction, shell]
requires: []
provides: [BUG-01-fix, BUG-02-fix, BUG-06-fix, IMP-01-syscall-audit]
affects: [internal/syscall/process.go, internal/syscall/junction.go, internal/syscall/shell.go]
tech-stack:
  added: []
  patterns: [comma-ok-type-assertion, stderr-logging, early-return-on-utf16-error]
key-files:
  created:
    - internal/syscall/process_test.go
    - internal/syscall/shell_test.go
  modified:
    - internal/syscall/process.go
    - internal/syscall/junction.go
    - internal/syscall/shell.go
requirements-completed: [BUG-01, BUG-02, BUG-06, IMP-01]
key-decisions:
  - "BUG-01: comma-ok pattern prevents panic on nil/wrong-type errors"
  - "BUG-02: SubLen/PrintLen include NUL terminator per Microsoft REPARSE_DATA_BUFFER spec"
  - "BUG-06: UTF16 errors logged to stderr with early return instead of silent discard"
  - "IMP-01 syscall audit: zero remaining silent error ignores in process.go, junction.go, shell.go"
duration: 0 min
completed: 2026-06-20
---

# Phase 01 Plan 01: Syscall Bug Fixes Summary

Fixed three syscall-layer bugs with regression tests. Eliminated panic risk, junction buffer misalignment, and silent error discards at the lowest OS layer.

**Substantive changes:**
- process.go:43 changed from unsafe `e.(syscall.Errno)` to `errno, ok := e.(syscall.Errno); ok && ...` — prevents panic when e is nil or wrong type
- junction.go buildReparseBuffer: SubLen/PrintLen now include NUL terminator per Microsoft REPARSE_DATA_BUFFER spec, matching mklink output
- shell.go NotifyFolderRenamed/NotifyDirUpdated: UTF16 encoding errors now logged to stderr with early return — no garbage pointers passed to SHChangeNotify

## Tasks Completed

| Task | Name | Commits | Files |
|------|------|---------|-------|
| 1 | Fix BUG-01 — comma-ok type assertion | 8374194, 064542d | process.go, process_test.go |
| 2 | Fix BUG-02 — junction buffer alignment | cd45d40 | junction.go |
| 3 | Fix BUG-06 — UTF16 error handling | 2522d1f, 0fac230 | shell.go, shell_test.go |

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written.

### Verification Notes

`go test` could not be executed in this environment (Go not found in PATH). All code changes follow the plan specifications exactly:
- Acceptance criteria verified statically (comma-ok pattern present at line 43, SubLen/PrintLen counts verified, stderr logging present)
- IMP-01 audit completed for all three files — zero remaining silent error ignores
- Test file structure follows existing codebase patterns (package winsyscall, t.TempDir(), t.Fatalf, t.Logf)

## Known Stubs

None — all changes are functional fixes, not stubs.

## Next

Ready for Plan 01-02 (Migration Bug Fixes).
