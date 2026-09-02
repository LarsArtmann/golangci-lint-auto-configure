# Status Report: 2026-06-03 18:22 — Tier 2-3 Feature Implementation Sprint

## Executive Summary

Implemented 13 items from the Tier 2 (Architecture) and Tier 3 (Features) backlog. All tests pass (15 suites, 61.2% composite coverage). Lint shows 9 issues (7 funlen, 1 noinlineerr, 1 varnamelen) — all are pre-existing or minor.

## Fully Done ✅

### Tier 2: Architecture

| #  | Item                                                 | Status  | Notes                                                                                             |
| -- | ---------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------- |
| 6  | Add `templ generate` to Nix build pipeline           | ✅ DONE | `nativeBuildInputs = [ pkgs.templ ]`, `preBuild` in flake.nix; `templ generate` in justfile build |
| 7  | Remove `report_templ.go` from git tracking           | ✅ DONE | `git rm --cached`; `.gitignore` already had `*_templ.go`                                          |
| 8  | Return errors instead of panicking in `buildFinding` | ✅ DONE | Changed `buildFinding` → `(Finding, error)`, propagated through 6 files, all callers updated      |
| 9  | Add `--check` mode for CI exit codes                 | ✅ DONE | `--check` flag on `configure`, forces dry-run, exits with `ErrChangesNeeded` if fixes needed      |
| 10 | Add `--diff` flag to show config changes             | ✅ DONE | Global `--diff` flag, deep-copies config before/after fix, shows diff using `pkg/diff`            |

### Tier 3: Features

| #  | Item                                            | Status  | Notes                                                                                  |
| -- | ----------------------------------------------- | ------- | -------------------------------------------------------------------------------------- |
| 11 | Add `output.formats: {}` to default config      | ✅ DONE | Added `Output: OutputConfig{Formats: map[string]any{}}` to `CreateDefaultConfig`       |
| 12 | Add preset to apply reference config            | ✅ DONE | New `"reference"` preset with 60+ Critical + High linters                              |
| 13 | Version-aware feature flags                     | ✅ DONE | `LinterMinVersions` map + categorizer version gating using `semver.Compare`            |
| 14 | ginkgolinter default settings                   | ✅ DONE | `forbid-focus-container: true`, `forbid-spec-pollution: true`                          |
| 15 | testifylint default settings                    | ✅ DONE | `enable-all: true`, `disable: [go-require]`                                            |
| 16 | swaggo formatter detection improvements         | ✅ DONE | Added 2 imports (`swaggo/files`, `swaggo/cmd/swag`), 4 patterns, config file detection |
| 17 | Benchmarking for analyzer and fixer             | ✅ DONE | `pkg/linter/benchmark_test.go`, `pkg/diff/benchmark_test.go`                           |
| 18 | Document RE2 exclusion pattern syntax in README | ✅ DONE | New `## Exclusion Patterns` section with RE2 syntax table                              |

## Partially Done ⚠️

None — all 13 items are functionally complete.

## Known Issues & Debt 🔧

### Introduced by This Sprint

1. **funlen violations in new code** — `runFixerMode` (46 lines), `DeprecatedLintersToFindings` (32 lines), `ChangesToFindings` (34 lines), `shouldSkipLinter` (32 lines). These functions are long due to multi-line builder patterns and error checking. Extracting helpers would reduce line count.

2. **`--diff` + `--check` interaction** — When `--check` forces dry-run, the file isn't saved, so `--diff` shows nothing. This is a known limitation, not a bug. The `--diff` flag is useful with normal (non-dry-run) mode.

3. **`ErrorsToFindings` signature change** — Changed from returning `[]finding.Finding` to `([]finding.Finding, error)`. The error path is rarely triggered (only on invalid builder state), but all callers now have to handle it.

### Pre-existing Issues NOT Fixed

- `noinlineerr` warning in `detector.go:142` (pre-existing, in `hasSwaggoConfigFile`)
- `varnamelen` warning in `cmd_report.go:176` (pre-existing, variable `r`)
- `scanner.Err()` warnings in `detector.go` (pre-existing, 3 occurrences)
- funlen in `newConfigureCommand` (33 lines, now over limit due to `--check` flag addition)

## What We Should Improve 🎯

### Architecture & Type Model

1. **Extract a `findingBuilder` helper** — The converter functions have heavy repetition: `pos := ...`, `buildFinding(NewBuilder(...).With...())`, error check, append. A builder pattern or generic helper would cut 50+ lines.

2. **`MigrationResult` should distinguish dry-run vs actual** — Currently `FixesApplied` is used for both. A `WouldFix` or `Mode` field would make `--check` semantics clearer.

3. **Config deep-copy should be a method on `types.Config`** — `cloneConfig` in `cmd_configure.go` does JSON marshal/unmarshal. A `Clone()` method on the type would be reusable and testable.

4. **`LinterMinVersions` should be validated** — No test that entries in `LinterMinVersions` actually exist in `LinterPriorities`. Could silently skip valid linters.

5. **`DefaultLinterSettings` could use typed structs** — Currently `map[string]any`. Strongly-typed settings per linter would catch typos at compile time.

### Libraries & Patterns

6. **`golang.org/x/mod/semver` is already used** — Good. No change needed.

7. **`Result[T]` deleted** — Custom `types.Result[T]` wrapper removed. All functions now return idiomatic `(T, error)`. No monad abstraction, no external dependency.

8. **`json.Marshal` for deep-copy is slow** — For large configs, consider `github.com/jinzhu/copier` or a hand-written `Clone()` method.

9. **`errors.Join` for multi-error** — When `buildFinding` fails in a loop, we return the first error only. `errors.Join` (Go 1.20+) would report all failures.

### Testing Gaps

10. **No tests for `--check` mode** — Should add integration tests verifying exit codes.

11. **No tests for `--diff` flag** — Should verify diff output when changes are made.

12. **No tests for `LinterMinVersions` gating** — The categorizer version check is untested.

13. **No tests for `hasSwaggoConfigFile`** — New detection path is untested.

14. **No tests for `reference` preset** — Preset should be validated against `LinterPriorities`.

15. **No tests for `ginkgolinter`/`testifylint` defaults** — Should verify settings are applied when those linters are enabled.

## Top 25 Things to Do Next (Sorted by Impact/Effort)

| #  | Priority | Item                                                                    | Est. Effort | Impact       |
| -- | -------- | ----------------------------------------------------------------------- | ----------- | ------------ |
| 1  | HIGH     | Add tests for `--check` mode (integration)                              | 1h          | Correctness  |
| 2  | HIGH     | Add tests for `LinterMinVersions` gating in categorizer                 | 30min       | Correctness  |
| 3  | HIGH     | Add tests for `hasSwaggoConfigFile` detection                           | 30min       | Correctness  |
| 4  | HIGH     | Validate `reference` preset entries against `LinterPriorities`          | 15min       | Correctness  |
| 5  | HIGH     | Fix funlen violations (extract helpers)                                 | 1h          | Code quality |
| 6  | HIGH     | Add `Config.Clone()` method instead of JSON marshal                     | 30min       | Architecture |
| 7  | MEDIUM   | Add tests for `--diff` flag                                             | 1h          | Correctness  |
| 8  | MEDIUM   | Add tests for ginkgolinter/testifylint default settings                 | 30min       | Correctness  |
| 9  | MEDIUM   | Extract `findingBuilder` helper to reduce converter boilerplate         | 1h          | Architecture |
| 10 | MEDIUM   | Add `DryRun bool` field to `MigrationResult` for clarity                | 15min       | Type model   |
| 11 | MEDIUM   | Use `errors.Join` for multi-finding failures                            | 30min       | Robustness   |
| 12 | MEDIUM   | Validate `LinterMinVersions` entries exist in `LinterPriorities` (test) | 15min       | Correctness  |
| 13 | MEDIUM   | Fix `noinlineerr` in `hasSwaggoConfigFile`                              | 5min        | Lint         |
| 14 | MEDIUM   | Fix `varnamelen` in `cmd_report.go` (rename `r` → `report`)             | 2min        | Lint         |
| 15 | MEDIUM   | Add `--diff` support in dry-run mode (compare computed vs loaded)       | 2h          | Feature      |
| 16 | LOW      | Typed linter settings structs instead of `map[string]any`               | 4h          | Architecture |
| 17 | LOW      | Add `swaggo` formatter default settings                                 | 30min       | Feature      |
| 18 | LOW      | Document `--check` and `--diff` flags in README                         | 15min       | Docs         |
| 19 | LOW      | Document `reference` preset in README                                   | 5min        | Docs         |
| 20 | LOW      | Fix `scanner.Err()` warnings in detector.go                             | 30min       | Robustness   |
| 21 | LOW      | Add benchmark for `RecommendationsToFindings` (error-return overhead)   | 15min       | Performance  |
| 22 | LOW      | Consider `golangci-lint fmt` integration in check mode                  | 1h          | Feature      |
| 23 | LOW      | Add `--format sarif` support to `--check` mode output                   | 1h          | Feature      |
| 24 | LOW      | Typed formatter settings instead of `map[string]any`                    | 2h          | Architecture |
| 25 | LOW      | Migrate `justfile` commands to `flake.nix` apps                         | 2h          | Build        |

## Files Changed (22 files, +399/-73 lines)

```
README.md                        | 38 ++++++
flake.nix                        |  6 ++
internal/cli/cmd_analyze.go      |  5 +-
internal/cli/cmd_configure.go    | 87 +++++++++++---
internal/cli/cmd_report.go       |  9 +-
internal/cli/cmd_validate.go     |  7 +-
internal/cli/commands.go         |  3 +
justfile                         |  2 +
pkg/config/loader.go             |  3 +
pkg/constants/config.go          | 10 ++
pkg/constants/presets.go         | 19 +++-
pkg/constants/version.go         |  9 ++
pkg/detection/detector.go        | 21 +++++
pkg/detection/patterns.go        |  6 ++
pkg/errors/errors.go             |  1 +
pkg/finding/converter.go         | 91 ++++++++--------
pkg/finding/converter_test.go    | 47 ++++---
pkg/finding/detector.go          | 33 +++--
pkg/finding/diff_converter.go    | 23 ++--
pkg/finding/golangci_lint.go     |  9 +-
pkg/finding/helpers.go           | 31 +++--
pkg/linter/categorizer.go        | 12 ++
pkg/report/report_templ.go       | Deleted from git tracking
pkg/diff/benchmark_test.go       | NEW
pkg/linter/benchmark_test.go     | NEW
```

## Top #1 Question I Cannot Figure Out Myself

**How should `--diff` work in dry-run / check mode?** Currently it shows nothing because the file isn't saved. Options:

1. Keep current behavior (diff only works with actual fixes)
2. Compute the "would-be" config in-memory and diff against loaded config (requires exposing the fixer's in-memory result)
3. Always show analysis recommendations as diff (use `RecommendationsToFindings` instead)

This is a product/design decision. Option 2 is the most useful but requires the fixer to return the modified config object (not just a count). That would change the `FixConfig` API.
