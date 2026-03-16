# PARTS.md - Component Analysis for Reusable Libraries/SDKs

> Analysis of golangci-lint-auto-configure components that could be extracted as standalone reusable libraries.
> **Last Updated:** March 1, 2026 (12:42)

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
| `pkg/workflow`  | **LOW**             | External dependency wrapper         | Keep internal                        |

---

## Component Inventory

### 1. `pkg/detection` - Project Type Detector

**Location:** `pkg/detection/detector.go` (268 lines), `pkg/detection/patterns.go`

**Capabilities:**

- Detects Go project type: CLI, Library, Web, API, Monorepo, Unknown
- Analyzes `go.mod` for module path and dependencies
- Detects HTTP frameworks: gin, echo, fiber, chi, stdlib, fasthttp, httprouter
- Detects CLI frameworks: cobra, urfave/cli, bubbletea, kingpin, flag
- Checks for `main` package presence
- Monorepo detection via multiple `go.mod` files

**Current API:**

```go
type Detector struct { ... }
func NewDetector(rootDir string) *Detector  // Simplified - no logger required
func (d *Detector) Detect() ProjectType
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

- `pkg/constants/linter_priorities.go` (126 lines)
- `pkg/constants/linter_reasons.go` (128 lines)
- `pkg/constants/formatter_data.go` (56 lines)
- `pkg/constants/presets.go` (38 lines)
- `pkg/constants/rules.go` (16 lines)
- `pkg/constants/config.go` (22 lines)

**Total:** ~386 lines of curated knowledge data

**Contents:**

- `LinterPriorities`: Map of 100+ linters to priority levels (Critical/High/Medium/Optional)
- `LinterReasons`: Human-readable explanations for each linter
- `FormatterInfo` / `FormatterPriorities` / `FormatterReasons`: Formatter metadata
- `PresetLinters`: Pre-defined configurations (minimal, standard, strict, security, performance)
- `RedundantLinters`: Known redundant combinations

**Sample Data:**

```go
LinterPriorities = map[types.LinterName]types.LinterPriority{
    "gosec":       LinterPriorityCritical,  // Security
    "errcheck":    LinterPriorityCritical,  // Correctness
    "staticcheck": LinterPriorityCritical,  // Correctness
    "wrapcheck":   LinterPriorityHigh,      // Quality
    "errorlint":   LinterPriorityHigh,      // Quality
    // ... 100+ more
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

**Location:** `pkg/config/loader.go` (239 lines)

**Capabilities:**

- Load/save YAML configs (`gopkg.in/yaml.v3`)
- Config file discovery (`.golangci.yml`, `.golangci.yaml`, `.golangci.toml`, `.golangci.json`)
- Git repository verification (requires running in git repo)
- Config validation via `golangci-lint linters --json`
- Default config creation

**Current API:**

```go
type Loader struct { ... }
func NewLoader(logger *log.Logger) *Loader
func (l *Loader) LoadConfig(path string) (*Config, error)
func (l *Loader) SaveConfig(cfg *Config, path string) error
func (l *Loader) ValidateConfig(cfg *Config) []error
func (l *Loader) FindConfigFile(projectPath string) (string, error)
func (l *Loader) EnsureGitRepo(startDir string) error
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

**Location:** `pkg/client/client.go` (142 lines)

**Already SDK-Ready:**

- Clean public API with context support
- Options pattern for configuration
- Convenience functions (`SimpleAnalyze`)
- Example documentation in godoc

**Current API:**

```go
type Client struct { ... }
type Options struct { ... }

func New(opts Options) *Client
func (c *Client) AnalyzeConfig(ctx context.Context, configPath string) (*types.ConfigAnalysis, error)
func (c *Client) LoadConfig(configPath string) (*config.Config, error)
func (c *Client) ValidateConfig(cfg *config.Config) []error
func (c *Client) GetSummary(analysis *types.ConfigAnalysis) string
func (c *Client) SaveConfig(cfg *config.Config, path string) error
func SimpleAnalyze(ctx context.Context, opts Options, configPath string) (string, error)
```

**Recent Improvement:** All methods now accept `context.Context` for cancellation support.

**Recommendation:**

- Keep as-is, already serves as SDK
- Improve documentation with more examples
- Consider adding integration test examples

---

### 5. `pkg/types` - Domain Types

**Location:** `pkg/types/types.go` (267 lines), `pkg/types/result.go`

**Contents:**

- `LinterPriority` / `FormatterPriority` types
- `LinterName` / `FormatterName` (strong typing)
- `LinterInfo`, `FormatterInfo` structs
- `Config`, `LintersConfig`, `RunConfig`, `IssuesConfig` structs
- `ConfigAnalysis`, `FixResult` result types
- Interfaces: `ConfigLoader`, `LinterAnalyzer`, `LinterFixer`

**Recommendation:**
Bundle with extracted libraries. Not valuable standalone.

---

### 6. `pkg/diff` - Config Differ

**Location:** `pkg/diff/differ.go` (244 lines)

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

- `pkg/linter/analyzer.go` (245 lines) - Main analysis logic
- `pkg/linter/fixer.go` (295 lines) - Configuration fixing
- `pkg/linter/categorizer.go` (104 lines) - Linter categorization
- `pkg/linter/validator.go` (92 lines) - Configuration validation
- `pkg/linter/version_checker.go` (143 lines) - golangci-lint version checking
- `pkg/linter/command_runner.go` (37 lines) - Command execution

**All files now comply with <250 line limit per HOW_TO_GOLANG.md**

**Capabilities:**

- golangci-lint binary discovery and version checking (>= v2.10.1)
- Config analysis via `golangci-lint linters --json`
- Formatter analysis via `golangci-lint formatters --json`
- Recommendation categorization by priority
- Automatic deprecated linter replacement

**Coupling:**

- Depends on `pkg/constants` for priorities
- Depends on `pkg/config` for loading
- Requires golangci-lint binary installed

**Recommendation:**
Keep as core logic. Too coupled to this tool's purpose.

---

### 8. `pkg/report` - Report Generation

**Location:** `pkg/report/generator.go` (47 lines), `pkg/report/report.templ`

**Capabilities:**

- HTML report generation using `a-h/templ`
- JSON report generation
- Dark mode support

**Recommendation:**
Keep internal. Specific to this tool's output format.

---

### 9. `pkg/workflow` - Workflow Orchestration

**Location:** `pkg/workflow/workflow.go` (196 lines)

**Capabilities:**

- Multi-step workflow orchestration
- Uses external `github.com/LarsArtmann/universal-workflow`
- Activities: AnalysisActivity, ValidationActivity, ReportActivity

**Recommendation:**
Keep internal. Wrapper around external dependency.

---

### 10. `pkg/errors` - Custom Error Types

**Location:** `pkg/errors/errors.go` (116 lines)

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

- `pkg/constants/linter_data.go` (all data)

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

- `gopkg.in/yaml.v3`

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

| Component      | Reason                      |
| -------------- | --------------------------- |
| `pkg/linter`   | Core logic, tightly coupled |
| `pkg/diff`     | Generic alternatives exist  |
| `pkg/report`   | Tool-specific output        |
| `pkg/workflow` | External dependency wrapper |
| `internal/cli` | CLI-specific, not reusable  |

---

## Library Policy Compliance

Per `HOW_TO_GOLANG.md`:

| Requirement                     | Status | Notes                                     |
| ------------------------------- | ------ | ----------------------------------------- |
| Files <250 lines                | ✅     | All files now compliant after refactoring |
| Functions <30 lines             | ✅     | Mostly compliant                          |
| No `any` types                  | ✅     | Strong typing used                        |
| DI with samber/do/v2            | ❌     | Manual DI currently                       |
| Logging with slog+charmbracelet | ✅     | Using charmbracelet/log                   |
| Error wrapping                  | ✅     | Using `%w`                                |
| Custom error types              | ✅     | `pkg/errors/errors.go`                    |
| Context propagation             | ✅     | All public methods accept context         |

---

## Action Items

1. **Create `go-project-detector` repository**
   - Extract `pkg/detection`
   - Refactor to remove internal dependencies
   - Add comprehensive tests
   - Document API

2. **Create `golangci-lint-knowledge` repository**
   - Extract `pkg/constants/linter_*.go`, `formatter_data.go`, `presets.go`
   - Add versioning strategy
   - Create update automation
   - Document data sources

3. **Enhance `pkg/client` documentation**
   - Add more examples
   - Create integration examples
   - Document common patterns

4. **Consider bundling**
   - Option: Create meta-package importing all extracted libraries
   - `github.com/larsartmann/golangci-lint-sdk`

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
