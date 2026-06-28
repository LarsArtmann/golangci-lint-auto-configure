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
go test -race ./pkg/... ./internal/...  # Run all tests (primary)
ginkgo -r --cover                      # Alternative: ginkgo directly
ginkgo -v ./pkg/...                    # Run with verbose output
ginkgo -r --focus="Name"              # Run specific tests by name
ginkgo -r --skip="Name"              # Skip specific tests
```

### Test Patterns

1. **Setup in `BeforeEach`**: Initialize test state
2. **BDD naming**: Use `Context` for scenarios, `It` for specific behaviors
3. **Matchers**: Use Gomega's `Expect().To(Equal())` syntax
4. **Table-driven tests**: Common for multiple test cases
5. **Error testing**: Test both success and failure paths

### Coverage

```bash
# Generate and open HTML coverage report
go test -coverprofile=coverage.out ./pkg/... ./internal/... && go tool cover -html=coverage.out
# Show coverage summary in terminal
go test -cover ./pkg/... ./internal/...
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
