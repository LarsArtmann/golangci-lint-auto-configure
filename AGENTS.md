# Agent Guide: golangci-lint-auto-configure

Concise, enduring context for AI sessions. For detail, see the linked docs at the bottom.

## What This Is

Go CLI that auto-configures and optimizes golangci-lint configs: analyzes, detects missing linters, recommends settings, auto-fixes, replaces deprecated linters, migrates v1→v2, and emits HTML/JSON/SARIF/finding reports.

## Commands (IMPORTANT — there is NO justfile)

The build system is **Nix flake + plain Go tooling**. Do not use `just` — older docs referencing it are stale.

```bash
# Build
nix build                              # reproducible build (preferred)
go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure

# Test (Ginkgo BDD specs, but run via go test)
go test -race ./pkg/... ./internal/...
ginkgo -r --cover                      # alternative: ginkgo directly

# Lint
golangci-lint run --config=.golangci.yml --timeout=5m

# Format (treefmt via Nix)
nix fmt                                # format Go, Nix, templ
nix flake check                        # runs format check + build + tests

# Dev shell (provides Go 1.26, golangci-lint, templ, gopls, etc.)
nix develop
```

## Tech Stack (non-obvious points)

- **Go 1.26**, Cobra CLI, charm.land log/lipgloss/fang, templ HTML reports, `go.yaml.in/yaml/v3`.
- **`GOEXPERIMENT=jsonv2` is required.** The codebase uses `encoding/json/v2` + `encoding/json/jsontext` (experimental stdlib in Go 1.26). The flake devShell, package build, and CI shell all set `GOEXPERIMENT=jsonv2` in `env`. Without it, `go build`/`go test` fail with "build constraints exclude all Go files in encoding/json/v2". buildflow auto-enables it for its own processes; when running `go` commands directly outside the devShell, export it manually.
- **json/v2 wire-format decoupling.** golangci-lint's JSON output uses **capitalized wrapper keys** (`"Enabled"`, `"Disabled"`) but **lowercase field keys** (`"name"`, `"description"`, `"autoFix"`). json/v2 is case-sensitive (unlike v1). Report types (`LinterInfo`, `FormatterInfo` in `pkg/types/types.go`) are tag-free (PascalCase for JSON reports). Dedicated wire-format structs in `pkg/linter/analyzer.go` (`golangciLinterEntry`, `golangciFormatterEntry`) match golangci-lint's wire format and convert to Report types after parsing.
- **Testing: Ginkgo v2 + Gomega (BDD)** — NOT standard `testing` style. Specs use `Describe`/`Context`/`It` + Gomega matchers. See `docs/references/testing-style-and-patterns.md`.
- **gogenfilter/v3**: auto-detects generated files to exclude from linting.
- **go-finding**: unified finding model (SARIF/JSON output).
- **go-error-family** (`v0.7.0`): structured error classification. Sentinel errors are registered with Families (Rejection/Conflict/Transient/Corruption/Infrastructure) in `pkg/errors/classification.go`. `ConfigError`, `ReportError`, and `MigrationError` implement the `Classified` interface (always Rejection). `AnalysisError` delegates to cause-chain sentinels for fine-grained Families. `Main()` uses `errorfamily.ExitCode(err)` for BSD sysexits exit codes instead of hardcoded `os.Exit(1)`.
- **Audit ledger** (`pkg/audit/`): append-only JSONL ledger in the OS cache dir (`~/.cache/golangci-lint-auto-configure/audit.jsonl`) that records every config mutation (linter enable/disable, formatter changes, settings pruning, re-enables). 90-day retention purge runs at the start of each configure run. Queried via the `audit` CLI subcommand (`--json`, `--since`, `--linter`, `--clear`). Disable with `--no-audit` flag or `GOLANGCI_LINT_AUTO_CONFIGURE_NO_AUDIT` env var.
- **Policy enforcement** (`pkg/policy/`): disable-reason sidecar file (`.golangci-lint-auto-configure.yml`) that justifies intentional linter disables. When a sidecar exists, the fixer re-enables linters in `linters.disable` that lack a justification entry (anti-gaming enforcement). Tool-level disabled linters (`constants.DisabledLinters`) are always exempt. When no sidecar exists, all disables are respected (backward compatible).

## Critical Gotchas (read these — they bite)

1. **No justfile.** Despite older docs, all `just <cmd>` references are stale. Use Nix/Go directly (see Commands above).

2. **templ output is committed.** `_templ.go` files are generated from `.templ` files and committed to git (un-ignored in `.gitignore`). After editing `pkg/report/report.templ`, run `templ generate` manually before `go build`. The Nix build no longer generates templ output (it uses the committed file directly).

3. **vendorHash update after go.mod changes.** `nix build` will fail with a hash mismatch. Procedure (run inside `nix develop` or export `GOEXPERIMENT=jsonv2` first — the code imports `encoding/json/v2` which won't compile without it):

   ```bash
   go mod tidy
   nix build 2>&1 | rg "got:"   # copy the got: sha256
   # paste into flake.nix vendorHash
   nix build                     # rebuild
   ```

4. **go-finding & gogenfilter replace in Nix.** `go.mod` uses published versions. Nix's `mkPreparedSource` (from go-nix-helpers) injects `replace` directives pointing to SSH-fetched local copies (the Go proxy doesn't cache these repos). `go mod tidy` runs ONLY in the go-modules FOD (has network via `__noChroot`); the main derivation sets `GOFLAGS=-mod=mod` to auto-reconcile from the FOD's proxy cache. Config-level findings use `Line: 1` (go-finding requires `Position.Line > 0`).

5. **Error classification via go-error-family.** `pkg/errors/classification.go` has an `init()` that registers all sentinel errors with their `errorfamily.Family`. To add a new sentinel: add it to the map in that file. `ConfigError`, `ReportError`, and `MigrationError` implement `Classified` → `Rejection` (type-level, checked before sentinels). `AnalysisError` does NOT implement `Classified` — its sentinels in the cause chain (e.g. `ErrVersionTooOld`) handle classification. Exit codes: Rejection/Conflict → 1, Transient → 75, Corruption → 65, Infrastructure → 69. `--json-errors` outputs via `errorfamily.Wrap().JSON()` (snake_case canonical schema with family/code/message/context/retryable).

6. **Fixer normalization counting.** Config mutations go through `configChangeRecorder` (`pkg/linter/fixer_recorder.go`), which wraps each mutation in a closure and auto-increments the counter. Used consistently in both `applyAllFixes` and `applyAndSave`. The `counts.total()==0` guard silently discards changes if the counter is zero. The fixer also injects `issues.max-issues-per-linter: 50` and `max-same-issues: 10` when absent (prevents golangci-lint's default `max-same-issues: 3` from hiding CI problems).

7. **Auto-injected linter safe defaults.** Linters that misbehave without explicit config get safe defaults injected by `injectDefaultSettings` (`pkg/linter/fixer_config.go`) when enabled and missing settings. Typed settings structs live in `pkg/constants/linter_settings.go` (`DefaultLinterSettings`, `DefaultFormatterSettings`) with compile-time-safe fields and yaml tags. Each struct implements `SettingsConverter.ToMap()` for config injection. Idempotent; never overwrites existing user settings.

8. **Config auto-creation requires a git repo.** `configure` creates a default `.golangci.yml` if missing. Git is mandatory for version-control safety.

9. **DI is manual.** No `internal/di/` directory (older docs lie). Dependencies wired manually in CLI commands. No DI framework.

10. **Linter priority data is static constants.** Priorities live in `pkg/constants/linter_priorities.go` + reasons in `linter_reasons.go`. Not dynamically computed from golangci-lint. To add/change a linter, edit both files (and `presets.go`/`rules.go` if relevant). `DisabledLinters` (`rules.go`) is `map[LinterName]string` mapping linter name → disable reason. Disabled linters must never have entries in `LinterPriorities` or `LinterReasons` — enforced by data integrity tests + `scripts/validate_linter_data.go`.

11. **Versioning is self-initializing.** `pkg/version/` reads ldflags with `runtime/debug.ReadBuildInfo()` fallback. `cli.Version` self-inits — no manual setup. All build targets (Nix, CI) inject via ldflags.

12. **Struct tag case policy.** Enforced by **tagliatelle** in `.golangci.yml` (`json: pascal`, `yaml: kebab`, `toml: kebab`). Three type families:

    | Family                                                                                                                                                                 | `json`/`cbor`                                               | `yaml`/`toml`           | Why                                                       |
    | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- | ----------------------- | --------------------------------------------------------- |
    | **Report types** (`pkg/types/types.go`: `ConfigAnalysis`, `LinterInfo`, …; `pkg/report/json_report_generator.go`: `JSONReport`; `pkg/config/merger.go`: `MergeResult`) | **PascalCase** — tags stripped, Go field names pass through | n/a (no yaml/toml tags) | Zero-tag-cost: Go-native, no browser consumers            |
    | **Config types** (`pkg/types/config_types.go`: `Config`, `RunConfig`, `LintersConfig`, …)                                                                              | **kebab**                                                   | **kebab** (locked)      | Round-trips `.golangci.yml` schema; one struct = one case |
    | **External format types** (`LinterList`, `golangciLintOutput`, `golangciLintVersion`, `GolangciLintIssue`)                                                             | **as-is** (matches golangci-lint wire format)               | n/a                     | Must match external JSON shape; tagliatelle-excluded      |

    **musttag linter:** `Marshal`/`Unmarshal` of tag-free report structs need `//nolint:musttag`. Test files are excluded from musttag in `.golangci.yml`.

13. **`.buildflow.yml` skips nixfmt-standalone.** `nixfmt-standalone` runs raw `nixfmt .` which ignores buildflow's exclude patterns and scans `.direnv/flake-inputs/` (symlinked third-party nix caches). It fails ~88% of the time. The `nix-fmt` step (treefmt) handles Nix formatting correctly and respects excludes. GitHub Actions workflows set `GOEXPERIMENT: jsonv2` at the job level for all Go-compiling jobs (`test-and-build`, `lint`, `govulncheck`, `release`).

14. **`WrapClassified` returns `*errorfamily.Error`, not `error`.** Callers that pass a potentially-nil error MUST check for nil first (`if err == nil { return nil }`) before calling `WrapClassified`/`WrapClassifiedf`. Otherwise, nil `*errorfamily.Error` gets boxed into a non-nil `error` interface (the typed-nil interface pitfall), causing the caller's error check to always be true. Two bugs from this were fixed: `RunFmtCommand` always reported failure, and `validate` never reported success. The return type was kept as `*errorfamily.Error` (intentional) so callers can access classified methods.

15. **Disable-reason enforcement is opt-in via sidecar.** When `.golangci-lint-auto-configure.yml` exists next to `.golangci.yml`, the fixer re-enables any linter in `linters.disable` that lacks a justification entry in the sidecar. Without a sidecar, all disables are respected. This prevents AI agents from silently disabling linters to game the lint gate. Tool-level disabled linters (`constants.DisabledLinters`: funcorder, noinlineerr, depguard) are always exempt from enforcement.

16. **Audit ledger is best-effort and lives outside the git tree.** The JSONL ledger at `~/.cache/golangci-lint-auto-configure/audit.jsonl` records every config mutation. Failures to write are logged but never block the configure run. Dry runs never emit entries. Query with the `audit` subcommand; purge with `--clear` or automatically via 90-day retention.

17. **json/v2 `omitempty` no longer omits `false`/`0`.** Under `encoding/json/v1`, `omitempty` dropped false bools, zero ints, and empty strings. Under `encoding/json/v2` it only omits nil pointers/interfaces and empty slices/maps/arrays/strings. To omit Go zero values (false, 0, ""), use **`omitzero`** instead. Report types (`LinterInfo.Fast`, `LinterInfo.AutoFix`, `FormatterInfo.AutoFix` in `pkg/types/types.go`) use `omitzero` for exactly this reason. Known latent follow-up: `pkg/types/config_types.go` still uses `bool`/`int + omitempty` on its `json:` tags — these emit `false`/`0` in **JSON-format** config output (`marshalConfig`, `pkg/config/loader.go`) but YAML output (the default) is unaffected. See `docs/references/json-v2.md`.

## Where to Find Detail

| Topic                                                  | Location                                        |
| ------------------------------------------------------ | ----------------------------------------------- |
| Directory structure & patterns                         | `docs/references/code-organization.md`          |
| Adding commands/linters, common tasks, troubleshooting | `docs/references/working-with-codebase.md`      |
| BDD testing, code style, CI/CD                         | `docs/references/testing-style-and-patterns.md` |
| Error handling patterns                                | `docs/references/error-handling.md`             |
| json/v2 migration & behavioral changes                 | `docs/references/json-v2.md`                    |
| gogenfilter & go-finding integration                   | `docs/references/integrations.md`               |
| User-facing usage                                      | `README.md`                                     |
| Feature inventory                                      | `FEATURES.md`                                   |
| Open work                                              | `TODO_LIST.md`                                  |
| Domain language                                        | `docs/DOMAIN_LANGUAGE.md`                       |
