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

-> See docs/references/code-organization.md for full directory structure.

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

-> See docs/references/code-organization.md for patterns (interface-based, strong typing, data-driven, go-finding).

## Testing Approach

-> See docs/references/testing-style-and-patterns.md for BDD testing patterns and coverage commands.
Uses Ginkgo v2 + Gomega. Run: just test (= ginkgo -r --cover).

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

Custom error types (`ConfigError`, `AnalysisError`, `ReportError`) in `pkg/errors/errors.go`.
Wrap errors with `%w`, use custom types for domain errors, provide actionable messages.

→ See [`docs/references/error-handling.md`](docs/references/error-handling.md) for full patterns and examples.

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

-> See docs/references/testing-style-and-patterns.md for code style, naming, and logging conventions.

## CI/CD Pipeline

-> See docs/references/testing-style-and-patterns.md for GitHub Actions and pre-commit hook details.

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
- Nix builds copy go-finding from `git+ssh://` flake input; `postPatch` redirects to `./go-finding-vendor`
- Local dev uses `../go-finding` directly; CI fetches via SSH from GitHub (private repo)

### 8. Cobra Deprecation

Previously used deprecated `cobra.ExactValidArgs()` — has been fixed. Check current codebase if similar warnings appear.

### 9. Config File Auto-Creation

`configure` creates a default `.golangci.yml` if missing. Requires git repository for version control safety.

### 10. Fixer Counting & Issues Normalization

Every config mutation in `applyAndSave` MUST increment `fixCounts.normalization` — otherwise the
`counts.total()==0` guard silently discards changes. The fixer injects
`issues.max-issues-per-linter: 50` and `max-same-issues: 10` when absent, preventing
golangci-lint's default `max-same-issues: 3` from hiding CI problems.

## Working with This Codebase

→ See [`docs/references/working-with-codebase.md`](docs/references/working-with-codebase.md) for:

- Adding new CLI commands, modifying linter priorities, adding linter data
- Common tasks (dev workflow, release prep, testing, fixing linter issues)
- Debugging and troubleshooting guides

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

-> See docs/references/testing-style-and-patterns.md for configuration discovery, analyzer, and fixer patterns.
