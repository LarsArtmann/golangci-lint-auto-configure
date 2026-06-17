# Cross-Project golangci-lint Config Audit & Auto-Fixer Improvement Plan

**Date:** 2026-06-17
**Dataset:** 184 golangci-lint config files across `~/projects/**` (136 active, excluding vendored/archived)
**Goal:** Compare real-world configs against the auto-fixer's behavior and identify concrete improvements.

---

## 1. Executive Summary

The auto-fixer is a **resounding success** at imposing a consistent house style: 100% of
versioned configs are on `v2`, ~95% share the same linter sets, formatters, Go experiment
build tags, and exclusion patterns. This is exactly what the tool was built to do.

However, the audit surfaced **one correctness bug** (uncounted fixes silently dropped) and
**several normalization gaps** that leave configs inconsistent. The highest-impact gap is
that the fixer **never normalizes the `issues` block** — 64 configs (47%) ship with
`issues: {}`, inheriting golangci-lint's low default of `max-same-issues: 3`, which hides
duplicate problems in CI.

---

## 2. The House Style That Emerged (Data)

Analyzing 136 active configs reveals a remarkably consistent pattern — strong evidence the
auto-fixer is the primary config author across this ecosystem.

### 2.1 Structural consistency

| Property                                         | Prevalence | Notes                                            |
| ------------------------------------------------ | ---------- | ------------------------------------------------ |
| `version: "2"`                                   | 100%       | Zero v1 configs remain. Migration feature works. |
| `run.timeout: 5m`                                | 92%        | 11 configs diverge (see gaps).                   |
| `run.allow-parallel-runners`                     | 98%        | Fixer enforces this.                             |
| `run.allow-serial-runners`                       | 98%        | Fixer enforces this.                             |
| `run.build-tags` (5 goexperiment tags)           | 93%        | Fixer injects all 5 experiment tags.             |
| `run.go: 1.26.x`                                 | ~90%       | Fixer sets local Go version.                     |
| `output.formats: {}`                             | 97%        | Explicit empty formats block.                    |
| `linters.exclusions.generated: lax`              | 93%        | Fixer injects this.                              |
| `formatters: [gci, goimports, gofumpt, golines]` | 97%        | Core formatters.                                 |
| `formatters.settings.golines.max-len: 120`       | 94%        | Fixer injects this default.                      |

### 2.2 Linter enablement — two regimes

The data shows **two distinct config regimes**, not one:

1. **Kitchen-sink regime (~55% of configs):** enable ~100+ linters (everything including
   `arangolint`, `clickhouselint`, `depguard`, `ireturn`, `mnd`, `wsl_v5`, etc.). These are
   configs produced by `configure --priority optional`.

2. **Curated regime (~45% of configs):** enable ~55-65 high+critical linters, deliberately
   skipping the noisier optional ones. Examples: `oxlint-auto-configure`,
   `cqrs-htmx/usermgmt`, `go-error-family`, `hierarchical-errors`.

Both regimes converge on the **same core ~50 linters** (the critical + high priority set).
The top-30 most-enabled linters all appear in 133-136 configs (97-100%):

```
misspell, unconvert, exhaustive, copyloopvar, nilerr, nolintlint, revive, gosec,
errorlint, nakedret, gocognit, goconst, errname, nilnesserr, gocyclo, funlen,
dupl, prealloc, sqlclosecheck, nestif, rowserrcheck, intrange, gocritic, bodyclose,
unparam, wrapcheck, ginkgolinter, forcetypeassert, cyclop, perfsprint
```

### 2.3 Settings that are near-universal (house defaults)

These settings appear so consistently they are effectively the house standard — and the
fixer already encodes them as `DefaultLinterSettings`:

| Linter            | Setting                                        | Prevalence |
| ----------------- | ---------------------------------------------- | ---------- |
| `exhaustruct`     | `exclude: [os/exec.Cmd]`                       | 97%        |
| `gocritic`        | `disabled-checks: [ifElseChain]`               | 95%        |
| `cyclop`          | `max-complexity: 12-25`                        | 94%        |
| `gomoddirectives` | `replace-local: true`                          | 90%        |
| `revive`          | disable `exported`, `package-comments`         | 89%        |
| `ireturn`         | `allow: [error, empty, anon, stdlib, generic]` | 88%        |
| `depguard`        | `rules.main.allow: [$gostd, $module]`          | 88%        |

---

## 3. Gaps Found (Prioritized)

### 🔴 P0 — Correctness Bug: Uncounted fixes silently dropped

**File:** `pkg/linter/fixer.go:242-263` (`applyAndSave`)

The fixer mutates the config in several places that **never increment `counts`**:

- `updateGoVersion` — sets `run.go`
- `updateRunnerSettings` — sets `allow-parallel-runners` / `allow-serial-runners`
- `updateBuildTags` — adds Go experiment tags
- `injectDefaultSettings` (via `updateConfigFromSets`) — injects linter/formatter defaults

Then at line 261:

```go
if counts.total() == 0 {
    return noFixesResult()  // ← discards ALL in-memory mutations without saving
}
```

**Impact:** A project that already has all linters/formatters enabled but is missing default
settings (e.g. `ginkgolinter`, `testifylint`), runner flags, or build tags will have those
fixes computed in memory and then **silently thrown away** because `counts.total() == 0`.
This is why 108 configs are missing `ginkgolinter` settings and 105 are missing `testifylint`
settings despite the tool having the correct defaults.

**Fix:** Make every mutating step return its change count and sum them. Alternatively (simpler):
compare the config before/after with a deep-equal check and save whenever it changed.

### 🔴 P0 — `issues` block never normalized

**Files:** `pkg/linter/fixer_config.go` (no issues updater exists)

The fixer has no logic to set `issues.max-issues-per-linter` or `issues.max-same-issues`.
These are only set by the loader when **creating a brand-new config** (`DefaultMaxIssuesPerLinter = 50`,
`DefaultMaxSameIssues = 10`).

**Impact (measured):**

| State                            | Configs  | Effect                                                                                 |
| -------------------------------- | -------- | -------------------------------------------------------------------------------------- |
| `issues: {}` (empty/absent)      | 64 (47%) | Inherits golangci-lint default `max-same-issues: 3` — **hides duplicate issues in CI** |
| Explicit `max-same-issues` set   | 72 (53%) | But values vary wildly: `{10, 5, 15, 50, 20, 0, 3}`                                    |
| Explicit `max-issues-per-linter` | 72       | Values vary: `{50, 100, 200, 0}`                                                       |

A project with `issues: {}` and a repeated `errcheck` violation will only see **3 of them** in
CI output — the rest are silently suppressed. This is a real quality blind spot.

**Fix:** Add an `updateIssuesSettings(cfg)` step to the config updater that injects
`max-issues-per-linter: 50` and `max-same-issues: 10` when absent, mirroring the loader defaults.

### 🟠 P1 — Stale default settings (symptom of the P0 bug)

Once the P0 bug is fixed, re-running `configure` will heal these:

| Missing setting       | Configs affected | Root cause                              |
| --------------------- | ---------------- | --------------------------------------- |
| `ginkgolinter` config | 108              | P0 bug: settings injected but not saved |
| `testifylint` config  | 105              | P0 bug: settings injected but not saved |
| `revive` rules        | 7                | Same                                    |
| `gomoddirectives`     | 7                | Same                                    |
| `gocritic`            | 7                | Same                                    |

### 🟡 P2 — Minor normalization gaps

| Gap                                      | Configs | Notes                                                                               |
| ---------------------------------------- | ------- | ----------------------------------------------------------------------------------- |
| `run.timeout != 5m`                      | 11      | Fixer only repairs _invalid_ timeouts, not non-standard ones. Consider normalizing. |
| Missing formatter exclusion paths        | 11      | `_templ.go` not excluded from formatters.                                           |
| `formatters.exclusions.generated != lax` | 9       |                                                                                     |
| Missing core formatters                  | 4       | A few configs lack `gci`/`gofumpt`/`goimports`.                                     |
| Missing `output.formats` block           | 3       |                                                                                     |

### 🟢 P3 — Strategic observations (not bugs)

1. **`run.modules-download-mode: readonly`** appears in 46 configs (34%). The fixer doesn't
   detect or set this. It could auto-detect vendored projects / CI-readonly environments and
   suggest it. Low priority — this is a deliberate per-project choice.

2. **`run.tests: true`** appears in 99 configs but the fixer never sets it. golangci-lint
   defaults to `tests: true` anyway, so this is cosmetic.

3. **`run.issues-exit-code: 1`** appears in 99 configs; the fixer doesn't normalize it. The
   loader defaults to `1` for new configs. Low impact.

4. **Preset mismatch:** The `reference` preset (57 linters) is the recommended starting point,
   but most real configs end up at 100+ linters (everything). Consider adding a documented
   "house" preset that matches the actual kitchen-sink regime most projects use, so
   `configure --preset house` reproduces the real-world standard exactly.

---

## 4. Recommended Action Plan (Pareto-ordered)

### Step 1 — Fix the silent-drop bug (P0, highest leverage)

File: `pkg/linter/fixer.go`

Make `updateGoVersion`, `updateRunnerSettings`, `updateBuildTags`,
`updateConfigFromSets`/`injectDefaultSettings` all report change counts. Sum all counts
before the `counts.total() == 0` guard. This single fix makes re-running `configure` heal
the 108 stale `ginkgolinter` + 105 stale `testifylint` configs automatically.

### Step 2 — Add `issues` block normalization (P0)

File: `pkg/linter/fixer_config.go`

Add `updateIssuesSettings(cfg *types.Config) int` that sets
`issues.max-issues-per-linter` and `issues.max-same-issues` to the loader defaults
(`50` / `10`) when they are zero/absent. Wire it into `applyAndSave`. This heals the 64
configs that currently hide duplicate CI issues.

### Step 3 — Normalize `run.timeout` (P2)

File: `pkg/linter/fixer_preflight.go`

Currently `needsDurationFix` only repairs _empty/invalid_ timeouts. Consider also offering
a `--normalize` flag (or making it default) that sets non-standard timeouts to `5m`.

### Step 4 — Consider a `house` preset (P3, strategic)

File: `pkg/constants/presets.go`

Capture the actual 100+ linter "kitchen sink" that most projects converge on, so the
explicit preset matches reality rather than the smaller `reference` set.

---

## 5. What the Fixer Already Does Well (no change needed)

These are working correctly and need no action — documented here to prevent regression:

- ✅ v1→v2 migration (zero v1 configs remain)
- ✅ Deprecated linter replacement (`wsl`→`wsl_v5`, `gomodguard`→`gomodguard_v2`, etc.)
- ✅ Core formatter injection (`gci`, `gofumpt`, `goimports`, `golines`)
- ✅ `swaggo` formatter auto-detection
- ✅ Redundant linter removal (`lll`→`golines`, `gofmt`→`gofumpt`)
- ✅ Go experiment build tags
- ✅ Generated-file exclusion via gogenfilter
- ✅ Default exclusion paths (`_templ.go`, `.gen.go`, `vendor/`)
- ✅ Default exclusion rules for `_test.go`
- ✅ `generated: lax` for both linters and formatters
- ✅ `golines.max-len: 120` default
- ✅ Go version auto-detection from toolchain
- ✅ Version-gated deprecation replacement (`gomodguard` only on v2.12.0+)

---

## 6. Methodology

A Python script (`/tmp/golangci-audit/aggregate.py`) walked all `~/projects/**/*.{yml,yaml}`
golangci-lint configs, parsed them with PyYAML, and aggregated: versions, enabled/disabled
linters, per-linter settings, formatters, run/issues/output blocks, and exclusions. A second
script (`gaps.py`) simulated the fixer's behavior on each config to count how often each fix
would still be needed. Vendored (`/vendor/`) and `archived/` configs were excluded from the
active dataset (136 configs) but included in the raw total (184).
