# Comprehensive Multi-Step Execution Plan

## golangci-lint-auto-configure

**Date:** 2026-03-26 21:55
**Purpose:** Brutally honest reflection + actionable improvement plan

---

## 1. BRUTALLY HONEST REFLECTION

### What I Forgot

| Item                                 | Impact | Why I Forgot                                                |
| ------------------------------------ | ------ | ----------------------------------------------------------- |
| **Broken test file**                 | HIGH   | Committed code with `errors.New` instead of `stderrors.New` |
| **go mod tidy**                      | MEDIUM | Didn't run after module name fix                            |
| **Consolidate duplicate git errors** | MEDIUM | Found but didn't fix                                        |
| **Check all tests pass**             | HIGH   | Assumed tests passed without verifying                      |
| **Review unstaged changes**          | MEDIUM | 55 files modified, committed without review                 |
| **gopls cache issues**               | LOW    | IDE showing stale errors                                    |

### What Could I Have Done Better

1. **Test before commit** - Always run `go test ./...` before committing
2. **Smaller commits** - 55 files in one commit is too large
3. **Review changes** - Should have reviewed each file before staging
4. **Verify build** - Should have verified `go build ./...` passes
5. **Check diagnostics** - Should have addressed all compiler errors

### What I Can Still Improve

1. **Create proper test coverage** - pkg/client, pkg/report, pkg/types have no tests
2. **Consolidate duplicate code** - Two Config structs, two FormatRecommendations
3. **Fix CLI test timeout** - Integration tests rebuild binary each time
4. **Remove unused dependencies** - 18+ unused deps in go.mod
5. **Add pre-commit optimization** - Skip slow tests in pre-commit hook

---

## 2. COMPREHENSIVE MULTI-STEP EXECUTION PLAN

### Phase 1: Critical Fixes (Do Immediately)

| Step | Task                      | Work | Impact | Existing Code |
| ---- | ------------------------- | ---- | ------ | ------------- |
| 1.1  | Fix broken errors_test.go | 5min | HIGH   | ✅ DONE       |
| 1.2  | Run go mod tidy           | 1min | HIGH   | N/A           |
| 1.3  | Verify all pkg tests pass | 2min | HIGH   | N/A           |
| 1.4  | Commit and push           | 1min | HIGH   | N/A           |

### Phase 2: Code Quality (This Session)

| Step | Task                             | Work  | Impact | Existing Code                           |
| ---- | -------------------------------- | ----- | ------ | --------------------------------------- |
| 2.1  | Consolidate duplicate git errors | 10min | MEDIUM | pkg/errors/errors.go + pkg/utils/git.go |
| 2.2  | Remove unused dependencies       | 5min  | MEDIUM | go.mod                                  |
| 2.3  | Add tests for pkg/client         | 30min | HIGH   | Use existing test patterns              |
| 2.4  | Add tests for pkg/types          | 20min | HIGH   | Use existing test patterns              |

### Phase 3: Architecture Improvements (Next Session)

| Step | Task                                  | Work  | Impact | Existing Code                     |
| ---- | ------------------------------------- | ----- | ------ | --------------------------------- |
| 3.1  | Document why two Config structs exist | 10min | MEDIUM | pkg/types + pkg/migration         |
| 3.2  | Consider unifying Config structs      | 2hr   | HIGH   | Complex - needs careful analysis  |
| 3.3  | Consolidate FormatRecommendations     | 15min | MEDIUM | pkg/linter + pkg/ui               |
| 3.4  | Add stringer for enums                | 30min | LOW    | LinterPriority, FormatterPriority |

### Phase 4: Test Infrastructure (Future)

| Step | Task                           | Work  | Impact | Existing Code              |
| ---- | ------------------------------ | ----- | ------ | -------------------------- |
| 4.1  | Fix CLI test timeout           | 1hr   | HIGH   | Use build tags             |
| 4.2  | Add integration test build tag | 30min | HIGH   | internal/cli/\*\_test.go   |
| 4.3  | Optimize pre-commit hook       | 15min | MEDIUM | scripts/pre-commit-hook.sh |
| 4.4  | Add test coverage reporting    | 30min | MEDIUM | Use go tool cover          |

### Phase 5: Library Improvements (Future)

| Step | Task                                      | Work  | Impact | Existing Code                |
| ---- | ----------------------------------------- | ----- | ------ | ---------------------------- |
| 5.1  | Consider go-cmd/cmd for command execution | 1hr   | MEDIUM | pkg/linter/command_runner.go |
| 5.2  | Standardize on afero for filesystem       | 2hr   | MEDIUM | Mixed os and afero usage     |
| 5.3  | Add stringer for enum String() methods    | 30min | LOW    | Multiple enums               |

---

## 3. SORTED BY WORK vs IMPACT

### High Impact, Low Work (Do First)

| # | Task                              | Work  | Impact |
| - | --------------------------------- | ----- | ------ |
| 1 | Run go mod tidy                   | 1min  | HIGH   |
| 2 | Verify all tests pass             | 2min  | HIGH   |
| 3 | Remove unused dependencies        | 5min  | MEDIUM |
| 4 | Consolidate git errors            | 10min | MEDIUM |
| 5 | Document Config struct separation | 10min | MEDIUM |

### High Impact, Medium Work (Do Second)

| # | Task                              | Work  | Impact |
| - | --------------------------------- | ----- | ------ |
| 6 | Add tests for pkg/client          | 30min | HIGH   |
| 7 | Add tests for pkg/types           | 20min | HIGH   |
| 8 | Add integration test build tags   | 30min | HIGH   |
| 9 | Consolidate FormatRecommendations | 15min | MEDIUM |

### Medium Impact, Medium Work (Do Third)

| #  | Task                     | Work  | Impact |
| -- | ------------------------ | ----- | ------ |
| 10 | Fix CLI test timeout     | 1hr   | HIGH   |
| 11 | Consider go-cmd/cmd      | 1hr   | MEDIUM |
| 12 | Optimize pre-commit hook | 15min | MEDIUM |

### Low Impact, Any Work (Do Last)

| #  | Task                             | Work  | Impact |
| -- | -------------------------------- | ----- | ------ |
| 13 | Add stringer for enums           | 30min | LOW    |
| 14 | Standardize on afero             | 2hr   | MEDIUM |
| 15 | Consider unifying Config structs | 2hr   | HIGH   |

---

## 4. EXISTING CODE CONSIDERATIONS

### Before Implementing, Check:

1. **pkg/client/client.go** - Already has `Client` struct with methods, just needs tests
2. **pkg/types/types.go** - Has all core types, just needs validation tests
3. **pkg/errors/errors.go** - Has custom error types, consolidate git errors here
4. **pkg/ui/formatter.go** - Has `FormatRecommendations`, remove duplicate from linter
5. **pkg/linter/command_runner.go** - Uses `os/exec`, could use go-cmd/cmd

### Don't Reinvent:

- Use existing `mo.Result[T]` for functional error handling
- Use existing `apperrors` package for all errors
- Use existing test patterns from `pkg/linter/analyzer_test.go`
- Use existing interface patterns from `pkg/types/types.go`

---

## 5. TYPE MODEL IMPROVEMENTS

### Current Issues

1. **Two Config structs** - `types.Config` and `migration.Config`
2. **Duplicate error definitions** - `ErrNotGitRepository` in two places
3. **Inconsistent naming** - Some use `LinterName`, some use `string`

### Proposed Improvements

```go
// 1. Consolidate git errors in pkg/errors
var (
    ErrNotGitRepository    = errors.New("not a git repository")
    ErrNotInGitWorkingTree = errors.New("not in a git working tree")
)

// 2. Remove duplicate from pkg/utils/git.go

// 3. Document why two Config structs exist:
// - types.Config: v2 output format (what we generate)
// - migration.Config: v1 input format (what we migrate from)

// 4. Consider type alias for clarity:
type V1Config = migration.Config  // Input format
type V2Config = types.Config      // Output format
```

### Type Safety Improvements

```go
// Already good:
type LinterName string    // Strongly typed
type FormatterName string // Strongly typed

// Consider adding:
type ConfigPath string    // Prevent typos in config paths
type Version string       // For golangci-lint version
```

---

## 6. WELL-ESTABLISHED LIBRARIES TO LEVERAGE

### Already Using (Good!)

| Library                   | Purpose                | Status    |
| ------------------------- | ---------------------- | --------- |
| `samber/mo`               | Functional programming | ✅ Active |
| `go-playground/validator` | Struct validation      | ✅ Active |
| `spf13/cobra`             | CLI framework          | ✅ Active |
| `spf13/afero`             | Virtual filesystem     | ✅ Active |
| `gopkg.in/yaml.v3`        | YAML handling          | ✅ Active |
| `charmbracelet/fang`      | Enhanced CLI           | ✅ Active |
| `charmbracelet/log`       | Structured logging     | ✅ Active |
| `charmbracelet/lipgloss`  | Terminal styling       | ✅ Active |

### Consider Adding

| Library                           | Purpose           | Benefit                                     |
| --------------------------------- | ----------------- | ------------------------------------------- |
| `go-cmd/cmd`                      | Command execution | Better testing, streaming output            |
| `stretchr/testify`                | Assertions        | More readable tests (alternative to gomega) |
| `golang.org/x/tools/cmd/stringer` | Enum String()     | Auto-generate String() methods              |
| `gotest.tools/v3`                 | Test utilities    | Better test assertions                      |

### Don't Add (Redundant)

| Library            | Reason                                   |
| ------------------ | ---------------------------------------- |
| `pkg/errors`       | Use stdlib `%w` wrapping + custom types  |
| `emperror`         | Overkill for our needs                   |
| `uber-go/multierr` | Can use stdlib `errors.Join` in Go 1.20+ |

---

## 7. IMMEDIATE ACTION ITEMS

### Execute Now (In Order):

1. ✅ Fix errors_test.go (DONE)
2. ⏳ Run go mod tidy
3. ⏳ Verify all pkg tests pass
4. ⏳ Consolidate git errors
5. ⏳ Commit and push
6. ⏳ Add tests for pkg/client
7. ⏳ Add tests for pkg/types
8. ⏳ Create status report

---

## 8. TOP #1 QUESTION

**Should we use build tags for integration tests?**

Current problem: CLI tests timeout because they rebuild the binary for each test case.

Options:

1. **Build tags** - `//go:build integration` - Run with `go test -tags=integration`
2. **Short flag** - Use `testing.Short()` and run with `go test -short`
3. **Shared binary** - Build once, share across tests (complex)
4. **Test containers** - Overkill for our needs

**Recommendation:** Option 1 (build tags) - Most explicit and controllable.

---

_Generated: 2026-03-26 21:55_
_Next Action: Run go mod tidy_
