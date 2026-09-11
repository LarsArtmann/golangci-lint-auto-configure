# Architecture Decision Records

This document captures key architectural decisions made during the development of golangci-lint-auto-configure.

The full decision records live in [`docs/adr/`](adr/) — one file per decision, named `ADR-NNN-<slug>.md`. The table below is the index.

| ADR | Decision |
|-----|----------|
| [ADR-008](ADR-008-Railway-Oriented-Programming.md) | Railway-Oriented Programming with Result Types |
| [ADR-009](ADR-009-MigrationResult-Error-Field.md) | MigrationResult Uses Error Field Instead of Success Bool |
| [ADR-010](ADR-010-Detector-Caching.md) | Detector Caching for Project Type Analysis |
| [ADR-011](ADR-011-Separate-Error-Types.md) | Separate Error Types for Different Domains |
| [ADR-012](ADR-012-Filesystem-Abstraction.md) | Filesystem Abstraction for Testability |
| [ADR-013](ADR-013-Context-Propagation.md) | Context Propagation for Cancellation |
| [ADR-014](ADR-014-Strong-Type-Aliases.md) | Strong Type Aliases for Linter and Formatter Names |
| [ADR-015](ADR-015-Cobra-CLI.md) | Command-Line Interface with Cobra |

## Future Considerations

### Fixer Module Split

The fixer has already been split into focused files:

- `fixer.go` — core orchestration
- `fixer_deprecated.go` — deprecated linter handling
- `fixer_config.go` — default settings injection
- `fixer_enforce.go` — disable-reason sidecar enforcement
- `fixer_formatters.go` — formatter management
- `fixer_preflight.go` — pre-flight normalization
- `fixer_recorder.go` — change counting/recording
- `fixer_results.go` — result aggregation
- `fixer_audit.go` — audit ledger integration

Further refinement could extract version-fixing and dry-run logic into their own files if `fixer.go` grows again.

### Generic Result Types

The codebase uses standard Go `(value, error)` returns. If a structured Result type becomes beneficial for carrying warnings/counts alongside errors, a project-specific `Result[T]` type (not `samber/mo`) could be introduced — see TODO_LIST.md for the proposal.
