# golangci-lint-auto-configure — Feature Audit

**Version:** v0.2.0+ (unreleased changes)
**Last Audited:** 2026-06-05

---

## CLI Commands

| Feature          | Command        | Status | Notes                                                      |
| ---------------- | -------------- | ------ | ---------------------------------------------------------- |
| Auto-configure   | `configure`    | Stable | Enables recommended linters, applies fixes                 |
| Analyze config   | `analyze`      | Stable | Reports missing/extra linters, supports SARIF/finding JSON |
| Validate config  | `validate`     | Stable | Checks YAML validity, supports SARIF output                |
| Generate report  | `report`       | Stable | HTML, JSON, SARIF, finding report formats                  |
| Migrate v1→v2    | `migrate`      | Stable | Migrates v1 configs to v2 format                           |
| Install hook     | `install-hook` | Stable | Installs git pre-commit hook                               |
| Shell completion | `completion`   | Stable | bash, zsh, fish, powershell                                |

## Auto-Configuration Features

| Feature                                               | Status | Notes                                      |
| ----------------------------------------------------- | ------ | ------------------------------------------ |
| 119 linter priorities (Critical/High/Medium/Optional) | Stable | `pkg/constants/linter_priorities.go`       |
| Linter reasons (human-readable)                       | Stable | `pkg/constants/linter_reasons.go`          |
| Priority-based filtering (`--priority`)               | Stable | configure command                          |
| Dry-run mode (`--dry-run`)                            | Stable | Shows what would change                    |
| CI check mode (`--check`)                             | Stable | Exit 0 if optimal, exit 1 if changes needed |
| Diff preview (`--diff`)                               | Stable | Shows config diff before applying          |
| Deprecated linter auto-replacement                    | Stable | wsl→wsl_v5, gomodguard→gomodguard_v2, etc. |
| Version-gated deprecation                             | Stable | gomodguard_v2 requires v2.12.0+            |
| Typecheck linter removal                              | Stable | Removes from enable/disable lists          |
| Invalid duration fix                                  | Stable | Fixes empty/invalid timeout values         |
| Multiple binary detection                             | Stable | Warns if multiple golangci-lint binaries   |

## Default Settings Injection

| Feature                                                | Status | Notes                              |
| ------------------------------------------------------ | ------ | ---------------------------------- |
| depguard defaults ($gostd, $module)                    | Stable | Prevents deny-all default          |
| ireturn defaults (error, empty, anon, stdlib, generic) | Stable | Reasonable interface return policy |
| gocritic defaults (ifElseChain disabled)               | Stable | Removes noisy checks               |
| exhaustruct defaults (os/exec.Cmd excluded)            | Stable | Common struct exemption            |
| revive defaults (exported, package-comments disabled)  | Stable | Noisy without config               |
| varnamelen defaults (short names, ignore flags)        | Stable | Common short variable exemptions   |
| gomoddirectives defaults (replace-local: true)         | Stable | Local dev support                  |
| cyclop defaults (max-complexity: 12)                   | Stable | Reasonable complexity threshold    |
| golines formatter defaults (max-len: 120)              | Stable | When enabled via lll replacement   |
| output.formats initialization                          | Stable | Empty map to prevent nil issues    |

## Exclusion Automation

| Feature                                                         | Status | Notes                                                                     |
| --------------------------------------------------------------- | ------ | ------------------------------------------------------------------------- |
| Default linter exclusion paths (\_templ.go$, .gen.go$, vendor/) | Stable | Always injected                                                           |
| Default formatter exclusion paths (\_templ.go$)                 | Stable | Always injected                                                           |
| Default test exclusion rules (6 linters for \_test.go)          | Stable | exhaustruct, testpackage, gochecknoglobals, funlen, cyclop, goconst       |
| Default unused text exclusion for test files                    | Stable | Suppresses unused false positives in tests                                |
| `generated: lax` auto-set                                       | Stable | Both linters and formatters                                               |
| gogenfilter dynamic scan                                        | Stable | Detects templ, protobuf, wire, moq, mockgen, stringer, sqlc, oapi-codegen |
| gogenfilter/v3 two-phase detection                              | Stable | Filename first, content second                                            |
| Deduplication of exclusion paths                                | Stable | MergeExclusionPaths                                                       |

## Presets

| Feature                                              | Status | Notes                              |
| ---------------------------------------------------- | ------ | ---------------------------------- |
| `minimal` preset (5 linters)                         | Stable | Essential only, fastest            |
| `standard` preset (8 linters)                        | Stable | Good balance for most projects     |
| `strict` preset (17 linters)                         | Stable | Maximum linting for CI/CD          |
| `security` preset                                    | Stable | Security-focused only              |
| `performance` preset                                 | Stable | Performance optimization           |
| `reference` preset (60+ linters)                     | Stable | All critical + high priority       |
| Auto-detect project type and select preset (`--detect`) | Stable | CLI, web, library, API, monorepo |

## Formatter Management

| Feature                                                   | Status | Notes                       |
| --------------------------------------------------------- | ------ | --------------------------- |
| Core formatters (gci, gofumpt, goimports)                 | Stable | Always enabled              |
| golines auto-enable (when lll detected)                   | Stable | Replaces redundant linter   |
| swaggo auto-detection                                     | Stable | Detects swag annotations    |
| Redundant formatter removal (gofmt when gofumpt)          | Stable | Superset detection          |
| Redundant linter removal (lll when golines)               | Stable | Formatter supersedes linter |
| Formatter ordering (gci→goimports→gofumpt→golines→swaggo) | Stable | Canonical order             |

## Build & Runner Settings

| Feature                                | Status | Notes                                                     |
| -------------------------------------- | ------ | --------------------------------------------------------- |
| Go version auto-detection              | Stable | Sets run.go to local version                              |
| GOEXPERIMENT build tags auto-injection | Stable | arenas, goroutineleakprofile, jsonv2, runtimesecret, simd |
| allow-parallel-runners enablement      | Stable | Always enabled                                            |
| allow-serial-runners enablement        | Stable | Always enabled                                            |
| Version field fix (empty → "2")        | Stable | Pre-flight check                                          |
| Benchmarking suite                     | Stable | analyzer and fixer benchmarks                             |

## Migration (v1 → v2)

| Feature                                         | Status | Notes                              |
| ----------------------------------------------- | ------ | ---------------------------------- |
| issues.exclude-rules → linters.exclusions.rules | Stable | Path-based migration               |
| issues.exclude-dirs → exclusions.paths          | Stable | Directory exclusions               |
| issues.exclude-files → exclusions.paths         | Stable | File exclusions                    |
| Linter-specific settings migration              | Stable | gci, cyclop, wrapcheck, etc.       |
| Removed settings cleanup                        | Stable | Removes deprecated linter settings |
| `--skip-validation` flag                        | Stable | Skips post-migration validation    |

## go-finding Integration

| Feature                                           | Status | Notes                                       |
| ------------------------------------------------- | ------ | ------------------------------------------- |
| LinterRecommendation → finding.Finding conversion | Stable | Priority-to-severity mapping                |
| ValidationError → finding.Finding conversion      | Stable | Config errors as findings                   |
| golangci-lint JSON → Findings parser              | Stable | Parse `golangci-lint run --out-format=json` |
| ConfigAnalysisDetector (pipeline.Detector)        | Stable | For CI/CD pipeline integration              |
| SARIF 2.1.0 output                                | Stable | CI/CD integration                           |
| go-finding Report JSON output                     | Stable | Structured with summary                     |
| Diff Change → Finding conversion                  | Stable | Config diff as findings                     |

## Reports

| Feature                   | Status | Notes                              |
| ------------------------- | ------ | ---------------------------------- |
| HTML report (templ-based) | Stable | Dark mode, responsive, color-coded |
| JSON report               | Stable | Machine-readable                   |
| SARIF report              | Stable | CI/CD integration                  |
| finding JSON report       | Stable | Unified data model                 |

## Project Detection

| Feature                                     | Status | Notes                                  |
| ------------------------------------------- | ------ | -------------------------------------- |
| Monorepo detection (multiple go.mod)        | Stable |                                        |
| CLI project detection (cobra, urfave/cli)   | Stable |                                        |
| Web project detection (gin, echo, net/http) | Stable |                                        |
| Library project detection                   | Stable |                                        |
| API service detection                       | Stable |                                        |
| swaggo annotation detection                 | Stable | `@Router`, `@Summary`, `@Tags`, etc.   |

## Error Handling

| Feature                                                      | Status | Notes                  |
| ------------------------------------------------------------ | ------ | ---------------------- |
| Custom error types (ConfigError, AnalysisError, ReportError) | Stable | `pkg/errors/errors.go` |
| Result type (railway-oriented)                               | Stable | `pkg/types/result.go`  |
| Error wrapping with context (%w)                             | Stable |                        |
| Structured logging (charmbracelet/log)                       | Stable |                        |
| Panic-free finding builder                                   | Stable | `pkg/finding/`         |

## Build & CI

| Feature                                 | Status | Notes                            |
| --------------------------------------- | ------ | -------------------------------- |
| Nix flake build                         | Stable | Reproducible builds              |
| justfile recipes                        | Stable | Primary build interface          |
| GitHub Actions CI (Go 1.25/1.26 matrix) | Stable |                                  |
| Pre-commit hook                         | Stable | golangci-lint, go-test, go-fmt   |
| Version injection via ldflags           | Stable | version, commit, date, treeState |
| Auto-tag workflow                       | Stable | Tags on merge to master          |
| templ generate in Nix build             | Stable | Generated code in pipeline       |
