# Comprehensive Status Report

**Date:** 2026-04-02 18:43\
**Branch:** master\
**Total Go LOC:** 10,681\
**Last 20 commits:** 7cd4d3b..689b2a0

---

## a) FULLY DONE

### Core Architecture

- **CLI framework** (Cobra + Fang): Root command, configure, analyze, validate, report, migrate, install-hook — all wired and functional
- **Config loading** (`pkg/config/loader.go`): Multi-format support (YAML/TOML/JSON), auto-detection by extension, in-memory fs support for testing, afero-based
- **Type system** (`pkg/types/types.go`): Strong types (`LinterName`, `FormatterName`, `LinterPriority`, `FormatterPriority`), comprehensive interfaces (`ConfigLoader`, `LinterAnalyzer`, `LinterFixer`), railway-oriented Result types
- **Error handling** (`pkg/errors/errors.go`): Custom error types (`ConfigError`, `AnalysisError`, `ReportError`) with structured fields
- **Linter data** (`pkg/constants/`): 125+ linters categorized across 4 priority levels (Critical/High/Medium/Optional) with human-readable reasons
- **Linter priorities**: 11 Critical, 50+ High, 45+ Medium, rest Optional
- **Formatter support**: 6 formatters (gci, gofmt, gofumpt, goimports, golines, swaggo) with priority system
- **Pre-flight checks** (`fixer_preflight.go`): Version field fix, invalid duration fix, deprecated linter removal, typecheck removal — all working
- **Deprecated linter replacements** (`rules.go`): 8 mappings (wsl→wsl_v5, deadcode→staticcheck, varcheck→staticcheck, etc.)
- **Redundant linter detection** (`rules.go`): lll→golines redundancy mapping
- **Disabled linters** (`rules.go`): funcorder explicitly disabled
- **Swaggo detection** (`pkg/detection/`): Checks go.mod imports + code annotations for swaggo patterns
- **Project type detection** (`pkg/detection/detector.go`): CLI, Library, Web, API, Monorepo — with caching
- **Git utilities** (`pkg/utils/git.go`): IsGitRepo, CheckGitRepo, CheckGitRepoWithTimeout
- **Retry logic** (`pkg/utils/retry.go`): Exponential backoff with context cancellation support
- **FormatterManager** (`fixer_formatters.go`): Extracted formatter logic — core formatters (gci, gofumpt, goimports), golines, swaggo, redundant gofmt removal, ordered output
- **v1→v2 migration** (`pkg/migration/`): Full migrator from golangci-lint v1 to v2 config format
- **HTML report generation** (`pkg/report/`): Templ-based with dark mode, responsive design
- **JSON report generation** (`pkg/report/`): Structured JSON output
- **Pre-commit hook** (`scripts/`): Installable git pre-commit hook
- **Preset system** (`pkg/constants/presets.go`): minimal (5), standard (8), strict (17), security (1), performance (4)
- **GOEXPERIMENT build tags**: Auto-adds goexperiment.jsonv2, goexperiment.simd, goexperiment.goroutineleakprofile
- **Parallel/Serial runner settings**: Auto-enables allow-parallel-runners and allow-serial-runners
- **Go version detection**: Auto-detects local Go version and sets run.go field
- **CI pipeline** (GitHub Actions): Go 1.25 + 1.26 matrix, golangci-lint action, Codecov
- **Justfile**: 13 commands (build, test, lint, fmt, tidy, etc.)
- **AGENTS.md**: Comprehensive project documentation for AI agents

### Today's Change: Default Priority → Optional

- Changed `--priority` default from `"high"` → `"optional"` in `internal/cli/cmd_configure.go:117`
- Updated help text to reflect new default
- This means `golangci-lint-auto-configure configure` now enables ALL linters by default
- Users can still use `--priority critical/high/medium` for conservative configurations

---

## b) PARTIALLY DONE

### Formatter Auto-Enable

- **Core formatters** (gci, gofumpt, goimports): Logic exists in `FormatterManager.EnableCoreFormatters()` BUT they're only enabled during `applyLintersFix`, not when creating a fresh default config via `CreateDefaultConfig()`
- **CreateDefaultConfig** only adds linters, not formatters — fresh configs get 0 formatters enabled until configure runs again

### Linter Settings Configuration

- `LintersConfig.Settings` field exists as `map[string]any` but the tool does NOT auto-configure linter-specific settings (e.g., funlen lines/statements thresholds, gocyclo complexity threshold)
- Users must manually configure these after running the tool

### DisabledLinters System

- Only `funcorder` is in the DisabledLinters map
- No clear documentation or mechanism for users to customize this list
- The user asked that `wsl` (deprecated) and `lll` (redundant with golines) should be disabled — these are handled by different mechanisms (DeprecatedLinters and RedundantLinters) but could be confusing

### Test Coverage

- Unit tests exist for: analyzer, categorizer, fixer, config loader, detector, differ, errors, utils
- Integration tests exist but are limited
- No end-to-end test that runs the full `configure` flow against a real golangci-lint binary
- Coverage data not available (can't run `just test` due to Go version mismatch in workspace)

---

## c) NOT STARTED

### DI Container

- `internal/di/` directory exists but is completely empty
- All dependency injection is manual in CLI commands

### Config Diffing Visualization

- `pkg/diff/differ.go` exists but there's no CLI command to show a nice diff of what changed
- Could be integrated into the configure command output

### Interactive Mode

- No interactive/prompt-based configuration where users can approve individual linters
- Would be valuable for the `--priority optional` default to let users opt-out of specific linters

### Config Profiles/Inheritance

- No support for base config + project-specific overrides
- Common in monorepos where teams share a base config

### Linter Conflict Detection

- No detection of mutually exclusive or conflicting linters
- e.g., both `gocyclo` and `cyclop` measure cyclomatic complexity — enabling both is redundant

### Auto-tune Linter Settings

- No intelligence around setting linter-specific thresholds based on project analysis
- e.g., setting funlen lines based on 90th percentile of existing functions

### Benchmark Suite

- `pkg/linter/analyzer_bench_test.go` and `pkg/detection/detector_bench_test.go` exist but benchmarks are minimal
- No performance regression tracking

### Config Backup/Restore

- No explicit backup mechanism beyond git
- Tool warns about git but doesn't create `.golangci.yml.bak` for non-git users

---

## d) TOTALLY FUCKED UP

### Build Environment — go.work Version Mismatch

- **CRITICAL BLOCKER**: `go.work` requires Go >= 1.26.1 but the system has Go 1.26.0
- This means `just test`, `just build`, `just lint` all FAIL
- LSP can't load packages → no type checking, no autocomplete, no diagnostics
- All 7 projects in the workspace show the same error
- **Fix**: Either upgrade Go to 1.26.1 or update go.work to require 1.26.0

### `fixer.go` References Non-Existent Code (LSP Warning)

- LSP shows warnings about `undefined: filepath` and `undefined: detection` in fixer.go lines 441-442
- This is likely because the LSP can't compile the project (go.work version issue)
- The code itself is correct — `filepath` is imported in `fixer_formatters.go` and `detection` is imported there too
- **Root cause**: LSP can't parse the package due to the Go version issue

### `pkg/utils/retry.go:50:18: undefined: op`

- LSP reports this but it's likely another victim of the Go version mismatch
- The actual code at line 50 uses `executeOperation()` which is correctly defined as a parameter
- **Root cause**: Same go.work version issue

---

## e) WHAT WE SHOULD IMPROVE

### Critical

1. **Fix Go version mismatch** — Without this, nothing compiles, tests don't run, linters can't check code
2. **End-to-end integration tests** — Test the full configure flow with a real golangci-lint binary
3. **Linter-specific settings auto-configuration** — funlen, gocyclo, etc. should get sensible defaults

### High Impact

4. **Formatter auto-enable in CreateDefaultConfig** — Fresh configs should include core formatters
5. **Remove duplicate linters from LinterPriorities** — `gofmt` is listed as Medium but is redundant when gofumpt is enabled (handled by RedundantLinters, but still confusing)
6. **Interactive mode for optional linters** — With `--priority optional` as default, users should be able to review and deselect
7. **Preset system overhaul** — Current presets are tiny (5-17 linters) compared to the 125+ available; presets should map to priority levels
8. **Conflict detection** — Warn about overlapping linters (gocyclo vs cyclop, gofmt vs gofumpt, etc.)

### Quality of Life

9. **Config diff output** — Show exactly what changed after configure runs
10. **Verbose logging improvement** — Current debug logs are verbose but not structured enough for debugging
11. **Error messages for common failures** — "parallel golangci-lint is running" should suggest `--allow-parallel-runners`
12. **Progress indicators** — The tool can be slow (30s+ for large projects); add spinners/progress bars

---

## f) Top #25 Things We Should Get Done Next

| #  | Task                                                                                                    | Priority | Effort | Impact                       |
| -- | ------------------------------------------------------------------------------------------------------- | -------- | ------ | ---------------------------- |
| 1  | Fix Go version mismatch (go.work → 1.26.0 or upgrade Go)                                                | CRITICAL | 5min   | Unblocks everything          |
| 2  | Run `just test` and verify all tests pass with new default priority                                     | CRITICAL | 10min  | Validates change             |
| 3  | Run `just lint` and fix any issues                                                                      | CRITICAL | 15min  | Code quality                 |
| 4  | Add core formatters (gci, gofumpt, goimports) to CreateDefaultConfig                                    | HIGH     | 30min  | Fresh configs get formatters |
| 5  | Add end-to-end test for full configure flow                                                             | HIGH     | 2hr    | Confidence in releases       |
| 6  | Detect and warn about conflicting/overlapping linters                                                   | HIGH     | 2hr    | Better user experience       |
| 7  | Auto-configure linter settings (funlen lines=80, gocyclo min-complexity=15, etc.)                       | HIGH     | 3hr    | Sensible defaults            |
| 8  | Add `--interactive` flag for reviewing individual linter selections                                     | HIGH     | 4hr    | User control                 |
| 9  | Show config diff after configure (what changed)                                                         | MEDIUM   | 2hr    | Transparency                 |
| 10 | Overhaul presets to be more comprehensive                                                               | MEDIUM   | 1hr    | Better defaults              |
| 11 | Add `--exclude` flag to skip specific linters during configure                                          | MEDIUM   | 1hr    | User control                 |
| 12 | Add config backup before modification (.golangci.yml.bak)                                               | MEDIUM   | 30min  | Safety net                   |
| 13 | Validate that enabled linters actually exist in golangci-lint                                           | MEDIUM   | 1hr    | Prevents config errors       |
| 14 | Add `golangci-lint-auto-configure diff` command for comparing configs                                   | MEDIUM   | 2hr    | Useful feature               |
| 15 | Write benchmarks for hot paths (CategorizeLinters, enableRecommendedLinters)                            | LOW      | 1hr    | Performance tracking         |
| 16 | Add structured JSON logging option                                                                      | LOW      | 1hr    | CI/CD integration            |
| 17 | Parallelize golangci-lint binary calls where possible                                                   | LOW      | 2hr    | Speed improvement            |
| 18 | Add shell completions for zsh/bash/fish                                                                 | LOW      | 2hr    | UX improvement               |
| 19 | Create a `--check` mode that exits with code 1 if config needs changes (for CI)                         | LOW      | 1hr    | CI/CD integration            |
| 20 | Add timeout for golangci-lint binary calls (prevent hanging)                                            | LOW      | 30min  | Robustness                   |
| 21 | Support for `.golangci.yml` schema validation against golangci-lint JSON schema                         | LOW      | 2hr    | Advanced validation          |
| 22 | Add `golangci-lint-auto-configure update` command to update linter priorities from latest golangci-lint | LOW      | 3hr    | Keep linter data fresh       |
| 23 | Remove `internal/di/` if unused or implement basic DI                                                   | LOW      | 30min  | Code cleanliness             |
| 24 | Add migration path documentation (v1→v2 config format guide)                                            | LOW      | 1hr    | User documentation           |
| 25 | Implement `--config-out` flag to write to a different file (non-destructive mode)                       | LOW      | 30min  | Safety + flexibility         |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the `--priority optional` default ALSO enable the `--preset` system, or should presets remain a separate opt-in path?**

Context: Now that the default priority is "optional" (enabling ALL linters), the preset system (`--preset standard/strict/minimal`) becomes somewhat contradictory — it enables a tiny subset (5-17 linters) while the default enables 125+. I can't determine if:

1. Presets should be removed/deprecated in favor of the priority system (since `--priority optional` achieves what `--preset strict` did, but better)
2. Presets should be redesigned as "opinionated curated sets" that add project-type-specific settings (e.g., `--preset web` enables web-specific linters + settings)
3. Both systems should coexist as-is (presets for quick starts, priority for fine-tuning)

This is a product/design decision that affects the CLI interface, documentation, and user mental model.

---

## Recent Commit History (Last 7)

| Commit    | Description                                                                                  |
| --------- | -------------------------------------------------------------------------------------------- |
| `7cd4d3b` | refactor: complete formatter management extraction to FormatterManager                       |
| `08ded63` | refactor: extract formatter logic into dedicated FormatterManager and improve error handling |
| `9ba7a9e` | refactor: extract formatter logic and improve error handling                                 |
| `e411472` | fix(build): add missing context parameter and default case                                   |
| `458701c` | feat(fixer): add core formatter auto-enable, runner settings, and build tags                 |
| `c6da5b3` | feat: improve formatter handling with core formatters and smarter deduplication              |
| `4248625` | feat(detection): add swaggo detection support and update build tag syntax                    |

---

## File Statistics

| Package          | Files | Key Types                                                                                            |
| ---------------- | ----- | ---------------------------------------------------------------------------------------------------- |
| `pkg/linter/`    | 8     | Analyzer, Fixer, FormatterManager                                                                    |
| `pkg/constants/` | 7     | LinterPriorities, LinterReasons, PresetLinters, DeprecatedLinters, DisabledLinters, RedundantLinters |
| `pkg/types/`     | 3     | Config, LinterPriority, ConfigLoader interface                                                       |
| `pkg/config/`    | 2     | Loader (CRUD for configs)                                                                            |
| `pkg/detection/` | 4     | Detector, ProjectType                                                                                |
| `pkg/migration/` | 7     | Migrator (v1→v2)                                                                                     |
| `internal/cli/`  | 9     | Commands, configure, analyze, validate, report                                                       |
| `pkg/report/`    | 3     | HTML + JSON generators                                                                               |
| `pkg/utils/`     | 4     | Git, Retry utilities                                                                                 |

---

_This report reflects the state as of 2026-04-02 18:43, including the `--priority optional` default change made in this session._
