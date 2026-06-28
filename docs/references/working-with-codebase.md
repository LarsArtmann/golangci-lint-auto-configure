# Working with This Codebase

## Adding a New CLI Command

1. Define command function in `internal/cli/cmd_*.go` (for configure/analyze/validate/report) or `internal/cli/cmd/*.go` (for migrate/install-hook/completion)
2. Wire up dependencies in `addSubCommands()` in `internal/cli/commands.go`
3. Add flags as needed
4. Write BDD tests in `internal/cli/commands_test.go` or `internal/cli/integration_test.go`
5. Run `go test -race ./pkg/... ./internal/...` to verify

## Modifying Linter Priorities

1. Edit `pkg/constants/linter_priorities.go`
2. Update `LinterPriorities` map
3. Edit `pkg/constants/linter_reasons.go` to update `LinterReasons` map
4. Consider updating presets in `pkg/constants/presets.go`
5. Run tests: `go test -race ./pkg/... ./internal/...`

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
4. Check test coverage: `go test -coverprofile=coverage.out ./pkg/... ./internal/... && go tool cover -html=coverage.out`
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

- `bin/`: Build output (cleaned by `rm -rf bin/`)
- `coverage.out`: Coverage profile (from `go test -coverprofile`)
- `coverage.html`: HTML coverage report (from `go tool cover -html`)
- `report.html`: Generated analysis report (from CLI)

## Common Tasks

### Full Development Workflow

```bash
go mod tidy                                          # Update dependencies
nix fmt                                              # Format code (Go, Nix, templ)
go test -race ./pkg/... ./internal/...               # Run tests
golangci-lint run --config=.golangci.yml --timeout=5m # Run linters
nix build                                            # Build binary (preferred)
# Or: go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure
./bin/golangci-lint-auto-configure --help
```

### Release Preparation

```bash
go test -race ./pkg/... ./internal/...                # All tests must pass
golangci-lint run --config=.golangci.yml --timeout=5m  # All linters must pass
# Check coverage:
go test -coverprofile=coverage.out ./pkg/... ./internal/... && go tool cover -func=coverage.out | grep total
git tag v0.1.0
nix build                                             # Build with version tag
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
golangci-lint run --config=.golangci.yml --timeout=5m

# Fix specific issues manually or use golangci-lint
golangci-lint run --fix

# Verify fixes
go test -race ./pkg/... ./internal/... && golangci-lint run --config=.golangci.yml --timeout=5m
```

## Troubleshooting

### Build Failures

1. Check Go version: `go version` (must be 1.26+)
2. Run `go mod tidy` to update dependencies
3. For Nix: after go.mod changes, update `vendorHash` in `flake.nix` (run `nix build`, copy the `got:` hash)

### Test Failures

1. Run with verbose output: `ginkgo -v ./pkg/...`
2. Check test coverage: `go test -coverprofile=coverage.out ./pkg/... ./internal/... && go tool cover -html=coverage.out`
3. Verify all dependencies installed: `go mod download`

### Linter Failures

1. Check golangci-lint version: `golangci-lint version` (must be v2.10.1+)
2. Run `nix fmt` (or `treefmt --ci`) to check formatting before linting
3. Check for deprecated APIs (e.g., cobra.ExactValidArgs)

### Runtime Issues

1. Enable verbose logging: `--verbose` flag
2. Check binary path: `which golangci-lint-auto-configure`
3. Verify config file exists: `ls .golangci.yml`
4. Check version info: `./bin/golangci-lint-auto-configure --help`
