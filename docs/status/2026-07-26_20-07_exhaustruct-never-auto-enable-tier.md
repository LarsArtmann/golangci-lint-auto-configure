# Status Report: exhaustruct NeverAutoEnable Tier

**Date:** 2026-07-26 20:07
**Session scope:** Move `exhaustruct` from auto-enabled (High priority, in reference preset) to a new `NeverAutoEnableLinters` tier — never auto-enabled, but respected + safe-defaulted when manually added.

---

## What Was Requested

> exhaustruct should NOT get enabled, if somebody adds it it should not be removed but it should always ignore *_test.go files

Three requirements:

1. **Never auto-enable** exhaustruct
2. **Never strip** exhaustruct if a user manually adds it
3. **Always ignore** `*_test.go` files (already satisfied via `DefaultExclusionRules`)

---

## a) FULLY DONE

### Core mechanism — `NeverAutoEnableLinters` map

| What                                                 | File                                    | Detail                                                                                                                                                                                                    |
| ---------------------------------------------------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| New `NeverAutoEnableLinters` map                     | `pkg/constants/rules.go`                | Third management tier: `exhaustruct` is the sole entry. Never auto-enabled, but not forcibly disabled.                                                                                                    |
| Removed exhaustruct from `PragmaticNoiseLinters`     | `pkg/constants/rules.go`                | Was 5 linters → now 4. exhaustruct is handled unconditionally now, not just with `--pragmatic`.                                                                                                           |
| `isNeverAutoEnable` skip in categorizer              | `pkg/linter/categorizer.go:53-55`       | Unconditional skip — independent of `--pragmatic` flag.                                                                                                                                                   |
| `isNeverAutoEnable` method                           | `pkg/linter/categorizer.go:68-76`       | Checks `NeverAutoEnableLinters` map, logs debug message.                                                                                                                                                  |
| Downgraded exhaustruct priority High → Medium        | `pkg/constants/linter_priorities.go:72` | No longer in the "all critical + high must be in reference preset" invariant.                                                                                                                             |
| Removed exhaustruct from `reference` preset          | `pkg/constants/presets.go:40`           | Was 62 linters → now 61.                                                                                                                                                                                  |
| Renamed `isToolLevelDisabled` → `isToolLevelManaged` | `pkg/linter/fixer_enforce.go:89-97`     | Now checks both `DisabledLinters` AND `NeverAutoEnableLinters`. Prevents the policy enforcer from stripping a manually-added exhaustruct from the disable list (or re-enabling it without justification). |

### Test coverage

| What                                                        | File                                          |
| ----------------------------------------------------------- | --------------------------------------------- |
| `TestIsToolLevelManaged` — renamed + added exhaustruct case | `pkg/linter/fixer_enforce_test.go:66-88`      |
| Categorizer pragmatic tests updated (4 linters, not 5)      | `pkg/linter/categorizer_test.go:161-194`      |
| New "NeverAutoEnable Linters" test context (2 specs)        | `pkg/linter/categorizer_test.go:196-215`      |
| `NeverAutoEnableLinters` data integrity block (5 specs)     | `pkg/constants/data_integrity_test.go:83-127` |

### Data integrity constraints enforced

- NeverAutoEnable entries MUST be in `LinterPriorities` (they can be manually enabled)
- NeverAutoEnable entries MUST be in `LinterReasons` (reports need them)
- NeverAutoEnable entries MUST have non-empty reason
- NeverAutoEnable MUST NOT overlap with `DisabledLinters`
- NeverAutoEnable MUST NOT overlap with `PragmaticNoiseLinters`

### Living docs updated

| File                            | What changed                                                                                                                       |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `FEATURES.md`                   | Pragmatic row (5→4), new NeverAutoEnableLinters row, exhaustruct defaults row notes "when manually enabled", reference count 62→61 |
| `README.md`                     | Removed `exhaustruct` from the reference preset linter list                                                                        |
| `AGENTS.md`                     | Gotcha #10 (three-tier management), Gotcha #19 (pragmatic is 4, exhaustruct is NeverAutoEnable)                                    |
| `pkg/linter/analyzer.go`        | `SetPragmatic` comment updated (5→4)                                                                                               |
| `internal/cli/cmd_configure.go` | `--pragmatic` help text (5→4)                                                                                                      |

### Verification

- `go build ./...` — passes
- `go test ./pkg/... ./internal/... ./cmd/...` — all packages green
- `golangci-lint run` on changed packages — 0 issues

---

## b) PARTIALLY DONE

### `scripts/validate_linter_data.go` — NOT updated

I read the script (line 91-109 checks `DisabledLinters` consistency) but did NOT add a `NeverAutoEnableLinters` validation block. The Ginkgo data integrity tests cover the constraints at test time, but the standalone script (`go run scripts/validate_linter_data.go`) is now incomplete — it validates `DisabledLinters` but not `NeverAutoEnableLinters`. The script is a secondary safety net; the Ginkgo tests are the real gate, but the script should mirror them for consistency.

### `CHANGELOG.md` — NOT updated

`CHANGELOG.md:14` still says "drops the 5 highest-noise linters (exhaustruct, ...)". The `[Unreleased]` section should document:

- New `NeverAutoEnableLinters` concept
- exhaustruct moved from auto-enabled to never-auto-enable
- `--pragmatic` now drops 4 linters instead of 5
- `reference` preset now 61 linters instead of 62

---

## c) NOT STARTED

1. **Integration test for manual exhaustruct enable.** No integration test verifies the full round-trip: user manually adds exhaustruct → configure runs → exhaustruct stays enabled → safe defaults injected → test exclusion rules present.
2. **CLI acceptance test for `--pragmatic`.** The planning doc (`docs/planning/2026-07-25_07-56_golangci-lint-friction-reduction-pareto-plan.md:F31`) has an open item: "Add CLI acceptance test: `--pragmatic` produces config without the noise linters." This was never done for the original 5, and is still not done for the new 4.
3. **`docs/DOMAIN_LANGUAGE.md`** — not checked for whether the three-tier linter management model should be documented there.

---

## d) TOTALLY FUCKED UP

Nothing. All tests pass, lint is clean, build succeeds. The core behavioral change is correct and verified.

**However — one design smell I should flag:**

The `DefaultExclusionRules` test-integrity check (`data_integrity_test.go:436`) asserts that exclusion-rule linters are NOT in `DisabledLinters`. It does NOT check against `NeverAutoEnableLinters`. This is **correct** — exhaustruct SHOULD be in exclusion rules (it's excluded from test files even when manually enabled). But the test description says "should not include any tool-level disabled linters" — which is now ambiguous since there are two tiers of "tool-level managed". The test name should say "DisabledLinters" specifically, not the vague "tool-level disabled".

---

## e) WHAT WE SHOULD IMPROVE

### Design improvements

1. **The three-tier naming is slightly confusing.** `DisabledLinters` (never enabled, forcibly disabled), `NeverAutoEnableLinters` (never auto-enabled, respected if manual), `PragmaticNoiseLinters` (opt-out via flag). The first two sound similar but have opposite semantics for manual users. Consider documenting the taxonomy in a single place (e.g., a comment block in `rules.go` or `DOMAIN_LANGUAGE.md`).

2. **`isToolLevelManaged` is doing double duty.** It checks two maps with different semantics. The function is correct, but callers might not realize it covers both "forcibly disabled" and "never auto-enabled but respected". The comment explains it, but a future reader might wonder why a never-auto-enable linter is exempt from policy enforcement.

3. **The `data_integrity_test.go:436` test name is stale.** It says "should not include any tool-level disabled linters" but `tool-level disabled` is ambiguous now that `isToolLevelManaged` covers two tiers.

### Process improvements

4. **I initially rushed the `linter_priorities.go` edit.** The `edit` tool failed twice because I didn't match the exact whitespace alignment of the Medium section (the `gofmt` alignment columns differ from what I typed). I fell back to `sed`, which worked but is a code-smell — I should have read the exact bytes with `cat -A` first. This wasted 3 tool calls.

5. **The `multiedit` on `FEATURES.md` had a cascading failure.** One of the 4 edits matched the wrong occurrence (the `revive defaults` line appeared twice after the first edit consumed the original), causing a 4th edit to fail. I had to manually fix it. I should have used more unique context strings in the multiedit.

---

## f) Up to 50 Things We Should Get Done Next

### Immediate follow-ups (this session's work)

1. **Update `CHANGELOG.md` `[Unreleased]`** — document the NeverAutoEnableLinters concept, exhaustruct move, pragmatic 5→4, reference 62→61
2. **Update `scripts/validate_linter_data.go`** — add `NeverAutoEnableLinters` consistency checks mirroring the Ginkgo tests
3. **Fix `data_integrity_test.go:436` test name** — "should not include any DisabledLinters" instead of vague "tool-level disabled"
4. **Add integration test** — user manually enables exhaustruct → `configure` preserves it + injects defaults + test exclusions
5. **Add CLI acceptance test for `--pragmatic`** — produces config without the 4 noise linters (F31 from planning doc)
6. **Verify `docs/DOMAIN_LANGUAGE.md`** — add three-tier linter management taxonomy if appropriate

### Short-term improvements

7. **Consider whether `ireturn` should also be NeverAutoEnable** — it has friction 1.3 but is in PragmaticNoiseLinters; evaluate if the opt-in flag is the right tier
8. **Consider whether `gochecknoglobals` should be NeverAutoEnable** — friction 5.1, second only to exhaustruct; currently in PragmaticNoiseLinters
9. **Audit all `docs/research/` and `docs/planning/` files** — they reference "5 noise linters" and "exhaustruct enabled by default" — decide whether to annotate as historical or update
10. **Review the `fixer_audit.go:56` site** — it looks up `DisabledLinters[linter]` for the audit reason when moving to disable; NeverAutoEnable linters won't have a reason there (correct, but verify the audit output is sensible)
11. **Consider a `--list-never-auto-enable` CLI flag** — transparency for users to see which linters the tool refuses to auto-enable
12. **Add `NeverAutoEnableLinters` to the `presets` command output** — so users running `golangci-lint-auto-configure presets` see which linters are excluded from auto-enable
13. **Review `pkg/report/report.templ`** — the HTML report might list recommended linters; verify exhaustruct no longer appears there
14. **Check `pkg/config/settings_validator.go:30`** — it validates linter settings keys against `LinterPriorities`; since exhaustruct is still in Priorities (Medium), this works, but verify
15. **Review the `analyze` command output** — when analyzing an existing config, does it still recommend exhaustruct if the user has it disabled? It shouldn't now.

### Documentation debt

16. **Update `docs/references/working-with-codebase.md`** — "Adding commands/linters" section should document the three-tier system
17. **Update `docs/references/code-organization.md`** — if it references the linter management maps
18. **Review all `docs/status/` reports** that reference "exhaustruct enabled by default" — annotate as historical
19. **Update `docs/research/validation-delta.md`** — its "Resolution" section says `--pragmatic` is the mechanism for exhaustruct friction; this is now outdated (NeverAutoEnable is the mechanism)
20. **Update `docs/research/2026-07-25_golangci-config-ecosystem-report.md`** — recommendation #1 says "reconsider enabling exhaustruct by default"; this is now resolved differently than the report suggests

### Test debt

21. **Add a test for `isToolLevelManaged` with a NeverAutoEnable linter in the enforcement flow** — verify it's exempt from sidecar re-enable
22. **Add a test verifying `updateConfigFromSets` preserves a manually-enabled exhaustruct** — it should NOT be moved to the disable list
23. **Add a test verifying `injectDefaultSettings` still fires for a manually-enabled exhaustruct** — safe defaults should be injected
24. **Golden report test** — verify exhaustruct doesn't appear in the "recommended linters" section of the HTML report
25. **Coverage check** — verify the new `isNeverAutoEnable` method is covered

### Architectural considerations

26. **Consider unifying the three maps into a typed enum** — `map[LinterName]LinterManagementTier` where tier is `Disabled | NeverAutoEnable | PragmaticNoise | Default`. Would eliminate the "check 3 maps" pattern.
27. **Consider whether `NeverAutoEnableLinters` should affect the `--recommend` flow** — currently `--recommend` applies presets, which don't include exhaustruct anyway, but verify
28. **Consider whether the `migrate` command should handle exhaustruct** — when migrating v1→v2, if exhaustruct is enabled in v1, should it be preserved? Currently yes (migrate doesn't use categorizer skip logic)
29. **Review interaction with `--check` mode** — `--check` exits 1 if changes needed; since exhaustruct is no longer recommended, configs with it disabled won't trigger a "you should enable exhaustruct" diff
30. **Consider `gosec` friction** — gosec has 590 nolints (friction ~3.7); evaluate if it should get the same treatment as exhaustruct

### CI/CD and release

31. **Cut a release** — `TODO_LIST.md:15` notes ~30 unreleased commits since v0.5.0; this change is significant enough to warrant a version bump
32. **Verify CI pipeline** — the `test-and-build` job sets `GOEXPERIMENT: jsonv2`; confirm the new tests pass in CI
33. **Verify coverage gate** — `cmd/coverage-check` enforces ≥60%; the new code paths should maintain or improve coverage

### Polish

34. **Rename `isToolLevelManaged` callers' comments** — `fixer_enforce.go` comment says "Tool-level managed linters" which is good, but `fixer_enforce_test.go` test descriptions still say "tool-level disabled" for the DisabledLinters entries — minor inconsistency
35. **Add `NeverAutoEnableLinters` to the `analyze --json` output** — so programmatic consumers can detect which linters are tool-managed
36. **Consider a `--force-enable exhaustruct` flag** — for users who want the old behavior (auto-enable exhaustruct with safe defaults)
37. **Review `pkg/finding/categories.go:66`** — exhaustruct is mapped to `CategoryTypeSafety`; verify this is still correct for a never-auto-enable linter
38. **Consider whether other high-friction linters from the baseline data should join NeverAutoEnable** — the baseline shows `recvcheck` (friction ~2.5) is also high-friction
39. **Update the `--help` output for `configure`** — mention that exhaustruct is never auto-enabled but supported when manual
40. **Consider migration path** — existing configs that were auto-generated with exhaustruct enabled will now have it "stranded" (enabled but never recommended); should the tool warn?

### Research

41. **Re-run the friction baseline measurement** — with exhaustruct now never-auto-enabled, re-measure nolint counts across the 160 sibling projects to validate the impact
42. **Survey sibling projects** — how many currently have exhaustruct in their enable list (auto-configured by this tool)? Those configs are now divergent from what the tool would produce.
43. **Evaluate if the `ExhaustructSettings` (14 stdlib excludes) are still worth maintaining** — if exhaustruct is never auto-enabled, fewer projects will use it, but the settings are still injected for manual enablers
44. **Consider `wrapcheck` as NeverAutoEnable candidate** — friction 1.8 but heavy nolint volume; evaluate
45. **Document the decision rationale** — why NeverAutoEnable instead of just removing from presets + priorities? (Answer: it preserves the "respect manual additions" requirement)

### Cleanup

46. **Remove exhaustruct from `PresetLinters["reference"]` test expectations** — already done, but verify no other test hardcodes the old reference list
47. **Review `internal/cli/cmd_configure_test.go`** — any tests that assert exhaustruct appears in configure output
48. **Review `internal/cli/integration_test.go`** — any integration tests that exhaustruct appears in output
49. **Stale docs in `docs/archive/`** — many archived reports reference exhaustruct as enabled; these are historical and should NOT be updated (per update-old-docs skill philosophy)
50. **Add a comment in `rules.go`** explaining the relationship between the three maps and when to use each

---

## g) Questions

### 1. Should existing auto-generated configs that have exhaustruct enabled be migrated?

When this tool previously auto-configured projects, it added exhaustruct to the enable list. Those configs now have a linter that the tool would never add today. Should `configure` proactively remove exhaustruct from `linters.enable` on the next run (treating it like a deprecated recommendation), or should it be left alone (respecting it as a "user choice" even though the tool originally added it)?

### 2. Should `gochecknoglobals` (friction 5.1) also move to NeverAutoEnable?

`gochecknoglobals` is the second-highest friction linter (5.1) and has no config knobs in golangci-lint v2. It's currently in `PragmaticNoiseLinters` (opt-out via `--pragmatic`). Given that exhaustruct (6.5) just moved to unconditional NeverAutoEnable, should gochecknoglobals (5.1) follow? Or is the `--pragmatic` opt-out the right tier for it?

### 3. Should the `analyze` command warn when a config has a NeverAutoEnable linter enabled?

When a user runs `golangci-lint-auto-configure analyze` on a config that has exhaustruct enabled, should the analysis output include an informational note like "exhaustruct is enabled — this linter is never auto-configured; safe defaults will be injected but test-file exclusions apply"? Or should the tool be silently permissive?
