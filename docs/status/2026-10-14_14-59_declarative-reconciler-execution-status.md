# Status Report — Declarative Reconciler Execution (SUPERB Plan)

**Date:** 2026-10-14 14:59
**Scope:** Execution of `docs/planning/2026-10-13_17-45_SUPERB-declarative-reconciler.md`
(27 tasks / 127 micro-tasks, 9 phases). This report covers what was executed
today, what is verified, what is half-wired, and what comes next.

---

## Executive Summary

The reconciler is **alive and byte-proven**: the declarative engine
(collect → normalize → render → plan → apply) reproduces the legacy engine's
output **byte-identically across all 11 corpus fixtures**, and is now the
**default engine** (the imperative pipeline is the opt-out). Phases P0–P4 are
complete and verified; P5 is functionally complete with a 12-spec tail of
legacy-parity fallout in progress. P6–P8 are not started.

**The headline deliverable** — `TestParityCorpusConverge` — is the flip gate
the plan demanded: the new engine cannot ship config drift without a red CI
job.

---

## a) Fully done (implemented AND verified)

| Phase             | Deliverable                                                                                                                                                                                                                                                                                                                                                                        | Verification                                                                        |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| P0/T01–T02        | Parity corpus: 11 fixtures (bare-minimal, typical, deprecated-linters, bad-settings-keys, overspecified-run-go, redundant-formatters, enforced-sidecar, never-enable-sidecar, pragmatic, tool-own, schema-fixture) with manifests + golden config/audit/result snapshots; `UPDATE_PARITY=1` regen; idempotency spec; CI `parity` job (pinned golangci-lint v2.14.0)                | `TestParityCorpus`, `TestParityCorpusIdempotent` green; CI job added to `ci.yml`    |
| P1/T03–T04        | Version kernel: `GolangciVersion` + `GoVersion` branded types (Parse-only, zero = unbounded), generic `Interval[V Ordered[V]]`, cross-axis misuse = compile error, deprecation shims, fuzz targets                                                                                                                                                                                 | `pkg/types` specs green; fuzz seeds pass                                            |
| P1/T05            | Call sites migrated behind the types (version checker, categorizer, run.go cap, GraduatedIn)                                                                                                                                                                                                                                                                                       | Full suite green; parity corpus green                                               |
| P2/T06–T07        | `LinterPolicy`/`FormatterPolicy` record types (closed Tier union, Gates, VersionedSettings, Evidence, supersession); **137-record policy table** + 6-record formatter table; init()-time structural validation (dup names, referential integrity, priority/rationale coupling)                                                                                                     | `policy_table_test.go`; init panics proven by the gofmt-missing-order bug it caught |
| P2/T08            | Compat views generating every legacy map from the table; **13 byte-golden views-parity tests** (`UPDATE_VIEWS=1`)                                                                                                                                                                                                                                                                  | `TestLegacyViewsGolden` green — views reproduce the old maps exactly                |
| P2/T09            | Integrity specs ported to the table; structurally-impossible map tests deleted (tier overlap, semver validity, disabled-has-no-priority — now compile-time/init-time facts); `scripts/validate_linter_data.go` rewritten for the table era                                                                                                                                         | green; validator runs clean                                                         |
| P2/T10            | Presets as **queries** over the table (`reference` = the Critical+High query — drift now structurally impossible; declared presets init-validated); descriptions derive counts                                                                                                                                                                                                     | `preset_queries_test.go` green: derived reference == legacy list exactly (61)       |
| P3/T11–T13        | `Env` collector (tool/build-Go/local-Go/project-Go/tech/generated/analyzer-enabled, fail-open semantics); pure `Desired()` renderer (tier/gate/validity/lang filtered, settings, run/issues defaults); **`MinGolangCILintVersion` DERIVED from the table** (hand constant deleted; generate-settings flag default follows derivation); pin-matrix v2.14.0 cell green locally       | `desired_test.go`, `derived_min_test.go`, pin matrix A1–A4 ✓                        |
| P4/T14–T16        | Merge strategy table (path-keyed DATA; `--force-settings` = table flip); `Change` type with stable keys + `ConflictError`; `PlanDesired`/`ApplyPlan`; property specs: idempotency (`merge(merge(x))==merge(x)`), no-mutation, stable keys, conflict detection                                                                                                                      | `merge_test.go` green; lint clean                                                   |
| P5/T17–T18 (core) | Pure normalizers (config-version, duration, typecheck removal, tool-disabled moves, bad-keys, deprecation replacement + settings migration + rule renames with dedupe, redundancy removal); `NormalizeConfig`/`PruneOrphanSettings`; `Converge` pipeline; two-phase driver (normalize → save → analyze → plan → apply → save) with legacy analyzer contract (exit 69/65 preserved) | corpus byte-green (below)                                                           |
| P5/T19 (core)     | Audit emission from `Change` records (`AuditActionsFor`); recorder still legacy-side                                                                                                                                                                                                                                                                                               | audit specs partially green (see b)                                                 |
| P5/T20 (flip)     | **Converge is the default engine**; legacy = `SetEngine(EngineLegacy)` opt-out; CLI specs updated to new markers; `--pragmatic` wired to the fixer                                                                                                                                                                                                                                 | full CLI suite green; parity corpus green on BOTH engines                           |

**Verified byte-parity:** all 11 corpus fixtures produce byte-identical
configs on the converge engine vs the legacy-captured goldens
(`TestParityCorpusConverge` green). The repo's own 513-line config
(`tool-own`) re-converges to itself — the self-config trap is dead for the
tool's own config.

---

## b) Partially done (started, working core, known tail)

1. **P5/T18–T21 legacy-parity tail — 12 failing specs in `pkg/linter`**
   (all are legacy-contract specs now running against the converge engine):
   - User-disabled respect: converge re-adds linters that are in the user's
     `disable` list (shouldRender has no user-disable filter) → 4 specs.
   - Orphan-settings pruning not wired into the pipeline (normalizer exists,
     `PruneOrphanSettings` is not called by the driver) → 2 specs.
   - Audit action mapping: converge emits `updated` where the ledger specs
     expect `moved-to-disable`, `pruned-settings`, `exclusion-linter-renamed`,
     `exclusion-linter-removed`, `re-enabled` for overlay adds → 6 specs.
     The normalizers for all three surfaces EXIST and are unit-tested; the
     wiring into `Converge`/`planDisableOverlay`/`AuditActionsFor` is the
     remaining work (~1–2 h).

2. **P5/T21 sidecar overlays**: never-enable + anti-gaming enforcement work
   through `planDisableOverlay` (corpus proves it), but the suppression
   ledger entry (`suppressed-re-enable`) and the integration specs' exact
   ledger shapes are not emitted yet.

3. **Lint tail**: 21 findings after the flip wave (15×mnd on generated-table
   literals, 4×varnamelen, 1×prealloc, 1×staticcheck) — mechanical, quarantined
   to new files.

---

## c) Not started

| Phase          | What                                                                                                        | Why it is safe to defer                                                                                                                                 |
| -------------- | ----------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| P5/T20.2–T20.3 | Delete the old imperative path (`applyAllFixes`, preflight, recorder + their specs)                         | Legacy is an explicit opt-out behind `EngineLegacy`; deletion cascades into ~dozens of spec files. The corpus gates any future deletion.                |
| P6/T22         | Findings = diff (`Deviation` adapter, detector rewrite, delete `WithPriority`)                              | Detector still works; its drift-risk is now contained by the shared policy table                                                                        |
| P6/T23         | `ConfigAnalysis` slim-down, `Evaluation` type, report-generator ports, golden regen                         | Reports consume the analyzer, which is untouched                                                                                                        |
| P7/T24         | Friction `Evidence` → tier derivation, `--pragmatic` as threshold                                           | Evidence data is ALREADY in the table (ratios + counts); only the derivation fn + tier-map deletion remain                                              |
| P7/T25         | Typed redundancy relation table                                                                             | Redundancy already moved into records (`SupersededBy`) + `RemoveRedundancies`; only the legacy `RedundantLinters` map remains as a view-shaped leftover |
| P8/T26         | Self-host end-to-end proof (`tool-own` green IS the core proof; remaining: sidecar-as-divergence docs)      |                                                                                                                                                         |
| P8/T27         | ADRs (reconciler architecture, two version axes), AGENTS.md gotcha rewrites (#2/#6/#28), FEATURES/CHANGELOG | After the parity tail is green                                                                                                                          |

---

## d) Fucked up (went wrong, and what it cost)

1. **Generator circular bootstrap** (~30 min): after the var→view swap, the
   table generator read the views (derived from the stale table) and silently
   self-perpetuated — dropping the two no-replacement deprecation records and
   poisoning priorities. **Cost:** one data-loss cascade, recovered from the
   frozen views-goldens. **Lesson encoded:** the generator now reads ONLY the
   frozen golden JSON; the goldens are the regeneration source of truth.
2. **UsesTech variable swap** (~20 min): `known, detected := m[k]` — I named
   the value "known" and the presence "detected", then trusted the names over
   the types and "fixed" the wrong thing three times. **Lesson encoded:** the
   fix is `result, known :=` with the semantics in the names.
3. **Priority zero-value hazard** (~25 min): `LinterPriorityCritical == 0`
   means a zero `LinterPolicy` silently claims Critical. Caught by the views
   golden, fixed by explicit `Priority: Optional` on every record + a
   Validate rule (priority ⇔ rationale coupling).
4. **Ginkgo tree-construction panic** (~10 min): moving a helper into
   `BeforeEach` that `Entry(...)` arguments evaluate at construction time.
   Fixed by inlining the parse calls.
5. **Stale-dump whiplash** (~15 min): the parity dump dir kept serving
   previous-run artifacts, sending me in circles twice. Fixed by
   `rm -rf` before surveys.

No data was lost; every mistake was caught by a gate the plan built
(corpus, views-goldens, init validation) — which is exactly why those gates
exist.

---

## e) Improvements found along the way (implemented)

1. **Views-goldens as regeneration source** — better than the plan's
   "embed old maps in tests": byte-goldens serve as parity contract AND
   generator input AND future regression pins.
2. **`DerivedMinToolVersion`** kills gotcha #28's hand-sync permanently, and
   the derivation ignores never-rendered tiers (exhaustruct_v5's v2.13.0
   settings provably do not constrain the minimum).
3. **Path-rank plan ordering** — plans are ordered by canonical section
   order with generation order preserved inside sections, making apply
   order part of the byte contract (global key-sort broke exclusion-rule
   ordering).
4. **Legacy-exact exclusion matching** — the planner replicates the legacy
   last-rule-wins map semantics; subtle but corpus-proven.
5. **Analyzer contract preserved** — the flip initially degraded
   missing-binary/garbage-version failures into silent successes; the driver
   now enforces the classified exit codes (69/65) before converging.

## e2) Still-open improvement proposals

1. `RedundantLinters` (lll→golines) should move into the linter records
   (`SupersededByFormatter`), completing T25.
2. `normalize.go`'s rule-rename loop mutates while iterating (the drop-rule
   path shifts indices) — needs a rework to collect-then-filter for safety.
3. The converge `MigrationResult.Message` differs from legacy wording; the
   UI layer already adapts, but SDK consumers string-matching messages would
   see churn (flagged in CHANGELOG when T27 lands).

---

## f) Next 50 things (priority order)

**P5 closeout (blocking green):**

1. Respect user-disable in `shouldRender` (skip records in observed disable).
2. Wire `PruneOrphanSettings` into the converge pipeline (post-apply or plan phase).
3. Extend `AuditActionsFor`: settings-remove → `pruned-settings`.
4. Map enable→disable moves → `moved-to-disable` + `removed-from-enable`.
5. Map exclusion-rule renames → `exclusion-linter-renamed`.
6. Map exclusion-rule removals → `exclusion-linter-removed`.
7. Map overlay enable-adds → `re-enabled`.
8. Emit `suppressed-re-enable` when never-enable blocks an unjustified disable.
9. Fix rule-rename iteration mutation (collect indices, then filter).
10. Green the 12 pkg/linter legacy-parity specs.
11. Green `FixConfig sidecar + ledger integration` specs (ledger shapes).
12. Full suite green (`ginkgo -r -p -race`).
13. Lint to zero (15 mnd on table literals → named consts or nolint-with-reason).
14. Pin-matrix re-run (all four cells) after the tail lands.
15. `nix flake check` (format + build + tests).

**P6 findings = diff:**
16. `Deviation` adapter: `Change` → `finding.Finding`.
17. Lint-mode driver over `PlanDesired` (report-only).
18. Port detector onto the plan (replace `WithPriority` loop prevention).
19. Delete `WithPriority` machinery.
20. Findings parity spec vs old detector on corpus.
21. `Evaluation` type (Env/Desired/Observed/Deviations).
22. `ConfigAnalysis` counts → derived methods.
23. Port templ report generator.
24. Port json/sarif generators.
25. `UPDATE_GOLDEN=1` regen + review.

**P7 derivations:**
26. Friction `Evidence` complete-ness pass (AsOf dates for the four noise linters).
27. Tier derivation fn + threshold constant.
28. `--pragmatic` as threshold parameter + parity spec.
29. Delete hand tier-view leftovers (Legacy view fns become derived-only).
30. Move `RedundantLinters` into linter records.
31. Referential integrity specs for redundancy targets.

**P8 self-host + cleanup:**
32. Document tool-own sidecar divergences (self-host overlay file).
33. Self-verify: re-run on own config = zero diff (CI-able).
34. Delete legacy engine (applyAllFixes/preflight/recorder + specs).
35. Delete `configChangeRecorder` + `counts.total()==0` guard.
36. Delete compat views + legacy map vars (T27.1) — consumers move to views/table.
37. Delete now-dead tests/helpers.
38. ADR: reconciler architecture decision.
39. ADR: two version axes.
40. AGENTS.md gotcha #2 rewrite (templ output unchanged, but reference the corpus).
41. AGENTS.md gotcha #6 rewrite (normalization counting is gone — diff IS the count).
42. AGENTS.md gotcha #28 rewrite (min-version is derived).
43. FEATURES.md reconciler section + deprecation map re-point.
44. CHANGELOG entry (engine flip, message-shape changes).
45. TODO_LIST harvest (blocked on foreign changes clearing).
46. Update `docs/references/working-with-codebase.md` (adding a linter = ONE record).
47. Delete `cmd/gen-policy-table` (bootstrap done) or keep with provenance header.
48. Schema-fixture regen check after any curated-settings change.
49. Coverage check: `go run ./cmd/coverage-check -min=65` (new engine files counted).
50. Final full CI dry-run: `nix flake check` + CI parity + pin matrix green.

---

## g) Questions

1. **Legacy deletion timing** — T20.2 says delete the old imperative engine in
   the same PR series as the flip. The flip is live and byte-proven; deleting
   the legacy path cascades into ~30 spec files that assert legacy internals.
   Delete now (big diff, cleaner end state), or hold the legacy engine one
   release as the documented opt-out and delete in a follow-up?

2. **Audit action parity** — the converge engine emits Change-shaped audit
   actions; six ledger specs expect the legacy action vocabulary. Map the new
   changes onto the legacy action names (byte-stable ledger for SDK
   consumers), or treat the ledger as engine-internal and update the specs to
   the new vocabulary?

3. **`--pragmatic` semantics for user-enabled noise linters** — today both
   engines keep user enables (settings included). The e2e spec
   (`cmd_pragmatic_test.go`) originally expected unconditional exclusion;
   should `--pragmatic` (a) keep user enables (current, documented), or
   (b) strip user-enabled noise linters too (the spec's original reading)?
