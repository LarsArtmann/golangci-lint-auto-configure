# Working with This Codebase

## Adding a New CLI Command

1. Define command function in `internal/cli/cmd_*.go` (for configure/analyze/validate/report) or `internal/cli/cmd/*.go` (for migrate/install-hook/completion)
2. Wire up dependencies in `addSubCommands()` in `internal/cli/commands.go`
3. Add flags as needed
4. Write BDD tests in `internal/cli/commands_test.go` or `internal/cli/integration_test.go`
5. Run `just test` to verify

## Modifying Linter Priorities

1. Edit `pkg/constants/linter_priorities.go`
2. Update `LinterPriorities` map
3. Edit `pkg/constants/linter_reasons.go` to update `LinterReasons` map
4. Consider updating presets in `pkg/constants/presets.go`
5. Run tests: `just test`

## Adding New Linter Data

1. Add linter to `pkg/constants/linter_priorities.go` (`LinterPriorities` map)
2. Set priority (Critical/High/Medium/Optional)
3. Add reason to `pkg/constants/linter_reasons.go` (`LinterReasons` map)
4. Consider if deprecated (add to `DeprecatedLinters` in `pkg/constants/rules.go`)
5. Regenerate reports if needed

## Updating HTML Report

1. Edit `pkg/report/report.templ`
2. Run `templ generate` to compile to Go
3. Test report generation: `./bin/golangci-lint-auto-configure report`
4. Verify HTML output in browser

## Debugging Issues

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
