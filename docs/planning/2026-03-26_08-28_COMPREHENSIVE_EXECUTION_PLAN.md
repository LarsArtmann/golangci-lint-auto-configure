# Comprehensive Execution Plan: golangci-lint-auto-configure UI/UX Improvements

**Created:** 2026-03-26  
**Status:** Planning Phase  
**Last Updated:** 2026-03-26

---

## Executive Summary

This plan addresses the UI/UX improvements for golangci-lint-auto-configure. The primary blocker is a corrupted Go toolchain installation caused by disk space exhaustion (100% full → 2.1GB free).

---

## Part 1: Honest Assessment of Current State

### What We Forgot / Did Wrong

1. **Disk Space Management**: Allowed disk to fill to 100%, corrupting Go toolchain
2. **Dependency on Broken Toolchain**: Can't verify code without working Go
3. **Stale Diagnostics**: LSP reports errors for files that don't exist (formatters.go, spinner.go, wizard.go were deleted)
4. **Global Variables in UI Package**: `gochecknoglobals` warnings for color constants

### Something Stupid We Do Anyway

1. **Keeping deps despite broken toolchain**: We keep adding deps (lipgloss, huh) without verifying the build works
2. **Planning more features when fundamentals are broken**: Should fix toolchain first before planning wizard/spinner

### What Could Be Better

1. **Immediate Fixes Before Adding Features**: Verify build works after each change
2. **Smaller PRs**: Large PRs with multiple file changes are harder to review and revert
3. **CI/CD Guardrails**: No automated checks for disk space or toolchain health
4. **Test Coverage**: Need to verify UI formatting functions have tests

### What Could Still Improve

1. **Type Safety**: Some areas could use more strong typing
2. **Error Messages**: User-facing errors could be more actionable
3. **Progress Feedback**: Long operations lack visual feedback

### Did We Lie to the User?

No, but we should be more explicit about the infrastructure issues blocking progress.

### How to Be Less Stupid

1. **Pre-flight Checks**: Always verify disk space before major operations
2. **Smaller Incremental Changes**: Commit working code, not work-in-progress
3. **Test Immediately**: Run tests after every change, not at the end

### Ghost Systems (Code Not Integrated)

1. **spinner.go, wizard.go, formatters.go**: These files were deleted but may have been partially integrated into the CLI commands. Need to verify.
2. **samber/mo Result types**: Well-integrated via `pkg/types/result.go`

### Scope Creep Trap

The original request was "Improve UI/UX" but expanded to include:

- Full TUI wizard implementation
- Progress spinners
- HTML report enhancements
- Multiple output formatters

**Recommended Scope Reduction**: Focus on making current styled output work reliably first.

### Did We Remove Something Useful?

The stale files (formatters.go, spinner.go, wizard.go) appeared to be incomplete implementations. No useful functionality was removed - they were causing build errors.

### Split Brains

1. **Two UI formatting approaches**:
   - `ui.FormatRecommendations()` in `pkg/ui/formatter.go`
   - `analyzer.FormatRecommendations()` in `pkg/linter/analyzer.go`

   These duplicate functionality. Should consolidate.

2. **Two summary formatting approaches**:
   - `ui.FormatSummary()` in `pkg/ui/formatter.go`
   - `analyzer.GetSummary()` in `pkg/linter/analyzer.go`

### Test Status

Tests were reported as passing (20 config tests, 19 CLI tests) before toolchain corruption. Current state unknown.

---

## Part 2: Comprehensive Multi-Step Execution Plan (Sorted by Impact/Effort)

### Phase 0: Infrastructure Recovery (BLOCKING - Must Fix First)

| Priority | Task                                   | Impact   | Effort | Customer Value      |
| -------- | -------------------------------------- | -------- | ------ | ------------------- |
| P0       | Fix Go toolchain (free disk space)     | Critical | 30min  | Unblocks all work   |
| P0       | Verify `go build ./...` works          | Critical | 10min  | Confirms recovery   |
| P0       | Run `go test ./pkg/... ./internal/...` | Critical | 15min  | Confirms tests pass |

### Phase 1: Quick Wins (High Impact, Low Effort)

| Priority | Task                                                | Impact | Effort | Customer Value        |
| -------- | --------------------------------------------------- | ------ | ------ | --------------------- |
| P1       | Remove stale LSP diagnostics (restart gopls)        | Medium | 5min   | Clean IDE             |
| P1       | Fix `gochecknoglobals` warnings in styled_output.go | Low    | 15min  | Cleaner build         |
| P1       | Verify HTML report compiles (run `templ generate`)  | Medium | 10min  | No template errors    |
| P1       | Clean up duplicate `FormatRecommendations` code     | Medium | 20min  | Less code to maintain |

### Phase 2: UI Polish (Medium Impact, Medium Effort)

| Priority | Task                                               | Impact | Effort | Customer Value   |
| -------- | -------------------------------------------------- | ------ | ------ | ---------------- |
| P2       | Add unit tests for UI formatting functions         | Medium | 30min  | Better coverage  |
| P2       | Add progress spinner for analyze operation         | Medium | 45min  | Better UX        |
| P2       | Improve error messages with actionable suggestions | Medium | 30min  | Better UX        |
| P2       | Add `--verbose` output for analysis steps          | Low    | 20min  | Better debugging |

### Phase 3: Optional Enhancements (Lower Priority)

| Priority | Task                             | Impact | Effort | Customer Value |
| -------- | -------------------------------- | ------ | ------ | -------------- |
| P3       | Interactive TUI wizard using huh | Low    | 2hr    | Advanced UX    |
| P3       | Interactive linter selection UI  | Low    | 2hr    | Advanced UX    |
| P3       | Custom presets editor            | Low    | 2hr    | Advanced UX    |

---

## Part 3: Detailed Task Breakdown (Sub-tasks)

### Phase 0: Infrastructure Recovery

#### 0.1: Fix Go Toolchain

- **Task**: Free disk space (currently 2.1GB free, 100% full)
- **Actions**:
  1. `rm -rf ~/go/pkg/mod/cache`
  2. `rm -rf ~/Library/Caches/go-build`
  3. `rm -rf bin/ coverage.* report.html`
  4. Verify `df -h /` shows >10GB free
- **Verification**: `go version && go build ./...`

#### 0.2: Verify Build Works

- **Task**: Confirm `go build ./...` succeeds
- **Actions**:
  1. Run `go build ./...`
  2. If fails, check specific error
  3. Fix as needed
- **Verification**: Build succeeds without errors

#### 0.3: Verify Tests Pass

- **Task**: Confirm all tests pass
- **Actions**:
  1. Run `go test ./pkg/... ./internal/...`
  2. Review test output
  3. Fix any failing tests
- **Verification**: All tests pass (0 failures)

### Phase 1: Quick Wins

#### 1.1: Clear Stale Diagnostics

- **Task**: Restart LSP to clear cached errors for deleted files
- **Actions**:
  1. Restart gopls: `lsp_restart` tool
  2. Or: Close/reopen project in IDE
- **Verification**: Diagnostics show 0 errors for non-existent files

#### 1.2: Fix Global Variable Warnings

- **Task**: Address `gochecknoglobals` warnings in styled_output.go
- **Actions**:
  1. Review `styled_output.go` lines 102, 106, etc.
  2. Either: Add `//nolint:gochecknoglobals` comments (if intentional)
  3. Or: Refactor to use function-local or package-level const
- **Current Code** (line 8-19):
  ```go
  var (
      PrimaryColor    = lipgloss.Color("#667eea")
      CriticalColor   = lipgloss.Color("#dc3545")
      ...
  )
  ```
- **Better Approach**: Use package-level const:
  ```go
  // Color definitions - intentional constants for consistent UI styling
  const (
      primaryColor    = "#667eea"
      criticalColor   = "#dc3545"
      ...
  )
  ```
- **Verification**: `golangci-lint run ./pkg/ui/...` shows 0 warnings

#### 1.3: Verify HTML Report Template

- **Task**: Ensure templ generates Go code without errors
- **Actions**:
  1. Run `templ generate`
  2. Check for syntax errors
  3. Fix any issues in `report.templ`
- **Verification**: `templ generate` succeeds, `report_templ.go` updated

#### 1.4: Consolidate Duplicate Formatting Code

- **Task**: Remove duplicate FormatRecommendations/GetSummary
- **Actions**:
  1. Review both implementations:
     - `pkg/ui/formatter.go:FormatRecommendations()`
     - `pkg/linter/analyzer.go:FormatRecommendations()`
  2. Keep one implementation (prefer `pkg/ui/` as it's UI layer)
  3. Remove duplicate from `analyzer.go`
  4. Update callers
  5. Update tests
- **Verification**: Build passes, tests pass

### Phase 2: UI Polish

#### 2.1: Add Unit Tests for UI Functions

- **Task**: Test UI formatting functions
- **Actions**:
  1. Create `pkg/ui/formatter_test.go`
  2. Test `FormatRecommendations()` with mock data
  3. Test `FormatSummary()` with mock data
  4. Test `FormatConfigHeader()` with various paths
  5. Test edge cases (empty analysis, all enabled, etc.)
- **Verification**: `go test ./pkg/ui/...` passes

#### 2.2: Add Progress Spinner

- **Task**: Show spinner during analyze operation
- **Actions**:
  1. Review existing `bubbles/spinner` usage (if any)
  2. Add spinner to `cmd_analyze.go` during analysis
  3. Use `charmbracelet/bubbles/spinner` package
  4. Handle terminal detection gracefully
- **Code Example**:

  ```go
  import "github.com/charmbracelet/bubbles/spinner"

  s := spinner.New()
  s.Spinner = spinner.Dot
  // In command loop:
  fmt.Fprint(os.Stdout, "\n")
  s.Start()
  // ... do work ...
  s.Stop()
  ```

- **Verification**: Spinner shows during analysis >3 seconds

#### 2.3: Improve Error Messages

- **Task**: Make errors more actionable
- **Actions**:
  1. Review error messages in `pkg/errors/errors.go`
  2. Add context to errors (file paths, suggestions)
  3. Add "Next Steps" to all error types
  4. Use `errors.Join` for multiple errors
- **Current Pattern**:
  ```go
  return fmt.Errorf("failed to analyze config: %w", err)
  ```
- **Better Pattern**:
  ```go
  return fmt.Errorf("failed to analyze config %s: %w\n\nSuggestion: Check if golangci-lint is installed (run: curl -sSfL https://get.golangci.org | sh)", configPath, err)
  ```
- **Verification**: Errors include actionable suggestions

#### 2.4: Add Verbose Analysis Steps

- **Task**: Show analysis steps in verbose mode
- **Actions**:
  1. Add debug logging in `analyzer.AnalyzeConfig()`
  2. Log each step: "Checking version...", "Running linters command...", etc.
  3. Use `logger.Debugf()` for step output
- **Verification**: `--verbose` shows analysis progress

### Phase 3: Optional Enhancements

#### 3.1: Interactive TUI Wizard

- **Task**: Add interactive linter selection using huh
- **Actions**:
  1. Create `pkg/ui/wizard.go` (replacing deleted file)
  2. Use `huh.NewForm()` for multi-step selection
  3. Add `--interactive` flag to configure command
  4. Handle terminal detection
- **Note**: Only implement if there's demonstrated user need

#### 3.2: Custom Presets Editor

- **Task**: Allow users to create custom linter presets
- **Actions**:
  1. Add `preset create` subcommand
  2. Store presets in `~/.config/golangci-lint-auto-configure/presets/`
  3. Allow preset sharing via config

---

## Part 4: Architectural Recommendations

### Use Established Libraries Better

| Library                  | Current Usage         | Recommended Usage                    |
| ------------------------ | --------------------- | ------------------------------------ |
| `samber/mo`              | Used for Result types | Good - keep                          |
| `charmbracelet/lipgloss` | Basic styling         | Good - consider more complex layouts |
| `charmbracelet/bubbles`  | Not used              | Consider for spinner, progress       |
| `charmbracelet/huh`      | Not used              | Consider for wizard if needed        |
| `cockroachdb/errors`     | Not used              | Consider for error wrapping          |
| `templ`                  | Used for HTML         | Good - keep                          |

### Architecture Patterns to Follow

1. **Separation of Concerns**: UI formatting in `pkg/ui/`, business logic in `pkg/linter/`
2. **Interface-Based Design**: Already following - `LinterAnalyzer`, `LinterFixer` interfaces
3. **Railway-Oriented Programming**: Using `samber/mo` Result types correctly
4. **Single Responsibility**: Each function does one thing

### Type Model Improvements

1. **Strong Typing for Priorities**: Already done with `LinterPriority` type
2. **Consider Option Types**: Use `samber/mo` Option[T] for nullable values
3. **Consider Either Types**: For operations that can fail with multiple error types

---

## Part 5: Implementation Priority Matrix

```
                    HIGH IMPACT
                        │
                        │
    ┌───────────────────┼───────────────────┐
    │                   │                   │
LOW │   Quick Wins      │   UI Polish       │
IMPACT                   │                   │
    │   - Clear diag    │   - Unit tests    │
    │   - Fix globals   │   - Progress      │
    │   - Templ gen     │   - Error msgs    │
    │   - Consolidate   │   - Verbose mode  │
    │                   │                   │
    ├───────────────────┼───────────────────┤
    │                   │                   │
HIGH│   Infrastructure   │   OPTIONAL        │
IMPACT                   │                   │
    │   - Fix Go env    │   - TUI wizard    │
    │   - Build works   │   - Preset editor │
    │   - Tests pass    │                   │
    │                   │                   │
    └───────────────────┴───────────────────┘
                        │
                        │
                    LOW IMPACT
```

---

## Part 6: Next Immediate Steps

### 1. Fix Infrastructure (BLOCKING)

```bash
# Free disk space
rm -rf ~/go/pkg/mod/cache
rm -rf ~/Library/Caches/go-build
rm -rf bin/ coverage.* report.html

# Verify
df -h /

# If >10GB free, rebuild
go build ./...
go test ./pkg/... ./internal/...
```

### 2. Clear Diagnostics

```bash
# Restart LSP
# Or just wait for IDE to refresh
```

### 3. Create Plan Document

```bash
# This document
mkdir -p docs/planning/
# ... (already done)
```

### 4. Execute Phase 1 Tasks

1. Fix global variables in styled_output.go
2. Run `templ generate`
3. Consolidate duplicate formatting code

### 5. Execute Phase 2 Tasks

1. Add unit tests for UI
2. Add progress spinner
3. Improve error messages

---

## Appendix: File Reference

### Key Files

| File                      | Purpose                            | Status               |
| ------------------------- | ---------------------------------- | -------------------- |
| `pkg/ui/styled_output.go` | Color definitions, styled messages | Needs polish         |
| `pkg/ui/formatter.go`     | Format functions                   | Needs tests          |
| `pkg/linter/analyzer.go`  | Core analysis logic                | Duplicate formatting |
| `pkg/linter/fixer.go`     | Config fixing logic                | Working              |
| `pkg/report/report.templ` | HTML report template               | Working              |
| `pkg/types/result.go`     | Result types                       | Well-designed        |

### Files to Delete

| File           | Reason                               |
| -------------- | ------------------------------------ |
| None currently | Previous stale files already deleted |

### Files to Create

| File                       | Purpose                        |
| -------------------------- | ------------------------------ |
| `pkg/ui/formatter_test.go` | Unit tests for UI              |
| `pkg/ui/spinner.go`        | Progress indicator (if needed) |

---

## Conclusion

**Primary Recommendation**: Fix infrastructure first, then execute Phase 1 quick wins. The UI package structure is solid - focus on making it work reliably before adding new features.

**Scope Control**: Resist adding TUI wizard or preset editor until core functionality is stable and tested.

**Quality Gates**:

- [ ] Build passes without errors
- [ ] All tests pass
- [ ] No golangci-lint warnings in UI package
- [ ] HTML report generates correctly
