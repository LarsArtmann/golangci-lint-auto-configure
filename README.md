# golangcli-linter-auto-configure

**Automatically configure and optimize golangci-lint with smart recommendations, missing linter detection, and auto-fixing capabilities.**

## 🎯 Purpose

This tool automatically configures golangci-lint for Go projects by:

- **Analyzing** existing golangci-lint configurations
- **Detecting** missing linters and formatters with smart categorization
- **Recommending** optimal linter settings based on project needs
- **Auto-fixing** common configuration issues (v2.8+ schema migrations)
- **Generating** comprehensive configuration templates
- **Providing** beautiful HTML reports via templ components

## ✨ Key Features

### Intelligent Linter Configuration

- **Smart categorization**: 4-tier priority system (Critical, High, Medium, Optional)
- **Auto-discovery**: Detects project characteristics (test framework, code generation tools)
- **Missing linter detection**: Identifies disabled linters with actionable recommendations
- **Formatter integration**: Checks for conflicting or redundant formatters

### Configuration Management

- **Auto-migration**: Migrates configs to golangci-lint v2.8+ schema
- **Backup & restore**: Automatic backup before modifications
- **Validation**: Verifies configurations with `golangci-lint config verify`
- **Template generation**: Creates optimized config templates based on project type

### Developer Experience

- **Universal Workflow integration**: Built on proven workflow orchestration library
- **Beautiful CLI**: Styled output with Fang framework
- **HTML reports**: Interactive templ-generated reports
- **Robust error handling**: Context-aware errors with recovery
- **Comprehensive testing**: Ginkgo BDD tests with high coverage

## 🚀 Installation

```bash
go install github.com/larsartmann/golangcli-linter-auto-configure/cmd/golangci-linter-auto-configure@latest
```

## 💻 Usage

### Basic Usage

```bash
# Auto-configure golangci-lint for current project
golangci-linter-auto-configure

# Show recommendations without applying changes
golangci-linter-auto-configure --dry-run

# Generate HTML report
golangci-linter-auto-configure --output-report report.html

# Verbose mode
golangci-linter-auto-configure --verbose
```

### Advanced Usage

```bash
# Migrate configuration to v2.8 schema
golangci-linter-auto-configure migrate --config .golangci.yml

# Generate optimized configuration
golangci-linter-auto-configure generate --config .golangci.yml --preset enterprise

# Validate configuration
golangci-linter-auto-configure validate --config .golangci.yml

# Show missing linters
golangci-linter-auto-configure check-missing
```

## 📋 Commands

| Command      | Description                                   |
| ------------ | --------------------------------------------- |
| `configure`  | Auto-configure golangci-lint (default)         |
| `migrate`    | Migrate config to v2.8+ schema                |
| `generate`    | Generate optimized configuration template         |
| `validate`    | Validate existing configuration                  |
| `check`       | Check for missing linters/formatters            |
| `report`      | Generate HTML report of current configuration    |

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Layer (Fang)                   │
├─────────────────────────────────────────────────────────────┤
│  Commands  │  Handlers  │  Validators  │  Reporters  │
├─────────────────────────────────────────────────────────────┤
│              Universal Workflow Integration                  │
├─────────────────────────────────────────────────────────────┤
│  WorkflowEngine  │  EventPublisher  │  ActivityHandlers  │
├─────────────────────────────────────────────────────────────┤
│                 Configuration Logic                       │
├─────────────────────────────────────────────────────────────┤
│  LinterAnalyzer  │  ConfigGenerator  │  Migrator  │
├─────────────────────────────────────────────────────────────┤
│                   Utilities                             │
└─────────────────────────────────────────────────────────────┘
```

## 🧪 Testing

```bash
# Run all tests
ginkgo -r --cover

# Run with verbose output
ginkgo -r -v

# Run specific suite
ginkgo ./pkg/config -r

# Generate coverage
ginkgo -r --coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📊 Linter Prioritization

### Critical (Always Enabled)
Security and correctness linters that should never be disabled:
- `errcheck` - Unchecked error detection
- `gosec` - Security vulnerability scanning
- `staticcheck` - Advanced static analysis
- `govet` - Go vet suspicious constructs
- `sloglint` - Consistent logging with slog
- `loggercheck` - Structured logging best practices

### High Value (Recommended)
Quality and maintainability linters:
- `wrapcheck` - Error wrapping
- `errorlint` - Error handling patterns
- `prealloc` - Slice preallocation optimization
- `exhaustive` - Enum exhaustiveness checks
- `revive` - Fast, configurable linter

### Medium Value (Optional)
Style and consistency linters:
- `misspell` - Typos detection
- `gocyclo` - Cyclomatic complexity
- `dupl` - Code duplication detection
- `lll` - Long line detection

### Optional (Niche)
Project-specific or opinionated linters

## 🔧 Tech Stack

- **Workflow**: [universal-workflow](https://github.com/LarsArtmann/universal-workflow)
- **CLI**: [charmbracelet/fang](https://github.com/charmbracelet/fang)
- **Config**: [spf13/viper](https://github.com/spf13/viper)
- **Testing**: [onsi/ginkgo](https://onsi.github.io/ginkgo/)
- **HTML**: [a-h/templ](https://github.com/a-h/templ)
- **Logging**: [charmbracelet/log](https://github.com/charmbracelet/log) + stdlib slog
- **YAML**: [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Ensure all tests pass (`ginkgo -r`)
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details

## 🔗 Related Projects

- [buildflow](https://github.com/larsartmann/buildflow) - Comprehensive build tooling
- [universal-workflow](https://github.com/LarsArtmann/universal-workflow) - Workflow orchestration
- [golangci-lint](https://github.com/golangci/golangci-lint) - Go linters aggregator
