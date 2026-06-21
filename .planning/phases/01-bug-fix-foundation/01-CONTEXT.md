# Phase 1: Bug Fix Foundation - Context

**Gathered:** 2026-06-21
**Status:** Ready for planning

<domain>
## Phase Boundary

Fix all 8 known bugs from CONCERNS.md with regression tests. Conduct targeted error handling audit on bug-related functions and known silent-failure sites. Apply 4 quality improvements (error audit, simulateProgress replacement, cancel wiring, ResolveConflict stub). Same-drive migration optimization. No new features — fix only what's broken and make the tool slightly better where it costs nothing.
</domain>

<decisions>
## Implementation Decisions

### Bug Fix Strategy
- **Ordering:** Severity-first — panic/crash bugs first (BUG-01 unsafe assertion, BUG-03 backup save, BUG-04 same-drive), then silent failures (BUG-05 dirSize, BUG-06 shell notify, BUG-08 security software), then dead code (BUG-02 junction debug buffer, BUG-07 percent)
- **Test approach:** Fix-then-test — modify code to eliminate the bug, then write a regression test that confirms the fix and prevents recurrence
- **Commit granularity:** One bug per commit — each commit contains the fix + its regression test

### Error Handling Audit
- **Scope:** Targeted — audit functions touched by bug fixes plus known silent-failure sites from CONCERNS.md. Not a full codebase scan.
- **Intentional ignores policy:** Log + comment — `if err != nil { log.Warn("..."); // intentional: non-critical }` — never silently swallow without explanation
- **Output:** Inline source changes only — no separate report file. Fixes ship as part of bug fix commits.

### Key Improvement Decisions
- **BUG-07 percent dead code:** Delete — percent is already calculated correctly in `migrate.go`; the `move.go:137` variable is just unused
- **IMP-02 simulateProgress:** Replace with Wails event listener — `useEffect` + `runtime.EventsOn("migration:progress")` drives real progress from backend events
- **IMP-04 ResolveConflict:** Return `MoveKitError{Code: "NOT_IMPLEMENTED"}` — explicit error instead of silently returning nil
- **BUG-08 security software:** PreCheck Warning field — detection results appear in warning output, informational only (don't block migration)

### Claude's Discretion
- All implementation choices not covered above are at Claude's discretion — use codebase conventions, ROADMAP success criteria, and safety judgment
</decisions>

<code_context>
## Existing Code Insights

### Files to Modify (from CONCERNS.md)
| Bug | File | Issue |
|-----|------|-------|
| BUG-01 | `internal/syscall/process.go:43` | Unsafe type assertion — panic risk |
| BUG-02 | `internal/syscall/junction.go` | Debug buffer size mismatch |
| BUG-03 | `internal/migrate/backup.go` | Save errors swallowed |
| BUG-04 | `internal/migrate/move.go` | Same-volume fallback missing |
| BUG-05 | `app.go:211` (dirSize) | Errors silently ignored |
| BUG-06 | `internal/syscall/shell.go` | UTF16 encoding errors ignored |
| BUG-07 | `internal/migrate/move.go:137` | Percent dead code |
| BUG-08 | `app.go` / `internal/migrate/check.go` | Security software detection discarded |

### Improvements
| Item | File | Change |
|------|------|--------|
| IMP-01 | `app.go`, `internal/syscall/`, `internal/migrate/` | Error handling audit (targeted) |
| IMP-02 | `frontend/src/App.tsx:139-181` | Replace simulateProgress with Wails event listener |
| IMP-03 | `frontend/src/components/` + `app.go` | Wire cancel button to CancelMigration() |
| IMP-04 | `app.go:513` | ResolveConflict error return |

### Established Patterns
- **MoveKitError** — structured error: Code + Message (Chinese) + Detail + Action
- **Test pattern** — `t.TempDir()` for isolation, `t.Cleanup()` for junction cleanup
- **Logging** — `migrate.Logger` with Info/Warn/Error levels, path masking
- **Cancellation** — atomic.Int32 flag checked during file iteration
</code_context>

<specifics>
## Specific Ideas

No specific visual or behavioral references — this is a code-quality phase. Follow existing codebase patterns and conventions.
</specifics>

<deferred>
## Deferred Ideas

- Full codebase error handling audit — Phase 3 (Go Backend Test Coverage) will cover this more thoroughly
- Interactive conflict resolution UI — would be a new capability, future phase
- Real percent calculation for cross-volume moves — requires deeper progress tracking refactor
- Security software blocking — UX decision for future phase: when/if to block vs warn
</deferred>
