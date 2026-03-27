package ui_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	uipkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/ui"
)

// assertFormatFixResult tests FormatFixResult with the given MigrationResult and expected substring.
func assertFormatFixResult(t *testing.T, result *types.MigrationResult, expected string) {
	t.Helper()
	output := uipkg.FormatFixResult(result)
	if !strings.Contains(output, expected) {
		t.Errorf("expected '%s' in result, got: %s", expected, output)
	}
}

func TestFormatRecommendations_AllEnabled(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{},
	}

	result := uipkg.FormatRecommendations(analysis)

	if !strings.Contains(result, "All recommended linters are already enabled") {
		t.Errorf("expected 'All recommended linters are already enabled' in result, got: %s", result)
	}
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

	if !strings.Contains(result, "Critical") {
		t.Errorf("expected 'Critical' section header in result, got: %s", result)
	}

	if !strings.Contains(result, "gosec") {
		t.Errorf("expected 'gosec' linter name in result, got: %s", result)
	}

	if !strings.Contains(result, "Security linter") {
		t.Errorf("expected linter reason in result, got: %s", result)
	}
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

	if !strings.Contains(result, "✓") {
		t.Errorf("expected '✓' for enabled linter in result, got: %s", result)
	}
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

	if !strings.Contains(result, "Summary") {
		t.Errorf("expected 'Summary' section header in result, got: %s", result)
	}

	if !strings.Contains(result, "Critical") {
		t.Errorf("expected 'Critical' label in result, got: %s", result)
	}
}

func TestFormatConfigHeader(t *testing.T) {
	result := uipkg.FormatConfigHeader("/path/to/config.yml")

	if !strings.Contains(result, "golangci-lint Configuration") {
		t.Errorf("expected 'golangci-lint Configuration' in result, got: %s", result)
	}

	if !strings.Contains(result, "/path/to/config.yml") {
		t.Errorf("expected config path in result, got: %s", result)
	}
}

func TestFormatDryRunWarning(t *testing.T) {
	result := uipkg.FormatDryRunWarning()

	if !strings.Contains(result, "DRY-RUN MODE") {
		t.Errorf("expected 'DRY-RUN MODE' in result, got: %s", result)
	}
}

func TestFormatFixResult_Success(t *testing.T) {
	result := &types.MigrationResult{
		FixesApplied: 5,
		Message:      "Applied 5 fixes",
	}
	assertFormatFixResult(t, result, "5 fixes")
}

func TestFormatFixResult_NoFixes(t *testing.T) {
	result := &types.MigrationResult{
		FixesApplied: 0,
		Message:      "No fixes needed",
	}
	assertFormatFixResult(t, result, "No fixes needed")
}

func TestFormatFixResult_Failure(t *testing.T) {
	result := &types.MigrationResult{
		Error:   errors.New("something went wrong"),
		Message: "Something went wrong",
	}

	output := uipkg.FormatFixResult(result)

	if !strings.Contains(output, "Fix failed") {
		t.Errorf("expected 'Fix failed' in result, got: %s", output)
	}
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

	if !strings.Contains(result, "✓") {
		t.Errorf("expected '✓' in result, got: %s", result)
	}

	if !strings.Contains(result, "test message") {
		t.Errorf("expected 'test message' in result, got: %s", result)
	}
}

func TestErrorMsg(t *testing.T) {
	result := uipkg.ErrorMsg("error message")

	if !strings.Contains(result, "✗") {
		t.Errorf("expected '✗' in result, got: %s", result)
	}
}

func TestWarningMsg(t *testing.T) {
	result := uipkg.WarningMsg("warning message")

	if !strings.Contains(result, "⚠") {
		t.Errorf("expected '⚠' in result, got: %s", result)
	}
}

func TestInfoMsg(t *testing.T) {
	result := uipkg.InfoMsg("info message")

	if !strings.Contains(result, "ℹ") {
		t.Errorf("expected 'ℹ' in result, got: %s", result)
	}
}

func TestCode(t *testing.T) {
	result := uipkg.Code("my-code")

	if !strings.Contains(result, "my-code") {
		t.Errorf("expected 'my-code' in result, got: %s", result)
	}
}

func TestSectionHeader(t *testing.T) {
	result := uipkg.SectionHeader("My Section")

	if !strings.Contains(result, "My Section") {
		t.Errorf("expected 'My Section' in result, got: %s", result)
	}

	if !strings.Contains(result, "━━") {
		t.Errorf("expected '━━' separator in result, got: %s", result)
	}
}
