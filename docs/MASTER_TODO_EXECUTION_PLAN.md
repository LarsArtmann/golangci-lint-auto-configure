# MASTER TODO EXECUTION PLAN
**Generated:** 2026-03-19  
**Total Tasks:** 40+  
**Estimated Time:** 6-8 hours  
**Max Task Duration:** 12 minutes

---

## EXECUTION PRIORITY MATRIX

| Priority | Category | Count | Customer Value |
|----------|----------|-------|----------------|
| P0 | Critical/Bugs | 8 | HIGH |
| P1 | High Impact | 12 | HIGH |
| P2 | Medium Impact | 15 | MEDIUM |
| P3 | Nice to Have | 10+ | LOW |

---

## PHASE 1: CRITICAL FIXES (P0) - DO FIRST

### Task 1.1: Fix wrapcheck errors in cmd_configure.go [8min]
**File:** `internal/cli/cmd_configure.go:159`  
**Impact:** Error context for debugging  
**Work:** Add fmt.Errorf wrapping with %w verb
```go
// Change: return err
// To: return fmt.Errorf("failed to configure: %w", err)
```
**Verification:** `golangci-lint run --fast | grep wrapcheck`

### Task 1.2: Fix wrapcheck errors in cmd_migrate.go [8min]
**File:** `internal/cli/cmd_migrate.go:83`  
**Impact:** Error context for debugging  
**Work:** Wrap migration errors with context
**Verification:** wrapcheck warnings reduced

### Task 1.3: Fix wrapcheck errors in cmd_analyze.go [8min]
**Files:** Lines 31, 40  
**Impact:** Error context for debugging  
**Work:** Wrap analysis errors
**Verification:** wrapcheck warnings reduced

### Task 1.4: Fix wrapcheck errors in cmd_report.go [8min]
**Files:** Lines 32, 41, 58, 65  
**Impact:** Error context for debugging  
**Work:** Wrap report generation errors
**Verification:** wrapcheck warnings reduced

### Task 1.5: Fix wrapcheck errors in cmd_validate.go [8min]
**Files:** Lines 41, 52, 66, 82  
**Impact:** Error context for debugging  
**Work:** Wrap validation errors
**Verification:** wrapcheck warnings reduced

### Task 1.6: Fix wrapcheck errors in commands.go [5min]
**File:** `internal/cli/commands.go:91`  
**Impact:** Error context for debugging  
**Work:** Wrap root command errors
**Verification:** wrapcheck warnings reduced

### Task 1.7: Fix wrapcheck errors in client.go [10min]
**Files:** Lines 67, 81, 118, 150  
**Impact:** Error context for library users  
**Work:** Wrap client API errors
**Verification:** wrapcheck warnings reduced

### Task 1.8: Fix wrapcheck errors in workflow.go [5min]
**File:** `pkg/workflow/workflow.go:44`  
**Impact:** Error context for workflow  
**Work:** Wrap workflow execution errors
**Verification:** wrapcheck warnings reduced

---

## PHASE 2: CODE QUALITY (P1) - HIGH IMPACT

### Task 2.1: Remove dead code - Builder.config field [5min]
**File:** `pkg/workflow/workflow.go:123`  
**Impact:** Cleaner code, less confusion  
**Work:** Remove unused `config` field from Builder struct
**Verification:** Build passes, tests pass

### Task 2.2: Remove dead code - workflowBuilder parameter [5min]
**File:** `internal/cli/cmd_configure.go`  
**Impact:** Cleaner code  
**Work:** Remove unused parameter
**Verification:** Build passes

### Task 2.3: Fix unused-parameter warnings in CLI [8min]
**Files:** All `internal/cli/cmd_*.go`  
**Impact:** Cleaner code, linter compliance  
**Work:** Rename unused parameters to `_`
**Verification:** `revive` linter warnings reduced

### Task 2.4: Add user-friendly validation error formatting [12min]
**File:** `pkg/types/validation.go`  
**Impact:** Better UX for config validation failures  
**Work:** 
```go
func FormatValidationError(err error) string {
    // Convert "Key: 'Config.Version' Error:Field validation..."
    // To: "Configuration error: 'version' is required"
}
```
**Verification:** Invalid configs show clear messages

### Task 2.5: Fix inline error handling (noinlineerr) [10min]
**Files:** CLI command files  
**Impact:** Code consistency  
**Work:** Change `if err := fn(); err != nil` to separate lines
**Verification:** `noinlineerr` warnings reduced

### Task 2.6: Fix varnamelen warnings (9 instances) [10min]
**Impact:** Code readability  
**Work:** Rename short variable names (c, tc, sb, f, wf)
**Verification:** `varnamelen` warnings reduced

### Task 2.7: Fix staticcheck SA4010 in fixer.go [8min]
**File:** `pkg/linter/fixer.go:196`  
**Impact:** Bug fix - append result unused  
**Work:** Fix messages slice handling
**Verification:** `staticcheck` passes

### Task 2.8: Fix errorlint warnings (type assertions) [10min]
**File:** `pkg/types/validation.go:57,84`  
**Impact:** Proper error handling  
**Work:** Use `errors.As()` instead of type assertion
**Verification:** `errorlint` passes

### Task 2.9: Fix funcorder warning in differ.go [8min]
**File:** `pkg/diff/differ.go:114`  
**Impact:** Code organization  
**Work:** Reorder methods (exported before unexported)
**Verification:** `funcorder` passes

### Task 2.10: Fix tagliatelle warnings (JSON casing) [10min]
**Files:** `pkg/linter/analyzer.go`, `pkg/config/loader.go`, etc.  
**Impact:** JSON API consistency  
**Work:** Change `json:"Enabled"` to `json:"enabled"`
**Verification:** `tagliatelle` passes

### Task 2.11: Fix redefines-builtin-id warnings [8min]
**Files:** `pkg/diff/differ.go` (multiple)  
**Impact:** Avoid shadowing built-in `new` function  
**Work:** Rename `new` parameter to `newConfig`
**Verification:** `revive` warnings reduced

### Task 2.12: Fix nestif warning in migrate.go [12min]
**File:** `internal/cli/cmd/migrate.go`  
**Impact:** Code complexity reduction  
**Work:** Refactor nested if statements
**Verification:** `nestif` passes

---

## PHASE 3: SOURCE CODE TODOS (P1-P2)

### Task 3.1: Extract LinterList type [10min]
**File:** `pkg/config/loader.go:3` - TODO line  
**Impact:** Type consistency  
**Work:** Create `LinterList` type alias in types package
**Verification:** Type used consistently

### Task 3.2: Add TOML/JSON config support [12min]
**File:** `pkg/config/loader.go:4` - TODO line  
**Impact:** User flexibility  
**Work:** Detect format from file extension, use appropriate decoder
**Verification:** Can load `.golangci.toml` and `.golangci.json`

### Task 3.3: Use io.Reader/Writer interfaces [12min]
**File:** `pkg/config/loader.go:5` - TODO line  
**Impact:** Testability  
**Work:** Add `LoadConfigFromReader(r io.Reader)` method
**Verification:** Tests can use strings.NewReader()

### Task 3.4: Extract default config values [8min]
**File:** `pkg/config/loader.go:7` - TODO line  
**Impact:** Maintainability  
**Work:** Create constants for default timeouts, paths
**Verification:** No magic strings in code

### Task 3.5: Use time.Duration for timeouts [10min]
**File:** `pkg/types/types.go:4` - TODO line  
**Impact:** Type safety  
**Work:** Change `Timeout string` to `Timeout time.Duration`
**Verification:** Proper duration parsing

### Task 3.6: Extract duplicate linter detection [10min]
**File:** `pkg/linter/fixer.go:5` - TODO line  
**Impact:** Code organization  
**Work:** Move validation to separate function
**Verification:** Clean separation of concerns

### Task 3.7: Add rollback mechanism [12min]
**File:** `pkg/linter/fixer.go:6` - TODO line  
**Impact:** Data safety  
**Work:** Create backup before save, restore on failure
**Verification:** Failed saves restore original config

### Task 3.8: Add AST parsing for project detection [12min]
**File:** `pkg/detection/detector.go:6` - TODO line  
**Impact:** Detection accuracy  
**Work:** Use go/ast instead of string matching
**Verification:** More accurate framework detection

---

## PHASE 4: DOCUMENTATION (P2)

### Task 4.1: Create docs/PATTERNS.md [12min]
**Impact:** Developer onboarding  
**Work:** Document Result type patterns, context usage, error handling
**Verification:** New devs understand patterns

### Task 4.2: Document context propagation pattern [8min]
**File:** `AGENTS.md`  
**Impact:** Consistent patterns  
**Work:** Add section on passing context from cobra
**Verification:** Pattern documented

### Task 4.3: Update README with timeout flags [8min]
**Impact:** User documentation  
**Work:** Document --fetch-timeout, --git-timeout
**Verification:** README accurate

---

## PHASE 5: ENHANCEMENTS (P3)

### Task 5.1: Add --fetch-timeout CLI flag [10min]
**Impact:** User control  
**Work:** Add flag, use in analyzer
**Verification:** Timeout works

### Task 5.2: Add --git-timeout CLI flag [10min]
**Impact:** User control  
**Work:** Add flag, use in git operations
**Verification:** Timeout works

### Task 5.3: Fix funlen warnings (3 functions) [12min]
**Files:** `cmd_installhook.go`, `cmd_validate.go`, etc.  
**Impact:** Code readability  
**Work:** Extract helper functions
**Verification:** Functions under 80 lines

### Task 5.4: Add parallel test markers [10min]
**Files:** Test files  
**Impact:** Faster tests  
**Work:** Add `t.Parallel()` to tests
**Verification:** Tests run in parallel

### Task 5.5: Fix testpackage warnings [8min]
**Files:** `*_test.go` files  
**Impact:** Test isolation  
**Work:** Change `package foo` to `package foo_test`
**Verification:** Tests use public API only

---

## TRACKING TABLE

| Task | Status | Time | Commit |
|------|--------|------|--------|
| 1.1 | ⏳ | - | - |
| 1.2 | ⏳ | - | - |
| 1.3 | ⏳ | - | - |
| 1.4 | ⏳ | - | - |
| 1.5 | ⏳ | - | - |
| 1.6 | ⏳ | - | - |
| 1.7 | ⏳ | - | - |
| 1.8 | ⏳ | - | - |
| 2.1 | ⏳ | - | - |
| 2.2 | ⏳ | - | - |
| ... | ... | ... | ... |

**Legend:** ⏳ Pending | 🔄 In Progress | ✅ Done | ❌ Skipped

---

## DAILY EXECUTION GOAL

**Day 1:** Complete Phase 1 (P0) + Phase 2 start  
**Day 2:** Complete Phase 2 + Phase 3 start  
**Day 3:** Complete Phase 3 + Phase 4  
**Day 4:** Complete Phase 5 + Final verification

---

## VERIFICATION CHECKLIST

After EACH task:
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Specific linter check passes
- [ ] Git commit with descriptive message

After EACH phase:
- [ ] Run full test suite
- [ ] Run golangci-lint
- [ ] Update tracking table
- [ ] Git push

---

*Plan generated: 2026-03-19*  
*Total estimated time: 6-8 hours*  
*Assisted-by: Crush <crush@charm.land>*
