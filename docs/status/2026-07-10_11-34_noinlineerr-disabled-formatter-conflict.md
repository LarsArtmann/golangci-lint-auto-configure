# Status: noinlineerr Linter Disabled — Formatter Conflict Resolution

**Date:** 2026-07-10 11:34
**Session scope:** Disable `noinlineerr` linter everywhere, add tests, verify
**Commit:** 0c6c4c7 (on master, pushed)
**Working tree:** clean
**Tests:** 16/16 packages ✅ | **Lint:** 2 pre-existing gosec issues (not from this session) | **Ginkgo focused:** 4/4 new tests ✅

---

## Executive Summary

User reported that `noinlineerr` conflicts with code formatters and should always be disabled. The linter disallows inline error handling patterns (`if err := foo(); err != nil { ... }`), but formatters like `gofumpt` and `goimports` routinely reformat these expressions, producing contradictory findings and noisy churn. The fix leverages the existing `DisabledLinters` mechanism (same pattern as `funcorder`): add to the set, remove dead priority/reason entries, add tests at every layer.

---

## a) FULLY DONE ✅

| #   | Item                                     | File(s)                                      | Details                                                                                                                                                                                        |
| --- | ---------------------------------------- | -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **noinlineerr added to DisabledLinters** | `pkg/constants/rules.go:52`                  | `types.NewSet[LinterName]("funcorder", "noinlineerr")` with rationale comment                                                                                                                  |
| 2   | **Dead priority entry removed**          | `pkg/constants/linter_priorities.go:118`     | Removed `"noinlineerr": types.LinterPriorityMedium` — disabled linters are never recommended, so priority is dead code (same as `funcorder` precedent)                                         |
| 3   | **Dead reason entry removed**            | `pkg/constants/linter_reasons.go:119`        | Removed `"noinlineerr": "Disallows inline error handling"` — same rationale                                                                                                                    |
| 4   | **Categorizer test added**               | `pkg/linter/categorizer_test.go:125-135`     | Verifies `noinlineerr` is never recommended by `CategorizeLinters`, following the exact `funcorder` test pattern                                                                               |
| 5   | **Fixer test added**                     | `pkg/linter/fixer_test.go:466-478`           | Verifies fixer moves `noinlineerr` from `enable` list to `disable` list; uses `ConfigLoader.LoadConfig` to parse and assert on structured lists (not fragile substring matching)               |
| 6   | **Data integrity tests added**           | `pkg/constants/data_integrity_test.go:58-71` | New `DisabledLinters` Describe block with 2 tests: (1) no disabled linter in `LinterPriorities`, (2) no disabled linter in `LinterReasons` — guards the invariant for all future additions     |
| 7   | **All tests pass**                       | verified                                     | `GOEXPERIMENT=jsonv2 CGO_ENABLED=1 go test -race ./pkg/... ./internal/...` — 16 packages, 0 failures                                                                                           |
| 8   | **Lint clean on changed files**          | verified                                     | `golangci-lint run` on `./pkg/constants/... ./pkg/linter/...` — 0 issues (2 pre-existing gosec G204 issues in `internal/cli/cmd_validate.go:260` and `pkg/config/loader.go:247` are unrelated) |
| 9   | **Committed**                            | 0c6c4c7                                      | 6 files changed, 53 insertions, 3 deletions                                                                                                                                                    |

---

## b) PARTIALLY DONE 🟡

| #   | Item                       | What's done | What's missing                                                                                                                                                                                                                                                                          |
| --- | -------------------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **AGENTS.md update**       | Not done    | The `DisabledLinters` set now has 2 entries. AGENTS.md item #10 says "Priorities live in `linter_priorities.go` + reasons in `linter_reasons.go`" but doesn't mention `DisabledLinters` or the invariant that disabled linters must not have priority/reason entries. Could add a note. |
| 2   | **Validation script sync** | Not checked | `scripts/validate_linter_data.go` checks priorities↔reasons consistency but does NOT check that `DisabledLinters` entries are absent from priorities/reasons. The new data integrity test covers this at test time, but the manual validation script is now redundant/incomplete.       |

---

## c) NOT STARTED ⬜

| #   | Item                                    | Why it matters                                                                                                                                                                                                                                                                                                                |
| --- | --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Stale doc references**                | 30 files in `docs/status/` + `docs/archive/` still mention `noinlineerr` as a recommended/medium-priority linter. These are historical status reports (snapshots in time) and arguably should NOT be updated — they're point-in-time records. But if there's a curated doc that lists recommended linters, it needs updating. |
| 2   | **FEATURES.md audit**                   | No check whether FEATURES.md mentions `noinlineerr` as a supported/recommended linter. (Quick grep showed no match, but worth verifying the disabled-linter feature is documented.)                                                                                                                                           |
| 3   | **README.md user-facing docs**          | No check whether README mentions `noinlineerr` or explains which linters are disabled and why. Users should know the tool actively disables certain linters.                                                                                                                                                                  |
| 4   | **`noinlineerr` reason documentation**  | The comment in `rules.go` explains the conflict, but there's no user-facing message. When the fixer disables `noinlineerr`, the user gets no explanation in the fix summary output. The fixer counts it as a normalization fix but doesn't surface "disabled noinlineerr because it conflicts with formatters."               |
| 5   | **Finding/report for disabled linters** | Deprecated linters get SARIF/JSON findings (`DeprecatedLintersToFindings` in `pkg/finding/converter.go`). Disabled linters get no such finding — the user only sees the linter moved to `disable:` in the config. Could add `DisabledLintersToFindings` for parity.                                                           |

---

## d) TOTALLY FUCKED UP ❌

Nothing. No regressions, no broken tests, no data loss. The one test failure during development (wrong assertion using `NotTo(ContainSubstring("- noinlineerr\n"))` when the linter correctly appears in the `disable:` list) was caught and fixed immediately before commit.

---

## e) WHAT WE SHOULD IMPROVE 🔧

| #   | Issue                                             | Impact                                                                                                                                                                                                                                                                                                                                                                         | Effort                                                                         |
| --- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ |
| 1   | **No user-facing message when disabling linters** | Users see `noinlineerr` silently move to `disable:` with no explanation. The fixer logs at Debug level but the fix summary doesn't call it out.                                                                                                                                                                                                                                | Medium — add a `DisabledLintersToFindings` converter or surface in fix summary |
| 2   | **`DisabledLinters` has no reason map**           | `DeprecatedLinters` has reasons, `RedundantLinters` has reasons, but `DisabledLinters` is just a `Set` with no machine-readable reason. The rationale is only in a code comment.                                                                                                                                                                                               | Low — convert to `map[LinterName]string` or add a parallel reasons map         |
| 3   | **Manual validation script diverging from tests** | `scripts/validate_linter_data.go` checks priorities↔reasons consistency manually. The new data integrity test now covers the `DisabledLinters` invariant at test time, but the script doesn't. Either update the script or consider replacing it with test-suite coverage entirely.                                                                                            | Low                                                                            |
| 4   | **Fixer test assertion approach**                 | First attempt used fragile substring matching (`NotTo(ContainSubstring("- noinlineerr\n"))`) which failed because the linter correctly appears in `disable:`. The fix uses `ConfigLoader.LoadConfig` to parse and assert on structured data — much better. Other fixer tests still use substring matching extensively. Consider migrating more tests to structured assertions. | Medium — systematic refactor                                                   |
| 5   | **Pre-existing gosec G204 issues**                | `internal/cli/cmd_validate.go:260` and `pkg/config/loader.go:247` have gosec G204 (subprocess with tainted input). These are pre-existing, not from this session, but they cause `golangci-lint run` to exit non-zero.                                                                                                                                                         | Low — add `//nolint:gosec` with justification or refactor                      |
| 6   | **No `DisabledLinters` mention in AGENTS.md**     | AGENTS.md item #10 discusses linter priority data but doesn't mention `DisabledLinters` or the invariant. Future contributors could re-add a disabled linter to priorities.                                                                                                                                                                                                    | Low — add a sentence to item #10                                               |

---

## f) Up to 50 Things to Get Done Next

### High Priority (directly related to this session's work)

1. Add `DisabledLinters` mention to AGENTS.md item #10 (linter priority data section)
2. Convert `DisabledLinters` from `Set[LinterName]` to `map[LinterName]string` with reason strings (parity with `DeprecatedLinters` and `RedundantLinters`)
3. Add `DisabledLintersToFindings` converter in `pkg/finding/converter.go` (parity with `DeprecatedLintersToFindings`)
4. Surface disabled-linter actions in fix summary output (so users see "disabled noinlineerr: conflicts with formatters")
5. Update `scripts/validate_linter_data.go` to check `DisabledLinters` invariant (or delete it if tests fully cover it)
6. Add `--explain-disabled` CLI flag that prints which linters are disabled and why
7. Verify FEATURES.md documents the disabled-linter behavior as a feature

### Medium Priority (code quality improvements noticed during this session)

8. Migrate more fixer tests from substring matching to structured `ConfigLoader.LoadConfig` assertions (especially the Typecheck Linter tests that use `NotTo(ContainSubstring("typecheck"))`)
9. Fix pre-existing gosec G204 issues in `cmd_validate.go:260` and `loader.go:247` (add `//nolint:gosec` with justification or refactor to use allowlisted command construction)
10. Add data integrity test for `RedundantLinters` — verify entries don't appear in `DisabledLinters` (no overlap between the two mechanisms)
11. Add data integrity test for `DeprecatedLinters` — verify entries don't appear in `DisabledLinters` (a linter shouldn't be both deprecated and disabled)
12. Add data integrity test verifying all `LinterPriorities` entries have matching `LinterReasons` entries (and vice versa) — currently only in the manual script
13. Consider adding `LinterToFormatter` reason to the fixer output when `RemoveRedundantLinters` removes `lll` (same "no user-facing message" problem as disabled linters)

### Low Priority (broader project health)

14. Audit all 30 stale doc references to `noinlineerr` in `docs/status/` and `docs/archive/` — decide policy: update or leave as historical snapshots
15. Add a "Disabled Linters" section to README.md explaining which linters the tool disables and why
16. Consider a `--list-disabled` CLI subcommand for discoverability
17. Review whether `funcorder` should also have a reason string (same `Set` limitation)
18. Check if any other linters conflict with formatters and should be added to `DisabledLinters` (e.g. `gofmt` as a linter vs `gofumpt` as a formatter — though `RemoveRedundantGofmt` already handles this)
19. Add integration test that runs the full CLI with `--json` output and verifies `noinlineerr` appears in disabled list
20. Consider whether the `DisabledLinters` comment in `rules.go` should reference specific formatter names or stay generic
21. Run `nix flake check` to verify Nix build still passes after the changes
22. Run `buildflow` to verify all 44 steps still pass
23. Consider adding a test that verifies `DisabledLinters` is non-empty (guard against accidental empty set)
24. Consider adding a test that verifies `DisabledLinters` entries are valid linter names (exist in golangci-lint's linter list)
25. Document the disabled-linter mechanism in `docs/references/working-with-codebase.md`
26. Review `pkg/linter/fixer_preflight.go` — should disabled linters also be removed pre-flight like deprecated linters? (Currently they're only handled at write-back in `fixer_config.go`)
27. Add test for `funcorder` in fixer (currently only tested in categorizer, not in fixer) — same gap that existed for `noinlineerr` before this session
28. Consider unifying `funcorder` and `noinlineerr` disable reasons into a structured map
29. Check if the HTML report generator includes disabled linters in its output
30. Verify `pkg/ui/formatter.go` correctly displays disabled linters in the CLI summary
31. Consider adding `noinlineerr` to the `noinlineerr` → disabled migration in `pkg/migration/` if there's a v1→v2 migration path
32. Review whether the `reference` preset in `presets.go` should explicitly exclude `DisabledLinters` (currently works by convention — disabled linters have no priority, so they're never critical/high)
33. Add a comment to `presets.go` noting that `DisabledLinters` are excluded by convention
34. Consider adding `DisabledLinters` to the `data_integrity_test.go` "Reference preset" Describe block (verify disabled linters are not in the reference preset)
35. Review the `fixer_results.go` output — does `dryRunResult` or `successResult` mention disabled linters?
36. Check if the `--dry-run` output shows what would be disabled (currently only shows fix count)
37. Consider adding a `DisabledLinters` field to `ConfigAnalysis` (currently only `DeprecatedLinters` is surfaced in analysis)
38. Add test verifying `noinlineerr` is not in the `reference` preset
39. Consider whether `noinlineerr` should be mentioned in `docs/DOMAIN_LANGUAGE.md` under linter categories
40. Run `templ generate` to verify HTML report doesn't need updating for disabled linters
41. Check if the JSON report includes disabled linters in its output
42. Consider adding a `--disable-reason` field to the JSON report for each disabled linter
43. Review whether SARIF output should include disabled-linter findings (currently only deprecated-linter findings)
44. Add a test that verifies the fixer is idempotent when `noinlineerr` is already in the disable list
45. Consider whether `enable-all` + `disable` pattern should also move `noinlineerr` to disable (currently `EnableAll` is not handled by the fixer)
46. Review `pkg/linter/fixer.go:enableRecommendedLinters` — ensure disabled linters can never be re-enabled by the recommendation engine
47. Add a guard in `enableRecommendedLinters` that explicitly checks `DisabledLinters` (defense in depth, even though categorizer already filters)
48. Consider logging at Info level (not Debug) when a linter is disabled due to being in `DisabledLinters`
49. Review whether the `configure` command should warn if the user's existing config has `noinlineerr` enabled before the fixer runs
50. Celebrate — the task is done and all tests pass

---

## g) Top 2 Questions I Cannot Answer Myself

### 1. Should historical status reports (docs/status/, docs/archive/) be updated when a linter is disabled?

**Context:** 30 files in `docs/status/` and `docs/archive/` mention `noinlineerr` as a medium-priority recommended linter. These are timestamped snapshots. Updating them would rewrite history; leaving them creates a potential disconnect for someone reading old reports.

**What I tried:** Checked whether `funcorder` (the other disabled linter) also has stale references — it does (19 files). This suggests the established precedent is to leave historical reports alone.

**Question:** Should I leave these as historical snapshots (consistent with `funcorder` precedent), or should there be a "stale reference policy" that updates or annotates them?

### 2. Should `DisabledLinters` be upgraded to carry machine-readable reason strings?

**Context:** `DeprecatedLinters` is a `map[LinterName]LinterReplacement` (carries reason + replacement). `RedundantLinters` is a `map[LinterName]LinterToFormatter` (carries reason + formatter). `DisabledLinters` is just a `Set[LinterName]` — no reason data. The rationale for disabling `noinlineerr` lives only in a code comment and is never surfaced to the user.

**What I tried:** Followed the existing `funcorder` precedent (which is also reasonless). But `funcorder` was disabled for different reasons (produced too many false positives) and the rationale may have been communicated elsewhere.

**Question:** Should I convert `DisabledLinters` to `map[LinterName]string` (or add a parallel `DisabledLinterReasons` map) so the fixer can surface "disabled noinlineerr: conflicts with formatters" to the user? Or is the current silent-move-to-disable behavior sufficient?

---

## Test Verification Details

### New Tests (4 total, all passing)

| Test                                                                                                | Suite     | File:Line                   | Focus Filter                | Result |
| --------------------------------------------------------------------------------------------------- | --------- | --------------------------- | --------------------------- | ------ |
| `DisabledLinters should not have entries in LinterPriorities`                                       | Constants | `data_integrity_test.go:59` | `--focus="DisabledLinters"` | ✅     |
| `DisabledLinters should not have entries in LinterReasons`                                          | Constants | `data_integrity_test.go:67` | `--focus="DisabledLinters"` | ✅     |
| `CategorizeLinters Disabled Linter Handling should skip noinlineerr as conflicting with formatters` | Analyzer  | `categorizer_test.go:126`   | `--focus="noinlineerr"`     | ✅     |
| `Fixer Disabled Linters should move noinlineerr from enable to disable`                             | Analyzer  | `fixer_test.go:467`         | `--focus="noinlineerr"`     | ✅     |

### Full Suite

| Package            | Status        |
| ------------------ | ------------- |
| `pkg/client`       | ✅            |
| `pkg/config`       | ✅            |
| `pkg/constants`    | ✅            |
| `pkg/detection`    | ✅            |
| `pkg/diff`         | ✅            |
| `pkg/errors`       | ✅            |
| `pkg/finding`      | ✅            |
| `pkg/gogenfilter`  | ✅            |
| `pkg/linter`       | ✅            |
| `pkg/migration`    | ✅            |
| `pkg/report`       | ✅            |
| `pkg/types`        | ✅            |
| `pkg/ui`           | ✅            |
| `pkg/utils`        | ✅            |
| `pkg/version`      | ✅            |
| `internal/cli`     | ✅            |
| `internal/cli/cmd` | no test files |

### Lint

| Scope                                                    | Issues       | Note                                                                              |
| -------------------------------------------------------- | ------------ | --------------------------------------------------------------------------------- |
| Changed packages (`pkg/constants/...`, `pkg/linter/...`) | 0            | Clean                                                                             |
| Full project                                             | 2 gosec G204 | Pre-existing in `cmd_validate.go:260` and `loader.go:247` — not from this session |

---

## Files Changed This Session

| File                                   | Change                                                       | Lines      |
| -------------------------------------- | ------------------------------------------------------------ | ---------- |
| `pkg/constants/rules.go`               | Added `noinlineerr` to `DisabledLinters` + rationale comment | +4/-1      |
| `pkg/constants/linter_priorities.go`   | Removed dead `noinlineerr` entry                             | -1         |
| `pkg/constants/linter_reasons.go`      | Removed dead `noinlineerr` entry                             | -1         |
| `pkg/constants/data_integrity_test.go` | New `DisabledLinters` Describe block (2 tests)               | +18        |
| `pkg/linter/categorizer_test.go`       | New `linterSetLllNoinline` var + noinlineerr skip test       | +12        |
| `pkg/linter/fixer_test.go`             | New "Disabled Linters" Context with fixer test               | +19        |
| **Total**                              |                                                              | **+53/-3** |

---

## Resolution (2026-07-25)

The follow-ups flagged here were resolved by the next session (`2026-07-10_14-13_DISABLED-LINTERS-REASON-UPGRADE.md`):

- **DisabledLinters reason map** (the "Set has no reason" gap): ✅ upgraded `Set` → `map[LinterName]string` (commit `31df177`); reason strings now surface in fixer/categorizer logs.
- **`scripts/validate_linter_data.go` invariant** (item f#5): ✅ added — enforces non-empty reasons + no-priority/reason-for-disabled checks.
- **gosec G204 nolints** (items e#5/f#9): ✅ `//nolint:gosec` added (commit `8d10df5`).
- **AGENTS.md #10** (DisabledLinters is static data): ✅ documented.

**Still open:** dry-run logging inconsistency; `funcorder` test gap. See `TODO_LIST.md`.
