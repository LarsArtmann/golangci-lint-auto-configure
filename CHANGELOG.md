# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `forcetypeassert` added to the default `_test.go` exclusion rules

### Changed

- golangci-lint configuration and linter constants updated
- Nix flake configuration and module dependencies updated

## [0.5.0] - 2026-07-23

### Added

- `audit` subcommand to query the config-mutation ledger (`--json`, `--since`, `--linter`, `--clear`)
- Audit ledger: append-only JSONL record of every config mutation in the OS cache dir, with 90-day retention purge
- Disable-reason policy enforcement via an opt-in `.golangci-lint-auto-configure.yml` sidecar (anti-gaming: re-enables linters in `linters.disable` that lack a justification entry)

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
