# Improvement Plan - golangci-lint-auto-configure

Generated: 2026-03-26

## Analysis Summary

### What Was Found

1. **Type Model Issues**: Duplicate `Config` structs, overlapping error types, missing abstractions
2. **Code Duplication**: Git repo checks, config loaders, deprecated linter handling
3. **Error Handling**: Inconsistent wrapping, missing context, mixed patterns
4. **Test Gaps**: 0% coverage in errors/, constants/, client/, workflow/, report/
5. **Architecture**: Fixer depends on concrete Analyzer, 313-line god method

### What We Did Well

- Interface-based design in pkg/types/types.go
- Strong typing with custom types (LinterName, FormatterName)
- Good use of samber/mo for functional patterns
- Proper separation of concerns in most packages

### What Could Be Better

- Consolidate duplicate types across packages
- Standardize error handling patterns
- Add missing test coverage
- Extract god methods into focused functions

---

## Prioritized Action Plan

### Tier 1: High Impact, Low Effort (Quick Wins)

| #   | Task                                              | Impact | Effort | File(s)                                      |
| --- | ------------------------------------------------- | ------ | ------ | -------------------------------------------- |
| 1.1 | Extract shared git repo check to pkg/utils/git.go | High   | Small  | loader.go:367-392, migrator.go:246-262       |
| 1.2 | Add tests for pkg/errors/ (0% → 80%+)             | High   | Small  | pkg/errors/\*\_test.go (new)                 |
| 1.3 | Standardize error wrapping to use apperrors       | Medium | Small  | loader.go, yaml_loader.go, command_runner.go |
| 1.4 | Fix Fixer to depend on LinterAnalyzer interface   | Medium | Small  | pkg/linter/fixer.go:17-20                    |

### Tier 2: High Impact, Medium Effort

| #   | Task                                                      | Impact | Effort | File(s)                                           |
| --- | --------------------------------------------------------- | ------ | ------ | ------------------------------------------------- |
| 2.1 | Consolidate Config types: migration uses types.Config     | High   | Medium | pkg/migration/config_types.go, pkg/types/types.go |
| 2.2 | Extract god method FixConfigResult into focused functions | High   | Medium | pkg/linter/fixer.go, pkg/linter/fixer\_\*.go      |
| 2.3 | Add tests for pkg/client/ (0% → 70%+)                     | High   | Medium | pkg/client/\*\_test.go (new)                      |
| 2.4 | Consolidate YAML loading to use config.Loader everywhere  | Medium | Medium | pkg/migration/yaml_loader.go                      |

### Tier 3: Medium Impact, Low Effort

| #   | Task                                                            | Impact | Effort | File(s)                              |
| --- | --------------------------------------------------------------- | ------ | ------ | ------------------------------------ |
| 3.1 | Create LinterMetadata struct consolidating priority+reason maps | Medium | Small  | pkg/constants/linter_data.go         |
| 3.2 | Add context to errors missing file paths                        | Medium | Small  | migrator.go:75, command_runner.go:31 |
| 3.3 | Add tests for pkg/constants/ (0% → 70%+)                        | Medium | Small  | pkg/constants/\*\_test.go (new)      |
| 3.4 | Add VersionError custom error type                              | Low    | Small  | pkg/errors/errors.go                 |

### Tier 4: Low Impact, Large Effort (Future Consideration)

| #   | Task                                                          | Impact | Effort | File(s)                                         |
| --- | ------------------------------------------------------------- | ------ | ------ | ----------------------------------------------- |
| 4.1 | Use go-git library instead of exec.Command for git operations | Low    | Large  | pkg/config/loader.go, pkg/migration/migrator.go |
| 4.2 | Create Recommendation interface for Linter/Formatter          | Low    | Small  | pkg/types/types.go                              |
| 4.3 | Add tests for pkg/workflow/ (0% → 60%+)                       | Medium | Medium | pkg/workflow/\*\_test.go (new)                  |
| 4.4 | Add tests for pkg/report/ (0% → 70%+)                         | Medium | Small  | pkg/report/\*\_test.go (new)                    |

---

## Implementation Order

### Phase 1: Foundation (This Session)

1. Extract shared git utilities
2. Add error package tests
3. Standardize error wrapping

### Phase 2: Consolidation (Next Session)

1. Consolidate Config types
2. Extract god methods
3. Add client package tests

### Phase 3: Polish (Future)

1. Create LinterMetadata struct
2. Add remaining test coverage
3. Consider library replacements

---

## Architecture Improvements

### Type Model Changes

**Current State:**

```
pkg/types/types.go:Config (v2 schema)
pkg/migration/config_types.go:Config (v1+v2 schema)
```

**Target State:**

```
pkg/types/types.go:Config (v2 schema - single source of truth)
pkg/types/types.go:ConfigV1 (v1 schema for migration only)
pkg/migration/ uses types.Config and types.ConfigV1
```

### Error Handling Standardization

**Current State:**

- Mix of `fmt.Errorf("...: %w", err)` and `apperrors.NewXxxError(...)`
- Missing context in some errors

**Target State:**

- All errors use apperrors package
- All errors include relevant context (file paths, operation names)
- Consistent wrapping pattern

### Interface Improvements

**Current State:**

```go
// Fixer depends on concrete type
type Fixer struct {
    analyzer *Analyzer  // concrete
}
```

**Target State:**

```go
// Fixer depends on interface
type Fixer struct {
    analyzer LinterAnalyzer  // interface
}
```

---

## Test Coverage Goals

| Package        | Current | Target | Priority |
| -------------- | ------- | ------ | -------- |
| pkg/errors/    | 0%      | 80%    | High     |
| pkg/client/    | 0%      | 70%    | High     |
| pkg/constants/ | 0%      | 70%    | Medium   |
| pkg/workflow/  | 0%      | 60%    | Medium   |
| pkg/report/    | 0%      | 70%    | Medium   |
| pkg/linter/    | ~60%    | 75%    | Medium   |

---

## Notes

- All changes should maintain backward compatibility
- Each change should be committed separately
- Run `just test && just lint` after each change
- Update AGENTS.md if patterns change significantly
