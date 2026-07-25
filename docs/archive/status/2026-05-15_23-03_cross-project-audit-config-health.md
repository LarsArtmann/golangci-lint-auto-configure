# Status Report: Cross-Project Audit + ConfigHealth + gogenfilter

**Date:** 2026-05-15 23:03
**Session Scope:** Cross-project golangci-lint config audit, ConfigHealth validation feature, gogenfilter integration
**Status:** Session complete, all work pushed to master

---

## A) FULLY DONE

### 1. Cross-Project golangci-lint Configuration Audit

**Commit:** `055cdab` | **File:** `docs/cross-project-golangci-lint-audit-report.md` (635 lines)

Analyzed **126 `.golangci.yml`/`.golangci.yaml`** files across all managed Go projects. Full content analysis + git history review + `golangci-lint-auto-configure analyze` runs.

**Key findings:**

- 8 configs have **broken YAML with duplicate linter entries** (GmbH worst: 56 duplicates + orphaned enable block)
- 4 configs mix **v1/v2 syntax** (`linters-settings:` alongside `version: "2"`)
- 14 configs are **missing critical linters** (errcheck, staticcheck, govet, or gosec)
- 33 configs use **both enable+disable** — a v2 anti-pattern
- ~60% are cookie-cutter 109-linter configs with zero project-specific tuning
- 2 configs exceed 800 lines of over-configuration (desire-secrets: 881, complaints-mcp: 822)
- All 126 configs are v2 format — no v1 stragglers remain

**Report sections:** Executive summary, 14 analytical sections covering structural problems, anti-patterns, missing critical linters, config size analysis, cluster analysis, formatter configuration, good patterns, bad patterns, recommended template, project-by-project severity assessment, and action items.

### 2. ConfigHealth Structural Validation Feature

**Commit:** `e74e6bd` | **Files:** `pkg/types/validation.go`, `pkg/types/health_test.go`, `internal/cli/cmd_validate.go`

Added `ConfigHealth` type with 4 detection rules wired into the `validate` command:

| Rule                      | Severity | What It Detects                     | Projects Affected |
| ------------------------- | -------- | ----------------------------------- | ----------------- |
| `duplicate-linter`        | CRITICAL | Duplicate entries in enable/disable | 8 projects        |
| `enable-disable-overlap`  | WARNING  | Same linter in both lists           | 33 projects       |
| `missing-critical-linter` | WARNING  | No errcheck/staticcheck/govet       | 14 projects       |
| `v1-syntax-in-v2`         | WARNING  | `linters-settings:` in v2 config    | 4 projects        |

**Architecture:**

- `ConfigHealth` struct with `IsHealthy()`, `CriticalIssues()`, `WarningIssues()` methods
- `CheckConfigHealth(cfg *Config)` — pure function, zero external dependencies
- `HealthIssue` struct with severity, rule, message, field, suggestion
- SARIF output via `healthIssuesToFindings()` in validate command
- 18 BDD test specs — all passing

### 3. gogenfilter Auto-Generated File Exclusion Scanning

**Commits:** `06f0491`, `b41f26e`, `305e125`, `344d354`, `7df3a30`, `00f0ed0`, `9ee558a`, `b1b2dd3`, `c078da0`
**Files:** `pkg/gogenfilter/scanner.go` (305 lines), `pkg/gogenfilter/scanner_test.go` (250 lines), `pkg/linter/fixer.go`, `pkg/linter/fixer_config.go`

Integrates `gogenfilter/v3` library to detect auto-generated Go files and inject exclusion patterns during `configure`. Detects: templ, protobuf, sqlc, wire, mockgen, moq, go-enum, stringer, deepcopy-gen, oapi-codegen, and generic codegen.

**Key decisions:**

- `ScanProject` accepts `fs.FS` interface for testability
- Returns individual regex path patterns (not pipe-delimited globs)
- `MergeExclusionPaths` handles deduplication and sorting
- `generated` count tracked in `fixCounts` struct
- Scan runs even when no other fixes exist (standalone value)
- Added `golines` settings with `max-len: 140` and regex pattern for generated templ files in own config

### 4. Audit Report Updated with Section 15

**Commit:** `435f961`

Added documentation linking audit findings to the new automated detection rules, with usage examples and architecture references.

### 5. Test Suite

**All 14 test suites passing, 251 total specs, composite coverage: 59.7%**

| Suite               | Specs | Coverage |
| ------------------- | ----- | -------- |
| CLI Commands        | 23    | 8.5%     |
| Config              | 37    | 65.9%    |
| Experiments         | 6     | 100.0%   |
| Errors              | 20    | 100.0%   |
| GoGenFilter Scanner | 15    | 62.9%    |
| Analyzer            | 39    | 78.4%    |
| Migration           | 37    | 66.8%    |
| Set                 | 38    | 57.7%    |
| Utils               | 16    | 94.6%    |
| Version             | 6     | 51.4%    |
| Differ              | 12    | 65.0%    |
| Linter Fixer        | 14    | 96.5%    |
| Detection           | 7     | 58.1%    |
| Categorizer         | 1     | —        |

---

## B) PARTIALLY DONE

### 1. Auto-fix for Duplicate Linters

The `validate` command now **detects** duplicate linters but the `configure` command does not **auto-fix** them. The detection is in place; the auto-fix step is not implemented. This would be a small change in `fixer_config.go` — deduplicating the enable list before writing it back.

**What's done:** Detection via `CheckConfigHealth` (CRITICAL severity)
**What's missing:** Auto-deduplication in `updateConfigFromSets()` or a pre-flight fix

### 2. Portfolio-Wide Validation Run

We identified all 126 configs and their issues, but did not run `golangci-lint-auto-configure validate --format sarif` across all projects to generate a portfolio-wide health report. The tooling is ready but the batch execution script was not created.

**What's done:** Tool supports health checks with SARIF output
**What's missing:** A script/CI step that iterates over all projects and aggregates results

### 3. Recommended Standard Template Adoption

Section 12 of the audit report defines a recommended standard config template (50 linters vs the current 110). This template was not implemented as a preset or enforced anywhere.

**What's done:** Template designed and documented
**What's missing:** Implementation as a `preset: recommended` option in the tool

---

## C) NOT STARTED

1. **`overly-long-config` health rule** — Warn when config exceeds 200 lines (desire-secrets: 881 lines)
2. **`enable-all-anti-pattern` health rule** — Warn on `enable-all: true` in any linter settings (16 projects)
3. **`irrelevant-linter` health rule** — Warn when framework-specific linters (arangolint, sqlclosecheck, zerologlint) are enabled without the framework present
4. **`inconsistent-go-version` health rule** — Warn when Go version is pinned to a patch (1.26.1 vs 1.26) or unset
5. **`unbounded-issues` health rule** — Warn when `max-issues-per-linter: 0` (3 projects)
6. **Auto-fix for enable+disable overlap** — Move overlapping linters from disable to enable, or remove disable list entirely
7. **Portfolio-wide batch validation script** — Script that runs `validate` across all 126 projects
8. **`preset: recommended` implementation** — Implement the Section 12 template as a preset
9. **Project-type-aware linter suggestions** — Use `pkg/detection` to suggest removing irrelevant linters based on project type
10. **Config size reduction** — Strip default values from configs over 200 lines
11. **4-space indentation normalizer** — Auto-fix 4-space indented configs to 2-space
12. **CI integration guide** — Document how to add `golangci-lint-auto-configure validate` to GitHub Actions
13. **Cross-project config drift detection** — Alert when a project's config diverges significantly from its cluster baseline
14. **`golangci-lint-auto-configure fix --health` command** — One-command fix for all detected health issues
15. **Health check integration in `configure` command** — Run health checks before/after configure and report remaining issues
16. **AGENTS.md update for new features** — Document ConfigHealth, gogenfilter, and audit findings
17. **Coverage improvement for CLI commands** — cmd_validate.go at 8.5% coverage needs integration tests for health checks
18. **Performance benchmarks for gogenfilter** — No benchmarks for the scanner on large projects
19. **SARIF integration test** — Verify SARIF output format is valid with a golden file test
20. **Nix flake integration for gogenfilter** — Ensure `vendorHash` is correct and `nix build` passes

---

## D) TOTALLY FUCKED UP

### Nothing is broken.

- Build: **OK**
- All 14 test suites: **PASSING** (251 specs, 0 failures)
- `git status`: **clean**
- All commits: **pushed to origin/master**
- No regressions introduced

### Near-misses (caught and fixed):

1. **Test config missing critical linters** — Initial `validate` test config only had `errcheck` + `gosec`, but `CheckConfigHealth` correctly flagged missing `staticcheck` + `govet`. Fixed by adding all 4 critical linters to the test config.
2. **gogenfilter glob patterns** — First implementation used pipe-delimited glob patterns (`dir1/**|dir2/**`) which are not valid YAML glob syntax. Fixed to return individual regex-compatible patterns.
3. **gogenfilter hardcoded `os.DirFS`** — First implementation accepted a `string` path directly, making it untestable without real filesystem. Refactored to accept `fs.FS` interface.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **`CheckConfigHealth` should accept a configurable set of "critical linters"** — Currently hardcoded `[errcheck, staticcheck, govet]`. Should come from `constants.LinterPriorities` filtered by `LinterPriorityCritical` or a dedicated `constants.RequiredLinters` set.

2. **`fixCounts.generated` naming is confusing** — It counts the number of exclusion path patterns added, not the number of generated files. Rename to `generatedExclusions` or document better.

3. **`HealthSeverity` duplicates `finding.Severity`** — The mapping from `HealthSeverity` to `finding.Severity` in `healthIssuesToFindings()` is a manual switch statement. Consider using `finding.Severity` directly or providing a `ToFindingSeverity()` method.

4. **gogenfilter scanner should support context cancellation** — `ScanProject` does a potentially long filesystem walk with no context support. For large monorepos this could block.

5. **ConfigHealth rules should be extensible** — Currently rules are private methods on `ConfigHealth`. Consider a `HealthRule` interface so users/projects can register custom rules.

### Code Quality

6. **CLI validate command at 8.5% coverage** — The new health check code in `cmd_validate.go` (SARIF output, issue formatting) has no integration tests. Should add at least basic tests for `checkConfigHealth` and `logHealthIssues`.

7. **gogenfilter scanner at 62.9% coverage** — Missing coverage for `sqlcPatterns` with `GetSQLOutputDirs`, `oapiPatterns` with `.gen.go` files, and error paths.

8. **`healthIssuesToFindings` silently skips build errors** — Uses `continue` on `Build()` error. Should at minimum log the error.

### Process

9. **No CI for golangci-lint-auto-configure itself** — The tool that validates other configs doesn't validate its own config's health in CI. Should add a step to `.github/workflows/ci.yml`.

10. **No CHANGELOG entries** — Three significant features shipped without changelog entries.

---

## F) Top 25 Things We Should Get Done Next

Sorted by impact/effort ratio (highest first):

| #   | Task                                                                            | Impact | Effort | Type        |
| --- | ------------------------------------------------------------------------------- | ------ | ------ | ----------- |
| 1   | Auto-deduplicate enable list in `updateConfigFromSets()`                        | High   | Low    | Bug fix     |
| 2   | Derive critical linters from `constants.LinterPriorities` instead of hardcoding | Medium | Low    | Refactor    |
| 3   | Add integration test for `validate` command health checks                       | High   | Low    | Testing     |
| 4   | Add `overly-long-config` health rule (>200 lines)                               | Medium | Low    | Feature     |
| 5   | Add `enable-all-anti-pattern` health rule                                       | Medium | Low    | Feature     |
| 6   | Add `inconsistent-go-version` health rule                                       | Medium | Low    | Feature     |
| 7   | Run `golangci-lint-auto-configure validate` across all 126 projects             | High   | Low    | Audit       |
| 8   | Implement `preset: recommended` from audit Section 12                           | High   | Medium | Feature     |
| 9   | Add `irrelevant-linter` health rule using `pkg/detection`                       | Medium | Medium | Feature     |
| 10  | Auto-fix enable+disable overlap in `configure`                                  | Medium | Low    | Feature     |
| 11  | Add `unbounded-issues` health rule (max-issues-per-linter: 0)                   | Low    | Low    | Feature     |
| 12  | Add context support to gogenfilter scanner                                      | Medium | Medium | Improvement |
| 13  | Add `ToFindingSeverity()` method on `HealthSeverity`                            | Low    | Low    | Refactor    |
| 14  | Log errors in `healthIssuesToFindings` instead of silent skip                   | Low    | Low    | Bug fix     |
| 15  | Improve gogenfilter scanner test coverage to >80%                               | Medium | Medium | Testing     |
| 16  | Add CI step to validate tool's own config with health checks                    | Medium | Low    | CI/CD       |
| 17  | Add `golangci-lint-auto-configure fix --health` command                         | High   | Medium | Feature     |
| 18  | Create project-type-aware linter suggestion layer                               | High   | High   | Feature     |
| 19  | Normalize 4-space indentation to 2-space during `configure`                     | Low    | Low    | Feature     |
| 20  | Strip default values from configs during `configure`                            | Medium | Medium | Feature     |
| 21  | Add SARIF golden file integration test                                          | Medium | Medium | Testing     |
| 22  | Update AGENTS.md with ConfigHealth, gogenfilter, audit findings                 | Low    | Low    | Docs        |
| 23  | Add performance benchmarks for gogenfilter on large projects                    | Low    | Medium | Testing     |
| 24  | Build cross-project config drift detection                                      | High   | High   | Feature     |
| 25  | Update vendorHash and verify `nix build` after go.mod changes                   | Medium | Low    | Build       |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should `golangci-lint-auto-configure fix --health` auto-fix health issues (remove duplicates, remove enable+disable overlaps, add missing critical linters) WITHOUT running the full `golangci-lint configure` flow (which requires `golangci-lint` binary + version check + full analysis)?**

The tension: `configure` requires golangci-lint installed and running in a git repo. But `CheckConfigHealth` is a pure function that needs only the config file. A `fix --health` mode that only does structural fixes (dedup, overlap removal, version normalization) would be:

- Much faster (no external binary needed)
- Runnable on any config file (no git repo needed)
- Safe to run in CI on config PRs

But it would create a second code path for config modification that diverges from the main `configure` flow. Is this the right architectural boundary, or should health fixes always be part of the full `configure` pipeline?

---

## Metrics Summary

| Metric                  | Value                           |
| ----------------------- | ------------------------------- |
| Commits this session    | 11                              |
| Files changed           | 17                              |
| Lines added             | 2,037                           |
| Lines removed           | 18                              |
| New test specs          | 33 (18 health + 15 gogenfilter) |
| Test suites             | 14/14 passing                   |
| Composite coverage      | 59.7%                           |
| Configs audited         | 126                             |
| Health rules added      | 4                               |
| Source files in project | 68                              |
| Test files in project   | 25                              |

---

## Git Log (Session Commits)

```
b1b2dd3 refactor(gogenfilter): accept fs.FS interface in ScanProject
9ee558a chore: exclude generated templ files from linting in own config
8cc9bc1 fix(gogenfilter): produce regex patterns instead of invalid globs
00f0ed0 fix(nix): update vendorHash for gogenfilter dependency
c078da0 fix(preset): apply generated exclusions in preset flow
b41f26e fix(fixer): include generated count in success message, simplify newFixCounts
305e125 fix(fixer): always run generated exclusion scan, even when no other fixes exist
7df3a30 fix(gogenfilter): return individual path entries instead of pipe-delimited patterns
344d354 docs(status): add comprehensive gogenfilter integration review
435f961 docs: add Section 15 to audit report for automated health checks
06f0491 feat(fixer): add auto-generated file exclusion scanning via gogenfilter
e74e6bd feat(validate): add ConfigHealth structural checks for config quality
055cdab docs: add cross-project golangci-lint configuration audit report
```
