# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-10

---

## High Priority

- [ ] Increase CLI integration test coverage (currently ~11% — integration tests exec the binary, which doesn't count toward go coverage)
- [ ] Add exit-code integration tests for Infrastructure (69) and Corruption (65) paths

## Medium Priority

- [ ] Adopt `HandleError` at CLI boundary (replaces slog — separate decision)
- [ ] Add `--diff` integration tests for addition/removal formatting
- [ ] Convert coverage-check.sh to a Go test (more portable, testable)

## Low Priority

- [ ] Register domain message templates when adopting `errorfamily.New()` constructors

## Completed

- [x] Migrate encoding/json v1 → v2 across all files (GOEXPERIMENT=jsonv2 enabled in flake.nix, CI workflows; wire-format decoupling structs added)
- [x] Upgrade go-finding v1.0.0 → v1.2.0 (branded FilePath types, Position.Line validation)
- [x] Fix `--diff` + `--check` interaction (was running fmt on restored config)
- [x] Add govulncheck security scanning to CI
- [x] Add per-package coverage threshold gate to CI (scripts/coverage-check.sh)
- [x] Add fuzz + property tests for Set operations (commutative, idempotent, subset)
- [x] Add fuzz test for config merger (FuzzMergeConfigInto + FuzzMergeIdempotent targeting merger.go, not types.Set)
- [x] Refactor showDiff from package-level variable to parameter (all call sites threaded)
- [x] Switch --json-errors to errorfamily.JSON() canonical output (snake_case, richer schema)
- [x] Add --quiet flag for CI output suppression (PersistentPreRunE log level)
- [x] Remove dead ConfigPath branded type (unused throughout codebase)
- [x] Add unit tests for cmd_configure internal functions (effectiveDryRunForCheckDiff, resolvePreset, etc.)
- [x] Add gogenfilter utility tests (MergeExclusionPaths, ExclusionPaths, shouldSkipDir, String)
- [x] Add `--json-errors` flag for structured JSON error output
- [x] Add exit-code integration tests (exit 0, exit 75 Transient, JSON output)
- [x] Add pkg/client smoke tests (New, LoadConfig, ValidateConfig, SaveConfig, SimpleFix)
- [x] Add `Config.Clone()` method (replace JSON marshal/unmarshal hack)
- [x] Add `DryRun bool` field on `MigrationResult`
- [x] Add `LinterMinVersions` validation test (ensure all entries exist in `LinterPriorities`)
- [x] Validate `reference` preset against `LinterPriorities`
- [x] Add `testifylint` default settings
- [x] Fix nix build: remove preBuild (commit \_templ.go), conditional go mod tidy in FOD only
- [x] Update `flake.nix` vendorHash for go-error-family dependency
- [x] Run `nix build` + `nix flake check` to verify full Nix pipeline
- [x] Fix go-finding v1.0.0 API breakage (branded RuleName/ToolName types)
- [x] Integrate go-error-family for semantic BSD sysexits exit codes
- [x] Implement `Classified` on ConfigError, ReportError, MigrationError
- [x] Document AnalysisError delegation decision (cause-chain classified)
- [x] Clear all 34 golangci-lint violations (makezero, noctx, nilerr, funlen, wrapcheck, etc.)
- [x] Fix FEATURES.md justfile lie (no justfile exists)
- [x] Document exit codes in README.md for CI/CD consumers
- [x] Verify all fmt.Errorf calls use %w (they do — 29-count was a grep artifact)
- [x] errors.Join already used for multi-finding in converter.go
- [x] Rewrite AGENTS.md to be lean (~75 lines) with corrected commands
- [x] Remove stale `just` command references from all living docs
- [x] Add `reference` preset (60+ critical + high priority linters)
- [x] Add `--check` mode for CI (exit 0 if optimal, exit 1 if changes needed)
- [x] Add `--diff` flag to show config changes before applying
- [x] P0: Add clickhouselint, move exportloopref to deprecated, fix validVersions
- [x] P1: Add missing v1 removed linters and alternative names
- [x] P2: Add cross-map data integrity tests
- [x] P3.1: Typed linter/formatter settings structs with SettingsConverter interface
- [x] P3.2: configChangeRecorder closure-based mutation counting
- [x] P3.3: Add format preset (minimal linters + core formatters)
