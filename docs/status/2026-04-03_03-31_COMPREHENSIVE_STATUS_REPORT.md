# Comprehensive Status Report

**Date:** 2026-04-03 03:31  
**Author:** Crush (AI Agent)  
**Branch:** master  
**Last Commit:** 29e4b54 — `docs(status): comprehensive status report 2026-04-03 01:48`

---

## Executive Summary

We are in the middle of a **large-scale lint violation cleanup** across the entire codebase. The project started with **51 lint issues** (24 funlen, 12 funcorder, 4 wrapcheck, 4 noinlineerr, plus exhaustruct, gochecknoglobals, goconst, godot, golines, nlreturn). After significant refactoring across **21 files** (+1246/-845 lines), we now have **58 lint issues** and **1 failing test**. The situation is worse than when we started — not because the approach is wrong, but because the refactoring was done too hastily without running `just lint && just test` between each change.

**Critical blocker:** `pkg/utils/retry.go` refactoring broke the "should fail after max retries" test. This must be fixed before any other work.

---

## A) FULLY DONE ✅

These items are complete and working:

| # | Item | Details |
|---|------|---------|
| 1 | **funlen: `pkg/migration/rules.go`** | Extracted 8 helper maps from `DefaultRules` |
| 2 | **funlen: `pkg/report/json_report_generator.go`** | Extracted `buildJSONReport`, `extractLinterNames` |
| 3 | **funlen: `pkg/ui/formatter.go`** | Extracted 6 helpers from `FormatRecommendations` |
| 4 | **funlen: `internal/cli/cmd/migrate.go`** | Extracted 3 helpers from `executeMigration` |
| 5 | **funlen: `pkg/migration/config_types.go`** | Extracted `migrateIssuesToTopLevel`, `moveIssuesFieldsToTemp` |
| 6 | **funlen: `pkg/migration/migrations.go`** | Extracted 6 helpers for migration logic |
| 7 | **funlen: `internal/cli/cmd_report.go`** | Extracted 3 helpers (but introduced new violations) |
| 8 | **funlen: `internal/cli/cmd_validate.go`** | Extracted 2 helpers (but introduced new violations) |
| 9 | **funlen: `pkg/diff/differ.go`** | Extracted 13 helpers (but introduced new violations) |
| 10 | **Status reports** | Two status reports written and committed |

---

## B) PARTIALLY DONE 🔧

These items were started but introduced new problems:

### 1. `pkg/utils/retry.go` — **BROKEN TEST** 🔥
- **What was done:** Extracted `handleRetry`, `waitWithBackoff`, `handleRetryParams`
- **What broke:** The refactoring changed the control flow. The original code had 3 distinct return paths (success, retry interrupted, failed after max retries, non-retryable). The new code conflates these, causing the "should fail after max retries" test to fail.
- **Root cause:** When `shouldRetry` returns true AND `attempt >= maxRetries`, the original code fell through to `return output, fmt.Errorf("failed after %d retries: %w", ...)`. The new `handleRetry` returns `(false, nil)` in this case, and the loop falls through to the wrong return.
- **Fix needed:** Revert to original logic or fix the `handleRetry` function to properly handle all cases.

### 2. `examples/api-usage/main.go` — funlen fixed, new issues
- Extracted 7 helpers from `main()` ✅
- **NEW:** exhaustive switch (missing Medium/Optional cases) ❌
- **NEW:** golines formatting ❌

### 3. `pkg/config/loader.go` — funlen fixed, new issues
- Extracted `fetchLintersWithFallback`, `detectGoVersion` ✅
- **NEW:** 2 funcorder violations (unexported methods after exported) ❌

### 4. `pkg/detection/detector.go` — funlen fixed, new issues
- Extracted 6 helpers from `HasSwaggo` ✅
- **NEW:** noinlineerr, wrapcheck violations persist ❌
- **NEW:** wsl_v5 violation ❌

### 5. `pkg/detection/detector_test.go` — funlen partially fixed
- Extracted helpers but `buildDetectTestCases` still 36 > 30 lines ❌
- **NEW:** thelper, varnamelen, noinlineerr violations ❌

### 6. `pkg/detection/detector_bench_test.go` — funlen fixed, new issues
- Extracted 2 helpers ✅
- **NEW:** 2 thelper violations ❌

### 7. `pkg/diff/differ_test.go` — funlen partially fixed
- Extracted test case types ✅
- `buildFormatChangesTests` (40>30) and `buildGetSummaryTests` (33>30) still too long ❌
- **NEW:** gochecknoglobals, golines, prealloc violations ❌

### 8. `pkg/migration/config_types.go` — funlen fixed, new issues
- Extracted 2 helpers ✅
- **NEW:** nolintlint (gocognit no longer needed), golines ❌

### 9. `pkg/migration/migrations.go` — funlen fixed, new issues
- Extracted 6 helpers ✅
- **NEW:** gochecknoglobals on `formatterNames` (was extracted to package level) ❌

---

## C) NOT STARTED ⏳

These violations existed before our changes and have not been touched:

| # | File | Line | Linter | Description |
|---|------|------|--------|-------------|
| 1 | `pkg/linter/fixer.go` | 194 | exhaustruct | `fixCounts{}` missing 4 fields |
| 2 | `pkg/linter/fixer.go` | 295 | gochecknoglobals | `goExperimentTags` package-level var |
| 3 | `pkg/linter/fixer.go` | 201 | golines | Line too long |
| 4 | `pkg/linter/fixer_formatters.go` | 158 | godot | Comment missing period |
| 5 | `pkg/linter/fixer_formatters.go` | 57,93 | nlreturn | Missing blank line before break/return |
| 6 | `pkg/linter/analyzer.go` | — | funcorder | 4 unexported methods after exported |
| 7 | `internal/cli/cmd_analyze.go` | 75 | wrapcheck | Unwrapped error from external package |
| 8 | `internal/cli/cmd_configure.go` | 64 | goconst | `standard` string appears 3 times |
| 9 | `internal/cli/cmd_report.go` | 96,116 | noinlineerr | 2 inline error assignments |
| 10 | `internal/cli/cmd_validate.go` | 59 | noinlineerr | 1 inline error assignment |

---

## D) TOTALLY FUCKED UP 💥

### 1. `pkg/utils/retry.go` — TEST FAILURE + REGRESSION
**Severity: CRITICAL — blocks all other work**

The refactoring from a clean, well-understood function into extracted helpers broke the observable behavior:

**Original behavior (working):**
```go
// When shouldRetry=true AND attempt >= maxRetries:
return output, fmt.Errorf("%s failed after %d retries: %w", name, config.MaxRetries, lastErr)

// When context canceled during backoff:
return nil, fmt.Errorf("%s retry interrupted (lastErr=%w): %w", name, lastErr, ctx.Err())

// When shouldRetry=false:
return output, err
```

**New behavior (broken):**
```go
// handleRetry returns (false, nil) when !shouldRetry || attempt >= maxRetries
// The loop continues, incrementing backoff, then falls through to:
return nil, fmt.Errorf("%s failed after %d retries: %w", name, config.MaxRetries, lastErr)
// But "persistent error" is returned, NOT "failed after 2 retries" because
// the original code returned output, err for non-retryable errors — now it wraps differently
```

**Test failure:**
```
Expected <string>: persistent error
to contain substring <string>: failed after 2 retries
```

### 2. Whack-a-mole pattern — New violations from fixes
We extracted functions to fix funlen, which introduced:
- **14 funcorder** violations (unexported methods placed before exported)
- **7 gochecknoglobals** violations (test tables moved to package level)
- **5 golines** formatting violations (new code not formatted)
- **2 nonamedreturns** violations
- **2 thelper** violations
- Multiple noinlineerr, wsl_v5, varnamelen violations

### 3. No incremental testing
The entire refactoring was done across 21 files without running `just lint && just test` between changes. This made it impossible to catch regressions early.

---

## E) WHAT WE SHOULD IMPROVE 📈

### Process Improvements

| # | Improvement | Why |
|---|------------|-----|
| 1 | **One file at a time** | Fix, test, lint, commit per file. Prevents cascading failures. |
| 2 | **Run `just lint && just test` after EVERY change** | Catch regressions immediately, not after 21 files. |
| 3 | **Don't extract to package-level globals** | Test tables should stay inside test functions or use `var _ = Describe()` pattern. |
| 4 | **Preserve exact behavior when extracting** | Copy the original return statements verbatim, don't restructure control flow. |
| 5 | **Check funcorder before committing** | Extracted unexported methods go AFTER the last exported method on the struct. |
| 6 | **Run `golines --write` before committing** | Formatting is trivial to fix but clutters the lint output. |
| 7 | **Use `//nolint:funlen` for genuinely complex functions** | Retry logic with context cancellation is inherently complex. Extracting it breaks semantics. |
| 8 | **Commit after each file** | Git is our undo button. We should use it. |

### Technical Improvements

| # | Improvement | Why |
|---|------------|-----|
| 1 | **Add pre-commit hook for lint** | Catch issues before they accumulate |
| 2 | **Configure golangci-lint with `fix` command** | Auto-fix trivial issues (golines, godot, nlreturn) |
| 3 | **Use `testify/suite` or table-driven tests** | Reduce boilerplate in test files |
| 4 | **Review golangci.yml config** | Some linters may be too strict for the project's current state |

---

## F) Top #25 Things We Should Get Done Next

### Priority 1: Fix the broken test (BLOCKER)
| # | Task | File | Effort |
|---|------|------|--------|
| 1 | **Revert `pkg/utils/retry.go` to original form and add `//nolint:funlen`** | `pkg/utils/retry.go` | 5 min |

### Priority 2: Fix newly introduced violations (our mess)
| # | Task | File | Effort |
|---|------|------|--------|
| 2 | Fix 2 funcorder violations | `pkg/config/loader.go` | 5 min |
| 3 | Fix 8 funcorder violations | `pkg/detection/detector.go` | 10 min |
| 4 | Fix 4 funcorder violations | `pkg/linter/analyzer.go` | 10 min |
| 5 | Add missing switch cases or default | `examples/api-usage/main.go` | 3 min |
| 6 | Remove unused `label` parameter | `pkg/diff/differ.go` | 2 min |
| 7 | Remove named returns from 2 functions | `pkg/diff/differ.go` | 2 min |
| 8 | Add `b.Helper()` to 2 bench helpers | `pkg/detection/detector_bench_test.go` | 2 min |
| 9 | Move `formatterNames` back to local scope | `pkg/migration/migrations.go` | 3 min |
| 10 | Remove `gocognit` from nolint directive | `pkg/migration/config_types.go` | 1 min |
| 11 | Fix wsl_v5 violations (2 locations) | `pkg/detection/detector.go`, `pkg/utils/retry.go` | 3 min |
| 12 | Fix remaining funlen in test helpers | `pkg/detection/detector_test.go`, `pkg/diff/differ_test.go` | 15 min |

### Priority 3: Fix pre-existing violations
| # | Task | File | Effort |
|---|------|------|--------|
| 13 | Fix exhaustruct on `fixCounts{}` | `pkg/linter/fixer.go` | 3 min |
| 14 | Fix gochecknoglobals on `goExperimentTags` | `pkg/linter/fixer.go` | 5 min |
| 15 | Fix 2 nlreturn violations | `pkg/linter/fixer_formatters.go` | 2 min |
| 16 | Fix godot on comment | `pkg/linter/fixer_formatters.go` | 1 min |
| 17 | Fix 3 noinlineerr violations | `cmd_report.go`, `cmd_validate.go` | 5 min |
| 18 | Fix goconst `standard` string | `cmd_configure.go` | 2 min |
| 19 | Fix wrapcheck on `AnalyzeConfig` | `cmd_analyze.go` | 2 min |
| 20 | Fix wrapcheck on `os.Open`, `scanner.Err()`, `filepath.Walk` | `detector.go` | 5 min |

### Priority 4: Formatting and cleanup
| # | Task | File | Effort |
|---|------|------|--------|
| 21 | Run `golines` on all affected files | 5 files | 2 min |
| 22 | Fix varnamelen in test files | `detector_test.go` | 3 min |
| 23 | Fix noinlineerr in test files | `detector_test.go` | 5 min |
| 24 | Fix gochecknoglobals in test files | `differ_test.go` | 10 min |
| 25 | Write final status report and commit everything | — | 5 min |

---

## G) Top #1 Question 🤔

### Should we `//nolint:funlen` the retry function?

The `pkg/utils/retry.go` `WithRetry` function is **inherently complex** — it has 3 distinct control flow paths (success, retry-with-backoff, context-cancelation, max-retries-exceeded, non-retryable-error). Every extraction attempt has broken one or more tests because the error wrapping semantics are subtle.

**My recommendation:** Yes, add `//nolint:funlen` and **revert to the original implementation**. The original function was 40 lines — only 10 over the limit — and was well-structured with clear comments. The extracted version has MORE total lines (101 vs 40) and is harder to understand.

**The alternative** is to carefully study all 4 test cases, understand the exact error wrapping expectations, and do a surgical extraction. But this has already been tried and failed twice in the previous session.

---

## Current Lint Violations Summary

**58 total violations across 17 categories:**

| Linter | Count | Files |
|--------|-------|-------|
| funcorder | 14 | loader.go, detector.go, analyzer.go |
| gochecknoglobals | 7 | detector_test.go, differ_test.go, fixer.go, migrations.go |
| noinlineerr | 9 | detector.go, detector_test.go, cmd_report.go, cmd_validate.go |
| golines | 5 | api-usage, differ_test, fixer.go, config_types.go |
| wrapcheck | 5 | cmd_analyze.go, detector.go (3), retry.go |
| varnamelen | 3 | detector_test.go (2), retry.go |
| nonamedreturns | 2 | differ.go |
| thelper | 2 | detector_bench_test.go |
| wsl_v5 | 2 | detector.go, retry.go |
| nlreturn | 2 | fixer_formatters.go |
| exhaustive | 1 | api-usage/main.go |
| exhaustruct | 1 | fixer.go |
| goconst | 1 | cmd_configure.go |
| godot | 1 | fixer_formatters.go |
| nolintlint | 1 | config_types.go |
| revive | 1 | differ.go |
| funlen | 1 | (will be 0 if we nolint retry.go) |

## Test Status

- **1 FAILING:** `pkg/utils/retry_test.go:69` — "should fail after max retries"
- **15 PASSING** in utils suite
- **All other suites passing** (8 of 9 suites green)

## Git Status

- **21 modified files** — +1246/-845 lines
- **Nothing committed** from this session's refactoring work
- **Branch:** master
- **No push done**
