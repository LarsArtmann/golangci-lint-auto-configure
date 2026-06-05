# Agent Guide: golangci-lint-auto-configure

This guide provides essential information for agents working on the golangci-lint-auto-configure codebase.

## Project Overview

**golangci-lint-auto-configure** is a Go CLI tool that automatically configures and optimizes golangci-lint configurations by:

- Analyzing existing golangci-lint configs
- Detecting missing linters with smart categorization
- Recommending optimal linter settings based on project type
- Auto-fixing configuration issues
- Automatically replacing deprecated linters with their successors
- Migrating v1 configs to v2 format (merged from golangci-config-migrator)
- Generating HTML/JSON/SARIF reports
- Converting results to go-finding unified model for pipeline and cross-tool integration

## Essential Commands

**Always use `just` for commands - they're the standard for this project:**

```bash
just build          # Build the CLI binary
just test           # Run all tests with coverage (uses ginkgo)
just test-coverage  # Show coverage summary
just coverage-html  # Generate and open HTML coverage report
just lint           # Run golangci-lint with .golangci.yml
just run            # Run the CLI
just clean          # Clean build artifacts (removes bin/)
just install        # Install to GOPATH/bin
just install-local  # Install locally with version ldflags
just fmt            # Format code with gofmt
just fmt-check      # Check formatting without modifying
just tidy           # Tidy go.mod
```

**Nix Commands:**

```bash
nix develop                          # Enter dev shell (all tools provided)
nix develop --command just test      # Run tests inside Nix shell
nix build                            # Build the CLI binary (reproducible)
nix flake check                      # Run all Nix checks
nix run . -- analyze                 # Run the CLI directly
nix flake update                     # Update all flake inputs
just nix-build                       # Build with Nix (just wrapper)
just nix-check                       # Run Nix checks (just wrapper)
just nix-update                      # Update flake inputs (just wrapper)
```

**vendorHash Update (after go.mod changes):**

```bash
just tidy                            # Tidy dependencies
nix build 2>&1 | tail -5             # Get expected hash from error
# Copy the "got:" hash into flake.nix vendorHash
nix build                            # Rebuild with correct hash
```

**CLI Commands (after build):**

```bash
./bin/golangci-lint-auto-configure configure [--priority critical|high|medium|optional] [--dry-run]
./bin/golangci-lint-auto-configure analyze [--config .golangci.yml]
./bin/golangci-lint-auto-configure validate [--config .golangci.yml]
./bin/golangci-lint-auto-configure report [--format html|json] [--output path]
./bin/golangci-lint-auto-configure migrate [--skip-validation]
./bin/golangci-lint-auto-configure install-hook
```

## Technology Stack

### Core Dependencies

- **Go**: 1.26+ (CI tests on 1.25 and 1.26, go.mod requires 1.26.0)
- **Cobra**: CLI command framework
- **Charmbracelet Log**: Structured logging
- **Charmbracelet Fang**: Enhanced CLI features
- **Ginkgo v2 + Gomega**: BDD testing framework (NOT standard Go testing)
- **Templ**: HTML template system for reports
- **YAML v3**: Configuration parsing (`go.yaml.in/yaml/v3`)
- **go-finding**: Unified finding model (local replace)
- **gogenfilter/v3**: Auto-generated code detection and filtering (`github.com/LarsArtmann/gogenfilter/v3`)

### External Tools Required

- **golangci-lint**: v2.10.1+ (auto-detected, minimum version enforced)
- **Go**: 1.26+ required for compilation

## Code Organization

### Directory Structure

```
golangci-lint-auto-configure/
├── cmd/
│   └── golangci-lint-auto-configure/
│       └── main.go                    # Entry point, sets version via ldflags
├── pkg/
│   ├── types/                        # Core type definitions and interfaces
│   │   ├── types.go                  # Main types: LinterPriority, Config, LinterInfo, etc.
│   │   └── result.go                # Result types
│   ├── constants/
│   │   ├── linter_priorities.go  # Linter priorities (119 entries)
│   │   ├── linter_reasons.go     # Human-readable explanations
│   │   ├── formatter_data.go     # Formatter priorities and reasons
│   │   ├── presets.go            # Pre-defined linter configurations
│   │   ├── rules.go              # Redundant/deprecated linter rules
│   │   ├── config.go             # Constants and config defaults
│   │   ├── version.go            # Min golangci-lint version
│   │   └── experiments.go        # Feature flags
│   ├── config/
│   │   ├── loader.go                # Load, save, validate golangci-lint configs
│   │   └── loader_test.go           # Config loading tests
│   ├── linter/
│   │   ├── analyzer.go              # Analyze configs, get recommendations
│   │   ├── fixer.go                 # Apply fixes to configs
│   │   ├── fixer_preflight.go       # Pre-flight checks before fixing
│   │   ├── fixer_formatters.go      # Formatter-specific fixing
│   │   ├── fixer_config.go          # Config construction for fixing
│   │   ├── fixer_deprecated.go      # Deprecated linter replacement
│   │   ├── fixer_results.go         # Fix result types
│   │   ├── categorizer.go           # Linter categorization
│   │   ├── version_checker.go       # golangci-lint version checking
│   │   ├── command_runner.go        # Command execution
│   │   └── *_test.go                # Tests
│   ├── detection/
│   │   ├── detector.go             # Detect project type (CLI, web, library, etc.)
│   │   └── detector_test.go        # Detector tests
│   ├── gogenfilter/
│   │   ├── scanner.go              # Scan project for auto-generated files using gogenfilter/v3
│   │   └── scanner_test.go         # Scanner tests
│   ├── diff/
│   │   ├── differ.go               # Compare two configs and show changes
│   │   └── differ_test.go          # Diff tests
│   ├── report/
│   │   ├── generator.go            # HTML report generation (templ-based)
│   │   ├── json_report_generator.go # JSON report generation
│   │   ├── report.templ            # HTML template (generates Go code)
│   │   └── report_templ.go         # Generated Go code from templ
│   ├── migration/                   # v1 to v2 config migration (merged from golangci-config-migrator)
│   │   ├── migrator.go             # Main migrator struct and logic
│   │   ├── migrations.go           # Migration helpers
│   │   ├── migrations_linters_settings.go # Linter-specific migrations
│   │   ├── config_types.go         # YAML config structs for v1/v2
│   │   ├── rules.go                # Migration rules
│   │   ├── validator.go            # Config validation
│   │   ├── yaml_loader.go          # Load/Save YAML configs
│   │   └── testdata/               # Test fixtures for migration
│   ├── errors/
│   │   └── errors.go              # Custom error types (package: apperrors)
│   ├── version/
│   │   └── version.go             # Structured version info with runtime/debug fallback
├── internal/
│   ├── cli/
│   │   ├── commands.go             # Root command + subcommand wiring
│   │   ├── cmd_configure.go        # Configure command
│   │   ├── cmd_analyze.go          # Analyze command
│   │   ├── cmd_validate.go         # Validate command
│   │   ├── cmd_report.go           # Report command
│   │   ├── cmd_builder.go          # CommandBuilder type
│   │   ├── cmd/                    # Separate package for some commands
│   │   │   ├── migrate.go          # Migrate command
│   │   │   ├── installhook.go      # Install-hook command
│   │   │   └── completion.go       # Completion command
│   │   ├── commands_test.go        # Command tests
│   │   └── integration_test.go     # Integration tests
├── pkg/finding/                   # go-finding integration
│   ├── converter.go               # Convert LinterRecommendations/ValidationErrors to finding.Finding
│   ├── golangci_lint.go           # Parse golangci-lint JSON output to Findings
│   ├── detector.go                # ConfigAnalysisDetector for pipeline integration
│   ├── diff_converter.go          # Convert diff.Change to Finding
│   └── helpers.go                 # LSP, filter, merge, groupBy helpers
├── pkg/ui/
│   ├── formatter.go               # Terminal output formatting
│   ├── styled_output.go           # Styled output with lipgloss v2
│   └── finding_formatter.go       # Terminal formatting for go-finding objects
├── docs/                          # Documentation and status reports
├── examples/                      # Example configurations for different project types
│   ├── minimal.golangci.yml
│   ├── standard.golangci.yml
│   ├── web-project.golangci.yml
│   ├── cli-project.golangci.yml
│   └── library.golangci.yml
├── reports/                       # Generated linter documentation
├── scripts/                       # Utility scripts
└── justfile                       # Build/test/lint commands (PRIMARY INTERFACE)
```

### Key Architectural Patterns

**1. Interface-Based Design (for testability)**
All major components implement interfaces defined in `pkg/types/types.go`:

- `ConfigLoader`: Composite interface in `pkg/types/types.go` — combines ConfigReader, ConfigWriter, ConfigDiscovery, ConfigValidator, ConfigInspector, ConfigCreator. Backed by `*Loader` in `pkg/config/loader.go` using a minimal `config.FS` interface for filesystem ops.
- `LinterAnalyzer`: Analyze configs, get recommendations
- `LinterFixer`: Apply fixes to configs

**2. Strong Typing with Custom Types**

- `LinterName` (string): Prevents typos in linter names
- `FormatterName` (string): Prevents typos in formatter names
- `LinterPriority` (int): Critical, High, Medium, Optional
- `FormatterPriority` (int): High, Medium, Low

**3. Separation of Concerns**

- `pkg/linter/analyzer.go`: Analysis logic only
- `pkg/linter/fixer.go`: Modification logic only
- `pkg/config/loader.go`: Config I/O only
- `internal/cli/commands.go`: CLI command wiring

**4. Data-Driven Configuration**

- `pkg/constants/` (multiple files) contains all linter metadata:
  - `linter_priorities.go`: Priority levels (119 linters)
  - `linter_reasons.go`: Human-readable reasons
  - `formatter_data.go`: Formatter priorities and reasons
  - `presets.go`: Pre-defined configurations
  - `rules.go`: Deprecated/redundant linter rules (`DeprecatedLinters`)
  - `version.go`: Minimum golangci-lint version
  - `experiments.go`: Feature flags

**5. go-finding Integration**

- `pkg/finding/` provides unified data model for static analysis results
- SARIF 2.1.0 output for CI/CD integration
- Priority-to-severity mapping (Critical→critical, High→error, etc.)

## Testing Approach

### Test Framework: Ginkgo + Gomega (BDD)

**NOT standard Go testing - this project uses BDD style:**

```go
import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestAnalyzer(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Analyzer Suite")
}

var _ = Describe("Analyzer", func() {
    var analyzer *linter.Analyzer

    BeforeEach(func() {
        logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
        analyzer = linter.NewAnalyzer(logger)
    })

    Context("Priority Filtering", func() {
        It("should filter linters by priority", func() {
            // ... test logic using Expect()
        })
    })
})
```

**Running Tests:**

```bash
just test                    # Run all tests (ginkgo -r --cover)
ginkgo -v ./pkg/...        # Run with verbose output
ginkgo -r --focus="Name"   # Run specific tests by name
ginkgo -r --skip="Name"    # Skip specific tests
```

### Test Patterns

1. **Setup in `BeforeEach`**: Initialize test state
2. **BDD naming**: Use `Context` for scenarios, `It` for specific behaviors
3. **Matchers**: Use Gomega's `Expect().To(Equal())` syntax
4. **Table-driven tests**: Common for multiple test cases
5. **Error testing**: Test both success and failure paths

### Coverage

```bash
just coverage-html  # Generate and open HTML coverage report
just test-coverage  # Show coverage summary in terminal
```

## Linter Priority System

### Priority Levels

```go
const (
    LinterPriorityCritical  // Security and correctness (ALWAYS enable)
    LinterPriorityHigh      // Quality and maintainability (recommended)
    LinterPriorityMedium   // Style and consistency (optional)
    LinterPriorityOptional  // May be too strict for most projects
)
```

### Example Linters

- **Critical**: gosec, errcheck, staticcheck, govet, errchkjson, musttag, sloglint, nilerr, noctx, loggercheck
- **High**: wrapcheck, errorlint, prealloc, unconvert, ineffassign, gocyclo, funlen, cyclop, gocognit, maintidx, exhaustive, exhaustruct, goconst, misspell, revive, nolintlint, forcetypeassert
- **Medium**: dupword, godot, goheader, gofmt, gci, varnamelen, lll, whitespace, wsl_v5, grouper, dogsled, makezero, thelper, exportloopref, paralleltest
- **Optional**: All other linters

### Adding New Linters

Update `pkg/constants/linter_priorities.go` and `linter_reasons.go`:

1. Add to `LinterPriorities` map
2. Add to `LinterReasons` map (human-readable explanation)
3. Consider adding to `PresetLinters` if appropriate

## Configuration Management

### Config File Locations (search order)

1. `.golangci.yml`
2. `.golangci.yaml`
3. `.golangci.toml`
4. `.golangci.json`

### Config Schema (golangci-lint v2)

```yaml
version: "2"
run:
  timeout: 10m
  go: ""
  tests: true
linters:
  enable: [...] # List of enabled linters
  disable: [...] # List of disabled linters
  settings:
    funlen:
      lines: 80
      statements: 50
issues:
  max-issues-per-linter: 100
  max-same-issues: 15
```

### Version Control

- **Git-based**: Tool requires running in a git repository
- Git provides full history, branching, and rollback capabilities
- Use `git restore` or `git checkout` to revert config changes if needed

## CLI Commands

All 7 subcommands are registered in `internal/cli/commands.go`:

| Command        | File                              | Description                  |
| -------------- | --------------------------------- | ---------------------------- |
| `configure`    | `internal/cli/cmd_configure.go`   | Auto-configure golangci-lint |
| `analyze`      | `internal/cli/cmd_analyze.go`     | Analyze configuration        |
| `validate`     | `internal/cli/cmd_validate.go`    | Validate configuration       |
| `report`       | `internal/cli/cmd_report.go`      | Generate reports             |
| `migrate`      | `internal/cli/cmd/migrate.go`     | Migrate v1 to v2 config      |
| `install-hook` | `internal/cli/cmd/installhook.go` | Install pre-commit hook      |
| `completion`   | `internal/cli/cmd/completion.go`  | Generate shell completion    |

## Project Type Detection

### Supported Types

```go
const (
    ProjectTypeUnknown   // Cannot determine
    ProjectTypeCLI      // Command-line tools
    ProjectTypeLibrary  // Reusable packages, SDKs
    ProjectTypeWeb     // HTTP servers, REST APIs
    ProjectTypeAPI      // API services
    ProjectTypeMonorepo // Multiple go.mod files
)
```

### Detection Logic (in `pkg/detection/detector.go`)

1. Check for monorepo: Multiple `go.mod` files
2. Analyze `go.mod` for module path and imports
3. Check for `main` package
4. Check for HTTP frameworks (net/http, gin, echo, etc.)
5. Check for CLI frameworks (cobra, urfave/cli, etc.)
6. Apply decision tree based on findings

## Error Handling Patterns

### Custom Error Types (in `pkg/errors/errors.go`)

```go
type ConfigError struct {
    Message string
    Path    string
    Cause   error
}

type AnalysisError struct {
    Message string
    File    string
    Cause   error
}

type ReportError struct {
    Message string
    Path    string
    Cause   error
}
```

### Error Handling Best Practices

1. **Wrap errors with context**: Use `%w` for wrapping, always preserve chain
2. **Use custom error types**: For domain-specific errors (ConfigError, AnalysisError)
3. **Provide actionable messages**: Include file paths, config names
4. **Log errors before returning**: Use structured logging with logger
5. **Recover from errors**: Try multiple strategies (e.g., JSON version parsing → text fallback)

Example from `pkg/linter/analyzer.go`:

```go
if err := json.Unmarshal(output, &versionInfo); err != nil {
    a.logger.Debugf("Failed to parse JSON version output, falling back to text: %v", err)
    return a.checkVersionText()  // Fallback strategy
}
```

## HTML Report Generation

### Using Templ (Not Standard HTML/Templates)

```bash
# Generate Go code from templ file
templ generate

# The .templ file generates Go code in pkg/report/report_templ.go
# Render with:
err = Report(data).Render(context.Background(), f)
```

### Report Data Structure

```go
type ReportData struct {
    Analysis *types.ConfigAnalysis
}
```

### Styling

- Inline CSS in `pkg/report/report.templ`
- Dark mode support via `@media (prefers-color-scheme: dark)`
- Responsive design with CSS grid
- Color-coded priority levels

## Linter Version Checking

### Minimum Version: v2.10.1

Version checking in `pkg/linter/version_checker.go`:

1. Try `golangci-lint version --json` first (more reliable)
2. Fallback to text parsing if JSON not supported
3. Use `golang.org/x/mod/semver` for comparison
4. Fail with helpful error if version too old

### Handling Deprecated Linters

Automatic replacement of deprecated linters:

- `wsl` → `wsl_v5` (deprecated since golangci-lint v2.2.0)
- Mappings in `pkg/constants/rules.go`:
  ```go
  DeprecatedLinters = map[types.LinterName]types.LinterReplacement{
      "wsl": {Replacement: "wsl_v5", Reason: "original wsl is deprecated since v2.2.0"},
  }
  ```

## Code Style and Conventions

### Naming

- **Packages**: Lowercase, single word, descriptive
  - `linter` (not `linters`)
  - `config` (not `configuration`)
- **Interfaces**: Simple nouns ending in capability (e.g., `ConfigLoader`, `LinterAnalyzer`)
- **Functions**: CamelCase, descriptive verbs
- **Variables**: CamelCase, descriptive
- **Constants**: PascalCase for exported, camelCase for internal
- **File names**: snake_case for packages, camelCase for tests

### Struct Field Tags

```go
type Config struct {
    Version    string           `yaml:"version"`
    Linters    LintersConfig    `yaml:"linters"`
    Run        RunConfig        `yaml:"run"`
}
```

### Error Patterns

```go
// Return wrapped errors
return fmt.Errorf("failed to load config: %w", err)

// Use custom error types
return errors.NewConfigError("failed to parse config", path, err)
```

### Logging

```go
logger.Infof("Processing configuration: %s", path)
logger.Debugf("Found %d linters", count)
logger.Warnf("Deprecated linter detected: %s", name)
logger.Errorf("Analysis failed: %v", err)
```

## CI/CD Pipeline

### GitHub Actions (`.github/workflows/ci.yml`)

**Test Matrix:**

- Go versions: 1.25, 1.26
- Tests: `go test -v -race ./pkg/... ./internal/...`
- Coverage: Uploads to Codecov

**Lint Job:**

- Uses `golangci/golangci-lint-action@v9`
- Config: `.golangci.yml`
- Timeout: 5m

### Pre-Commit Hooks (`.pre-commit-config.yaml`)

```bash
# Install hooks
pre-commit install

# Run hooks manually
pre-commit run --all-files
```

Built-in hooks:

1. `golangci-configure`: Analyze config (dry-run)
2. `golangci-lint`: Run linter on changed files
3. `go-test`: Run tests
4. `go-fmt`: Check formatting
5. Standard pre-commit hooks (trailing whitespace, YAML check, etc.)

## Important Gotchas

### 1. Use `just` Commands, Not Manual Commands

- `just test` (NOT `go test ...`)
- `just build` (NOT `go build ...`)
- `just lint` (NOT manual golangci-lint commands)

### 2. Ginkgo Testing, Not Standard Go Testing

- Uses BDD style with `Describe`/`Context`/`It`
- Uses Gomega matchers, not testify assertions
- `just test` runs `ginkgo -r --cover`, NOT `go test`

### 3. Templ Requires Code Generation

- `.templ` files must be compiled to Go code
- Run `templ generate` (not in justfile, manual when needed)
- Generated file: `pkg/report/report_templ.go`

### 4. Dependency Injection Directory Does Not Exist

- `internal/di/` referenced in older docs but does not exist
- Dependency injection is manual in CLI commands
- No DI framework like samber/do or wire

### 5. Linter Priority Data is in Constants

- Linter priorities split across `pkg/constants/linter_priorities.go` and `linter_reasons.go`
- Modify those files to change linter behavior
- Not dynamically computed from golangci-lint

### 6. Versioning

- `pkg/version/` package manages version info (version, commit, date, treeState)
- `runtime/debug.ReadBuildInfo()` provides automatic VCS fallback when ldflags are not set
- Ldflags take precedence over buildinfo fallback
- All build targets (justfile, Nix, Docker) inject version via ldflags to `pkg/version.version/commit/date/treeState`
- `cli.Version` is self-initializing from `version.Get().Short()` — no manual setup needed
- `--version` output shows: Version, Commit, Built, Tree state
- Justfile shared variables: `VERSION`, `COMMIT`, `DATE`, `TREE_STATE`

### 7. go-finding is Nix Flake Input

- `go.mod` has local replace: `replace github.com/larsartmann/go-finding => ../go-finding`
- For Nix builds, `flake.nix` copies go-finding from a `git+ssh://` flake input into the source tree
- The `postPatch` hook redirects the replace directive to `./go-finding-vendor`
- Local dev still uses the `../go-finding` sibling directory directly
- CI fetches go-finding via SSH from GitHub (private repo)

### 8. Cobra Deprecation

- Previously used deprecated `cobra.ExactValidArgs()` — has been fixed
- Check current codebase if similar deprecation warnings appear

### 9. Config File Auto-Creation

- `configure` command creates default config if missing
- Uses `.golangci.yml` as default path
- Requires git repository for version control safety

## Working with This Codebase

### Adding a New CLI Command

1. Define command function in `internal/cli/cmd_*.go` (for configure/analyze/validate/report) or `internal/cli/cmd/*.go` (for migrate/install-hook/completion)
2. Wire up dependencies in `addSubCommands()` in `internal/cli/commands.go`
3. Add flags as needed
4. Write BDD tests in `internal/cli/commands_test.go` or `internal/cli/integration_test.go`
5. Run `just test` to verify

### Modifying Linter Priorities

1. Edit `pkg/constants/linter_priorities.go`
2. Update `LinterPriorities` map
3. Edit `pkg/constants/linter_reasons.go` to update `LinterReasons` map
4. Consider updating presets in `pkg/constants/presets.go`
5. Run tests: `just test`

### Adding New Linter Data

1. Add linter to `pkg/constants/linter_priorities.go` (`LinterPriorities` map)
2. Set priority (Critical/High/Medium/Optional)
3. Add reason to `pkg/constants/linter_reasons.go` (`LinterReasons` map)
4. Consider if deprecated (add to `DeprecatedLinters` in `pkg/constants/rules.go`)
5. Regenerate reports if needed

### Updating HTML Report

1. Edit `pkg/report/report.templ`
2. Run `templ generate` to compile to Go
3. Test report generation: `./bin/golangci-lint-auto-configure report`
4. Verify HTML output in browser

### Debugging Issues

1. Enable verbose mode: `--verbose` flag or `logger.SetLevel(log.DebugLevel)`
2. Check golangci-lint version: `golangci-lint version`
3. Verify config file exists and is valid: `golangci-lint-auto-configure validate`
4. Check test coverage: `just coverage-html`
5. Run specific tests: `ginkgo -r --focus="TestName"`

## Scripts Directory

- `pre-commit-hook.sh`: Git pre-commit hook script (also installed by CLI command)
- `validate_linter_doc.sh`: Validate linter documentation
- `verify_linter_count.sh`: Verify linter count matches expectations

## Documentation

- `README.md`: User-facing documentation
- `docs/`: Developer documentation and status reports
- `examples/`: Example configurations for different project types
- `reports/`: Auto-generated linter documentation (one .md per linter)

## Build Artifacts

- `bin/`: Build output (cleaned by `just clean`)
- `coverage.out`: Coverage profile (from `just test`)
- `coverage.html`: HTML coverage report (from `just coverage-html`)
- `report.html`: Generated analysis report (from CLI)

## Common Tasks

### Full Development Workflow

```bash
just tidy           # Update dependencies
just fmt           # Format code
just test           # Run tests
just lint           # Run linters
just build          # Build binary
just install-local  # Install with version
./bin/golangci-lint-auto-configure --help
```

### Release Preparation

```bash
just test           # All tests must pass
just lint           # All linters must pass
just coverage-html  # Check coverage
git tag v0.1.0
just install-local  # Build with version tag
```

### Adding a New Test

```go
var _ = Describe("New Feature", func() {
    Context("When X happens", func() {
        It("should do Y", func() {
            // Arrange
            input := "test"

            // Act
            result := Process(input)

            // Assert
            Expect(result).To(Equal("expected"))
        })
    })
})
```

### Fixing Linter Issues

```bash
# See what's wrong
just lint

# Fix specific issues manually or use golangci-lint
golangci-lint run --fix

# Verify fixes
just test && just lint
```

## Key Dependencies to Understand

- **spf13/cobra**: CLI command framework (patterns, flags, subcommands)
- **charmbracelet/log**: Structured logging (log.Infof, Debugf, Warnf, Errorf)
- **onsi/ginkgo/v2**: BDD testing framework (Describe, Context, It, BeforeEach)
- **onsi/gomega**: BDD assertions (Expect().To(Equal(), HaveLen(), BeNil()))
- **a-h/templ**: HTML templating (generates Go code from .templ files)
- **LarsArtmann/gogenfilter/v3**: Auto-generated code detection and filtering
- **go.yaml.in/yaml/v3**: YAML parsing (Unmarshal, Marshal)
- **golang.org/x/mod/semver**: Semantic versioning (Compare, IsValid)
- **samber/mo**: **REMOVED** — `types.Result[T]` wrapper deleted; all functions now return idiomatic `(T, error)`. No external dependency for monad types.
- **spf13/afero**: **REMOVED** — replaced with minimal `config.FS` interface backed by `os` package. No external dependency for filesystem abstraction.

## Project-Specific Patterns

### Configuration Discovery Pattern

```go
// Search order
configFile, err := configLoader.FindConfigFile(".")
if err != nil {
    return fmt.Errorf("no config found: %w", err)
}

// Or get default if missing
configFile = configLoader.FindOrGetDefaultConfigPath(".")
```

### Analyzer Usage Pattern

```go
analyzer := linter.NewAnalyzer(logger)
err := analyzer.FindBinary(ctx)
if err != nil {
    return err
}
err = analyzer.CheckVersion(ctx)
if err != nil {
    return err
}
analysis, err := analyzer.AnalyzeConfig(ctx, configFile)
```

### Fixer Usage Pattern

```go
fixer := linter.NewFixer(logger, analyzer)
result, err := fixer.FixConfig(ctx, configPath, priority, dryRun)
if err != nil {
    return err
}
logger.Infof("Result: %s", result.Message)
```

## Troubleshooting

### Build Failures

1. Check Go version: `go version` (must be 1.26+)
2. Run `just tidy` to update dependencies
3. Check for local replace in go.mod (`go-finding => ../go-finding` may need adjustment)

### Test Failures

1. Run `just test` with verbose output: `ginkgo -v ./pkg/...`
2. Check test coverage: `just coverage-html`
3. Verify all dependencies installed: `just deps`

### Linter Failures

1. Check golangci-lint version: `golangci-lint version` (must be v2.10.1+)
2. Run `just fmt-check` before `just lint`
3. Check for deprecated APIs (e.g., cobra.ExactValidArgs)

### Runtime Issues

1. Enable verbose logging: `--verbose` flag
2. Check binary path: `which golangci-lint-auto-configure`
3. Verify config file exists: `ls .golangci.yml`
4. Check version info: `./bin/golangci-lint-auto-configure --help`

## gogenfilter Integration

The tool uses [gogenfilter/v3](https://github.com/LarsArtmann/gogenfilter) to automatically detect and exclude auto-generated Go files from linting.

### How It Works

1. During `configure`, the scanner (`pkg/gogenfilter/scanner.go`) walks all `.go` files in the project
2. Uses gogenfilter's two-phase detection (filename first, content second) to identify generated files
3. Derives glob patterns for each detected generator type
4. Injects patterns into both `linters.exclusions.paths` and `formatters.exclusions.paths`
5. Also sets `linters.exclusions.generated: lax` as a belt-and-suspenders approach

### Supported Generators

| Generator    | Pattern                                   | Detection Method               |
| ------------ | ----------------------------------------- | ------------------------------ |
| sqlc         | output dirs (from sqlc.yaml)              | Filename + content + config    |
| templ        | `**/*_templ.go`                           | Filename suffix + content      |
| protobuf     | `**/*.pb.go`                              | Filename suffix + content      |
| go-enum      | `**/*_enum.go`                            | Filename suffix + content      |
| deepcopy-gen | `**/zz_generated.*.go`                    | Filename prefix + content      |
| wire         | `**/wire_gen.go`                          | Filename suffix + content      |
| moq          | `**/*_moq.go`                             | Filename suffix + content      |
| mockgen      | `**/*_mock.go`                            | Filename suffix + content      |
| stringer     | `**/*_string.go`                          | Content marker                 |
| oapi-codegen | `**/*.gen.go`                             | Content marker                 |
| Generic      | (no pattern, handled by `generated: lax`) | `// Code generated by` comment |

### Key Files

| File                              | Purpose                                                                    |
| --------------------------------- | -------------------------------------------------------------------------- |
| `pkg/gogenfilter/scanner.go`      | Scan project, detect generators, derive exclusion patterns                 |
| `pkg/gogenfilter/scanner_test.go` | BDD tests for scanner                                                      |
| `pkg/linter/fixer_config.go`      | `updateGeneratedExclusions` method that integrates scanner into fixer flow |

### Integration Point

The scanner runs during `configure` (non-dry-run) in `Fixer.applyAndSave()`. The `configUpdater.updateGeneratedExclusions` method:

1. Scans the project directory using `gogenfilter.ScanProject()`
2. Sets `linters.exclusions.generated: lax` if empty
3. Sets `formatters.exclusions.generated: lax` if empty
4. Merges detected patterns into existing exclusion paths (deduplicating)

## go-finding Integration

The project uses [go-finding](https://github.com/larsartmann/go-finding) as a unified data model for static analysis results.

### Key Files

| File                            | Purpose                                                                            |
| ------------------------------- | ---------------------------------------------------------------------------------- |
| `pkg/finding/converter.go`      | Convert domain types (LinterRecommendation, ValidationError) to `finding.Finding`  |
| `pkg/finding/golangci_lint.go`  | Parse `golangci-lint run --out-format=json` output to Findings                     |
| `pkg/finding/detector.go`       | `ConfigAnalysisDetector` implementing `pipeline.Detector` for pipeline integration |
| `pkg/finding/diff_converter.go` | Convert `diff.Change` and `MigrationResult` to Findings                            |
| `pkg/finding/helpers.go`        | LSP, filter, merge, groupBy helper utilities                                       |
| `pkg/ui/finding_formatter.go`   | Terminal text formatting for go-finding objects                                    |

### Output Formats

- **`analyze --format sarif`**: SARIF 2.1.0 output (CI/CD integration)
- **`analyze --format finding`**: go-finding Report JSON (structured, with summary)
- **`report --format sarif`**: SARIF report file (`report.sarif.json`)
- **`report --format finding`**: go-finding Report file (`report.finding.json`)
- **`validate --format sarif`**: Validation errors as SARIF

### Priority-to-Severity Mapping

| LinterPriority | finding.Severity |
| -------------- | ---------------- |
| Critical       | `critical`       |
| High           | `error`          |
| Medium         | `warning`        |
| Optional       | `info`           |

### Linter-to-Category Mapping

Linters are mapped to go-finding categories: security, correctness, performance, complexity, duplication, error-handling, style, testing, type-safety, structure, configuration.

### go.mod Note

`go-finding` uses a local replace directive:

```
replace github.com/larsartmann/go-finding => ../go-finding
```

## External References

- **golangci-lint**: https://github.com/golangci/golangci-lint
- **Ginkgo**: https://onsi.github.io/ginkgo/
- **Gomega**: https://onsi.github.io/gomega/
- **Templ**: https://templ.guide/
- **Cobra**: https://github.com/spf13/cobra
- **go-finding**: https://github.com/larsartmann/go-finding
- **gogenfilter**: https://github.com/LarsArtmann/gogenfilter
