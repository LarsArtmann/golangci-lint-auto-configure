# ADR-017: Declarative Desired-State Reconciler as the Default Configure Engine

**Status:** Accepted (default engine since the corpus flip; legacy imperative path scheduled for deletion)\
**Date:** 2026-10-14

## Context

The configure engine was an imperative mutator pipeline: a dozen ordered
passes each inspected the config, mutated it, and incremented hand-maintained
counters. Counting, audit, and dry-run reporting were three independently
maintained projections of the same work, and they drifted (the split-brain
that motivated the exclusion-rule `RuleKey` dedup fix). Detect mode and apply
mode shared nothing structural, so "check said optimal but apply changed
things" was possible in principle.

## Decision

Rebuild configuration as a **desired-state reconciler** with five pure
stages:

1. **Collect** — `Env` names every environment fact (tool version, build Go,
   local Go, project Go, tech flags, analyzer-enabled set).
2. **Normalize** — pure normalizers repair unloadable shapes (version,
   duration, typecheck, tool-disabled moves, bad settings keys, deprecated
   supersession, exclusion-rule renames). Changes are the count, the audit
   trail, and the dry-run report at once.
3. **Render** — `Desired(env, options...)` projects the 137-record policy
   table into the desired configuration. No I/O, fully deterministic.
4. **Plan** — `PlanDesired` diffs observed against desired through a
   path-keyed **merge-strategy table** (`ensure-present`, `union`,
   `prefer-user`, `prefer-desired`, `cap-go-version`). `--force-settings` is
   a table flip, not a flag threaded through mutators.
5. **Apply** — `ApplyPlan` materializes the plan onto a clone. Audit = diff:
   `AuditActionsFor` projects every applied `Change` into exactly one ledger
   entry.

**Findings = diff.** The `DeviationDetector` runs the same plan phase in
report-only mode (`Fixer.PlanDeviations`) and converts each `Change` into one
finding. Lint mode and apply mode consume the identical change set, so
detect/repair drift is structurally impossible. The old
`WithPriority` alignment knob and its detect→repair-loop prevention are
deleted with the recommendation-based detector.

**Tiers derive from measurement.** The pragmatic-noise tier is not a
hand-maintained map: `types.DeriveTier` promotes an unmanaged linter whose
measured friction (`Evidence.Ratio`, nolints per config) reaches
`PragmaticNoiseFrictionThreshold` (0.8). `--pragmatic` is the user-facing
threshold, not a list membership check.

## Evidence

- **Parity corpus**: 11 fixtures captured from the legacy engine
  (`pkg/linter/testdata/parity-corpus/`). `TestParityCorpusConverge` proves
  the reconciler reproduces the legacy engine's saved-config bytes
  byte-identically on every fixture, including the tool's own 500+ line
  config.
- **Pin matrix**: `scripts/e2e-pin-matrix.sh v2.14.0 v2.12.0` green on the
  new engine (below-min refusal, at-min load, `--check` idempotence).
- **Self-host**: this repository's own `.golangci.yml` converges to itself
  (`configure --check` exit 0) with the sidecar
  `.golangci-lint-auto-configure.yml` documenting the intentional
  divergences (depguard disable + never-enable).

## Consequences

- Adding a linter = editing one policy record (interval, rationale,
  settings, friction evidence). Views over the table keep the ~40 legacy
  consumers compiling until the cleanup phase deletes them.
- Normalizer changes are auditable by construction; the hand-maintained
  `configChangeRecorder` counters are dead on the default path.
- The legacy imperative pipeline remains as an explicit opt-out
  (`SetEngine(EngineLegacy)`) during the migration window; its deletion is
  tracked in TODO_LIST.md together with the compat-view removal.

## References

- Plan: `docs/planning/2026-10-13_17-45_SUPERB-declarative-reconciler.md`
- Parity gate: `pkg/linter/parity_corpus_test.go`
- Strategy table: `pkg/linter/merge.go` (`DefaultMergeStrategies`)
