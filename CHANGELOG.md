# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Structured versioning via `pkg/version` package with version, commit, date, and tree state
- `runtime/debug.ReadBuildInfo()` fallback for automatic VCS detection when ldflags are not set
- All build targets (justfile, Nix, Docker) inject version metadata via ldflags
- Rich `--version` output showing Version, Commit, Built date, and Tree state
- BDD test suite for version package (6 specs)

### Changed

- `cli.Version` is now self-initializing from `version.Get().Short()` (single source of truth)
- Test runner uses `go run github.com/onsi/ginkgo/v2/ginkgo` to ensure version matches go.mod
- Nix devShell installs ginkgo from go.mod instead of nixpkgs

### Fixed

- Nix vendorHash updated for new `pkg/version/` package
- Docker build now injects version ldflags (previously showed `dev`)
- Eliminated Ginkgo CLI/library version mismatch warning (nixpkgs v2.28.1 vs go.mod v2.28.3)
- Eliminated version split brain (3 sources of truth reduced to 1)

### Removed

- `main.version` global variable (replaced by `pkg/version` package)
- `cli.VersionInfo` global variable (unused)

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
