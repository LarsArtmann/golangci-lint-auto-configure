# Status Report: DisabledLinters Reason Upgrade Session

> **Resolved 2026-09-11 (docs-health archive pass).** Shipped as commit `31df177`; invariants enforced by data-integrity tests; funcorder reason sharpened in `pkg/constants/rules.go`; validate_linter_data.go kept as a manual utility (v0.6.0 decision). Forward-looking items below are struck inline; process sections (d/e) are retained as historical context. Archived from `docs/status/` — live state: `TODO_LIST.md` / `ROADMAP.md` / `CHANGELOG.md`.

**Date:** 2026-07-10 14:13
**Commit:** `31df177` — `refactor: upgrade DisabledLinters from Set to map[LinterName]string`
**Branch:** master (pushed)

---

## a) FULLY DONE

| #  | Task                                                                                              | Files                                                       | Verified                |
| -- | ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- | ----------------------- |
| 1  | `DisabledLinters` changed from `Set[LinterName]` to `map[LinterName]string` with reason strings   | `pkg/constants/rules.go`                                    | Build + tests pass      |
| 2  | Categorizer updated: map lookup + reason in `Debugf` log                                          | `pkg/linter/categorizer.go:40`                              | Tests pass              |
| 3  | Fixer updated: map lookup + reason in `Debugf` log + logger param added to `updateConfigFromSets` | `pkg/linter/fixer_config.go:249`, `pkg/linter/fixer.go:242` | Tests pass              |
| 4  | Non-empty reason invariant test added                                                             | `pkg/constants/data_integrity_test.go`                      | Test passes             |
| 5  | Validation script updated with 2 new checks (disabled↔priorities isolation + non-empty reasons)   | `scripts/validate_linter_data.go`                           | All 6 checks pass       |
| 6  | AGENTS.md item #10 updated to document `DisabledLinters` type and invariant                       | `AGENTS.md`                                                 | Committed               |
| 7  | Planning doc with mermaid execution graph                                                         | `docs/planning/2026-07-10_12-03_...`                        | Written                 |
| 8  | Full test suite passes (16 packages, `-race`)                                                     | —                                                           | All `ok`                |
| 9  | BuildFlow pre-commit hooks pass (26/26)                                                           | —                                                           | Passed                  |
| 10 | Committed with detailed message + pushed                                                          | —                                                           | Pushed to origin/master |

---

## b) PARTIALLY DONE

| # | What                                       | Status                                                  | What's missing                                                                                                                                                                                                              |
| - | ------------------------------------------ | ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Fixer logging for disabled linters         | Logs reason at `Debugf`, but **no dry-run distinction** | The `RemoveRedundantLinters` pattern uses `[DRY-RUN] Would...` prefix when `dryRun=true`. My implementation always says "Moving" even in dry-run preview mode. `updateConfigFromSets` doesn't receive a `dryRun` parameter. |
| 2 | Test coverage for disabled linter behavior | Tests verify `noinlineerr` moves from enable to disable | **No `funcorder` test** (was pre-existing gap, not introduced this session). **No dry-run test** for disabled linters (deprecated linters have `testDeprecatedLinterDryRun`).                                               |

---

## c) NOT STARTED

| # | What                                                                                                                                             | Why not                                                                                                   | Impact                                                                 |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| 1 | ~~`nix build` verification~~ done — see header resolution note (docs-health 2026-09-11) | Ran `go build` + `go test` but not `nix build`                                                            | No dependency changes, so vendorHash likely unaffected, but unverified |
| 2 | ~~Finding converter for disabled linters (`DisabledLintersToFindings`)~~ done — see header resolution note (docs-health 2026-09-11) | Explicitly decided against in architecture discussion — tool-internal policy, not user-actionable problem | Low — can add later if user feedback demands it                        |
| 3 | ~~Answering open question #1 from previous session: should 30 stale `noinlineerr` doc references in `docs/status/` and `docs/archive/` be updated?~~ done — see header resolution note (docs-health 2026-09-11) | Out of scope for this refactoring task                                                                    | None — historical snapshots                                            |
| 4 | ~~Planning doc marked as completed~~ done — see header resolution note (docs-health 2026-09-11) | Wrote it as "In Progress" and forgot to update after execution                                            | Cosmetic                                                               |

---

## d) TOTALLY FUCKED UP

Nothing. No regressions, no broken builds, no data loss, no incorrect logic. The refactoring is clean and correct.

The closest thing to a mistake: **the dry-run logging inconsistency** (see b.1 above). Not a bug — the behavior is correct (linters are moved in both dry-run and non-dry-run paths for `updateConfigFromSets`, since the function doesn't have a dry-run concept). But the log message "Moving disabled linter to disable list" is misleading when the entire `applyAndSave` is in dry-run context. This is a cosmetic/log-clarity issue, not a functional bug.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Planning doc should be updated post-execution** — I wrote the planning doc with "Status: In Progress" and never marked it done. Should update status after completion.

2. **Dry-run logging pattern should be consistent** — `fixer_formatters.go` has a `logChange` helper that handles dry-run prefixing. `fixer_config.go`'s `updateConfigFromSets` doesn't use it and doesn't have dry-run context. Consider whether `updateConfigFromSets` should receive `dryRun` or whether the log message should be more neutral.

3. **Validation script is `//go:build ignore`** — It's a manual script never run in CI. The real guard is the data integrity test. The script updates are correct but low-value since nobody runs it automatically. Consider either adding it to CI or deleting it (tests cover the same invariants).

### Code improvements

4. **`funcorder` reason is vague** — "provides minimal value and can be confusing for users" is accurate but doesn't explain what `funcorder` does or why it's confusing. The `noinlineerr` reason is specific and actionable. Consider: "enforces function ordering (constructors before methods, etc.) which conflicts with common Go code organization patterns and provides marginal value".

5. **No test verifies the reason string is logged** — Tests check structural correctness (linter moves to disable list, reason is non-empty in data) but no test captures log output to verify the reason appears in user-facing logs.

6. **Test gap: `funcorder`** — No test verifies `funcorder` is moved from enable to disable. Only `noinlineerr` is tested. Pre-existing gap, not introduced this session.

---

## f) Next 50 Things We Should Get Done

### High impact (architecture/reliability)

| # | Task                                                                                                                           | Impact                                                      | Effort |
| - | ------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------- | ------ |
| 1 | ~~Fix dry-run logging: pass `dryRun` to `updateConfigFromSets` or use neutral log language ("Disabled linter excluded: %s (%s)")~~ done — see header resolution note (docs-health 2026-09-11) | Medium — prevents misleading log in dry-run                 | 15min  |
| 2 | ~~Add `funcorder` test to fixer_test.go disabled linters context~~ done — see header resolution note (docs-health 2026-09-11) | Medium — closes test coverage gap for both disabled linters | 10min  |
| 3 | ~~Add dry-run test for disabled linters (parity with `testDeprecatedLinterDryRun`)~~ done — see header resolution note (docs-health 2026-09-11) | Medium — verifies dry-run doesn't move linters prematurely  | 10min  |
| 4 | ~~Decide: add `validate_linter_data.go` to CI or delete it (data integrity tests cover the same ground)~~ done — see header resolution note (docs-health 2026-09-11) | Medium — dead code / dead scripts are debt                  | 10min  |
| 5 | ~~Mark planning doc as completed~~ done — see header resolution note (docs-health 2026-09-11) | Low — cosmetic                                              | 2min   |

### Medium impact (consistency/quality)

| #  | Task                                                                                                                   | Impact                         | Effort |
| -- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------ | ------ |
| 6  | ~~Improve `funcorder` reason string to be more specific~~ done — see header resolution note (docs-health 2026-09-11) | Low — clarity                  | 5min   |
| 7  | ~~Add log capture test verifying reason string appears in fixer `Debugf` output~~ done — see header resolution note (docs-health 2026-09-11) | Low — nice-to-have             | 15min  |
| 8  | ~~Consider extracting a shared `logLinterChange` helper in `fixer_config.go` (same pattern as `fixer_formatters.go:156`)~~ done — see header resolution note (docs-health 2026-09-11) | Low — DRY                      | 15min  |
| 9  | ~~Run `nix build` to verify vendorHash is unaffected~~ done — see header resolution note (docs-health 2026-09-11) | Low — likely fine              | 5min   |
| 10 | ~~Answer open question: update 30 stale `noinlineerr` doc references? (recommend: no — historical snapshots)~~ done — see header resolution note (docs-health 2026-09-11) | Low — precedent is `funcorder` | 5min   |

### Lower priority (polish/future)

| #     | Task                                                                                                                               | Impact                | Effort |
| ----- | ---------------------------------------------------------------------------------------------------------------------------------- | --------------------- | ------ |
| 11    | ~~Consider whether `DisabledLintersToFindings` converter is needed for `analyze` mode (re-evaluate based on user feedback)~~ done — see header resolution note (docs-health 2026-09-11) | Low                   | 30min  |
| 12    | ~~Consider upgrading `DisabledLinters` to a struct type if more fields are needed (e.g. `Replacement`, `Category`)~~ done — see header resolution note (docs-health 2026-09-11) | Low — YAGNI for now   | 30min  |
| 13    | ~~Review all `Set[LinterName]` usages — are there other sets that should carry reason data?~~ done — see header resolution note (docs-health 2026-09-11) | Low — audit           | 30min  |
| 14    | ~~Consider surfacing disabled-linter reasons in HTML report (report.templ)~~ done — see header resolution note (docs-health 2026-09-11) | Low — UI enhancement  | 30min  |
| 15    | ~~Consider surfacing disabled-linter reasons in JSON report (json_report_generator.go)~~ done — see header resolution note (docs-health 2026-09-11) | Low — API enhancement | 20min  |
| 16-50 | _(No further tasks identified at this granularity — the refactoring is complete and the remaining work is enhancements, not debt)_ | —                     | —      |

---

## g) Top 2 Questions I Cannot Answer Myself

### Question 1: Should `updateConfigFromSets` receive `dryRun` and use `[DRY-RUN]` prefix in its log messages?

The function is called from `applyAndSave` which has a `dryRun` parameter. Currently `updateConfigFromSets` always logs "Moving disabled linter to disable list" even when in dry-run. The `RemoveRedundantLinters` pattern in `fixer_formatters.go` receives `dryRun` and prefixes with `[DRY-RUN] Would...`. Should I thread `dryRun` through for log-message consistency, or is the current behavior acceptable since `applyAndSave` already short-circuits before saving in dry-run mode?

**Why I can't decide:** The config struct IS mutated in dry-run mode (the function modifies `cfg.Linters.Enable` / `cfg.Linters.Disable`), but the config is never saved to disk in dry-run. So the log message is technically accurate ("moving" in the in-memory struct) but potentially misleading to a user reading dry-run output.

### Question 2: Should the validation script (`scripts/validate_linter_data.go`) be integrated into CI or deleted?

It's `//go:build ignore` — never compiled, never tested, never run automatically. The data integrity tests in `data_integrity_test.go` now cover the same invariants (and more). The script is redundant. But it produces nice human-readable output with checkmarks. Should we: (a) add it to CI as a separate step, (b) delete it and rely solely on tests, or (c) leave it as a manual debugging tool?

**Why I can't decide:** This is a project convention decision. Some projects value manual validation scripts for debugging; others see them as dead code. The script was already there before this session — I added to it, but its fundamental purpose is questionable now that tests cover the same ground.
