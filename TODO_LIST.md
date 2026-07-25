# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-25

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`; long-term ideas live in `ROADMAP.md`.

---

## High Priority

| Task                                                                                            | Impact                                             | Effort  | Evidence                                                                                  |
| ----------------------------------------------------------------------------------------------- | -------------------------------------------------- | ------- | ----------------------------------------------------------------------------------------- |
| Split `cmd_configure.go` (581 lines, 8+ concerns) into focused files                           | High — largest SRP violation in the codebase       | 3–4h    | `internal/cli/cmd_configure.go` is 581 lines; fixer was already split, configure was not  |
| Extend docs-integrity test to cover ALL hardcoded counts in FEATURES.md (not just presets)      | High — prevents all documentation drift            | 1h      | `pkg/constants/docs_integrity_test.go` infrastructure exists; only preset counts covered  |
| Add a separate `golangci-lint run` (no `--fix`) CI step                                         | Medium-High — catches issues BuildFlow `--fix` hides | 30min   | No `golangci-lint run` step in `.github/workflows/ci.yml`                                 |
| Reconcile funlen defaults: code injects 60/40, project config uses 30/20                        | Medium — values mismatch between tool and self      | 30min   | `pkg/constants/linter_settings.go` FunlenSettings vs `.golangci.yml` funlen config        |

## Medium Priority

| Task                                                                  | Impact                                                       | Effort  | Evidence                                                                        |
| --------------------------------------------------------------------- | ------------------------------------------------------------ | ------- | ------------------------------------------------------------------------------- |
| Adopt `HandleError` at the CLI boundary (replaces slog)              | Medium — structured error output at the system boundary       | 1–2h    | No `HandleError` in `internal/cli/` today                                       |
| Add `--no-color` flag for CI/scripting output                         | Medium — enables accurate plain-text examples in docs         | 1h      | README example output is hand-simplified because real output has ANSI codes     |
| Pin golangci-lint version in CI to match devShell (v2.12.2)           | Medium — prevents version drift between local and CI          | 30min   | CI uses `go-version-file: go.mod` for Go, but golangci-lint version is unpinned |
| Convert `scripts/coverage-check.sh` to a Go test                     | Medium — more portable, runs in `go test`                    | 1h      | `scripts/coverage-check.sh` is still bash                                       |
| Full README.md claim-by-claim audit (all ~500 lines)                 | Medium — only ~100 lines audited; minor drift may remain     | 2h      | README.md is ~500 lines; this session fixed Requirements, output, CI, links     |
| Extract ARCHITECTURE.md inline ADRs to individual `docs/adr/` files  | Medium — ADRs live in two places (split brain)                | 1–2h    | `docs/adr/` has 5 files; ARCHITECTURE.md has 8 inline ADRs                      |
| Add markdown linter (`markdownlint-cli2`) to Nix devShell and CI     | Medium — catches broken `<details>` blocks, link rot          | 1h      | No markdown linter in `flake.nix` or CI                                         |
| Add property-based JSON round-trip tests for report types            | Medium — catches serialization regressions                    | 2h      | Report types use json/v2 omitzero; no round-trip property tests exist           |
| Add HTML report snapshot/golden tests                                | Medium — guards against silent templ regressions              | 1–2h    | `pkg/report/report.templ` has no golden tests                                   |

## Low Priority

| Task                                                                        | Impact | Effort  | Evidence                                                   |
| --------------------------------------------------------------------------- | ------ | ------- | ---------------------------------------------------------- |
| Register domain message templates for `errorfamily.New()` constructors      | Low    | 30min   | `pkg/errors/classification.go` uses bare sentinels         |
| Build an error-code governance registry (~40 ad-hoc exit codes, no test)    | Low    | 2h      | Exit codes scattered across CLI commands; no central map   |
| Consolidate `ValidationError` + `HealthIssue` (overlapping types)           | Low    | 1–2h    | `pkg/types/types.go:199` and `pkg/types/validation.go:118` |
| Split the composite `ConfigLoader` God Object interface (8+ methods)        | Low    | 2–3h    | `pkg/types/types.go:223` — combines 6 sub-interfaces       |
| Add `flake.lock` drift detection to CI                                      | Low    | 30min   | No drift check in `.github/workflows/ci.yml`               |
| Consolidate/archive the 100+ July status reports in `docs/status/`          | Low    | 1h      | `docs/archive/status/` + `docs/status/` have 100+ files    |
