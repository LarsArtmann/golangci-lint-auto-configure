# Status Report — 2026-05-16 19:54 CEST

**Session Focus:** Architecture cleanup, SRP violations, domain method enrichment, type-safety improvements

---

## A. FULLY DONE

### 1. MergeExclusionPaths → types.Set[string] (commit `89803ec`)

Replaced 20-line manual map-based dedup in `pkg/gogenfilter/scanner.go:275-294` with existing `types.Set[string]`:

**Before:** 20 lines: manual map creation + append + sort
**After:** 4 lines: `types.NewSet(existing...)`, `Add()`, `ToSortedSlice()`

| Metric         | Before | After                              |
| -------------- | ------ | ---------------------------------- |
| Lines          | 20     | 4                                  |
| Imports added  | 0      | `types` (already in package graph) |
| Tests modified | 0      | 0                                  |

### 2. ConfigHealth Critical Linters Derivation (commit `2b0fb0f`)

**Problem:** `checkMissingCriticalLinters` hardcoded `["errcheck", "staticcheck", "govet"]`
**Solution:** Added `CriticalLinters()` in `pkg/constants` + new `CheckConfigHealthWithCriticalLinters` signature

| File                                 | Change                                                                         |
| ------------------------------------ | ------------------------------------------------------------------------------ |
| `pkg/constants/linter_priorities.go` | Added `CriticalLinters()` returning curated 3-linter set                       |
| `pkg/types/validation.go`            | New `CheckConfigHealthWithCriticalLinters` accepting `[]string` param          |
| `internal/cli/cmd_validate.go`       | Calls `CheckConfigHealthWithCriticalLinters(cfg, constants.CriticalLinters())` |
| `pkg/finding/helpers.go`             | Added `SeverityFromHealthSeverity` (eliminates 8-line manual switch in CLI)    |
| `pkg/finding/categories.go`          | Removed deprecated `gomodguard` mapping                                        |

**Key decision:** `CriticalLinters()` returns curated 3-liner set, not all `LinterPriorityCritical` linters. All 11 critical linters (including `paralleltest`, `sloglint`) would be too aggressive for health checks.

### 3. Silent Skip Fix in healthIssuesToFindings (commit `31cce92`)

**Bug:** `finding.NewBuilder(...).Build()` error → bare `continue` (data loss: health issue silently dropped from SARIF)
**Fix:** Threaded `*log.Logger` through `outputHealthSARIF` and `healthIssuesToFindings`:

```go
if err != nil {
    logger.Warnf("⚠️  Failed to build finding for rule %q: %v", issue.Rule, err)
    continue
}
```

### 4. ConfigHealth Domain Methods (commit `206c20a`)

Added 3 methods to `ConfigHealth` + migrated 7 test call sites:

| Method                                    | Purpose              | Tests         |
| ----------------------------------------- | -------------------- | ------------- |
| `IssuesByRule(rule string) []HealthIssue` | Filter by rule name  | ✅ Added test |
| `HasRule(rule string) bool`               | Check rule existence | ✅ Added test |
| `CountBySeverity(sev HealthSeverity) int` | Count by severity    | ✅ Added test |

Removed standalone `filterByRule` test helper (5 lines).

### 5. ProjectType.Preset() (commit `c9d617a`)

**Moved** `presetForProjectType()` from CLI layer to detection package as value method:

| Before (CLI)                                     | After (Detection)                             |
| ------------------------------------------------ | --------------------------------------------- |
| `presetForProjectType(projectType)`              | `projectType.Preset()`                        |
| `defaultPreset`, `presetStrict` constants in CLI | Eliminated (2 unused consts removed)          |
| No unit tests                                    | `TestProjectType_Preset` covering all 6 types |

**SRP violation fixed:** Preset selection is a property of project type, not a CLI concern.

### 6. ConfigAnalysis Convenience Methods (commit `ca6b00d`)

| Method                          | Use Case                             |
| ------------------------------- | ------------------------------------ |
| `TotalRecommendations() int`    | Reporting: "Found N recommendations" |
| `EnabledLinterNames() []string` | Logging, comparison, set operations  |

---

## B. PARTIALLY DONE

Nothing. All scope items completed.

---

## C. NOT STARTED

From the comprehensive plan, these were deprioritized:

1. **Extract `applyPreset` from `cmd_configure.go`** — File is now 395 lines (down from 420 after removing preset logic). Still above the 350-line target, but `applyPreset` depends on many CLI-local types. Requires careful extraction.
2. **//go:generate stringer for `HealthSeverity`** — Low impact; `String()` method already exists and covers all cases.
3. **Package doc comments for uncovered packages** — `pkg/report/` and `internal/cli/cmd/` already have basic package docs in existing files.
4. **`ConfigAnalysis.String()`** — Not needed yet; `Summary()` method in `pkg/ui/formatter.go` handles display.

---

## D. TOTALLY FUCKED UP

Nothing. All code builds, tests pass, lint is clean.

**Close call:** `CriticalLinters()` initially returned all 11 critical-priority linters, which broke integration tests (config with 4 linters suddenly flagged for missing 7). Fixed by returning curated 3-liner set.

---

## E. WHAT WE SHOULD IMPROVE

### Architecture

1. **`cmd_configure.go` still 395 lines** — Extract `applyPreset`, `runConfigure`, and preset flow to separate files. This is the largest file and has the most functions.

2. **`ConfigAnalysis` needs more convenience methods**:
   - `TotalLinters() int` — total enabled + disabled
   - `HasDeprecatedLinters() bool` — boolean for quick check
   - `IsEmpty() bool` — no analysis performed

3. **`HealthSeverity` should use `//go:generate stringer`** — Remove manual `String()` switch.

### Code Quality

4. **`HealthIssue` could have `WithSeverity`, `WithRule`, `WithMessage` builder methods** — matches go-finding builder pattern.

5. **`ConfigHealth.addIssue` is unexported** — Should it be exported for custom health rule extensions?

---

## F. Top #25 Things to Get Done Next

| #  | Priority | Task                                                                       | Value              |
| -- | -------- | -------------------------------------------------------------------------- | ------------------ |
| 1  | HIGH     | Extract `applyPreset` from `cmd_configure.go` to `cmd_configure_preset.go` | File size          |
| 2  | HIGH     | Extract `runConfigure` from `cmd_configure.go` to `cmd_configure_core.go`  | File size          |
| 3  | MEDIUM   | Add `//go:generate stringer` for `HealthSeverity`                          | Standard Go        |
| 4  | MEDIUM   | Add `ConfigAnalysis.TotalLinters()`, `HasDeprecatedLinters()`              | API completeness   |
| 5  | MEDIUM   | Create `FEATURES.md`                                                       | Project clarity    |
| 6  | MEDIUM   | Create `TODO_LIST.md`                                                      | Execution roadmap  |
| 7  | MEDIUM   | Improve `internal/cli` coverage (8.6%)                                     | Core path testing  |
| 8  | MEDIUM   | Improve `pkg/report` coverage (0%)                                         | Report reliability |
| 9  | MEDIUM   | Improve `pkg/finding` coverage (56.0%)                                     | Finding pipeline   |
| 10 | LOW      | Add `ConfigAnalysis.String()` debug method                                 | DX                 |
| 11 | LOW      | Document 0%-coverage packages                                              | Documentation      |
| 12 | LOW      | Add `HealthIssue` builder pattern                                          | Consistency        |
| 13 | LOW      | Tag `v0.1.0` release                                                       | Release mgmt       |
| 14 | LOW      | Add goreleaser config                                                      | Distribution       |
| 15 | LOW      | Evaluate nix flake-based releases                                          | Build system       |

---

## Metrics Summary

| Metric                     | Before Session                       | After Session                                  | Δ             |
| -------------------------- | ------------------------------------ | ---------------------------------------------- | ------------- |
| Composite Coverage         | 59.7%                                | 59.9-60.1%                                     | +0.2~+0.4%    |
| Files Over 350 Lines       | 14                                   | 13 (cmd_configure down)                        | -1            |
| Lint Issues                | 1 (gochecknoglobals)                 | 0                                              | ✅ Fixed      |
| Deprecated linter mappings | 1 (gomodguard)                       | 0                                              | ✅ Fixed      |
| ConfigHealth methods       | 2 (CriticalIssues, WarningIssues)    | 5 (+ IssuesByRule, HasRule, CountBySeverity)   | +3            |
| ConfigAnalysis methods     | 0                                    | 2 (+ TotalRecommendations, EnabledLinterNames) | +2            |
| Manual switch statements   | 2 (severity mapping, preset mapping) | 0                                              | ✅ Eliminated |
| Free test helpers          | 1 (`filterByRule`)                   | 0                                              | ✅ Eliminated |

---

## Commit Log (This Session)

```
ca6b00d feat(types): add TotalRecommendations and EnabledLinterNames to ConfigAnalysis
31cce92 fix(validate): log warning when health issue finding fails to build
c9d617a refactor(detection): add ProjectType.Preset() method; remove presetForProjectType from CLI
206c20a feat(types): add ConfigHealth domain methods; migrate tests to use IssuesByRule
2b0fb0f feat(health-checks): derive critical linters from constants; add SeverityFromHealthSeverity
89803ec refactor(gogenfilter): use types.Set[string] instead of manual map dedup in MergeExclusionPaths
```

---

_Generated at 2026-05-16 19:54 CEST_
