# ADR-015: Command-Line Interface with Cobra

**Status:** Accepted\
**Date:** 2026-03-26

## Context

The application is a CLI tool with multiple subcommands.

## Decision

Use `spf13/cobra` for CLI framework:

```go
cmd := &cobra.Command{
    Use:   "configure",
    Short: "Automatically configure and optimize golangci-lint",
    RunE: func(cmd *cobra.Command, _ []string) error {
        return runConfigure(cmd.Context(), logger, analyzer, configLoader,
            priority, preset, dryRun, configPath, check, showDiff)
    },
}
```

## Consequences

**Positive:**

- Standard Go CLI patterns
- Built-in help, flags, and subcommand support
- Well-maintained library

**Negative:**

- Additional dependency
- Some deprecated APIs in older versions

---
