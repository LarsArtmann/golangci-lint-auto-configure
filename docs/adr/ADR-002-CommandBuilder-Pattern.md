# ADR-002: CommandBuilder Pattern for CLI Commands

**Status:** Accepted  
**Date:** 2026-04-09  
**Author:** Lars Artmann (@larsartmann)

---

## Context

Our CLI has multiple commands (configure, analyze, validate, report, migrate, install-hook) that all require the same dependencies: logger, analyzer, and configLoader. The original implementation scattered dependency creation throughout the commands, leading to:

1. **Code duplication**: Each command created its own dependencies
2. **Inconsistent initialization**: Different commands initialized dependencies differently
3. **Testing difficulties**: Hard to inject mocks for testing
4. **Violation of DRY**: Same setup code repeated in multiple places

---

## Decision

We will implement a `CommandBuilder` pattern that centralizes dependency management for CLI commands.

### Implementation

```go
// internal/cli/cmd_builder.go
type CommandBuilder struct {
    logger       *log.Logger
    analyzer     *linter.Analyzer
    configLoader *config.Loader
}

func NewCommandBuilder(
    logger *log.Logger,
    analyzer *linter.Analyzer,
    configLoader *config.Loader,
) *CommandBuilder {
    return &CommandBuilder{...}
}

// Build creates a cobra.Command with dependencies accessible
func (b *CommandBuilder) Build(use, short string, runE func(*cobra.Command, []string) error, options ...func(*cobra.Command)) *cobra.Command

// Accessor methods for commands to use
func (b *CommandBuilder) Logger() *log.Logger
func (b *CommandBuilder) Analyzer() *linter.Analyzer
func (b *CommandBuilder) ConfigLoader() *config.Loader

// Fluent options for common patterns
func WithLong(long string) func(*cobra.Command)
func (b *CommandBuilder) WithStringFlag(name, shorthand, value, usage string) func(*cobra.Command)
```

---

## Consequences

### Positive

- **Single Source of Truth**: One place to create dependencies
- **Consistent Initialization**: All commands use same initialization logic
- **Testability**: Easy to inject mocks via CommandBuilder constructor
- **Reduced Boilerplate**: Commands don't repeat dependency setup
- **Extensibility**: New dependencies added in one place

### Negative

- **Additional Abstraction**: Developers must understand the pattern
- **Indirection**: Commands access dependencies through builder methods

### Trade-offs Accepted

- Slight learning curve outweighed by maintainability benefits

---

## Usage Examples

### Before (scattered dependencies)

```go
// internal/cli/commands.go
func NewConfigureCommand() *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            logger := log.New(cmd.ErrOrStderr())
            analyzer := linter.NewAnalyzer(logger)  // Created here
            configLoader := config.NewLoader(logger)  // Created here
            // ... use them
        },
    }
}

func NewAnalyzeCommand() *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            logger := log.New(cmd.ErrOrStderr())
            analyzer := linter.NewAnalyzer(logger)  // Duplicated!
            // ... use them
        },
    }
}
```

### After (with CommandBuilder)

```go
// internal/cli/commands.go
func NewRootCommand() *cobra.Command {
    // Dependencies created ONCE
    logger := log.NewWithOptions(os.Stderr, log.Options{Level: log.InfoLevel})
    analyzer := linter.NewAnalyzer(logger)
    configLoader := config.NewLoader(logger)
    builder := NewCommandBuilder(logger, analyzer, configLoader)

    rootCmd := &cobra.Command{...}
    rootCmd.AddCommand(newConfigureCommand(builder))
    rootCmd.AddCommand(newAnalyzeCommand(builder))
    // ...
}

func newConfigureCommand(b *CommandBuilder) *cobra.Command {
    return b.Build(
        "configure",
        "Configure golangci-lint",
        func(cmd *cobra.Command, args []string) error {
            b.Logger().Info("Configuring...")  // Use builder accessors
            // ...
        },
        b.WithStringFlag("priority", "p", "", "Filter by priority"),
    )
}
```

---

## Migration Strategy

1. **Phase 1**: Create CommandBuilder with core dependencies (logger, analyzer, configLoader)
2. **Phase 2**: Migrate high-traffic commands (configure, analyze)
3. **Phase 3**: Migrate remaining commands (validate, report, migrate, install-hook)
4. **Phase 4**: Add helper methods as needed (WithStringFlag, WithBoolFlag, etc.)

---

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| Dependency Injection Container | Full DI, automatic wiring | Overkill for CLI, adds complexity | Rejected |
| Global Variables | Simple, no passing | Testing nightmare, hidden dependencies | Rejected |
| Manual Passing | Explicit, clear | Verbose, repetitive | Rejected |
| **CommandBuilder Pattern** | Balanced, testable, DRY | Slight abstraction cost | **Accepted** |

---

## Related Decisions

- ADR-001: Set[T] Type - Used in commands that handle linter sets
- ADR-003: Interface-Based Design - Enables mocking for CommandBuilder testing

---

## References

- [CommandBuilder Implementation](../../internal/cli/cmd_builder.go)
- [Command Usage](../../internal/cli/commands.go)

---

_Accepted by: Lars Artmann_  
_Date: 2026-04-09_
