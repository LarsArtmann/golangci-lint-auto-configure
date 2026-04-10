package ui_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	uipkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/ui"
)

// assertContains verifies that result contains the expected substring.
func assertContains(t *testing.T, result, expected string) {
	t.Helper()

	if !strings.Contains(result, expected) {
		t.Errorf("expected %q in result, got: %s", expected, result)
	}
}

// assertFormatFixResult tests FormatFixResult with the given MigrationResult and expected substring.
func assertFormatFixResult(t *testing.T, result *types.MigrationResult, expected string) {
	t.Helper()

	output := uipkg.FormatFixResult(result)
	assertContains(t, output, expected)
}

// newMigrationResult creates a MigrationResult with fixes applied.
func newMigrationResult(fixes int, msg string) *types.MigrationResult {
	return &types.MigrationResult{
		FixesApplied: fixes,
		Message:      msg,
	}
}

// newMigrationErrorResult creates a MigrationResult with an error.
func newMigrationErrorResult(errMsg string) *types.MigrationResult {
	return &types.MigrationResult{
		Error:   errors.New(errMsg),
		Message: errMsg,
	}
}

func TestFormatRecommendations_AllEnabled(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{},
	}

	result := uipkg.FormatRecommendations(analysis)

	assertContains(t, result, "All recommended linters are already enabled")
}

func TestFormatRecommendations_WithCriticalLinter(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{
				Name:     "gosec",
				Priority: types.LinterPriorityCritical,
				Reason:   "Security linter",
			},
		},
		EnabledLinters: []types.LinterInfo{},
	}

	result := uipkg.FormatRecommendations(analysis)

	assertContains(t, result, "Critical")
	assertContains(t, result, "gosec")
	assertContains(t, result, "Security linter")
}

func TestFormatRecommendations_WithEnabledLinter(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{
				Name:     "gosec",
				Priority: types.LinterPriorityCritical,
				Reason:   "Security linter",
			},
		},
		EnabledLinters: []types.LinterInfo{
			{Name: "gosec"},
		},
	}

	result := uipkg.FormatRecommendations(analysis)

	assertContains(t, result, "✓")
}

func TestFormatSummary(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		CriticalCount:    2,
		HighValueCount:   3,
		MediumValueCount: 4,
		OptionalCount:    1,
		EnabledLinters:   []types.LinterInfo{{Name: "lint1"}, {Name: "lint2"}},
		DisabledLinters:  []types.LinterInfo{{Name: "lint3"}},
	}

	result := uipkg.FormatSummary(analysis)

	assertContains(t, result, "Summary")
	assertContains(t, result, "Critical")
}

func TestFormatConfigHeader(t *testing.T) {
	result := uipkg.FormatConfigHeader("/path/to/config.yml")

	assertContains(t, result, "golangci-lint Configuration")
	assertContains(t, result, "/path/to/config.yml")
}

func TestFormatDryRunWarning(t *testing.T) {
	result := uipkg.FormatDryRunWarning()

	assertContains(t, result, "DRY-RUN MODE")
}

func TestFormatFixResult_Success(t *testing.T) {
	result := newMigrationResult(5, "Applied 5 fixes")
	assertFormatFixResult(t, result, "5 fixes")
}

func TestFormatFixResult_NoFixes(t *testing.T) {
	result := newMigrationResult(0, "No fixes needed")
	assertFormatFixResult(t, result, "No fixes needed")
}

func TestFormatFixResult_Failure(t *testing.T) {
	result := newMigrationErrorResult("something went wrong")

	output := uipkg.FormatFixResult(result)
	assertContains(t, output, "Fix failed")
}

func TestPriorityBadge(t *testing.T) {
	tests := []struct {
		priority int
		expected string
	}{
		{0, "CRITICAL"},
		{1, "HIGH"},
		{2, "MEDIUM"},
		{3, "OPTIONAL"},
	}

	for _, tt := range tests {
		result := uipkg.PriorityBadge(tt.priority)

		if !strings.Contains(result, tt.expected) {
			t.Errorf("PriorityBadge(%d) = %q, want containing %q", tt.priority, result, tt.expected)
		}
	}
}

func TestSuccessMsg(t *testing.T) {
	result := uipkg.SuccessMsg("test message")

	assertContains(t, result, "✓")
	assertContains(t, result, "test message")
}

func TestErrorMsg(t *testing.T) {
	result := uipkg.ErrorMsg("error message")

	assertContains(t, result, "✗")
}

func TestWarningMsg(t *testing.T) {
	result := uipkg.WarningMsg("warning message")

	assertContains(t, result, "⚠")
}

func TestInfoMsg(t *testing.T) {
	result := uipkg.InfoMsg("info message")

	assertContains(t, result, "ℹ")
}

func TestCode(t *testing.T) {
	result := uipkg.Code("my-code")

	assertContains(t, result, "my-code")
}

func TestSectionHeader(t *testing.T) {
	result := uipkg.SectionHeader("My Section")

	assertContains(t, result, "My Section")
	assertContains(t, result, "━━")
}
