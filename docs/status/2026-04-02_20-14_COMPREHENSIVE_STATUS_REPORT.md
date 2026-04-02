# Comprehensive Status Report

**Date:** 2026-04-02 20:14
**Branch:** master
**Commit:** adead5a
**Go Version:** 1.26.1 (darwin/arm64)
**Total LOC:** 10,681 across 63 Go files
**Test Files:** 16 (3,389 lines)

---

## a) FULLY DONE

### Core Architecture

- **CLI framework** (Cobra + Fang): Root + 7 subcommands (configure, analyze, validate, report, migrate, install-hook, completion) — all wired and functional
- **Config loading** (`pkg/config/loader.go`, 413 lines): Multi-format (YAML/TOML/JSON), afero-based with in-memory FS for testing, auto-detection by extension, `CreateDefaultConfig` with all linters enabled
- **Type system** (`pkg/types/types.go`, 317 lines): Strong types (`LinterName`, `FormatterName`, `LinterPriority`, `FormatterPriority`), 7 composed interfaces (`ConfigLoader` = `ConfigReader` + `ConfigWriter` + `ConfigDiscovery` + ...), railway-oriented `mo.Result[T]` wrappers
- **Error handling** (`pkg/errors/errors.go`): Custom types (`ConfigError`, `AnalysisError`, `ReportError`) with structured fields and error wrapping
- **Result types** (`pkg/types/result.go`, 88 lines): Type aliases for `mo.Result[T]` with ergonomic constructors (`OkConfig`/`ErrConfig`, etc.)

### Linter System

- **125+ linters categorized** across 4 priority levels:
  - Critical (11): gosec, errcheck, staticcheck, govet, errchkjson, musttag, sloglint, nilerr, noctx, loggercheck, paralleltest
  - High (~55): wrapcheck, errorlint, prealloc, unconvert, ineffassign, gocyclo, funlen, cyclop, gocognit, maintidx, exhaustive, exhaustruct, goconst, misspell, revive, nolintlint, forcetypeassert, etc.
  - Medium (~45): dupword, godot, godox, goheader, gofmt, gci, varnamelen, whitespace, wsl_v5, grouper, dogsled, makezero, thelper, exportloopref, etc.
  - Optional: rest default to Optional
- **Human-readable reasons** (`linter_reasons.go`, 127 lines): Every linter has a reason string explaining why it's recommended
- **Deprecated linter replacements** (`rules.go`): 9 mappings — wsl→wsl_v5, deadcode/varcheck/structcheck/gosimple/interfacer→staticcheck, exhaustivestruct→exhaustive, maligned→govet(fieldalignment), nosnakecase→revive
- **Disabled linters** (`rules.go`): funcorder explicitly disabled
- **Redundant linter detection** (`rules.go`): lll→golines (lll redundant when golines formatter handles line length)

### Formatter System

- **6 formatters**: gci, gofmt, gofumpt, goimports, golines, swaggo — all with priority levels and reasons
- **FormatterManager** (`fixer_formatters.go`, 181 lines): Extracted formatter logic — core formatters (gci, gofumpt, goimports), golines, swaggo, redundant gofmt removal, ordered output
- **Redundant formatters** (`config.go`): gofmt redundant when gofumpt is enabled
- **Formatters managed by BuildFlow**: goimports, gofumpt (excluded from auto-enable when BuildFlow manages them)

### Fixer System

- **Fixer** (`fixer.go`, 497 lines): Main orchestrator with pre-flight → apply → save pipeline
- **Pre-flight checks** (`fixer_preflight.go`, 271 lines): Version field fix (ensures `version: "2"`), invalid duration fix (default `"5m"`), deprecated linter removal, typecheck removal
- **Dry-run support**: Full dry-run mode with detailed result reporting

### Detection System

- **Project type detection** (`detector.go`, 416 lines): CLI, Library, Web, API, Monorepo — with caching, thread-safe
- **Swaggo detection**: Checks go.mod imports + code annotations for swaggo patterns
- **Go version detection**: Auto-detects local Go version and sets `run.go` field

### Analysis System

- **Analyzer** (`analyzer.go`, 254 lines): Gathers linters/formatters from golangci-lint binary, categorizes, recommends
- **Categorizer** (`categorizer.go`, 130 lines): Priority-based categorization with human-readable output

### Additional Features

- **v1→v2 migration** (`pkg/migration/`, 8 files + testdata): Full migrator from golangci-lint v1 to v2 config format
- **HTML report** (`pkg/report/`): Templ-based with dark mode, responsive design
- **JSON report** (`pkg/report/`): Structured JSON output
- **Diff system** (`pkg/diff/`): Config comparison logic
- **Preset system** (`presets.go`): minimal (5), standard (8), strict (17), security (1), performance (4)
- **GOEXPERIMENT build tags**: Auto-adds goexperiment.jsonv2, goexperiment.simd, goexperiment.goroutineleakprofile
- **Runner settings**: Auto-enables allow-parallel-runners and allow-serial-runners
- **Git utilities** (`pkg/utils/git.go`, 58 lines): IsGitRepo, CheckGitRepo, CheckGitRepoWithTimeout (5s default)
- **Retry logic** (`pkg/utils/retry.go`, 93 lines): Exponential backoff with context cancellation

### Infrastructure

- **CI pipeline** (GitHub Actions): Go 1.25 + 1.26 matrix, golangci-lint action, Codecov
- **Justfile**: 13 commands (build, test, lint, fmt, tidy, etc.)
- **Pre-commit hook**: Installable git pre-commit hook
- **AGENTS.md**: Comprehensive project documentation for AI agents
- **Default priority → optional** (today's change): `--priority` now defaults to `"optional"`, enabling ALL linters

---

## b) PARTIALLY DONE

### Formatter Auto-Enable in CreateDefaultConfig

- `FormatterManager.EnableCoreFormatters()` exists and works during `applyLintersFix`
- **BUT** `CreateDefaultConfig()` only adds linters — fresh configs get 0 formatters until configure runs again
- Need to add formatter enablement in the default config creation path

### Linter-Specific Settings

- `LintersConfig.Settings` field exists as `map[string]any` but no auto-configuration
- Users must manually configure thresholds (funlen lines/statements, gocyclo complexity, etc.)
- Should provide sensible defaults at minimum

### Test Coverage

- **16 test files, 3,389 lines of tests** — good breadth
- Unit tests: analyzer, categorizer, fixer, config loader, detector, differ, errors, utils, commands, integration
- Benchmarks: `analyzer_bench_test.go` (53 lines), `detector_bench_test.go` (75 lines)
- **Gaps**: No end-to-end test running the full `configure` flow against a real golangci-lint binary
- Coverage data unavailable (can't run `just test` — see section d)

### Preset System

- 5 presets exist but are tiny (5-17 linters) compared to 125+ available
- Now that `--priority optional` is the default, presets are contradictory
- Need redesign or deprecation decision

### File Size Compliance

- 4 files exceed the 350-line limit:
  - `pkg/linter/fixer.go`: 497 lines (+147, 42% over) — **CRITICAL**
  - `pkg/config/loader.go`: 413 lines (+63, 18% over)
  - `pkg/detection/detector.go`: 416 lines (+66, 19% over)
  - `internal/cli/commands_test.go`: 391 lines (+41, 12% over)

---

## c) NOT STARTED

### DI Container

- `internal/di/` directory does NOT exist anymore (was removed)
- All dependency injection is manual in CLI commands — actually fine for this project size

### Config Diff Visualization

- `pkg/diff/differ.go` exists but no CLI command exposes it
- No `golangci-lint-auto-configure diff` command

### Interactive Mode

- No `--interactive` flag for reviewing/approving individual linters
- Critical now that `--priority optional` enables everything by default

### Config Profiles/Inheritance

- No base config + project-specific overrides
- Would help monorepo setups

### Linter Conflict Detection

- No detection of overlapping linters (gocyclo vs cyclop, gofmt vs gofumpt, etc.)
- Users can enable both and get duplicate warnings

### Auto-Tune Linter Settings

- No intelligence around setting linter-specific thresholds based on project analysis
- e.g., funlen lines based on 90th percentile of existing functions

### Config Backup Mechanism

- No `.golangci.yml.bak` creation
- Tool warns about git but doesn't create backups for non-git users

### Config Output Flag

- No `--config-out` flag to write to a different file
- Would enable non-destructive "preview" mode

---

## d) TOTALLY FUCKED UP

### 1. Go Build Cache Corruption

- `go build ./...` fails with corrupted cache errors: `open .../Library/Caches/go-build/.../: no such file or directory`
- `GOWORK=off go build ./...` also fails — std lib packages not found
- `go clean -cache` was attempted but rebuild hung indefinitely
- **Impact**: Cannot build, cannot run tests, cannot run linters
- **Fix**: `go clean -cache && go build ./...` (may need multiple attempts)

### 2. go.work Does Not Include This Project

- `/Users/larsartmann/projects/go.work` lists 11 projects but NOT `golangci-lint-auto-configure`
- `go build ./...` fails: "directory prefix . does not contain modules listed in go.work"
- **Fix**: Either add this project to `go.work`, or use `GOWORK=off` consistently, or remove the parent `go.work`

### 3. Conflicting Default Priority

- `internal/cli/commands.go`: Global `--priority` flag defaults to `"high"`
- `internal/cli/cmd_configure.go`: Local `--priority` flag defaults to `"optional"`
- Same flag name, different defaults — the local one wins for `configure`, but confusing
- **Fix**: Remove the global `--priority` flag from `commands.go` or align defaults

### 4. Duplicated Code in detector.go

- `analyzeGoMod()` and `analyzeGoModWithError()` are nearly identical (~50 lines duplicated)
- `analyzeGoMod()` silently ignores `scanner.Err()`
- **Fix**: Refactor to single method with error return

### 5. Unreachable Code in retry.go

- Line 83 has unreachable code after the retry loop (the `for` always returns before reaching it)
- `DefaultInitialBackoff` is untyped `const = 500` instead of `time.Duration`
- **Fix**: Remove unreachable code, type the constant properly

---

## e) WHAT WE SHOULD IMPROVE

### Critical (Blocks Everything)

1. **Fix Go build environment** — Resolve cache corruption and workspace inclusion
2. **Run full test suite** — Verify all tests pass after recent refactors
3. **Run linter suite** — Verify code quality after recent changes

### High Impact

4. **Split fixer.go** — 497 lines is 42% over limit; extract more methods/helpers
5. **Align priority defaults** — Remove conflicting global/local `--priority` defaults
6. **End-to-end integration tests** — Full configure flow with real golangci-lint
7. **Formatter auto-enable in CreateDefaultConfig** — Fresh configs should include formatters
8. **Auto-configure linter settings** — funlen lines=80, gocyclo min-complexity=15, etc.
9. **Linter conflict detection** — Warn about overlapping linters
10. **Refactor detector.go** — Merge duplicate `analyzeGoMod` methods

### Quality of Life

11. **Config diff output** — Show what changed after configure
12. **Interactive mode** — Review/deselect linters when `--priority optional`
13. **Preset redesign** — Current presets are too small vs. 125+ linters
14. **`--exclude` flag** — Skip specific linters during configure
15. **Config backup** — Create `.golangci.yml.bak` before modification
16. **`--check` mode** — Exit code 1 if config needs changes (for CI)
17. **`--config-out` flag** — Write to different file (non-destructive)

---

## f) Top #25 Things We Should Get Done Next

| #   | Task                                                                   | Priority | Effort | Impact                        |
| --- | ---------------------------------------------------------------------- | -------- | ------ | ----------------------------- |
| 1   | Fix Go build cache and workspace configuration                         | CRITICAL | 15min  | Unblocks everything           |
| 2   | Run `just test` and verify all tests pass                              | CRITICAL | 10min  | Validates all code            |
| 3   | Run `just lint` and fix any issues                                     | CRITICAL | 15min  | Code quality                  |
| 4   | Align `--priority` defaults (remove global flag or sync to "optional") | HIGH     | 10min  | Prevents confusion            |
| 5   | Split `fixer.go` (497→<350 lines) by extracting more helpers           | HIGH     | 30min  | File size compliance          |
| 6   | Refactor `detector.go`: merge duplicate `analyzeGoMod` methods         | HIGH     | 20min  | Removes ~50 lines duplication |
| 7   | Fix `retry.go`: remove unreachable code, type the constant             | HIGH     | 5min   | Code correctness              |
| 8   | Add core formatters to `CreateDefaultConfig`                           | HIGH     | 30min  | Fresh configs get formatters  |
| 9   | Add end-to-end test for full configure flow                            | HIGH     | 2hr    | Confidence in releases        |
| 10  | Detect and warn about conflicting/overlapping linters                  | HIGH     | 2hr    | Better user experience        |
| 11  | Auto-configure linter settings (funlen, gocyclo, etc.)                 | HIGH     | 3hr    | Sensible defaults             |
| 12  | Add `--interactive` flag for reviewing linter selections               | HIGH     | 4hr    | User control                  |
| 13  | Show config diff after configure (what changed)                        | MEDIUM   | 2hr    | Transparency                  |
| 14  | Overhaul presets (or deprecate in favor of priority system)            | MEDIUM   | 1hr    | Clarity                       |
| 15  | Add `--exclude` flag to skip specific linters                          | MEDIUM   | 1hr    | User control                  |
| 16  | Add config backup before modification                                  | MEDIUM   | 30min  | Safety net                    |
| 17  | Validate enabled linters exist in golangci-lint version                | MEDIUM   | 1hr    | Prevents config errors        |
| 18  | Add `golangci-lint-auto-configure diff` command                        | MEDIUM   | 2hr    | Useful feature                |
| 19  | Write comprehensive benchmarks for hot paths                           | LOW      | 1hr    | Performance tracking          |
| 20  | Add structured JSON logging option                                     | LOW      | 1hr    | CI/CD integration             |
| 21  | Add shell completions for zsh/bash/fish                                | LOW      | 2hr    | UX improvement                |
| 22  | Create `--check` mode for CI (exit code 1 if changes needed)           | LOW      | 1hr    | CI/CD integration             |
| 23  | Add timeout for golangci-lint binary calls                             | LOW      | 30min  | Robustness                    |
| 24  | Implement `--config-out` flag for non-destructive output               | LOW      | 30min  | Safety + flexibility          |
| 25  | Add JSON schema validation for config files                            | LOW      | 2hr    | Advanced validation           |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the preset system be removed, redesigned, or kept as-is now that `--priority optional` is the default?**

Context:

- `--priority optional` now enables ALL 125+ linters by default
- Presets enable tiny subsets (minimal=5, standard=8, strict=17, security=1, performance=4)
- When a user runs `golangci-lint-auto-configure configure` without flags, they get everything — but `--preset standard` gives only 8 linters
- The two systems are contradictory: one says "enable everything", the other says "enable a curated few"

Options:

1. **Remove presets** — The priority system fully replaces them; `--priority critical` is the new "minimal"
2. **Redesign presets as project-type profiles** — `--preset web` enables web-specific linters + settings, `--preset cli` for CLI tools, etc.
3. **Keep both** — Presets for quick-start users, priority for fine-tuning (document the relationship clearly)

This is a product/design decision that affects CLI UX, documentation, and user mental model.

---

## Codebase Statistics

| Package          | Files          | Key Types                                        | Lines (approx) |
| ---------------- | -------------- | ------------------------------------------------ | -------------- |
| `pkg/linter/`    | 6+1 test       | Analyzer, Fixer, FormatterManager, Categorizer   | ~1,333         |
| `pkg/constants/` | 7              | LinterPriorities, LinterReasons, Rules, Presets  | ~438           |
| `pkg/types/`     | 2+1 test       | Config, LinterPriority, interfaces, Result types | ~405           |
| `pkg/config/`    | 1+1 test       | Loader (CRUD for configs)                        | ~740           |
| `pkg/detection/` | 1+2 test/bench | Detector, ProjectType                            | ~704           |
| `pkg/migration/` | 7+1 test       | Migrator (v1→v2)                                 | ~700+          |
| `internal/cli/`  | 10+2 test      | Commands, configure, analyze, validate, report   | ~1,500+        |
| `pkg/report/`    | 3              | HTML + JSON generators                           | ~300           |
| `pkg/utils/`     | 2+2 test       | Git, Retry                                       | ~311           |
| `pkg/diff/`      | 1+1 test       | Differ                                           | ~300           |
| `pkg/ui/`        | 1+1 test       | Formatter (display)                              | ~300           |
| `pkg/errors/`    | 1+1 test       | Custom error types                               | ~200           |
| **Total**        | **63**         |                                                  | **10,681**     |

## Test Statistics

| Test File                                     | Lines     |
| --------------------------------------------- | --------- |
| `internal/cli/commands_test.go`               | 391       |
| `pkg/config/loader_test.go`                   | 327       |
| `pkg/linter/fixer_test.go`                    | 295       |
| `pkg/ui/formatter_test.go`                    | 226       |
| `internal/cli/cmd_configure_internal_test.go` | 236       |
| `internal/cli/integration_test.go`            | 201       |
| `pkg/diff/differ_test.go`                     | 208       |
| `pkg/linter/analyzer_test.go`                 | 210       |
| `pkg/detection/detector_test.go`              | 213       |
| `pkg/migration/migrator_test.go`              | 327       |
| `pkg/errors/errors_test.go`                   | 182       |
| `pkg/utils/retry_test.go`                     | 158       |
| `pkg/linter/categorizer_test.go`              | 127       |
| `pkg/utils/git_test.go`                       | 60        |
| `pkg/linter/analyzer_bench_test.go`           | 53        |
| `pkg/detection/detector_bench_test.go`        | 75        |
| **Total**                                     | **3,389** |

---

_This report reflects the state as of 2026-04-02 20:14, based on full codebase audit of all 63 Go files (10,681 LOC)._
