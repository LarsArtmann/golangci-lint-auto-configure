# Session Self-Review: Invocation Tracking / CQRS-ES / OTEL Audit

> **Scope:** Brutal self-review of the 2026-11-02 audit session only ("Do we
> track all invocations?", "How are we doing on Command- & Event-Sourcing?",
> "How are we doing on OTEL?"), plus everything noticed along the way. No new
> research beyond the session's own observations. Point-in-time snapshot —
> stale by design.

**Session summary:** Three-question audit answered with code + empirical
evidence (isolated `XDG_CACHE_HOME` ledgers). One real bug found and fixed
(`--check --diff` phantom audit entries), one spec added, four docs updated,
one TODO row harvested. Test suites for the touched packages green.

---

## a) FULLY DONE

1. **Q1 audit — invocation tracking, answered with an evidence matrix.**
   Verified empirically, not just from code: plain `configure` apply writes
   exactly one ledger entry per Change (98 changes → 98 entries); no-op
   configure, `analyze`, `report`, `--check`, `--dry-run` write zero;
   `migrate` rewrites the whole config with zero entries
   (`pkg/migration/migrator.go:171` has no ledger); `configure --preset`
   same (`cmd_configure_preset.go:127`); SDK/embedded path
   (`pkg/client/client.go:50` — fixer built without `SetLedger`, and
   `Options` has no recorder field) is invisible, including BuildFlow-driven
   repairs. Conclusion: **mutation coverage on 1 of 4 mutating paths; no
   invocation (run-lifecycle) events at all.**
2. **Q2 audit — CQRS/Event-Sourcing assessment.** ADR-017 reconciler: `Change`
   records as immutable identified facts (pkg/linter/merge.go:29), CQS split
   via the narrow `DeviationPlanner` port (pkg/finding/detector.go:16),
   findings/counts/audit all projections of one Change set, ledger as
   append-only notification log with run streams (`run_id`, `repo_hash`).
   Verdict: principles strongly present, machinery (replay, aggregates, event
   store) deliberately absent — correct for a CLI where git is the real event
   store for the config aggregate.
3. **Q3 audit — OTEL.** Zero `go.opentelemetry.io` in go.mod/go.sum;
   SECURITY.md:29 documents local-only/no-telemetry as posture; the tool
   polices OTel in consumers (recommends `spancheck`) while having none
   itself — appropriate. Archived 2026-07-08 report's "~~Add OTel trace
   spans~~ done" strike-through is misleading (resolved via errorfamily, not
   OTel).
4. **Bug fixed: `configure --check --diff` phantom audit entries.** Check+diff
   temporarily saves the converged config, the fixer audited during that
   save, then the config was restored — every drifted CI run appended a full
   set of false `added-to-enable` entries, poisoning
   `audit.PreviouslyAutoEnabled` cycle detection and violating the documented
   "dry runs never emit entries" invariant (AGENTS gotcha 17). Fix:
   `prepareFixerRun` passes `noAudit: flags.NoAudit || flags.Check`
   (internal/cli/cmd_configure_fixer.go:75). Verified before (96 phantom
   entries on one drifted run) and after (0 entries, config restored, real
   apply still writes 96). Regression spec with a self-validating control
   run (internal/cli/cmd_check_test.go:89).
5. **`lll` silent-drop TODO row independently confirmed resolved.** Fresh
   repro (repo's own config + `lll` re-added to `linters.enable`, sandbox
   git repo): removal now emits a Change AND a `removed-from-enable` ledger
   entry with the golines-redundancy reason, and `--check` reports the same
   single change — check/run parity holds. (The 2026-10-17
   `mergeNormalizerChanges` union fix closed it; a parallel docs-health
   session struck the row the same day.)
6. **Docs updated:** AGENTS gotcha 17 (check-mode suppression + coverage
   caveat), CHANGELOG entry added (plus consolidated a duplicated
   `### Fixed` heading), ROADMAP stale sub-item corrected
   (`exclusion-rule-merged` audit entries shipped and spec-pinned by
   `exclusion_merge_ledger_internal_test.go`; open parts kept),
   TODO_LIST row added for the coverage gaps.
7. **Verification discipline on the fix:** `internal/cli` + `pkg/linter`
   suites green with `-race`; treefmt/gofmt clean; repo-lint diff vs stashed
   master baseline = zero new findings; full build OK.

## b) PARTIALLY DONE

1. **Test breadth.** Only the two affected packages were run. The preferred
   full gate (`ginkgo -r -p -race` over pkg/internal/cmd, or
   `nix flake check`) was NOT run. Justified-ish for a two-file Go change,
   but not the house gold standard.
2. **Q2 report-shape claim.** "Reports are projections of Changes" was
   verified for the findings path (`DeviationsToFindings`) only; the JSON
   report's exposure of the raw planned Change set was never inspected.
3. **Q3 rests on documentation + dependency graph.** SECURITY.md's "no
   third-party network calls" was not verified empirically (no egress
   test); charm.land dependency telemetry surfaces were not audited.
4. **Harvest routing.** The main gap got its TODO_LIST row during the
   session; section (f) below routes the remainder in-report only — the
   concurrent docs-health session owns TODO_LIST/FEATURES edits right now,
   so further file edits were deliberately deferred.
5. **Exit-code nuance noticed, not chased:** `configure --check` (changes
   needed) exited 75 in one test, while the isolated `--check --diff` repro
   exited 1. Same semantic condition, different exit codes across check
   modes — observed, unexplained, un-pinned.

## c) NOT STARTED

1. **Closing the actual coverage gaps** (migrate/preset/SDK audit entries,
   recorder injection API, run-lifecycle events) — designed and documented,
   not implemented. Deliberate: each needs a small API/vocabulary decision
   and touches the privacy surface (USER-GATED).
2. **FEATURES.md entry** for the audit-ledger behavior change — skipped to
   avoid colliding with the concurrent session's uncommitted FEATURES.md
   edits.
3. **Re-checking ci.yml's dogfood step** (`configure --check` with NO_AUDIT
   set) against the new suppression — relied on the TODO_LIST note, never
   opened the workflow file. Should be a no-op, but "should" is not
   "verified".
4. **Correcting the archived 2026-07-08 report's misleading OTel
   strike-through** — identified, no annotation written (archive is
   docs-health ANNOTATE territory).

## d) TOTALLY FUCKED UP

1. **`lll` fixture engineering — three attempts.** First inserted `lll` into
   `formatters.enable` (wrong list; the run then "failed" fmt with "lll is
   not a formatter" while configure itself said "optimal", which briefly
   looked like ANOTHER silent-mutation bug and cost investigation time).
   Second attempt broke YAML indentation (parse rejection). Third worked.
   Lesson: read the fixture structure (or the spec's builder) before
   hand-crafting YAML.
2. **Flag guesswork.** Guessed `--show-diff` before reading the actual
   registration (`--diff`, internal/cli/commands.go:245) — one wasted
   compound repro and a wrong assumption about which command had run.
3. **Hand-crafted exclusion-merge fixture didn't trigger the merge** (0
   entries) because its RuleKey didn't match `DefaultExclusionRules[0]`;
   conclusion was rescued by reading the pinned spec
   (`exclusion_merge_ledger_internal_test.go`) instead of building the
   fixture the spec builds. Right conclusion, sloppy empiricism.
4. **Tool misuse noise:** `rg --include` (not an rg flag), `rg -rn`
   (hook-blocked replace trap), a grep-based exit-code capture that
   reported the pipe's exit, not the tool's. Small, repeated,
   avoidable.
5. **Chased stale LSP wsl_v5 warnings** briefly before confirming against
   the repo's own linter run that my files added zero findings vs the
   master baseline (the diagnostics were stale; the 2 wsl_v5 + 61 wrapcheck
   findings pre-exist on master under local v2.15).

## e) WHAT WE SHOULD IMPROVE

1. **Fixture-first empiricism:** when a behavior has a pinned spec, copy the
   spec's fixture construction before inventing YAML by hand (lll and
   exclusion-merge both burned cycles here).
2. **Close with the full gate:** affected-packages-green is necessary, not
   sufficient; budget the ~50s parallel ginkgo run (or `nix flake check`)
   before declaring done.
3. **Exit-code contract deserves pinning:** BSD sysexits mapping is a
   user-facing contract; the 75-vs-1 observation suggests check-mode paths
   diverge. One table-driven spec per command path would have caught this
   class.
4. **Pre-existing lint drift under v2.15** (61 wrapcheck + 2 wsl_v5 on
   master locally, CI green on pinned v2.14): version skew or config drift —
   needs a triage decision, not silence.
5. **Keep the new AGENTS gotcha-17 coverage caveat alive:** it now documents
   which paths are unaudited; any new mutating path must update it.

## f) Things to get done next (routed from this session)

Already tracked (do not duplicate):

1. **[tracked: TODO_LIST 2026-11-02 row]** Close audit-ledger coverage gaps:
   migrate/preset/SDK mutations unaudited, no run-lifecycle events.
2. **[tracked: ROADMAP exclusion-merge bullet]** `--reset-exclusions`,
   `--validate-merge`, versioned default exclusion rules (remaining open
   parts).

New / unrouted:

3. Recorder injection API: `client.Options.Recorder audit.Recorder` so
   embedded consumers (BuildFlow) can opt into auditing; document the
   privacy-preserving NoopRecorder default.
4. Decide the ledger's action vocabulary for the missing paths
   (`preset-applied`, `migrated-v1-to-v2`) OR explicitly declare the ledger
   fixer-path-only in README/`--help` so the invariant is honest by
   documentation.
5. Run-lifecycle events decision: a per-run `run-completed` entry (mode,
   change count, tool version) would make "invocation tracking" real;
   weigh ledger noise against 90-day retention and the no-telemetry
   posture.
6. Chase and pin the 75-vs-1 exit code difference for "changes needed"
   across check modes; add exit-code specs per command path.
7. FEATURES.md: audit-ledger row update (check-mode suppression) —
   coordinate with the concurrent docs-health session.
8. Verify ci.yml's dogfood step still behaves as intended under check-mode
   ledger suppression (expected no-op; confirm by reading the workflow).
9. Lint version skew triage: run the CI-pinned golangci-lint (v2.14)
   locally on master; if green, the 61 wrapcheck findings are v2.15
   behavior drift — decide bump-or-curate.
10. Empirically verify SECURITY.md's "no third-party network calls" (sandbox
    egress test) and audit charm.land deps for any telemetry hooks.
11. Inspect whether JSON/HTML reports expose the planned Change set (not
    just findings); if not, consider exposing the plan for machine
    consumers.
12. docs-health ANNOTATE the archived 2026-07-08 report: item 48's
    "~~OTel trace spans~~ done" strike-through is false as written.
13. Give the recurring "add OTEL" idea a durable home: one ROADMAP line
    recording the decision (local-only posture) so 2026-03-era TODO
    resurrections stop.
14. One spec pinning that `installhook` never mutates configs (completes the
    mutation-path map from this audit).
15. pkg/client doc comment noting: sidecar enforcement travels with the
    fixer (works embedded), but ledger-based cycle detection silently
    degrades to off without a recorder — consumers should know.

## g) Questions I can NOT figure out myself

1. **Ledger scope (USER-GATED design):** should the invariant become "every
   mutation by this tool is audited" (extend to migrate/preset/SDK), or is
   fixer-path-only the intended scope? This changes the privacy/noise
   surface of a documented local-only tool — your call, not mine.
2. **Invocation tracking appetite:** do you want per-machine run-lifecycle
   events (run frequency, no-op rate), or does "collects no telemetry"
   extend to deliberately NOT knowing invocation frequency locally either?
3. **Lint-version canonicality:** when local v2.15 and CI-pinned v2.14
   disagree by 61 findings, which is truth — bump CI to v2.15 and fix the
   findings, or hold v2.14 and treat local v2.15 output as noise until the
   pin moves?

---

**Files this session actually changed:**
`internal/cli/cmd_configure_fixer.go` (fix),
`internal/cli/cmd_check_test.go` (spec),
`AGENTS.md` (gotcha 17), `CHANGELOG.md` (entry + heading fix),
`ROADMAP.md` (stale sub-item), `TODO_LIST.md` (coverage-gap row).
All picked up by the auto-commit daemon.
