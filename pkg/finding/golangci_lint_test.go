package finding

import (
	"encoding/json"
	"testing"

	finding "github.com/larsartmann/go-finding"
)

func TestParseGolangciLintJSON(t *testing.T) {
	input := `{
		"Issues": [
			{
				"FromLinter": "gosec",
				"Text": "G101: Hardcoded credentials",
				"Pos": {"Filename": "main.go", "Line": 42, "Column": 5},
				"Severity": "warning"
			},
			{
				"FromLinter": "errcheck",
				"Text": "Error return value is not checked",
				"Pos": {"Filename": "handler.go", "Line": 10, "Column": 1},
				"Severity": "error"
			},
			{
				"FromLinter": "funlen",
				"Text": "Function is too long (45 > 30)",
				"Pos": {"Filename": "service.go", "Line": 100, "Column": 1},
				"Severity": ""
			}
		]
	}`

	findings, err := ParseGolangciLintJSON([]byte(input))
	if err != nil {
		t.Fatalf("ParseGolangciLintJSON failed: %v", err)
	}

	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	assertLintFinding(
		t,
		findings[0],
		"gosec",
		"G101: Hardcoded credentials",
		"main.go",
		42,
		5,
		finding.SeverityWarning,
		finding.CategorySecurity,
	)
	assertLintFinding(
		t,
		findings[1],
		"errcheck",
		"Error return value is not checked",
		"handler.go",
		10,
		1,
		finding.SeverityError,
		finding.CategoryCorrectness,
	)
	assertLintFinding(
		t,
		findings[2],
		"funlen",
		"Function is too long (45 > 30)",
		"service.go",
		100,
		1,
		finding.SeverityWarning,
		finding.CategoryComplexity,
	)
}

func TestParseGolangciLintJSONEmpty(t *testing.T) {
	input := `{"Issues": []}`

	findings, err := ParseGolangciLintJSON([]byte(input))
	if err != nil {
		t.Fatalf("ParseGolangciLintJSON failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty issues, got %d", len(findings))
	}
}

func TestParseGolangciLintJSONInvalid(t *testing.T) {
	_, err := ParseGolangciLintJSON([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGolangciLintSeverityToFinding(t *testing.T) {
	tests := []struct {
		input    string
		expected finding.Severity
	}{
		{"error", finding.SeverityError},
		{"warning", finding.SeverityWarning},
		{"warn", finding.SeverityWarning},
		{"info", finding.SeverityInfo},
		{"", finding.SeverityWarning},
		{"unknown", finding.SeverityWarning},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := golangciLintSeverityToFinding(tt.input)
			assertEqualSeverity(t, "golangciLintSeverityToFinding", tt.input, got, tt.expected)
		})
	}
}

func TestParseGolangciLintJSONRoundTrip(t *testing.T) {
	input := `{
		"Issues": [
			{
				"FromLinter": "govet",
				"Text": "printf: non-constant format string",
				"Pos": {"Filename": "cmd/main.go", "Line": 15, "Column": 2},
				"Severity": "warning"
			}
		]
	}`

	findings, err := ParseGolangciLintJSON([]byte(input))
	if err != nil {
		t.Fatalf("ParseGolangciLintJSON failed: %v", err)
	}

	report := finding.NewReport(finding.ToolInfo{
		Name:    "golangci-lint",
		Version: "2.10.0",
	})
	report.AddFindings(findings)
	report.ComputeSummary()

	sarif, sarifErr := report.ToSARIF()
	if sarifErr != nil {
		t.Fatalf("ToSARIF failed: %v", sarifErr)
	}

	var sarifMap map[string]json.RawMessage

	jsonErr := json.Unmarshal(sarif, &sarifMap)
	if jsonErr != nil {
		t.Fatalf("SARIF is not valid JSON: %v", jsonErr)
	}

	if _, ok := sarifMap["runs"]; !ok {
		t.Error("SARIF missing 'runs' field")
	}
}

func assertLintFinding(
	t *testing.T,
	f finding.Finding,
	linter string,
	message string,
	file string,
	line int,
	column int,
	severity finding.Severity,
	category finding.Category,
) {
	t.Helper()

	if f.Rule != finding.RuleName(linter) {
		t.Errorf("expected rule %q, got %q", linter, f.Rule)
	}

	if f.Message != message {
		t.Errorf("expected message %q, got %q", message, f.Message)
	}

	if f.Position.File != file {
		t.Errorf("expected file %q, got %q", file, f.Position.File)
	}

	if f.Position.Line != line {
		t.Errorf("expected line %d, got %d", line, f.Position.Line)
	}

	if f.Position.Column != column {
		t.Errorf("expected column %d, got %d", column, f.Position.Column)
	}

	if f.Severity != severity {
		t.Errorf("expected severity %v, got %v", severity, f.Severity)
	}

	if f.Category != category {
		t.Errorf("expected category %v, got %v", category, f.Category)
	}

	if f.ToolName != "golangci-lint" {
		t.Errorf("expected tool name golangci-lint, got %s", f.ToolName)
	}
}
