# Status Report: 2026-04-01 04:30 — Funlen Enforcement (30 Lines Max)

## Mission

Enforce `funlen` linter with `lines: 30` and `statements: 20` across the entire codebase — no exceptions.

---

## A) FULLY DONE

| #   | Item                              | Details                                                                                                                                                         |
| --- | --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `.golangci.yml` threshold updated | `funlen.lines: 80→30`, `funlen.statements: 50→20` — **staged but NOT committed** (pre-commit hook failed due to golangci-lint cache corruption, not our change) |

---

## B) PARTIALLY DONE

| #   | Item                   | Status                                                       | Blocker                                                                                                                                                                           |
| --- | ---------------------- | ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `.golangci.yml` commit | Staged, not committed                                        | Pre-commit hook's golangci-lint step hit a Go build cache error (`no such file or directory`). The **change itself is correct**. Needs `go clean -cache` or `--no-verify` commit. |
| 2   | Violation audit        | **32 violations identified across 18 files** (see section D) | None — audit is complete, fixes not started                                                                                                                                       |

---

## C) NOT STARTED

| #   | Item                                                         | Details                                                                                                                                                                     |
| --- | ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Add `RecommendedLinterSettings` to `pkg/constants/config.go` | Single source of truth for linter settings like funlen thresholds. Currently, settings are only in YAML files — no Go code knows the recommended values.                    |
| 2   | Update 4 example config files                                | `examples/standard.golangci.yml` (80→30), `examples/web-project.golangci.yml` (80→30), `examples/library.golangci.yml` (50→30), `examples/cli-project.golangci.yml` (60→30) |
| 3   | Update funlen reason in `pkg/constants/linter_reasons.go`    | Change `"Detect long functions"` → something mentioning 30-line threshold                                                                                                   |
| 4   | Remove funlen from `.golangci.yml` exclusion rules           | Lines 261, 264, 267, 270, 274, 331 exclude funlen for specific files — these should be removed so funlen is enforced everywhere                                             |
| 5   | Refactor 32 violating functions across 18 files              | The bulk of the work — see violation list below                                                                                                                             |
| 6   | Update `AGENTS.md` documentation                             | Lines referencing `funlen: lines: 80, statements: 50`                                                                                                                       |
| 7   | Full test + lint verification                                | Run `just test` and `just lint` to confirm everything passes                                                                                                                |

---

## D) TOTALLY FUCKED UP / BLOCKERS

### Pre-commit hook golangci-lint cache corruption

The commit failed because the pre-commit hook runs `golangci-lint run --fix` which hit:

```
could not load export data: open /Users/larsartmann/Library/Caches/go-build/.../0995d4002b0d12a49bfc552d60cebfd45687dcb9320320501f66ddffc9263630-d: no such file or directory
```

**Root cause:** Go build cache corruption (unrelated to our change). **Fix:** `go clean -cache` before committing.

### The 32 funlen violations (full list)

| #   | File                                             | Function                              | Lines  | Stmt   | Severity  |
| --- | ------------------------------------------------ | ------------------------------------- | ------ | ------ | --------- |
| 1   | `pkg/migration/rules.go:21`                      | `DefaultRules`                        | **79** | —      | 🔴 SEVERE |
| 2   | `internal/cli/cmd_analyze.go:39`                 | `newAnalyzeCommand`                   | **65** | —      | 🔴 SEVERE |
| 3   | `internal/cli/cmd_report.go:14`                  | `newReportCommand`                    | **62** | —      | 🔴 SEVERE |
| 4   | `internal/cli/cmd_configure.go:60`               | `newConfigureCommand`                 | **61** | —      | 🔴 SEVERE |
| 5   | `pkg/ui/formatter.go:21`                         | `FormatRecommendations`               | **39** | —      | 🟡        |
| 6   | `pkg/linter/categorizer.go:10`                   | `CategorizeLinters`                   | **46** | —      | 🔴        |
| 7   | `internal/cli/cmd/completion.go:8`               | `NewCompletionCommand`                | **47** | —      | 🔴        |
| 8   | `internal/cli/commands.go:48`                    | `NewRootCommand`                      | **50** | —      | 🔴        |
| 9   | `pkg/utils/retry.go:34`                          | `WithRetry`                           | **42** | —      | 🟡        |
| 10  | `internal/cli/cmd_configure_internal_test.go:52` | `TestParsePriorityParam`              | **53** | —      | 🔴        |
| 11  | `pkg/linter/command_runner.go:19`                | `runCommandWithRetry`                 | **35** | —      | 🟡        |
| 12  | `pkg/report/json_report_generator.go:42`         | `GenerateJSONReport`                  | **37** | —      | 🟡        |
| 13  | `pkg/diff/differ.go:114`                         | `compareEnabled`                      | **37** | —      | 🟡        |
| 14  | `pkg/config/loader.go:282`                       | `CreateDefaultConfig`                 | **34** | —      | 🟡        |
| 15  | `pkg/diff/differ.go:78`                          | `compareRunSettings`                  | **33** | —      | 🟡        |
| 16  | `internal/cli/cmd_configure.go:209`              | `ensureConfigFile`                    | **32** | —      | 🟡        |
| 17  | `pkg/detection/detector_bench_test.go:11`        | `setupBenchmarkProject`               | **35** | —      | 🟡        |
| 18  | `pkg/linter/fixer_preflight.go:253`              | `calculateDryRunResultWithDeprecated` | **40** | —      | 🟡        |
| 19  | `pkg/migration/migrations.go:130`                | `migrateIssuesFlags`                  | **35** | —      | 🟡        |
| 20  | `pkg/migration/migrations.go:169`                | `migrateFormatters`                   | **31** | —      | 🟡        |
| —   | _Statements-only violations:_                    |                                       |        |        |           |
| 21  | `pkg/linter/fixer_preflight.go:122`              | `preFixDeprecatedLinters`             | —      | **33** | 🟡        |
| 22  | `pkg/linter/fixer_preflight.go:192`              | `preFixTypecheck`                     | —      | **26** | 🟡        |
| 23  | `pkg/linter/analyzer.go:145`                     | `FormatRecommendations`               | —      | **31** | 🟡        |
| 24  | `pkg/migration/migrations.go:225`                | `migrateOutputProperties`             | —      | **25** | 🟡        |
| 25  | `pkg/migration/config_types.go:29`               | `UnmarshalYAML`                       | —      | **32** | 🟡        |
| 26  | `internal/cli/cmd_configure.go:127`              | `runConfigure`                        | —      | **34** | 🟡        |
| 27  | `internal/cli/cmd_configure.go:261`              | `applyPreset`                         | —      | **21** | 🟢        |
| 28  | `pkg/detection/detector.go:91`                   | `detect`                              | —      | **21** | 🟢        |
| 29  | `pkg/detection/detector.go:148`                  | `analyzeGoMod`                        | —      | **29** | 🟡        |
| 30  | `pkg/diff/differ.go:165`                         | `FormatChanges`                       | —      | **25** | 🟡        |
| 31  | `pkg/diff/differ.go:210`                         | `GetSummary`                          | —      | **21** | 🟢        |
| 32  | `pkg/linter/analyzer.go:69`                      | `AnalyzeConfigResult`                 | —      | **22** | 🟡        |

**Breakdown by area:**

| Area             | Count | Worst offender                     |
| ---------------- | ----- | ---------------------------------- |
| `internal/cli/`  | 9     | `newAnalyzeCommand` (65 lines)     |
| `pkg/migration/` | 6     | `DefaultRules` (79 lines)          |
| `pkg/linter/`    | 6     | `CategorizeLinters` (46 lines)     |
| `pkg/diff/`      | 4     | `compareEnabled` (37 lines)        |
| `pkg/config/`    | 1     | `CreateDefaultConfig` (34 lines)   |
| `pkg/detection/` | 3     | `setupBenchmarkProject` (35 lines) |
| `pkg/report/`    | 1     | `GenerateJSONReport` (37 lines)    |
| `pkg/ui/`        | 1     | `FormatRecommendations` (39 lines) |
| `pkg/utils/`     | 1     | `WithRetry` (42 lines)             |

---

## E) WHAT WE SHOULD IMPROVE

### Architecture Improvements

1. **`RecommendedLinterSettings` constant** — Add a strongly-typed map in `pkg/constants/config.go` that defines recommended linter settings (funlen lines/statements, gocyclo complexity, etc.). This becomes the single source of truth used by:
   - The fixer when writing new configs
   - `CreateDefaultConfig()` in the loader
   - Example config generation
   - Validation against outdated settings

2. **Remove `map[string]any` for linter settings** — `LintersConfig.Settings` is `map[string]any`. This means no type safety, no IDE completion, easy typos. Consider a `LinterSettings` struct or at least typed accessors. (Big refactor, lower priority.)

3. **Consistent exclusion management** — The `.golangci.yml` has 6 funlen exclusions for specific files. Rather than maintaining exclusion lists, refactor the code to comply. Exclusions should be rare exceptions, not the norm.

### Process Improvements

4. **Pre-commit hook reliability** — The golangci-lint cache corruption shouldn't block commits. Consider adding `go clean -cache` as a pre-step, or using `--no-verify` with a follow-up lint check.

5. **Example configs should auto-sync** — The 4 example configs have different funlen thresholds (50, 60, 80, 80). They should derive from the same source of truth.

---

## F) TOP 25 THINGS TO DO NEXT (Priority Order)

### Immediate (must do to complete the funlen task)

1. **Fix Go build cache** — `go clean -cache` to unblock commits
2. **Commit the `.golangci.yml` change** — The threshold update that's already staged
3. **Remove funlen from ALL exclusion rules in `.golangci.yml`** — 6 exclusion entries (lines 261, 264, 267, 270, 274, 331)
4. **Refactor `DefaultRules()` (79 lines!)** — `pkg/migration/rules.go` — split into sub-rule functions
5. **Refactor `newAnalyzeCommand` (65 lines)** — `internal/cli/cmd_analyze.go` — extract flag setup
6. **Refactor `newReportCommand` (62 lines)** — `internal/cli/cmd_report.go` — extract flag setup
7. **Refactor `newConfigureCommand` (61 lines)** — `internal/cli/cmd_configure.go` — extract flag setup
8. **Refactor `NewCompletionCommand` (47 lines)** — `internal/cli/cmd/completion.go` — extract completion logic
9. **Refactor `CategorizeLinters` (46 lines)** — `pkg/linter/categorizer.go` — extract category logic
10. **Refactor `NewRootCommand` (50 lines)** — `internal/cli/commands.go` — extract subcommand registration
11. **Refactor `TestParsePriorityParam` (53 lines)** — `internal/cli/cmd_configure_internal_test.go` — use table-driven tests
12. **Refactor `WithRetry` (42 lines)** — `pkg/utils/retry.go` — simplify control flow
13. **Refactor `calculateDryRunResultWithDeprecated` (40 lines)** — `pkg/linter/fixer_preflight.go`
14. **Refactor `FormatRecommendations` in `pkg/ui/` (39 lines)** — `pkg/ui/formatter.go`
15. **Refactor `GenerateJSONReport` (37 lines)** — `pkg/report/json_report_generator.go`
16. **Refactor remaining 16 violations** — All the 31-37 line functions
17. **Add `RecommendedLinterSettings` constant** — `pkg/constants/config.go` for single source of truth
18. **Update 4 example configs** — `examples/*.golangci.yml` funlen thresholds
19. **Update funlen reason** — `pkg/constants/linter_reasons.go`
20. **Update AGENTS.md** — Funlen documentation reference

### Post-completion verification

21. **Run `just test`** — Verify all tests pass after refactoring
22. **Run `just lint`** — Verify zero funlen violations
23. **Run `golangci-lint run --enable-only funlen`** — Final targeted check
24. **Git commit with detailed message** — One commit per logical change
25. **Git push** — Push all changes to remote

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should the `funlen` exclusions be removed entirely, or should some files keep permanent exclusions?**

Currently excluded from funlen:

- `internal/cli/cmd/installhook.go` — likely long because of Cobra command wiring
- `internal/cli/cmd/migrate.go` — migration CLI command
- `internal/cli/cmd_validate.go` — validation CLI command
- `pkg/detection/detector_test.go` — test file
- `pkg/diff/differ_test.go` — test file
- `pkg/linter/fixer.go` — the main fixer logic

**My recommendation:** Remove ALL exclusions and refactor the code to comply. The 30-line limit forces good decomposition. But if you want to keep test file exclusions (tests can be naturally longer with table-driven setups), that's a valid trade-off.

**Decision needed:** Enforce funlen on test files too, or exclude `*_test.go` from funlen?

---

## Summary

| Metric                       | Value                                                         |
| ---------------------------- | ------------------------------------------------------------- |
| Task completion              | **~5%** (threshold updated, nothing refactored yet)           |
| Violations found             | **32** across 18 files                                        |
| Worst offender               | `DefaultRules()` at 79 lines (2.6x over limit)                |
| Files needing refactor       | 18 source files                                               |
| Estimated refactoring effort | **Medium-Large** (most are straightforward extractions)       |
| Architecture debt            | `map[string]any` for linter settings, no centralized defaults |
| Blocked?                     | No — just need `go clean -cache` and then proceed             |
