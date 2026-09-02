# Pareto Plan v2: golangci-lint Friction Reduction

**Date:** 2026-07-25 07:56 · **v2 revision:** 2026-07-25
**Author:** Pareto-planning session, grounded in `docs/research/2026-07-25_golangci-config-ecosystem-report.md`
**Status:** PROPOSED — awaiting approval before execution

> **Update (2026-07-25, later session):** this plan was **EXECUTED** the same
> day. Phases 0–4 shipped (C0–C9, C11–C17); C18 (validation) produced
> `docs/research/validation-delta.md`; **C19 (version bump + tag) is the only
> unfinished task** — now tracked in `TODO_LIST.md`. Per-task status in
> [Execution resolution](#execution-resolution-2026-07-25) at the end of this
> file. The DoD checklist in §6 is superseded by that resolution table.

> **What changed in v2?** A critical re-review of v1 against the _actual_ codebase found
> **6 correctness defects + 3 structural gaps**. They are listed in [§7 Revision history](#7-revision-history--v1--v2-changelog)
> and fixed throughout. v1 is preserved in git history (`62baa44`).

---

## 0. Context — why this plan exists

A cross-project audit of **160 first-party `.golangci.yml` configs** (sibling projects under `/home/lars/projects/`) plus **5 201 `//nolint` directives** in non-test Go code produced hard evidence of where linting friction concentrates. Full data: `docs/research/2026-07-25_golangci-config-ecosystem-report.md`. The headline:

| Linter             | enabled in | `//nolint` | friction (nolint÷enable) | verdict                                                                          | has config knobs?           |
| ------------------ | ---------: | ---------: | -----------------------: | -------------------------------------------------------------------------------- | --------------------------- |
| `exhaustruct`      |        146 |    **952** |                  **6.5** | Most-hated — forces exhaustive struct literals on every `http.Server{}`, `Cmd{}` | `exclude []` ✅             |
| `gochecknoglobals` |        153 |        773 |                      5.1 | Fights standard Go patterns (registries, sentinels)                              | **none** ❌                 |
| `gosec`            |        154 |        589 |                      3.8 | Security linter, many false positives in test fixtures                           | `excludes []` ✅            |
| `errcheck`         |        143 |        443 |                      3.1 | Noisy on `defer Close()`, `fmt.Fprint`                                           | `exclude-functions` ✅      |
| `wrapcheck`        |        150 |        265 |                      1.8 | Demands every error be wrapped                                                   | `ignore-sigs` ✅            |
| `ireturn`          |        136 |        171 |                      1.3 | Conflicts with common interface-returning APIs                                   | `allow []` ✅ (already set) |
| `funlen`           |        151 |        126 |                     0.83 | Default `60/40` too strict vs. house style `200/100`                             | `lines/statements` ✅       |

**The goal:** cut friction (nolint spam) **without sacrificing real defect detection.** The zero-friction linters (`copyloopvar`, `errorlint`, `intrange`, `nilnesserr`, `unconvert`, `unparam`, `sloglint`…) stay untouched.

### How the default enable set actually works (v1 got this wrong)

The `~109`-linter enable set is **NOT** a static list and **NOT** derived from `PresetLinters` (those are tiny: `strict`=17, `reference`≈40). It is **dynamically computed at runtime** by `pkg/linter/categorizer.go:21` (`CategorizeLinters`):

> `enable-set = (golangci-lint linters → Disabled) − Deprecated − DisabledLinters − RedundantLinters − ProjectSpecific − LinterMinVersions`

with `--priority optional` (the CLI default, `cmd_configure.go:96`) admitting every surviving linter. **This is why v1's "pragmatic = strict minus 5 linters" was impossible** — `exhaustruct`, `gochecknoglobals`, `ireturn`, `wrapcheck` are not even _in_ `strict`. To drop a linter from the dynamic set you either (a) add it to `DisabledLinters` (`rules.go:111`), (b) lower its `LinterPriorities` entry and raise `--priority`, or (c) add a new skip-rule in `shouldSkipLinter`. v2 uses (c) for the opt-in `--pragmatic` flag.

### What we will NOT do (anti-verschlimmbessern guardrails)

- **NOT removing `exhaustruct`/`gosec` from defaults.** They catch real bugs. We curate _excludes_, not presence.
- **NOT removing `gochecknoglobals` from defaults** either (v2 decision, see C5b). It catches real accidental globals; its 5.1 friction is handled by the opt-in `--pragmatic` flag, not by changing everyone's defaults.
- **NOT changing the formatter set** `{gci, gofumpt, goimports, golines}` — validated winning stack (128/160).
- **NOT touching `issues: (50, 10)`** — validated standard (128/160).
- **NOT rewriting the preset engine** — `--pragmatic` is a skip-rule, not a new static preset.
- **NOT ripping out v1 config support** — declare maintenance-only, don't delete.
- **NOT reverting the auto-commit daemon's work** or any change we didn't author.

---

## 1. Pareto Breakdown

### 1% that delivers 51% of the result

**Expand the `exhaustruct` default `Exclude` list with common stdlib structs.**

`DefaultLinterSettings["exhaustruct"]` (`pkg/constants/linter_settings.go:156`) has exactly **one** entry: `os/exec.Cmd`. Every project manually re-derives the same list (`net/http.Server`, `net/http.Client`, `net/http.Request`, `time.Ticker`, `bytes.Buffer`, `sync.WaitGroup`…). This single change eliminates the #1 friction source: **952 nolints across 146 configs**. One map literal, ~12 lines.

### 4% that delivers 64% of the result

The 1% change **plus four more** targeting the next friction leaders (v1 only did three — it skipped errcheck):

1. **High-friction test-file exclusions** (`pkg/constants/config.go:75`) for `gosec`, `errcheck`, `wrapcheck` — the three above friction 1.5 with heavy test-fixture noise.
2. **Tune `funlen`** from `60/40` to house style `200/100` (`linter_settings.go:191`) **and reconcile this repo's own `.golangci.yml` which uses `30/20`** (the 3-way mismatch flagged in `TODO_LIST.md`). Note: this _intentionally diverges_ from golangci-lint upstream default (the existing comment says "matching upstream" — that comment must be updated).
3. **Add `GosecSettings` typed struct + curated excludes** (G104 unhandled error in CLI main, G304 file-by-variable in loaders).
4. **Add `ErrcheckSettings` typed struct + curated `exclude-functions`** (`(*os.File).Close`, `(io.Closer).Close`, `fmt.Fprint*`) — v1 gave gosec a settings struct but left errcheck (higher count of _configurable_ friction) with only test-exclusion. **This is v2's Pareto-fairness fix.**

### 20% that delivers 80% of the result

5. **`--pragmatic` flag** (`pkg/linter/categorizer.go` `shouldSkipLinter`) that drops the 5 noise leaders `{exhaustruct, gochecknoglobals, ireturn, wrapcheck, funlen}` from the _dynamic_ enable set on demand. This is the **correct** mechanism v1's static-preset approach couldn't be. It gives projects an escape hatch that isn't "disable everything."
6. **Resolve the sidecar deadlock.** `.golangci-lint-auto-configure.yml` has **0 adoption** across 160 projects. Promote or de-emphasize — decision + docs only.
7. **`golangci-lint run` (no `--fix`) CI gate** (already `TODO_LIST.md` Medium-High). BuildFlow's `--fix` silently swallows unfixable issues.
8. **Publish findings** in `FEATURES.md`, `AGENTS.md`, `CHANGELOG.md`.

### The other 20% (to reach 100%)

9. **Lock the formatter quadruple as a documented `house` formatter preset** + test.
10. **Declare v1 config support maintenance-only** (0 live v1 configs).
11. **Test-debt payoff** (from `TODO_LIST.md`). _Note:_ the HIGH-priority "split `cmd_configure.go` (581 lines)" is tracked there but is an SRP refactor, not friction reduction — explicitly out of scope here.

### The closing loop v1 forgot (Phase 0 + Phase 5)

Without a **before/after measurement**, "reduce friction" is unverifiable. v2 adds:

- **Phase 0 (baseline):** freeze current per-linter nolint counts + 5 representative sibling configs as the _before_ snapshot.
- **Phase 5 (validate):** regenerate those 5 configs with the new tool, re-measure, report the delta. **This is the only proof the plan worked.**

---

## 2. Coarse Plan — tasks 30–100 min each

Sorted by **impact ÷ effort** (descending). `Tier` = Pareto bucket. `Verdict` = quick-win / strategic / hygiene. **[v1→v2]** marks changed tasks.

| #   | Task                                                                                                                           | Tier  | Impact    | Effort | Verdict   | Key files                                          |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | ----- | --------- | ------ | --------- | -------------------------------------------------- |
| C0  | **[NEW]** Freeze baseline friction metrics (per-linter nolint counts + 5 sibling-config "before" snapshots)                    | loop  | —         | 30m    | hygiene   | `/tmp/friction.py`, `docs/research/`               |
| C1  | Expand `exhaustruct` default `Exclude` with stdlib structs + data-integrity test                                               | 1%    | Very High | 60m    | quick-win | `linter_settings.go:156`, `data_integrity_test.go` |
| C2  | **[SPLIT]** High-friction test-file exclusions: `gosec`, `errcheck`, `wrapcheck`                                               | 4%    | Very High | 25m    | quick-win | `config.go:75`                                     |
| C2b | **[NEW]** Low-friction test-file exclusions: `ireturn`, `recvcheck`, `contextcheck`, `exhaustive`                              | 4%    | Medium    | 20m    | quick-win | `config.go:75`                                     |
| C3  | **[EXPANDED]** Tune `funlen` 60/40→200/100 **+ reconcile this repo's own `.golangci.yml` (30/20) + update "upstream" comment** | 4%    | High      | 30m    | quick-win | `linter_settings.go:191`, `.golangci.yml:165`      |
| C4  | Add `GosecSettings` typed struct + curated excludes + injection + test                                                         | 4%    | High      | 60m    | quick-win | `linter_settings.go`, `fixer_config.go`            |
| C4b | **[NEW]** Add `ErrcheckSettings` typed struct + curated `exclude-functions` (`Close`, `Fprint*`) + test                        | 4%    | High      | 60m    | quick-win | `linter_settings.go`, `fixer_config.go`            |
| C5  | **[REWRITTEN]** `--pragmatic` flag: extend `shouldSkipLinter` with noise-skip set on the _dynamic_ enable set                  | 20%   | High      | 90m    | strategic | `categorizer.go:40`, CLI flag                      |
| C5b | **[NEW]** `gochecknoglobals` decision: keep in defaults, document in `--pragmatic` skip-set (no config knobs exist)            | 20%   | Medium    | 30m    | strategic | `AGENTS.md`, `categorizer.go`                      |
| C6  | Sidecar adoption decision: promote vs de-emphasize — implement chosen path                                                     | 20%   | Medium    | 45m    | strategic | `pkg/policy/`, `README.md`                         |
| C7  | Add `golangci-lint run` (no `--fix`) CI step (flake lint output + GH Actions)                                                  | 20%   | High      | 30m    | quick-win | `flake.nix`, `.github/workflows/`                  |
| C8  | Publish findings: update `FEATURES.md`, `AGENTS.md`, `CHANGELOG.md`                                                            | 20%   | Medium    | 60m    | strategic | docs                                               |
| C9  | Lock formatter quadruple as `house` formatter preset + test                                                                    | other | Medium    | 60m    | strategic | `presets.go`, `config.go:30`                       |
| C10 | Declare v1 maintenance-only in docs + `ROADMAP.md`                                                                             | other | Low       | 45m    | hygiene   | docs                                               |
| C11 | Tests for audit/policy code paths (`cmd_audit.go`, `fixer_enforce.go`, `newRunLedger`)                                         | other | High      | 90m    | hygiene   | `internal/cli/`, `pkg/linter/`                     |
| C12 | Exit-code integration tests for Infrastructure(69) + Corruption(65)                                                            | other | Medium    | 45m    | hygiene   | `internal/cli/exit_code_test.go`                   |
| C13 | Increase CLI integration test coverage (~11%)                                                                                  | other | Medium    | 60m    | ongoing   | `internal/cli/`                                    |
| C14 | Convert `scripts/coverage-check.sh` → Go test                                                                                  | other | Low       | 30m    | hygiene   | `scripts/`                                         |
| C15 | Adopt `HandleError` at CLI boundary (replaces slog)                                                                            | other | Medium    | 60m    | hygiene   | `internal/cli/`                                    |
| C16 | `funcorder` test gap closure                                                                                                   | other | Low       | 30m    | hygiene   | tests                                              |
| C17 | Register domain message templates with `errorfamily.New()`                                                                     | other | Low       | 30m    | hygiene   | `pkg/errors/classification.go`                     |
| C18 | **[NEW]** Validate: regenerate 5 sample sibling configs w/ new tool, re-measure nolints, write delta vs C0 baseline            | loop  | Very High | 60m    | strategic | `/tmp/friction.py`, sample repos                   |
| C19 | **[NEW]** Rollout: version bump, CHANGELOG, git tag, README note, announce new defaults                                        | loop  | High      | 45m    | strategic | `version`, `CHANGELOG.md`, `README.md`             |

**Totals:** 21 tasks · ~16.9h. Quick-wins (C1, C2, C3, C4, C4b, C7) = ~3.9h for the bulk of friction reduction. The measurement loop (C0+C18) = 1.5h and is what makes the result _provable_.

---

## 3. Fine Plan — tasks ≤ 12 min each

Each coarse task decomposed into atomic, independently-verifiable steps. Sorted within tier by dependency then impact.

| #   | Fine task                                                                                                               | Parent | Est | Files                              |
| --- | ----------------------------------------------------------------------------------------------------------------------- | ------ | --- | ---------------------------------- |
| F0a | Re-run `/tmp/friction.py`; freeze per-linter nolint table as `baseline.md`                                              | C0     | 10m | `docs/research/`                   |
| F0b | Snapshot 5 representative sibling `.golangci.yml` (1 tiny, 2 median, 2 large) as "before"                               | C0     | 10m | sample configs                     |
| F0c | Pick the 5 projects + record their _current_ nolint counts per target linter                                            | C0     | 10m | research                           |
| F1  | Enumerate stdlib structs to exclude (grep top-20 exhaustruct nolint targets across sibling repos)                       | C1     | 10m | research                           |
| F2  | Add 12–18 stdlib struct entries to `ExhaustructSettings.Exclude`                                                        | C1     | 8m  | `linter_settings.go:156`           |
| F3  | Add data-integrity test: `ExhaustructSettings.Exclude` contains required stdlib entries                                 | C1     | 12m | `data_integrity_test.go`           |
| F4  | Add unit test: generated config contains expanded exhaustruct exclude                                                   | C1     | 12m | `fixer_test.go`                    |
| F5  | Run `go test ./pkg/constants/... ./pkg/linter/...` — verify green                                                       | C1     | 5m  | —                                  |
| F6  | Add `gosec`, `errcheck`, `wrapcheck` to `DefaultExclusionRules[0].Linters` test-file list                               | C2     | 6m  | `config.go:75`                     |
| F7  | Update `DefaultExclusionRules` integrity test to assert the 3 high-friction linters excluded                            | C2     | 10m | `data_integrity_test.go`           |
| F8  | Run `go test ./pkg/constants/...` — verify green                                                                        | C2     | 5m  | —                                  |
| F9  | Add `ireturn`, `recvcheck`, `contextcheck`, `exhaustive` to test-file exclusion list                                    | C2b    | 8m  | `config.go:75`                     |
| F10 | Extend integrity test for the 4 low-friction linters                                                                    | C2b    | 8m  | `data_integrity_test.go`           |
| F11 | Change `FunlenSettings` default `Lines:60`→`200`, `Statements:40`→`100`                                                 | C3     | 5m  | `linter_settings.go:191`           |
| F12 | Update the `//nolint:mnd` comment: remove "matching upstream", state "house style (200/100); diverges from upstream"    | C3     | 4m  | `linter_settings.go:192`           |
| F13 | Fix this repo's own `.golangci.yml` funlen `30/20`→`200/100` (resolve `TODO_LIST.md` High item)                         | C3     | 5m  | `.golangci.yml:165`                |
| F14 | Grep sibling configs: confirm 200/100 is the dominant override (validate assumption)                                    | C3     | 10m | research                           |
| F15 | Update funlen default unit test assertion to 200/100                                                                    | C3     | 8m  | `linter_settings_internal_test.go` |
| F16 | Define `GosecSettings{Excludes []string}` struct + `ToMap()`                                                            | C4     | 10m | `linter_settings.go`               |
| F17 | Add `_ SettingsConverter = GosecSettings{}` compile-time check                                                          | C4     | 3m  | `linter_settings.go:34`            |
| F18 | Populate `DefaultLinterSettings["gosec"]` with curated excludes (G104, G304, …)                                         | C4     | 10m | `linter_settings.go:147`           |
| F19 | Add `GosecSettings` round-trip unit test (marshal→unmarshal)                                                            | C4     | 10m | `linter_settings_internal_test.go` |
| F20 | Add generated-config regression test: gosec excludes present when gosec enabled                                         | C4     | 12m | `fixer_test.go`                    |
| F21 | Define `ErrcheckSettings{ExcludeFunctions []string, CheckBlank bool, …}` + `ToMap()`                                    | C4b    | 10m | `linter_settings.go`               |
| F22 | Add `_ SettingsConverter = ErrcheckSettings{}` compile-time check                                                       | C4b    | 3m  | `linter_settings.go:34`            |
| F23 | Populate `DefaultLinterSettings["errcheck"]` with curated `exclude-functions` (`(io.Closer).Close`, `fmt.Fprint*`, …)   | C4b    | 10m | `linter_settings.go:147`           |
| F24 | Add `ErrcheckSettings` round-trip unit test                                                                             | C4b    | 10m | `linter_settings_internal_test.go` |
| F25 | Add generated-config regression test: errcheck exclude-functions present                                                | C4b    | 12m | `fixer_test.go`                    |
| F26 | Run `go test ./pkg/...` — verify all of C1–C4b green together                                                           | C4b    | 8m  | —                                  |
| F27 | **[v2 core]** Add `PragmaticNoiseLinters` const set `{exhaustruct, gochecknoglobals, ireturn, wrapcheck, funlen}`       | C5     | 8m  | `categorizer.go` / new const       |
| F28 | **[v2 core]** Extend `shouldSkipLinter` to skip `PragmaticNoiseLinters` when `--pragmatic` set (NOT when unset)         | C5     | 12m | `categorizer.go:40`                |
| F29 | Add `--pragmatic` bool flag to `configure` cmd (`cmd_configure.go`)                                                     | C5     | 10m | `internal/cli/cmd_configure.go`    |
| F30 | Add `shouldSkipLinter` unit tests: pragmatic on→skips 5, off→keeps all (default unchanged)                              | C5     | 12m | `categorizer_test.go`              |
| F31 | Add CLI acceptance test: `--pragmatic` produces config without the 5 noise linters                                      | C5     | 12m | CLI tests                          |
| F32 | Document `--pragmatic` in `FEATURES.md` + `--help` text + `AGENTS.md`                                                   | C5     | 10m | docs                               |
| F33 | **[v2 core]** Write `gochecknoglobals` rationale: friction 5.1, no config knobs, kept in defaults, in `--pragmatic` set | C5b    | 12m | `AGENTS.md`                        |
| F34 | Add data-integrity test: `PragmaticNoiseLinters` ⊆ known-linters and disjoint from `DisabledLinters`                    | C5b    | 10m | `data_integrity_test.go`           |
| F35 | Audit current sidecar usage: confirm 0 adoption, read `pkg/policy/` enforcement path                                    | C6     | 12m | `pkg/policy/`                      |
| F36 | Make sidecar decision (promote vs de-emphasize) — write 1-paragraph rationale                                           | C6     | 10m | planning                           |
| F37 | Implement chosen sidecar path: README section OR de-emphasize note in docs                                              | C6     | 12m | `README.md` / docs                 |
| F38 | Find lint output in `flake.nix`; locate where `--fix` is passed                                                         | C7     | 8m  | `flake.nix`                        |
| F39 | Add `golangci-lint run ./...` (no `--fix`) as separate `lint-check` output                                              | C7     | 12m | `flake.nix`                        |
| F40 | Add matching GH Actions step (if workflow exists)                                                                       | C7     | 10m | `.github/workflows/`               |
| F41 | Verify `nix build` + new lint output succeeds                                                                           | C7     | 10m | —                                  |
| F42 | Write `FEATURES.md` section: "Friction-driven defaults" with the data table                                             | C8     | 12m | `FEATURES.md`                      |
| F43 | Update `AGENTS.md` gotcha #7 with new exhaustruct/gosec/errcheck/funlen defaults                                        | C8     | 10m | `AGENTS.md`                        |
| F44 | Add `CHANGELOG.md` entry summarizing friction-reduction changes                                                         | C8     | 10m | `CHANGELOG.md`                     |
| F45 | Define `house` formatter preset in `PresetFormatters` = `{gci, gofumpt, goimports, golines}`                            | C9     | 8m  | `presets.go:56`                    |
| F46 | Add `house` to preset descriptions + integrity test                                                                     | C9     | 10m | `presets.go`, tests                |
| F47 | Add `CoreFormatters` alignment test (config.go:30 vs preset)                                                            | C9     | 10m | tests                              |
| F48 | Write v1 maintenance-only note in `ROADMAP.md`                                                                          | C10    | 10m | `ROADMAP.md`                       |
| F49 | Add v1 deprecation banner to migration docs                                                                             | C10    | 12m | `docs/references/`                 |
| F50 | Add `AGENTS.md` line: "v1 configs: maintenance-only, 0 live instances"                                                  | C10    | 5m  | `AGENTS.md`                        |
| F51 | Read `cmd_audit.go` + `fixer_enforce.go` to map untested branches                                                       | C11    | 12m | source                             |
| F52 | Write `cmd_audit_test.go`: ledger read/filter/clear happy path                                                          | C11    | 12m | `internal/cli/`                    |
| F53 | Write audit error-path test (corrupt JSONL, missing file)                                                               | C11    | 12m | `internal/cli/`                    |
| F54 | Write `fixer_enforce_test.go`: sidecar re-enable logic                                                                  | C11    | 12m | `pkg/linter/`                      |
| F55 | Write `newRunLedger` test: mutation recording + retention purge                                                         | C11    | 12m | `pkg/audit/`                       |
| F56 | Add Infrastructure(69) exit-code integration test                                                                       | C12    | 12m | `exit_code_test.go`                |
| F57 | Add Corruption(65) exit-code integration test                                                                           | C12    | 12m | `exit_code_test.go`                |
| F58 | Identify 3 lowest-covered CLI funcs (from coverage report)                                                              | C13    | 12m | coverage                           |
| F59 | Add integration tests for those 3 funcs                                                                                 | C13    | 12m | `internal/cli/`                    |
| F60 | Read `scripts/coverage-check.sh`; map logic to Go                                                                       | C14    | 8m  | `scripts/`                         |
| F61 | Write `coverage_check_test.go` replicating the threshold logic                                                          | C14    | 12m | tests                              |
| F62 | Locate CLI error-handling sites using raw `slog.Error`                                                                  | C15    | 10m | `internal/cli/`                    |
| F63 | Replace 3–5 sites with `HandleError`                                                                                    | C15    | 12m | `internal/cli/`                    |
| F64 | Write `funcorder` test closing flagged gap                                                                              | C16    | 12m | tests                              |
| F65 | Register domain message templates via `errorfamily.New()`                                                               | C17    | 12m | `classification.go`                |
| F66 | **[NEW]** Regenerate the 5 baseline sibling configs (F0b) with the new tool binary                                      | C18    | 12m | sample repos                       |
| F67 | **[NEW]** Re-run friction extraction on the 5 sample repos' _code_ against the regenerated configs                      | C18    | 12m | `/tmp/friction.py`                 |
| F68 | **[NEW]** Write `validation-delta.md`: per-linter nolint count before→after + % reduction                               | C18    | 12m | `docs/research/`                   |
| F69 | **[NEW]** If delta < target (DoD), diagnose which exclude is missing and loop back to C1/C4/C4b                         | C18    | 12m | —                                  |
| F70 | **[NEW]** Bump version (patch/minor per semver), update `CHANGELOG.md` release section                                  | C19    | 10m | `version`, `CHANGELOG.md`          |
| F71 | **[NEW]** Tag release; add README "what changed" callout for the friction-driven defaults                               | C19    | 12m | `README.md`                        |

**Totals:** 71 fine tasks · ~12h of atomic work (parallelizable across waves).

---

## 4. Execution Graph

```mermaid
flowchart TD
    RES["Research: 160 configs + 5201 nolints<br/>docs/research/2026-07-25_*"]
    BASE["C0: freeze baseline<br/>F0a–F0c"]

    subgraph P1["Phase 1 — 1% / 51%  (the gate)"]
        C1["C1: exhaustruct stdlib exclude<br/>F1–F5"]
    end
    subgraph P2["Phase 2 — 4% / 64%  (parallel after C1)"]
        C2["C2: high-friction test excl<br/>gosec/errcheck/wrapcheck<br/>F6–F8"]
        C2b["C2b: low-friction test excl<br/>F9–F10"]
        C3["C3: funlen 200/100 + repo reconcile<br/>F11–F15"]
        C4["C4: gosec curated excludes<br/>F16–F20"]
        C4b["C4b: errcheck exclude-functions<br/>F21–F25"]
    end
    subgraph P3["Phase 3 — 20% / 80%"]
        C5["C5: --pragmatic flag (dynamic skip)<br/>F27–F32"]
        C5b["C5b: gochecknoglobals rationale<br/>F33–F34"]
        C6["C6: sidecar decision<br/>F35–F37"]
        C7["C7: CI no-fix gate<br/>F38–F41"]
        C8["C8: publish findings<br/>F42–F44"]
    end
    subgraph P4["Phase 4 — other 20% / 100%"]
        C9["C9: house formatter preset<br/>F45–F47"]
        C10["C10: v1 maintenance-only<br/>F48–F50"]
        C11["C11: audit/policy tests<br/>F51–F55"]
        C12["C12: exit-code tests<br/>F56–F57"]
        C13["C13: CLI coverage<br/>F58–F59"]
        C14["C14: coverage-check→Go<br/>F60–F61"]
        C15["C15: HandleError<br/>F62–F63"]
        C16["C16: funcorder gap<br/>F64"]
        C17["C17: domain templates<br/>F65"]
    end
    subgraph P5["Phase 5 — Validate & Rollout  (NEW)"]
        C18["C18: validate delta vs baseline<br/>F66–F69"]
        C19["C19: version bump + tag + announce<br/>F70–F71"]
    end

    RES --> BASE
    BASE --> P1
    C1 --> C2
    C1 --> C4
    C1 --> C4b
    C2 & C2b & C3 & C4 & C4b --> C5
    C5 --> C5b
    C5 --> C8
    C6 --> C8
    C7 --> C8
    C8 --> P4
    C8 --> C18
    C18 -->|"delta ≥ target"| C19
    C18 -.->|"delta < target → fix excludes"| P2

    P1 --> VERIFY1["go test ./pkg/constants/... ./pkg/linter/..."]
    P2 --> VERIFY2["go test ./pkg/... + nix build"]
    P3 --> VERIFY3["nix flake check + lint"]
    P4 --> VERIFY4["full suite + coverage report"]

    VERIFY1 -.->|"if red, fix immediately"| P1
    VERIFY2 -.->|"if red, fix immediately"| P2

    classDef phase0 fill:#e2e8f0,stroke:#475569,color:#0f172a
    classDef phase1 fill:#dcfce7,stroke:#16a34a,color:#14532d
    classDef phase2 fill:#dbeafe,stroke:#2563eb,color:#1e3a8a
    classDef phase3 fill:#fef9c3,stroke:#ca8a04,color:#713f12
    classDef phase4 fill:#f3e8ff,stroke:#9333ea,color:#581c87
    classDef phase5 fill:#ffe4e6,stroke:#e11d48,color:#4c0519
    classDef verify fill:#fee2e2,stroke:#dc2626,color:#7f1d1d

    class BASE phase0
    class C1 phase1
    class C2,C2b,C3,C4,C4b phase2
    class C5,C5b,C6,C7,C8 phase3
    class C9,C10,C11,C12,C13,C14,C15,C16,C17 phase4
    class C18,C19 phase5
    class VERIFY1,VERIFY2,VERIFY3,VERIFY4 verify
```

**Sequencing rules:**

- **C0 (baseline) runs first** — it freezes the numbers C18 will compare against. No changes yet.
- **Phase 1 is the gate.** Nothing else starts until C1 is green — cheapest, highest-impact, and its exclude-list pattern is reused by C4/C4b.
- **Phase 2 is parallel _after C1_.** (v1 wrongly called it "parallel-safe" while drawing C1→C2/C4 edges — the edges mean C2/C4/C4b wait on C1. C3 and C2b are fully independent.) Different files (`config.go` vs `linter_settings.go`), no merge conflicts.
- **Phase 3 depends on Phase 2** being green — the `--pragmatic` flag's value is the tuned defaults from C1–C4b.
- **Phase 4 is independent** test-debt + docs; can run anytime.
- **Phase 5 closes the loop.** C18 regenerates sample configs and measures the delta vs C0. If the reduction misses the DoD target, **loop back** to Phase 2 to expand excludes — do NOT proceed to rollout.
- Every phase ends with verification (`go test` / `nix build` / `nix flake check`). Red → fix immediately, never carry broken state forward.

---

## 5. Risk & rollback

| Risk                                              | Mitigation                                                                                                                                                                                        |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Expanded exhaustruct exclude hides a real bug     | Exclude list is stdlib-only + project types added per-project (existing pattern). Data-integrity test pins the canonical list.                                                                    |
| funlen 200/100 too permissive                     | It's the _empirical_ house style (dominant override). **Diverges from golangci-lint upstream** — documented in the updated comment. Projects wanting stricter can set it. Reversible in one line. |
| gosec/errcheck excludes weaken real checks        | Excludes target test-path + known-benign codes/signatures only, via existing `DefaultExclusionRules` + `exclude-functions`. Production code still fully scanned.                                  |
| `--pragmatic` flag误用 (users enable by accident) | Off by default; `shouldSkipLinter` only skips the 5 when flag is explicitly set. Default behavior byte-identical to today.                                                                        |
| gochecknoglobals decision is wrong                | Decision is docs-only + opt-in flag membership. No code deletion, no default change. Fully reversible.                                                                                            |
| Validation delta misses target                    | C18 has an explicit loop-back edge to Phase 2. Rollout (C19) is _gated_ on meeting the quantitative DoD.                                                                                          |
| Sidecar decision is wrong                         | Decision is documented + reversible (docs-only change). No code deletion.                                                                                                                         |
| Tool changes don't propagate to 160 projects      | v2 adds C19 (rollout: version bump + tag + announce). Sample re-gen in C18 proves propagation works.                                                                                              |
| Auto-commit daemon mid-execution                  | Expected behavior (per `AGENTS.md`). We commit our own logical units with detailed messages; daemon fills gaps.                                                                                   |

**Every change is reversible** (config-map edits, additive flag, additive presets, docs). No irreversible operations. No `rm`, no `git reset`, no force-push.

---

## 6. Definition of Done (now quantitative)

- [ ] Phase 0–3 complete; `go test -race ./pkg/... ./internal/...` green
- [ ] `nix build` + `nix flake check` green; `golangci-lint run` clean on this repo
- [ ] **Quantitative (C18):** regenerated 5 sample configs yield, vs C0 baseline:
  - [ ] ≥ **40%** fewer `exhaustruct` nolints
  - [ ] ≥ **25%** fewer `gosec` nolints
  - [ ] ≥ **20%** fewer `errcheck` nolints
  - [ ] net `funlen` nolint reduction (200/100 default)
- [ ] `--pragmatic` flag works: default behavior unchanged; with flag, the 5 noise linters absent from generated config
- [ ] `FEATURES.md`, `AGENTS.md`, `CHANGELOG.md` reflect the new friction-driven defaults
- [ ] Ecosystem report annotated with "actions taken" (non-destructive, per `update-old-docs`)
- [ ] Version bumped + tagged release (C19); committed with detailed messages

---

## 7. Revision history — v1 → v2 changelog

A critical re-review of v1 against the **actual codebase** (verified by reading `presets.go`, `config.go`, `linter_settings.go`, `categorizer.go`, `rules.go`, and this repo's own `.golangci.yml`) found these defects. All fixed in v2:

| # | v1 defect                                                                                                                                                                       | v2 fix                                                                                                                                                                           |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **C5 was factually impossible.** "pragmatic = `strict` minus 5 linters" — but `exhaustruct`, `gochecknoglobals`, `ireturn`, `wrapcheck` are not in `strict` (only `funlen` is). | **C5 rewritten**: `--pragmatic` flag extends `shouldSkipLinter` on the _dynamic_ enable set. Verified the ~109 set is runtime-computed (`categorizer.go:21`), not a static list. |
| 2 | **`gochecknoglobals` (#2 friction, 5.1) had zero direct task.** The plan treated #1, #3, #4 but silently skipped #2.                                                            | **C5b added**: explicit decision (keep in defaults — it has _no config knobs_ per golangci-lint docs; route through `--pragmatic`).                                              |
| 3 | **Pareto asymmetry:** `gosec` (3.8) got a settings struct; `errcheck` (3.1, supports `exclude-functions`) got only test-exclusion.                                              | **C4b added**: `ErrcheckSettings` with curated `exclude-functions` (`Close`, `Fprint*`).                                                                                         |
| 4 | **No measurement loop.** "Reduce friction" was unverifiable — no before/after numbers.                                                                                          | **C0 (baseline) + C18 (validate)** added with a loop-back edge if the delta misses target. **DoD made quantitative.**                                                            |
| 5 | **funlen 3-way mismatch ignored.** Tool injects `60/40`; this repo's own `.golangci.yml` uses `30/20`; house style `200/100`. v1 only did 60→200.                               | **C3 expanded**: 60→200, fix repo's `30/20`, and update the misleading "matching upstream" comment (`linter_settings.go:192`).                                                   |
| 6 | **No rollout.** Tool changes don't auto-propagate to 160 sibling projects.                                                                                                      | **C19 added**: version bump + tag + announce.                                                                                                                                    |
| 7 | **Graph contradicted text.** "Phase 2 parallel-safe" but edges `C1→C2`,`C1→C4` mean C2/C4 wait on C1.                                                                           | Graph redrawn: C3 & C2b are the only fully-independent Phase-2 tasks; C2/C4/C4b depend on C1.                                                                                    |
| 8 | **C2 spread thin** across 7 linters mixing friction 3.8 (gosec) with 0.84 (exhaustive).                                                                                         | **Split into C2** (gosec/errcheck/wrapcheck, the >1.5 trio) **+ C2b** (the <1.3 remainder).                                                                                      |
| 9 | **`ireturn` already configured** (`Allow:[error,empty,anon,stdlib,generic]` at `linter_settings.go:148`) — pragmatic removing it is low-value.                                  | Noted in the friction table; `ireturn` stays in `--pragmatic` set but flagged as already-well-handled.                                                                           |

v1 is preserved in git at `62baa44`.

---

## Execution resolution (2026-07-25)

This plan was executed the same day it was proposed. Task-by-task outcome:

| Task   | Outcome | Notes                                                                                                            |
| ------ | ------- | ---------------------------------------------------------------------------------------------------------------- |
| C0     | ✅ Done | Baseline frozen in `docs/research/baseline.md`                                                                   |
| C1     | ✅ Done | `ExhaustructSettings.Exclude` expanded to 14 stdlib structs + data-integrity test                                |
| C2/C2b | ✅ Done | High + low friction test-file exclusions added to `DefaultExclusionRules[0]` (now 14 linters)                    |
| C3     | ✅ Done | `funlen` → 200/100 (house style); repo's own `.golangci.yml` regenerated via the tool                            |
| C4     | ✅ Done | `GosecSettings` typed struct; G104 later removed (too broad — see C4 note)                                       |
| C4b    | ✅ Done | `ErrcheckSettings` with curated `exclude-functions`                                                              |
| C5/C5b | ✅ Done | `--pragmatic` flag + `PragmaticNoiseLinters` skip-set; gochecknoglobals kept in defaults                         |
| C6     | ✅ Done | Sidecar de-emphasized (ROADMAP non-goal; README note added)                                                      |
| C7     | ✅ Done | CI already runs `golangci-lint` without `--fix` (verified)                                                       |
| C8     | ✅ Done | Findings published in FEATURES/AGENTS/CHANGELOG                                                                  |
| C9     | ✅ Done | `house` formatter preset locked (4 formatters) + alignment test                                                  |
| C10    | ✅ Done | v1 declared maintenance-only in ROADMAP + AGENTS                                                                 |
| C11    | ✅ Done | Audit/policy tests written (`cmd_audit_test.go`, `fixer_enforce_test.go`)                                        |
| C12    | ✅ Done | Exit-code integration tests for Infrastructure(69) + Corruption(65)                                              |
| C13    | ✅ Done | CLI coverage raised (~11% → 27.9%)                                                                               |
| C14    | ✅ Done | `scripts/coverage-check.sh` → `cmd/coverage-check` (Go); old script deleted                                      |
| C15    | ✅ Done | `HandleError` adopted at CLI boundary                                                                            |
| C16    | ⚠️ Open  | `funcorder` test gap — not closed this cycle                                                                     |
| C17    | ✅ Done | 27 domain message templates registered via `errorfamily.New()`                                                   |
| C18    | ✅ Done | `docs/research/validation-delta.md` written (errcheck 27.3%, gosec 23.7%, exhaustruct 2.3% — target unrealistic) |
| C19    | ❌ Open | Version bump + tag NOT done — requires semver decision. Now in `TODO_LIST.md` (High Priority)                    |

**Caveat carried forward:** the `RuleKey()` dedup means new default-exclusion
linters only reach **new/regenerated** configs; 88 machine-generated sibling
configs keep the old list until their rule is re-injected (see ROADMAP "Config
propagation"). The validation delta (C18) was measured via grep estimates, not
full config regeneration — a proper re-measurement remains open.
