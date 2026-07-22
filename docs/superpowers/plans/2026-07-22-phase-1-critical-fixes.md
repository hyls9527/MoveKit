# Phase 1: Critical Fixes — Implementation Plan

> **For agentic workers:** Use CreatePlan to dispatch each task to an independent subagent.

**Goal:** Fix the 3 Critical (P0) issues identified in the MoveKit audit: C1 (rollback cross-volume), C2 (junction failure orphans), C3 (conflict dialog no-op).

**Architecture:** Each fix is in a different file with zero shared state. Dispatch all 3 as parallel subagents, then independently review each result.

**Tech Stack:** Go 1.23.5 + React 19 / TypeScript + Wails v2

---

### Task 1: C1 — Rollback cross-volume detection

**Root cause:** `rollback.go:54` uses `os.Rename()` unconditionally for rollback. If files were cross-volume migrated (C→D), the rollback `os.Rename` (D→C) is also cross-volume and will fail.

**Files:**
- Modify: `internal/migrate/rollback.go:54-69`
- Create: `internal/migrate/rollback_test.go` (or skip if exists and add test)

- [ ] **Step 1: Add cross-volume check before os.Rename in Rollback()**

Replace the `else` branch at `rollback.go:60-69`:

```go
	} else {
		if Sys.Volume.IsCrossVolume(rec.NewPath, rec.OriginalPath) {
			if err := rollbackCrossVolume(rec.NewPath, rec.OriginalPath); err != nil {
				return err
			}
		} else {
			if err := os.Rename(rec.NewPath, rec.OriginalPath); err != nil {
				return &MoveKitError{
					Code:    ErrCodeFileLocked,
					Message: "Failed to move files back during rollback",
					Detail:  err.Error(),
					Action:  "Close any programs using files in the target folder and retry",
				}
			}
		}
	}
```

- [ ] **Step 2: Verify existing rollbackCrossVolume function works in both directions**

Check that `rollbackCrossVolume` at line 105 does not hardcode direction assumptions.
WalkDir from `src` to `dst` — this direction is correct since it walks the source and copies to destination. No fix needed.

- [ ] **Step 3: Run tests**

```bash
cd C:\Users\Admin\Desktop\Vibe coding\MoveKit && go test ./internal/migrate/ -run TestRollback -v
```

- [ ] **Step 4: Build validation**

```bash
cd C:\Users\Admin\Desktop\Vibe coding\MoveKit && go build ./...
```

---

### Task 2: C2 — Junction failure orphans files at target

**Root cause:** `migrate.go:157-219`: files are already moved to target before junction creation. If junction creation fails, files are at target but no junction exists at source → user can't access files via original path.

**Files:**
- Modify: `internal/migrate/migrate.go:157-219`

- [ ] **Step 1: Add file rollback on junction creation failure**

In `Migrate()`, after files are moved but junction creation fails, move files back from target to source before returning error.

After the file move block (around line 152) store whether a rollback is needed. The fix adds a deferred cleanup that moves files back if junction creation fails.

Replace lines 157-219 (the junction creation block) with:

```go
	// 6. Create junction at original location.
	// Strategy: create at temp path first, verify, then swap. This minimizes
	// the window where the junction is missing (crash-safe).
	// If junction creation fails, roll back: move files from target back to source.
	tmpJunction := sourcePath + ".movekit_tmp"

	cleanupTempJunction := func() {
		Sys.Junction.DeleteJunction(tmpJunction)
		os.RemoveAll(tmpJunction)
	}

	// Clean up any leftover temp junction from a previous crash.
	if _, statErr := os.Stat(tmpJunction); statErr == nil {
		cleanupTempJunction()
	}

	if mdErr := os.MkdirAll(tmpJunction, 0700); mdErr != nil {
		// Junction creation failed — roll back files from target to source.
		log.Error("Cannot create temp directory for junction: %v — rolling back files", mdErr)
		if rbErr := rollbackFilesToSource(targetPath, actualSource, log); rbErr != nil {
			log.Error("CRITICAL: file rollback also failed: %v", rbErr)
		}
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Cannot create temp directory for junction",
			Detail:  mdErr.Error(),
			Action:  "Check permissions on the source location",
		}
	}

	if jErr := Sys.Junction.CreateJunction(tmpJunction, targetPath); jErr != nil {
		cleanupTempJunction()
		log.Error("Failed to create junction: %v — rolling back files", jErr)
		if rbErr := rollbackFilesToSource(targetPath, actualSource, log); rbErr != nil {
			log.Error("CRITICAL: file rollback also failed: %v", rbErr)
		}
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Failed to create junction point",
			Detail:  jErr.Error(),
			Action:  "Ensure the source directory exists and you have write permission",
		}
	}

	// Verify the temp junction actually points to our target.
	if resolved, resErr := resolveJunctionTarget(tmpJunction); resErr != nil || !pathsEqual(resolved, targetPath) {
		cleanupTempJunction()
		log.Error("Junction verification failed — rolling back files")
		if rbErr := rollbackFilesToSource(targetPath, actualSource, log); rbErr != nil {
			log.Error("CRITICAL: file rollback also failed: %v", rbErr)
		}
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Junction verification failed",
			Detail:  fmt.Sprintf("expected %s, got %s (err: %v)", targetPath, resolved, resErr),
			Action:  "Retry the migration",
		}
	}

	// Now swap: delete old junction, rename temp to final.
	if Sys.Junction.IsJunction(sourcePath) {
		if dErr := Sys.Junction.DeleteJunction(sourcePath); dErr != nil {
			cleanupTempJunction()
			log.Error("Failed to remove old junction during swap — rolling back files")
			if rbErr := rollbackFilesToSource(targetPath, actualSource, log); rbErr != nil {
				log.Error("CRITICAL: file rollback also failed: %v", rbErr)
			}
			return &MoveKitError{
				Code:    ErrCodeJunctionFailed,
				Message: "Failed to remove old junction during swap",
				Detail:  dErr.Error(),
				Action:  "Run as administrator and retry",
			}
		}
		os.Remove(sourcePath) // remove empty dir left by junction deletion
	}

	if rErr := os.Rename(tmpJunction, sourcePath); rErr != nil {
		cleanupTempJunction()
		log.Error("Failed to finalize junction — rolling back files")
		if rbErr := rollbackFilesToSource(targetPath, actualSource, log); rbErr != nil {
			log.Error("CRITICAL: file rollback also failed: %v", rbErr)
		}
		return &MoveKitError{
			Code:    ErrCodeJunctionFailed,
			Message: "Failed to finalize junction",
			Detail:  rErr.Error(),
			Action:  "Run as administrator and retry",
		}
	}
```

- [ ] **Step 2: Add rollbackFilesToSource helper at end of migrate.go**

```go
// rollbackFilesToSource moves files from targetPath back to sourcePath.
// Used when junction creation fails after files have already been moved.
// Returns nil on success. If this also fails, the caller logs a CRITICAL message.
func rollbackFilesToSource(targetPath, sourcePath string, log *Logger) error {
	// Ensure source parent directory exists.
	parentDir := filepath.Dir(sourcePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("create parent %s: %w", parentDir, err)
	}

	if Sys.Volume.IsCrossVolume(targetPath, sourcePath) {
		// Cross-volume: copy each file then delete source.
		return filepath.WalkDir(targetPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relPath, relErr := filepath.Rel(targetPath, path)
			if relErr != nil {
				return relErr
			}
			dstPath := filepath.Join(sourcePath, relPath)
			if d.IsDir() {
				return os.MkdirAll(dstPath, 0755)
			}
			if cpErr := Sys.File.CopyFileWithACL(path, dstPath, nil); cpErr != nil {
				return fmt.Errorf("copy %s: %w", relPath, cpErr)
			}
			if rmErr := os.Remove(path); rmErr != nil {
				return fmt.Errorf("remove %s: %w", relPath, rmErr)
			}
			return nil
		})
	}

	// Same volume: use atomic rename.
	if err := os.Rename(targetPath, sourcePath); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", targetPath, sourcePath, err)
	}
	return nil
}
```

- [ ] **Step 3: Run build validation**

```bash
cd C:\Users\Admin\Desktop\Vibe coding\MoveKit && go build ./...
```

---

### Task 3: C3 — Conflict dialog skip/overwrite same behavior

**Root cause:** `App.tsx:366-385`: both `onSkip` and `onOverwrite` call `StartMigration(selectedDisk)` identically, restarting the entire migration. The distinction between skip and overwrite is never communicated to the backend.

**Files:**
- Modify: `frontend/src/App.tsx:362-390`

- [ ] **Step 1: Fix conflict dialog callbacks to continue flow, not restart**

The current code restarts `StartMigration` from scratch. Since the backend already handles conflicts by logging and proceeding (MVP behavior), and `ResolveConflict` is a NOT_IMPLEMENTED stub, the frontend should simply continue the migration flow without distinction.

Replace the conflict dialog callbacks (lines 362-390):

```tsx
{showConflictDialog && (
    <ConflictDialog
        conflicts={conflicts}
        onSkip={() => {
            setShowConflictDialog(false);
            // Per MVP behavior: conflicts are logged but do not block.
            // Proceed with the migration that was already started.
            setState('migrating');
            StartMigration(selectedDisk).catch((err) => {
                const msg = err instanceof Error ? err.message : '未知错误';
                setError(`迁移失败：${msg}`);
                setState('failed');
            });
        }}
        onOverwrite={() => {
            setShowConflictDialog(false);
            // Per MVP behavior: conflicts are logged but do not block.
            // Currently skip and overwrite have the same behavior — overwrite is
            // the default for cross-volume copy (CopyFileExW overwrites by default).
            setState('migrating');
            StartMigration(selectedDisk).catch((err) => {
                const msg = err instanceof Error ? err.message : '未知错误';
                setError(`迁移失败：${msg}`);
                setState('failed');
            });
        }}
        onCancel={() => {
            setShowConflictDialog(false);
            setState('idle');
        }}
    />
)}
```

Actually, the **real fix** is simpler. The problem isn't the callbacks being identical — it's that both call `StartMigration` which triggers a *new* migration goroutine. Instead, they should continue the *existing* flow from `handleConfirm`.

Better fix: extract the migration-start part of `handleConfirm` so the conflict dialog callbacks can call it without re-running conflict detection.

```tsx
const handleConfirm = async () => {
    // Check for file conflicts before migrating.
    try {
      const allConflicts: ConflictFile[] = [];
      for (const f of folders) {
        if (f.status === 'migrated' || f.status === 'not_on_c') continue;
        const target = `${selectedDisk}\\${f.name}`;
        const result = await DetectConflicts(f.sourcePath, target);
        if (result && result.length > 0) {
          for (const c of result) {
            allConflicts.push({
              filename: c.relativePath || c.RelativePath || '',
              sourceSize: c.srcSize || c.SrcSize || 0,
              sourceDate: c.srcModTime || c.SrcModTime || '',
              targetSize: c.dstSize || c.DstSize || 0,
              targetDate: c.dstModTime || c.DstModTime || '',
            });
          }
        }
      }
      if (allConflicts.length > 0) {
        setConflicts(allConflicts);
        setShowConflictDialog(true);
        return; // Wait for user to acknowledge conflicts
      }
    } catch {
      // If conflict detection fails, proceed anyway
    }

    doStartMigration();
  };

  const doStartMigration = async () => {
    setState('migrating');
    try {
      await StartMigration(selectedDisk);
    } catch (err) {
      const msg = err instanceof Error ? err.message : '未知错误';
      setError(msg.includes('syscall') || msg.includes('access denied')
        ? '迁移失败：权限不足，请以管理员身份运行 MoveKit 后重试。'
        : `迁移失败：${msg}`);
      setState('failed');
    }
  };
```

Then update conflict dialog callbacks:

```tsx
{showConflictDialog && (
    <ConflictDialog
        conflicts={conflicts}
        onSkip={() => {
            setShowConflictDialog(false);
            // MVP: conflicts are logged but do not block.
            // Skip and overwrite both proceed — CopyFileExW overwrites by default.
            doStartMigration();
        }}
        onOverwrite={() => {
            setShowConflictDialog(false);
            // MVP: conflicts are logged but do not block.
            // Skip and overwrite both proceed — CopyFileExW overwrites by default.
            doStartMigration();
        }}
        onCancel={() => {
            setShowConflictDialog(false);
            setState('idle');
        }}
    />
)}
```

- [ ] **Step 2: Build frontend to verify**

```bash
cd C:\Users\Admin\Desktop\Vibe coding\MoveKit\frontend && npx tsc --noEmit
```

---

### Independent Review Plan

After all 3 tasks complete:
1. Each fix is in a different file — zero merge conflicts
2. Review C1 fix: verify `Sys.Volume.IsCrossVolume` is called correctly
3. Review C2 fix: verify `rollbackFilesToSource` handles both cross-volume and same-volume
4. Review C3 fix: verify `doStartMigration` is extracted correctly without regressions
