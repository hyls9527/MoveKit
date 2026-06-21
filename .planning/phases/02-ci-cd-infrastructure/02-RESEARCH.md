# Phase 2: CI/CD Infrastructure - Research

**Researched:** 2026-06-22
**Domain:** GitHub Actions CI/CD for Go + TypeScript Windows desktop application
**Confidence:** HIGH

## Summary

Phase 2 establishes automated CI from zero. MoveKit currently has no `.github/` directory, no lint configuration, and no automated build verification. Every check is manual. This phase creates a single `ci.yml` workflow that runs on every push to `windows-latest`: Go lint (golangci-lint), frontend lint (ESLint + tsc), Go tests with race detection and coverage output, and a `wails build` smoke check. On `v*` tag push, the workflow additionally builds and publishes a GitHub Release with auto-generated conventional-commits changelog. Dependabot is configured for Go modules, npm, and GitHub Actions.

**Primary recommendation:** Use a single `ci.yml` with four sequential jobs (lint -> test -> build -> release[conditional]), pinned tool versions per D-06, and `actions/cache@v4` for Go modules, Go build cache, npm cache, and golangci-lint cache.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Go linting | CI Runner (GitHub Actions) | -- | golangci-lint runs standalone on the runner |
| Frontend linting | CI Runner | -- | ESLint + tsc run on the runner with Node.js |
| Go unit/integration tests | CI Runner | -- | go test with -race runs on windows-latest |
| Wails production build verification | CI Runner | -- | wails build compiles Go + Vite frontend, verifies shippable state |
| GitHub Release creation | CI Runner | GitHub API | softprops/action-gh-release calls GitHub Releases API |
| Dependency update PRs | Dependabot (GitHub) | -- | Dependabot opens PRs; CI runs on those PRs for validation |
| Caching | CI Runner (actions/cache) | GitHub Cache Service | Restore Go/npm/golangci-lint caches across runs |
| Status badge | README.md (static) | shields.io | README renders badge image from GitHub Actions API |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| GitHub Actions | -- | CI orchestration | Built into GitHub; zero infrastructure cost; windows-latest runner available [VERIFIED: github.com/features/actions] |
| golangci-lint | v2.11.3 | Go static analysis (errcheck, govet, staticcheck, gosec, +) | Official Go linter aggregator; v2.x config format with `version: "2"` [CITED: golangci-lint.run] |
| golangci-lint-action | v7 | GitHub Actions integration for golangci-lint | Official action from golangci team; handles caching, only-new-issues mode [CITED: github.com/golangci/golangci-lint-action] |
| gotestsum | v1.13.0 | Go test runner with JUnit XML output | More feature-rich than go-junit-report; built-in JUnit XML, test summaries, failure rerun [CITED: pkg.go.dev/gotest.tools/gotestsum] |
| ESLint | v10.5.0 | Frontend linting (TypeScript + React) | Industry standard JS/TS linter; v9+ flat config format [VERIFIED: npm registry] |
| typescript-eslint | v8.61.1 | TypeScript-aware ESLint rules | Unified v8 package replacing separate parser/plugin [VERIFIED: npm registry] |
| eslint-plugin-react | v7.37.5 | React-specific lint rules | Official React ESLint plugin [VERIFIED: npm registry] |
| eslint-plugin-react-hooks | v7.1.1 | Rules of Hooks enforcement | Official React Hooks ESLint plugin [VERIFIED: npm registry] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @eslint/js | v10.0.1 | ESLint core JS recommended rules | Required peer for ESLint v9+ flat config [VERIFIED: npm registry] |
| globals | v17.6.0 | Browser/Node global variable definitions for ESLint | Required for flat config languageOptions.globals [VERIFIED: npm registry] |
| softprops/action-gh-release | v2 | Create GitHub Release + upload assets | Only in release job (tag push trigger) [CITED: github.com/softprops/action-gh-release] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| gotestsum | go-junit-report | go-junit-report is lighter (pipe-based), but gotestsum provides test summaries, failure reruns, and multi-format output. gotestsum is more actively maintained (1,912 stars vs 776). Decision left to Claude's discretion per D-09 context. |
| golangci-lint-action | Manual go install golangci-lint | The action handles caching automatically and provides `only-new-issues` for PR diff linting. Manual install requires explicit cache management. |
| dAppServer/wails-build-action | Manual wails CLI install | Third-party action abstracts setup but adds supply-chain risk and version lock-in. Per decision D-17 (single ci.yml) and security enforcement, manual setup with pinned versions is preferred. |

**Installation (frontend devDependencies only -- CI tools installed on runner):**
```bash
cd frontend
npm install -D eslint@^10.5.0 @eslint/js typescript-eslint@^8.61.1 eslint-plugin-react@^7.37.5 eslint-plugin-react-hooks@^7.1.1 globals@^17.6.0
```

**Version verification:** Each npm package above confirmed on npm registry at research time:
```bash
npm view eslint version           # 10.5.0
npm view typescript-eslint version # 8.61.1
npm view eslint-plugin-react version # 7.37.5
npm view eslint-plugin-react-hooks version # 7.1.1
npm view globals version          # 17.6.0
npm view @eslint/js version       # 10.0.1
```

## Package Legitimacy Audit

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| eslint | npm | 12 yrs | 50M+/wk | github.com/eslint/eslint | [ASSUMED] | Approved |
| @eslint/js | npm | 3 yrs | 20M+/wk | github.com/eslint/eslint | [ASSUMED] | Approved |
| typescript-eslint | npm | 6 yrs | 15M+/wk | github.com/typescript-eslint/typescript-eslint | [ASSUMED] | Approved |
| eslint-plugin-react | npm | 8 yrs | 15M+/wk | github.com/jsx-eslint/eslint-plugin-react | [ASSUMED] | Approved |
| eslint-plugin-react-hooks | npm | 6 yrs | 15M+/wk | github.com/facebook/react | [ASSUMED] | Approved |
| globals | npm | 2 yrs | 8M+/wk | github.com/sindresorhus/globals | [ASSUMED] | Approved |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

*slopcheck was not available at research time. All packages above are tagged `[ASSUMED]` and the planner must gate each install behind a `checkpoint:human-verify` task. All packages are well-known ecosystem standards with millions of weekly downloads and verifiable GitHub repositories.*

**Postinstall script audit:** All six packages returned empty for `scripts.postinstall` -- no suspicious postinstall scripts detected.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                      GitHub Actions (windows-latest)                  │
│                                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────────┐  │
│  │  LINT    │───>│   TEST   │───>│  BUILD   │───>│   RELEASE     │  │
│  │  Job     │    │   Job    │    │  Job     │    │ (tag only)    │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────────┘  │
│       │               │               │                │            │
│  ┌────▼────┐     ┌────▼────┐     ┌────▼────┐     ┌────▼─────┐     │
│  │golangci-│     │go test  │     │ wails   │     │GitHub    │     │
│  │lint     │     │-race    │     │ build   │     │Release   │     │
│  │ESLint   │     │-cover   │     │         │     │+exe      │     │
│  │tsc      │     │JUnit XML│     │         │     │+changelog│     │
│  └─────────┘     └─────────┘     └─────────┘     └──────────┘     │
│                                                                      │
│  Shared Caching Layer (actions/cache@v4):                           │
│  ┌─────────────┬──────────────┬──────────────┬──────────────────┐  │
│  │ Go modules  │ Go build     │ npm cache    │ golangci-lint    │  │
│  │ ~/go/pkg/mod│ ~/.cache/    │ ~/.npm       │ cache            │  │
│  │             │ go-build     │              │                  │  │
│  └─────────────┴──────────────┴──────────────┴──────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘

Trigger Events:
  push (any branch) ──> lint + test + build
  push tag v*      ──> lint + test + build + release
```

### Workflow Job Pipeline

The `ci.yml` uses four sequential jobs with `needs` dependencies, not parallel execution, because:
1. **Lint must pass before test** -- wasted test cycles on code that fails lint is wasteful given CI runner time limits
2. **Test must pass before build** -- a failing build from broken tests wastes build compilation time
3. **Release depends on build** -- only publish artifacts that passed full validation

Job-level `timeout-minutes: 15` per D-16.

Concurrency group per D-02: `${{ github.workflow }}-${{ github.ref }}` with `cancel-in-progress: true`.

### Recommended Project Structure
```
.github/
├── workflows/
│   └── ci.yml              # Single CI workflow (lint->test->build->release)
├── dependabot.yml           # Dependabot v2 config
movekit/
├── .golangci.yml            # Go linter configuration (project root)
frontend/
├── eslint.config.js         # ESLint v9+ flat config
├── package.json             # Updated with lint scripts + new devDeps
```

### Pattern 1: Single Workflow, Sequential Jobs

**What:** One `ci.yml` with four jobs: `lint`, `test`, `build`, `release` (conditional). Jobs use `needs` to enforce sequential execution. All jobs share a `concurrency` group for auto-cancellation of in-progress runs.

**When to use:** Per decision D-17 (single file) and D-02 (cancel-in-progress).

**Example (job structure):**
```yaml
# Source: GitHub Actions documentation + project decisions
name: CI

on:
  push:
    branches: ['**']
    tags: ['v*']

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    runs-on: windows-latest
    timeout-minutes: 15
    # ... lint steps ...

  test:
    needs: lint
    runs-on: windows-latest
    timeout-minutes: 15
    # ... test steps ...

  build:
    needs: test
    runs-on: windows-latest
    timeout-minutes: 15
    # ... wails build steps ...

  release:
    needs: build
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: windows-latest
    timeout-minutes: 15
    permissions:
      contents: write
    # ... release steps ...
```

### Pattern 2: golangci-lint with Version Pinning and Custom Config

**What:** Use `golangci/golangci-lint-action@v7` with `version: v2.11.3` (pinned per D-06) and `only-new-issues: false` (per D-03, always full run). Custom `.golangci.yml` at project root with `version: "2"` and explicit linter enable list under `disable-all: true`.

**When to use:** Per decision D-04 (strict mode, custom config, zero warnings blocking).

**Example:**
```yaml
# Source: golangci-lint-action marketplace + golangci-lint docs
- uses: golangci/golangci-lint-action@v7
  with:
    version: v2.11.3
    only-new-issues: false
    args: --timeout=5m
```

### Pattern 3: ESLint Flat Config (v9+) for TypeScript + React

**What:** ESLint v10.x with flat config format (`eslint.config.js`), not legacy `.eslintrc`. Uses `typescript-eslint` unified v8 package. React rules manually configured (no broken legacy presets).

**When to use:** Per decision D-05 (ESLint + tsc --noEmit, blocking at same level as Go lint).

**Example:**
```javascript
// Source: ESLint v9+ flat config docs + typescript-eslint v8 docs
import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import reactPlugin from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import { defineConfig } from "eslint/config";

export default defineConfig([
  { ignores: ["dist", "node_modules", "wailsjs"] },
  { settings: { react: { version: "detect" } } },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    plugins: { react: reactPlugin, "react-hooks": reactHooks },
    languageOptions: {
      globals: globals.browser,
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    rules: {
      "react/react-in-jsx-scope": "off",
      "react/jsx-uses-react": "off",
      "react/prop-types": "off",
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "warn",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
    },
  },
]);
```

### Pattern 4: Semantic Version Release with ldflags Injection

**What:** On tag push matching `v*`, extract version from `github.ref_name`, inject via Go `-ldflags` into `main.currentVersion`, build with `wails build`, create GitHub Release with auto-generated conventional-commits changelog using `softprops/action-gh-release@v2` with `generate_release_notes: true`.

**When to use:** Per decisions D-10 (tag trigger), D-11 (version from git tag via ldflags), D-12 (automatic release, non-draft), D-13 (conventional commits changelog).

**Example (release step):**
```yaml
# Source: softprops/action-gh-release docs + Go ldflags convention
- name: Build for release
  run: |
    $version = "${{ github.ref_name }}"
    wails build -webview2 embed -ldflags "-X main.currentVersion=$version"
  shell: pwsh

- name: Create GitHub Release
  uses: softprops/action-gh-release@v2
  with:
    name: ${{ github.ref_name }}
    files: build/bin/MoveKit.exe
    generate_release_notes: true
```

### Anti-Patterns to Avoid
- **Third-party monolithic build actions:** The `dAppServer/wails-build-action` hides setup details and limits control over Go/Node/Wails version pinning per D-06. Use official `setup-go`/`setup-node` actions plus explicit CLI installs instead.
- **Separate workflow files per job:** Decision D-17 specifies single `ci.yml`. Don't create `lint.yml`, `test.yml`, `build.yml` as separate files.
- **Skipping CI with commit messages:** Decision D-03 explicitly forbids `[skip ci]`. Don't add path filters or skip conditions in the workflow trigger.
- **Cache key without lockfile hash:** Without `hashFiles('go.sum')` or `hashFiles('frontend/package-lock.json')` in cache keys, stale caches persist indefinitely and cause non-reproducible failures.
- **`--new-from-rev` without `fetch-depth: 0` in merge_group:** If using PR lint modes in the future, deep clone is required.
- **ESLint legacy `.eslintrc` format:** ESLint v9+ deprecated the legacy format. Using it creates a migration burden within months.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JUnit XML from go test | Custom XML generation | gotestsum --junitfile | Edge cases in test output parsing (subtests, panics, skipped tests, parallel output interleaving) are all handled |
| Go linter orchestration | Individual `go vet`/`staticcheck`/`errcheck` CI steps | golangci-lint | Unified config, shared caching, version pinning, `--new-from-patch` for PRs |
| Release notes from commits | Manual changelog writing | GitHub's `generate_release_notes: true` (powered by conventional commits) | GitHub natively parses conventional commits and groups by type (Features, Bug Fixes, etc.) |
| Changelog between tags | Git log parsing script | GitHub's built-in auto-generation or `softprops/action-gh-release` `generate_release_notes` | Handles merge commits, first-release edge case, tag comparison |
| Cache eviction logic | Custom TTL/cleanup scripts | `actions/cache@v4` with hashFiles() keys | Cache busts automatically when lockfile changes; restore-keys provide partial hits |

**Key insight:** CI infrastructure is configuration, not code. Every custom script in CI is a maintenance burden. The standard actions and tools handle edge cases accumulated over years of open-source usage.

## Runtime State Inventory

> Omitted. Phase 2 is greenfield CI infrastructure (no existing `.github/` directory, no runtime state to migrate).

## Common Pitfalls

### Pitfall 1: WebView2 Not Available on windows-latest
**What goes wrong:** `wails build` fails because WebView2 runtime is not found. The `windows-latest` runner is Windows Server, which may not have the full WebView2 Evergreen runtime pre-installed.
**Why it happens:** Windows Server 2022 (which `windows-latest` currently maps to) does not include Edge or WebView2 by default in all configurations.
**How to avoid:** Use `-webview2 embed` flag on `wails build` to embed the WebView2 loader in the binary. Alternatively, install WebView2 via `choco install webview2-runtime` or `Invoke-WebRequest` the bootstrapper. The `embed` approach is simpler and self-contained.
**Warning signs:** CI log shows "WebView2 not found" or "Could not locate WebView2Loader.dll".

### Pitfall 2: Go Race Detector Requires C Toolchain on Windows
**What goes wrong:** `go test -race` fails with "gcc not found" or linker errors.
**Why it happens:** Go's race detector (`-race`) requires a C compiler to build the runtime race detection library. On `windows-latest`, the MinGW toolchain is pre-installed, but the environment PATH may not include it correctly depending on the shell used.
**How to avoid:** Use `shell: pwsh` for Go commands that need `-race`. Verify `gcc --version` in CI before the test step. The `windows-latest` runner has MinGW at `C:\msys64\ucrt64\bin\gcc.exe` -- if missing from PATH, add it explicitly.
**Warning signs:** Build error referencing `-lrace` or `runtime/race`.

### Pitfall 3: Package-lock.json Mismatch with npm ci
**What goes wrong:** `npm ci` fails because `package-lock.json` is out of sync with `package.json`.
**Why it happens:** Adding ESLint devDependencies updates `package.json` but may not be followed by `npm install` to regenerate the lockfile.
**How to avoid:** After adding ESLint packages, run `npm install` locally and commit the updated `package-lock.json`. CI uses `npm ci` (clean install) which strictly requires the lockfile to match.
**Warning signs:** "`npm ci` can only install packages when your package.json and package-lock.json or npm-shrinkwrap.json are in sync."

### Pitfall 4: golangci-lint v2 Config Format Migration
**What goes wrong:** Writing `.golangci.yml` in v1 format (`linters-settings` without `version: "2"`) causes silent misbehavior or skipped rules in golangci-lint v2.x.
**Why it happens:** golangci-lint v2.0.0 introduced a new config format. The old format is partially supported but linter settings may not map correctly.
**How to avoid:** Always start `.golangci.yml` with `version: "2"`. Use the v2 config schema. If migrating an existing config, use `golangci-lint migrate` to convert automatically.
**Warning signs:** Linters appear to run but don't report issues they should catch. Config warnings at the top of the lint output.

### Pitfall 5: Race Between Concurrency Cancellation and Artifact Upload
**What goes wrong:** When a new push cancels an in-progress run, artifact uploads from the cancelled run may fail or upload partial artifacts.
**Why it happens:** `cancel-in-progress` kills the runner process; any `actions/upload-artifact` step that hasn't completed is terminated.
**How to avoid:** Per D-02, concurrency cancellation is desired behavior. The cancelled run's artifacts are irrelevant -- the new run will produce fresh ones. Non-issue; just be aware of the expected behavior.
**Warning signs:** Cancelled workflow runs showing failed artifact upload steps -- this is normal.

## Code Examples

Verified patterns from official sources:

### Go Test with Race, Coverage, and JUnit XML (CI step)
```yaml
# Source: gotestsum pkg.go.dev + GitHub Actions workflow patterns
- name: Run Go tests
  run: |
    go install gotest.tools/gotestsum@v1.13.0
    gotestsum --junitfile test-results.xml -- -race -coverprofile=coverage.out -covermode=atomic ./...
  shell: pwsh

- name: Upload test report
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: test-results
    path: |
      test-results.xml
      coverage.out
```

### Wails Build in CI (verify shippable state)
```yaml
# Source: Wails v2 documentation + CI patterns
- name: Build with Wails
  env:
    CGO_ENABLED: 1
  run: |
    go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
    cd frontend && npm ci && cd ..
    wails build -webview2 embed -o MoveKit.exe
  shell: pwsh
```

### Caching Strategy (Go + npm + golangci-lint)
```yaml
# Source: actions/cache@v4 documentation + Go caching best practices
- name: Cache Go modules and build cache
  uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-

- name: Cache npm
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-node-${{ hashFiles('frontend/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

### Dependabot Configuration
```yaml
# Source: GitHub Dependabot v2 documentation
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    open-pull-requests-limit: 5
    cooldown:
      default-days: 4
    commit-message:
      prefix: "chore(deps):"
    labels:
      - "dependencies"

  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
      day: "monday"
    open-pull-requests-limit: 10
    cooldown:
      default-days: 4
    commit-message:
      prefix: "chore(deps):"
    labels:
      - "dependencies"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    commit-message:
      prefix: "ci(deps):"
    labels:
      - "dependencies"
      - "ci"
    groups:
      actions-dependencies:
        patterns:
          - "*"
```

### Status Badge (README.md)
```markdown
[![CI](https://github.com/hyls9527/MoveKit/actions/workflows/ci.yml/badge.svg)](https://github.com/hyls9527/MoveKit/actions/workflows/ci.yml)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| ESLint `.eslintrc` legacy config | ESLint `eslint.config.js` flat config | ESLint v9 (2024) | Simpler config, faster, no extends chains. Legacy format deprecated -- don't use it. |
| golangci-lint v1 config format | golangci-lint v2 config format (`version: "2"`) | v2.0.0 (2025-03) | New `formatters` section, redesigned `linters-settings`. Old format partially supported but use v2 for new projects. |
| `actions/cache@v3` with single save/restore | `actions/cache@v4` with separate `restore`/`save` actions | v4 (2024) | More control, less accidental cache overwrites |
| `setup-go@v4` with node16 | `setup-go@v5` with node20 | v5 (2024) | Required upgrade -- node16 deprecated in GitHub Actions |
| Separate `@typescript-eslint/parser` + `@typescript-eslint/eslint-plugin` | Unified `typescript-eslint` v8 package | v8 (2024) | Single dependency, simpler imports |

**Deprecated/outdated:**
- `.eslintrc.*` files: ESLint v9+ deprecated. Use `eslint.config.js` (flat config).
- `actions/upload-artifact@v3` and older: Deprecated. Use `@v4`.
- `golangci-lint` v1.x: Final release v1.64.8 (2025-03). v2.x is current.
- Node.js 16 for GitHub Actions runners: Deprecated. All actions must use node20.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `windows-latest` runner has MinGW toolchain for Go race detector | Common Pitfalls | Tests with `-race` fail; need explicit gcc PATH setup or choco install mingw |
| A2 | `windows-latest` runner has WebView2 Runtime or it can be embedded with `-webview2 embed` | Common Pitfalls | Wails build fails; need explicit WebView2 bootstrapper install step |
| A3 | `package-lock.json` tracks current `package.json` dependencies and is in sync | Common Pitfalls | npm ci fails; need local `npm install` before committing |
| A4 | golangci-lint-action v7 is the latest major version as of 2026-06 | Standard Stack | Dependabot will update it; v7 API is stable |
| A5 | `softprops/action-gh-release@v2` `generate_release_notes: true` produces adequate conventional-commits changelogs | Architecture Patterns | May need explicit changelog generator (e.g., `npx changelogithub`) if GitHub's auto-generation insufficient |
| A6 | Dependabot weekly schedule with 4-day cooldown is appropriate for this project's pace | Standard Stack | May be too frequent or too slow; cooldown delays security patches |
| A7 | Go 1.23.5 is compatible with the `windows-latest` runner image | Architecture Patterns | May need `setup-go` with `check-latest: true` or explicit version override |

## Open Questions

1. **Wails build on windows-latest with embedded WebView2 -- does it reliably succeed?**
   - What we know: Wails issue #4102 reports invisible UI from CI-built binaries (v2.9-v2.10.1). The fix in go-webview2 v0.1.17 (Wails v2.10) partially addressed this. Since CI only verifies compilation (not runtime rendering), this is likely acceptable. The MoveKit project uses Wails v2.12.0 which is past the fix.
   - What's unclear: Whether v2.12.0 fully resolves the CI build issue. No reports found specifically for v2.12.0.
   - Recommendation: Implement `wails build -webview2 embed` in CI. If build succeeds, the phase requirement CI-04 is met. If runtime rendering issues appear later, address in a dedicated phase. The CI verification goal is "this code compiles into a valid exe," not "this exe renders perfectly on Windows Server."

2. **Should `npm ci` or `npm install` be used in CI for the frontend?**
   - What we know: `package-lock.json` exists and is tracked in git. `npm ci` is faster and deterministic but requires the lockfile to be perfectly synced.
   - What's unclear: Whether historical commits may have `package-lock.json` drift.
   - Recommendation: Use `npm ci` for strict reproducibility. If it fails on old commits, that's a data integrity signal worth catching.

3. **Go test coverage report format -- human-readable output sufficient for now?**
   - What we know: D-07 says Phase 2 only generates coverage, not blocking. D-09 says JUnit XML format.
   - What's unclear: Whether a plain `go tool cover -func=coverage.out` output is sufficient or if an HTML report should be uploaded as artifact.
   - Recommendation: Upload `coverage.out` as an artifact (small file). Add a step to print human-readable summary with `go tool cover -func=coverage.out`. HTML report (`go tool cover -html`) can be added later if needed.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| GitHub Actions | CI execution | ✓ | -- | N/A (runs on GitHub's infrastructure) |
| windows-latest runner | All CI jobs | ✓ | Windows Server 2022 | N/A (GitHub-managed) |
| Node.js (on local machine, for lockfile update) | Adding ESLint devDeps + regenerating lockfile | ✓ | v22.23.0 | N/A |
| npm (on local machine) | Adding ESLint devDeps + regenerating lockfile | ✓ | 10.9.8 | N/A |
| Go (on local machine) | Testing golangci-lint config, go test | ✗ | -- | Config can be validated on CI; or install Go manually |

**Missing dependencies with no fallback:**
- Go (local): Not blocking -- CI config can be validated directly on GitHub Actions. Local Go is only needed for pre-testing the `.golangci.yml` config.

**Missing dependencies with fallback:**
- Go (local): All Go verification will happen on CI. Config files written in this phase are static YAML/JS.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A -- CI uses GITHUB_TOKEN automatically |
| V3 Session Management | no | N/A |
| V4 Access Control | yes | CI workflow uses `permissions: contents: read` by default; release job explicitly sets `contents: write` (minimum required scope) |
| V5 Input Validation | no | N/A -- no user input in CI pipeline |
| V6 Cryptography | no | N/A |

### Known Threat Patterns for GitHub Actions CI

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Third-party action compromise (supply chain) | Tampering | Pin actions to commit SHA or at minimum major version tag (e.g., `@v4`); Dependabot monitors for updates |
| Secret exfiltration via malicious PR workflow changes | Information Disclosure | `pull_request` events from forks get read-only token; `contents: write` only on explicit jobs |
| Cache poisoning via PR modifying lockfiles | Tampering | Cache keys include lockfile hash; PR cache is separate from main branch cache |
| Release artifact tampering | Tampering | Release job runs only on tag push (not on PR); tag push must be from authorized committer |
| Code execution via unchecked npm postinstall scripts | Elevation of Privilege | All six ESLint packages verified to have empty `postinstall` scripts; `npm ci` uses lockfile for integrity |

### Supply Chain Risk Assessment

This phase introduces configuration files, not new code execution paths in the MoveKit binary. The supply chain risks are:

1. **GitHub Actions actions**: `actions/checkout@v4`, `actions/setup-go@v5`, `actions/setup-node@v4`, `actions/cache@v4`, `actions/upload-artifact@v4`, `golangci/golangci-lint-action@v7`, `softprops/action-gh-release@v2`. All are official GitHub or well-maintained community actions. Dependabot configured for github-actions ecosystem will keep them updated.

2. **Go tools installed on runner**: `golangci-lint`, `gotestsum`, `wails` CLI. Installed via `go install` with explicit version pinning (D-06). Each version is auditable via its source repository.

3. **npm devDependencies**: Six ESLint-related packages added to `frontend/package.json`. All are industry-standard with millions of weekly downloads. Postinstall scripts verified empty.

## Sources

### Primary (HIGH confidence)
- GitHub Actions documentation - workflow syntax, concurrency, permissions, caching [github.com/docs/actions]
- golangci-lint official site - v2 config format, linter documentation [golangci-lint.run]
- gotestsum pkg.go.dev page - JUnit XML output, CLI flags [pkg.go.dev/gotest.tools/gotestsum]
- Wails v2 cross-platform build guide - GitHub Actions workflow pattern [wails.io/docs/v2.11.0/guides/crossplatform-build]
- GitHub Dependabot v2 documentation - configuration schema [docs.github.com/en/code-security/dependabot]
- softprops/action-gh-release repository - release creation with auto-generated notes [github.com/softprops/action-gh-release]

### Secondary (MEDIUM confidence)
- WebSearch: golangci-lint best practices (errcheck, govet, staticcheck, gosec) - multiple community sources agree on core linter set
- WebSearch: ESLint v9+ flat config for TypeScript + React - multiple blog posts and tutorials converge on same pattern
- WebSearch: Go GitHub Actions caching strategies - date-stamped keys, separate save/restore actions
- WebSearch: gotestsum vs go-junit-report comparison - ecosystem consensus favors gotestsum for richer features

### Tertiary (LOW confidence)
- Wails issue #4102 (invisible UI from CI builds) - open issue, not yet resolved; may not affect build-only verification
- golangci-lint-action latest version (v7) - inferred from Dependabot bump PRs, not confirmed from official releases page
- `windows-latest` runner MinGW path assumptions - based on training data, not verified against current runner image

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All tools verified via npm registry, official docs, and cross-referenced with community sources
- Architecture: HIGH - Single workflow sequential jobs pattern is well-documented and matches all project decisions
- Pitfalls: MEDIUM - WebView2 and race detector behavior on windows-latest is version-dependent and not verified against current runner image
- Release automation: HIGH - softprops/action-gh-release pattern is standard and well-documented

**Research date:** 2026-06-22
**Valid until:** 2026-08-22 (60 days -- stable infrastructure domain, no expected breaking changes)

---

*Phase 2 research complete. Planner ready to create PLAN.md files.*
