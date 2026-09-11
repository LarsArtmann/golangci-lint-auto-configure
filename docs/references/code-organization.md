## Code Organization

### Directory Structure

```text
golangci-lint-auto-configure/
├── cmd/
│   ├── golangci-lint-auto-configure/
│   │   └── main.go                    # Entry point, sets version via ldflags
│   └── coverage-check/
│       └── main.go                    # Coverage threshold gate (used by CI)
├── pkg/
│   ├── types/                        # Core type definitions and interfaces
│   │   ├── types.go                  # Main types: LinterPriority, Config, LinterInfo, etc.
│   │   ├── config_types.go           # Config structs for .golangci.yml schema
│   │   ├── validation.go             # ValidationResult, HealthIssue, HealthSeverity
│   │   ├── clone.go                  # Deep-copy for Config (prevents aliasing)
│   │   └── set.go                    # Generic Set[T] with full algebra
│   ├── constants/
│   │   ├── linter_priorities.go  # Linter priorities (110 entries)
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
│   │   ├── errors.go              # Custom error types (package: apperrors)
│   │   ├── classification.go     # go-error-family Family registration + Classified interface
│   │   └── doc.go                # Package documentation
│   ├── audit/                      # Append-only JSONL audit ledger for config mutations
│   │   └── ledger.go              # Ledger + NoopRecorder (90-day retention)
│   ├── policy/                     # Disable-reason sidecar enforcement (.golangci-lint-auto-configure.yml)
│   │   └── policy.go              # Policy loader + anti-gaming enforcement
│   ├── client/                     # golangci-lint CLI client wrapper
│   │   └── client.go              # Runs `golangci-lint linters` / version commands
│   ├── utils/                      # Shared utilities
│   │   ├── git.go                 # Git repo detection (version-control safety)
│   │   └── retry.go               # Exponential backoff with context cancellation
│   ├── version/
│   │   └── version.go             # Structured version info with runtime/debug fallback
├── internal/
│   ├── cli/
│   │   ├── commands.go             # Root command + subcommand wiring
│   │   ├── cmd_configure.go        # Configure command
│   │   ├── cmd_analyze.go          # Analyze command
│   │   ├── cmd_validate.go         # Validate command
│   │   ├── cmd_report.go           # Report command
│   │   ├── cmd_audit.go            # Audit query command (--json, --since, --clear)
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
└── flake.nix                      # Build/test/lint via Nix (PRIMARY INTERFACE)
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
  - `linter_priorities.go`: Priority levels (110 linters)
  - `linter_reasons.go`: Human-readable reasons
  - `linter_settings.go`: Typed default settings structs with `SettingsConverter` interface
  - `formatter_data.go`: Formatter priorities and reasons
  - `presets.go`: Pre-defined configurations
  - `rules.go`: Deprecated/redundant linter rules (`DeprecatedLinters`)
  - `version.go`: Minimum golangci-lint version
  - `experiments.go`: Feature flags

**5. go-finding Integration**

- `pkg/finding/` provides unified data model for static analysis results
- SARIF 2.1.0 output for CI/CD integration
- Priority-to-severity mapping (Critical→critical, High→error, etc.)

**6. Typed Settings Pattern (SettingsConverter)**

- `pkg/constants/linter_settings.go` defines typed Go structs for each linter/formatter that has default settings
- Each struct implements `SettingsConverter` interface with `ToMap() map[string]any`
- `ToMap()` converts via YAML marshal/unmarshal round-trip (handles kebab-case tag conversion)
- `DefaultLinterSettings` and `DefaultFormatterSettings` are `map[Name]SettingsConverter`
- `fixer_config.go` calls `.ToMap()` at injection point to populate runtime config
- Compile-time checks (`var _ SettingsConverter = DepguardSettings{}`) ensure compliance
- Runtime config types (`LintersConfig.Settings`) remain `map[string]any` for round-trip safety
