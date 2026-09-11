# Pipeline Comparison: `golangci-lint-auto-configure` vs `go-finding` vs `BuildFlow`

> **🔄 RETROACTIVE UPDATE — 2026-07-16**
>
> The 5 actionable recommendations at the end of this report have the following status:
>
> | # | Recommendation                            | Status       | Details                                                                                                                                                                  |
> | - | ----------------------------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
> | 1 | Borrow go-finding's coverage-gate pattern | ✅ Done      | Coverage threshold gate added to CI. Originally `scripts/coverage-check.sh`; replaced by `cmd/coverage-check/main.go` (portable Go program with BDD tests) in July 2025. |
> | 2 | Add fuzz target on config merger/fixer    | ✅ Done      | `FuzzMergeConfigInto` + `FuzzMergeIdempotent` in `pkg/config/merger_fuzz_test.go`                                                                                        |
> | 3 | Add govulncheck to CI                     | ✅ Done      | CI job added in `.github/workflows/ci.yml`                                                                                                                               |
> | 4 | Do NOT adopt go-finding/pipeline/         | ✅ Validated | Domain mismatch confirmed — config mutation ≠ source-byte editing                                                                                                        |
> | 5 | Do NOT grow into a DAG                    | ✅ Validated | BuildFlow owns that layer; our value is focus                                                                                                                            |
>
> All 5 recommendations resolved. The architectural analysis remains evergreen. This was a one-time research doc; no ongoing maintenance needed.

**Date:** 2026-07-06
**Scope:** Deep multi-dimensional comparison of the "pipeline" concept across three related LarsArtmann Go projects.

---

## The core insight: these are three _different kinds_ of pipelines on a dependency stack

The word "pipeline" means something genuinely different in each project, and they sit in a **layered relationship**:

```text
                 go-finding (Finding/Report model + pipeline/ SDK)
                ╱        ╱              ╲              ╲
go-structure-linter   golangci-lint-     BuildFlow    hierarchical-errors
(model + pipeline)    auto-configure     (model only)  (model only)
                      (model only)
```

> **Broader ecosystem note:** This report originally compared only three projects. Two sibling reports — `go-structure-linter/docs/pipeline-comparison.md` and `hierarchical-errors/docs/pipeline-comparison.md` — revealed two additional consumers (go-structure-linter, hierarchical-errors) and one **critical correction** documented in [Appendix B](#appendix-b-corrections-from-cross-referencing-sibling-reports). The tables below focus on the original three; the broader five-project picture lives in the appendices.

| Project                                 | What "pipeline" means                                                                     | Topology                                       | Operates on                                       | Scope                       |
| --------------------------------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------------------------- | --------------------------- |
| **golangci-lint-auto-configure** (ours) | A sequential CLI fix-flow: `load → preflight-fix → analyze → apply-fixes → save → report` | **Linear, single-pass**                        | ONE golangci-lint config file                     | Single domain (lint config) |
| **go-finding**                          | A reusable `pipeline/` library: `detect → process → triage → apply → verify`              | **Iterative loop** (≤5 iterations), byte-level | Many findings across many files                   | Generic remediation library |
| **BuildFlow**                           | A build-orchestration **DAG** of ~90 steps with data-flow scheduling                      | **Directed acyclic graph**                     | An entire monorepo (multi-language, multi-module) | Whole-project CI            |

**Key finding:** `golangci-lint-auto-configure` and `BuildFlow` both import `github.com/larsartmann/go-finding v1.0.0` for the _finding model_ (`Finding`, `Report`, SARIF/JSON conversion) but **neither consumes `go-finding/pipeline/`** — each rolls its own remediation flow. However, **go-structure-linter DOES consume both** the model and the pipeline (see [Appendix B](#appendix-b-corrections-from-cross-referencing-sibling-reports)).

---

## 1. Architecture & execution model

| Dimension       | golangci-lint-auto-configure                                                  | go-finding                                                                               | BuildFlow                                                                                                                |
| --------------- | ----------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **Engine**      | Hand-written linear sequence in `pkg/linter/fixer.go:29` `FixConfig`          | Self-contained loop in `pipeline/pipeline.go:146` `Run`                                  | Delegated to **`Azure/go-workflow`** (`execution/pipeline.go:42`) — "ZERO custom orchestration"                          |
| **Scheduling**  | Implicit call order                                                           | Iteration counter (`MaxIterations`, default 5)                                           | **DAG with data-flow edges** (Bazel/make `outputsOverlapInputs` model, `domain/data_flow.go:22`); native cycle detection |
| **Concurrency** | Minimal — `errgroup` for parallel linter+formatter parsing (`analyzer.go:97`) | `errgroup` for parallel detectors (`pipeline_detect.go:113`); opt-in `ParallelDetectors` | Global `MaxConcurrency` (default 4) lease-bucket scheduler; per-module fan-out capped at 4                               |
| **Single-use?** | No (re-runnable)                                                              | **Yes** — `ran bool` guard returns `errAlreadyRan` (`pipeline.go:17`)                    | No — resumable/re-runnable                                                                                               |
| **Idempotency** | Core design — every mutation checks "already set" first                       | Loop terminates on `len(findings)==0` → `ReasonStable`                                   | Healed-step correction + resume checkpoint                                                                               |
| **Determinism** | Provider list static (`linter_priorities.go` constants)                       | Detectors sorted; edits applied descending-by-offset                                     | Providers sorted by tool name before DAG build                                                                           |

---

## 2. The remediation loop in detail

### Ours — `FixConfig` (`fixer.go:29`), linear 7-stage

```text
LoadConfig → GetLintersEnabled → detectVersion → runPreFlightChecks (4 mutations,
saved to disk so golangci-lint will run) → analyzeAndFix → applyLintersFix →
applyAllFixes+applyAndSave → runFmt
```

Pre-flight is **unusual**: it mutates-and-saves _before_ analysis (timeout, version, deprecated linters, typecheck) because golangci-lint refuses to start otherwise. Every save must increment `counts.normalization` or a `total()==0` guard **silently drops** all changes (`fixer.go:250`, AGENTS.md gotcha #6).

### go-finding — `Pipeline.Run` (`pipeline.go:146`), iterative

```text
for iterations < MaxIterations:
    Detect (partial/graceful-degradation) → Process (FindingTransformer chain)
    → [len==0? break: ReasonStable]
    → Triage (Direct/Suggest/None by FixStrategy) → Apply (byte-level, descending)
    → LineShiftMap updates remaining findings → OnIteration
optional: CorrelateFindings, VerifyAfterFix (re-run detectors)
```

This is a genuine **convergent loop** — it re-detects after applying fixes, handles line shifts, and detects byte-level edit conflicts. None of that exists in our pipeline.

### BuildFlow — `runPipeline` (`execution/pipeline.go:42`), 10-stage DAG lifecycle

```text
AugmentPath → resolveStore → (single-step OR build-mode filter) → circuit-breaker
→ resume-filter → BuildProjectState → health-checks → Build DAG →
Workflow.Do(ctx) → reportSkipped/Finish/convertResults/healState/audit-snapshot
```

Node _topologies_ vary per tool: `detect→If(hasFindings)→repair`, `diagnose→Switch(case)→repair`, per-module fan-out via embedded workflows, or standalone. Data-flow edges are _derived_ from write→read overlap; formatter ordering is explicit `DependsOn`.

---

## 3. Retries, errors & resilience

| Aspect                | Ours                                                                                                              | go-finding                                                                                                                               | BuildFlow                                                                                                                                           |
| --------------------- | ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Retry**             | `utils.WithRetry` 3-phase on commands; only on `"parallel golangci-lint is running"` output (`command_runner.go`) | `RetryDetector` wraps each detector: exp backoff `2^n×100ms` capped 5s + ±25% jitter; 3 retries; context errors never retried            | `cenkalti/backoff/v4` exp backoff, **shared `RetryBudget`** (atomic) across all steps; only `Transient` family retries (fail-open for plain errors) |
| **Conflict handling** | None (single file)                                                                                                | **Two modes**: position-based (`conflict.go:35`) + byte-level (`ByteLevelConflictDetection`); `AnalyzeConflicts` returns conflict groups | DAG-level: `flow.If`/`Switch` branch on detection results; healed-step mechanism flips failed→success                                               |
| **Partial failure**   | `--check` exits Conflict(1) if fixes>0; pre-flight saves incrementally                                            | `GracefulDegradation` collects per-detector errors into `PartialResult.Errors`, continues (`partial.go`)                                 | `SkipAsError` (strict mode); circuit breaker skips chronically-failing steps (>80% fail over 5+ runs)                                               |
| **Error model**       | `go-error-family` v0.5.1 — `Classified` interface + sentinel registration → BSD sysexit codes (1/65/69/75)        | Sentinel errors (`ErrAlreadyRan`, `ErrPartialDetection`)                                                                                 | `go-error-family` classification gates retry; `buildflow_error` hierarchical types                                                                  |
| **Crash recovery**    | None — git is the safety net (`IsGitRepo` mandatory)                                                              | `FileBackup` (temp dir, FNV-1a hashed, atomic restore, `RollbackAll`)                                                                    | **Resume store**: SQLite checkpoint per-step; staleness check on git HEAD; `MarkStepHealed` corrects checkpoint                                     |

---

## 4. Extensibility & plugin model

| Aspect                  | Ours                                                                                                                                         | go-finding                                                                                                                   | BuildFlow                                                                                                                                                                   |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Extension points**    | `types.ConfigLoader` & `types.LinterAnalyzer` **interfaces** (DI seams); `migration.Validator` swappable; `config.FS` filesystem abstraction | `Detector` interface; `FindingTransformer`; `FixProvider` chain; `StageHook`; `TriageFunc`; `DetectorRegistry` (thread-safe) | `Tool` = Spec + FileContract + **6 capability interfaces** (Detector/Repairer/Generator/CommandExecutor/SystemChecker/HealthChecker); function adapters (`DetectorFunc`...) |
| **How to add a "step"** | Edit static constants (`linter_priorities.go` + `linter_reasons.go` + presets) — AGENTS.md gotcha #10                                        | Implement `Detector` + register                                                                                              | One constructor in `tools/providers/*.go` + one line in `CreateAllProviders`; no registry/loader (Go `plugin` deleted as Nix/Windows-incompatible)                          |
| **DI framework**        | **Manual** (no `internal/di/` — AGENTS.md gotcha #9)                                                                                         | None (constructor injection)                                                                                                 | **`samber/do` v2** — lazy singletons, `do.Package` grouping, auto-shutdown                                                                                                  |
| **Provider count**      | ~119 linter priorities, 6 presets, 10 deprecated mappings (all static data)                                                                  | ~3 default fix providers + AST provider                                                                                      | **~90 v3 providers** across Go/Rust/Python/JS/Nix/Proto                                                                                                                     |

---

## 5. Caching, persistence & observability

| Aspect                   | Ours                     | go-finding                                                                                            | BuildFlow                                                                                                                                                        |
| ------------------------ | ------------------------ | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Caching**              | None                     | `goast.Provider` caches parsed AST per file-content hash                                              | **Content-addressed cache** (`CacheKey{CheckerName,OptionsHash,FilesHash}`); file + SQLite persisters; `CachedChecker` wrapper; success derived (no split-brain) |
| **Persistence**          | Config file (0600)       | In-memory + file backups                                                                              | **SQLite** (`modernc.org/sqlite`, no CGO, WAL): resume checkpoint + timings store; `dbstore` module                                                              |
| **Metrics**              | `fixCounts` (6 counters) | Thread-safe `Metrics` (stage durations, detector times, findings/fixes counts) → immutable `Snapshot` | Per-step `StepTelemetryFunc`; timings table; **workflow audit log** exportable to 10 formats (json/mermaid/d2/dot/plantuml...); PostHog telemetry (opt-out)      |
| **Telemetry philosophy** | Local logging only       | Local logging + callbacks (`OnFinding`/`OnFix`/`OnIteration`)                                         | Privacy-first PostHog, self-hostable, `BUILDFLOW_TELEMETRY_DISABLED`                                                                                             |

---

## 6. CI/CD & build maturity

| Aspect                | Ours                                                                                               | go-finding                                                                                                         | BuildFlow                                                                                         |
| --------------------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------- |
| **CI jobs**           | 4 (test+build, lint, nix, summary)                                                                 | **8** (test matrix, coverage, lint, govulncheck, module-isolation, stress `-count=20`, dupl, benchmark-regression) | 4 (build-and-test self-check, lint-and-format, security, module-tests)                            |
| **Coverage gate**     | Uploaded to Codecov, no threshold                                                                  | **Per-package thresholds enforced**: core 98%, pipeline 95%, CLI 90%, total 93%                                    | 60% threshold                                                                                     |
| **Release security**  | GoReleaser v2 + **Cosign + Syft SBOM**                                                             | GoReleaser v2 + **Cosign + Syft SBOM** (linux/darwin/win × amd64/arm64)                                            | GoReleaser v2 + Cosign + SBOM + Homebrew tap                                                      |
| **Nix**               | `flake-parts` + `treefmt-nix`; `mkPreparedSource` injects private deps; `vendorHash` manual update | `flake-parts`; `postPatch = "rm -f go.work"` for `GOWORK=off`                                                      | `flake-parts`; `mkPreparedSource`; devShell overrides `go()` to warn against root `go test ./...` |
| **Fuzzing**           | None                                                                                               | **23 fuzz targets** + corpus (`testdata/fuzz/`); proves optimized edit-application = naive O(n²)                   | 9 fuzz targets                                                                                    |
| **Stress/regression** | None                                                                                               | `go test -race -count=20` + benchstat vs `baseline.txt` (25% regression gate)                                      | Self-check: `./buildflow --build-mode full`                                                       |

---

## 7. The go-finding relationship — who consumes what

This is worth calling out explicitly because it's a key architectural decision:

| Consumer                         | Uses go-finding core (`Finding`/`Report`/SARIF)?                                                     | Uses go-finding `pipeline/`?                                |
| -------------------------------- | ---------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| **golangci-lint-auto-configure** | Yes — `pkg/finding/` is a thin adapter (severity/category mapping, SARIF/JSON report, diff→findings) | **No** — has its own `fixer.go` linear pipeline             |
| **BuildFlow**                    | Yes — `gofinding.Finding`/`*gofinding.Report` is the interchange type in `FindingReport`             | **No** — delegates remediation to the DAG/tool capabilities |
| **go-structure-linter**          | Yes — `types.Issue = finding.Finding` (type alias, zero conversion)                                  | **Yes** — `pipeline.RunDetectOnly` + `RunFixPipeline`       |
| **hierarchical-errors**          | Yes — `pkg/finding/` bridge via `ViolationToFinding`                                                 | **No** — linear 17-stage static-analysis pipeline           |

> **Correction (see [Appendix B](#appendix-b-corrections-from-cross-referencing-sibling-reports)):** An earlier version of this report claimed `go-finding/pipeline/` had "zero downstream consumers." That was wrong — it was based on checking only our project and BuildFlow. **go-structure-linter is a genuine consumer** of both the model and the pipeline SDK. The real question is not "does anyone use it?" but "why don't the config/build/error-analysis verticals use it?"

### Why we don't use `go-finding/pipeline/`

go-finding's pipeline operates on `Finding`s with `BeforeCode`/`AfterCode` byte edits — it fixes _source code_. Our fixer operates on YAML config semantics (enable/disable linters, merge maps) — it fixes _config_. The **domains genuinely don't overlap**. The same applies to hierarchical-errors (AST-based error analysis) and BuildFlow (whole-tool orchestration). go-structure-linter is the natural fit because it _does_ produce source-level findings with byte-level fix data. The non-use by the other three verticals is architecturally justified, not an oversight.

---

## 8. Testing strategy comparison

| Aspect              | Ours                                                   | go-finding                                                                                                  | BuildFlow                                                                                |
| ------------------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| **Framework**       | Ginkgo v2 + Gomega (BDD)                               | Ginkgo v2 + Gomega (BDD)                                                                                    | Ginkgo v2 + Gomega (BDD); **testify banned** (CI gate blocks new imports)                |
| **Test files**      | ~56 `_test.go` files                                   | 58+ test files + `*_extra_test.go` + `*_bugfix_test.go`                                                     | 266 test files across all workspace modules                                              |
| **Fuzz tests**      | None                                                   | **23 fuzz targets** across 8 files with corpus in `testdata/fuzz/`                                          | 9 fuzz targets across 6 files                                                            |
| **Property tests**  | None                                                   | `property_test.go` — 6 properties (monotonicity, partition completeness, ID round-trip, merge preservation) | Snapshot tests (`__snapshots__/`), compile tests                                         |
| **Benchmarks**      | 5 bench files (linter, analyzer, diff, detection, set) | 6 bench files with `benchmarks/baseline.txt` regression tracking                                            | `workflow_bench_test.go`, persister + scroll renderer benches                            |
| **Integration/E2E** | `integration_test.go` in CLI                           | `pipeline/integration_test.go`, `cmd/go-finding/e2e_test.go`                                                | 9+ E2E/integration test files (circuit breaker, retry budget, resume cycle, concurrency) |
| **Race detection**  | `go test -race` + Nix `checks.race`                    | `go test -race` + CI stress (`-count=20`)                                                                   | `just test-race`                                                                         |

---

## 9. Domain model comparison

| Aspect                                       | Ours                                               | go-finding                                                                                             | BuildFlow                                                                                                                |
| -------------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| **Core type**                                | `types.Config` (tri-tagged struct: json/toml/yaml) | `Finding` (immutable value type with branded fields)                                                   | `Tool` (Specification + FileContract + capabilities)                                                                     |
| **Branded types**                            | `LinterPriority` (int enum)                        | `ID`, `RuleName`, `ToolName`, `FilePath` (4 distinct string types)                                     | `ToolName`, `BinaryName`, `StepName`, `Language`, `FilePattern`, `Manifest`, `ModuleRole`, `RunID` (8+ branded strings)  |
| **"Make impossible states unrepresentable"** | Partial — `Set[T]` with algebra, `RuleKey()` dedup | Strong — immutable findings, branded types, `NormalizeFixStrategy` eliminates empty-string split-brain | Strongest — constructor-validated value objects, `ActionResult` auto-computes `Remaining`, `StepResultStatus` singletons |
| **Collections**                              | `types.Set[T]` (generic, full algebra)             | `IntervalIndex[T]` (generic sorted-slice overlap)                                                      | `ValidationResult[T PrioritizedViolation]` (generic with severity counts)                                                |
| **Module/monorepo structure**                | Single module                                      | **4 modules** (core/pipeline/analysis/CLI) via `go.work` + replace                                     | **26+ modules** via `go.work`; strict acyclic dependency graph                                                           |

---

## 10. Verdict

- **Ours is the simplest and most focused**: a linear, single-file, idempotent config-mutation pipeline. Correct for its domain. Its sophistication lives in _domain knowledge_ (119 linter priorities, version-gated deprecation, presets) rather than in _orchestration_. Weaknesses: no crash recovery beyond git, no metrics/observability, no fuzzing, manual DI.
- **go-finding is the most reusable & best-tested**: a generic, fuzz-validated, convergent remediation loop with byte-level conflict detection. Its `pipeline/` package has one confirmed consumer (go-structure-linter) and the model layer is used by all four downstream projects. The question of whether the config/build/error verticals _should_ adopt it is answered: no — domain mismatch (see [Appendix B](#appendix-b-corrections-from-cross-referencing-sibling-reports)).
- **BuildFlow is the most powerful & complex**: a true DAG orchestrator delegating hard problems (scheduling, retry, DI, audit) to battle-tested libraries, with resume, caching, circuit-breaking, and 10-format audit logging. The clear "top of stack" — and the one where "pipeline" most matches the conventional meaning.

The three are **complementary layers**, not competitors: go-finding defines the interchange + remediation primitive, BuildFlow orchestrates at repo scale, and our tool is a focused vertical that reports through go-finding's model. The healthy move is to keep boundaries clean and resist the temptation to grow our linear fixer into a DAG — that's BuildFlow's job.

---

## Actionable recommendations for our project

1. **Borrow go-finding's coverage-gate pattern** — add per-package coverage thresholds to CI (go-finding enforces core 98%, pipeline 95%, CLI 90%). We currently upload to Codecov with no threshold.
2. **Add a fuzz target on the config merger/fixer** — the `merger.go` multi-config merge and `fixer.go` normalization are pure functions ripe for fuzzing. go-finding proves this pattern works (23 fuzz targets).
3. **Add `govulncheck` to CI** — go-finding and BuildFlow both run it; we don't.
4. **Do NOT adopt `go-finding/pipeline/`** — the domain mismatch is real (config mutation vs source-byte editing). Our linear pipeline is correct for a single config file.
5. **Do NOT grow into a DAG** — BuildFlow owns that layer. Our value is focus.

---

## Appendix A: Broader ecosystem — five projects at a glance

After cross-referencing `go-structure-linter/docs/pipeline-comparison.md` and `hierarchical-errors/docs/pipeline-comparison.md`, the full ecosystem picture is:

```text
                 go-finding (Finding/Report model + pipeline/ SDK)
                ╱        ╱              ╲              ╲
go-structure-linter   golangci-lint-     BuildFlow    hierarchical-errors
(model + pipeline)    auto-configure     (model only)  (model only)
                      (model only)
```

| Project                          | Pipeline model  | Stages | Uses go-finding model? | Uses go-finding pipeline/? | Unique strength                                        |
| -------------------------------- | --------------- | ------ | ---------------------- | -------------------------- | ------------------------------------------------------ |
| **go-finding**                   | Iterative loop  | 5      | — (is the foundation)  | — (is the foundation)      | Byte-level fix engine, SARIF round-trip, 23 fuzz       |
| **golangci-lint-auto-configure** | Linear          | 7      | Yes                    | No                         | 119 linter priorities, version-gated deprecation       |
| **BuildFlow**                    | DAG             | 14     | Yes                    | No                         | 90+ providers, data-flow edges, resume, healing        |
| **go-structure-linter**          | Adapter over GF | ~3     | Yes (type alias)       | **Yes**                    | 65 structure rules, two-path fix (safe vs raw)         |
| **hierarchical-errors**          | Linear          | 17     | Yes                    | No                         | Error hierarchy graphs, 10 runtime plugins, 11 formats |

### Dimensions the original three-project comparison missed

These came to light only after reading the sibling reports:

| Dimension                 | Details                                                                                                                                                                                                                                |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Runtime plugin system** | hierarchical-errors is the **only** project with a runtime plugin system — 10 lifecycle hooks (pre-analysis → post-output), auto-init builtins, `--disable-plugins`. All others are compile-time only.                                 |
| **Watch mode**            | hierarchical-errors and BuildFlow both have `--watch` (file watcher, 500ms debounce, mutex overlap prevention). Neither ours nor go-finding has this.                                                                                  |
| **Caching as a spectrum** | Not binary. hierarchical-errors sits in the middle: otter file cache (max 1000 files/50MB) + LRU package cache (100 entries/500MB) + RelationshipTracker adjacency maps. Between our "none" and BuildFlow's SQLite.                    |
| **Output format breadth** | hierarchical-errors has **11 formats** (tree, json, html, full, hierarchy, dot, mermaid, agent, sarif, go-finding-sarif, go-finding-json) — including DOT graphviz and Mermaid diagrams. go-finding has 6, BuildFlow has 6, we have 4. |
| **Plugin safety split**   | go-structure-linter has two fix paths: `FixService` (safe: backup + validate + automatic restore) vs `RunFixPipeline` (raw, no safety — SDK only). Worth studying as a pattern.                                                        |
| **Build-mode durations**  | BuildFlow: full ~5-10min, fast ~5-30sec, pre-commit ~5-10sec, dev ~5-30sec, lightning ~1-2sec. Concrete numbers missing from the original report.                                                                                      |

---

## Appendix B: Corrections from cross-referencing sibling reports

This appendix documents factual errors in the original version of this report, discovered after reading `go-structure-linter/docs/pipeline-comparison.md` and `hierarchical-errors/docs/pipeline-comparison.md`.

### B1. CRITICAL: "go-finding/pipeline/ has zero consumers" — FALSE

**Original claim (Section 7):** "go-finding/pipeline/ is a fully-built, fuzzed, conflict-aware remediation engine that neither downstream consumer actually uses."

**Reality:** **go-structure-linter is a genuine consumer** of both the model and the pipeline SDK:

- `pipeline.RunDetectOnly(ctx, path, opts, factory)` — detect-only mode with `MaxIterations=1`
- `RunFixPipeline()` / `FixWithPipeline()` — SDK fix path via `ApplyFixesFromTriage`
- `types.Issue = finding.Finding` — type alias, zero conversion

The original research only checked our `go.mod` and BuildFlow's `go.mod` for `go-finding/pipeline` imports. go-structure-linter was not in scope. The corrected consumer table is in [Section 7](#7-the-go-finding-relationship--who-consumes-what) above.

**Impact on conclusions:** The "who uses this?" investment question is answered — go-structure-linter uses it. The real question is narrower: "why don't the config/build/error verticals use it?" — and the answer is domain mismatch (config mutation vs source-byte editing vs AST analysis vs tool orchestration).

### B2. The ecosystem diagram was incomplete

The original three-node diagram omitted go-structure-linter and hierarchical-errors. The corrected five-node diagram is in [Appendix A](#appendix-a-broader-ecosystem--five-projects-at-a-glance).

### B3. Missing dimensions

The original report missed runtime plugins (hierarchical-errors), watch mode (hierarchical-errors + BuildFlow), the caching middle tier (hierarchical-errors' otter/LRU), and output-format breadth (hierarchical-errors' 11 formats). These are documented in [Appendix A](#appendix-a-broader-ecosystem--five-projects-at-a-glance).

### B4. What the original report got right (unchanged)

- **The domain-mismatch analysis** remains valid: go-finding/pipeline/ fixes _source bytes_ (`BeforeCode`/`AfterCode`); our fixer fixes _config semantics_. That's why we don't use it.
- **The CI/CD maturity comparison** with concrete job counts, coverage thresholds, and fuzz targets — both sibling reports omit CI entirely.
- **The "keep boundaries clean" recommendation** — reinforced by go-structure-linter's successful adapter pattern.
