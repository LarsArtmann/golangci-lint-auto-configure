# Status Report: Never-Enable Enforcement Fix & Regression Loop Prevention

> **Date:** 2026-07-30 23:22
> **Session scope:** Fix BUG 1 (never-enable bypass in enforcement), update docs, review auto-commits
> **Previous session:** `2026-07-30_22-39_regression-loop-prevention-never-enable-cycle-detection.md`

---

## What This Session Did

The previous session implemented two-layer regression loop prevention (audit-ledger cycle detection + `never-enable` sidecar section) but left a **critical bug** unfixed: `enforceDisableReasons` bypassed `never-enable`. This session:

1. **Fixed BUG 1** — Added `IsNeverEnable` check to `tryReEnableLinter` in `fixer_enforce.go`
2. **Wrote 2 tests** for the fix (enforcement-level + unit-level)
3. **Updated 5 docs** (README, FEATURES, DOMAIN_LANGUAGE, AGENTS gotcha #27, feedback resolution)
4. **Reviewed auto-committed error-family migrations** — legitimate, kept
5. **Refactored test** to satisfy gocognit complexity limit (extracted sub-test to standalone function)

---

## a) FULLY DONE

| Item | Details |
|------|---------|
| **BUG 1 fix** | `tryReEnableLinter` now checks `f.pol.IsNeverEnable(linter)` before re-enabling (`fixer_enforce.go:75`). Returns false + debug log. Never-enable takes priority over anti-gaming enforcement. |
| **Test: enforcement-level** | `TestEnforceDisableReasons_NeverEnableOverridesUnjustified` — linter in both `disabled` (unjustified) + `never-enable` stays disabled; only unjustified non-never-enable linters re-enabled |
| **Test: unit-level** | `TestTryReEnableLinter_NeverEnable` — standalone test (extracted from `TestTryReEnableLinter` to satisfy gocognit ≤25) |
| **README.md** | New `never-enable:` subsection with YAML example, explanation of both layers, how cycle detection works alongside the sidecar |
| **FEATURES.md** | Two new rows: `never-enable sidecar section` + `Regression loop detection`; audit date bumped to 2026-07-30 |
| **DOMAIN_LANGUAGE.md** | Three new glossary terms (Sidecar, Never-Enable, Regression Loop), two value objects (DisableJustification, AuditEntry), updated Policy bounded context description |
| **AGENTS.md gotcha #27** | Added paragraph: never-enable checked in both code paths (recommendation + enforcement); explains why enforcement check is critical |
| **Feedback resolution doc** | Updated section (d) with enforcement-path detail; added 2 new test names to Tests section |
| **Error-family migration review** | Reviewed `cmd_audit.go`, `cmd_configure_config.go`, `cmd_presets.go` diffs. All replace bare `fmt.Errorf` with `WrapClassified`/`WrapRejectionf`/`WrapCorruptionf`. Consistent with project architecture (AGENTS.md gotcha #5). **Kept — no action needed.** |
| **Full test suite** | 18 packages pass (pkg + internal). 0 failures. |
| **Lint** | 0 issues on changed packages (linter, policy, audit). gocognit resolved by extraction. |
| **Coverage** | policy 88.5%, audit 78.8%, linter 85.0% — all above 60% gate |
| **Build** | `GOEXPERIMENT=jsonv2 go build ./...` — clean |

---

## b) PARTIALLY DONE

| Item | What's done | What's missing |
|------|-------------|----------------|
| **Test coverage of never-enable** | Unit tests cover `tryReEnableLinter` + `enforceDisableReasons` + `enableRecommendedLinters` in isolation | No integration test through the full `FixConfig` pipeline (see section c) |
| **Documentation** | README, FEATURES, DOMAIN_LANGUAGE, AGENTS updated | `docs/references/working-with-codebase.md` not checked for stale sidecar references |

---

## c) NOT STARTED

| Item | Why it matters |
|------|----------------|
| **Integration test: full `FixConfig` flow** | The BUG 1 scenario was a pipeline-ordering bug (`enforceDisableReasons` runs AFTER `enableRecommendedLinters` in `applyAndSave`). Unit tests prove each function works, but don't prove the pipeline as a whole respects never-enable end-to-end. A test that creates a real sidecar file + config, calls `FixConfig`, and asserts the linter is absent from the output would close this gap. |
| **Deprecated linter replacement → never-enable interaction** | `replaceLinters` (deprecated handler) adds successor linters to the enable set. If a deprecated linter's successor is in `never-enable`, does the replacement respect it? **Not tested.** Likely NOT — `replaceLinters` operates on `linterSet` directly and doesn't check `f.pol`. |
| **Never-enable linter already in `enable`** | Contradictory state: user has a linter in both `enable` and `never-enable`. Current behavior: tool doesn't remove it (never-enable only blocks adding, not removes existing). No warning logged. Could confuse users. |
| **`--pragmatic` + `never-enable` composition test** | Both filter the recommendation set independently. Should compose correctly but untested together. |
| **Updating previous status report** | `docs/status/2026-07-30_22-39_...md` still lists BUG 1 as unfixed and 3 questions as unanswered. |

---

## d) TOTALLY FUCKED UP

Nothing in this session. The previous session left BUG 1 unfixed, but that's now resolved. No regressions introduced.

**One self-criticism:** I should have written the integration test BEFORE declaring done. The unit tests give false confidence — they prove `tryReEnableLinter` works in isolation but the actual bug was about the pipeline calling it at the wrong time. If `enforceDisableReasons` had been called from a different location, or if a future refactor moves it, the unit tests would still pass while the pipeline breaks.

---

## e) WHAT WE SHOULD IMPROVE

### Process

1. **Always write integration tests for pipeline-ordering bugs.** BUG 1 was fundamentally about call ordering in `applyAndSave`. Unit tests on individual functions cannot catch this class of bug. The test must exercise the full `FixConfig` → save → reload cycle.

2. **Review the entire mutation pipeline before declaring a feature done.** The previous session identified BUG 1 but didn't fix it. When a bug is identified during development, fix it immediately — don't defer to a "next steps" list.

3. **The auto-commit daemon obscures change attribution.** Commits like `573ccd4 refactor(cli): improve configure fixer command implementation` don't describe what actually changed. This makes review difficult. Consider squashing or amending daemon commits with descriptive messages.

### Architecture

4. **`never-enable` should be checked in ALL code paths that add to `enable`.** Currently checked in `enableRecommendedLinters` and `tryReEnableLinter`. But `replaceLinters` (deprecated handler) and preset application also add linters. A centralized `canAddToEnable(linter)` guard would be more robust than scattering checks.

5. **Contradictory state detection.** A linter in both `enable` and `never-enable` is a user error. The tool should warn (not silently ignore the contradiction).

6. **`ledgerReader` duck-typing risk.** If a wrapper struct around `*Ledger` doesn't forward `PreviouslyAutoEnabled`, cycle detection silently disables. Consider adding it to the `audit.Recorder` interface proper (breaking change, but safer).

### Testing

7. **Table-style test for `tryReEnableLinter`.** The current 4 sub-tests/standalone tests are repetitive. A table-driven approach with cases (tool-managed, justified, never-enable, unjustified) would be more maintainable.

8. **Policy package has no test for `never-enable` + `disabled` overlap.** What if a linter is in both sections? Current behavior: `IsJustified` checks `Disabled` map only; `IsNeverEnable` checks `NeverEnable` map only. No conflict detection.

---

## f) Up to 50 Things to Get Done Next

### Critical (blocks trust in the feature)

1. **Write integration test through full `FixConfig` flow** — real sidecar file + config on disk, call `FixConfig`, assert never-enable linter absent from saved config
2. **Test deprecated linter replacement respects never-enable** — if `wsl` → `wsl_v5` and `wsl_v5` is in never-enable, the replacement must not add it
3. **Consider centralized `canAddToEnable(linter)` guard** — single chokepoint instead of scattered checks in `enableRecommendedLinters`, `tryReEnableLinter`, `replaceLinters`

### High value

4. **Add warning for contradictory state** — linter in both `enable` and `never-enable`
5. **Test `--pragmatic` + `never-enable` composition**
6. **Update previous status report** (`2026-07-30_22-39_...md`) to mark BUG 1 as resolved
7. **Check `docs/references/working-with-codebase.md`** for stale sidecar references
8. **Add `never-enable` to the `audit` subcommand output** — show which linters are suppressed and why
9. **Add `never-enable` validation to `validate` command** — warn if a never-enable linter has settings still in the config (orphaned settings)

### Medium value

10. **Table-driven refactor of `tryReEnableLinter` tests** — reduce duplication
11. **Test policy parsing of sidecar with both `disabled` and `never-enable`** — ensure both parse correctly
12. **Add `never-enable` count to `configure` output summary** — "Skipped 2 never-enable linters"
13. **Document the three-tier linter governance in README** — Disabled vs NeverAutoEnable vs PragmaticNoise vs never-enable (sidecar) — these are confusingly similar
14. **Add `never-enable` to `--dry-run` output** — show which linters would be skipped
15. **Consider `never-enable` glob/pattern support** — e.g., `go*` to block all go-prefixed linters (YAGNI? maybe)
16. **Add integration test for cycle detection** — write to real ledger, run configure, verify suppression
17. **Test cycle detection after 90-day purge** — ledger entries expire, cycle detection stops working (expected, but should be documented in tests)
18. **Add `--list-never-enable` flag** — show current never-enable entries from sidecar
19. **Consider machine-readable sidecar validation** — JSON schema for `.golangci-lint-auto-configure.yml`
20. **Add never-enable to HTML report** — show suppressed linters in the report

### Low value / polish

21. **Rename `DisableJustification` to `LinterJustification`** — it's reused for both `disabled` and `never-enable`, the name is misleading
22. **Add `NeverEnableJustification` accessor to enforcement log** — when skipping, log the reason from the sidecar
23. **Consistent log formatting** — enforcement uses `📋`, cycle detection uses `⚠️`, never-enable recommendation uses `Debugf`. Standardize.
24. **Add `never-enable` to `presets` output** — show which preset linters are blocked by sidecar
25. **Test sidecar with empty `never-enable:` section** — YAML parses to nil map, should be no-op
26. **Test sidecar with unknown linter in `never-enable`** — should be silently ignored (not a recommendation anyway)
27. **Add `--strict-never-enable` flag** — error if a never-enable linter is found in `enable` (current: silent)
28. **Document never-enable in `docs/references/error-handling.md`** — if applicable
29. **Consider `never-enable-formatters` section** — same concept for formatters
30. **Add never-enable interaction test with `--check` mode** — does `--check` report never-enable linters as "missing"?

### Infrastructure / maintenance

31. **Vendor hash update** — if go.mod changed, `nix build` will need new vendorHash
32. **Run `nix flake check`** — full reproducibility verification
33. **Run full `golangci-lint run` on entire project** — not just changed packages
34. **Check `cmd/coverage-check` threshold** — confirm overall coverage didn't drop
35. **Update CHANGELOG.md** with never-enable feature
36. **Consider adding never-enable to the pre-commit hook** — warn if sidecar is missing never-enable for known-problematic linters
37. **Add never-enable to TODO_LIST.md** as completed
38. **Review if `ireturn` should move to `NeverAutoEnableLinters`** — decided NO this session, but worth a deeper analysis across 160 projects
39. **Audit all code paths that mutate `cfg.Linters.Enable`** — ensure none bypass never-enable
40. **Add fuzz test for policy parsing** — malformed YAML edge cases
41. **Benchmark never-enable check** — `IsNeverEnable` is called per-linter; verify no perf impact on large configs
42. **Consider caching policy parse result** — `loadPolicy` reads file every run; could cache by mtime
43. **Add never-enable to SARIF report** — as informational findings
44. **Test never-enable with relative config path** — `loadPolicy` uses `filepath.Dir(configPath)`
45. **Document the priority order: tool-level > never-enable > justified > unjustified** — in code comments and docs
46. **Add `--explain-never-enable <linter>` flag** — show why a linter is in never-enable (reads sidecar)
47. **Consider never-enable expiry** — sidecar entries with optional `until:` date (YAGNI?)
48. **Add never-enable statistics to audit ledger** — count of suppressed re-enables over time
49. **Test concurrent configure runs** — ledger read/write race conditions
50. **Add never-enable to `install-hook` generated pre-commit script** — document the sidecar in the hook output

---

## g) Questions (that I CANNOT figure out myself)

### 1. Should `replaceLinters` (deprecated linter replacement) respect `never-enable`?

**Context:** When a deprecated linter is replaced (e.g., `wsl` → `wsl_v5`), the successor is added to the enable set via `handler.replaceLinters`. This code path does NOT check `f.pol.IsNeverEnable`. If `wsl_v5` is in `never-enable`, it would be added anyway — same regression loop, different entry point.

**Why I can't decide:** The deprecated replacement is a direct mapping (old → new), not a recommendation. Blocking it could leave the user with a deprecated linter enabled. But adding it creates the same loop the feature is designed to prevent.

**What I'd do if you say yes:** Add `f.pol.IsNeverEnable` check in `replaceLinters` or extract a `canAddToEnable` guard.

### 2. Should the integration test use a real `*audit.Ledger` (writes to disk) or a test double?

**Context:** The integration test for the full `FixConfig` flow needs to verify cycle detection works end-to-end. This requires either a real ledger writing to a temp directory, or extending the test double to satisfy `ledgerReader`. The real ledger tests the actual JSONL read/write path but is slower and more brittle. The test double is faster but doesn't prove the real ledger's `PreviouslyAutoEnabled` works in the pipeline.

**Why I can't decide:** Both approaches have merit. The existing `ledger_test.go` already tests `PreviouslyAutoEnabled` with real files at the audit-package level. But the wiring between `SetLedger` → `reader` field → `enableRecommendedLinters` is only proven if the real type flows through.

### 3. Should never-enable linters that are already in `enable` be forcibly removed?

**Context:** If a user has `godoclint` in `linters.enable` AND in `never-enable:` in the sidecar, this is contradictory. Currently the tool silently ignores the contradiction (never-enable only blocks adding, never removes). Three options: (a) warn and leave it, (b) warn and remove it, (c) error and abort.

**Why I can't decide:** Option (b) is the most useful (self-healing) but could surprise users who intentionally have it in both (transitioning state). Option (a) is safest but the contradiction could mask a misconfiguration. This is a product decision about user intent.
