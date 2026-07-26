# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [0.6.0] - 2026-07-27

A large release covering the friction-reduction Pareto plan and the
subsequent quality-debt cleanup. Grouped by theme.

### Added — Friction reduction (data-driven default tuning)

- `--pragmatic` flag: drops the 4 highest-noise linters (gochecknoglobals, wrapcheck, ireturn, funlen) from the dynamic enable set
- `GosecSettings` typed struct with curated excludes (G304, G115) — reduces gosec false-positive friction while preserving unhandled-error detection (errcheck handles known-benign cases surgically)
- `ErrcheckSettings` typed struct with curated `exclude-functions` (`Close`, `fmt.Fprint*`, Builder writes) — reduces errcheck friction by 20%+
- `WrapcheckSettings` typed struct with a curated `ignore-sigs` list
- Friction-driven defaults: expanded `exhaustruct` exclude list with 14 stdlib structs (net/http.Client/Server/Request/Response/Transport/Cookie, net.TCPAddr/Dialer, slog.HandlerOptions, sync.WaitGroup, bytes.Buffer, time.Ticker/Timer, os/exec.Cmd)
- `house` formatter preset (4 formatters: gci, goimports, gofumpt, golines) — the validated winning stack across 128/160 sibling projects
- `forcetypeassert` added to the default `_test.go` exclusion rules
- `gosec`, `errcheck`, `wrapcheck`, `ireturn`, `recvcheck`, `contextcheck`, `exhaustive` added to default `_test.go` exclusion rules

### Added — Architecture & type safety

- `CommandResult` type: optional structured return for CLI commands carrying a user-facing message and explicit exit code alongside the standard error
- `ConfigLoader` God Object interface decomposed into 6 focused sub-interfaces (ConfigReader, ConfigWriter, ConfigDiscoverer, ConfigValidator, ConfigInspector, ConfigCreator)
- `coreLinters` + `withCore()` pattern eliminating 6× duplication of the core-linter list across detection patterns
- `ValidationError.ToHealthIssue()` conversion bridging the two reporting types
- Settings key validation: soft warnings for unknown linter settings keys at config load time
- `cmd/generate-settings`: generates 88 settings structs from the golangci-lint JSON Schema (codegen infrastructure replacing hand-maintained structs)

### Added — UX & preset ergonomics

- Multi-preset support: repeated `--preset` flags merge linters/formatters with deduplication (`--preset minimal --preset security`)
- `--recommend` flag: analyzes the project and applies multiple recommended presets (implies `--detect`)
- `--detect` mode for the format preset: `--preset format --detect` auto-enables swaggo when Swagger annotations are found
- `--no-color` flag for CI/scripting output (sets `NO_COLOR=1`)
- `--json` output for the `presets` command (structured preset details)

### Added — Error handling

- `HandleError` function at the CLI boundary (replaces raw `slog.Error` calls with classified, user-friendly rendering)
- Domain message templates: 27 Wix-style error messages (What/Why/Fix/WayOut) registered with `errorfamily.New()`

### Added — CI, build & tooling

- All 21 GitHub Actions pinned to immutable commit SHAs across 4 workflow files (supply-chain security)
- `cmd/coverage-check` binary added to Nix `subPackages` and exposed as `nix run .#coverage-check`
- `vendorHash.nix` — dedicated file for Go vendor hash (cleaner diffs on dependency updates)
- CI retry logic for nix build (3 attempts) and resilient magic-nix-cache handling
- Dependabot automation (`.github/dependabot.yml`) for GitHub Actions and Go modules
- `git-cliff` config (`cliff.toml`) for changelog generation from conventional commits
- Coverage-check tool documented in README under "Development Tools"
- `pkg/audit/`, `pkg/policy/`, `pkg/client/`, `pkg/utils/` added to code-organization.md directory tree
- `docs/status/README.md`: consolidating index for historical status reports
- `docs/architecture-understanding/`: dated architecture review documenting package structure, strengths, and concerns

### Added — Tests

- Policy-enforcement tests (`fixer_enforce_test.go`, 14 tests) and audit CLI tests (`cmd_audit_test.go`, ~20 tests) for previously zero-coverage code paths
- Exit-code integration tests for Infrastructure (69) and Corruption (65)
- JSON round-trip tests for all report types (`pkg/types/json_roundtrip_test.go`)
- HTML report golden snapshot test + 8 structural invariant tests
- HTML report CSS regression tests (color golden values + cross-format consistency)
- Coverage-check integration tests for `run()` threshold logic
- Docs-integrity test cross-checking FEATURES.md preset counts against `pkg/constants/presets.go` (fails CI on drift)
- `DefaultExclusionRules` data-integrity tests (valid names, no disabled overlap, no duplicates)

### Changed

- `cmd_configure.go` split from 680 lines into 4 focused files (193/208/182/133 lines)
- `funlen` default thresholds changed from `60/40` to `200/100` (house style — dominant override across 160 sibling projects; diverges from golangci-lint upstream)
- `CoreFormatters` expanded from 3 to 4 formatters (`gci`, `goimports`, `gofumpt`, `golines`) — now matches the `house` preset's validated winning stack
- CI workflow switched from a Go version matrix to `go-version-file: go.mod`; `flake.lock` drift detection added
- README medium linter count corrected from "50+" to "48"; linter-priority tiers corrected (ineffassign/gocyclo/misspell/revive moved to High)
- Stale `//nolint:legacyerrors` directives removed (linter not configured — was producing warnings)
- golangci-lint configuration and linter constants updated
- Nix flake configuration and module dependencies updated

### Fixed

- G104 removed from `GosecSettings.Excludes` defaults — it was too broad, suppressing ALL unhandled-error findings from gosec rather than just the curated Close/Fprint* family that errcheck handles surgically
- `audit` subcommand broken on first run (no ledger exists yet): `os.IsNotExist` does not unwrap `fmt.Errorf %w` chains, so a missing ledger returned an error instead of nil. Changed to `errors.Is(err, os.ErrNotExist)`
- Coverage-check ghost file (`scripts/coverage-check.sh`) deleted — replaced by the Go `cmd/coverage-check`

### Changed — exhaustruct NeverAutoEnable reclassification

- `exhaustruct` (highest-friction linter across 160 sibling projects, 6.5 nolint ratio) moved from auto-enabled (High priority, in the `reference` preset) to a new **`NeverAutoEnableLinters`** tier: it is never recommended or auto-enabled, but it is **never stripped** if a user manually adds it. Safe default settings (14 stdlib struct excludes) and `_test.go` exclusion rules are still injected when it is manually enabled
- New **linter management tiers**: `DisabledLinters` (forcibly disabled), `NeverAutoEnableLinters` (never recommended, respected if manual), and `PragmaticNoiseLinters` (opt-out via `--pragmatic`) — three disjoint maps with data-integrity tests enforcing consistency
- `--pragmatic` now drops 4 linters (was 5): `exhaustruct` graduated from `PragmaticNoiseLinters` to `NeverAutoEnableLinters` (handled unconditionally now, not just with the flag)
- `reference` preset now 61 linters (was 62): `exhaustruct` removed; `exhaustruct` priority downgraded High → Medium
- Sidecar policy enforcer (`isToolLevelDisabled` → `isToolLevelManaged`): now exempts both `DisabledLinters` and `NeverAutoEnableLinters` from re-enable enforcement, so a manually-disabled `exhaustruct` is respected

## [0.5.0] - 2026-07-23

### Added

- `audit` subcommand to query the config-mutation ledger (`--json`, `--since`, `--linter`, `--clear`)
- Audit ledger: append-only JSONL record of every config mutation in the OS cache dir, with 90-day retention purge
- Disable-reason policy enforcement via an opt-in `.golangci-lint-auto-configure.yml` sidecar (anti-gaming: re-enables linters in `linters.disable` that lack a justification entry)
- Detection findings now aligned with the repairer's priority threshold, preventing lower-priority linters from surfacing as critical findings

### Changed

- `depguard` linter disabled repo-wide — superseded by the dedicated `library-policy` tool (AST-based banned-library governance)
- CI workflows hardened: cancel-in-progress, `paths-ignore`, `timeout-minutes`
- Auto-tag workflow gated to version files only (`flake.nix`, `package.nix`)
- Error inspection modernized for the Go json/v2 era (`errors.AsType` generic assertions)
- LF line endings enforced across the repository

### Fixed

- `repair`/`configure` no longer re-enables user-disabled linters — root cause was `updateConfigFromSets` rebuilding `linters.disable` from scratch on every run; orphaned `settings.<linter>` blocks are now pruned too
- Optional boolean report fields (`LinterInfo.Fast`, `LinterInfo.AutoFix`, `FormatterInfo.AutoFix`) omit `false` again under `encoding/json/v2` by using `omitzero` instead of `omitempty`

## [0.4.0] - 2026-07-16

### Fixed

- Formatter and linter detection now gated by redundancy and project context, avoiding spurious auto-detection

### Changed

- Categorizer helper methods extracted for `funlen` compliance

## [0.3.0] - 2026-07-16

A large release consolidating the SUPERB quality sprint, the json/v2 + structured-error migrations, the PascalCase report-types refactor, and the P0–P3 architecture/data-accuracy fixes.

### Added

- `--check` flag for CI integration: exit 0 if config is optimal, exit 1 if changes are needed
- `--diff` flag to preview config changes before applying
- `--quiet` flag for CI output suppression (errors only)
- `--json-errors` flag for structured JSON error output (canonical snake_case schema via `errorfamily.JSON()`)
- `presets` command to list all available presets
- `reference` preset (60+ critical and high priority linters)
- `format` preset (core formatters + essential linters)
- Version-gated deprecation system (`LinterMinVersions`) for linters requiring specific golangci-lint versions
- `output.formats: {}` initialization in default config to prevent nil map issues
- swaggo formatter auto-detection (6 annotation patterns)
- Benchmarking suite for analyzer and fixer performance testing
- `pkg/version` package with `runtime/debug.ReadBuildInfo()` fallback
- Auto-tag workflow for automated releases on merge to master
- Multiple golangci-lint binary detection with version warnings
- Integration tests verifying PascalCase JSON output for `analyze` and `report` commands
- Serialization tests for PascalCase report types and kebab-case config types
- Fuzz tests for config merger (`FuzzMergeConfigInto`, `FuzzMergeIdempotent`) and property tests for Set algebra
- `Config.Clone()` deep-copy method replacing JSON marshal/unmarshal hack
- `configChangeRecorder` closure-based mutation counting for fixer normalization
- Typed linter/formatter settings structs with `SettingsConverter` interface
- Semantic BSD sysexits exit codes via go-error-family (Rejection/Conflict/Corruption/Infrastructure/Transient)
- Sentinel error classification registry (`pkg/errors/classification.go`)
- `DisabledLinters` as `map[LinterName]string` (linter name → disable reason)
- P0–P3 architecture & data-accuracy fixes: added `clickhouselint`, moved `exportloopref` to deprecated, replaced hardcoded version whitelist, typed settings, recorder pattern

### Changed

- **BREAKING:** `analyze --format json` and `report --format json` now emit **PascalCase** JSON keys (e.g. `ConfigPath`, `EnabledLinters`) instead of snake_case/camelCase
- Struct tag case policy enforced by tagliatelle: report types use PascalCase (tag-free), config types use kebab-case (round-trip `.golangci.yml` schema)
- Config types extracted from `pkg/types/types.go` into `pkg/types/config_types.go` for precise tagliatelle enforcement
- **BREAKING:** All `fmt.Errorf` calls migrated to `go-error-family` structured errors (`errorfamily.Wrap*` constructors)
- **BREAKING:** `encoding/json` v1 migrated to `encoding/json/v2` (`GOEXPERIMENT=jsonv2` required); wire-format decoupling structs added to match golangci-lint's JSON output
- `cli.Version` is now self-initializing from `version.Get().Short()` (single source of truth)
- Test runner uses `go run github.com/onsi/ginkgo/v2/ginkgo` to ensure version matches go.mod
- `charmbracelet/fang` migrated from v1 to `charm.land/fang/v2`
- `showDiff` refactored from package-level variable to threaded parameter
- `--json-errors` output switched to `errorfamily.JSON()` canonical snake_case schema

### Fixed

- go-finding panics on `FixStrategyDirect` and invalid tags replaced with proper error returns
- Docker build now injects version ldflags (previously showed `dev`)
- Eliminated Ginkgo CLI/library version mismatch warning
- Eliminated version split brain (3 sources of truth reduced to 1)
- go-finding API compatibility: `Merge` → `Combine`
- Registered 10 previously-unclassified sentinel errors with the error family system
- JSON report file permission hardened from `0o644` (world-readable) to `0o600`
- Cleared all 34 golangci-lint violations

### Removed

- `main.version` global variable (replaced by `pkg/version` package)
- `cli.VersionInfo` global variable (unused)
- `noinlineerr` linter disabled (conflicts with code formatters)
- `justfile` (superseded by the existing `flake.nix`)

## [0.2.0] - 2026-06-03

### Added

- Comprehensive default linter settings injection (depguard, ireturn, gocritic, exhaustruct, revive, varnamelen, gomoddirectives, cyclop)
- Default formatter settings (golines max-len: 120)
- Default exclusion paths for `_templ.go` and `vendor/`
- Default test exclusion rules (6 linters for `_test.go` files)
- gogenfilter/v3 integration for dynamic auto-generated file detection (templ, protobuf, wire, moq, mockgen, stringer, sqlc, oapi-codegen)
- HTML and JSON report generation tests
- `gitleaks` allowlist support
- `FEATURES.md` and `TODO_LIST.md` project tracking
- Pre-commit hook installation command
- Status report documentation system

### Changed

- Error types refactored from type aliases to distinct structs for `errors.As` discrimination
- `updateExclusionRules` simplified from O(n\*m) to O(n+m) using set-based lookup
- Nix flake modernized with fileset-based source filtering
- Dependency management files added for better reproducibility

### Fixed

- exhaustruct, gci, gocritic, and golines lint violations resolved
- Nix build uses shortRev for readable store paths
- All semantic code clones eliminated at threshold 45+

### Removed

- 3 dead backward-compat validation functions from `pkg/types/validation.go`
- Redundant `newFixCounts()` constructor (Go zero-init is sufficient)

## [0.1.0] - 2026-01-01

### Added

- CLI with 7 subcommands: configure, analyze, validate, report, migrate, install-hook, completion
- Linter priority system with 119 linters across 4 tiers (Critical, High, Medium, Optional)
- Project type detection (CLI, Library, Web, API, Monorepo)
- v1 to v2 config migration (merged from golangci-config-migrator)
- HTML/JSON/SARIF report generation
- go-finding integration for unified static analysis model
- Nix flake build system
- BDD test suite with Ginkgo/Gomega
- Pre-commit hook support
- Auto-merge for multiple config files
- Deprecated linter auto-replacement (e.g., `wsl` → `wsl_v5`)
- MIT license
