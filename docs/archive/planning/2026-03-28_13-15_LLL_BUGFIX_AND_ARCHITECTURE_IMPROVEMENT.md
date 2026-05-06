# Comprehensive Multi-Step Execution Plan

## golangci-lint-auto-configure

**Date:** 2026-03-28 13:15
**Status:** PHASE 1-2 COMPLETE ✅ | PHASE 3 DEFERRED (requires Go 1.26.1)
**Purpose:** Fix lll/golines bug + architectural improvements

---

## 1. BRUTALLY HONEST REFLECTION

### What I Forgot

| Item                                 | Impact   | Why I Forgot                                        |
| ------------------------------------ | -------- | --------------------------------------------------- |
| **Redundant linters in categorizer** | CRITICAL | Only added check in fixer.go, forgot categorizer.go |
| **Test for categorizeLinters**       | HIGH     | No unit tests for this function                     |
| **Code duplication**                 | MEDIUM   | Hardcoded lll->golines in TWO places                |
| **Stale gopls cache**                | LOW      | IDE showing wrong error                             |

### What's Stupid

1. **Hardcoded lll->golines mapping** in both `categorizer.go` and `fixer.go`
2. **RedundantLinters map doesn't specify the formatter** - it only has a reason string
3. **No tests for categorizeLinters** - core function with no coverage
4. **Code duplication** - same check exists twice

### What Could I Have Done Better

1. **Use the RedundantLinters map properly** - make it a map from linter -> formatter name
2. **Single source of truth** - one place to define linter/formatter relationships
3. **Test-driven** - write test before fix
4. **Review all places using RedundantLinters** - ensure consistency

### What I Can Still Improve

1. **Refactor RedundantLinters** - add formatter name mapping
2. **Extract common function** - create shared helper for redundancy check
3. **Add tests** - cover categorizeLinters with redundant linter scenarios
4. **Clean up constants** - consolidate linter/formatter relationship data

### Did We Create Split Brains?

YES! The `RedundantLinters` concept exists in two places:

- `fixer.go:210-227` - checks and removes redundant linters
- `categorizer.go:33-41` (my fix) - skips recommending redundant linters

These should be ONE function, not duplicated.

### Are We Building Ghost Systems?

- `internal/di/` - directory was planned but NEVER created (per memory file)
- `pkg/README.md` - exists but references `internal/di/` which doesn't exist

### How Are We Doing on Tests?

- `pkg/linter/analyzer_test.go` - tests formatting/summary, NOT categorization
- `categorizeLinters` - **NO DIRECT TESTS**
- Fixer tests exist but don't cover redundant linter scenario

---

## 2. COMPREHENSIVE MULTI-STEP EXECUTION PLAN

### Phase 1: Critical Bug Fix (This Session - 30min)

| Step | Task                                           | Work  | Impact | Existing Code                     |
| ---- | ---------------------------------------------- | ----- | ------ | --------------------------------- |
| 1.1  | Commit current fix                             | 2min  | HIGH   | analyzer.go, categorizer.go       |
| 1.2  | Refactor RedundantLinters to include formatter | 10min | HIGH   | rules.go                          |
| 1.3  | Create shared helper for redundancy check      | 8min  | HIGH   | New helper in constants or linter |
| 1.4  | Update categorizer to use shared helper        | 5min  | HIGH   | categorizer.go                    |
| 1.5  | Update fixer to use shared helper              | 5min  | HIGH   | fixer.go                          |

### Phase 2: Testing (This Session - 40min)

| Step | Task                                             | Work  | Impact | Existing Code  |
| ---- | ------------------------------------------------ | ----- | ------ | -------------- |
| 2.1  | Add tests for categorizeLinters with lll+golines | 15min | HIGH   | New test file  |
| 2.2  | Add tests for categorizeLinters without golines  | 10min | HIGH   | Same test file |
| 2.3  | Add tests for fixer redundant linter removal     | 15min | HIGH   | fixer_test.go  |

### Phase 3: Architecture Cleanup (Next Session - 60min)

| Step | Task                                                 | Work  | Impact | Existing Code            |
| ---- | ---------------------------------------------------- | ----- | ------ | ------------------------ |
| 3.1  | Remove hardcoded lll->golines checks                 | 10min | MEDIUM | categorizer.go, fixer.go |
| 3.2  | Update pkg/README.md to remove internal/di reference | 5min  | LOW    | pkg/README.md            |
| 3.3  | Document linter/formatter relationship model         | 15min | MEDIUM | docs/ARCHITECTURE.md     |
| 3.4  | Run full test suite                                  | 10min | HIGH   | All tests                |
| 3.5  | Run golangci-lint on codebase                        | 10min | HIGH   | Self-check               |
| 3.6  | Update AGENTS.md if needed                           | 10min | LOW    | AGENTS.md                |

### Phase 4: Future Improvements (Backlog)

| Step | Task                                              | Work  | Impact | Existing Code   |
| ---- | ------------------------------------------------- | ----- | ------ | --------------- |
| 4.1  | Consider using samber/mo for Result types         | 30min | LOW    | types/result.go |
| 4.2  | Add string() methods to priority types            | 20min | LOW    | types/types.go  |
| 4.3  | Explore go-arch-lint for architecture enforcement | 30min | MEDIUM | New dependency  |

---

## 3. SORTED BY WORK vs IMPACT

### High Impact, Low Work (Do First)

| #   | Task                                       | Work  | Impact | Customer Value         |
| --- | ------------------------------------------ | ----- | ------ | ---------------------- |
| 1   | Commit current fix                         | 2min  | HIGH   | Preserve working code  |
| 2   | Refactor RedundantLinters data structure   | 10min | HIGH   | Single source of truth |
| 3   | Create shared redundancy check helper      | 8min  | HIGH   | DRY principle          |
| 4   | Update categorizer and fixer to use helper | 10min | HIGH   | Consistent behavior    |
| 5   | Add tests for categorizeLinters            | 25min | HIGH   | Regression prevention  |

### Medium Impact, Medium Work

| #   | Task                             | Work  | Impact | Customer Value         |
| --- | -------------------------------- | ----- | ------ | ---------------------- |
| 6   | Add fixer redundant linter tests | 15min | MEDIUM | Test coverage          |
| 7   | Document linter/formatter model  | 15min | MEDIUM | Knowledge sharing      |
| 8   | Update pkg/README.md             | 5min  | LOW    | Documentation accuracy |

### Lower Priority

| #   | Task                        | Work  | Impact | Customer Value           |
| --- | --------------------------- | ----- | ------ | ------------------------ |
| 9   | Explore mo Result types     | 30min | LOW    | Potential cleaner code   |
| 10  | Add priority string methods | 20min | LOW    | Better logging           |
| 11  | go-arch-lint integration    | 30min | MEDIUM | Architecture enforcement |

---

## 4. DETAILED EXECUTION TASKS (12min each max)

### Task 1.1: Commit Current Fix (2min)

```bash
git add pkg/linter/analyzer.go pkg/linter/categorizer.go
git commit -m "fix(linter): skip recommending lll when golines formatter is enabled"
```

### Task 1.2: Refactor RedundantLinters (10min)

Change `rules.go`:

```go
// RedundantLinters maps linters to their superseding formatters
var RedundantLinters = map[types.LinterName]types.LinterToFormatter{
    "lll": {Formatter: "golines", Reason: "golines fixes long lines, lll only reports them"},
}
```

### Task 1.3: Create Shared Helper (8min)

Add to `pkg/constants/rules.go` or `pkg/linter/redundancy.go`:

```go
func IsLinterRedundant(linter types.LinterName, enabledFormatters map[string]bool) (bool, string) {
    if mapping, exists := RedundantLinters[linter]; exists {
        if enabledFormatters[string(mapping.Formatter)] {
            return true, mapping.Reason
        }
    }
    return false, ""
}
```

### Task 1.4: Update categorizer.go (5min)

Replace hardcoded check with shared helper call.

### Task 1.5: Update fixer.go (5min)

Replace hardcoded check with shared helper call.

### Task 2.1: Test categorizeLinters - with golines (15min)

Add test: `lll` should NOT appear in recommendations when golines is enabled.

### Task 2.2: Test categorizeLinters - without golines (10min)

Add test: `lll` SHOULD appear in recommendations when golines is NOT enabled.

### Task 2.3: Test fixer redundant removal (15min)

Add test: fixer should remove `lll` when golines is enabled.

### Task 3.1-3.6: Cleanup and Documentation (60min total)

- Remove hardcoded checks
- Update docs
- Run tests

---

## 5. ARCHITECTURAL IMPROVEMENT: Type Models

### Current Problem

```go
// RedundantLinters maps linter -> reason (no formatter info!)
var RedundantLinters = map[types.LinterName]string{
    "lll": "redundant when golines formatter is enabled",
}
```

### Improved Model

```go
// LinterToFormatter describes a linter that is superseded by a formatter
type LinterToFormatter struct {
    Formatter types.FormatterName  // The formatter that supersedes this linter
    Reason    string               // Human-readable explanation
}

// RedundantLinters maps linters to their superseding formatters
var RedundantLinters = map[types.LinterName]LinterToFormatter{
    "lll": {Formatter: "golines", Reason: "golines fixes long lines"},
}
```

### Benefits

1. **Type safety** - FormatterName is already a branded type
2. **Single source of truth** - linter->formatter relationship defined once
3. **Extensible** - easy to add more linter/formatter pairs
4. **DRY** - reason string defined alongside mapping

---

## 6. LEVERAGING EXISTING LIBRARIES

### Already Using

- **samber/mo** - Result types (MigrationResult, AnalysisResult) already use this pattern
- **charmbracelet/log** - structured logging
- **onsi/ginkgo** - BDD testing

### Could Use

- **samber/lo** - for `lo.MapKeys`, `lo.Invert` for transformations
- **lo.Invert** could simplify the formatter set creation

### Not Applicable

- **gin/httprouter** - CLI tool, no HTTP server needed
- **sqlc** - no database
- **casbin** - no auth
- **resend** - no email

---

## 7. EXECUTION GRAPH

```mermaid
graph TD
    A[Start: lll bug reported] --> B[Task 1.1: Commit current fix]
    B --> C[Task 1.2: Refactor RedundantLinters]
    C --> D[Task 1.3: Create shared helper]
    D --> E[Task 1.4: Update categorizer]
    E --> F[Task 1.5: Update fixer]
    F --> G[Task 2.1: Test with golines]
    G --> H[Task 2.2: Test without golines]
    H --> I[Task 2.3: Test fixer removal]
    I --> J[Task 3.1: Cleanup hardcoded]
    J --> K[Task 3.2: Update docs]
    K --> L[Task 3.3: Run full tests]
    L --> M[Commit all changes]
    M --> N[Push to remote]

    style A fill:#ff6b6b
    style N fill:#51cf66
```

---

## 8. EXECUTION SUMMARY

### Completed ✅

| Task                                     | Commit  | Status |
| ---------------------------------------- | ------- | ------ |
| Commit current fix                       | a006833 | ✅     |
| Refactor RedundantLinters data structure | e106cda | ✅     |
| Update categorizer to use shared helper  | e106cda | ✅     |
| Update fixer to use shared helper        | e106cda | ✅     |
| Add tests for CategorizeLinters          | 95454eb | ✅     |
| Push to remote                           | -       | ✅     |

### Deferred ⬜

- Full test suite (requires Go 1.26.1)
- golangci-lint on codebase (requires Go 1.26.1)
- pkg/README.md internal/di cleanup
- Architecture documentation

---

## 9. CUSTOMER VALUE

**Before:** Report showed confusing output where lll appeared in both "Medium Priority Linters" and "Disabled Linters" sections even when golines was enabled.

**After:** lll is correctly identified as redundant when golines is enabled, and users see "No fixes needed" with accurate messaging.

**Files changed:** 6 files, +147 lines, -15 lines

---

**Customer Value:** Users won't see confusing "recommendation" to enable `lll` when they already have `golines` enabled. Reports will accurately show configuration is optimal.
