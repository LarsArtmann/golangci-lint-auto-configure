package finding

import (
	"encoding/json"
	"errors"
	"testing"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func assertEqualSeverity(t *testing.T, name string, input any, got, expected finding.Severity) {
	t.Helper()

	if got != expected {
		t.Errorf("%s(%v) = %v, want %v", name, input, got, expected)
	}
}

func assertPositionLine(t *testing.T, findings []finding.Finding, idx, expectedLine int, msg string) {
	t.Helper()

	if findings[idx].Position.Line != expectedLine {
		t.Errorf(msg, expectedLine, findings[idx].Position.Line)
	}
}

func TestPriorityToSeverity(t *testing.T) {
	tests := []struct {
		input    types.LinterPriority
		expected finding.Severity
	}{
		{types.LinterPriorityCritical, finding.SeverityCritical},
		{types.LinterPriorityHigh, finding.SeverityError},
		{types.LinterPriorityMedium, finding.SeverityWarning},
		{types.LinterPriorityOptional, finding.SeverityInfo},
		{types.LinterPriority(99), finding.SeverityInfo},
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := PriorityToSeverity(tt.input)
			assertEqualSeverity(t, "PriorityToSeverity", tt.input, got, tt.expected)
		})
	}
}

func TestFormatterPriorityToSeverity(t *testing.T) {
	tests := []struct {
		input    types.FormatterPriority
		expected finding.Severity
	}{
		{types.FormatterPriorityHigh, finding.SeverityError},
		{types.FormatterPriorityMedium, finding.SeverityWarning},
		{types.FormatterPriorityLow, finding.SeverityInfo},
		{types.FormatterPriority(99), finding.SeverityInfo},
	}
	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := FormatterPriorityToSeverity(tt.input)
			assertEqualSeverity(t, "FormatterPriorityToSeverity", tt.input, got, tt.expected)
		})
	}
}

func TestLinterCategory(t *testing.T) {
	tests := []struct {
		name     types.LinterName
		expected finding.Category
	}{
		{"gosec", finding.CategorySecurity},
		{"govet", finding.CategoryCorrectness},
		{"errcheck", finding.CategoryCorrectness},
		{"prealloc", finding.CategoryPerformance},
		{"gocyclo", finding.CategoryComplexity},
		{"dupl", finding.CategoryDuplication},
		{"wrapcheck", finding.CategoryErrorHandling},
		{"misspell", finding.CategoryStyle},
		{"paralleltest", finding.CategoryTesting},
		{"exhaustive", finding.CategoryTypeSafety},
		{"sloglint", finding.CategoryStructure},
		{"unknown-linter", finding.CategoryConfiguration},
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			got := LinterNameToCategory(tt.name)
			if got != tt.expected {
				t.Errorf("LinterNameToCategory(%s) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestRecommendationsToFindings(t *testing.T) {
	recommendations := []types.LinterRecommendation{
		{Name: "gosec", Priority: types.LinterPriorityCritical, Reason: "Security scanner"},
		{Name: "funlen", Priority: types.LinterPriorityHigh, Reason: "Long function detector"},
		{Name: "gci", Priority: types.LinterPriorityMedium, Reason: "Import organizer"},
	}

	findings := RecommendationsToFindings(recommendations, ".golangci.yml")

	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	assertFinding(
		t,
		findings[0],
		"missing-linter",
		"gosec",
		finding.SeverityCritical,
		finding.CategorySecurity,
	)
	assertFinding(
		t,
		findings[1],
		"missing-linter",
		"funlen",
		finding.SeverityError,
		finding.CategoryComplexity,
	)
	assertFinding(
		t,
		findings[2],
		"missing-linter",
		"gci",
		finding.SeverityWarning,
		finding.CategoryStyle,
	)

	for _, f := range findings {
		if f.Position.File != ".golangci.yml" {
			t.Errorf("expected position file .golangci.yml, got %s", f.Position.File)
		}

		if f.Suggestion == "" {
			t.Errorf("expected non-empty suggestion for %v", f.Tags)
		}
	}
}

func TestRecommendationsToFindingsEmpty(t *testing.T) {
	findings := RecommendationsToFindings(nil, ".golangci.yml")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for nil input, got %d", len(findings))
	}
}

func TestFormatterRecommendationsToFindings(t *testing.T) {
	recommendations := []types.FormatterRecommendation{
		{Name: "gofumpt", Priority: types.FormatterPriorityHigh, Reason: "Stricter formatting"},
		{Name: "golines", Priority: types.FormatterPriorityHigh, Reason: "Long line fixer"},
	}

	findings := FormatterRecommendationsToFindings(recommendations, ".golangci.yml")

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	for _, f := range findings {
		assertFinding(
			t,
			f,
			"missing-formatter",
			"",
			finding.SeverityError,
			finding.CategoryStyle,
		)
	}

	if len(findings[0].Tags) == 0 || string(findings[0].Tags[0]) != "gofumpt" {
		t.Errorf("expected tag gofumpt, got %v", findings[0].Tags)
	}
}

func TestDeprecatedLintersToFindings(t *testing.T) {
	linters := []types.LinterInfo{
		{Name: "wsl", Description: "Deprecated whitespace linter"},
		{Name: "deadcode", Description: "Removed linter"},
	}

	findings := DeprecatedLintersToFindings(linters, ".golangci.yml")

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	assertFinding(
		t,
		findings[0],
		"deprecated-linter",
		"wsl",
		finding.SeverityWarning,
		finding.CategoryMigration,
	)

	if findings[0].Suggestion == "" {
		t.Error("expected non-empty suggestion for deprecated linter")
	}

	if findings[1].Suggestion == "" {
		t.Error("expected non-empty suggestion for removed linter")
	}
}

func TestValidationErrorsToFindings(t *testing.T) {
	errors := []types.ValidationError{
		{Field: "run.timeout", Message: "timeout is required", Line: 5},
		{Field: "version", Message: "must be 2"},
	}

	findings := ValidationErrorsToFindings(errors, ".golangci.yml")

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	assertFinding(
		t,
		findings[0],
		"validation-error",
		"run-timeout",
		finding.SeverityError,
		finding.CategoryConfiguration,
	)

	assertPositionLine(t, findings, 0, 5, "expected line %d, got %d")

	assertPositionLine(t, findings, 1, 0, "expected line %d for no-line error, got %d")
}

func TestErrorsToFindings(t *testing.T) {
	errors := []error{
		errors.New("file not found"),
		errors.New("permission denied"),
	}

	results := ErrorsToFindings(errors, "config.yml")

	if len(results) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(results))
	}

	for idx, found := range results {
		if found.Rule != "validation-error" {
			t.Errorf("findings[%d]: expected rule validation-error, got %s", idx, found.Rule)
		}

		if found.Severity != finding.SeverityError {
			t.Errorf("findings[%d]: expected error severity, got %v", idx, found.Severity)
		}

		if found.Category != finding.CategoryConfiguration {
			t.Errorf("findings[%d]: expected configuration category, got %v", idx, found.Category)
		}

		if found.Position.File != "config.yml" {
			t.Errorf("findings[%d]: expected file config.yml, got %s", idx, found.Position.File)
		}

		if found.Message == "" {
			t.Errorf("findings[%d]: expected non-empty message", idx)
		}
	}
}

func TestErrorsToFindingsEmpty(t *testing.T) {
	findings := ErrorsToFindings(nil, "config.yml")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for nil input, got %d", len(findings))
	}
}

func TestAnalysisToReport(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		ConfigPath: ".golangci.yml",
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "gosec", Priority: types.LinterPriorityCritical, Reason: "Security"},
		},
		FormatterRecommendations: []types.FormatterRecommendation{
			{Name: "gofumpt", Priority: types.FormatterPriorityHigh, Reason: "Formatting"},
		},
		DeprecatedLinters: []types.LinterInfo{
			{Name: "wsl", Description: "Deprecated"},
		},
		CriticalCount:    1,
		HighValueCount:   0,
		MediumValueCount: 0,
		OptionalCount:    0,
		DeprecatedCount:  1,
	}

	report := AnalysisToReport(analysis, "v0.5.0")

	if report.Tool.Name != toolName {
		t.Errorf("expected tool name %s, got %s", toolName, report.Tool.Name)
	}

	if report.Tool.Version != "v0.5.0" {
		t.Errorf("expected version v0.5.0, got %s", report.Tool.Version)
	}

	if len(report.Findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(report.Findings))
	}

	assertReportSummaryTotal(t, report, 3)

	if report.Summary.BySeverity[finding.SeverityCritical] != 1 {
		t.Errorf("expected 1 critical in summary, got %d", report.Summary.BySeverity[finding.SeverityCritical])
	}

	if report.Summary.ByCategory[finding.CategorySecurity] != 1 {
		t.Errorf("expected 1 security in summary, got %d", report.Summary.ByCategory[finding.CategorySecurity])
	}
}

func TestAnalysisToSARIF(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		ConfigPath: ".golangci.yml",
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "gosec", Priority: types.LinterPriorityCritical, Reason: "Security"},
			{Name: "funlen", Priority: types.LinterPriorityHigh, Reason: "Long functions"},
		},
	}

	sarifJSON, err := AnalysisToSARIF(analysis, "v0.5.0")
	if err != nil {
		t.Fatalf("AnalysisToSARIF failed: %v", err)
	}

	var sarif map[string]json.RawMessage

	err = json.Unmarshal(sarifJSON, &sarif)
	if err != nil {
		t.Fatalf("SARIF is not valid JSON: %v", err)
	}

	if _, ok := sarif["version"]; !ok {
		t.Error("SARIF missing 'version' field")
	}

	if _, ok := sarif["$schema"]; !ok {
		t.Error("SARIF missing '$schema' field")
	}

	if _, ok := sarif["runs"]; !ok {
		t.Error("SARIF missing 'runs' field")
	}
}

func TestAnalysisToReportEmpty(t *testing.T) {
	analysis := &types.ConfigAnalysis{
		ConfigPath: ".golangci.yml",
	}

	report := AnalysisToReport(analysis, "dev")

	assertReportSummaryTotal(t, report, 0)

	if len(report.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(report.Findings))
	}
}

func assertReportSummaryTotal(t *testing.T, report *finding.Report, expected int) {
	t.Helper()

	if report.Summary.Total != expected {
		t.Errorf("expected summary total %d, got %d", expected, report.Summary.Total)
	}
}

func assertFinding(
	t *testing.T,
	f finding.Finding,
	expectedRule string,
	expectedTag string,
	expectedSeverity finding.Severity,
	expectedCategory finding.Category,
) {
	t.Helper()

	if f.Rule != expectedRule {
		t.Errorf("expected rule %q, got %q", expectedRule, f.Rule)
	}

	if expectedTag != "" {
		found := false

		for _, tag := range f.Tags {
			if string(tag) == expectedTag {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("expected tag %q in %v", expectedTag, f.Tags)
		}
	}

	if f.Severity != expectedSeverity {
		t.Errorf("expected severity %v, got %v", expectedSeverity, f.Severity)
	}

	if f.Category != expectedCategory {
		t.Errorf("expected category %v, got %v", expectedCategory, f.Category)
	}

	if f.ToolName != toolName {
		t.Errorf("expected tool name %q, got %q", toolName, f.ToolName)
	}

	if f.Message == "" {
		t.Error("expected non-empty message")
	}

	if f.ID == "" {
		t.Error("expected non-empty ID")
	}
}
