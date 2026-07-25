# Domain Language

A **Ubiquitous Language** for `golangci-lint-auto-configure` — shared across users, contributors, and AI agents.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a user, define it here.

## Glossary

| Term                  | Definition                                                                                | Context                                      |
| --------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------- |
| Linter                | A static analysis tool that checks Go source code for issues                              | golangci-lint integrates many linters        |
| Formatter             | A tool that reformats Go source code (gofumpt, goimports, gci, golines, swaggo)           | Distinct from linters — changes code style   |
| Linter Priority       | A tier (Critical, High, Medium, Optional) ranking how important a linter is               | Drives `--priority` filtering                |
| Linter Recommendation | A suggestion to enable/disable a specific linter, with reason and severity                | Output of `analyze`, converted to Findings   |
| Preset                | A curated set of linters for a specific use case (minimal, standard, strict, etc.)        | Applied via `--preset` flag                  |
| Project Type          | Classification of the target project (CLI, Library, Web, API, Monorepo)                   | Auto-detected via `--detect`                 |
| Fixer                 | The engine that applies config mutations: enable linters, fix settings, remove deprecated | `pkg/linter/fixer.go` — linear, idempotent   |
| Config Mutation       | Any change the fixer makes to a golangci-lint config file                                 | Tracked by `configChangeRecorder`            |
| Normalization         | Pre-flight fixes applied before analysis (timeout, version, deprecated linters)           | Saved to disk before golangci-lint can run   |
| Finding               | A unified issue representation (from go-finding) with severity, rule, file, message       | Used for SARIF/JSON output                   |
| Config Analysis       | The full result of analyzing a config: enabled/disabled linters, recommendations          | `ConfigAnalysis` struct, output of `analyze` |
| Deprecation           | A linter that has been superseded by a newer version (e.g. wsl → wsl_v5)                  | Auto-replaced by the fixer                   |
| Version-Gated         | A linter that requires a minimum golangci-lint version to be available                    | `LinterMinVersions` map                      |
| Exclusion Path        | A regex pattern that excludes files from linting (e.g. `_templ\.go$`, `vendor/`)          | RE2 syntax, injected into config             |
| Migration             | Converting a v1 golangci-lint config to v2 schema format                                  | `migrate` command                            |
| Validation            | Checking a config for correctness (YAML validity, schema compliance)                      | `validate` command                           |

## Entities

Objects with identity and lifecycle.

| Term             | Definition                                                              | Context                                       |
| ---------------- | ----------------------------------------------------------------------- | --------------------------------------------- |
| Config           | A parsed golangci-lint configuration (v1 or v2 schema)                  | `pkg/types/config_types.go` — the core domain |
| ConfigAnalysis   | Full analysis result: enabled/disabled linters, recommendations, counts | `pkg/types/types.go`                          |
| MigrationResult  | Outcome of a v1→v2 migration: fixes applied count, message, next steps   | `pkg/types/types.go`                          |
| ValidationResult | Outcome of config validation: valid flag, errors                        | `pkg/types/types.go`                          |
| ConfigHealth     | Health-check result with severity-scored issues                         | `pkg/types/validation.go`                     |

## Value Objects

Immutable objects defined by attributes.

| Term                 | Definition                                                          | Context                             |
| -------------------- | ------------------------------------------------------------------- | ----------------------------------- |
| LinterPriority       | An int enum: Critical (0), High (1), Medium (2), Optional (3)       | `pkg/types/types.go` — branded type |
| LinterName           | A branded string type for linter names (prevents string confusion)  | `pkg/types/types.go`                |
| FormatterName        | A branded string type for formatter names                           | `pkg/types/types.go`                |
| LinterRecommendation | A value: linter name, priority, reason                             | Output of analysis, input to fixer  |
| LinterReplacement    | A deprecated linter and its recommended successor                   | `pkg/types/types.go`                |
| HealthIssue          | A single config health problem with severity and message            | `pkg/types/validation.go`           |
| HealthSeverity       | Severity level for health issues (Critical, Warning, Info)          | `pkg/types/validation.go`           |
| GoExperiment         | A Go build experiment (arenas, jsonv2, etc.) to inject as build tag | `pkg/types/types.go`                |
| Change               | A single diff change (addition, removal, or modification of a config line) | `pkg/diff/differ.go`                |
| GolangciLintIssue    | A single issue from `golangci-lint run --out-format json` output    | `pkg/finding/golangci_lint.go`      |
| Set[T]               | A generic set with full algebra (union, intersection, difference)   | `pkg/types/set.go`                  |

## Commands

Actions the system can perform.

| Term         | Definition                                                   | Context                |
| ------------ | ------------------------------------------------------------ | ---------------------- |
| configure    | Auto-configure golangci-lint: analyze, recommend, fix, save  | Default command        |
| analyze      | Analyze config and report missing/extra linters (no changes) | Read-only command      |
| validate     | Check config validity (YAML, schema, health)                 | Read-only command      |
| report       | Generate HTML/JSON/SARIF/finding report                      | Output command         |
| audit        | Query the config-mutation audit ledger                       | Query command          |
| migrate      | Convert v1 config to v2 schema                               | Transformation command |
| install-hook | Install git pre-commit hook                                  | Setup command          |
| presets      | List all available presets with descriptions                 | Informational command  |
| completion   | Generate shell completion script (bash/zsh/fish/powershell)  | Setup command          |

## Bounded Contexts

Subsystems with distinct vocabulary.

| Context        | Description                                                              | Key Package           |
| -------------- | ------------------------------------------------------------------------ | --------------------- |
| Configuration  | Loading, parsing, merging, and saving golangci-lint config files         | `pkg/config/`         |
| Analysis       | Detecting missing/extra linters, computing recommendations               | `pkg/linter/`         |
| Fixing         | Applying mutations to config (enable/disable linters, inject settings)   | `pkg/linter/fixer.go` |
| Finding        | Converting analysis results to go-finding Findings for SARIF/JSON output | `pkg/finding/`        |
| Migration      | Transforming v1 configs to v2 schema                                     | `pkg/migration/`      |
| Detection      | Auto-detecting project type (CLI, Web, Library, API, Monorepo)           | `pkg/detection/`      |
| Reporting      | Generating HTML/JSON/SARIF/finding output                                | `pkg/report/`         |
| Error Handling | Structured error classification with BSD sysexits exit codes             | `pkg/errors/`         |
| Audit Trail    | Append-only JSONL ledger recording every config mutation                | `pkg/audit/`           |
| Policy         | Disable-reason sidecar enforcement (anti-gaming)                        | `pkg/policy/`          |
| Linter Data    | Static constants: priorities, reasons, presets, deprecated mappings      | `pkg/constants/`      |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
