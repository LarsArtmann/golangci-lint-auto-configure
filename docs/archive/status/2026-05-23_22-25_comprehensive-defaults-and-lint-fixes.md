# Status Report — 2026-05-23 Session 2

**Generated:** 2026-05-23 22:25 UTC
**Branch:** master (2 commits ahead of origin)
**Last Release:** v0.2.0 (3f2681b)
**Commits Since Release:** 2

---

## Session Summary

Two feature commits landed today, adding comprehensive default settings and
exclusion automation that bring the tool's output in line with the reference
config in project-dependency-graph.

---

## A) FULLY DONE

### Commit 1: `7723dd6` — Default exclusion paths

Added unconditional injection of `_templ\.go$`, `vendor/` into
`linters.exclusions.paths` and `_templ\.go$` into
`formatters.exclusions.paths`. Also fixed pre-existing golines lint in
`loader.go`.

### Commit 2: `df8927c` — Comprehensive default settings + lint fixes

| Category           | Item                   | Detail                                                                                                                       |
| ------------------ | ---------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Linter defaults    | `revive`               | Disable `exported`, `package-comments`                                                                                       |
| Linter defaults    | `varnamelen`           | `ignore-map-index-ok`, `ignore-type-assert-ok`, 11 common short names                                                        |
| Linter defaults    | `gomoddirectives`      | `replace-local: true`                                                                                                        |
| Linter defaults    | `cyclop`               | `max-complexity: 12`                                                                                                         |
| Formatter defaults | `golines`              | `max-len: 120`                                                                                                               |
| Exclusion paths    | `.gen.go$`             | Added to default linter exclusion paths                                                                                      |
| Exclusion rules    | test files             | `linters.exclusions.rules` for `_test.go`: exhaustruct, testpackage, gochecknoglobals, funlen, cyclop, goconst, unused(text) |
| Lint fix           | `analyzer.go`          | Extracted `summaryParts()` — GetSummary 31→19 lines                                                                          |
| Lint fix           | `finding_formatter.go` | Extracted `writeFindingGroup()` — FormatFindings 31→17 lines                                                                 |
| Lint fix           | `finding_formatter.go` | Renamed loop var to fix varnamelen                                                                                           |

### Previously Done (v0.2.0 and earlier)

- 7 CLI commands (configure, analyze, validate, report, migrate, install-hook, completion)
- 119 linter priorities with reasons
- gogenfilter integration for generated file exclusion
- v1 to v2 config migration
- Deprecated linter auto-replacement (wsl, gomodguard, deadcode, etc.)
- Version-gated deprecation (gomodguard_v2 needs v2.12.0+)
- go-finding integration (SARIF, JSON, unified model)
- HTML report generation (templ-based)
- Build tags auto-injection (GOEXPERIMENT flags)
- Depguard/ireturn/gocritic/exhaustruct defaults
- Semantic deduplication to zero clones
- Nix flake build + justfile recipes
- CI (Go 1.25/1.26 matrix, lint, coverage)
- Pre-commit hook support

---

## B) PARTIALLY DONE

| Area                    | What's Done                  | What's Missing                                                                                                                   |
| ----------------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| CLI integration tests   | `integration_test.go` exists | Coverage 9.0% — most command paths untested                                                                                      |
| Report generation       | HTML + JSON done             | `pkg/report` 0.0% coverage                                                                                                       |
| Default config creation | `CreateDefaultConfig` exists | No exclusion paths/rules/settings in default config                                                                              |
| gocritic defaults       | 4 disabled checks            | `dupImport`, `octalLiteral`, `whyNoLint` are already disabled by golangci-lint (warnings in lint output) — should remove these 3 |

---

## C) NOT STARTED

1. **FEATURES.md** — No feature inventory file exists
2. **TODO_LIST.md** — No consolidated task list
3. **`gocritic` cleanup** — 3 disabled-checks (`dupImport`, `octalLiteral`, `whyNoLint`) are already disabled by default in golangci-lint, causing warnings. Should remove them.
4. **Default config enrichment** — `CreateDefaultConfig()` in `loader.go` doesn't include exclusion paths, rules, or linter/formatter settings. Fresh configs miss all the good stuff.
5. **CLI `--check` mode** — Exit 1 if config needs changes (CI/CD gate)
6. **Example config updates** — `examples/` doesn't reflect current defaults
7. **Coverage improvements** — `pkg/report` (0%), `internal/cli` (9%), `pkg/finding` (56%), `pkg/gogenfilter` (59.8%) all need more tests
8. **AGENTS.md length** — 912 lines, go-structure-linter says max 377
9. **coverage.out location** — Should be in `/coverage/` directory
10. **report_templ.go committed** — Generated file in git
11. **Binaries in repo** — `bin/golangci-lint-auto-configure`, `golangci-lint-auto-configure`, `result` tracked
12. **gitleaks false positives** — `reports/gosec.md` triggers API key patterns (docs, not secrets)
13. **charmbracelet/fang v1** — Library policy says use v2 (`charm.land/fang/v2`)

---

## D) TOTALLY FUCKED UP

**Nothing is broken.** All tests pass, all lints pass (0 issues), working tree
clean.

### Pre-existing buildflow hook failures (NOT in our code):

| Hook Step             | Issue                                                       | Severity                   |
| --------------------- | ----------------------------------------------------------- | -------------------------- |
| `todo-check`          | 2 TODO/NOTE comments in codebase                            | Info-level, not actionable |
| `gitleaks`            | False positives in `reports/gosec.md` (linter docs)         | False positive             |
| `go-structure-linter` | AGENTS.md too long, binaries in repo, coverage.out location | Pre-existing               |
| `library-policy`      | fang v1 → v2 migration                                      | Pre-existing               |

---

## E) WHAT WE SHOULD IMPROVE

### Critical

1. **`gocritic` disabled-checks cleanup** — Remove `dupImport`, `octalLiteral`, `whyNoLint` from defaults (golangci-lint already disables them, causing noisy warnings on every lint run)

2. **Default config enrichment** — `CreateDefaultConfig()` should inject exclusion paths, rules, and default settings so fresh configs are production-ready

3. **CLI test coverage at 9.0%** — The CLI wiring layer is where most bugs hide

### Medium

4. **AGENTS.md trim** — 912 lines is 535 over limit; extract detailed docs to referenced files
5. **FEATURES.md** — No feature inventory exists
6. **TODO_LIST.md** — No consolidated task list
7. **Report package at 0.0%** — Completely untested

### Nice to Have

8. **Example configs** — Stale, don't reflect current defaults
9. **`--check` CI mode** — Would enable `golangci-lint-auto-configure` as a CI gate
10. **Benchmarking** — Only one bench test

---

## F) TOP #25 THINGS TO DO NEXT

| #  | Priority | Task                                                                     | Impact                         |
| -- | -------- | ------------------------------------------------------------------------ | ------------------------------ |
| 1  | Critical | Remove 3 already-disabled gocritic checks from defaults (stop warnings)  | Clean lint output              |
| 2  | Critical | Enrich `CreateDefaultConfig()` with exclusion paths, rules, and settings | Fresh configs production-ready |
| 3  | Critical | Increase CLI integration test coverage (9.0% → 50%+)                     | Trust in command wiring        |
| 4  | High     | Add report package tests (0.0% → 50%+)                                   | HTML output untested           |
| 5  | High     | Trim AGENTS.md from 912 to ≤377 lines                                    | Fast AI context loading        |
| 6  | High     | Create `FEATURES.md` with full feature audit                             | Feature tracking               |
| 7  | High     | Create `TODO_LIST.md` from all status reports                            | Consolidated backlog           |
| 8  | High     | Increase gogenfilter scanner coverage (59.8% → 80%+)                     | Error path coverage            |
| 9  | High     | Increase migration coverage (66.8% → 80%+)                               | Edge case coverage             |
| 10 | Medium   | Add `--check` mode for CI (exit 1 if config needs changes)               | CI/CD gate                     |
| 11 | Medium   | Update example configs to reflect current defaults                       | Accurate examples              |
| 12 | Medium   | Add `.gitleaks.toml` to allowlist `reports/` false positives             | Clean hook                     |
| 13 | Medium   | Migrate `charmbracelet/fang` v1 → v2 (`charm.land/fang/v2`)              | Library policy                 |
| 14 | Medium   | Remove `report_templ.go` from git tracking                               | Generated file hygiene         |
| 15 | Medium   | Move `coverage.out` to `coverage/` directory                             | Structure lint                 |
| 16 | Medium   | Remove tracked binaries from git                                         | Structure lint                 |
| 17 | Medium   | Add `output.formats: {}` to default config creation                      | Explicit config                |
| 18 | Low      | Add `swaggo` formatter detection improvements                            | Better auto-detect             |
| 19 | Low      | Add benchmarking for analyzer and fixer                                  | Perf regression                |
| 20 | Low      | Document exclusion pattern syntax (RE2 regex) in README                  | User education                 |
| 21 | Low      | Add `ginkgolinter` default settings if any exist                         | Already High priority          |
| 22 | Low      | Consider adding `testifylint` default settings                           | Test quality                   |
| 23 | Low      | Add `--diff` flag to show config changes before applying                 | User safety                    |
| 24 | Low      | Add preset to apply reference config (project-dependency-graph pattern)  | One-shot setup                 |
| 25 | Low      | Vendor `vendor/` in formatter exclusions too? (open question)            | Consistency                    |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should `CreateDefaultConfig()` in `loader.go` call the same
`updateGeneratedExclusions`, `updateExclusionRules`, and
`injectDefaultSettings` pipeline that the fixer uses?**

Currently `CreateDefaultConfig()` creates a bare config with just linters
enabled and basic issues settings. It doesn't inject any of the new defaults
(exclusion paths, rules, linter/formatter settings). But the fixer flow
(`applyAndSave`) does inject them.

This means:

- `golangci-lint-auto-configure configure` → gets full defaults (via fixer)
- `golangci-lint-auto-configure configure` on a **new project** with no config → `CreateDefaultConfig()` runs first → then fixer runs → defaults applied

So in practice the fixer pipeline handles it. But if someone calls
`CreateDefaultConfig()` directly (e.g. programmatic API), they get a bare
config. The question is: should `CreateDefaultConfig()` be enriched
independently, or is the two-step flow (create → fix) the intended design?

---

## Test & Lint Summary

```
Ginkgo: 14/14 suites PASSED
Lint:   0 issues
Coverage: 61.2% composite (up from 60.6%)

Key packages:
  pkg/errors      — 100.0%
  pkg/diff        —  96.5%
  pkg/utils       —  94.6%
  pkg/constants   —  80.0%  (was 80.0%)
  pkg/linter      —  81.5%  (was 79.5%  ↑ +2.0)
  pkg/ui          —  67.7%  (was 67.7%)
  pkg/config      —  65.9%
  pkg/migration   —  66.8%
  pkg/detection   —  65.5%
  pkg/gogenfilter —  59.8%
  pkg/finding     —  56.0%
  pkg/types       —  59.4%
  pkg/version     —  51.4%
  internal/cli    —   9.0%
  pkg/report      —   0.0%
```

---

## Commits This Session

| Hash      | Message                                                                                      |
| --------- | -------------------------------------------------------------------------------------------- |
| `df8927c` | feat(defaults): add comprehensive default settings, exclusion rules, and fix lint violations |
| `7723dd6` | feat(exclusions): add default exclusion paths for \_templ.go and vendor/                     |

## Files Changed Since v0.2.0

```
10 files changed, 656 insertions(+), 73 deletions(-)
```

| File                          | Change                                                                                                                                          |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/constants/config.go`     | +54 — DefaultLinterExclusionPaths, DefaultFormatterExclusionPaths, DefaultExclusionRules, DefaultFormatterSettings, 4 new DefaultLinterSettings |
| `pkg/linter/fixer_config.go`  | +113 — updateExclusionRules, injectDefaultFormatterSettings, formatter defaults wiring                                                          |
| `pkg/linter/fixer_test.go`    | +208 — 12 new BDD tests (exclusion paths, rules, linter defaults, formatter defaults)                                                           |
| `pkg/linter/analyzer.go`      | +14/-26 — Extracted summaryParts() helper (funlen fix)                                                                                          |
| `pkg/ui/finding_formatter.go` | +17/-22 — Extracted writeFindingGroup() helper (funlen fix + varnamelen fix)                                                                    |
| `pkg/linter/fixer.go`         | +1 — Wire updateExclusionRules into applyAndSave                                                                                                |
| `pkg/config/loader.go`        | +5/-2 — Break long error format string (golines fix)                                                                                            |
| `flake.nix` / `flake.lock`    | Auto-updated by pre-commit hook                                                                                                                 |
