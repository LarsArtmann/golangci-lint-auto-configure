# PARTS.md - Component Analysis for Reusable Libraries/SDKs

> Analysis of golangci-lint-auto-configure components that could be extracted as standalone reusable libraries.
> **Last Updated:** April 30, 2026

## Executive Summary

| Component       | Extraction Priority | Unique Value                        | Recommendation                       |
| --------------- | ------------------- | ----------------------------------- | ------------------------------------ |
| `pkg/detection` | **HIGH**            | No equivalent exists                | Extract as `go-project-detector`     |
| `pkg/constants` | **HIGH**            | Curated security-focused priorities | Extract as `golangci-lint-knowledge` |
| `pkg/client`    | **MEDIUM**          | Already SDK-ready                   | Keep, document as public API         |
| `pkg/config`    | **MEDIUM**          | Git-based version control           | Extract with detection               |
| `pkg/types`     | **LOW**             | Domain types                        | Bundle with extracted libs           |
| `pkg/diff`      | **LOW**             | Generic alternative exists          | Keep internal                        |
| `pkg/report`    | **LOW**             | Specific to this tool               | Keep internal                        |
| `pkg/finding`   | **LOW**             | go-finding integration layer        | Keep internal                        |
| `pkg/migration` | **LOW**             | v1→v2 migration logic               | Keep internal                        |
| `pkg/ui`        | **LOW**             | Terminal output formatting          | Keep internal                        |

---

## Component Inventory

### 1. `pkg/detection` - Project Type Detector

**Location:** `pkg/detection/detector.go` (377 lines), `pkg/detection/patterns.go` (92 lines)

**Capabilities:**

- Detects Go project type: CLI, Library, Web, API, Monorepo, Unknown
- Analyzes `go.mod` for module path and dependencies
- Detects HTTP frameworks: gin, echo, fiber, chi, stdlib, fasthttp, httprouter
- Detects CLI frameworks: cobra, urfave/cli, bubbletea, kingpin, flag
- Detects API patterns: swaggo annotations, API code patterns
- Checks for `main` package presence
- Monorepo detection via multiple `go.mod` files

**Current API:**

```go
type Detector struct { ... }
func NewDetector(rootDir string) *Detector
func (d *Detector) Detect() ProjectType
func (d *Detector) HasSwaggo() (bool, error)
```

**Key Improvement:** API simplified to not require logger injection - pure detection logic.

**Alternatives Research:**

- **None found** - No standalone Go library for project type detection
- IDE plugins do this internally but don't expose as library
- Build tools (goreleaser, etc.) have hardcoded assumptions

**Unique Value Proposition:**

1. **No equivalent exists** - This is the ONLY library that systematically detects Go project types
2. Framework-aware recommendations become possible
3. Could be used by: IDE plugins, scaffolding tools, CI/CD analyzers, documentation generators

**Extraction Recommendation:**

```
github.com/larsartmann/go-project-detector
```

---

### 2. `pkg/constants` - Linter Knowledge Base

**Location:** Split into multiple files for maintainability:

- `pkg/constants/linter_priorities.go` (127 lines)
- `pkg/constants/linter_reasons.go` (127 lines)
- `pkg/constants/formatter_data.go` (60 lines)
- `pkg/constants/presets.go` (41 lines)
- `pkg/constants/rules.go` (55 lines)
- `pkg/constants/config.go` (44 lines)
- `pkg/constants/version.go` (7 lines)
- `pkg/constants/experiments.go` (50 lines)
- `pkg/constants/experiments_test.go` (90 lines)

**Total:** ~601 lines of curated knowledge data

**Contents:**

- `LinterPriorities`: Map of 119 linters to priority levels (Critical/High/Medium/Optional)
- `LinterReasons`: Human-readable explanations for each linter
- `FormatterInfo` / `FormatterPriorities` / `FormatterReasons`: Formatter metadata
- `PresetLinters`: Pre-defined configurations (minimal, standard, strict, security, performance)
- `RedundantLinters`: Known redundant combinations
- `MinGolangCILintVersion`: Minimum required golangci-lint version (`v2.10.1`)
- `Experiments`: Feature flags for experimental linter settings

**Sample Data:**

```go
LinterPriorities = map[types.LinterName]types.LinterPriority{
    "gosec":       LinterPriorityCritical,  // Security
    "errcheck":    LinterPriorityCritical,  // Correctness
    "staticcheck": LinterPriorityCritical,  // Correctness
    "wrapcheck":   LinterPriorityHigh,      // Quality
    "errorlint":   LinterPriorityHigh,      // Quality
    // ... 119 total entries
}
```

**Alternatives:**

- **golangci-lint built-in presets**: Simple enable-all/disable-all patterns
- **golangci-lint documentation**: Static markdown, not machine-readable
- **Community configs**: Scattered across repos, not versioned

**Unique Value Proposition:**

1. **Machine-readable curated knowledge** - Priorities based on security/quality best practices
2. **Version-controlled** - Track changes as linters evolve
3. **Human explanations** - `LinterReasons` provides context
4. **Extensible** - Add custom priorities per project/organization

**Use Cases:**

- Linter recommendation engines
- CI/CD quality gates
- Security auditing tools
- IDE plugins suggesting linters
- Documentation generators

**Extraction Recommendation:**

```
github.com/larsartmann/golangci-lint-knowledge
```

---

### 3. `pkg/config` - Config Management

**Location:** `pkg/config/loader.go` (439 lines)

**Capabilities:**

- Load/save YAML/TOML/JSON configs (`go.yaml.in/yaml.v3`, `pelletier/go-toml/v2`)
- Config file discovery (`.golangci.yml`, `.golangci.yaml`, `.golangci.toml`, `.golangci.json`)
- Git repository verification (requires running in git repo)
- Config validation via `golangci-lint linters --json`
- Default config creation
- Result types for functional error handling (`mo.Result`)

**Current API:**

```go
type Loader struct { ... }
func NewLoader(logger *log.Logger) *Loader
func NewLoaderWithFS(logger *log.Logger, fs afero.Fs) *Loader
func (l *Loader) LoadConfig(path string) (*Config, error)
func (l *Loader) FindConfigFile(startDir string) (string, error)
func (l *Loader) FindAllConfigFiles(startDir string) []string
func (l *Loader) FindOrGetDefaultConfigPath(startDir string) string
func (l *Loader) SaveConfig(config *Config, path string) error
func (l *Loader) ValidateConfig(config *Config) []error
func (l *Loader) EnsureGitRepo(ctx context.Context, startDir string) error
func (l *Loader) IsGitRepo(ctx context.Context, startDir string) bool
func (l *Loader) GetAllLinterNames(ctx context.Context) ([]string, error)
func (l *Loader) CreateDefaultConfig(ctx context.Context) *Config
func (l *Loader) GetLintersEnabled(config *Config) []string
func (l *Loader) GetLintersDisabled(config *Config) []string
```

**Alternatives:**

- **Direct YAML parsing**: No validation, no git verification, no discovery
- **golangci-lint internal**: Not exposed as library
- **go-yaml/yaml**: Low-level, no domain knowledge

**Unique Value Proposition:**

1. **Domain-aware** - Knows golangci-lint config structure
2. **Git-based safety** - Requires git repo, uses git for version control
3. **Validation** - Uses actual golangci-lint binary for validation
4. **Multi-format** - Supports YAML, TOML, JSON

**Extraction Recommendation:**
Bundle with `go-project-detector` or create separate:

```
github.com/larsartmann/golangci-lint-config
```

---

### 4. `pkg/client` - High-Level SDK

**Location:** `pkg/client/client.go` (232 lines)

**Already SDK-Ready:**

- Clean public API with context support
- Options pattern for configuration
- Convenience functions (`SimpleAnalyze`, `SimpleFix`)
- Fix functionality with `FixConfig` and `FixOptions`

**Current API:**

```go
type Client struct { ... }
type Options struct { ... }
type FixOptions struct { ... }

func New(opts Options) *Client
func (c *Client) AnalyzeConfig(ctx context.Context, configPath string) (*types.ConfigAnalysis, error)
func (c *Client) LoadConfig(configPath string) (*config.Config, error)
func (c *Client) ValidateConfig(cfg *config.Config) []error
func (c *Client) GetSummary(analysis *types.ConfigAnalysis) string
func (c *Client) SaveConfig(cfg *config.Config, path string) error
func (c *Client) FixConfig(ctx context.Context, configPath string, opts FixOptions) (*types.MigrationResult, error)
func SimpleFix(ctx context.Context, opts Options, configPath string, dryRun bool) (*types.MigrationResult, error)
func SimpleAnalyze(ctx context.Context, opts Options, configPath string) (string, error)
```

**All methods accept `context.Context` for cancellation support.**

**Recommendation:**

- Keep as-is, already serves as SDK
- Improve documentation with more examples
- Consider adding integration test examples

---

### 5. `pkg/types` - Domain Types

**Location:** `pkg/types/types.go` (385 lines), `pkg/types/result.go` (101 lines)

**Contents:**

- `LinterPriority` / `FormatterPriority` types
- `LinterName` / `FormatterName` (strong typing)
- `LinterInfo`, `FormatterInfo` structs
- `Config`, `LintersConfig`, `RunConfig`, `IssuesConfig` structs
- `ConfigAnalysis`, `FixResult`, `MigrationResult` result types
- Interfaces: `ConfigLoader`, `LinterAnalyzer`, `LinterFixer`
- Result types for functional error handling

**Recommendation:**
Bundle with extracted libraries. Not valuable standalone.

---

### 6. `pkg/diff` - Config Differ

**Location:** `pkg/diff/differ.go` (275 lines)

**Capabilities:**

- Compare two golangci-lint configs
- Change types: Added, Removed, Modified
- Human-readable diff formatting

**Alternatives:**

- **pmezard/go-difflib**: Generic text diff
- **kylelemons/godebug/diff**: Generic diff
- **sergi/go-diff**: Unified diff

**Assessment:**
Our differ provides semantic understanding (knows what changed in config structure), but generic diff libraries are sufficient for most use cases.

**Recommendation:**
Keep internal. Not enough unique value for extraction.

---

### 7. `pkg/linter` - Analyzer & Fixer

**Location:** Split into multiple files for maintainability:

- `pkg/linter/analyzer.go` (282 lines) - Main analysis logic
- `pkg/linter/fixer.go` (268 lines) - Configuration fixing
- `pkg/linter/fixer_preflight.go` (268 lines) - Pre-flight checks before fixing
- `pkg/linter/categorizer.go` (137 lines) - Linter categorization
- `pkg/linter/fixer_formatters.go` (197 lines) - Formatter-specific fixing
- `pkg/linter/fixer_config.go` (124 lines) - Config construction for fixing
- `pkg/linter/command_runner.go` (97 lines) - Command execution
- `pkg/linter/version_checker.go` (136 lines) - golangci-lint version checking
- `pkg/linter/fixer_deprecated.go` (93 lines) - Deprecated linter replacement
- `pkg/linter/fixer_results.go` (73 lines) - Fix result types

**Capabilities:**

- golangci-lint binary discovery and version checking (>= v2.10.1)
- Config analysis via `golangci-lint linters --json`
- Formatter analysis via `golangci-lint formatters --json`
- Recommendation categorization by priority
- Automatic deprecated linter replacement
- Pre-flight validation before config changes
- Formatter-aware config fixing

**Coupling:**

- Depends on `pkg/constants` for priorities
- Depends on `pkg/config` for loading
- Requires golangci-lint binary installed

**Recommendation:**
Keep as core logic. Too coupled to this tool's purpose.

---

### 8. `pkg/report` - Report Generation

**Location:** `pkg/report/generator.go` (48 lines), `pkg/report/json_report_generator.go` (86 lines), `pkg/report/report_templ.go` (453 lines)

**Capabilities:**

- HTML report generation using `a-h/templ`
- JSON report generation
- Dark mode support

**Recommendation:**
Keep internal. Specific to this tool's output format.

---

### 9. `pkg/finding` - go-finding Integration

**Location:**

- `pkg/finding/converter.go` (222 lines) - Convert domain types to `finding.Finding`
- `pkg/finding/golangci_lint.go` (107 lines) - Parse golangci-lint JSON output to Findings
- `pkg/finding/detector.go` (62 lines) - `ConfigAnalysisDetector` for pipeline integration
- `pkg/finding/diff_converter.go` (94 lines) - Convert `diff.Change` to Finding
- `pkg/finding/helpers.go` (50 lines) - LSP, filter, merge, groupBy helpers

**Capabilities:**

- Convert `LinterRecommendation` and `ValidationError` to `finding.Finding`
- Parse `golangci-lint run --out-format=json` output into Findings
- SARIF 2.1.0 output generation (via go-finding)
- Priority-to-severity mapping (Critical→critical, High→error, Medium→warning, Optional→info)
- Pipeline integration via `ConfigAnalysisDetector`

**Recommendation:**
Keep internal. Integration layer between this tool and go-finding.

---

### 10. `pkg/migration` - v1 to v2 Config Migration

**Location:** `pkg/migration/migrator.go`, `pkg/migration/migrations.go`, `pkg/migration/migrations_linters_settings.go`, `pkg/migration/config_types.go`, `pkg/migration/rules.go`, `pkg/migration/validator.go`, `pkg/migration/yaml_loader.go`, `pkg/migration/testdata/`

**Capabilities:**

- Migrate golangci-lint v1 configs to v2 format
- Linter-specific setting migrations
- Config validation during migration
- YAML loading and saving

**Recommendation:**
Keep internal. Specific to v1→v2 migration which is a one-time operation.

---

### 11. `pkg/ui` - Terminal Output Formatting

**Location:** `pkg/ui/formatter.go` (126 lines), `pkg/ui/styled_output.go` (104 lines), `pkg/ui/finding_formatter.go` (86 lines)

**Capabilities:**

- Terminal text formatting for analysis results
- Styled output with color and layout (using lipgloss v2)
- Finding formatter for go-finding objects

**Recommendation:**
Keep internal. UI layer specific to this CLI tool.

---

### 12. `pkg/errors` - Custom Error Types

**Location:** `pkg/errors/errors.go` (167 lines)

**Package name:** `apperrors`

**Types:**

- `ConfigError` - Configuration-related errors
- `AnalysisError` - Analysis-related errors
- `ReportError` - Report generation errors
- Helper functions: `IsConfigError`, `IsAnalysisError`, `IsReportError`

**Recommendation:**
Bundle with extracted libraries. Pattern worth reusing but not standalone library.

---

## Extraction Roadmap

### Phase 1: High-Value Extraction

#### 1.1 `go-project-detector`

```
github.com/larsartmann/go-project-detector
```

**Scope:**

- `pkg/detection` (refactored)
- `pkg/types` (relevant parts only)

**API:**

```go
package detector

type ProjectType string

const (
    ProjectTypeCLI      ProjectType = "cli"
    ProjectTypeLibrary  ProjectType = "library"
    ProjectTypeWeb      ProjectType = "web"
    ProjectTypeAPI      ProjectType = "api"
    ProjectTypeMonorepo ProjectType = "monorepo"
    ProjectTypeUnknown  ProjectType = "unknown"
)

type Result struct {
    Type      ProjectType
    Framework string
    Module    string
    HasMain   bool
}

type Detector interface {
    Detect(path string) (Result, error)
}
```

**Dependencies:**

- None (pure Go)

---

#### 1.2 `golangci-lint-knowledge`

```
github.com/larsartmann/golangci-lint-knowledge
```

**Scope:**

- `pkg/constants/linter_priorities.go`, `linter_reasons.go`
- `pkg/constants/formatter_data.go`, `presets.go`, `rules.go`, `config.go`, `version.go`

**API:**

```go
package knowledge

type Priority int

const (
    PriorityCritical Priority = iota
    PriorityHigh
    PriorityMedium
    PriorityOptional
)

type LinterInfo struct {
    Name        string
    Priority    Priority
    Reason      string
    Deprecated  bool
    Replacement string
}

type Preset struct {
    Name        string
    Description string
    Linters     []string
}

var AllLinters map[string]LinterInfo
var Presets map[string]Preset

func GetByPriority(p Priority) []LinterInfo
func GetSecurityLinters() []LinterInfo
func GetQualityLinters() []LinterInfo
```

**Dependencies:**

- None (pure data)

---

### Phase 2: Medium-Value Extraction

#### 2.1 `golangci-lint-config`

```
github.com/larsartmann/golangci-lint-config
```

**Scope:**

- `pkg/config/loader.go` (refactored)
- `pkg/types` (config-related parts)

**Dependencies:**

- `go.yaml.in/yaml/v3`
- `pelletier/go-toml/v2`
- `spf13/afero`
- `samber/mo`

---

### Phase 3: SDK Enhancement

#### 3.1 Enhance `pkg/client`

- Add more convenience functions
- Create example programs
- Improve godoc documentation
- Add integration test suite as examples

---

## Non-Extraction Candidates

These components should remain internal to golangci-lint-auto-configure:

| Component       | Reason                           |
| --------------- | -------------------------------- |
| `pkg/linter`    | Core logic, tightly coupled      |
| `pkg/diff`      | Generic alternatives exist       |
| `pkg/report`    | Tool-specific output             |
| `pkg/finding`   | Integration layer for go-finding |
| `pkg/migration` | One-time v1→v2 migration         |
| `pkg/ui`        | CLI terminal output              |
| `internal/cli`  | CLI-specific, not reusable       |

---

## Conclusion

**Primary extraction candidates:**

1. `pkg/detection` → `go-project-detector` (no equivalent exists)
2. `pkg/constants` → `golangci-lint-knowledge` (curated, machine-readable data)

**Secondary candidates:** 3. `pkg/config` → `golangci-lint-config` (domain-aware config handling)

**Keep as-is:** 4. `pkg/client` (already SDK-ready, needs better docs)

The unique value of this project lies in:

- **Project type detection** - No equivalent exists
- **Curated linter knowledge** - Machine-readable, security-focused priorities
- **Domain-aware abstractions** - Understands golangci-lint ecosystem
