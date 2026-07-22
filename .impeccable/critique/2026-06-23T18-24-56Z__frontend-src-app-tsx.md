---
target: MoveKit frontend app
total_score: 32
p0_count: 0
p1_count: 3
timestamp: 2026-06-23T18-24-56Z
slug: frontend-src-app-tsx
---
## Design Health Score

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 3 | Migration progress excellent, but rollback has no feedback |
| 2 | Match Between System and Real World | 4 | Chinese clear, Windows native feel, FAQ explains terms |
| 3 | User Control and Freedom | 4 | Cancel, rollback, skip, Esc coverage complete |
| 4 | Consistency and Standards | 4 | Component style, color, spacing highly consistent |
| 5 | Error Prevention | 3 | OneDrive/cloud/NTFS checks excellent, missing disk space precheck |
| 6 | Recognition Rather Than Recall | 3 | Status badges clear, but header icon buttons lack text labels |
| 7 | Flexibility and Efficiency of Use | 2 | Single flow path, no shortcuts, no batch operations |
| 8 | Aesthetic and Minimalist Design | 4 | Clean, focused, no decorative noise |
| 9 | Error Recovery | 3 | Rollback available, but error messages show raw exception text |
| 10 | Help and Documentation | 3 | FAQ + first-run guide + defender info, good coverage |
| **Total** | | **32/40** | **Good** |

## Anti-Patterns Verdict

**Not AI-generated.** No gradient text, no hero-metric template, no identical card grids, no tiny uppercase eyebrows. Windows Fluent Design is an intentional brand choice. Colors are restrained. Only minor signal: 5 modals use `backdrop-blur-sm`.

**Deterministic scan**: 0 findings. Code is clean.

## What's Working

1. **Error prevention check chain**: OneDrive KFM → cloud sync → NTFS check → disk type filter → confirm dialog. Intercepts almost all foreseeable issues before dangerous operations.
2. **Emotional curve design**: Anxiety → reassurance → confirmation → execution → release. Each phase has corresponding emotional support.
3. **Component consistency**: All cards 8px radius + 1px border, all buttons unified, all dialogs follow Header → Body → Footer.

## Priority Issues

**[P1] DiskSelector missing ARIA radio roles**: Custom radio group has no `role="radiogroup"` and `role="radio"`. Screen readers cannot identify this as a radio group. WCAG 4.1.2 violation.

**[P1] HelpPanel GitHub URL points to wrong repository**: Links point to `nousresearch/movekit` instead of `hyls9527/MoveKit`. Users clicking feedback links will land on wrong repo.

**[P1] Rollback operation has no loading feedback**: Button immediately restores after click with no loading state. User doesn't know if rollback is executing.

**[P2] Error messages show raw exception text**: `catch (err) { setError(err.message) }` exposes technical details. Normal users seeing "syscall access denied" will be confused.

**[P2] Modal backdrop-blur-sm violates DESIGN.md**: 5 modals use glassmorphism effect, DESIGN.md explicitly prohibits it.

## Persona Red Flags

**Alex (Power User)**: No keyboard shortcuts. Forced confirm dialog. Cannot select individual folders. High abandonment risk for experienced users.

**Jordan (First-Timer)**: Header icon buttons rely on emoji only. Migration warning doesn't explain why. FirstRunGuide is good.

**Sam (Accessibility)**: DiskSelector cannot be operated by screen readers (P1). Header icon buttons missing aria-label. Color contrast passes.

## Minor Observations

- `formatBytes` duplicated across 4 components
- HelpPanel uses inline style `height: calc(100% - 120px)` instead of Tailwind
- Confirm dialog ⚠️ emoji renders inconsistently across systems
- Missing `prefers-reduced-motion` support
