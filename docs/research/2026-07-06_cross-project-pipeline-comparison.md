# Pipeline Comparison: `golangci-lint-auto-configure` vs `go-finding` vs `BuildFlow`

**Date:** 2026-07-06
**Scope:** Deep multi-dimensional comparison of the "pipeline" concept across three related LarsArtmann Go projects.

---

## The core insight: these are three _different kinds_ of pipelines on a dependency stack

The word "pipeline" means something genuinely different in each project, and they sit in a **layered relationship**:

```
   BuildFlow          ← whole-repo DAG orchestrator (top)
       │ uses go-finding core (Finding/Report), NOT pipeline/
       ▼
   go-finding         ← finding model + reusable fix-pipeline LIBRARY (middle)
       │
       ▼
   golangci-lint-     ← single-domain config-mutation pipeline (consumer)
   auto-configure       uses go-finding core for reporting, NOT pipeline/
```

| Project                                 | What "pipeline" means                                                                     | Topology                                       | Operates on                                       | Scope                       |
| --------------------------------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------- | ------------------------------------------------- | --------------------------- |
| **golangci-lint-auto-configure** (ours) | A sequential CLI fix-flow: `load → preflight-fix → analyze → apply-fixes → save → report` | **Linear, single-pass**                        | ONE golangci-lint config file                     | Single domain (lint config) |
| **go-finding**                          | A reusable `pipeline/` library: `detect → process → triage → apply → verify`              | **Iterative loop** (≤5 iterations), byte-level | Many findings across many files                   | Generic remediation library |
| **BuildFlow**                           | A build-orchestration **DAG** of ~90 steps with data-flow scheduling                      | **Directed acyclic graph**                     | An entire monorepo (multi-language, multi-module) | Whole-project CI            |

**Critical finding:** Both `golangci-lint-auto-configure` and `BuildFlow` import `github.com/larsartmann/go-finding v1.0.0` for the _finding model_ (`Finding`, `Report`, SARIF/JSON conversion) — but **neither consumes `go-finding/pipeline/`**. Each rolls its own remediation flow. That's the single most important fact in this comparison.

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

```
LoadConfig → GetLintersEnabled → detectVersion → runPreFlightChecks (4 mutations,
saved to disk so golangci-lint will run) → analyzeAndFix → applyLintersFix →
applyAllFixes+applyAndSave → runFmt
```

Pre-flight is **unusual**: it mutates-and-saves _before_ analysis (timeout, version, deprecated linters, typecheck) because golangci-lint refuses to start otherwise. Every save must increment `counts.normalization` or a `total()==0` guard **silently drops** all changes (`fixer.go:250`, AGENTS.md gotcha #6).

### go-finding — `Pipeline.Run` (`pipeline.go:146`), iterative

```
for iterations < MaxIterations:
    Detect (partial/graceful-degradation) → Process (FindingTransformer chain)
    → [len==0? break: ReasonStable]
    → Triage (Direct/Suggest/None by FixStrategy) → Apply (byte-level, descending)
    → LineShiftMap updates remaining findings → OnIteration
optional: CorrelateFindings, VerifyAfterFix (re-run detectors)
```

This is a genuine **convergent loop** — it re-detects after applying fixes, handles line shifts, and detects byte-level edit conflicts. None of that exists in our pipeline.

### BuildFlow — `runPipeline` (`execution/pipeline.go:42`), 10-stage DAG lifecycle

```
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

## 7. The go-finding relationship — the biggest gap

This is worth calling out explicitly because it's a latent architectural decision:

| Consumer                         | Uses go-finding core (`Finding`/`Report`/SARIF)?                                                     | Uses go-finding `pipeline/`?                                |
| -------------------------------- | ---------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| **golangci-lint-auto-configure** | Yes — `pkg/finding/` is a thin adapter (severity/category mapping, SARIF/JSON report, diff→findings) | **No** — has its own `fixer.go` linear pipeline             |
| **BuildFlow**                    | Yes — `gofinding.Finding`/`*gofinding.Report` is the interchange type in `FindingReport`             | **No** — delegates remediation to the DAG/tool capabilities |

So **`go-finding/pipeline/` is a fully-built, fuzzed, conflict-aware remediation engine that neither downstream consumer actually uses.** Our `pkg/linter/fixer.go` and BuildFlow's `execution/` reimplement remediation flows independently. That's either (a) deliberate — our domain is _config mutation_ not _source-byte editing_, so the byte-level edit engine doesn't fit; or (b) an opportunity to consolidate. Given go-finding's pipeline operates on `Finding`s with `BeforeCode`/`AfterCode` byte edits, and our fixer operates on YAML config semantics (enable/disable linters, merge maps), the **domains genuinely don't overlap** — go-finding's pipeline fixes _source code_, our pipeline fixes _config_. The non-use is justified.

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
- **go-finding is the most reusable & best-tested**: a generic, fuzz-validated, convergent remediation loop with byte-level conflict detection — but its `pipeline/` package has zero in-repo consumers, raising a "who uses this?" question worth answering before more investment.
- **BuildFlow is the most powerful & complex**: a true DAG orchestrator delegating hard problems (scheduling, retry, DI, audit) to battle-tested libraries, with resume, caching, circuit-breaking, and 10-format audit logging. The clear "top of stack" — and the one where "pipeline" most matches the conventional meaning.

The three are **complementary layers**, not competitors: go-finding defines the interchange + remediation primitive, BuildFlow orchestrates at repo scale, and our tool is a focused vertical that reports through go-finding's model. The healthy move is to keep boundaries clean and resist the temptation to grow our linear fixer into a DAG — that's BuildFlow's job.

---

## Actionable recommendations for our project

1. **Borrow go-finding's coverage-gate pattern** — add per-package coverage thresholds to CI (go-finding enforces core 98%, pipeline 95%, CLI 90%). We currently upload to Codecov with no threshold.
2. **Add a fuzz target on the config merger/fixer** — the `merger.go` multi-config merge and `fixer.go` normalization are pure functions ripe for fuzzing. go-finding proves this pattern works (23 fuzz targets).
3. **Add `govulncheck` to CI** — go-finding and BuildFlow both run it; we don't.
4. **Do NOT adopt `go-finding/pipeline/`** — the domain mismatch is real (config mutation vs source-byte editing). Our linear pipeline is correct for a single config file.
5. **Do NOT grow into a DAG** — BuildFlow owns that layer. Our value is focus.
