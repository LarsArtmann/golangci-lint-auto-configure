# golangci-lint-auto-configure

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

### Quick Start with Nix (Recommended)

1. [Install Nix](https://nixos.org/download.html)
2. Enable Flakes: `mkdir -p ~/.config/nix && echo 'experimental-features = nix-command flakes' >> ~/.config/nix/nix.conf`
3. Clone this repo
4. Run `nix develop` — you're ready to go

All tools (Go 1.26, ginkgo, golangci-lint, just, templ) are provided automatically.

```bash
nix build                            # Build the CLI binary
./result/bin/golangci-lint-auto-configure --version
nix run . -- analyze                 # Run directly
```

### Install without Nix

```bash
# Install the latest version
go install github.com/larsartmann/golangci-lint-auto-configure/cmd/golangci-lint-auto-configure@latest

# Or build from source
git clone https://github.com/larsartmann/golangci-lint-auto-configure
cd golangci-lint-auto-configure
go build -o /usr/local/bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure
```

## Requirements

### With Nix (Recommended)

- **Nix** with Flakes enabled — all other tools are provided automatically

### Without Nix

- **Go**: 1.26+
- **golangci-lint**: v2.10.1+ (tool checks version automatically)
- **Git**: Must run inside a git repository (for version control)
- **ginkgo**: For running tests (`go install github.com/onsi/ginkgo/v2/ginkgo@latest`)
- **just**: For running recipes (`brew install just` or `go install github.com/casey/just@latest`)
- **templ**: For report template generation (`go install github.com/a-h/templ/cmd/templ@latest`)

## Usage

### Analyze Your Configuration

See what linters you're missing:

```bash
# Analyze current directory
golangci-lint-auto-configure analyze

# Analyze specific config
golangci-lint-auto-configure analyze --config .golangci.yml

# Verbose output with debug logs
golangci-lint-auto-configure analyze --verbose
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

### SARIF Output (GitHub Code Scanning)

Generate [SARIF](https://sarifweb.azurewebsites.net/) output for GitHub Code Scanning, Azure DevOps, or any SARIF-compatible tool:

```bash
# Analyze and output SARIF
golangci-lint-auto-configure analyze --format sarif

# Generate SARIF report file
golangci-lint-auto-configure report --format sarif --output results.sarif

# Validate and output SARIF
golangci-lint-auto-configure validate --format sarif

# Unified finding JSON (go-finding format)
golangci-lint-auto-configure analyze --format finding
```

**GitHub Actions integration:**

```yaml
- name: Run analysis
  run: golangci-lint-auto-configure analyze --format sarif > results.sarif

- name: Upload to GitHub Code Scanning
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: results.sarif
```

### Auto-Configure Your Project

Automatically enable recommended linters:

```bash
# Dry-run to see what would change
golangci-lint-auto-configure configure --dry-run

# Apply changes
golangci-lint-auto-configure configure

# Configure with specific priority level
golangci-lint-auto-configure configure --priority critical   # Security only
golangci-lint-auto-configure configure --priority high       # Recommended (default)
golangci-lint-auto-configure configure --priority medium     # Include style linters
golangci-lint-auto-configure configure --priority optional   # All linters
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
golangci-lint-auto-configure validate --config .golangci.yml
```

### Generate Reports

Create JSON reports for CI/CD:

```bash
golangci-lint-auto-configure report --output analysis.json --format json
```

## Example Workflows

### New Project Setup

```bash
# 1. Analyze what's missing
golangci-lint-auto-configure analyze

# 2. Apply recommendations
golangci-lint-auto-configure configure --priority high

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
golangci-lint-auto-configure configure --dry-run
golangci-lint-auto-configure configure
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
          go-version: "1.26"
      - name: Install golangci-lint-auto-configure
        run: go install github.com/larsartmann/golangci-lint-auto-configure/cmd/golangci-lint-auto-configure@latest
      - name: Auto-configure
        run: golangci-lint-auto-configure configure
      - name: Run linters
        run: golangci-lint run ./...
```

## Commands

| Command        | Description                                    |
| -------------- | ---------------------------------------------- |
| `configure`    | Auto-configure golangci-lint (default command) |
| `analyze`      | Analyze configuration and show recommendations |
| `validate`     | Validate existing configuration                |
| `report`       | Generate JSON/HTML/SARIF report                |
| `migrate`      | Migrate configuration from v1 to v2 schema     |
| `install-hook` | Install pre-commit hook for git                |
| `completion`   | Generate shell completion script               |

## Flags

| Flag            | Description                                               |
| --------------- | --------------------------------------------------------- |
| `-c, --config`  | Path to golangci-lint config file                         |
| `-d, --dry-run` | Show what would be done without making changes            |
| `--priority`    | Minimum priority level (critical, high, medium, optional) |
| `-v, --verbose` | Enable verbose output                                     |
| `--format`      | Output format (html, json, sarif, finding)                |
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
git clone https://github.com/larsartmann/golangci-lint-auto-configure
cd golangci-lint-auto-configure
go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure

# Run locally
./bin/golangci-lint-auto-configure --help
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
- [go-finding](https://github.com/LarsArtmann/go-finding) - Unified finding model and SARIF output for static analysis tools
- [universal-workflow](https://github.com/LarsArtmann/universal-workflow) - Workflow orchestration
