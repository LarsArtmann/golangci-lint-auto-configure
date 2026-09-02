# Status Report: SUPERB Plan Execution — Phase 1-3 In Progress

**Date:** 2026-07-26 09:41
**Session:** Architecture & Data Model Improvement Plan execution
**Plan:** `docs/planning/2026-07-26_06-34_SUPERB-architecture-data-model-improvement-plan.md`

---

## Executive Summary

Executing the 16-task Pareto plan to make the architecture and data model "superb." Completed Phase 1 (Safe Type-Safety Wins) and Phase 2 (Cleanup & Completion) fully. Phase 3 (Architectural Decoupling) is partially done — MT9 (decouple finding) and MT10 (remove aliases) are complete, MT11 (invert linter→config) is mid-implementation with a **broken build** that needs immediate fixing. Phase 4 is not started.

**Current state: BUILD IS BROKEN.** The auto-commit daemon committed a state where `newConfigUpdater` is called with the old signature (1 arg) but the function now requires 2 args (logger + GoVersionProvider). One call site in `fixer_config.go:218` was not updated.

---

## a) FULLY DONE (Verified Green)

### Phase 1 — Safe Type-Safety Wins (1% → 51%) ✅

| Task   | Description                                                                           | Status                   |
| ------ | ------------------------------------------------------------------------------------- | ------------------------ |
| MT1    | `NewConfig()` constructor with functional options (`pkg/types/config_constructor.go`) | ✅ Done, 85 specs pass   |
| MT2    | Brand `Version` type (`type Version string`) with `Valid()`/`Compare()` using semver  | ✅ Done                  |
| MT3    | `LintersConfig.Enable`/`Disable` changed from `[]string` to `[]LinterName`            | ✅ Done, ~30 files fixed |
| MT4    | `FormattersConfig.Enable`/`Disable` changed from `[]string` to `[]FormatterName`      | ✅ Done, ~15 files fixed |
| GATE 1 | `go build ./...` + `go test -race ./pkg/... ./internal/...` — all 18 packages green   | ✅ Passed                |

### Phase 2 — Cleanup & Completion (4% → 64%) ✅

| Task   | Description                                                                                                                                                | Status    |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| MT5    | Branded types propagated: `policy.Policy.Disabled` map key → `LinterName`, `report.JSONReport` linter slices → `[]LinterName`                              | ✅ Done   |
| MT6    | `settingsToMap` panic renamed to `mustSettingsToMap` with descriptive panic messages (intentional: static structs cannot fail to marshal)                  | ✅ Done   |
| MT7    | `TriState` enum created in `pkg/migration/tristate.go`; 7 `*bool` fields in migration types replaced                                                       | ✅ Done   |
| MT8    | Dead code removed (`EnableGolinesFormatter`), stale `G104` removed from `.golangci.yml`, `golines` added to `format` preset (aligns with `CoreFormatters`) | ✅ Done   |
| GATE 2 | All 18 packages green with race detector                                                                                                                   | ✅ Passed |

---

## b) PARTIALLY DONE

### Phase 3 — Architectural Decoupling (20% → 80%) 🔧

| Task | Description                                                                                                                                                                                                         | Status                           |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- |
| MT9  | Decouple `pkg/finding` from `pkg/linter` — `ConfigAnalysisDetector` now takes `ConfigAnalyzer` interface (defined in `pkg/finding`) instead of `*linter.Analyzer`                                                   | ✅ Done                          |
| MT10 | Remove type-alias re-exports — `type Config = types.Config` block deleted from `pkg/config/loader.go`; all ~10 external callers updated to `types.Config`                                                           | ✅ Done                          |
| MT11 | Invert `pkg/linter` → `pkg/config` dependency — `GoVersionProvider` func type injected, constants moved to `pkg/constants`, `config.DefaultMaxIssuesPerLinter`/`DefaultMaxSameIssues` duplicated to `pkg/constants` | ⚠️ **IN PROGRESS — BUILD BROKEN** |
| MT12 | Dissolve `pkg/client` god-package                                                                                                                                                                                   | ❌ Not started                   |

**The break:** `pkg/linter/fixer_config.go:218` has a second call to `newConfigUpdater(f.logger)` that was not updated to `newConfigUpdater(f.logger, f.goVersionProvider)`. The auto-commit daemon committed this broken state.

**The fix (1 line):** Change `newConfigUpdater(f.logger)` to `newConfigUpdater(f.logger, f.goVersionProvider)` at line 218 of `fixer_config.go`. Then wire `SetGoVersionProvider(config.GetLocalGoVersion)` in `pkg/client/client.go` alongside the CLI wiring already done in `cmd_configure_fixer.go`.

---

## c) NOT STARTED

### Phase 4 — Deep Refactors (→ 100%)

| Task   | Description                                                                  | Risk |
| ------ | ---------------------------------------------------------------------------- | ---- |
| MT13   | Replace 12 global CLI flag vars with `CommandContext` struct                 | MED  |
| MT14   | Config unification design spike (ADR-006)                                    | LOW  |
| MT15   | Implement Config unification (`migration.Config` → shim over `types.Config`) | HIGH |
| MT16   | `SettingsMap` wrapper with typed getters                                     | MED  |
| GATE 4 | `nix flake check` (full hermetic build)                                      | —    |

---

## d) TOTALLY FUCKED UP 🚨

1. **BUILD IS CURRENTLY BROKEN.** The auto-commit daemon committed a state where `newConfigUpdater` signature changed (now requires 2 args) but one call site was not updated. `GOEXPERIMENT=jsonv2 go build ./...` fails with `not enough arguments in call to newConfigUpdater`. This needs to be fixed IMMEDIATELY before any other work.

2. **`mapKeys` function may now be dead code.** The `loadPresetConfig` and `applyPresetFormatters` functions were rewritten to use `types.NewSet`/`types.ToSortedSlice` instead of `map[string]struct{}` + `mapKeys()`. The `mapKeys` generic function in `cmd_configure_config.go` may have zero remaining callers. Should be checked and removed if dead.

3. **Duplicate constants.** `DefaultMaxIssuesPerLinter` and `DefaultMaxSameIssues` now exist in BOTH `pkg/config/loader.go` (lines 45-47) and `pkg/constants/config.go`. The config package ones should be removed and all internal callers updated to use `constants.DefaultMaxIssuesPerLinter`.

4. **GATE 3 not passed.** Phase 3 cannot be gated until MT11 is completed and MT12 is done.

---

## e) WHAT WE SHOULD IMPROVE

1. **Stop relying on sed for bulk type changes.** The sed-based approach for MT3/MT4 (LinterName/FormatterName propagation) caused cascading test breakages that took ~15 iterative fix cycles. A better approach: use `gopls rename` or LSP-based refactoring, or at minimum run `go test` after each file edit instead of batching.

2. **The auto-commit daemon is dangerous during active refactoring.** It committed a broken build. Consider disabling it during multi-file refactors or adding a pre-commit build check.

3. **Test the build after EVERY file change, not after batching.** Several times I made 5+ edits then discovered the build was broken and had to bisect.

4. **Phase 3 MT11 scope was underestimated.** The plan said "90min MED risk" but the `config` → `constants` constant migration + `GoVersionProvider` injection + `SetGoVersionProvider` wiring touches the Fixer struct, configUpdater, 3 call sites, and client.go. It should have been split into two tasks.

5. **Missing test coverage for the `Version` type.** `MT2` added `Valid()` and `Compare()` methods but no dedicated tests were written for `Compare()` (semver ordering).

6. **Missing test for `TriState` YAML unmarshaling.** `MT7` added `UnmarshalYAML` but no test verifies that `nil → Unspecified`, `true → Enabled`, `false → Disabled`.

---

## f) Next 50 Things to Get Done

### Immediate (Block everything)

1. **Fix the broken build** — update `newConfigUpdater` call at `fixer_config.go:218`
2. Wire `SetGoVersionProvider(config.GetLocalGoVersion)` in `pkg/client/client.go`
3. Remove duplicate `DefaultMaxIssuesPerLinter`/`DefaultMaxSameIssues` from `pkg/config/loader.go`
4. Check if `mapKeys` in `cmd_configure_config.go` is now dead code; remove if so
5. Run full test suite to verify GATE 2 is restored

### Phase 3 Completion

6. **MT12:** Analyze `pkg/client` public API and callers
7. **MT12:** Move client wiring to `internal/cli` or give focused facade
8. **MT12:** Update `examples/api-usage/main.go` if it imports client
9. **GATE 3:** `go build + go test -race + golangci-lint run`
10. **GATE 3:** Verify `pkg/finding` no longer imports `pkg/linter`
11. **GATE 3:** Verify `pkg/linter` no longer imports `pkg/config`
12. **GATE 3:** `grep -r "config.Config\b" internal/ pkg/` returns zero alias hits

### Phase 4 — Deep Refactors

13. **MT13:** Design `CommandContext` struct (flags + logger + analyzer refs)
14. **MT13:** Replace 12 global `var` declarations in `commands.go`
15. **MT13:** Update `newConfigureCommand` to use `CommandContext`
16. **MT13:** Update `newAnalyzeCommand` + `newValidateCommand`
17. **MT13:** Update `newReportCommand` + `newAuditCommand` + presets
18. **MT13:** Update `cmd_configure_fixer.go` + `cmd_configure_preset.go`
19. **MT13:** Run full build + test
20. **MT14:** Diff `types.Config` vs `migration.Config` — list every field difference
21. **MT14:** Design the v1→v2 normalizing shim
22. **MT14:** Write ADR `docs/adr/ADR-006-Unify-Config-Type.md`
23. **MT14:** Identify migration test coverage gaps
24. **MT15:** Create `migration.ShimConfig` type wrapping `types.Config`
25. **MT15:** Implement v1→v2 field normalization
26. **MT15:** Update `migrator.go` to use shim
27. **MT15:** Delete duplicate sub-structs from `migration/config_types.go`
28. **MT15:** Fix all consumer compiler errors
29. **MT15:** Run full migration test suite + golden files
30. **MT16:** Design `SettingsMap` type with typed `GetString`/`GetSlice`/`GetMap`
31. **MT16:** Implement `SettingsMap` with centralized type assertions
32. **MT16:** Migrate `pkg/types/clone.go` to use `SettingsMap`
33. **MT16:** Migrate `pkg/config/merger_helpers.go`
34. **MT16:** Migrate `pkg/config/settings_validator.go`
35. **MT16:** Migrate `pkg/linter/fixer_config.go` settings access
36. **GATE 4:** `nix flake check` (full hermetic build + format + test)

### Quality Follow-ups

37. Write tests for `Version.Compare()` (semver ordering edge cases)
38. Write tests for `TriState.UnmarshalYAML` (nil/true/false)
39. Write tests for `NewConfig()` options (all `With*` functions)
40. Run `golangci-lint run` on the repo's own config — verify zero issues
41. Run `UPDATE_GOLDEN=1 go test ./pkg/report/...` — review HTML golden diff
42. Update `AGENTS.md` with the new type system (LinterName/FormatterName/Version/TriState)
43. Update `docs/references/json-v2.md` with the `omitzero` vs `omitempty` status after branded types
44. Update `TODO_LIST.md` — remove completed items (format preset, dead code, G104)
45. Update `CHANGELOG.md` with all Phase 1-3 changes
46. Run `go mod tidy` (semver dependency added by Version type)
47. Update `vendorHash` in `flake.nix` after go.mod change
48. Run `nix build` to verify hermetic build
49. Cut release `v0.6.0` (additive type changes + format preset alignment = minor bump)
50. Run full `nix flake check` as final GATE 4

---

## g) Questions I Cannot Answer Myself

1. **Should `pkg/client` be dissolved entirely or kept as a focused facade?** The plan says "dissolve into `internal/cli` or give focused API" but `examples/api-usage/main.go` imports it — removing it is a breaking API change for external consumers. Do you want to keep a public `pkg/client` with a narrowed API, or dissolve it and update the example?

2. **Should `DefaultMaxIssuesPerLinter`/`DefaultMaxSameIssues` live in `pkg/constants` or `pkg/config`?** I moved them to `constants` to break the linter→config dependency, but they were originally config-package constants. Keeping both creates a split-brain. The "right" home depends on whether you consider these "config schema defaults" (config) or "tool-wide constants" (constants).

3. **Should the `format` preset now align exactly with `CoreFormatters` (4 formatters) or remain intentionally different?** I added `golines` to align them (the plan recommended this), but the old `format` preset was documented as the "minimal formatter" preset. If you want them different, I should revert and document the intentional divergence instead.
