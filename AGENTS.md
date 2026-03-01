# Agent Guide: golangci-linter-auto-configure

This guide provides essential information for agents working on the golangci-linter-auto-configure codebase.

## Project Overview

**golangci-linter-auto-configure** is a Go CLI tool that automatically configures and optimizes golangci-lint configurations by:

- Analyzing existing golangci-lint configs
- Detecting missing linters with smart categorization
- Recommending optimal linter settings based on project type
- Auto-fixing configuration issues
- Automatically replacing deprecated linters with their successors
- Generating HTML/JSON reports

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

**CLI Commands (after build):**

```bash
./bin/golangci-linter-auto-configure configure [--priority critical|high|medium|optional] [--dry-run]
./bin/golangci-linter-auto-configure analyze [--config .golangci.yml]
./bin/golangci-linter-auto-configure validate [--config .golangci.yml]
./bin/golangci-linter-auto-configure report [--format html|json] [--output path]
./bin/golangci-linter-auto-configure migrate [--skip-validation]
./bin/golangci-linter-auto-configure restore --backup-path .golangci.yml.backup
./bin/golangci-linter-auto-configure install-hook
```

## Technology Stack

### Core Dependencies

- **Go**: 1.25+ (CI tests on 1.25 and 1.26)
- **Cobra**: CLI command framework
- **Charmbracelet Log**: Structured logging
- **Charmbracelet Fang**: Enhanced CLI features
- **Ginkgo v2 + Gomega**: BDD testing framework (NOT standard Go testing)
- **Templ**: HTML template system for reports
- **YAML v3**: Configuration parsing
- **Universal Workflow**: Workflow orchestration (local replace)

### External Tools Required

- **golangci-lint**: v2.10.1+ (auto-detected, minimum version enforced)
- **Go**: 1.25+ required for compilation

## Code Organization

### Directory Structure

```
golangci-linter-auto-configure/
├── cmd/
│   └── golangci-linter-auto-configure/
│       └── main.go                    # Entry point, sets version via ldflags
├── pkg/
│   ├── types/                        # Core type definitions and interfaces
│   │   ├── types.go                  # Main types: LinterPriority, Config, LinterInfo, etc.
│   │   └── result.go                # Result types
│   ├── constants/
│   │   └── linter_data.go           # Linter priorities, reasons, presets, replacements
│   ├── config/
│   │   ├── loader.go                # Load, save, validate golangci-lint configs
│   │   └── loader_test.go           # Config loading tests
│   ├── linter/
│   │   ├── analyzer.go              # Analyze configs, get recommendations
│   │   ├── fixer.go                # Apply fixes to configs
│   │   └── analyzer_test.go        # Analyzer tests
│   ├── detection/
│   │   ├── detector.go             # Detect project type (CLI, web, library, etc.)
│   │   └── detector_test.go        # Detector tests
│   ├── diff/
│   │   ├── differ.go               # Compare two configs and show changes
│   │   └── differ_test.go          # Diff tests
│   ├── report/
│   │   ├── generator.go            # HTML report generation (templ-based)
│   │   ├── json_report_generator.go # JSON report generation
│   │   └── report.templ           # HTML template (generates Go code)
│   ├── workflow/
│   │   └── workflow.go            # Workflow orchestration (uses universal-workflow)
│   └── errors/
│       └── errors.go              # Custom error types (ConfigError, AnalysisError)
├── internal/
│   ├── cli/
│   │   ├── commands.go             # All CLI command definitions
│   │   └── commands_test.go        # Command tests
│   └── di/                        # Dependency injection (currently empty)
├── examples/                       # Example configurations for different project types
├── docs/                          # Documentation and status reports
├── reports/                       # Generated linter documentation
├── scripts/                       # Utility scripts
└── justfile                       # Build/test/lint commands (PRIMARY INTERFACE)
```

### Key Architectural Patterns

**1. Interface-Based Design (for testability)**
All major components implement interfaces defined in `pkg/types/types.go`:

- `ConfigLoader`: Load, save, validate configs
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

- `pkg/constants/linter_data.go` contains all linter metadata:
  - Priority levels
  - Human-readable reasons
  - Preset configurations
  - Deprecated linter replacements

**5. Workflow Orchestration**

- `pkg/workflow/workflow.go` uses `universal-workflow` for complex operations
- Supports multi-step operations: analyze → validate → report

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

Update `pkg/constants/linter_data.go`:

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

### Backup Strategy

- Automatic backup before any modification: `<config>.backup`
- Backup path returned in result, logged to user
- Restore command: `golangci-linter-auto-configure restore --backup-path <path>`

## Project Type Detection

### Supported Types

```go
const (
    ProjectTypeUnknown   // Cannot determine
    ProjectTypeCLI      // Command-line tools
    ProjectTypeLibrary  // Reusable packages, SDKs
    ProjectTypeWeb     // HTTP servers, REST APIs
    ProjectTypeAPI      # API services
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

Version checking in `pkg/linter/analyzer.go`:

1. Try `golangci-lint version --json` first (more reliable)
2. Fallback to text parsing if JSON not supported
3. Use `golang.org/x/mod/semver` for comparison
4. Fail with helpful error if version too old

### Handling Deprecated Linters

Automatic replacement of deprecated linters:

- `wsl` → `wsl_v5` (deprecated since golangci-lint v2.2.0)
- Mappings in `pkg/constants/linter_data.go`:
  ```go
  LinterReplacements = map[types.LinterName]LinterReplacement{
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

### 4. Dependency Injection Directory is Empty

- `internal/di/` exists but is unused
- Dependency injection is manual in CLI commands
- No DI framework like samber/do or wire

### 5. Linter Priority Data is in Constants

- All linter priorities in `pkg/constants/linter_data.go`
- Modify that file to change linter behavior
- Not dynamically computed from golangci-lint

### 6. Version String Injected at Build Time

- `main.version` variable injected via ldflags
- Justfile `install-local` does: `go build -ldflags "-X main.version=$VERSION"`
- Default value is "dev" if not set

### 7. Universal Workflow is Local Replace

- `go.mod` has: `replace github.com/LarsArtmann/universal-workflow => /Users/larsartmann/projects/universal-workflow`
- Path is user-specific, needs adjustment for different developers
- CI may not work with this local replace

### 8. Deprecated Cobra Usage

- `cobra.ExactValidArgs()` is deprecated (detected in commands.go:692)
- Should use `MatchAll(ExactArgs(n), OnlyValidArgs)` instead

### 9. Test Error in detector_test.go:98

- Error: "no new variables on left side of :="
- Warning currently present in project diagnostics
- Needs fixing before considering codebase clean

### 10. Config File Auto-Creation

- `configure` command creates default config if missing
- Uses `.golangci.yml` as default path
- Creates backup before modifying any existing config

## Working with This Codebase

### Adding a New CLI Command

1. Define command function in `internal/cli/commands.go`
2. Wire up dependencies in `NewRootCommand()`
3. Add flags as needed
4. Write BDD tests in `internal/cli/commands_test.go`
5. Run `just test` to verify

### Modifying Linter Priorities

1. Edit `pkg/constants/linter_data.go`
2. Update `LinterPriorities` map
3. Update `LinterReasons` map
4. Consider updating presets
5. Run tests: `just test`

### Adding New Linter Data

1. Add linter to `pkg/constants/linter_data.go`
2. Set priority (Critical/High/Medium/Optional)
3. Write reason for recommendation
4. Consider if deprecated (add to `LinterReplacements`)
5. Regenerate reports if needed

### Updating HTML Report

1. Edit `pkg/report/report.templ`
2. Run `templ generate` to compile to Go
3. Test report generation: `./bin/golangci-linter-auto-configure report`
4. Verify HTML output in browser

### Debugging Issues

1. Enable verbose mode: `--verbose` flag or `logger.SetLevel(log.DebugLevel)`
2. Check golangci-lint version: `golangci-lint version`
3. Verify config file exists and is valid: `golangci-linter-auto-configure validate`
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
./bin/golangci-linter-auto-configure --help
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
- **gopkg.in/yaml.v3**: YAML parsing (Unmarshal, Marshal)
- **golang.org/x/mod/semver**: Semantic versioning (Compare, IsValid)
- **samber/mo**: Functional programming utilities (monads, option types)

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
err := analyzer.FindBinary()
if err != nil {
    return err
}
err = analyzer.CheckVersion()
if err != nil {
    return err
}
analysis, err := analyzer.AnalyzeConfig(configFile)
```

### Fixer Usage Pattern

```go
fixer := linter.NewFixer(logger, analyzer)
result, err := fixer.FixConfig(configFile, priority, dryRun)
if err != nil {
    return err
}
logger.Infof("Result: %s", result.Message)
```

### Workflow Usage Pattern

```go
workflowBuilder := workflow.NewBuilder(logger, analyzer)
wf, err := workflowBuilder.BuildAutoConfigureWorkflow(configPath, dryRun, false, outputPath)
run, err := wf.Execute(ctx)
```

## Troubleshooting

### Build Failures

1. Check Go version: `go version` (must be 1.25+)
2. Run `just tidy` to update dependencies
3. Check for local replace in go.mod (may need adjustment)

### Test Failures

1. Run `just test` with verbose output: `ginkgo -v ./pkg/...`
2. Check test coverage: `just coverage-html`
3. Verify all dependencies installed: `just deps`

### Linter Failures

1. Check golangci-lint version: `golangci-lint version` (must be 2.8.0+)
2. Run `just fmt-check` before `just lint`
3. Check for deprecated APIs (e.g., cobra.ExactValidArgs)

### Runtime Issues

1. Enable verbose logging: `--verbose` flag
2. Check binary path: `which golangci-linter-auto-configure`
3. Verify config file exists: `ls .golangci.yml`
4. Check version info: `./bin/golangci-linter-auto-configure --help`

## External References

- **golangci-lint**: https://github.com/golangci/golangci-lint
- **Ginkgo**: https://onsi.github.io/ginkgo/
- **Gomega**: https://onsi.github.io/gomega/
- **Templ**: https://templ.guide/
- **Cobra**: https://github.com/spf13/cobra
- **Universal Workflow**: https://github.com/LarsArtmann/universal-workflow
