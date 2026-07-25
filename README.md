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

All tools (Go 1.26, ginkgo, golangci-lint, templ) are provided automatically.

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
- **golangci-lint**: v2.10.1+ minimum (v2.12.2+ recommended; tool warns if below recommended)
- **Git**: Must run inside a git repository (for version control)
- **ginkgo**: For running tests (`go install github.com/onsi/ginkgo/v2/ginkgo@latest`)
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

━━ golangci-lint Configuration ━━

  Config: .golangci.yml

✓ All recommended linters are already enabled
━━ Summary ━━

  Critical:        CRITICAL  0
  High Priority:   HIGH      0
  Medium Priority: MEDIUM    0
  Optional:        OPTIONAL  0
  Enabled:         ✓ 109
  Disabled:        ✗ 5
```

When issues are found, the summary shows disabled linters grouped by priority with explanations.
(Actual output includes color coding via lipgloss.)

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

# Use a preset linter set
golangci-lint-auto-configure configure --preset minimal      # Essential only
golangci-lint-auto-configure configure --preset standard     # Balanced (default behavior)
golangci-lint-auto-configure configure --preset strict       # Maximum coverage
golangci-lint-auto-configure configure --preset reference    # All critical + high linters
golangci-lint-auto-configure configure --preset format       # Core formatters + essential linters

# Auto-detect project type and apply appropriate preset
golangci-lint-auto-configure configure --detect

# CI check mode — exit 1 if changes needed, 0 if optimal
golangci-lint-auto-configure configure --check

# Preview config changes as a diff before applying
golangci-lint-auto-configure configure --diff

# Combine flags
golangci-lint-auto-configure configure --check --diff        # CI diff preview
golangci-lint-auto-configure configure --dry-run --diff      # Preview without saving
```

> **Note:** When combining `--check` with `--diff`, the tool temporarily applies changes to compute the diff, then restores the original config before exiting. The file on disk is never modified.

**Automatic Deprecation Handling:**
The tool automatically detects and replaces deprecated linters with their recommended successors:

- `wsl` → `wsl_v5` (original wsl is deprecated since golangci-lint v2.2.0)
- `gomodguard` → `gomodguard_v2` (requires golangci-lint v2.12.0+)

**Safety Features:**

- **Git-based version control** - Requires running in a git repository
- Git provides full history and rollback capabilities
- Idempotent - safe to run multiple times

### Audit Trail

Every config change is recorded in an append-only audit ledger at
`~/.cache/golangci-lint-auto-configure/audit.jsonl` (outside the git tree,
tamper-resistant across branch switches). Review what automated runs changed:

```bash
# Show recent config changes in a table
golangci-lint-auto-configure audit

# Filter by linter or time
golangci-lint-auto-configure audit --linter mnd --since 7d

# JSON output for scripts
golangci-lint-auto-configure audit --json

# Clear the ledger
golangci-lint-auto-configure audit --clear
```

Entries older than 90 days are automatically purged. Disable the ledger with
`--no-audit` or `GOLANGCI_LINT_AUTO_CONFIGURE_NO_AUDIT=1`.

### Disable-Reason Enforcement (Anti-Gaming)

To prevent AI agents from silently disabling linters to claim "0 findings,"
create a `.golangci-lint-auto-configure.yml` sidecar file next to your
`.golangci.yml`. When this file exists, the tool enforces that every linter in
`linters.disable` has a justification entry — unjustified disables are
re-enabled automatically.

```yaml
# .golangci-lint-auto-configure.yml
disabled:
  mnd:
    reason: "false-positives in file permissions like 0o644 and display widths"
    category: false-positives
  varnamelen:
    reason: "idiomatic short names like tc, r, wg, mu are standard Go"
    category: convention
```

Categories: `false-positives`, `superseded`, `convention`, `performance`, `other`.

Without a sidecar file, all disables are respected (backward compatible). Commit
the sidecar to git so the team shares a baseline — agents must add entries to
justify new disables, which is visible in code review.

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
          go-version-file: go.mod
      - name: Install golangci-lint-auto-configure
        run: go install github.com/larsartmann/golangci-lint-auto-configure/cmd/golangci-lint-auto-configure@latest
      - name: Check config is optimal
        run: golangci-lint-auto-configure configure --check
      - name: Run linters
        run: golangci-lint run ./...
```

## Exit Codes

The CLI uses BSD `sysexits.h` exit codes for CI/CD integration, powered by [go-error-family](https://github.com/larsartmann/go-error-family).

| Code | Constant                  | Meaning        | When                                                                          |
| ---- | ------------------------- | -------------- | ----------------------------------------------------------------------------- |
| 0    | `EX_OK`                   | Success        | Command completed successfully                                                |
| 1    | `EX_USAGE` / `EX_DATAERR` | User error     | Bad input, missing config, invalid version, `--check` detected needed changes |
| 65   | `EX_DATAERR`              | Corruption     | Unparseable golangci-lint output (broken installation)                        |
| 69   | `EX_UNAVAILABLE`          | Infrastructure | golangci-lint binary not found in PATH                                        |
| 75   | `EX_TEMPFAIL`             | Transient      | Temporary failure (retry in CI)                                               |

CI pipelines can branch on these codes:

```bash
golangci-lint-auto-configure configure --check
exit_code=$?
case $exit_code in
  0)  echo "Config is optimal" ;;
  1)  echo "Changes needed or user error" ;;
  65) echo "Corrupted golangci-lint installation" ;;
  69) echo "golangci-lint not installed" ;;
  75) echo "Transient failure, retry" ;;
esac
```

## Commands

| Command        | Description                                    |
| -------------- | ---------------------------------------------- |
| `configure`    | Auto-configure golangci-lint (default command) |
| `analyze`      | Analyze configuration and show recommendations |
| `validate`     | Validate existing configuration                |
| `report`       | Generate JSON/HTML/SARIF report                |
| `audit`        | Show the audit trail of config changes         |
| `migrate`      | Migrate configuration from v1 to v2 schema     |
| `install-hook` | Install pre-commit hook for git                |
| `completion`   | Generate shell completion script               |

## Flags

| Flag              | Description                                                                        |
| ----------------- | ---------------------------------------------------------------------------------- |
| `-c, --config`    | Path to golangci-lint config file                                                  |
| `-d, --dry-run`   | Show what would be done without making changes                                     |
| `--check`         | CI mode: exit 1 if changes needed, 0 if optimal                                    |
| `--diff`          | Show diff of config changes before applying                                        |
| `--priority`      | Minimum priority level (critical, high, medium, optional)                          |
| `--preset`        | Use a preset (minimal, standard, strict, security, performance, reference, format) |
| `--detect`        | Auto-detect project type and select appropriate preset                             |
| `-v, --verbose`   | Enable verbose output                                                              |
| `--format`        | Output format (html, json, sarif, finding)                                         |
| `--output`        | Output path for report file                                                        |
| `--no-auto-merge` | Disable automatic merging of multiple config files                                 |
| `--no-audit`      | Skip writing to the audit ledger                                                   |

## Project-Specific Examples

The `examples/` directory contains optimized configurations for different project types:

- **minimal.golangci.yml** - Small projects, fast linting
- **standard.golangci.yml** - Most projects, balanced coverage
- **web-project.golangci.yml** - HTTP servers, REST APIs
- **cli-project.golangci.yml** - Command-line tools
- **library.golangci.yml** - Reusable packages, SDKs

## Linter Priorities

The tool classifies 100+ linters into four priority levels. Below is a curated
highlight of the most important ones. Run `golangci-lint-auto-configure analyze`
to see the full classification for your project.

### Critical (Always Enable)

Security and correctness linters that should never be disabled:

- `gosec` - Security vulnerability scanning
- `errcheck` - Unchecked error detection
- `staticcheck` - Advanced static analysis
- `govet` - Go vet suspicious constructs
- `musttag` - Enforces struct tags for JSON/XML/YAML marshaling
- `noctx` - Detects functions that don't use context
- `sloglint` - Structured logging best practices

<details><summary>All 11 Critical linters</summary>

`loggercheck`, `gosec`, `errcheck`, `staticcheck`, `govet`, `errchkjson`, `musttag`, `sloglint`, `nilerr`, `noctx`, `paralleltest`

</details>

### High Value (Recommended)

Quality and maintainability linters:

- `errorlint` - Error handling patterns
- `exhaustive` - Enum exhaustiveness checks
- `wrapcheck` - Error wrapping validation
- `ineffassign` - Detects unused assignments
- `forcetypeassert` - Detects forced type assertions
- `revive` - Fast, configurable, extensible linter
- `misspell` - Typos detection
- `gocyclo` - Cyclomatic complexity

<details><summary>All 50+ High value linters</summary>

`wrapcheck`, `errorlint`, `prealloc`, `unconvert`, `ineffassign`, `gocyclo`, `funlen`, `cyclop`, `gocognit`, `maintidx`, `exhaustive`, `exhaustruct`, `goconst`, `misspell`, `revive`, `nolintlint`, `forcetypeassert`, `gocritic`, `unused`, `bodyclose`, `contextcheck`, `dupl`, `durationcheck`, `errname`, `gochecknoglobals`, `gochecknoinits`, `gosmopolitan`, `interfacebloat`, `nestif`, `nilnil`, `nakedret`, `predeclared`, `reassign`, `rowserrcheck`, `spancheck`, `sqlclosecheck`, `testifylint`, `thelper`, `unparam`, `wastedassign`, `copyloopvar`, `ginkgolinter`, `gochecksumtype`, `intrange`, `mirror`, `perfsprint`, `protogetter`, `usetesting`, `recvcheck`, `nilnesserr`, `zerologlint`

</details>

### Medium Value (Optional)

Style and consistency linters:

- `varnamelen` - Variable name length rules
- `tagliatelle` - Struct tag style enforcement
- `mnd` - Magic number detector
- `usestdlibvars` - Detects stdlib variable usage
- `nonamedreturns` - Enforces named returns policy

<details><summary>All 50+ Medium value linters</summary>

`dupword`, `godot`, `godox`, `goheader`, `varnamelen`, `whitespace`, `wsl_v5`, `grouper`, `dogsled`, `makezero`, `asciicheck`, `bidichk`, `containedctx`, `decorder`, `forbidigo`, `godoclint`, `gomoddirectives`, `gomodguard`, `gomodguard_v2`, `ireturn`, `lll`, `mnd`, `nlreturn`, `nonamedreturns`, `promlinter`, `tagliatelle`, `testpackage`, `tparallel`, `unqueryvet`, `usestdlibvars`, `asasalint`, `canonicalheader`, `err113`, `exptostd`, `fatcontext`, `gocheckcompilerdirectives`, `goprintffuncname`, `iface`, `inamedparam`, `iotamixing`, `modernize`, `nosprintfhostport`, `tagalign`, `testableexamples`, `importas`, `arangolint`, `embeddedstructfieldcheck`, `clickhouselint`

</details>

## Exclusion Patterns

golangci-lint v2 uses [RE2 regex](https://github.com/google/re2/wiki/Syntax) for exclusion paths. The tool automatically injects these defaults into your config:

**Linter exclusion paths:**

```yaml
linters:
  exclusions:
    paths:
      - '_templ\.go$' # Templ generated files
      - '\.gen\.go$' # Generic generated files
      - "vendor/" # Vendored dependencies
```

**Formatter exclusion paths:**

```yaml
formatters:
  exclusions:
    paths:
      - '_templ\.go$' # Templ generated files
```

**Common RE2 patterns:**

| Pattern             | Matches                                   |
| ------------------- | ----------------------------------------- |
| `_test\.go`         | All test files                            |
| `_templ\.go$`       | Templ generated files (end of filename)   |
| `\.gen\.go$`        | Generic generated files (end of filename) |
| `\.pb\.go$`         | Protobuf generated files                  |
| `vendor/`           | Vendored dependencies (prefix match)      |
| `^(cmd\|internal)/` | Files in cmd/ or internal/ directories    |
| `.*_string\.go$`    | Stringer generated files                  |

Note: Unlike standard regex, RE2 anchors like `$` are literal — the pattern is matched against the full file path. Unanchored patterns (like `vendor/`) match any path containing that substring.

## Testing

```bash
# Run all tests with race detection

go test -race ./pkg/... ./internal/...

# Run with verbose output
ginkgo -v ./...

# Generate and view HTML coverage report
go test -coverprofile=coverage.out ./pkg/... ./internal/... && go tool cover -html=coverage.out

# Check test coverage summary
go test -cover ./pkg/... ./internal/...
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
2. Create a feature branch (`git switch -c feature/amazing-feature`)
3. Make your changes and add tests
4. Ensure all tests pass (`go test -race ./pkg/... ./internal/...`)
5. Run linters (`golangci-lint run --config=.golangci.yml --timeout=5m`)
6. Submit a pull request

## License

MIT License - see LICENSE file for details

## Related Projects

- [golangci-lint](https://github.com/golangci/golangci-lint) - The Go linters aggregator
- [go-finding](https://github.com/LarsArtmann/go-finding) - Unified finding model and SARIF output for static analysis tools
