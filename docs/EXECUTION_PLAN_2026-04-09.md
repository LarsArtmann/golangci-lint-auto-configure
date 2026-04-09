# Comprehensive Analysis & Execution Plan

**Date:** 2026-04-09  
**Status:** In Progress  
**Current Commit:** a0b9e24

---

## Part 1: What Was Forgotten / Could Be Improved

### From Previous Session

1. **Commit Granularity**: Changes were bundled together rather than being committed individually as each pattern was migrated. This makes reviewing harder.

2. **Status File Updates**: Did not create a new status report file documenting the completed work (now addressed in this document).

3. **CommandBuilder Decision**: The CommandBuilder pattern was applied inconsistently:
   - ✅ analyze, configure, report, validate now use CommandBuilder
   - ❌ migrate, completion, install-hook still use traditional approaches
   - **Decision:** Complete the pattern for all commands (medium effort, removes dependency threading)

4. **Pre-existing Test Failure**: `migrateIssuesExcludeFiles` test has been failing since before my changes - needs investigation/fixing.

5. **File Size Warnings**: Multiple files exceed the 350 line limit:
   - pkg/config/merger.go: 685 lines (+335, 95.7% over) - CRITICAL
   - pkg/migration/migrator_test.go: 640 lines (+290, 82.9% over) - CRITICAL
   - pkg/linter/fixer.go: 466 lines (+116, 33.1% over) - Warning

---

## Part 2: Current State Analysis

### Architecture Strengths

1. **Strong Type System**: Extensive use of strongly-typed wrappers (LinterName, ConfigPath, Version, etc.)
2. **Interface-Based Design**: Good separation of concerns with ConfigLoader, LinterAnalyzer, LinterFixer interfaces
3. **Generic Set[T] Implementation**: Clean, reusable set operations with Union, Difference, Intersect, Equal
4. **CommandBuilder Pattern**: Emerging pattern for CLI command construction (partially applied)
5. **Functional Options Pattern**: Used in CommandBuilder.Build() with variadic options

### Architecture Weaknesses

1. **Inconsistent Patterns**: CommandBuilder partially applied creates cognitive load
2. **Large Files**: Several files exceed maintainability thresholds (350 lines)
3. **Missing Set Operations**: Some Set[T] methods could be more useful (IsSubset, IsSuperset)
4. **Limited Error Wrapping**: Some errors lack context for debugging
5. **Unused Interface Components**: ConfigLoader combines many interfaces, some unused

### Dependencies Analysis

**Current Libraries:**

- Cobra (CLI) - Standard, well-maintained
- Charmbracelet Log (logging) - Modern, structured
- Ginkgo/Gomega (testing) - BDD style, comprehensive
- Afero (filesystem abstraction) - Good for testing
- Templ (HTML generation) - Type-safe templates
- Validator (validation) - Industry standard
- samber/mo (functional programming) - Underutilized

**Potential Additions:**

- lo (lodash for Go) - functional utilities
- errgroup - parallel error handling
- golang.org/x/sync/singleflight - deduplication

---

## Part 3: Multi-Step Execution Plan

### Priority Matrix: Work vs Impact

| Task                            | Work   | Impact | Priority | Category      |
| ------------------------------- | ------ | ------ | -------- | ------------- |
| Complete CommandBuilder pattern | Medium | High   | 1        | Architecture  |
| Add Set[T].IsSubset/IsSuperset  | Low    | Medium | 2        | Types         |
| Extract merger subcomponents    | High   | High   | 3        | Refactoring   |
| Use errgroup for parallel ops   | Medium | Medium | 4        | Concurrency   |
| Migrate to mo.Option types      | Medium | Low    | 5        | Types         |
| Add comprehensive Set tests     | Low    | Medium | 6        | Testing       |
| Document architecture decisions | Low    | High   | 7        | Documentation |
| Fix pre-existing test failure   | Medium | High   | 8        | Bug Fix       |

---

## Phase 1: Quick Wins (Low Work, High/Medium Impact)

### Task 1.1: Complete CommandBuilder Pattern ⭐

**Work:** Medium | **Impact:** High | **Priority:** 1

**Current State:**

- ✅ analyze, configure, report, validate use CommandBuilder
- ❌ migrate, completion, install-hook use traditional pattern

**Action:**

1. Migrate `newMigrateCommand` to CommandBuilder
2. Migrate `newCompletionCommand` to CommandBuilder
3. Migrate `newInstallHookCommand` to CommandBuilder
4. Update tests as needed

**Rationale:** Consistency reduces cognitive load. The pattern removes dependency threading through parameters.

---

### Task 1.2: Add Set[T] Utility Methods

**Work:** Low | **Impact:** Medium | **Priority:** 2

**Proposed Additions:**

```go
// IsSubset returns true if all items in s are in other
func (s Set[T]) IsSubset(other Set[T]) bool

// IsSuperset returns true if all items in other are in s
func (s Set[T]) IsSuperset(other Set[T]) bool

// IsProperSubset returns true if s is a subset and not equal
func (s Set[T]) IsProperSubset(other Set[T]) bool

// IsProperSuperset returns true if s is a superset and not equal
func (s Set[T]) IsProperSuperset(other Set[T]) bool
```

**Use Cases:**

- Comparing enabled linters between configs
- Checking if one exclusion set covers another

---

### Task 1.3: Add Comprehensive Set Tests

**Work:** Low | **Impact:** Medium | **Priority:** 6

**Current Coverage:** Union, Difference, Intersect, Equal, basic operations

**Missing:**

- Edge cases (empty sets, nil handling)
- Property-based tests
- Benchmarks for large sets

---

## Phase 2: Structural Improvements

### Task 2.1: Refactor Merger.go ⭐⭐

**Work:** High | **Impact:** High | **Priority:** 3

**Current:** 685 lines (95.7% over limit)

**Proposed Extraction:**

```
pkg/config/
  merger.go              # Core orchestration (~200 lines)
  merger_run.go          # RunConfig merging (~80 lines)
  merger_linters.go      # LintersConfig merging (~100 lines)
  merger_formatters.go   # FormattersConfig merging (~80 lines)
  merger_issues.go       # IssuesConfig merging (~80 lines)
  merger_output.go       # OutputConfig merging (~60 lines)
```

**Benefits:**

- Each file under 100 lines
- Easier to test individual merge strategies
- Clear separation of concerns

---

### Task 2.2: Use errgroup for Parallel Operations

**Work:** Medium | **Impact:** Medium | **Priority:** 4

**Candidates for Parallelization:**

1. Loading multiple config files in MergeConfigs
2. Backup creation operations
3. Validation checks

**Example:**

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)
for _, path := range configPaths {
    path := path // capture loop var
    g.Go(func() error {
        return validateConfig(ctx, path)
    })
}
if err := g.Wait(); err != nil {
    return err
}
```

---

## Phase 3: Type System Enhancements

### Task 3.1: Leverage samber/mo Package

**Work:** Medium | **Impact:** Low-Medium | **Priority:** 5

**Current Usage:** Minimal

**Opportunities:**

```go
// Replace pointer returns with Option types
func FindConfig(dir string) mo.Option[string]
func GetLinterVersion() mo.Option[Version]

// Use Result type for fallible operations
func LoadConfig(path string) mo.Result[*Config]
```

**Benefits:**

- Explicit null handling
- Composable error handling
- Railway-oriented programming

---

### Task 3.2: Add Stringer Implementations

**Work:** Low | **Impact:** Low | **Priority:** - (Nice to have)

**Types missing String() or better formatting:**

- MergeResult
- MigrationResult
- ValidationResult

---

## Phase 4: Bug Fixes & Quality

### Task 4.1: Fix Pre-existing Test Failure ⭐

**Work:** Medium | **Impact:** High | **Priority:** 8

**Failure:** `migrateIssuesExcludeFiles` - YAML parsing error

**Investigation Needed:**

1. Check test fixture YAML format
2. Verify v2ConfigWithExcludeFiles helper generates valid YAML
3. Fix indentation or quoting issues

---

## Phase 5: Documentation

### Task 5.1: Document Architecture Decisions ⭐

**Work:** Low | **Impact:** High | **Priority:** 7

**Create ADRs (Architecture Decision Records):**

1. ADR-001: Use of Generic Set[T] type
2. ADR-002: CommandBuilder pattern for CLI
3. ADR-003: Interface-based design for testability
4. ADR-004: BDD testing with Ginkgo/Gomega

---

## Phase 6: Established Libraries Integration

### Libraries to Consider

| Library                            | Use Case             | Current Status                       |
| ---------------------------------- | -------------------- | ------------------------------------ |
| samber/lo                          | Functional utilities | Not used, could replace some loops   |
| golang.org/x/sync/errgroup         | Parallel operations  | Not used, good for config loading    |
| golang.org/x/sync/singleflight     | Deduplication        | Not used, good for repeated analysis |
| github.com/hashicorp/go-multierror | Error aggregation    | Could improve error handling         |

---

## Execution Order Recommendation

```
Week 1 (Immediate):
├── Task 1.1: Complete CommandBuilder
├── Task 1.2: Add Set[T] methods
└── Task 4.1: Fix test failure

Week 2 (Structural):
├── Task 2.1: Refactor merger.go
└── Task 1.3: Add Set tests

Week 3 (Enhancement):
├── Task 2.2: Use errgroup
└── Task 3.1: Leverage mo package

Week 4 (Documentation):
└── Task 5.1: Document architecture
```

---

## Verification Checklist

After each task:

- [ ] All tests pass (`just test`)
- [ ] Build succeeds (`go build ./...`)
- [ ] Lint passes (`just lint`)
- [ ] No new file size violations
- [ ] Commit with descriptive message
- [ ] Update this document

---

## Questions for User

1. **CommandBuilder Priority:** Should I prioritize completing CommandBuilder for all commands, or is the current partial state acceptable?

2. **Merger Refactoring:** The merger.go file is 685 lines (95% over limit). Should I prioritize splitting it into smaller files?

3. **Library Adoption:** Would you prefer to adopt samber/lo for functional utilities, or keep the codebase dependency-light?

4. **Test Failure:** The `migrateIssuesExcludeFiles` test has been failing. Should I prioritize fixing this pre-existing issue?

5. **Scope:** Should I proceed with implementing all phases, or focus on a specific subset?
