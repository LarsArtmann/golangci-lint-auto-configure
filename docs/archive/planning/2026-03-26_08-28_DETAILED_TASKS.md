# Detailed Tasks Breakdown

**Plan Reference:** `2026-03-26_08-28_COMPREHENSIVE_EXECUTION_PLAN.md`\
**Created:** 2026-03-26

---

## Task Summary Table

| #   | Task                         | Phase | Priority | Effort | Impact   | Status  |
| --- | ---------------------------- | ----- | -------- | ------ | -------- | ------- |
| 0.1 | Fix Go toolchain (free disk) | 0     | P0       | 30min  | Critical | Pending |
| 0.2 | Verify build works           | 0     | P0       | 10min  | Critical | Pending |
| 0.3 | Run tests                    | 0     | P0       | 15min  | Critical | Pending |
| 1.1 | Clear stale diagnostics      | 1     | P1       | 5min   | Medium   | Pending |
| 1.2 | Fix global variable warnings | 1     | P1       | 15min  | Low      | Pending |
| 1.3 | Verify templ generates       | 1     | P1       | 10min  | Medium   | Pending |
| 1.4 | Consolidate duplicate code   | 1     | P1       | 20min  | Medium   | Pending |
| 2.1 | Add UI unit tests            | 2     | P2       | 30min  | Medium   | Pending |
| 2.2 | Add progress spinner         | 2     | P2       | 45min  | Medium   | Pending |
| 2.3 | Improve error messages       | 2     | P2       | 30min  | Medium   | Pending |
| 2.4 | Add verbose mode             | 2     | P2       | 20min  | Low      | Pending |
| 3.1 | TUI wizard (optional)        | 3     | P3       | 2hr    | Low      | Future  |
| 3.2 | Preset editor (optional)     | 3     | P3       | 2hr    | Low      | Future  |

---

## Detailed Task Specifications

### Phase 0: Infrastructure Recovery

#### Task 0.1: Fix Go Toolchain

**Description**: Free disk space and restore Go toolchain functionality.

**Root Cause**: Disk was at 100% capacity, corrupting Go build cache.

**Steps**:

```bash
# 1. Clean Go caches
rm -rf ~/go/pkg/mod/cache
rm -rf ~/Library/Caches/go-build

# 2. Clean build artifacts
rm -rf bin/
rm -f coverage.*
rm -f report.html

# 3. Verify disk space
df -h /

# 4. If still <10GB, clean more:
rm -rf ~/go/pkg/mod
go clean -cache

# 5. Download modules
go mod download

# 6. Verify build
go build ./...
```

**Verification**:

- `df -h /` shows >10GB free
- `go version` works
- `go build ./...` succeeds

**Time Estimate**: 30 minutes

---

#### Task 0.2: Verify Build Works

**Description**: Confirm `go build ./...` completes without errors.

**Steps**:

```bash
go build ./...
```

**Expected Output**: No errors (warnings OK)

**If Fails**:

1. Check specific error message
2. Verify Go version: `go version` (should be 1.26+)
3. Verify module cache: `go env GOMODCACHE`
4. Try `go mod tidy` then `go mod download`

**Verification**: Build completes with exit code 0

**Time Estimate**: 10 minutes

---

#### Task 0.3: Run Tests

**Description**: Confirm all existing tests pass.

**Steps**:

```bash
go test ./pkg/... ./internal/...
```

**Expected Output**: All tests pass

```
ok      github.com/larsartmann/golangcli-linter-auto-configure/pkg/config
ok      github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter
ok      github.com/larsartmann/golangcli-linter-auto-configure/internal/cli
...
```

**If Fails**:

1. Check specific test failures
2. Run `go test -v ./pkg/...` for verbose output
3. Fix failing tests before proceeding

**Verification**: All tests pass (0 failures)

**Time Estimate**: 15 minutes

---

### Phase 1: Quick Wins

#### Task 1.1: Clear Stale Diagnostics

**Description**: Restart LSP to clear cached errors for deleted files.

**Root Cause**: LSP caches file references even after deletion.

**Steps**:

1. Use `lsp_restart` tool, OR
2. Close and reopen project in IDE, OR
3. Wait for IDE to refresh (automatic)

**Verification**: Diagnostics no longer reference formatters.go, spinner.go, wizard.go

**Time Estimate**: 5 minutes

---

#### Task 1.2: Fix Global Variable Warnings

**Description**: Address `gochecknoglobals` warnings in styled_output.go.

**Current State** (lines 8-19):

```go
//nolint:gochecknoglobals
var (
    PrimaryColor    = lipgloss.Color("#667eea")
    CriticalColor   = lipgloss.Color("#dc3545")
    HighColor       = lipgloss.Color("#fd7e14")
    MediumColor     = lipgloss.Color("#ffc107")
    OptionalColor   = lipgloss.Color("#17a2b8")
    SuccessColor    = lipgloss.Color("#28a745")
    MutedColor      = lipgloss.Color("#6c757d")
    HeadingColor    = lipgloss.Color("#ffffff")
    SubtextColor    = lipgloss.Color("#a0a0a0")
    CardBorderColor = lipgloss.Color("#4a4a6a")
)
```

**Problem**: Global variables cause linter warnings.

**Solution Option A - Keep globals with better comment**:

```go
// Colors is a map of named colors for consistent UI styling.
// These are intentionally global to allow style composition across functions.
// Using lipgloss.Color values requires runtime initialization.
var Colors = struct {
    Primary, Critical, High, Medium, Optional, Success, Muted, Heading, Subtext, CardBorder lipgloss.Color
}{
    Primary:    lipgloss.Color("#667eea"),
    Critical:   lipgloss.Color("#dc3545"),
    // ...
}
```

**Solution Option B - Use const strings, convert at use site**:

```go
// Color values as hex strings (const for compile-time safety)
const (
    colorPrimary    = "#667eea"
    colorCritical   = "#dc3545"
    // ...
)

// Helper function to convert at use site
func primaryColor() lipgloss.Color { return lipgloss.Color(colorPrimary) }
```

**Recommendation**: Option B is cleaner - lipgloss.Color() is cheap at runtime.

**Verification**: `golangci-lint run ./pkg/ui/...` shows 0 gochecknoglobals warnings

**Time Estimate**: 15 minutes

---

#### Task 1.3: Verify Templ Generates

**Description**: Ensure HTML report template compiles correctly.

**Steps**:

```bash
templ generate
```

**Expected Output**: `pkg/report/report_templ.go` regenerated without errors

**Current Potential Issue** (from report.templ line 150):

```templ
style={ fmt.Sprintf("width:%d%%", getCoveragePercent(data.Analysis)) }
```

This should work - `fmt` is imported at top of file.

**Verification**: `templ generate` succeeds, no errors

**Time Estimate**: 10 minutes

---

#### Task 1.4: Consolidate Duplicate Formatting Code

**Description**: Remove duplicate FormatRecommendations/GetSummary implementations.

**Current State**:

1. `pkg/ui/formatter.go:FormatRecommendations()` - Groups recommendations by priority, shows enabled/disabled status
2. `pkg/linter/analyzer.go:FormatRecommendations()` - Similar but different output format

**Problem**: Two sources of truth, potential for divergence.

**Decision Needed**: Which to keep?

| Option                | Pros                   | Cons                      |
| --------------------- | ---------------------- | ------------------------- |
| Keep UI version       | Separation of concerns | Must pass analysis object |
| Keep Analyzer version | Already has tests      | Mixes UI concerns         |

**Recommendation**: Keep `pkg/ui/formatter.go` version since it's in the UI layer. The Analyzer's `FormatRecommendations` should be removed or renamed to `FormatPlainRecommendations`.

**Steps**:

1. Compare both implementations
2. Decide which formatting logic to keep
3. Remove duplicate from `analyzer.go`
4. Update any callers
5. Ensure tests still pass

**Verification**: Build passes, tests pass

**Time Estimate**: 20 minutes

---

### Phase 2: UI Polish

#### Task 2.1: Add UI Unit Tests

**Description**: Add test coverage for UI formatting functions.

**File to Create**: `pkg/ui/formatter_test.go`

**Test Cases**:

```go
package ui

import (
    "testing"
    "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

func TestFormatRecommendations_AllEnabled(t *testing.T) {
    // Arrange
    analysis := &types.ConfigAnalysis{
        EnabledLinters: []types.LinterInfo{
            {Name: "gosec", Description: "Security"},
        },
        LinterRecommendations: []types.LinterRecommendation{},
    }

    // Act
    result := FormatRecommendations(analysis)

    // Assert
    // Should contain "All recommended linters are already enabled"
}

func TestFormatRecommendations_WithRecommendations(t *testing.T) {
    // Arrange
    analysis := &types.ConfigAnalysis{
        LinterRecommendations: []types.LinterRecommendation{
            {
                Name:     "gosec",
                Priority: types.LinterPriorityCritical,
                Reason:   "Security linter",
            },
        },
    }

    // Act
    result := FormatRecommendations(analysis)

    // Assert
    // Should contain "Critical" section header
    // Should contain "gosec" in code styling
}

func TestFormatSummary(t *testing.T) {
    // Test summary formatting
}

func TestFormatConfigHeader(t *testing.T) {
    // Test header formatting
}

func TestFormatDryRunWarning(t *testing.T) {
    // Test dry-run warning
}

func TestFormatFixResult_Success(t *testing.T) {
    // Test success result formatting
}

func TestFormatFixResult_NoFixes(t *testing.T) {
    // Test "no fixes needed" result
}

func TestFormatFixResult_Failure(t *testing.T) {
    // Test failure result
}
```

**Verification**: `go test ./pkg/ui/... -v` passes

**Time Estimate**: 30 minutes

---

#### Task 2.2: Add Progress Spinner

**Description**: Show spinner during long-running analysis.

**File to Modify**: `internal/cli/cmd_analyze.go`

**Current Code** (simplified):

```go
// Perform analysis
analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
if err != nil {
    return fmt.Errorf("failed to analyze config: %w", err)
}
```

**Enhanced Code**:

```go
import (
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
)

// Create spinner
s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("206"))

// Start spinner in background
done := make(chan struct{})
go func() {
    s.Start()
    <-done
}()

// Perform analysis
analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
close(done)

if err != nil {
    return fmt.Errorf("failed to analyze config: %w", err)
}
```

**Note**: Consider using bubbles' `Spinner` type with tea.Model for proper terminal handling.

**Alternative (simpler)**: Just show a simple text spinner:

```go
fmt.Fprint(os.Stdout, "Analyzing... ")
for range time.Tick(time.Second) {
    fmt.Fprint(os.Stdout, ".")
    if analysisComplete {
        fmt.Fprint(os.Stdout, " Done!\n")
        break
    }
}
```

**Verification**: Spinner visible during analysis

**Time Estimate**: 45 minutes

---

#### Task 2.3: Improve Error Messages

**Description**: Make errors more actionable with suggestions.

**File to Modify**: `pkg/errors/errors.go`

**Current Pattern**:

```go
type AnalysisError struct {
    Message string
    Path    string
    Cause   error
}
```

**Enhanced Pattern**:

```go
type AnalysisError struct {
    Message   string
    Path      string
    Cause     error
    Suggestion string  // NEW: Actionable next step
}

func (e AnalysisError) Error() string {
    msg := fmt.Sprintf("analysis error in %s: %s", e.Path, e.Message)
    if e.Cause != nil {
        msg = fmt.Sprintf("%s: %v", msg, e.Cause)
    }
    if e.Suggestion != "" {
        msg = fmt.Sprintf("%s\n\nSuggestion: %s", msg, e.Suggestion)
    }
    return msg
}
```

**Usage Example**:

```go
return AnalysisError{
    Message:   "golangci-lint not found",
    Path:       configPath,
    Cause:      err,
    Suggestion: "Install golangci-lint: curl -sSfL https://get.golangci.org | sh",
}
```

**Verification**: Errors include suggestions when printed

**Time Estimate**: 30 minutes

---

#### Task 2.4: Add Verbose Analysis Steps

**Description**: Log each analysis step in verbose mode.

**File to Modify**: `pkg/linter/analyzer.go`

**Current Code**:

```go
func (a *Analyzer) AnalyzeConfigResult(ctx context.Context, configPath string) types.AnalysisResult {
    if err := a.FindBinary(ctx); err != nil {
        return types.ErrAnalysis(err)
    }
    // ... rest of analysis
}
```

**Enhanced Code**:

```go
func (a *Analyzer) AnalyzeConfigResult(ctx context.Context, configPath string) types.AnalysisResult {
    a.logger.Debugf("Step 1/4: Finding golangci-lint binary...")
    if err := a.FindBinary(ctx); err != nil {
        return types.ErrAnalysis(err)
    }
    a.logger.Debugf("Step 2/4: Checking version...")
    if err := a.CheckVersion(ctx); err != nil {
        return types.ErrAnalysis(err)
    }
    a.logger.Debugf("Step 3/4: Running linters command...")
    // ... analysis
    a.logger.Debugf("Step 4/4: Categorizing recommendations...")
    // ... categorization
    return types.OkAnalysis(analysis)
}
```

**Verification**: `--verbose` shows step-by-step progress

**Time Estimate**: 20 minutes

---

### Phase 3: Optional Enhancements

#### Task 3.1: TUI Wizard (Optional)

**Description**: Interactive linter selection using huh.

**Status**: Consider only if demonstrated user need exists.

**Time Estimate**: 2 hours

---

#### Task 3.2: Preset Editor (Optional)

**Description**: Allow users to create custom linter presets.

**Status**: Future enhancement, not in current scope.

**Time Estimate**: 2 hours

---

## Execution Order

1. **Phase 0** (Must complete first):
   - 0.1 → 0.2 → 0.3

2. **Phase 1** (Can run in parallel):
   - 1.1 → 1.2 → 1.3 → 1.4

3. **Phase 2** (Sequential, depends on Phase 1):
   - 2.1 → 2.2 → 2.3 → 2.4

4. **Phase 3** (Future, only if needed):
   - Skip unless user requests

---

## Time Summary

| Phase     | Tasks  | Total Time          |
| --------- | ------ | ------------------- |
| Phase 0   | 3      | ~55 minutes         |
| Phase 1   | 4      | ~50 minutes         |
| Phase 2   | 4      | ~2 hours 5 minutes  |
| Phase 3   | 2      | ~4 hours (optional) |
| **Total** | **13** | **~3.5 hours**      |

---

## Customer Value Impact

| Phase   | Customer Value Delivered                  |
| ------- | ----------------------------------------- |
| Phase 0 | Unblocks all work, restores functionality |
| Phase 1 | Clean codebase, fewer warnings            |
| Phase 2 | Better UX, reliability, test coverage     |
| Phase 3 | Advanced features (low priority)          |
