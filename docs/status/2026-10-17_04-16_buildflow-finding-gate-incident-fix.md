# Status Report: BuildFlow Finding-Gate Incident — Tag Crash, Detect/Apply Drift, Stale Binary

**Date:** 2026-10-17 04:16 CEST
**Session scope:** Diagnose and fix the failed `buildflow -s golangci-lint-auto-configure` run; close the loop end-to-end (repo fix → config convergence → BuildFlow binary refresh → green rerun).
**Verdict:** User's exact failing command now reports `✓ BuildFlow passed 1/1, 0 findings`. One real product bug found and fixed; two adjacent landmines defused.

---

## TL;DR

The pasted failure (`[corruption:finding.build]` — tag `exhaustruct->exhaustruct-v5` invalid) was the visible tip of a **three-layer incident**:

1. **Tag crash** — rename elements were fed to go-finding as tags; `>` violates the tag convention. Fixed earlier the same night (3ab5c92, `tagSafe` in `pkg/finding/deviation.go`), but the incident was never closed with a rerun.
2. **Real product bug (found this session)** — `finalizeConvergence` (`pkg/linter/fixer_converge.go`) **overwrote** `result.NormalizerChanges` with phase-one changes, silently discarding the redundancy pass (`RemoveRedundancies`, lll↔golines) that runs _only_ inside `Converge`. Result: apply reported _"already converged"_ while detect (report-only path) still planned the removal — a warning finding **no repair could ever clear**. This contradicted AGENTS.md gotcha 6's claim that "lint/apply drift is structurally dead": the parity corpus proved render parity, not normalizer-accounting parity. Live in published **v0.10.0** since ~Sep.
3. **Stale BuildFlow binary** — BuildFlow runs this repo's code **in-process** (SDK provider), pinned by BuildFlow's `flake.lock` input, not by the working tree. The installed binary predated the tag guard _and_ the supersession filter, so after the config was converged it flipped to a phantom _"lll is missing from linters.enable"_ finding. Refreshed: lock → d4447de, binary 5ccf8ea → **7dcb161**.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                   | Evidence                                                                                                                                        |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Root-caused the three-layer incident (tag crash → convergence overwrite → stale binary)                                                                                                                                | Full trace in this report; repro via in-process SDK probe                                                                                       |
| 2  | **Fixed the detect/apply divergence**: `mergeNormalizerChanges` unions phase-one persisted changes with Converge's in-memory normalizer changes by `Change.Key()`, replacing the discard-overwrite                     | `pkg/linter/fixer_converge.go:180-223`, commit 3703be4                                                                                          |
| 3  | Regression specs: apply parity (dry-run count == plan count), apply-clears-drift, **one-round convergence**, plus a 4-case union table                                                                                 | `pkg/linter/fixer_supersession_apply_test.go` (3 BDD specs, all pass), `pkg/linter/fixer_converge_merge_internal_test.go`                       |
| 4  | Converged the repo's own config: `lll` removed from `linters.enable` (superseded by golines)                                                                                                                           | Surgical 1-line diff in `.golangci.yml`; audit ledger entry `removed-from-enable` with redundancy reason (verified via `audit --linter lll`)    |
| 5  | Fixed 2 lint warnings on sight: `intrange` (deviation.go:171), `gci` import order (desired_supersession_test.go)                                                                                                       | `nix fmt` → 0 changed; `golangci-lint run` on `pkg/linter` + `pkg/finding` → **0 issues**                                                       |
| 6  | Full test suite green                                                                                                                                                                                                  | `ginkgo -r -p -race --skip-package=cmd/` → 18 suites SUCCESS (~1m45s)                                                                           |
| 7  | Pushed master per BuildFlow's flake sibling-push discipline (fast-forward, no force)                                                                                                                                   | `e0fcafe..d4447de` on origin/master                                                                                                             |
| 8  | Refreshed BuildFlow: `nix flake update golangci-lint-auto-configure` → lock e0fcafe→**d4447de**, `nix build`, profile reinstall (entry named `BuildFlow`, `--priority 6`)                                              | `buildflow version` → **7dcb161** embedding this repo at d4447de                                                                                |
| 9  | Final verification with the user's exact command, cold result cache                                                                                                                                                    | `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s golangci-lint-auto-configure` → **✓ BuildFlow passed 1/1, 0 findings**; binary-freshness warning gone |
| 10 | Memory updated: gotcha 6 extended (union-accounting rule + incident reference), **new gotcha 32** (BuildFlow in-process embedding, phantom-finding mode, full refresh loop incl. `nix profile remove BuildFlow` quirk) | `AGENTS.md` (uncommitted at report time)                                                                                                        |
| 11 | Cleanup: throwaway probe `tmp_sdkcheck/` and `/tmp` config backup removed                                                                                                                                              | `git status` clean except AGENTS.md                                                                                                             |

## b) PARTIALLY DONE

| Item                                    | What works                                     | What remains                                                                                                                                                             | Blocker                               | Effort |
| --------------------------------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------- | ------ |
| AGENTS.md documentation of the incident | Gotchas 6/32 written                           | Daemon commit pending at report time                                                                                                                                     | None (daemon)                         | S      |
| BuildFlow-side hygiene                  | `flake.lock` updated, binary rebuilt+installed | Lock change uncommitted in `~/projects/BuildFlow` (its own daemon should pick it up); installed 7dcb161 is a dirty-tree snapshot of HEAD 8971aec, not a clean HEAD build | Concurrent BuildFlow session activity | S      |
| Incident closure for the SDK contract   | In-repo parity specs added                     | No **cross-repo** contract test in BuildFlow asserting detector-zero-findings ⇔ FixConfig-converged; BuildFlow provider source unchanged                                 | None                                  | M      |
| Fleet protection from the v0.10.0 bug   | This repo + local BuildFlow fixed              | Published v0.10.0 still carries the bug; no release tagged; other fleet consumers' BuildFlow installs not audited                                                        | Release-timing decision (question g1) | M      |

## c) NOT STARTED (observed this session, deliberately not touched)

| Item                                                                                                                       | Why not started                                                   | Priority |
| -------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- | -------- |
| Tag + release the fixed version (go-release flow)                                                                          | Release timing is user policy; also needs version-number decision | High     |
| Harvest this report's (f) into `TODO_LIST.md`/`ROADMAP.md` (docs-health HARVEST)                                           | User said "then wait for instructions"                            | High     |
| CHANGELOG entry for the incident/fix                                                                                       | Same harvest gate                                                 | High     |
| Full BuildFlow pipeline rerun (all steps, not just the single step)                                                        | Single-step green was the session's goal; full run ~5-10 min      | Medium   |
| Audit of other `Converge`-only passes for the same discard class (codebase-wide sweep for `result.X =` overwrite patterns) | New follow-up work identified while fixing                        | High     |
| SDK-consumer documentation (`docs/references/integrations.md`) for the normalizer-accounting contract                      | Not blocking locally                                              | Medium   |
| Verify `VerifyConfigLoads` post-apply tripwire fired on the converged save                                                 | Not observed in output; low risk                                  | Medium   |
| Update stale `docs/references/` pages that still describe pre-supersession desired-state semantics                         | Scope discipline (user: report only this session)                 | Low      |

## d) TOTALLY FUCKED UP (radical honesty)

1. **The published v0.10.0 carries a silent-repair-suppression bug and is unreleased-fixed.** Any consumer whose BuildFlow binary embeds v0.10.0 gets detect findings that repair cannot clear — for ANY current or future `Converge`-only normalizer change, not just `lll`. Severity: quality-gate noise + false "converged" confidence; no data loss. Mitigation: local binary now fixed; fleet-wide fix blocked only on release (g1).
2. **AGENTS.md gotcha 6 claimed "detect/apply drift is structurally dead" — false for weeks.** The ADR-017 parity corpus proves byte-identical renders, but nothing proved normalizer-change _accounting_. The claim itself may have anchored earlier sessions away from checking the apply path.
3. **My session process had real detours:**
   - I initially nearly concluded "stale binary + already-fixed, nothing to do" — the decisive experiment (replicating BuildFlow's exact SDK calls in-process) came only after several CLI-side dead ends. Reading BuildFlow's provider source FIRST would have saved ~3 buildflow runs (~6 min) and most of the speculation.
   - The earlier same-night fix (tag guard) left the incident half-closed: guard committed, config healed, but no rerun/closure — the user had to paste the failure to trigger it.
   - The root-cause fix is buried under a daemon auto-commit message (`chore: auto-commit 4 changed file(s)`) — future archaeologists get no signal without this report.
   - Minor: the throwaway probe needed 2 compile iterations (wrong logger type; readable provider source existed before I guessed signatures).
4. **Phantom-finding flip-flop was nearly misdiagnosed twice.** The `lll stale` → `lll missing` direction flip is the signature of binary/config drift, and I ran on config-side hypotheses too long before pinning the binary snapshot version (`flake.lock` rev vs build rev).

## e) WHAT WE SHOULD IMPROVE

1. **Incident closure rule is not self-enforcing.** A fix that makes a failing gate green must end with a rerun of the gate in the same session (buildflow skill step 7 says exactly this — it wasn't applied to the tag-guard fix). Concrete: treat any committed lint/gate fix as incomplete until `buildflow -s <step>` is green.
2. **Consumer-first debugging.** When a BuildFlow step misbehaves, read the provider source (`~/projects/BuildFlow/tools/providers/*`) before touching the consumer repo — the invocation shape (in-process SDK vs subprocess, pinned version, env) determined everything here. Candidate: a short section in the buildflow skill or AGENTS.md gotcha (done for this repo via gotcha 32).
3. **Parity must cover accounting, not just rendering.** Extend the ADR-017 invariant: any short-circuit ("already converged") must be derived from the same change set the report-only path produces. Concrete follow-up: an invariant spec that runs `FixConfig` (dry-run) and `PlanDeviations` on the repo's own config in CI and asserts equality.
4. **Stale-binary detection should be loud, not advisory.** BuildFlow's binary-freshness warning existed and was correct — but it didn't stop the phantom finding from being reported. Candidate BuildFlow-provider improvement: annotate findings with the embedded consumer-repo rev.
5. **Daemon commit messages hide incident history.** Root-cause fixes deserve descriptive messages; the harness (rule 6: no manual commit without ask) currently guarantees bad messages for AI-driven fixes. Policy decision needed (g2).
6. **LSP diagnostics mislead under stale cache** (5 phantom warnings persisted after files were formatted/deleted). Trust `nix fmt` + `golangci-lint run` as ground truth; worth remembering rather than fixing.

## f) TOP 50 NEXT TASKS (ranked by impact; brainstorm — most beyond #15 are ROADMAP fuel for docs-health HARVEST)

| #  | Task                                                                                                                                                                                          | Impact   | Effort | Category           |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------------ |
| 1  | Harvest (f) into TODO_LIST/ROADMAP + CHANGELOG entry for the incident                                                                                                                         | Critical | S      | Documentation      |
| 2  | Tag + release the fix version (merge-union + tag guard); verify pkg.go.dev propagation                                                                                                        | Critical | M      | Bug/Release        |
| 3  | Update BuildFlow's go.mod pin past v0.10.0 once released; rebuild + reinstall                                                                                                                 | Critical | S      | Cleanup            |
| 4  | Codebase sweep for the discard-overwrite class: `result.<Field> = <other>` where `<other>` is a subset computed earlier (`fixer_converge.go`, `internal/cli`)                                 | High     | M      | Bug-hunt           |
| 5  | Add CI invariant: `PlanDeviations` empty ⇔ `FixConfig` dry-run converged on the repo's own `.golangci.yml`                                                                                    | High     | S      | Quality            |
| 6  | File/document the stale-binary phantom-finding mode in the BuildFlow repo (with refresh procedure)                                                                                            | High     | S      | Documentation      |
| 7  | Cross-repo SDK parity contract test in BuildFlow (detector ⇔ repairer on a fixture config)                                                                                                    | High     | M      | Quality            |
| 8  | Guard: any new normalizer added to `Converge` must have an apply-parity spec (test helper or lint note in ADR-017)                                                                            | High     | M      | Quality            |
| 9  | Audit which fleet machines/CI runners run BuildFlow binaries embedding pre-fix snapshots                                                                                                      | High     | M      | Cleanup            |
| 10 | Re-run the FULL buildflow pipeline (all steps) to confirm repo-wide green post-fix                                                                                                            | Medium   | S      | Quality            |
| 11 | Verify `VerifyConfigLoads` tripwire behavior on the converged save (was it exercised?)                                                                                                        | Medium   | S      | Quality            |
| 12 | Spec the never-enable × redundancy interplay (lll in sidecar never-enable + config carries lll)                                                                                               | Medium   | S      | Quality            |
| 13 | Spec `--pragmatic` × supersession interaction (`tierAdmits` × `dropSuperseded`)                                                                                                               | Medium   | S      | Quality            |
| 14 | Review `Change.Key()` collision surface: Key excludes Kind — assert plan validator catches two kinds on one key                                                                               | Medium   | M      | Bug-hunt           |
| 15 | Sweep `pkg/client` test coverage for `FixOptions.Priority` paths                                                                                                                              | Medium   | S      | Quality            |
| 16 | Confirm installed BuildFlow binary rebuilt from clean HEAD (7dcb161 dirty snapshot vs 8971aec) once BuildFlow tree settles                                                                    | Medium   | S      | Cleanup            |
| 17 | Clear legacy `/home/lars/.cache/buildflow/buildflow.db` after migration completes (advisory warning)                                                                                          | Low      | S      | Cleanup            |
| 18 | Decide finding-severity semantics: is "missing from enable" info correct when the linter was removed-by-supersession? (Currently would re-flag configs that pre-date a supersession decision) | Medium   | S      | Decision           |
| 19 | Confirm with owner: priority bounds ADDS only — removals must never be priority-gated (assert in specs)                                                                                       | Medium   | S      | Decision           |
| 20 | Add property tests: `mergeNormalizerChanges` idempotent + order-stable + duplicate-collapsing                                                                                                 | Medium   | M      | Quality            |
| 21 | Canonical ordering for merged normalizer changes (mirror `sortChanges`)                                                                                                                       | Low      | S      | Cleanup            |
| 22 | Update `docs/references/integrations.md`: SDK consumer contract for normalizer-change accounting                                                                                              | Medium   | S      | Documentation      |
| 23 | Update `docs/references/working-with-codebase.md`/testing refs for the new spec files                                                                                                         | Low      | S      | Documentation      |
| 24 | Add README troubleshooting entry: "finding replays after fix → refresh BuildFlow binary"                                                                                                      | Low      | S      | Documentation      |
| 25 | Extend gotcha 32 with the exact `git log <rev>` verification command for the embedded snapshot                                                                                                | Low      | S      | Documentation      |
| 26 | CI canary: run `analyze --format finding` on repo config expecting zero findings (drift tripwire)                                                                                             | Medium   | M      | Quality            |
| 27 | Evaluate `Fixer.ParityCheck(ctx, path)` exported helper (Evaluate vs FixConfig-dry diff) for SDK consumers                                                                                    | Low      | M      | Feature            |
| 28 | Surface normalizer-vs-plan split in configure's user-facing summary table (already in dry-run message)                                                                                        | Low      | S      | Feature            |
| 29 | `audit` subcommand: filter by action type (`--action removed-from-enable`)                                                                                                                    | Low      | S      | Feature            |
| 30 | Property-test `deviationTags`/`tagSafe` over random elements (rename notation fuzz)                                                                                                           | Low      | S      | Quality            |
| 31 | Draft v0.10.1 release notes from this report (go-release skill flow)                                                                                                                          | Medium   | S      | Documentation      |
| 32 | Check whether any OTHER published version family (v0.9.x) carries the overwrite bug; retract if needed                                                                                        | Medium   | S      | Release            |
| 33 | Verify golden HTML report handles the new finding shapes (lll removal) — suite passed; add explicit note to golden-refresh checklist                                                          | Low      | S      | Quality            |
| 34 | `docs/DOMAIN_LANGUAGE.md`: add "supersession"/"redundancy pass" terms                                                                                                                         | Low      | S      | Documentation      |
| 35 | Sweep docs for stale `just` references (gotcha 1)                                                                                                                                             | Low      | S      | Cleanup            |
| 36 | Confirm markdownlint CI passes on AGENTS.md/report changes (separate workflow)                                                                                                                | Low      | S      | Quality            |
| 37 | Investigate BuildFlow-side: findings should carry the embedded consumer-repo rev for stale-binary forensics                                                                                   | Low      | M      | Feature (upstream) |
| 38 | Consider dry-run-only double-run in `convergenceOutcome` as a parity assertion (debug mode)                                                                                                   | Low      | M      | Quality            |
| 39 | Review `convergenceOutcome` audit behavior: should no-op runs be ledgered for observability?                                                                                                  | Low      | M      | Decision           |
| 40 | Grep for remaining TODOs older than 1 week in touched files (house rule)                                                                                                                      | Low      | S      | Cleanup            |
| 41 | Machine hygiene: `~/projects/BuildFlow-pr44`, `BuildFlow.vendor.bak` — delete or archive                                                                                                      | Low      | S      | Cleanup            |
| 42 | Add `docs/status/README.md` index note if the convention isn't documented                                                                                                                     | Low      | S      | Documentation      |
| 43 | Double-check `omitzero` fields on any new report-facing structs (none added; confirm)                                                                                                         | Low      | S      | Quality            |
| 44 | Post-release: watch `buildflow timings` for convergence duration regressions                                                                                                                  | Low      | S      | Quality            |
| 45 | Evaluate promoting the "replicate the consumer's exact SDK calls" probe into a reusable script                                                                                                | Medium   | M      | Quality            |
| 46 | `erraudit` pass over the new `mergeNormalizerChanges` + specs (quarterly review due ~2027-01 anyway)                                                                                          | Low      | S      | Quality            |
| 47 | Confirm the audit ledger recorded BOTH fix paths in the incident (exclusion rename + lll removal) via `audit --json`                                                                          | Low      | S      | Verification       |
| 48 | Add the SDK-probe snippet (sanitized) to `docs/references/` as a debugging recipe                                                                                                             | Low      | S      | Documentation      |
| 49 | Re-verify gotcha 31's standing-warning list still accurate post-incident (binary-freshness now green)                                                                                         | Low      | S      | Documentation      |
| 50 | Ask the 3 questions in (g) and unblock items 2/3/5 accordingly                                                                                                                                | High     | S      | Decision           |

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **Release policy:** The overwrite bug is live in published v0.10.0 and every BuildFlow binary built against it. Do you want me to run the go-release flow now (tag v0.10.1 with the merge-union + tag-guard fixes, then bump BuildFlow's go.mod pin + rebuild)? I tried to self-answer via `git tag`/go.mod history and found no policy doc beyond the release-process reference.
2. **Commit-message policy:** The root-cause fix landed under a daemon `chore: auto-commit 4 changed file(s)` message. Should root-cause fixes get descriptive manual commits going forward (and should I amend THIS one), or is daemon-history-as-is acceptable? Rule 6 prevents me from deciding.
3. **Fleet scope:** Is the local machine the only BuildFlow consumer whose gate matters for this repo (my refresh fixed it here), or should I audit/refresh other machines' and CI runners' BuildFlow installs against the stale-snapshot class? I have no inventory of BuildFlow installs beyond this machine.

---

## Session timeline (for correlation)

- ~01:24 audit: `lll removed-from-enable` **did not exist yet**; earlier sessions: tag-guard fix (3ab5c92), supersession filter (0946975), exhaustruct config heal (d099f44, 02:58 push point e0fcafe)
- 02:59 BuildFlow rerun (this session): crash gone; `lll stale` warning appears
- 03:05 `--fix` detect→repair→verify: repair no-ops, warning persists → divergence hunt
- 03:0x-03:2x SDK-boundary probe reproduces; root cause `finalizeConvergence` overwrite; fix + specs; config converged; lint zeroed
- 03:37 refreshed BuildFlow binary; user's command: **✓ passed 1/1, 0 findings**
- 04:16 this report

_Report format: `.md` per explicit user instruction (skill default is HTML — flagged per spec). Section (f) is HARVEST-ready but not harvested, per "then wait for instructions"._
