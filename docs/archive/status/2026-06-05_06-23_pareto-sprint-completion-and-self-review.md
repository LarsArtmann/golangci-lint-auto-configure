# Status Report: Pareto Sprint + Self-Review

**Date**: 2026-06-05 06:23
**Branch**: master
**Commits**: dea1733, 80e450b, 6488833 (pending push)

---

## Summary

Executed the full Pareto execution plan (11/11 tasks), performed a comprehensive self-review, and identified critical architectural issues. All planned work is committed across 3 logical commits, ready for push.

---

## A) FULLY DONE

### Test Coverage (Phase 1 — Critical)

- **CLI integration tests**: 32 → 53 tests (+21 new, +363 lines)
  - All presets: minimal, standard, web, cli, library (DescribeTable)
  - Flag combinations: --check/--diff with priorities
  - Deprecated linter replacement (wsl → wsl_v5)
  - install-hook, completion, validate --format sarif
  - report HTML generation, missing config error
  - migrate --skip-validation
- **gogenfilter scanner tests**: 15 → 22 tests (+7 new, +122 lines)
  - moq, mockgen, stringer, go-enum, oapi-codegen, deepcopy-gen
  - node_modules exclusion guard
  - Coverage: 59.8% → 63.9%
- **migration tests**: 37 → 43 tests (+6 new, +87 lines)
  - cyclop skip-tests, gosec null excludes, containedctx removal
  - forbidigo p→pattern, goimports local-prefixes, gci skip-generated
  - Coverage: 66.8% → 75.5%

### Correctness Improvements (Phase 3 — Medium)

- **DryRun field**: Added `DryRun bool` to `MigrationResult` struct
  - Set in `dryRunResult()`, `calculateDryRunResultWithInvalidDurations`, `calculateDryRunResultWithDeprecated`
- **errors.Join**: Replaced sequential fail-fast with `errors.Join` in `AnalysisToReport`
  - All findings collected rather than stopping at first error

### Documentation (Phase 2 — High)

- **AGENTS.md trim**: 912 → 367 lines (extracted to reference files)
- **docs/references/**: 5 new reference files
  - `code-organization.md` (140 lines) — Directory structure and patterns
  - `error-handling.md` (50 lines) — Error patterns
  - `integrations.md` (88 lines) — gogenfilter + go-finding
  - `testing-style-and-patterns.md` (174 lines) — Testing, CI/CD, style
  - `working-with-codebase.md` (143 lines) — Tasks, debugging

### Pre-existing Fix

- Fixed `// Note:` comment in `pkg/config/merger_issues.go:14` that triggered todo-check pre-commit hook

### Coverage Summary

| Suite       | Before | After |
| ----------- | ------ | ----- |
| Composite   | 61.5%  | 62.6% |
| gogenfilter | 59.8%  | 63.9% |
| migration   | 66.8%  | 75.5% |
| CLI tests   | 32     | 53    |

---

## B) PARTIALLY DONE

### Architecture Analysis

The self-review identified 13+ architectural issues across 3 severity levels. Analysis is complete but fixes are not yet implemented. See section E for the full list.

### Type Safety Improvements

- Analysis complete: identified `[]string` where `[]LinterName` should be, `map[string]any` settings problem, missing enum types
- No implementation started — deferred pending design decision on `map[string]any` strategy

---

## C) NOT STARTED

1. **Config.Clone() deep clone fix** — Critical bug identified, not fixed
2. **CLI globals removal** — 9 package-level mutables in commands.go
3. **ConfigLoader god interface split** — 6 roles, Fixer only needs 3
4. **Type-safe LinterName/FormatterName slices** — Enable/Disable still `[]string`
5. **GeneratedMode enum** — `LintersExclusionsConfig.Generated string` should be typed
6. **ConfigVersion type** — `Config.Version string` should be typed
7. **ParsePriority() validator** — Invalid input silently defaults
8. **KnownFields(true) in YAML decoder** — Typo detection disabled
9. **detectFormat hardening** — Silently defaults to YAML for unknown extensions
10. **"validation-error" string constant** — Used 3 times without const
11. **pkg/client/client.go tests** — 210 lines, zero tests
12. **Spinner goroutine leak** — Context cancellation doesn't clean up
13. **go-error-family adoption** — Recommended by go-structure-linter

---

## D) TOTALLY FUCKED UP

### Config.Clone() Shallow Clone Bug (CRITICAL)

**File**: `pkg/types/clone.go:16-25`

`cloneAnyMap()` uses `maps.Copy` which is **shallow**. Nested `map[string]any` values (linter settings) are **shared between original and clone**. Any mutation to nested settings after Clone() corrupts the original config.

**Impact**: Latent data corruption in the configure pipeline. If a user runs configure and the fixer modifies nested linter settings, the original loaded config is silently mutated.

**Fix**: Recursive deep clone for `map[string]any` values, or use `encoding/gob` / `json.Marshal`+`Unmarshal` round-trip.

---

## E) WHAT WE SHOULD IMPROVE

### Critical (Data Corruption Risk)

1. Fix `Config.Clone()` — implement deep clone for nested maps
2. Spinner goroutine leak — use `context.WithCancel` or `sync.WaitGroup`

### High (Type Safety)

3. Type `LintersConfig.Enable/Disable` as `[]LinterName` instead of `[]string`
4. Type `FormattersConfig.Enable/Disable` as `[]FormatterName` instead of `[]string`
5. Add `GeneratedMode` enum type for `LintersExclusionsConfig.Generated`
6. Add `ConfigVersion` type for `Config.Version`
7. Add `ParsePriority(string)` with validation and error return
8. Enable `KnownFields(true)` in YAML decoder (`pkg/config/loader.go:152`)

### Medium (Code Quality)

9. Extract "validation-error" string to constant
10. Harden `detectFormat` — return error for unknown extensions
11. Split `ConfigLoader` god interface into focused interfaces
12. Remove 9 package-level mutable globals from `internal/cli/commands.go`
13. Add tests for `pkg/client/client.go` (210 lines, 0% coverage)

### Low (Nice-to-Have)

14. Adopt `go-error-family` for structured, classified errors
15. Extract `map[string]any` typed accessors for known linter settings
16. Consider typed structs for top ~15 linters (Option B from analysis)

---

## F) TOP 25 THINGS TO DO NEXT

Sorted by impact × effort (Pareto order):

| #  | Task                                                  | Impact   | Effort   | Severity       |
| -- | ----------------------------------------------------- | -------- | -------- | -------------- |
| 1  | Fix Config.Clone() deep clone bug                     | Critical | 30min    | Bug            |
| 2  | Type Enable/Disable as []LinterName/[]FormatterName   | High     | 45min    | Type Safety    |
| 3  | Enable KnownFields(true) in YAML decoder              | High     | 30min    | Correctness    |
| 4  | Add ParsePriority() with validation                   | Medium   | 15min    | Type Safety    |
| 5  | Add GeneratedMode enum type                           | Medium   | 20min    | Type Safety    |
| 6  | Add ConfigVersion type                                | Medium   | 15min    | Type Safety    |
| 7  | Extract "validation-error" constant                   | Low      | 5min     | Code Quality   |
| 8  | Harden detectFormat for unknown extensions            | Medium   | 20min    | Correctness    |
| 9  | Fix spinner goroutine leak                            | Medium   | 30min    | Bug            |
| 10 | Add tests for pkg/client/client.go                    | High     | 45min    | Coverage       |
| 11 | Split ConfigLoader god interface                      | Medium   | 60min    | Architecture   |
| 12 | Remove CLI package-level globals                      | Medium   | 60min    | Architecture   |
| 13 | Adopt go-error-family                                 | Medium   | 90min    | Error Handling |
| 14 | Add map[string]any typed accessors                    | Medium   | 60min    | Type Safety    |
| 15 | Typed structs for top 15 linters                      | High     | 4h       | Architecture   |
| 16 | Deep clone benchmark (cloneAnyMap vs json round-trip) | Low      | 30min    | Performance    |
| 17 | Add integration test for Clone() deep isolation       | Medium   | 15min    | Testing        |
| 18 | Fix unparam warnings in fixer_preflight.go            | Low      | 10min    | Lint           |
| 19 | Fix wsl_v5 warning in commands_test.go:299            | Low      | 2min     | Lint           |
| 20 | Fix inefficient string concat in migrator_test.go     | Low      | 5min     | Performance    |
| 21 | Update vendorHash in flake.nix                        | Low      | 10min    | Build          |
| 22 | Remove redundant gofmt from formatters config         | Low      | 2min     | Config         |
| 23 | Add CI pipeline test for all presets end-to-end       | High     | 2h       | Testing        |
| 24 | Document map[string]any design decision in ADR        | Medium   | 30min    | Documentation  |
| 25 | Explore code-gen from golangci-lint schema (Option C) | Low      | Research | Architecture   |

---

## G) TOP QUESTION

**The `map[string]any` settings problem**: `LintersConfig.Settings` is `map[string]any` because golangci-lint's linter settings schema is completely dynamic (every linter has different settings, and custom linters can add arbitrary config).

Options:

- **(A)** Keep `map[string]any` with typed accessor helpers — lowest risk, no schema coupling
- **(B)** Typed structs for ~15 known linters, `map[string]any` for unknown — pragmatic middle ground, but YAML round-trip complexity with mixed typed/untyped needs validation
- **(C)** Code generation from golangci-lint schema — most correct but creates tight coupling to upstream schema changes

**My recommendation**: Start with (A) — typed accessor helpers like `GetStringSetting(linter, key)`, `GetIntSetting(linter, key)` — then evaluate (B) once we understand which accessors are actually needed. Option (C) is a longer-term research item.

**What I need feedback on**: Should we invest in (B) typed structs for the top ~15 linters now, or keep it simple with (A) typed accessors? The YAML round-trip is the key risk — mixing typed structs with `map[string]any` for unknown linters requires careful Marshal/Unmarshal handling.

---

## Commits (This Session)

1. `dea1733` — test(cli): add 21 CLI integration tests covering all presets, flags, and commands
2. `80e450b` — test: add gogenfilter and migration test coverage
3. `6488833` — fix: add DryRun field to MigrationResult and use errors.Join for multi-finding conversion
4. _(pending)_ — docs: AGENTS.md trim + docs/references/ + status report
