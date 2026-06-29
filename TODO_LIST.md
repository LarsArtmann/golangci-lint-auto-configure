# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-06-29

---

## High Priority

- [ ] Increase CLI integration test coverage (currently ~9%)
- [ ] Increase gogenfilter scanner coverage (currently 63.9%)
- [ ] Add exit-code integration tests for full `Main()` → `os.Exit()` path

## Medium Priority

- [ ] Add `--check` mode integration tests (exit codes, flag combinations)
- [ ] Add `--diff` flag integration tests
- [ ] Fix `--diff` + `--check` interaction (diff shows nothing in check mode)
- [ ] Add `LinterMinVersions` validation test (ensure all entries exist in `LinterPriorities`)
- [ ] Validate `reference` preset against `LinterPriorities`
- [ ] Add `testifylint` default settings
- [ ] Adopt `HandleError` at CLI boundary (replaces slog — separate decision)
- [ ] Add `--json` error output flag (`errorfamily.Error.JSON()`)

## Low Priority

- [ ] Add `Config.Clone()` method (replace JSON marshal/unmarshal hack)
- [ ] Add `pkg/client` smoke tests (or resolve intent: public API vs internal)
- [ ] Add `DryRun bool` field on `MigrationResult` (clarify "would fix" vs "did fix")
- [ ] Register domain message templates when adopting `errorfamily.New()` constructors

## Completed

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
