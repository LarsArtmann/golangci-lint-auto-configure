# Integration Plan: go-composable-business-types/id Library

**Date:** 2026-03-17  
**Status:** Analysis Complete - Ready for Implementation  
**Priority:** Medium (Code Quality Enhancement)

---

## Executive Summary

The `go-composable-business-types/id` package provides **branded, strongly-typed identifiers** that prevent mixing different entity IDs at compile time. This analysis identifies specific opportunities to use this library in the golangci-lint-auto-configure project to enhance type safety and prevent accidental ID misuse.

### Key Findings

| Aspect                 | Assessment                                                                                    |
| ---------------------- | --------------------------------------------------------------------------------------------- |
| **Current ID Types**   | 3 string-based type aliases (`LinterName`, `FormatterName`) with minimal compile-time safety  |
| **Workflow IDs**       | Already using external `universal-workflow` types (`WorkflowID`, `ActivityID`)                |
| **Opportunity**        | High - branded IDs would prevent mixing linter/formatter names and provide better type safety |
| **Integration Effort** | Low - minimal breaking changes, mostly internal type updates                                  |
| **Benefit**            | Medium-High - prevents entire class of bugs at compile time                                   |

---

## 1. Current State Analysis

### 1.1 Existing String-Based Types (pkg/types/types.go:79-96)

```go
// LinterName is a strongly-typed linter name to prevent typos.
type LinterName string

func (ln LinterName) String() string {
    return string(ln)
}

// FormatterName is a strongly-typed formatter name to prevent typos.
type FormatterName string

func (fn FormatterName) String() string {
    return string(fn)
}
```

**Problems with Current Approach:**

1. **No Branding**: `LinterName("gosec")` and `FormatterName("gofmt")` are interchangeable - both are just `string` underneath
2. **No Compile-Time Safety**: Can accidentally pass a `FormatterName` where `LinterName` is expected:
   ```go
   func ProcessLinter(name LinterName) { ... }
   formatter := FormatterName("gofmt")
   ProcessLinter(formatter) // Compiles! Runtime bug.
   ```
3. **Limited Operations**: Only has `String()` method - no equality, comparison, or serialization helpers

### 1.2 Workflow ID Usage (pkg/workflow/workflow.go:139-174)

The project already uses external workflow types:

```go
workflowID := types.WorkflowID("golangci-lint-auto-configure")
wf.Step(types.ActivityID("analyze-config"), ...)
```

These come from `universal-workflow` and may already use branded ID patterns.

### 1.3 Configuration Path Handling

Configuration paths are passed as raw strings throughout:

```go
// pkg/workflow/workflow.go
ConfigPath string

// pkg/linter/analyzer.go
func (a *Analyzer) AnalyzeConfig(ctx context.Context, configPath string) (*types.ConfigAnalysis, error)
```

---

## 2. The go-composable-business-types/id Library

### 2.1 Core Concept

The library uses **phantom types** (brand types) to create distinct identifier types that cannot be accidentally mixed:

```go
// Definition
type ID[B any, V comparable] struct{ value V }

// Usage
type LinterBrand struct{}
type LinterName = id.ID[LinterBrand, string]

linterID := id.NewID[LinterBrand]("gosec")
```

### 2.2 Key Features

| Feature                 | Description                           | Benefit                                     |
| ----------------------- | ------------------------------------- | ------------------------------------------- |
| **Type Safety**         | Different brands = incompatible types | Compile-time prevention of ID mixing        |
| **Serialization**       | JSON, SQL, Binary, Text, Gob support  | Seamless integration with storage/transport |
| **Zero Value Handling** | `IsZero()`, `Or()` methods            | Proper handling of unset IDs                |
| **Comparison**          | `Equal()`, `Compare()` methods        | Type-safe equality and ordering             |
| **Formatting**          | `String()`, `Format()` methods        | Clean string representation                 |

### 2.3 Supported Value Types

The `ID[B, V]` type supports:

- `string` (most common for identifiers)
- All integer types: `int`, `int8`, `int16`, `int32`, `int64`
- All unsigned types: `uint`, `uint8`, `uint16`, `uint32`, `uint64`

---

## 3. Integration Opportunities

### 3.1 Primary Targets for Branded IDs

| Current Type           | Proposed Type                                   | Usage Locations                                  | Impact                                     |
| ---------------------- | ----------------------------------------------- | ------------------------------------------------ | ------------------------------------------ |
| `LinterName string`    | `LinterName = id.ID[LinterBrand, string]`       | 15+ files, type definitions, function signatures | High - prevents linter/formatter confusion |
| `FormatterName string` | `FormatterName = id.ID[FormatterBrand, string]` | 8+ files, similar to above                       | High - prevents formatter/linter confusion |
| `ConfigPath string`    | `ConfigPath = id.ID[ConfigBrand, string]`       | workflow, loader, analyzer packages              | Medium - prevents path/string confusion    |

### 3.2 Files Requiring Updates

**Core Types:**

- `pkg/types/types.go` - Update `LinterName`, `FormatterName` definitions
- `pkg/types/result.go` - May need ID type adjustments

**Usage Locations (LinterName):**

- `pkg/constants/linter_data.go` - All linter name constants
- `pkg/linter/analyzer.go` - Function parameters and returns
- `pkg/linter/fixer.go` - Linter name handling
- `pkg/config/loader.go` - Linter list operations

**Usage Locations (FormatterName):**

- `pkg/constants/linter_data.go` - Formatter constants
- `pkg/config/loader.go` - Formatter list operations

**Configuration Paths:**

- `pkg/workflow/workflow.go` - `ConfigPath` field
- `pkg/linter/analyzer.go` - `AnalyzeConfig` parameter
- `pkg/config/loader.go` - Multiple functions

---

## 4. Implementation Strategy

### 4.1 Phase 1: Foundation (Minimal Risk)

Create a new `pkg/types/ids.go` file with branded ID definitions:

```go
package types

import "github.com/larsartmann/go-composable-business-types/id"

// Brand types (unexported to prevent external instantiation)
type linterBrand struct{}
type formatterBrand struct{}
type configPathBrand struct{}

// LinterName is a branded, strongly-typed linter identifier.
// Prevents accidentally using formatter names or other strings where linter names are expected.
type LinterName = id.ID[linterBrand, string]

// NewLinterName creates a new LinterName from a string.
func NewLinterName(name string) LinterName {
    return id.NewID[linterBrand](name)
}

// FormatterName is a branded, strongly-typed formatter identifier.
type FormatterName = id.ID[formatterBrand, string]

// NewFormatterName creates a new FormatterName from a string.
func NewFormatterName(name string) FormatterName {
    return id.NewID[formatterBrand](name)
}

// ConfigPath is a branded, strongly-typed configuration file path.
type ConfigPath = id.ID[configPathBrand, string]

// NewConfigPath creates a new ConfigPath from a string.
func NewConfigPath(path string) ConfigPath {
    return id.NewID[configPathBrand](path)
}
```

### 4.2 Phase 2: Update Type Definitions (Breaking Change)

Modify `pkg/types/types.go`:

```go
// Remove the old definitions:
// type LinterName string
// type FormatterName string

// Add imports and use the new branded types:
import "github.com/larsartmann/go-composable-business-types/id"

type linterBrand struct{}
type formatterBrand struct{}

type LinterName = id.ID[linterBrand, string]
type FormatterName = id.ID[formatterBrand, string]

// Constructor functions for convenience
func NewLinterName(name string) LinterName {
    return id.NewID[linterBrand](name)
}

func NewFormatterName(name string) FormatterName {
    return id.NewID[formatterBrand](name)
}
```

### 4.3 Phase 3: Update Constants (pkg/constants/linter_data.go)

Current:

```go
LinterPriorities = map[types.LinterName]types.LinterPriority{
    "gosec": types.LinterPriorityCritical,
    // ...
}
```

Updated:

```go
LinterPriorities = map[types.LinterName]types.LinterPriority{
    types.NewLinterName("gosec"): types.LinterPriorityCritical,
    // ...
}
```

Or use a helper:

```go
func l(name string) types.LinterName { return types.NewLinterName(name) }

LinterPriorities = map[types.LinterName]types.LinterPriority{
    l("gosec"): types.LinterPriorityCritical,
    // ...
}
```

### 4.4 Phase 4: Update Function Signatures

**Before:**

```go
func (a *Analyzer) GetLintersByPriority(
    recommendations []LinterRecommendation,
    priority LinterPriority,
) []LinterRecommendation
```

**After:** (No change needed - the type alias means existing code compiles!)

The beauty of type aliases is that `LinterName` still works everywhere, but now provides additional compile-time safety.

---

## 5. Benefits of Integration

### 5.1 Compile-Time Safety Examples

**Current (Unsafe):**

```go
linter := types.LinterName("gosec")
formatter := types.FormatterName("gofmt")

// This compiles but is wrong:
analyzer.ProcessLinter(formatter) // No error!
```

**With Branded IDs:**

```go
linter := types.NewLinterName("gosec")
formatter := types.NewFormatterName("gofmt")

// This is a compile error:
analyzer.ProcessLinter(formatter) // ERROR: cannot use formatter (type FormatterName) as type LinterName
```

### 5.2 Additional Benefits

| Benefit                | Description                                                              |
| ---------------------- | ------------------------------------------------------------------------ |
| **JSON Serialization** | Automatic proper JSON handling - IDs serialize as their underlying value |
| **Zero Value Safety**  | `IsZero()` method for checking unset IDs                                 |
| **Default Values**     | `Or()` method for providing defaults                                     |
| **Type-Safe Equality** | `Equal()` method prevents accidental comparison with wrong types         |
| **Comparison Support** | `Compare()` enables sorting of IDs                                       |

### 5.3 Serialization Examples

```go
// JSON
linter := types.NewLinterName("gosec")
data, _ := json.Marshal(linter)
// Output: "gosec"

// Zero value serializes to null
var empty types.LinterName
data, _ := json.Marshal(empty)
// Output: null

// Unmarshal
var restored types.LinterName
json.Unmarshal([]byte(`"gosec"`), &restored)
```

---

## 6. Migration Considerations

### 6.1 Breaking Changes

| Change              | Impact                                       | Mitigation                               |
| ------------------- | -------------------------------------------- | ---------------------------------------- |
| Type alias change   | Minimal - type alias maintains compatibility | Gradual migration with type constructors |
| JSON serialization  | None - same format as string                 | Verified by existing tests               |
| Function signatures | None - type alias is transparent             | No changes needed                        |

### 6.2 Testing Strategy

1. **Unit Tests**: Ensure all existing tests pass with new types
2. **Serialization Tests**: Verify JSON/YAML serialization works correctly
3. **Integration Tests**: Test full workflow with branded IDs

### 6.3 Rollback Plan

If issues arise, simply revert to the old type definitions:

```go
// Before (branded)
type LinterName = id.ID[linterBrand, string]

// After (rollback)
type LinterName string
```

---

## 7. Alternative: NanoId Integration

### 7.1 When to Use NanoId

For **new identifiers** (not linter/formatter names), consider the `nanoid` package:

```go
import (
    "github.com/larsartmann/go-composable-business-types/id"
    "github.com/larsartmann/go-composable-business-types/nanoid"
)

type RunID = id.ID[runBrand, nanoid.NanoId]

func NewRunID() RunID {
    return id.NewID[runBrand](nanoid.NewNanoId())
}
```

### 7.2 NanoId Benefits

- URL-safe (uses `A-Za-z0-9_-`)
- Cryptographically secure
- 21 characters = 126 bits entropy
- Collision-resistant

---

## 8. Implementation Checklist

### Phase 1: Setup

- [ ] Add `go-composable-business-types` to go.mod
- [ ] Create `pkg/types/ids.go` with branded type definitions
- [ ] Add constructor functions (`NewLinterName`, `NewFormatterName`)

### Phase 2: Core Types

- [ ] Update `pkg/types/types.go` to use branded types
- [ ] Update `LinterInfo.Name` field type
- [ ] Update `FormatterInfo.Name` field type

### Phase 3: Constants

- [ ] Update `pkg/constants/linter_data.go` to use constructors
- [ ] Update `pkg/constants/linter_reasons.go`
- [ ] Update `pkg/constants/formatter_data.go` (if exists)

### Phase 4: Implementation

- [ ] Update `pkg/linter/analyzer.go`
- [ ] Update `pkg/linter/fixer.go`
- [ ] Update `pkg/config/loader.go`

### Phase 5: Testing

- [ ] Run `just test` to verify all tests pass
- [ ] Run `just lint` to verify no new issues
- [ ] Test JSON serialization/deserialization

---

## 9. References

### go-composable-business-types Library

- **Repository**: `/Users/larsartmann/projects/go-composable-business-types`
- **Package**: `github.com/larsartmann/go-composable-business-types/id`
- **Key File**: `id/id.go` - Full implementation of branded ID type

### Current Project Types

- **Location**: `pkg/types/types.go:79-96`
- **Current Types**: `LinterName`, `FormatterName` (string-based)

### Usage Locations

- **Workflow**: `pkg/workflow/workflow.go:139-174`
- **Analyzer**: `pkg/linter/analyzer.go`
- **Constants**: `pkg/constants/linter_data.go`

---

## 10. Conclusion

Integrating `go-composable-business-types/id` would significantly enhance the type safety of the golangci-lint-auto-configure project by:

1. **Preventing bugs** at compile time (mixing linter/formatter names)
2. **Providing rich serialization support** out of the box
3. **Enabling better API design** with type-safe operations
4. **Maintaining backward compatibility** through type aliases

### Recommendation

**Proceed with Phase 1 implementation** (creating the branded type definitions). The risk is minimal, the benefit is clear, and the integration can be done incrementally without breaking existing functionality.

The effort is low (~2-3 hours), and the long-term maintainability benefits are substantial.

---

**Next Steps:**

1. Review this plan with the team
2. Create feature branch for implementation
3. Start with `pkg/types/ids.go` foundation
4. Gradually migrate constants and usage sites
