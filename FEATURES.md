# golangci-lint-auto-configure — Feature Audit

**Version:** v0.5.0+ (unreleased changes)
**Last Audited:** 2026-07-25

Status vocabulary: `FULLY_FUNCTIONAL` · `PARTIALLY_FUNCTIONAL` · `BROKEN` · `PLANNED`.

---

## CLI Commands

| Feature          | Command        | Status           | Notes                                                                                                    |
| ---------------- | -------------- | ---------------- | -------------------------------------------------------------------------------------------------------- |
| Auto-configure   | `configure`    | FULLY_FUNCTIONAL | Enables recommended linters, applies fixes, backs up config                                              |
| List presets     | `presets`      | FULLY_FUNCTIONAL | Lists all presets with descriptions                                                                      |
| Analyze config   | `analyze`      | FULLY_FUNCTIONAL | Reports missing/extra linters, supports SARIF/finding JSON                                               |
| Validate config  | `validate`     | FULLY_FUNCTIONAL | Checks YAML validity, supports SARIF output                                                              |
| Generate report  | `report`       | FULLY_FUNCTIONAL | HTML, JSON, SARIF, finding report formats                                                                |
| Migrate v1→v2    | `migrate`      | FULLY_FUNCTIONAL | Migrates v1 configs to v2 format                                                                         |
| Query audit log  | `audit`        | FULLY_FUNCTIONAL | Query config-mutation ledger (`--json`, `--since`, `--linter`, `--clear`); tested in `cmd_audit_test.go` |
| Install hook     | `install-hook` | FULLY_FUNCTIONAL | Installs git pre-commit hook                                                                             |
| Shell completion | `completion`   | FULLY_FUNCTIONAL | bash, zsh, fish, powershell                                                                              |

## Auto-Configuration Features

| Feature                                                | Status           | Notes                                                                                                     |
| ------------------------------------------------------ | ---------------- | --------------------------------------------------------------------------------------------------------- |
| Linter priority system (Critical/High/Medium/Optional) | FULLY_FUNCTIONAL | `pkg/constants/linter_priorities.go` + `linter_reasons.go`                                                |
| Priority-based filtering (`--priority`)                | FULLY_FUNCTIONAL | configure command                                                                                         |
| Pragmatic mode (`--pragmatic`)                         | FULLY_FUNCTIONAL | Drops 5 highest-noise linters (exhaustruct, gochecknoglobals, wrapcheck, ireturn, funlen) from enable set |
| Dry-run mode (`--dry-run`)                             | FULLY_FUNCTIONAL | Shows what would change                                                                                   |
| CI check mode (`--check`)                              | FULLY_FUNCTIONAL | Exit 0 if optimal, exit 1 if changes needed                                                               |
| Diff preview (`--diff`)                                | FULLY_FUNCTIONAL | Shows config diff before applying (threaded as parameter, not global var)                                 |
| Deprecated linter auto-replacement                     | FULLY_FUNCTIONAL | wsl→wsl_v5, gomodguard→gomodguard_v2, etc.                                                                |
| Version-gated deprecation                              | FULLY_FUNCTIONAL | gomodguard_v2 requires v2.12.0+                                                                           |
| Typecheck linter removal                               | FULLY_FUNCTIONAL | Removes from enable/disable lists                                                                         |
| Invalid duration fix                                   | FULLY_FUNCTIONAL | Fixes empty/invalid timeout values                                                                        |
| Multiple binary detection                              | FULLY_FUNCTIONAL | Warns if multiple golangci-lint binaries                                                                  |

## Disable-Respect & Audit Policy

| Feature                              | Status           | Notes                                                                                                                         |
| ------------------------------------ | ---------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Preserve user `linters.disable`      | FULLY_FUNCTIONAL | `repair`/`configure` never re-adds a disabled linter; orphaned settings pruned (`fixer_config.go`)                            |
| Audit ledger (config-mutation JSONL) | FULLY_FUNCTIONAL | Append-only `~/.cache/.../audit.jsonl`; 90-day retention; tested in `ledger_test.go`                                          |
| Disable-reason sidecar enforcement   | FULLY_FUNCTIONAL | `.golangci-lint-auto-configure.yml` justifies disables (anti-gaming); `pkg/linter/fixer_enforce.go` + `fixer_enforce_test.go` |
| Tool-level disabled linters exempt   | FULLY_FUNCTIONAL | `constants.DisabledLinters`: funcorder, noinlineerr, depguard (`pkg/constants/rules.go`)                                      |

## Error Handling & Exit Codes

| Feature                            | Status           | Notes                                                                                     |
| ---------------------------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| Semantic exit codes (BSD sysexits) | FULLY_FUNCTIONAL | via go-error-family: Rejection(1), Conflict(1), Corruption(65), Infrastructure(69)        |
| Error classification registry      | FULLY_FUNCTIONAL | All sentinel errors mapped to families in `pkg/errors/classification.go`                  |
| `--json-errors` flag               | FULLY_FUNCTIONAL | Structured JSON via errorfamily.JSON() (snake_case keys matching SARIF ecosystem)         |
| `--quiet` flag                     | FULLY_FUNCTIONAL | Suppresses all output except errors (for CI pipelines)                                    |
| Exit-code test coverage            | FULLY_FUNCTIONAL | Exit 0, 1 (Rejection), 65 (Corruption), 69 (Infrastructure) tested in `exit_code_test.go` |

## Default Settings Injection

| Feature                                                  | Status           | Notes                                                                                                                                                           |
| -------------------------------------------------------- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Typed settings structs (SettingsConverter interface)     | FULLY_FUNCTIONAL | Compile-time safety in `pkg/constants/linter_settings.go`                                                                                                       |
| depguard defaults ($gostd, $module)                      | FULLY_FUNCTIONAL | Prevents deny-all default (depguard now tool-disabled; kept for reference)                                                                                      |
| ireturn defaults (error, empty, anon, stdlib, generic)   | FULLY_FUNCTIONAL | Reasonable interface return policy                                                                                                                              |
| gocritic defaults (ifElseChain disabled)                 | FULLY_FUNCTIONAL | Removes noisy checks                                                                                                                                            |
| exhaustruct defaults (14 stdlib structs excluded)        | FULLY_FUNCTIONAL | net/http.Client/Server/Request/Response/Transport/Cookie, net.TCPAddr/Dialer, slog.HandlerOptions, sync.WaitGroup, bytes.Buffer, time.Ticker/Timer, os/exec.Cmd |
| revive defaults (exported, package-comments disabled)    | FULLY_FUNCTIONAL | Noisy without config                                                                                                                                            |
| varnamelen defaults (short names, ignore flags)          | FULLY_FUNCTIONAL | Common short variable exemptions                                                                                                                                |
| gomoddirectives defaults (replace-local: true)           | FULLY_FUNCTIONAL | Local dev support                                                                                                                                               |
| cyclop defaults (max-complexity: 12)                     | FULLY_FUNCTIONAL | Reasonable complexity threshold                                                                                                                                 |
| funlen defaults (lines: 200, statements: 100)            | FULLY_FUNCTIONAL | House style; dominant override across 160 projects; diverges from upstream                                                                                      |
| mnd defaults (ignored-numbers: 0, 1, 2, 100)             | FULLY_FUNCTIONAL | Reduces magic-number noise for common values                                                                                                                    |
| golines formatter defaults (max-len: 120)                | FULLY_FUNCTIONAL | When enabled via lll replacement                                                                                                                                |
| gosec defaults (G304, G115 excluded)                     | FULLY_FUNCTIONAL | Curated excludes for common false-positive security findings                                                                                                    |
| errcheck defaults (exclude-functions for Close, Fprint*) | FULLY_FUNCTIONAL | Curated exclude-functions reducing defer/fmt noise                                                                                                              |
| output.formats initialization                            | FULLY_FUNCTIONAL | Empty map to prevent nil issues                                                                                                                                 |

## Exclusion Automation

| Feature                                                         | Status           | Notes                                                                                                                                                          |
| --------------------------------------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Default linter exclusion paths (\_templ.go$, .gen.go$, vendor/) | FULLY_FUNCTIONAL | Always injected                                                                                                                                                |
| Default formatter exclusion paths (\_templ.go$)                 | FULLY_FUNCTIONAL | Always injected                                                                                                                                                |
| Default test exclusion rules (14 linters for \_test.go)         | FULLY_FUNCTIONAL | exhaustruct, testpackage, gochecknoglobals, funlen, cyclop, goconst, forcetypeassert, gosec, errcheck, wrapcheck, ireturn, recvcheck, contextcheck, exhaustive |
| Default unused text exclusion for test files                    | FULLY_FUNCTIONAL | Suppresses unused false positives in tests                                                                                                                     |
| `generated: lax` auto-set                                       | FULLY_FUNCTIONAL | Both linters and formatters                                                                                                                                    |
| gogenfilter dynamic scan                                        | FULLY_FUNCTIONAL | Detects templ, protobuf, wire, moq, mockgen, stringer, sqlc, oapi-codegen                                                                                      |
| gogenfilter/v3 two-phase detection                              | FULLY_FUNCTIONAL | Filename first, content second                                                                                                                                 |
| Deduplication of exclusion paths                                | FULLY_FUNCTIONAL | MergeExclusionPaths                                                                                                                                            |

## Presets

| Feature                                                 | Status           | Notes                               |
| ------------------------------------------------------- | ---------------- | ----------------------------------- |
| `minimal` preset (5 linters)                            | FULLY_FUNCTIONAL | Essential only, fastest             |
| `standard` preset (8 linters)                           | FULLY_FUNCTIONAL | Good balance for most projects      |
| `strict` preset (20 linters)                            | FULLY_FUNCTIONAL | Maximum linting for CI/CD           |
| `security` preset                                       | FULLY_FUNCTIONAL | Security-focused only               |
| `performance` preset                                    | FULLY_FUNCTIONAL | Performance optimization            |
| `reference` preset (62 linters)                         | FULLY_FUNCTIONAL | All critical + high priority        |
| `format` preset (5 linters + 3 formatters)              | FULLY_FUNCTIONAL | Core formatters + essential linters |
| Auto-detect project type and select preset (`--detect`) | FULLY_FUNCTIONAL | CLI, web, library, API, monorepo    |

> Linter counts verified against `pkg/constants/presets.go` as of 2026-07-25.

## Formatter Management

| Feature                                                   | Status           | Notes                       |
| --------------------------------------------------------- | ---------------- | --------------------------- |
| Core formatters (gci, gofumpt, goimports)                 | FULLY_FUNCTIONAL | Always enabled              |
| golines auto-enable (when lll detected)                   | FULLY_FUNCTIONAL | Replaces redundant linter   |
| swaggo auto-detection                                     | FULLY_FUNCTIONAL | Detects swag annotations    |
| Redundant formatter removal (gofmt when gofumpt)          | FULLY_FUNCTIONAL | Superset detection          |
| Redundant linter removal (lll when golines)               | FULLY_FUNCTIONAL | Formatter supersedes linter |
| Formatter ordering (gci→goimports→gofumpt→golines→swaggo) | FULLY_FUNCTIONAL | Canonical order             |

## Build & Runner Settings

| Feature                                | Status           | Notes                                                     |
| -------------------------------------- | ---------------- | --------------------------------------------------------- |
| Go version auto-detection              | FULLY_FUNCTIONAL | Sets run.go to local version                              |
| GOEXPERIMENT build tags auto-injection | FULLY_FUNCTIONAL | arenas, goroutineleakprofile, jsonv2, runtimesecret, simd |
| allow-parallel-runners enablement      | FULLY_FUNCTIONAL | Always enabled                                            |
| allow-serial-runners enablement        | FULLY_FUNCTIONAL | Always enabled                                            |
| Version field fix (empty → "2")        | FULLY_FUNCTIONAL | Pre-flight check                                          |
| Benchmarking suite                     | FULLY_FUNCTIONAL | analyzer and fixer benchmarks                             |

## Migration (v1 → v2)

| Feature                                         | Status           | Notes                              |
| ----------------------------------------------- | ---------------- | ---------------------------------- |
| issues.exclude-rules → linters.exclusions.rules | FULLY_FUNCTIONAL | Path-based migration               |
| issues.exclude-dirs → exclusions.paths          | FULLY_FUNCTIONAL | Directory exclusions               |
| issues.exclude-files → exclusions.paths         | FULLY_FUNCTIONAL | File exclusions                    |
| Linter-specific settings migration              | FULLY_FUNCTIONAL | gci, cyclop, wrapcheck, etc.       |
| Removed settings cleanup                        | FULLY_FUNCTIONAL | Removes deprecated linter settings |
| `--skip-validation` flag                        | FULLY_FUNCTIONAL | Skips post-migration validation    |

## go-finding Integration

| Feature                                           | Status           | Notes                                       |
| ------------------------------------------------- | ---------------- | ------------------------------------------- |
| LinterRecommendation → finding.Finding conversion | FULLY_FUNCTIONAL | Priority-to-severity mapping                |
| ValidationError → finding.Finding conversion      | FULLY_FUNCTIONAL | Config errors as findings                   |
| golangci-lint JSON → Findings parser              | FULLY_FUNCTIONAL | Parse `golangci-lint run --out-format=json` |
| ConfigAnalysisDetector (pipeline.Detector)        | FULLY_FUNCTIONAL | For CI/CD pipeline integration              |
| SARIF 2.1.0 output                                | FULLY_FUNCTIONAL | CI/CD integration                           |
| go-finding Report JSON output                     | FULLY_FUNCTIONAL | Structured with summary                     |
| Diff Change → Finding conversion                  | FULLY_FUNCTIONAL | Config diff as findings                     |

## Reports

| Feature                   | Status           | Notes                              |
| ------------------------- | ---------------- | ---------------------------------- |
| HTML report (templ-based) | FULLY_FUNCTIONAL | Dark mode, responsive, color-coded |
| JSON report               | FULLY_FUNCTIONAL | Machine-readable (PascalCase keys) |
| SARIF report              | FULLY_FUNCTIONAL | CI/CD integration                  |
| finding JSON report       | FULLY_FUNCTIONAL | Unified data model                 |

## Project Detection

| Feature                                     | Status           | Notes                                |
| ------------------------------------------- | ---------------- | ------------------------------------ |
| Monorepo detection (multiple go.mod)        | FULLY_FUNCTIONAL |                                      |
| CLI project detection (cobra, urfave/cli)   | FULLY_FUNCTIONAL |                                      |
| Web project detection (gin, echo, net/http) | FULLY_FUNCTIONAL |                                      |
| Library project detection                   | FULLY_FUNCTIONAL |                                      |
| API service detection                       | FULLY_FUNCTIONAL |                                      |
| swaggo annotation detection                 | FULLY_FUNCTIONAL | `@Router`, `@Summary`, `@Tags`, etc. |

## Error Handling

| Feature                                                                      | Status           | Notes                  |
| ---------------------------------------------------------------------------- | ---------------- | ---------------------- |
| Custom error types (ConfigError, AnalysisError, ReportError, MigrationError) | FULLY_FUNCTIONAL | `pkg/errors/errors.go` |
| Error wrapping with context (%w)                                             | FULLY_FUNCTIONAL |                        |
| Structured logging (charmbracelet/log)                                       | FULLY_FUNCTIONAL |                        |
| Panic-free finding builder                                                   | FULLY_FUNCTIONAL | `pkg/finding/`         |

## Build & CI

| Feature                             | Status           | Notes                                                               |
| ----------------------------------- | ---------------- | ------------------------------------------------------------------- |
| Nix flake build                     | FULLY_FUNCTIONAL | Reproducible builds                                                 |
| GitHub Actions CI (Go 1.26)         | FULLY_FUNCTIONAL |                                                                     |
| Pre-commit hook                     | FULLY_FUNCTIONAL | golangci-lint, go-test, go-fmt                                      |
| Version injection via ldflags       | FULLY_FUNCTIONAL | version, commit, date, treeState                                    |
| Auto-tag workflow                   | FULLY_FUNCTIONAL | Tags on merge to master                                             |
| Committed templ output (\_templ.go) | FULLY_FUNCTIONAL | No build-time generation needed                                     |
| Govulncheck security scanning       | FULLY_FUNCTIONAL | CI job runs govulncheck ./...                                       |
| Coverage threshold gate             | FULLY_FUNCTIONAL | cmd/coverage-check (Go program, 60% threshold)                      |
| Fuzz + property tests               | FULLY_FUNCTIONAL | Set algebra invariants (commutative, idempotent, subset)            |
| `--json-errors` flag                | FULLY_FUNCTIONAL | JSON error output for CI/CD                                         |
| `encoding/json/v2` migration        | FULLY_FUNCTIONAL | All files migrated; GOEXPERIMENT=jsonv2 in flake.nix + CI workflows |
