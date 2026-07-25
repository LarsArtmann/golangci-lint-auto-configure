# Status Report: Deep Audit + Targeted Fix Sprint

**Date**: 2026-06-05 07:14
**Branch**: master
**Commits**: 7 new (9015511..aac15fa)

---

## Summary

Deep codebase audit followed by targeted fixes for bugs, type safety improvements, and code quality. All changes verified with full test suite (15 suites pass).

---

## A) FULLY DONE

### Bug Fixes (Critical)

- **Config.Clone() deep clone** (`pkg/types/clone.go`): Replaced shallow `maps.Copy` with recursive `cloneAnyMap` that handles nested `map[string]any` and `[]any`. Added 2 proof tests that mutate nested values and verify original isolation.
- **ExclusionRuleConfig shallow clone** (`pkg/types/clone.go`): Added `cloneExclusionRules` that deep-clones each rule's `Linters []string` slice.
- **Differ Disable list comparison** (`pkg/diff/differ.go`): Added `compareDisabled` method — the differ was silently ignoring all changes to Disable lists.
- **Version double-v-prefix** (`pkg/linter/version_checker.go`): Used `TrimPrefix`+add instead of conditional `HasPrefix` to prevent `"vv2.10.1"` from JSON output that already includes `"v"`.

### Type Safety

- **ParseLinterPriority** (`pkg/types/types.go`): Added `ParseLinterPriority(s string) (LinterPriority, error)` that returns an error for invalid input instead of silently defaulting. CLI's `ParsePriorityParam` delegates to it.
- **Removed unused CLI constants**: `priorityCritical/High/Medium/Optional` in `cmd_configure.go` — now unused after delegating to `ParseLinterPriority`.

### Code Quality

- **RuleID constants** (`pkg/finding/converter.go`): Extracted 6 finding RuleID strings to named constants (`RuleIDMissingLinter`, `RuleIDMissingFormatter`, `RuleIDDeprecatedLinter`, `RuleIDValidationError`, `RuleIDGenericError`, `RuleIDConfigFix`). Fixed `ErrorsToFindings` using `"validation-error"` for generic errors → now uses `RuleIDGenericError`.
- **Health rule constants** (`pkg/types/validation.go`): Extracted 4 health rule names to constants (`RuleDuplicateLinter`, `RuleEnableDisableOverlap`, `RuleMissingCriticalLinter`, `RuleV1SyntaxInV2`).
- **Cached strings.NewReplacer** (`pkg/finding/converter.go`): Moved from per-call allocation to package-level `var linterTagReplacer`.
- **Removed always-nil error returns** (`pkg/linter/fixer_preflight.go`): `calculateDryRunResultWithInvalidDurations` and `calculateDryRunResultWithDeprecated` now return only `*types.MigrationResult`, fixing unparam lint warnings.
- **detectFormat empty extension** (`pkg/config/loader.go`): Handles `""` extension (files with no extension) correctly as YAML.

---

## B) PARTIALLY DONE

### Audit Complete, Implementation Deferred

The full audit identified 30+ issues across the codebase. The following were analyzed but not implemented due to blast radius:

- `[]string` → `[]LinterName` for `LintersConfig.Enable/Disable` (13 files, high risk)
- `ConfigVersion` type for `Config.Version` (20+ references)
- `GeneratedMode` enum for `LintersExclusionsConfig.Generated`

---

## C) NOT STARTED

1. **Type Enable/Disable as []LinterName** — Analyzed, deferred (13 files touched)
2. **ConfigVersion type** — Analyzed, deferred (20+ references)
3. **GeneratedMode enum** — Analyzed, deferred
4. **ConfigLoader god interface split** — Not started
5. **CLI package-level globals removal** — Not started (9 mutable globals in commands.go)
6. **pkg/client/client.go tests** — Not started (210 lines, 0% coverage)
7. **Spinner goroutine leak** — Not started
8. **KnownFields(true) in YAML decoder** — Not started
9. **go-error-family adoption** — Not started
10. **map[string]any typed accessors** — Not started (design decision needed)

---

## D) TOTALLY FUCKED UP

Nothing new. All prior critical issues either fixed or documented.

---

## E) WHAT WE SHOULD IMPROVE

### Critical (Still Open)

1. **pkg/client/client.go has zero tests** — 210 lines, 7 exported methods, public API surface, 0% coverage. This is the package external consumers use.

### High Priority

2. **Type Enable/Disable as []LinterName** — Safe to do now that `ParseLinterPriority` pattern is established. Use `type LinterName string` transparently with YAML.
3. **KnownFields(true)** — Enable YAML typo detection. Will catch misspelled config keys.
4. **CLI globals removal** — 9 mutable globals make parallel testing impossible.

### Medium Priority

5. **map[string]any typed accessors** — `GetStringSetting(linter, key)`, `GetIntSetting(linter, key)` for safe linter settings access.
6. **go-error-family** — Structured error classification for better error handling.
7. **ConfigVersion type** — Prevent `"3"` from silently passing validation.

---

## F) TOP 25 THINGS TO DO NEXT

| #   | Task                                                           | Impact   | Effort   |
| --- | -------------------------------------------------------------- | -------- | -------- |
| 1   | Add tests for pkg/client/client.go                             | Critical | 45min    |
| 2   | Type Enable/Disable as []LinterName                            | High     | 45min    |
| 3   | Enable KnownFields(true) in YAML decoder                       | High     | 30min    |
| 4   | Add ConfigVersion type for Config.Version                      | Medium   | 30min    |
| 5   | Add GeneratedMode enum                                         | Medium   | 20min    |
| 6   | Add map[string]any typed accessors                             | Medium   | 60min    |
| 7   | Remove CLI package-level globals                               | Medium   | 60min    |
| 8   | Split ConfigLoader god interface                               | Medium   | 60min    |
| 9   | Fix spinner goroutine leak                                     | Medium   | 30min    |
| 10  | Adopt go-error-family                                          | Medium   | 90min    |
| 11  | Detector file handle leak (filepath.Walk defer)                | Medium   | 30min    |
| 12  | gogenfilter scanner: log filter errors                         | Low      | 10min    |
| 13  | gogenfilter scanner: Windows path fix                          | Low      | 15min    |
| 14  | detector.go: scanner.Err() silently discarded                  | Low      | 10min    |
| 15  | analyzer.go: FindBinary ignores context                        | Low      | 15min    |
| 16  | analyzer.go: duplicate FindBinary+CheckVersion calls           | Low      | 20min    |
| 17  | categorizer.go: empty version allows min-version-gated linters | Medium   | 20min    |
| 18  | Add ToSortedSlice consistency (nil vs empty)                   | Low      | 10min    |
| 19  | Errors package: consolidate repetitive boilerplate             | Low      | 30min    |
| 20  | Error field names inconsistent (Path/File/Config)              | Low      | 15min    |
| 21  | Loader.getAllLinterNames only returns enabled linters          | Medium   | 20min    |
| 22  | Fix differ test coverage for new compareDisabled               | High     | 15min    |
| 23  | Document map[string]any design decision in ADR                 | Medium   | 30min    |
| 24  | Explore code-gen from golangci-lint schema                     | Low      | Research |
| 25  | Add CI pipeline test for all presets end-to-end                | High     | 2h       |

---

## G) TOP QUESTION

**The `map[string]any` settings access pattern**: Currently `LintersConfig.Settings` is `map[string]any` and consumers do blind type assertions like `settings["gocritic"].(map[string]any)`. Options:

- **(A)** Typed accessor helpers (`GetLinterSettings(linter) map[string]any`, `GetInt(linter, key) int`) — safe, minimal change
- **(B)** Typed structs for top ~15 linters — most safe, but YAML round-trip with mixed typed/untyped needs validation
- **(C)** Code generation from golangci-lint schema — most correct, creates tight coupling

My recommendation: Start with (A) — add typed accessors. Evaluate (B) once we understand actual usage patterns. (C) is a longer-term research item.

---

## Commits (This Session)

1. `9015511` — fix: deep clone nested map[string]any in Config.Clone()
2. `27831ad` — refactor: extract RuleID constants and cache strings.NewReplacer
3. `12deb6f` — feat: add ParseLinterPriority with validation
4. `075b1bd` — fix: add Disable list comparison to config differ
5. `4d01e8f` — fix: version double-v-prefix, unparam warnings, health rule constants, detectFormat
6. `aac15fa` — fix: deep clone ExclusionRuleConfig.Linters slice in Clone()
