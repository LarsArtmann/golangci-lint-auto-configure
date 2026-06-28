# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-06-28

---

## Critical Priority

- [ ] Update `flake.nix` vendorHash for go-error-family dependency
- [ ] Run `nix build` + `nix flake check` to verify full Nix pipeline

## High Priority

- [ ] Propagate error-family classification to all CLI command handlers
- [ ] Increase CLI integration test coverage (currently 9.0%)
- [ ] Increase gogenfilter scanner coverage (currently 63.9%)

## Medium Priority

- [ ] Add `--check` mode integration tests (exit codes, flag combinations)
- [ ] Add `--diff` flag integration tests
- [ ] Fix `--diff` + `--check` interaction (diff shows nothing in check mode)
- [ ] Add `LinterMinVersions` validation test (ensure all entries exist in `LinterPriorities`)
- [ ] Validate `reference` preset against `LinterPriorities`
- [ ] Decide whether `vendor/` should be in formatter exclusions
- [ ] Add `ginkgolinter` default settings if any exist
- [ ] Add `testifylint` default settings

## Low Priority

- [ ] Add `Config.Clone()` method (replace JSON marshal/unmarshal hack)
- [ ] Add `pkg/client` smoke tests (or resolve intent: public API vs internal)
- [ ] Use `errors.Join` for multi-finding failures (currently returns first error only)
- [ ] Add `DryRun bool` field on `MigrationResult` (clarify "would fix" vs "did fix")

## Completed

- [x] Fix go-finding v1.0.0 API breakage (branded RuleName/ToolName types, unexported Report.findings)
- [x] Integrate go-error-family for semantic BSD sysexits exit codes
- [x] Rewrite AGENTS.md to be lean (~75 lines) with corrected commands (no justfile exists)
- [x] Remove stale `just` command references from all living docs
- [x] Trim AGENTS.md from 912 to ≤377 lines (current: 373)
- [x] Increase migration coverage from 66.8% to 75.5%
- [x] Fix gofumpt formatting in integration_test.go

- [x] Fix all lint violations — round 2 (7 funlen, 1 noinlineerr, 1 varnamelen → 0 issues)

- [x] Fix error types: use distinct structs instead of type aliases for errors.As discrimination
- [x] Remove 3 dead backward-compat validation functions from pkg/types/validation.go
- [x] Simplify updateExclusionRules from O(n\*m) to O(n+m) using set-based lookup
- [x] Remove redundant newFixCounts() constructor (Go zero-init is sufficient)
- [x] Add report package tests (HTML and JSON generators)
- [x] Migrate charmbracelet/fang v1 → v2 (charm.land/fang/v2)
- [x] Enrich CreateDefaultConfig() with default settings, exclusion paths, rules
- [x] Add comprehensive default linter settings (revive, varnamelen, gomoddirectives, cyclop)
- [x] Add default formatter settings (golines max-len: 120)
- [x] Add default exclusion rules for test files
- [x] Add default exclusion paths for \_templ.go and vendor/
- [x] Fix all lint violations (funlen, varnamelen, exhaustruct, gci, golines, gocritic)
- [x] Create FEATURES.md feature audit
- [x] Remove `report_templ.go` from git tracking (generated file)
- [x] Add `output.formats: {}` to default config
- [x] Add `reference` preset (60+ critical + high priority linters)
- [x] Add swaggo formatter auto-detection (6 annotation patterns)
- [x] Add benchmarking for analyzer and fixer
- [x] Document exclusion pattern syntax (RE2 regex) in README
- [x] Add `--check` mode for CI (exit 0 if optimal, exit 1 if changes needed)
- [x] Add `--diff` flag to show config changes before applying
