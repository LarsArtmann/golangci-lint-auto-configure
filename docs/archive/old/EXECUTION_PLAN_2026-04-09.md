# Comprehensive Execution Plan: Code Quality & Architecture Improvements

**Date:** 2026-04-09\
**Status:** Ready for Execution\
**Priority:** High Impact, Incremental Delivery

---

## Phase 1: What I Forgot / Could Improve

### Critical Issues Discovered

1. **Go Build Cache Corruption**
   - Build cache has missing files, preventing compilation
   - Need to clear and rebuild cache
   - Impact: BLOCKING - cannot run tests

2. **LSP False Positives**
   - categorizer_test.go shows errors at lines 36, 110
   - These are likely golangci-lint-ls artifacts, not real errors
   - Need to verify actual compilation status

3. **File Size Violations** (9 files over 350 lines)
   - merger.go: 689 lines (+339, 96.9% over) - CRITICAL
   - migrator_test.go: 648 lines (+298, 85.1% over) - HIGH
   - fixer.go: 466 lines (+116, 33.1% over) - MEDIUM
   - loader.go: 422 lines (+72, 20.6% over) - LOW

4. **Missing Architecture Documentation**
   - No ADRs for key decisions
   - Future maintainers lack context
   - Technical debt accumulating

---

## Phase 2: Multi-Step Execution Plan

### Step 1: Fix Build Environment (BLOCKING) ⚡

**Work:** Low | **Impact:** CRITICAL

- [ ] Clear Go build cache
- [ ] Verify all packages compile
- [ ] Run full test suite
- [ ] Fix any actual compilation errors

**Verification:** `go build ./... && ginkgo -r ./...`

---

### Step 2: Create ADR-001 - Set[T] Type Decision 📝

**Work:** Low | **Impact:** HIGH

Document:

- Why generic Set[T] was introduced
- Comparison with map[T]struct{}
- Performance considerations
- Migration strategy from old sets

**Template:** Use ADR template from architecture standards

---

### Step 3: Create ADR-002 - CommandBuilder Pattern 📝

**Work:** Low | **Impact:** HIGH

Document:

- Problem: CLI commands had scattered dependency management
- Solution: CommandBuilder with common dependencies
- Benefits: Testability, consistency, reduced boilerplate
- Tradeoffs: Slight increase in abstraction

---

### Step 4: Create ADR-003 - Interface-Based Design 📝

**Work:** Low | **Impact:** HIGH

Document:

- ConfigLoader interface
- LinterAnalyzer interface
- LinterFixer interface
- Benefits for testing and mocking

---

### Step 5: Refactor merger.go - Extract Helper Functions 🔧

**Work:** Medium | **Impact:** HIGH

**Strategy:** Extract internal helpers first, then config sections

Files to create:

- `merger_helpers.go`: mergeSettingsMaps, mergeStringSlices, mergePaths
- Keep merger.go for public API and orchestration

**Benefits:**

- Reduces merger.go by ~100 lines
- Separates concerns
- Easier to test helpers independently

---

### Step 6: Refactor merger.go - Extract Config Section Mergers 🔧

**Work:** High | **Impact:** HIGH

**Strategy:** Create merger sections by config type

Files to create:

- `merger_run.go`: Run config merging (timeout, tests, etc.)
- `merger_linters.go`: Linters enable/disable/Settings
- `merger_formatters.go`: Formatters configuration
- `merger_issues.go`: Issues configuration
- `merger_output.go`: Output configuration
- `merger_exclusions.go`: Exclusions configuration

**Pattern:**

```go
// merger_run.go
package config

func (cm *Merger) mergeRunConfig(primary, secondary *RunConfig) int {
    // Implementation
}
```

**Verification:** All existing tests pass without modification

---

### Step 7: Leverage samber/mo for Better Type Safety 🔄

**Work:** Medium | **Impact:** MEDIUM

**Current Pattern (pointer returns):**

```go
func (cm *Merger) MergeConfigs(configPaths []string) (*Config, *MergeResult, error)
```

**Improved Pattern (Option types):**

```go
func (cm *Merger) MergeConfigs(configPaths []string) (mo.Option[Config], mo.Option[MergeResult], error)
```

**Benefits:**

- Explicit nil handling
- Functional operations (Map, FlatMap)
- Better composition

**Start with:** New functions, migrate existing gradually

---

### Step 8: Use errgroup for Parallel Config Loading ⚡

**Work:** Medium | **Impact:** MEDIUM

**Current:** Sequential loading in MergeConfigs
**Improved:** Parallel loading with errgroup

```go
import "golang.org/x/sync/errgroup"

func (cm *Merger) loadConfigsParallel(paths []string) ([]Config, error) {
    var g errgroup.Group
    configs := make([]Config, len(paths))

    for i, path := range paths {
        g.Go(func() error {
            config, err := cm.loadConfig(path)
            if err != nil {
                return err
            }
            configs[i] = config
            return nil
        })
    }

    return configs, g.Wait()
}
```

**Benefits:**

- Faster merging with multiple configs
- Clean error handling
- Bounded concurrency

---

### Step 9: Add Set[T] Benchmarks 📊

**Work:** Low | **Impact:** LOW

Create `pkg/types/set_bench_test.go`:

- Benchmark Add/Contains for small sets (10 items)
- Benchmark Add/Contains for medium sets (100 items)
- Benchmark Add/Contains for large sets (10000 items)
- Benchmark Intersect/Difference/Union

**Benefits:**

- Performance regression detection
- Optimization guidance
- Documentation of performance characteristics

---

### Step 10: Consolidate Duplicate Validation Logic 🔧

**Work:** Medium | **Impact:** MEDIUM

**Current:** Validation scattered across packages
**Target:** Centralized validation in `pkg/types/validation.go`

Extract:

- Config validation
- Linter name validation
- Path validation

**Benefits:**

- Single source of truth
- Reusable validation
- Consistent error messages

---

### Step 11: Refactor migrator_test.go - Extract Helpers 🔧

**Work:** Medium | **Impact:** MEDIUM

**Current:** 648 lines with many helper functions
**Target:** Extract to `pkg/migration/testhelpers.go`

Helpers to extract:

- `v2ConfigWithExcludeFiles`
- `v2ConfigWithExcludeDirs`
- `runMigration`
- `testMigrationWithConfig`
- `testSimpleMigration`

**Benefits:**

- Reduces test file size
- Reusable test fixtures
- Clearer test intent

---

### Step 12: Optimize Pre-Commit Hook Performance ⚡

**Work:** Medium | **Impact:** MEDIUM

**Current Issues:**

- Library policy scanner: ~2.2s
- AST analyzer: failing with unknown command
- gitleaks: scanning all commits

**Improvements:**

- Cache library policy results
- Fix AST analyzer command
- Skip slow scans on doc-only changes

---

### Step 13: Create ADR-004 - BDD Testing Approach 📝

**Work:** Low | **Impact:** MEDIUM

Document:

- Why Ginkgo/Gomega over standard testing
- Describe/Context/It pattern
- Table-driven tests with Ginkgo
- When to use standard testing

---

### Step 14: Add Property-Based Tests 🧪

**Work:** High | **Impact:** LOW

Consider adding `testing/quick` or `pgregory/rapid` for:

- Set operation properties (commutativity, associativity)
- Config parsing round-trips
- Migration idempotency

**Benefits:**

- Find edge cases
- Document invariants
- Regression prevention

---

### Step 15: Extract detector.go Subcomponents 🔧

**Work:** Medium | **Impact:** LOW

**Current:** 372 lines handling multiple detection strategies
**Target:** Split by detection type

Files:

- `detector_framework.go`: Web framework detection
- `detector_project.go`: Project type detection
- `detector_patterns.go`: Pattern matching

---

## Phase 3: Execution Priority Matrix

| Step | Task                      | Work   | Impact   | Priority |
| ---- | ------------------------- | ------ | -------- | -------- |
| 1    | Fix build environment     | Low    | CRITICAL | P0       |
| 2    | ADR-001: Set[T]           | Low    | HIGH     | P1       |
| 3    | ADR-002: CommandBuilder   | Low    | HIGH     | P1       |
| 4    | ADR-003: Interfaces       | Low    | HIGH     | P1       |
| 5    | Extract merger helpers    | Medium | HIGH     | P1       |
| 6    | Extract merger sections   | High   | HIGH     | P1       |
| 7    | samber/mo integration     | Medium | MEDIUM   | P2       |
| 8    | errgroup parallel loading | Medium | MEDIUM   | P2       |
| 9    | Set benchmarks            | Low    | LOW      | P3       |
| 10   | Validation consolidation  | Medium | MEDIUM   | P2       |
| 11   | migrator_test helpers     | Medium | MEDIUM   | P2       |
| 12   | Pre-commit optimization   | Medium | MEDIUM   | P2       |
| 13   | ADR-004: BDD Testing      | Low    | MEDIUM   | P2       |
| 14   | Property-based tests      | High   | LOW      | P3       |
| 15   | detector subcomponents    | Medium | LOW      | P3       |

---

## Phase 4: Reflection on Existing Code

### What We Already Have ✅

1. **Set[T] Type** (`pkg/types/set.go`)
   - Comprehensive methods: Add, Remove, Contains, Union, Intersect, Difference
   - Set operations: IsSubset, IsSuperset, IsProperSubset, IsProperSuperset
   - Generic, type-safe
   - **Use for:** Any collection operations, deduplication

2. **CommandBuilder** (`internal/cli/cmd_builder.go`)
   - Dependency injection pattern
   - Common dependencies: logger, analyzer, configLoader
   - **Use for:** New CLI commands

3. **samber/mo** (in go.mod)
   - Option[T] for nullable values
   - Result[T] for fallible operations
   - **Use for:** Better error handling, functional composition

4. **errgroup** (in go.mod via golang.org/x/sync)
   - Parallel error handling
   - **Use for:** Parallel config loading

5. **Interface Design** (`pkg/types/types.go`)
   - ConfigLoader, LinterAnalyzer, LinterFixer
   - **Use for:** Test mocking, dependency injection

### What We Should Add 📦

1. **singleflight** (golang.org/x/sync/singleflight)
   - Deduplicate concurrent calls
   - **Use for:** Config loading cache

2. **lo** (samber/lo)
   - Functional utilities: Map, Filter, Reduce
   - **Consider for:** Collection operations (but we have Set[T])

---

## Phase 5: Type Model Improvements

### Current Architecture

```go
// Returns pointers, can be nil
func (cm *Merger) MergeConfigs(configPaths []string) (*Config, *MergeResult, error)

// Returns slices, empty if none
func (cm *Merger) mergeStringSlices(primary, secondary []string) int
```

### Improved Architecture

```go
// Returns Option types, explicit about emptiness
func (cm *Merger) MergeConfigs(configPaths []string) (mo.Option[Config], mo.Option[MergeResult], error)

// Returns Result type for fallible operations
func (cm *Merger) LoadConfig(path string) mo.Result[Config]

// Set[T] for collections
type LinterSet = types.Set[types.LinterName]
```

### Benefits

1. **Type Safety:** nil handling is explicit
2. **Composition:** Map/FlatMap chain operations
3. **Clarity:** Return types document behavior
4. **Testing:** Easier to assert on Option/Result states

---

## Phase 6: How to Use Established Libraries

### samber/mo Integration Strategy

**Step 1: New Functions**

```go
func LoadConfigSafe(path string) mo.Result[Config] {
    config, err := LoadConfig(path)
    if err != nil {
        return mo.Err[Config](err)
    }
    return mo.Ok(config)
}
```

**Step 2: Gradual Migration**

- Keep old functions for backward compatibility
- Add new functions with mo types
- Deprecate old functions over time

**Step 3: Full Migration**

- Remove pointer returns
- Use Option for optional values
- Use Result for fallible operations

### errgroup Integration Strategy

**Step 1: Identify Parallelizable Operations**

- Config loading in MergeConfigs
- Validation checks
- File operations

**Step 2: Create Parallel Versions**

```go
func (cm *Merger) MergeConfigsParallel(configPaths []string) (*Config, *MergeResult, error)
```

**Step 3: Benchmark and Compare**

- Measure sequential vs parallel
- Only use parallel when beneficial

---

## Execution Checklist

### Before Starting

- [ ] Clear Go build cache
- [ ] Verify all tests pass
- [ ] Commit current state

### During Execution

- [ ] Run tests after each change
- [ ] Commit each self-contained change
- [ ] Update documentation

### After Completion

- [ ] Run full test suite
- [ ] Update status report
- [ ] Push to remote
- [ ] Review changes

---

## Top Question: How to Split merger.go Without Breaking Changes?

### Challenge

merger.go has:

- Public API: Merger struct, MergeConfigs, MergeResult
- Private helpers: mergeSettingsMaps, mergeStringSlices, mergePaths
- Config section mergers: mergeRunConfig, mergeLintersConfig, etc.

### Proposed Solution

**Step 1: Extract Private Helpers**

```
pkg/config/
├── merger.go (public API, reduced by ~60 lines)
└── merger_helpers.go (private helpers, ~60 lines)
```

**Step 2: Extract Config Section Interfaces**

```
pkg/config/
├── merger.go (orchestration)
├── merger_run.go (RunConfig merging)
├── merger_linters.go (LintersConfig merging)
└── ...
```

**Step 3: Maintain Backward Compatibility**

- Keep Merger struct in merger.go
- Keep MergeConfigs in merger.go
- Use internal functions from other files

### Verification

After refactoring:

```bash
go build ./...
ginkgo -r ./pkg/config/...
# All tests should pass without modification
```

---

## Summary

This execution plan provides:

1. **Immediate fixes** (build cache, ADRs)
2. **High-impact refactoring** (merger.go split)
3. **Type system improvements** (samber/mo integration)
4. **Performance optimizations** (errgroup, benchmarks)
5. **Documentation** (ADRs, property tests)

**Recommended Start:** Steps 1-6 (build fix + ADRs + merger refactoring)

**Estimated Time:** 2-3 focused sessions
**Impact:** Significant code quality and maintainability improvements

---

_Plan generated by Crush AI Assistant_
_Date: 2026-04-09_
_Assisted-by: Kimi K2.5 via Crush <crush@charm.land>_
