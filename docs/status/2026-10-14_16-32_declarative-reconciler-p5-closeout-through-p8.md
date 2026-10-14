# Status: Declarative Reconciler Execution — Session 2 (P5 close-out through P8)

**Date:** 2026-10-14 16:32
**Branch:** master (clean tree; auto-commit daemon has been committing throughout)
**Plan:** `docs/planning/2026-10-13_17-45_SUPERB-declarative-reconciler.md` (27 tasks / 127 micro-tasks)
**Predecessor report:** `docs/status/2026-10-14_14-59_declarative-reconciler-execution-status.md`

---

## a) What got done since the last report

### P5 closed out (converge engine byte-green + legacy contract green)

- **User-disable filter**: `WithUserDisabled` option + `shouldRender` skip; `disableMinusEnable`
  computes it in `Converge` (enable/disable conflicts are NOT user disables).
- **Enable/disable conflict resolution**: `planEnableDisableConflicts` removes every disable
  entry whose linter ends up enabled (legacy enable-wins contract), deduped against the
  anti-gaming overlay's own removals.
- **Orphaned-settings pruning**: `planOrphanSettings` plans `pruned-settings` removals for
  settings blocks outside the final enable set (legacy prune passes, unified into one plan).
- **Audit = diff, completed vocabulary**: `AuditActionsFor` is table-driven
  (`auditActionMap`); mappings now cover `moved-to-disable`, `removed-from-enable`,
  `pruned-settings`, `re-enabled` (overlay adds with "unjustified" reason prefix),
  `suppressed-re-enable` (never-enable blocks, audit-only event via the new
  `ConvergeResult.Suppressed`), `exclusion-linter-renamed` (ledger names the predecessor),
  `exclusion-linter-removed`, and `exclusion-rule-merged` (reason names the added linters).
  `ledgerActionFor` extended to 12 mappings. `ledgerActionMap` replaces the 27-statement
  switch.
- **Exclusion-rule sweep rewritten** (`RenameDeprecatedInExclusionRules` +
  `sweepExclusionRuleLinters` + `keepSweptRule` + `sweepRename` + `sweptDropChange` +
  `keepSweptEntry`): legacy-exact semantics — rename with seen-dedupe, drop without
  successor, text/source-scoped rules die with their last linter, path-scoped rules
  survive; the iterate-while-mutating bug is gone (collect-then-filter).
- **All 12 previously failing specs green**; full suite green; `golangci-lint run` driven
  from 69 findings to **0 issues** (path constants `Path*` in merge.go, `planPathRank`
  derived from an ordered slice, funlen extractions in merge/apply/normalize/converge/
  fixer_converge, gocyclo table-driven audit mapping, prealloc/varnamelen/wsl/exhaustive/
  gocritic/godoclint/ineffassign/staticcheck fixes, deprecated `NormalizeGoMajorMinor`
  call site migrated to `types.ParseGoVersion`).
- **Pin matrix re-run**: `scripts/e2e-pin-matrix.sh v2.14.0 v2.12.0` → ALL CHECKS PASSED
  (rescue, config verify at min, `--check` idempotent).

### P6 — Findings = diff (T22) + Evaluation and reports (T23)

- **`Deviation` adapter**: `pkg/finding/deviation.go` — `DeviationsToFindings` converts each
  reconciler `Change` into one finding (rule per config path, severity by kind, before/after
  codes, tag-safe linter tags, category via `LinterNameToCategory`).
- **Detector rewritten**: `ConfigAnalysisDetector` + `WithPriority` + both priority-filter
  helpers DELETED; new `DeviationDetector` (name `config-deviation`) runs
  `Fixer.PlanDeviations` (report-only plan) through the adapter. Specs rewritten.
- **`Evaluation` type + `Fixer.Evaluate`** (`pkg/linter/evaluation.go`): Env + Desired +
  Observed + Deviations + Suppressed + Converged in one report input.
  `evaluateConvergence`/`loadForEvaluation`/`prepareEvaluation`/`collectEvaluationEnv`
  helpers keep every function under the lint caps. `PlanDeviations` now delegates to
  `Evaluate`.
- **ConfigAnalysis slim-down**: the five count fields are derived methods
  (`CountByPriority`/`CriticalCount`/…); `calculateRecommendationCounts` deleted; wire shape
  intentionally loses the count keys (specs updated to pin the new shape).
- **Report generators ported**: `report --format json/sarif/finding/html` all run through
  `Fixer.Evaluate` now. HTML template rewritten deviation-centric (summary cards,
  converged banner, deviation tables, enabled/disabled lists) — golden regenerated and the
  diff reviewed (palette-synced to the shared CLI colors). `JSONReport` carries
  `Deviations`/`Suppressed`/`Converged` summary. `AnalysisToReport`/`AnalysisToSARIF` remain
  for the `analyze` command only (wire-compatible by design this release).
- **`EvaluationToFindings/Report/SARIF`** in pkg/finding, including a
  `suppressed-re-enable` finding rule for never-enable blocks.
- Error-code registry updated (new `detector.plan`, `evaluation.*`, `report.evaluate`;
  dead `detector.analyze`, `detector.convert_validation`, `report.analyze*` removed).

### P7 — Derivations and relations

- **Friction-derived tiers (T24)**: `types.DeriveTier` + `types.PragmaticNoiseFrictionThreshold`
  (0.8 nolints/config). The table now loads as `linterPolicyRecords` → `LinterPolicies =
  deriveLinterTiers(...)` (dependency-ordered var init). The four hand-maintained
  pragmatic-noise `Tier:` stanzas are DELETED from the table, the generator's
  pragmaticNoise branch and the `pragmatic_noise_linters.json` golden are deleted, and the
  views golden was regenerated from the derived tier. `--pragmatic` is the threshold's user
  handle. Pinned by `pkg/constants/tier_derivation_test.go` (exact derived set, threshold
  boundary 0.79/0.83, explicit tiers win, measured rationale present).
- **Redundancy as typed relations (T25)**: `constants.RedundancyRelations`
  (`RedundancyLinter`: lll→golines; `RedundancyFormatter`: gofmt→gofumpt) with
  `RedundanciesFor` and `ValidateProblem`. Init-time referential integrity panics on
  dangling targets. `RemoveRedundancies` now iterates the table
  (`removeRedundantLinters`/`removeRedundantFormatters`/`enabledFormatterSet`) — the
  hardcoded engine branches are gone.

### P8 — Self-host (T26) + docs (T27 partial)

- **Self-hosted**: `.golangci-lint-auto-configure.yml` documents the repo's intentional
  divergence (depguard: justified disable + never-enable). `configure --config
  .golangci.yml --check` exits 0 with the sidecar loaded (1 justified disable, 1
  never-enable) and leaves the config untouched — the self-config trap is dead.
- **ADRs**: `docs/adr/ADR-017-Declarative-Reconciler-Engine.md` (accepted, evidence-linked)
  and `docs/adr/ADR-018-Two-Version-Axes-Branded-Types.md`; both registered in
  `docs/ARCHITECTURE.md`.
- **AGENTS.md gotchas rewritten** for the new reality: #6 (reconciler is the default engine,
  counting is the diff), #10 (one policy record; maps are views; noise tier derived),
  #19 (`--pragmatic` by threshold, pinned set), #28 (min version is derived).
- **CHANGELOG.md** Unreleased: Added (reconciler, audit vocabulary, findings=diff,
  user-disable respect, derived tiers, relation table, self-host) + Changed
  (evaluation-based reports, derived counts) entries.
- **FEATURES.md**: 8 new/updated rows (reconciler, policy table, derived tiers, merge
  strategies, deviation reports, self-host, version axes, priority via table).
- **TODO_LIST.md**: two High-priority rows added with precise execution scope (legacy
  engine deletion; compat-view deletion) — see c).

---

## b) Current verified state (all green at 16:29-16:32)

| Gate | Result |
| --- | --- |
| `go build ./...` | OK |
| Full suite `ginkgo -r -p -race --skip-package=cmd/` (18 suites) | PASS (1m10s) |
| Parity corpus (`TestParityCorpus*`, incl. converge flip gate) | ok |
| `golangci-lint run --config=.golangci.yml` | **0 issues** |
| `nix fmt` (treefmt: gofumpt/golines/nixfmt) | clean |
| Self-host `configure --config .golangci.yml --check` | exit 0, zero diff |
| Pin matrix v2.14.0 (min v2.12.0) | ALL CHECKS PASSED |
| `git status` | clean (auto-committed) |

Deletions verified by grep: `WithPriority`, `ConfigAnalysisDetector`, the priority filters,
the count fields, the hand-maintained noise tiers, and the pragmaticNoise golden are all
gone.

---

## c) NOT done (the honest list)

1. **Legacy imperative engine deletion** (the big one, deliberately deferred as ONE future
   commit): `SetEngine`/`Engine` field, `applyLintersFix`/`applyAllFixes`/`applyAndSave`,
   `runPreFlightChecks`, `checkDryRunEarlyReturns` + dry-run calculators, `fixCounts`,
   `configChangeRecorder`, `fixer_audit.go` snapshot diffing (`recordConfigChanges`,
   `recordExclusionLintersCleanup`), `fixer_enforce.go` (legacy enforcement; the converge
   overlay replaced it), the legacy parts of `fixer_deprecated.go` (`preFixDeprecatedLinters`,
   `cleanupExclusionLinters` — superseded by `RenameDeprecatedInExclusionRules`), the
   cycle-detection reader (`Fixer.reader`, `ledgerReader`, `audit.PreviouslyAutoEnabled`
   consumption), legacy-only tests (`fixer_enforce_test.go` legacy suites,
   `fixer_config_internal_test.go` legacy specs, `TestParityCorpus` +
   `TestParityCorpusIdempotent` — keep `TestParityCorpusConverge`), and the CI `parity` job
   repoint. Est. ~1.7k lines plus test debris. Everything it needs to be safe (corpus,
   pin matrix, self-host) is already green on the new engine.
2. **Compat-view deletion** (T27.1 remainder): the legacy maps in
   `linter_policy_views.go`/`rules.go`/`config.go` (priorities, reasons, deprecated,
   disabled, never-auto-enable, noise, min-versions, project-specific, settings maps) and
   their views-parity goldens — after porting the ~40 remaining consumers to
   `PolicyRecord`/table queries.
3. **`analyze` command still on `AnalyzeConfig`/`ConfigAnalysis`** (text/json wire kept
   intentionally stable this release; SARIF/finding ports exist but are wired only into
   `report`). Decide and execute in the next release window.
4. **Pin matrix other cells** (only the v2.14.0 cell was run this session).
5. **AGENTS.md re-audit of remaining gotchas** that still describe the two-engine world
   (#21 golden workflow fine; #26/#30 legacy mentions; #28 generate-settings flag-default
   sync not re-verified against the derived minimum).
6. **`scripts/validate_linter_data.go`**: still referenced by docs; needs a pass to confirm
   it reads only table/views (was out of scope today).
7. **Evaluation golden corpus**: only the HTML golden exists; no golden fixtures for the
   JSON/SARIF evaluation outputs.
8. **`ExclusionRuleConfig` dedupe hardening** (dedupe of successor listed twice in one rule
   is spec-pinned, but no fuzz coverage on the sweep).
9. **Roadmap/README user-facing docs** still describe the legacy engine as default in
   places (README quickstart output examples).
10. **`DeriveTier` AsOf sourcing** (T24.1): Evidence rows for funlen/gochecknoglobals/
    ireturn/wrapcheck carry zero timestamps; only exhaustruct_v5 has a real measurement.

## d) Totally fucked up (so it does not happen again)

1. **Var-init-order bug in T24**: put `deriveLinterTiers()` in `init()` — but the legacy
   view vars (`PragmaticNoiseLinters = tierView(...)`) initialize BEFORE `init()` runs, so
   the derived set was empty and the FEATURES docs-integrity spec caught it. Fixed by
   restructuring to `linterPolicyRecords` → derived `LinterPolicies` (dependency-ordered).
   Lesson: in this package, derived state must be built through var-initialization
   dependencies, never `init()`.
2. **Blind python replaces after formatting**: at least four replace attempts silently
   no-op'd (old_string drift after gofmt/golines reformatted the block), including one that
   burned three cycles on `evaluateConvergence` funlen. Lesson: after any formatting pass,
   re-read the exact block; prefer the edit tool for surgery and python only for verified
   mechanical sweeps.
3. **sed constant sweep rewrote the definitions**: `"linters.enable"` → `PathLintersEnable`
   also rewrote the const block itself into `PathLintersEnable = PathLintersEnable`
   (initialization cycle). Caught by build immediately.
4. **Typo in a multiedit** (`renameBadSettingsKeys(...)utors...`) — build caught it.
5. **`convergePhaseOne` error-swallowing signature**: first version returned a
   `*types.MigrationResult` as the error carrier, which DROPPED the error entirely
   (`migrationError` returns `(nil, err)`). Caught by reading `migrationError`'s signature;
   reworked to return `error` via `analysisError`.
6. **Golden regen ordering churn**: deleted `pragmatic_noise_linters.json`, then ran the
   suite BEFORE the derivation fix — UPDATE_VIEWS regenerated it as `{}` and had to be
   regenerated again after the var-init fix. Regen goldens only after the behavior fix is
   verified.
7. **exhaustruct literals vs config escape hatch**: completed the
   `configFindingParams` literals partially, then added an `.golangci.yml`
   ignore-pattern for the struct instead of finishing the sweep. Defensible (optional-field
   params struct), but it is a config escape hatch for what was half-mechanical work.

## e) Improvements for next session

1. Run the FULL suite after each phase-level change, not only at phase boundaries —
   cross-package fallout (ui colors test, errors codes test, JSON wire specs) surfaced only
   in batch runs.
2. Use `nix fmt` (treefmt owns gofumpt AND golines) instead of chasing golines findings
   individually with `gofmt -w` (which does not apply golines).
3. When moving derivation/init logic between var init and `init()`, check all
   package-level var initializers that read the derived state FIRST.
4. Grep the exact current bytes of a block before any scripted replace; never replace
   after a formatting pass without re-reading.
5. Keep deletions (legacy engine) as one atomic commit with its test deletions — half-
   deleted legacy states cannot build.

## f) Next steps (prioritized)

1. Delete the legacy engine (scope in c.1) as one atomic commit; delete
   `SetEngine`/`Engine`, repoint the CI parity job to `TestParityCorpusConverge`, drop the
   legacy parity tests.
2. Delete `Fixer.reader`/`ledgerReader` cycle detection and the dry-run early-return
   machinery with it; update `audit` subcommand docs (PreviouslyAutoEnabled remains queryable
   history).
3. Port the ~40 remaining view consumers to the table and delete the compat views +
   views-parity goldens (c.2).
4. Port the `analyze` command to `Evaluation` (or explicitly freeze it as the
   observation surface) — resolve the split input story (c.3).
5. Add JSON/SARIF evaluation golden fixtures (c.7).
6. Run the remaining pin-matrix cells (v2.12.0 binary) + a fleet-style smoke over 3-5
   sibling repos' real configs (c.4, mirrors TODO_LIST "fleet sweep").
7. Fill Evidence AsOf/measurement provenance for the four derived noise linters (c.10).
8. Re-audit AGENTS.md gotchas #26/#28/#30 and README quickstart against the one-engine
   reality (c.5, c.9).
9. Confirm `cmd/generate-settings -min-tool-version` default equals the derived minimum
   (gotcha #28 rule).
10. Re-run `scripts/validate_linter_data.go` (or retire it) once views are gone (c.6).
11. Dogfood: run `configure` (non-check) on the repo once and commit the resulting
    ledger-visible audit trail as a demo of `audit` output.
12-50. Remaining plan micro-tasks all live in the plan doc §6 (T27.3-T27.6 docs polish,
    ROADMAP harvest for "no CUE / no key-level merge surgery" non-goals, docs-health sweep
    to fold this report, CHANGELOG release-date stamping when cutting the next version).

## g) Questions for Lars

1. **Legacy engine deletion timing**: cut a release from today's green state first (dual-
   engine escape hatch preserved for one release), or delete the legacy path immediately as
   the next commit (my recommendation: delete now — the corpus + pin matrix + self-host all
   prove the new engine, and the escape hatch is untested surface)?
2. **`analyze --format json` wire**: keep the `ConfigAnalysis` shape indefinitely (this
   release's choice, zero consumer breakage) or move it to `Evaluation` in the next major
   and accept that downstream CI parsers must update?
3. **Suppressed re-enables**: currently they land only in the audit ledger (plus the new
   finding rule in `report`). Should `configure --check` also surface them in its human
   output, or is ledger+report enough?
