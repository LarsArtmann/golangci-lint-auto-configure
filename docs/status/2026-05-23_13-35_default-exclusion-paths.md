# Status Report — 2026-05-23

**Generated:** 2026-05-23 13:35 UTC  
**Branch:** master  
**Last Release:** v0.2.0 (3f2681b)  
**Total LoC:** ~16,467 Go (excluding generated `_templ.go`)

---

## A) FULLY DONE

### Default Exclusion Paths (this session)

**Problem:** `*_templ.go` and `vendor/` files were NOT always excluded from linting. The tool relied entirely on dynamic gogenfilter scanning — if no generated files were detected, no exclusions were added. This meant fresh projects or projects without detectable generators got zero exclusion paths.

**Solution:** Added hardcoded default exclusion paths that are unconditionally injected into every config:

| Section                       | Paths Added              |
| ----------------------------- | ------------------------ |
| `linters.exclusions.paths`    | `_templ\.go$`, `vendor/` |
| `formatters.exclusions.paths` | `_templ\.go$`            |

**Files changed:**

- `pkg/constants/config.go` — `DefaultLinterExclusionPaths` + `DefaultFormatterExclusionPaths`
- `pkg/linter/fixer_config.go` — `updateGeneratedExclusions` now injects defaults before dynamic scan
- `pkg/linter/fixer_test.go` — 4 new BDD tests

**Test results:** 14/14 suites PASS, 46/46 linter specs PASS

### Previously Completed (v0.2.0 and earlier)

| Feature                                                                                  | Status |
| ---------------------------------------------------------------------------------------- | ------ |
| 7 CLI commands (configure, analyze, validate, report, migrate, install-hook, completion) | Done   |
| 119 linter priorities with reasons                                                       | Done   |
| gogenfilter integration for generated file exclusion                                     | Done   |
| v1 → v2 config migration                                                                 | Done   |
| Deprecated linter auto-replacement (wsl→wsl_v5, gomodguard→gomodguard_v2, etc.)          | Done   |
| Version-gated deprecation (gomodguard_v2 needs v2.12.0+)                                 | Done   |
| go-finding integration (SARIF, JSON, unified model)                                      | Done   |
| HTML report generation (templ-based)                                                     | Done   |
| Build tags auto-injection (GOEXPERIMENT flags)                                           | Done   |
| Depguard/ireturn/gocritic/exhaustruct defaults                                           | Done   |
| Semantic deduplication to zero clones                                                    | Done   |
| Nix flake build + justfile recipes                                                       | Done   |
| CI (Go 1.25/1.26 matrix, lint, coverage)                                                 | Done   |
| Pre-commit hook support                                                                  | Done   |

---

## B) PARTIALLY DONE

| Area                             | What's Done                               | What's Missing                                                                                                                                               |
| -------------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Default test exclusion rules** | Paths (`_templ.go`, `vendor/`)            | No default `linters.exclusions.rules` for test files (e.g. `exhaustruct`, `funlen`, `cyclop` exclusions for `_test.go`) — project-dependency-graph has these |
| **CLI integration tests**        | `internal/cli/integration_test.go` exists | Coverage is 9.0% — most command paths untested                                                                                                               |
| **Report generation**            | HTML + JSON done                          | `pkg/report` shows 0.0% coverage                                                                                                                             |
| **Migration test coverage**      | 66.8%                                     | Edge cases for complex v1 configs                                                                                                                            |

---

## C) NOT STARTED

1. **Default exclusion rules for test files** — `project-dependency-graph/.golangci.yml` has `linters.exclusions.rules` excluding `exhaustruct`, `testpackage`, `gochecknoglobals`, `funlen`, `cyclop`, `goconst` for `_test\.go` paths. This tool does NOT inject these.

2. **`revive` default settings** — The reference config has `revive.rules` to disable `exported` and `package-comments`. Not in `DefaultLinterSettings`.

3. **`varnamelen` default settings** — Reference config has `ignore-map-index-ok: true`, `ignore-names`, `ignore-type-assert-ok: true`. Not in defaults.

4. **`gomoddirectives` default settings** — Reference has `replace-local: true`. Not in defaults.

5. **`golines` formatter default settings** — Reference has `max-len: 120`. Not in defaults (only enabled when `lll` detected).

6. **`cyclop` default max-complexity** — Reference has `max-complexity: 12`. Not in defaults.

7. **`gocritic` ifElseChain** — Reference disables only `ifElseChain`, we also disable `dupImport`, `octalLiteral`, `whyNoLint` (stricter).

8. **Output formats config** — Reference has `output.formats: {}`. Not injected.

9. **`allow-parallel-runners` and `allow-serial-runners`** — Done! Already injected by `updateRunnerSettings`.

10. **Features.md audit** — No `FEATURES.md` file exists.

11. **TODO_LIST.md** — No comprehensive TODO list.

12. **Exclusion path for `.gen.go` files** — oapi-codegen and other `.gen.go` generators are only excluded if dynamically detected. No default path.

13. **Swaggo formatter auto-detection** — Only enabled if `swag` annotations found. Not a default.

14. **Config validation strictness** — `validate` command exists but doesn't cross-check against reference patterns like project-dependency-graph.

---

## D) TOTALLY FUCKED UP

**Nothing is broken.** All tests pass, all linter issues are pre-existing (not introduced this session). The codebase is stable at v0.2.0 with clean master.

**Pre-existing lint warnings (NOT introduced this session):**

| File                              | Linter  | Issue                                     |
| --------------------------------- | ------- | ----------------------------------------- |
| `pkg/config/loader.go:269`        | golines | Line too long                             |
| `pkg/linter/command_runner.go:41` | funlen  | `runGolangciLintLinters` is 38 lines > 30 |
| `pkg/ui/finding_formatter.go:12`  | funlen  | `FormatFindings` is 31 lines > 30         |

These are minor and existed before this session.

---

## E) WHAT WE SHOULD IMPROVE

### Critical (High Impact)

1. **Default test exclusion rules** — Every Go project needs test file exemptions for certain linters. This should be automated, not manually configured. The reference config proves it.

2. **Default linter settings are incomplete** — `revive`, `varnamelen`, `gomoddirectives`, `cyclop`, `golines` all need sensible defaults matching the reference project pattern.

3. **CLI test coverage at 9.0%** — The CLI layer is where all the wiring happens. Integration tests should cover every command path.

4. **Report package at 0.0% coverage** — Completely untested templ-based HTML generation.

### Medium Impact

5. **No FEATURES.md** — No documented feature inventory. Hard to track what's done vs planned.

6. **No TODO_LIST.md** — No consolidated task list. Work is scattered across status reports.

7. **gogenfilter scanner coverage at 59.8%** — Error paths and edge cases untested.

8. **Migration coverage at 66.8%** — Complex v1 config edge cases may not be covered.

### Nice to Have

9. **Example configs are stale** — `examples/` doesn't reflect current default settings injection.

10. **No benchmarking** — Only one bench test (`analyzer_bench_test.go`). Performance regression not tracked.

---

## F) TOP #25 THINGS TO DO NEXT

| #   | Priority | Task                                                                                      | Impact                     |
| --- | -------- | ----------------------------------------------------------------------------------------- | -------------------------- |
| 1   | Critical | Add default `linters.exclusions.rules` for test files (exhaustruct, funlen, cyclop, etc.) | Every project needs this   |
| 2   | Critical | Add `revive` default settings (disable `exported`, `package-comments`)                    | Linter noisy without it    |
| 3   | Critical | Add `varnamelen` default settings (ignore-map-index-ok, common short names)               | Linter noisy without it    |
| 4   | Critical | Add `gomoddirectives` default settings (`replace-local: true`)                            | Local dev needs this       |
| 5   | Critical | Add `cyclop` default settings (`max-complexity: 12`)                                      | Default is too strict      |
| 6   | High     | Add `golines` default settings (`max-len: 120`) when enabled                              | Consistent line length     |
| 7   | High     | Create `FEATURES.md` with full feature audit                                              | Track what exists          |
| 8   | High     | Create `TODO_LIST.md` from all status reports                                             | Consolidated task tracking |
| 9   | High     | Increase CLI integration test coverage (9.0% → 50%+)                                      | Trust in command wiring    |
| 10  | High     | Add report package tests (0.0% → 50%+)                                                    | HTML output untested       |
| 11  | High     | Fix pre-existing lint: `command_runner.go` funlen (extract helpers)                       | Clean lint baseline        |
| 12  | High     | Fix pre-existing lint: `finding_formatter.go` funlen (extract helpers)                    | Clean lint baseline        |
| 13  | High     | Fix pre-existing lint: `loader.go:269` golines (break long line)                          | Clean lint baseline        |
| 14  | Medium   | Add default exclusion for `.gen.go` files (oapi-codegen pattern)                          | Common generator           |
| 15  | Medium   | Increase gogenfilter scanner coverage (59.8% → 80%+)                                      | Error path coverage        |
| 16  | Medium   | Increase migration coverage (66.8% → 80%+)                                                | Edge case coverage         |
| 17  | Medium   | Update example configs to reflect current default injection                               | Accurate examples          |
| 18  | Medium   | Add `exhaustruct` test-file exclusion by default (not just via gogenfilter)               | Noisy in tests             |
| 19  | Medium   | Add `output.formats: {}` to default config creation                                       | Explicit config            |
| 20  | Low      | Add `swaggo` formatter detection improvements                                             | Better auto-detection      |
| 21  | Low      | Add config validation cross-check against reference patterns                              | Consistency                |
| 22  | Low      | Add benchmarking for analyzer and fixer                                                   | Performance regression     |
| 23  | Low      | Add `ginkgolinter` to recommended test linters                                            | Test quality               |
| 24  | Low      | Document exclusion pattern syntax (RE2 regex) in README                                   | User education             |
| 25  | Low      | Add `--check` mode for CI (exit 1 if config needs changes)                                | CI/CD integration          |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should `vendor/` be excluded from `formatters.exclusions.paths` too?**

Currently the implementation adds `vendor/` only to `linters.exclusions.paths` (linters) but NOT to `formatters.exclusions.paths`. The reference config (`project-dependency-graph/.golangci.yml`) also only excludes `vendor/` from linters paths, not formatters paths. However, formatting vendor code is generally undesirable (modifies vendored dependencies, creates noisy diffs). The question is:

- Is there a reason golangci-lint or the reference project intentionally ALLOWS formatting vendor code?
- Or should we add `vendor/` to `DefaultFormatterExclusionPaths` too for safety?

I erred on the side of matching the reference config exactly, but this could be wrong.

---

## Test Summary

```
Ginkgo ran 14 suites — ALL PASSED
Composite coverage: 60.6% of statements

Key packages:
  pkg/errors    — 100.0%
  pkg/diff      — 96.5%
  pkg/utils     — 94.6%
  pkg/constants — 80.0%
  pkg/linter    — 79.5%
  pkg/config    — 65.9%
  pkg/detection — 65.5%
  pkg/migration — 66.8%
  pkg/ui        — 67.7%
  pkg/gogenfilter — 59.8%
  pkg/finding   — 56.0%
  pkg/types     — 59.4%
  internal/cli  — 9.0%
  pkg/report    — 0.0%
```

---

## Files Changed This Session

| File                         | Change                                                                        |
| ---------------------------- | ----------------------------------------------------------------------------- |
| `pkg/constants/config.go`    | Added `DefaultLinterExclusionPaths` and `DefaultFormatterExclusionPaths`      |
| `pkg/linter/fixer_config.go` | Refactored `updateGeneratedExclusions` to inject defaults before dynamic scan |
| `pkg/linter/fixer_test.go`   | Added 4 BDD tests for default exclusion paths                                 |
