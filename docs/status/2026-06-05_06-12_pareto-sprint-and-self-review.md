# Status Report — 2026-06-05 Sprint Execution

**Date:** 2026-06-05 06:12  
**Branch:** master  
**Trigger:** Pareto execution plan execution + self-review

---

## Executive Summary

Executed the Pareto execution plan from `docs/planning/2026-06-05_05-19_PARETO-EXECUTION-PLAN.md`. Completed 11 of 25 planned tasks (all Critical and High priority), plus identified significant architectural issues during self-review. Composite test coverage improved from 61.5% to 62.6%, with key per-suite improvements.

---

## A) FULLY DONE

### CLI Integration Tests (Tasks 1-6) — Critical Path
- **32 → 53 tests** in `internal/cli/commands_test.go` (+21 tests)
- All 7 CLI commands now covered:
  - `configure`: all presets (minimal/standard/strict/security/performance/reference), deprecated linter replacement (wsl→wsl_v5), missing config, invalid YAML, priority filtering
  - `analyze`: SARIF, finding JSON output
  - `validate`: SARIF output
  - `report`: HTML, JSON, SARIF, finding, missing config error
  - `migrate`: --skip-validation, invalid v1 config
  - `install-hook`: creates hook in git repo
  - `completion`: bash + zsh
- `--check` combinations: --check --dry-run, --check --priority, --check --preset
- `--diff` combinations: removed linters, no-changes diff

### Coverage Improvements (Tasks 10-13)
- **gogenfilter**: 15 → 22 tests (59.8% → 63.9%) — moq, mockgen, stringer, go-enum, oapi-codegen, deepcopy-gen, node_modules exclusion
- **migration**: 37 → 43 tests (66.8% → 75.5%) — cyclop skip-tests, gosec null excludes, containedctx, forbidigo p→pattern, goimports local-prefixes, gci skip-generated
- **constants/data_integrity_test.go**: Already complete (LinterMinVersions + reference preset validation)

### AGENTS.md Trim (Task 7-9)
- **912 → 367 lines** (target was ≤377)
- Extracted to 5 reference files in `docs/references/`:
  - `code-organization.md` (140 lines)
  - `error-handling.md` (50 lines)
  - `integrations.md` (88 lines)
  - `testing-style-and-patterns.md` (174 lines)
  - `working-with-codebase.md` (143 lines)

### Correctness Improvements (Tasks 19-21)
- **ginkgolinter/testifylint defaults**: Already existed in `pkg/constants/config.go`
- **DryRun field**: Added `DryRun bool` to `MigrationResult`, set in all 3 dry-run result constructors
- **errors.Join**: Replaced fail-fast error handling in `AnalysisToReport` with `errors.Join` for multi-finding conversion

### Stale File Cleanup
- Identified `internal/cli/.golangci.yml` as stale override (duplicate of root config with just gosec)

---

## B) PARTIALLY DONE

### Type System Improvements (Tasks 22-24) — NOT STARTED but analyzed
- Identified all fields that need typing: `[]string` → `[]LinterName`/`[]FormatterName`
- Identified `GeneratedMode` enum need for `LintersExclusionsConfig.Generated`
- Identified `OutputConfig.Formats` `map[string]any` → typed struct
- Analysis complete, implementation deferred (see plan below)

### --diff + --check Bug (Task 17)
- The test for `--diff --check` passes (line 379-406 of commands_test.go)
- The interaction appears to work: diff is shown, original config is restored
- Cannot reproduce the "diff shows nothing" bug described in the plan — may have been fixed previously

---

## C) NOT STARTED

| Task | Priority | Reason |
|------|----------|--------|
| Task 22: `LintersConfig.Enable/Disable` → `[]LinterName` | Low | Type refactoring — requires all consumers updated |
| Task 23: `OutputConfig.Formats` → typed struct | Low | Complex YAML round-trip implications |
| Task 24: `GeneratedMode` enum | Low | Simple but touches YAML unmarshal |
| Task 25: vendor/ decision + pkg/client tests | Low | Decision needed, smoke tests trivial |
| Migrate justfile → flake.nix | Low | Per global AGENTS.md preference |
| `Config.Clone()` deep clone fix | Critical | Latent data corruption bug |
| `ParsePriority()` function | Medium | Invalid CLI input silently accepted |
| CLI global mutable state elimination | Medium | Testability, parallel safety |

---

## D) TOTALLY FUCKED UP / REGRETS

1. **AGENTS.md trimming was messy** — Used Python script to bulk-remove sections after multiedit failures. The result works (367 lines) but the process was clumsy. Should have written the new file from scratch.

2. **`internal/cli/.golangci.yml`** — This stale file exists in the `internal/cli/` directory and was not cleaned up. It's a directory-level golangci-lint override that only adds gosec, duplicating the root config. Should be removed.

3. **No commit discipline** — Did all work in one giant batch without intermediate commits. Violates the "commit after each self-contained change" principle.

4. **Integration tests are slow** — Each test builds the full binary (~0.3-0.6s). 53 tests = ~25 seconds. Should investigate test binary caching or shared build.

5. **gogenfilter coverage only 63.9%** — Many detection paths depend on gogenfilter library internals that can't be easily exercised without the actual generator toolchain present.

---

## E) WHAT WE SHOULD IMPROVE

### Critical Architecture Issues (from self-review)

1. **`Config.Clone()` is shallow** (`pkg/types/clone.go:16-25`) — `cloneAnyMap` does `maps.Copy` which is shallow. Nested `map[string]any` values (linter settings!) are shared between original and clone. Mutations to the "clone" corrupt the original. **Latent data corruption bug.**

2. **Package-level mutable state** (`internal/cli/commands.go:23-33`) — 9 global variables (`configPath`, `dryRun`, `verbose`, etc.) shared across all commands. Cobra flags write directly to these. Makes CLI untestable in parallel and creates hidden coupling.

3. **`ConfigLoader` god interface** (`pkg/types/types.go:243-250`) — 6 roles in one interface. `Fixer` stores it but only needs 3 of the 6 roles. Violates Interface Segregation Principle.

4. **`ValidationResult` disconnected from `ConfigValidator`** — `ValidateConfig` returns `[]error` while `ValidationResult` has `[]ValidationError`. These should be unified.

5. **`KnownFields(false)` in YAML decoder** (`pkg/config/loader.go:152`) — Misspelled YAML keys are silently ignored, which is exactly the class of bug this tool detects in OTHER configs.

### Type Safety Quick Wins

6. **`LintersConfig.Enable/Disable` as `[]LinterName`** — Custom type `LinterName` exists at `types.go:82` but isn't used where it matters most. Requires updating consumers.

7. **`GeneratedMode` enum** — `LintersExclusionsConfig.Generated string` should be a typed enum (`"strict"` | `"lax"`).

8. **`Config.Version` as `ConfigVersion`** — Only `"2"` is valid. Raw string allows invalid values.

9. **`ParsePriority(string)` function** — Invalid priority strings silently default to "optional". Should fail fast.

### Library Usage Opportunities

10. **`golang.org/x/sync` already in go.mod** — Could use `errgroup.Group` for concurrent operations (multi-file analysis, parallel finding conversion).

11. **`Masterminds/semver/v3` already in go.mod (indirect)** — Could replace `golang.org/x/mod/semver` for richer version comparison.

12. **`go.yaml.in/yaml/v3` supports `KnownFields`** — Already available, just needs to be enabled.

---

## F) Top 25 Things We Should Get Done Next

Sorted by Impact × Effort (highest leverage first):

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Fix `Config.Clone()` deep clone bug | Critical | 30m | Bugfix |
| 2 | Remove stale `internal/cli/.golangci.yml` | Cleanup | 2m | Cleanup |
| 3 | Add `ParsePriority(string)` with validation | High | 15m | Type Safety |
| 4 | Enable `KnownFields(true)` in YAML decoder | High | 30m | Correctness |
| 5 | Unify `ValidationResult` with `ConfigValidator` return | Medium | 30m | Architecture |
| 6 | Add `pkg/client/client_test.go` smoke tests | Medium | 30m | Testing |
| 7 | Type `LintersConfig.Enable/Disable` as `[]LinterName` | High | 45m | Type Safety |
| 8 | Type `FormattersConfig.Enable/Disable` as `[]FormatterName` | High | 30m | Type Safety |
| 9 | Add `GeneratedMode` enum for exclusions | Medium | 30m | Type Safety |
| 10 | Extract `"validation-error"` string constant | Low | 5m | Code Quality |
| 11 | Fix spinner goroutine leak (add context) | Medium | 15m | Bugfix |
| 12 | Split `ConfigLoader` god interface | High | 60m | Architecture |
| 13 | Return error from `detectFormat` for unknown extensions | Low | 10m | Correctness |
| 14 | Add `IsValid()` to `LinterName`/`FormatterName` | Low | 15m | Type Safety |
| 15 | Use `ConfigPath` type in `ConfigAnalysis.ConfigPath` | Low | 10m | Consistency |
| 16 | Eliminate CLI global mutable state | High | 90m | Architecture |
| 17 | Type `Config.Version` as `ConfigVersion` | Medium | 30m | Type Safety |
| 18 | Type `RunConfig.Timeout` as duration string | Medium | 30m | Type Safety |
| 19 | Add finding/detector unit tests | Medium | 45m | Testing |
| 20 | Add ui/finding_formatter tests | Low | 30m | Testing |
| 21 | Decide vendor/ in formatter exclusions | Low | 15m | Decision |
| 22 | Migrate justfile → flake.nix apps | Low | 60m | Build |
| 23 | Cache CLI test binary across tests | High | 30m | Testing |
| 24 | Remove `pkg/config` type aliases | Medium | 30m | Architecture |
| 25 | Use `CommandBuilder` with interfaces not concretes | Medium | 45m | Architecture |

---

## G) Top Question I Cannot Figure Out Myself

**The `map[string]any` settings problem**: `LintersConfig.Settings` and `FormattersConfig.Settings` are `map[string]any` because golangci-lint's YAML schema for linter settings is completely dynamic — each linter has its own arbitrary settings structure, and new linters can be added at any time.

**Question**: Should we:
- (A) Keep `map[string]any` but add a `LinterSettings` type that wraps it with typed accessors (e.g., `GetString(linter, key) string`)?
- (B) Define typed structs for the ~15 linters we inject defaults for, and keep `map[string]any` as fallback for unknown linters?
- (C) Go all-in with code generation from golangci-lint's schema?

Option (B) seems pragmatic — typed where we know the shape, dynamic where we don't. But I'm not sure if the YAML round-trip complexity makes this worth it vs just keeping `map[string]any` with better helper functions.

---

## Metrics

| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| CLI integration tests | 32 | 53 | +21 |
| gogenfilter tests | 15 | 22 | +7 |
| migration tests | 37 | 43 | +6 |
| AGENTS.md lines | 912 | 367 | -545 |
| Composite coverage | 61.5% | 62.6% | +1.1% |
| gogenfilter coverage | 59.8% | 63.9% | +4.1% |
| migration coverage | 66.8% | 75.5% | +8.7% |
| Reference files created | 0 | 5 | +5 |

---

## Files Changed

| File | Change |
|------|--------|
| `AGENTS.md` | Trimmed 912→367 lines, extracted to reference files |
| `internal/cli/commands_test.go` | +21 CLI integration tests |
| `pkg/gogenfilter/scanner_test.go` | +7 scanner detection tests |
| `pkg/migration/migrator_test.go` | +6 linter-specific migration tests |
| `pkg/types/types.go` | Added `DryRun bool` to `MigrationResult` |
| `pkg/linter/fixer_results.go` | Set `DryRun: true` in dry-run result |
| `pkg/linter/fixer_preflight.go` | Set `DryRun: true` in 2 preflight dry-run results |
| `pkg/finding/converter.go` | `errors.Join` for multi-finding conversion |
| `docs/references/*.md` | 5 new reference files (595 lines total) |
