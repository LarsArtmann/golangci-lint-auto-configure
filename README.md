# golangci-linter-auto-configure

**Automatically configure, optimize, and maintain golangci-lint configurations for Go projects with smart linter recommendations, deprecation handling, and auto-fixing.**

## Purpose

This tool automatically configures golangci-lint for Go projects by:

- **Analyzing** existing golangci-lint configurations
- **Detecting** missing linters with smart categorization
- **Recommending** optimal linter settings based on project type
- **Auto-fixing** configuration issues
- **Automatically replacing deprecated linters** with their recommended successors
- **Generating** configuration templates

## Installation

```bash
# Install the latest version
go install github.com/larsartmann/golangcli-linter-auto-configure/cmd/golangci-linter-auto-configure@latest

# Or build from source
git clone https://github.com/larsartmann/golangcli-linter-auto-configure
cd golangci-linter-auto-configure
go build -o /usr/local/bin/golangci-linter-auto-configure ./cmd/golangci-linter-auto-configure
```

## Requirements

- **Go**: 1.25+
- **golangci-lint**: v2.10.1+ (tool checks version automatically)
- **Git**: Must run inside a git repository (for version control)

## Usage

### Analyze Your Configuration

See what linters you're missing:

```bash
# Analyze current directory
golangci-linter-auto-configure analyze

# Analyze specific config
golangci-linter-auto-configure analyze --config .golangci.yml

# Verbose output with debug logs
golangci-linter-auto-configure analyze --verbose
```

**Output Example:**

```
INFO Analyzing configuration: .golangci.yml

🚨 2 CRITICAL linter(s) are disabled (should ALWAYS be enabled):
  - musttag: Enforces struct tags for JSON/XML/YAML marshaling
  - noctx: Check whether function uses a non-inherited context

⚠️  1 HIGH VALUE linter(s) are disabled:
  - exhaustruct: Check if all struct fields are initialized

Summary: Found 16 disabled linters
```

### Auto-Configure Your Project

Automatically enable recommended linters:

```bash
# Dry-run to see what would change
golangci-linter-auto-configure configure --dry-run

# Apply changes
golangci-linter-auto-configure configure

# Configure with specific priority level
golangci-linter-auto-configure configure --priority critical   # Security only
golangci-linter-auto-configure configure --priority high       # Recommended (default)
golangci-linter-auto-configure configure --priority medium     # Include style linters
golangci-linter-auto-configure configure --priority optional   # All linters
```

**Automatic Deprecation Handling:**
The tool automatically detects and replaces deprecated linters with their recommended successors:

- `wsl` → `wsl_v5` (original wsl is deprecated since golangci-lint v2.2.0)

**Safety Features:**

- **Git-based version control** - Requires running in a git repository
- Git provides full history and rollback capabilities
- Idempotent - safe to run multiple times

### Validate Configuration

Check if your config is valid:

```bash
golangci-linter-auto-configure validate --config .golangci.yml
```

### Generate Reports

Create JSON reports for CI/CD:

```bash
golangci-linter-auto-configure report --output analysis.json --format json
```

## Example Workflows

### New Project Setup

```bash
# 1. Analyze what's missing
golangci-linter-auto-configure analyze

# 2. Apply recommendations
golangci-linter-auto-configure configure --priority high

# 3. Verify with golangci-lint
golangci-lint run ./...
```

### Using Example Configurations

```bash
# Copy an example config
cp examples/web-project.golangci.yml .golangci.yml

# Customize if needed
vim .golangci.yml

# Run the tool to optimize
golangci-linter-auto-configure configure --dry-run
golangci-linter-auto-configure configure
```

### CI/CD Integration

```yaml
# .github/workflows/lint.yml
name: Lint

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - name: Install golangci-linter-auto-configure
        run: go install github.com/larsartmann/golangcli-linter-auto-configure/cmd/golangci-linter-auto-configure@latest
      - name: Auto-configure
        run: golangci-linter-auto-configure configure
      - name: Run linters
        run: golangci-lint run ./...
```

## Commands

| Command     | Description                                    |
| ----------- | ---------------------------------------------- |
| `configure` | Auto-configure golangci-lint (default command) |
| `analyze`   | Analyze configuration and show recommendations |
| `validate`  | Validate existing configuration                |
| `report`    | Generate JSON/HTML report                      |
| `migrate`   | Migrate config to v2.8+ schema                 |

## Flags

| Flag            | Description                                               |
| --------------- | --------------------------------------------------------- |
| `-c, --config`  | Path to golangci-lint config file                         |
| `-d, --dry-run` | Show what would be done without making changes            |
| `--priority`    | Minimum priority level (critical, high, medium, optional) |
| `-v, --verbose` | Enable verbose output                                     |
| `--format`      | Output format for report (html, json)                     |
| `--output`      | Output path for report file                               |

## Project-Specific Examples

The `examples/` directory contains optimized configurations for different project types:

- **minimal.golangci.yml** - Small projects, fast linting
- **standard.golangci.yml** - Most projects, balanced coverage
- **web-project.golangci.yml** - HTTP servers, REST APIs
- **cli-project.golangci.yml** - Command-line tools
- **library.golangci.yml** - Reusable packages, SDKs

## Linter Priorities

### Critical (Always Enable)

Security and correctness linters that should never be disabled:

- `gosec` - Security vulnerability scanning
- `errcheck` - Unchecked error detection
- `staticcheck` - Advanced static analysis
- `govet` - Go vet suspicious constructs
- `ineffassign` - Detects unused assignments

### High Value (Recommended)

Quality and maintainability linters:

- `errorlint` - Error handling patterns
- `exhaustive` - Enum exhaustiveness checks
- `wrapcheck` - Error wrapping validation
- `forcetypeassert` - Detects forced type assertions

### Medium Value (Optional)

Style and consistency linters:

- `gocyclo` - Cyclomatic complexity
- `misspell` - Typos detection
- `revive` - Fast, configurable linter
- `varnamelen` - Variable name length rules

## Testing

```bash
# Run all tests
ginkgo -r --cover

# Run with verbose output
ginkgo -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Building from Source

```bash
# Clone and build
git clone https://github.com/larsartmann/golangcli-linter-auto-configure
cd golangci-linter-auto-configure
go build -o bin/golangci-linter-auto-configure ./cmd/golangci-linter-auto-configure

# Run locally
./bin/golangci-linter-auto-configure --help
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Ensure all tests pass (`ginkgo -r`)
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Related Projects

- [golangci-lint](https://github.com/golangci/golangci-lint) - The Go linters aggregator
- [universal-workflow](https://github.com/LarsArtmann/universal-workflow) - Workflow orchestration
