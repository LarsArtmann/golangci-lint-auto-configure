# ADR-003: Interface-Based Design for Testability

**Status:** Accepted  
**Date:** 2026-04-09  
**Author:** Lars Artmann (@larsartmann)

---

## Context

Our codebase interacts with external dependencies (filesystem, git, golangci-lint binary) and internal components (config loading, analysis, fixing). Testing these components in isolation is difficult without proper abstractions.

Prior to this decision, we had:

1. **Direct dependencies**: Functions created their own dependencies inline
2. **Concrete implementations**: No way to inject mocks for testing
3. **Integration tests only**: Couldn't unit test components in isolation
4. **Brittle tests**: Tests broke when dependencies changed

---

## Decision

We will define small, focused interfaces for all major components and depend on interfaces, not concrete implementations.

### Implementation

```go
// pkg/types/types.go

// ConfigReader defines the interface for reading golangci-lint configurations.
type ConfigReader interface {
    LoadConfig(path string) (*Config, error)
    FindConfigFile(startDir string) (string, error)
}

// ConfigWriter defines the interface for writing golangci-lint configurations.
type ConfigWriter interface {
    SaveConfig(config *Config, path string) error
}

// ConfigValidator defines the interface for validating configurations.
type ConfigValidator interface {
    ValidateConfig(config *Config) []error
}

// ConfigLoader combines all config operations for backward compatibility.
type ConfigLoader interface {
    ConfigReader
    ConfigWriter
    ConfigValidator
    ConfigDiscovery
    ConfigInspector
    ConfigCreator
    GitChecker
}

// LinterAnalyzer defines the interface for analyzing golangci-lint configurations.
type LinterAnalyzer interface {
    AnalyzeConfig(ctx context.Context, configPath string) (*ConfigAnalysis, error)
    FindBinary(ctx context.Context) error
    CheckVersion(ctx context.Context) error
    FormatRecommendations(analysis *ConfigAnalysis) string
    GetSummary(analysis *ConfigAnalysis) string
}

// LinterFixer defines the interface for fixing golangci-lint configurations.
type LinterFixer interface {
    FixConfig(ctx context.Context, configPath string, priority LinterPriority, dryRun bool) (*MigrationResult, error)
}
```

### Key Design Decisions

1. **Small, Focused Interfaces**: Each interface does one thing well (SRP)
2. **Composable**: Larger interfaces compose smaller ones (ConfigLoader example)
3. **Minimal Surface Area**: Only methods actually needed by consumers
4. **Interface Segregation**: No fat interfaces, consumers depend on what they use

---

## Consequences

### Positive

- **Testability**: Easy to create mocks for unit testing
- **Flexibility**: Swap implementations without changing consumers
- **Documentation**: Interfaces document expected behavior
- **Decoupling**: Consumers don't depend on implementation details

### Negative

- **Indirection**: More layers of abstraction
- **Boilerplate**: Need adapters/concrete implementations

### Trade-offs Accepted

- Slight complexity increase for significant testability gains

---

## Usage Examples

### Before (direct dependency)

```go
func ProcessConfig(path string) error {
    loader := config.NewLoader(logger)  // Hardcoded dependency
    cfg, err := loader.LoadConfig(path)
    // ...
}

// Testing: Cannot mock, must use real filesystem
```

### After (interface dependency)

```go
func ProcessConfig(loader ConfigLoader, path string) error {
    cfg, err := loader.LoadConfig(path)  // Depends on interface
    // ...
}

// Testing: Easy to mock
func TestProcessConfig(t *testing.T) {
    mockLoader := &mockConfigLoader{
        LoadConfigFunc: func(path string) (*Config, error) {
            return &Config{Version: "2"}, nil
        },
    }
    err := ProcessConfig(mockLoader, ".golangci.yml")
    // Assert behavior...
}
```

---

## Interface Segregation Example

### Before: Fat Interface

```go
type ConfigLoader interface {
    LoadConfig(path string) (*Config, error)
    SaveConfig(config *Config, path string) error
    FindConfigFile(startDir string) (string, error)
    ValidateConfig(config *Config) []error
    GetLintersEnabled(config *Config) []string
    GetLintersDisabled(config *Config) []string
    CreateDefaultConfig(ctx context.Context) *Config
    GetAllLinterNames(ctx context.Context) ([]string, error)
    EnsureGitRepo(ctx context.Context, startDir string) error
}

// Consumer only needs LoadConfig but depends on everything
```

### After: Composed Interfaces

```go
// Consumer only needs ConfigReader
type ConfigAnalyzer struct {
    reader ConfigReader  // Only what I need
}

// Consumer needs both read and write
type ConfigManager struct {
    reader ConfigReader
    writer ConfigWriter
}

// Full loader still available for convenience
type ConfigLoader interface {
    ConfigReader
    ConfigWriter
    // ... all others
}
```

---

## Migration Strategy

1. **Phase 1**: Define interfaces for new components
2. **Phase 2**: Refactor existing components to implement interfaces
3. **Phase 3**: Update consumers to depend on interfaces
4. **Phase 4**: Create mock implementations for testing

---

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| No interfaces, concrete types | Simple, direct | Un-testable, tightly coupled | Rejected |
| Mocking frameworks | Powerful mocks | External dependency, complexity | Rejected |
| **Explicit interfaces** | Clear contracts, no deps | More code to maintain | **Accepted** |

---

## Testing with Interfaces

### Mock Implementation Example

```go
// test/mock_config_loader.go
type MockConfigLoader struct {
    LoadConfigFunc  func(path string) (*Config, error)
    SaveConfigFunc  func(config *Config, path string) error
    // ... other methods
}

func (m *MockConfigLoader) LoadConfig(path string) (*Config, error) {
    if m.LoadConfigFunc != nil {
        return m.LoadConfigFunc(path)
    }
    return nil, errors.New("not implemented")
}

// ... implement other methods
```

### Table-Driven Test Example

```go
func TestAnalyze(t *testing.T) {
    tests := []struct {
        name    string
        loader  *MockConfigLoader
        wantErr bool
    }{
        {
            name: "valid config",
            loader: &MockConfigLoader{
                LoadConfigFunc: func(string) (*Config, error) {
                    return &Config{Version: "2"}, nil
                },
            },
        },
        {
            name: "config not found",
            loader: &MockConfigLoader{
                LoadConfigFunc: func(string) (*Config, error) {
                    return nil, os.ErrNotExist
                },
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            analyzer := NewAnalyzer(tt.loader)
            _, err := analyzer.AnalyzeConfig(context.Background(), ".golangci.yml")
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

---

## Related Decisions

- ADR-001: Set[T] Type - Uses interfaces for set operations
- ADR-002: CommandBuilder Pattern - Depends on interfaces for dependency injection

---

## References

- [Interface Definitions](../../pkg/types/types.go) (lines 237-288)
- [LinterAnalyzer Interface](../../pkg/types/types.go) (lines 371-379)
- [LinterFixer Interface](../../pkg/types/types.go) (lines 381-383)
- [Go Interface Best Practices](https://github.com/golang/go/wiki/CodeReviewComments#interfaces)

---

_Accepted by: Lars Artmann_  
_Date: 2026-04-09_
