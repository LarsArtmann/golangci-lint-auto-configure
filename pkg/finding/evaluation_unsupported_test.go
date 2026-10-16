package finding

import (
	"strings"
	"testing"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestEvaluationToFindingsReportsUnsupportedRecommendations(t *testing.T) {
	evaluation := &linter.Evaluation{
		ConfigPath: ".golangci.yml",
		Env: linter.Env{
			Tool: types.MustParseGolangciVersion("v2.14.0"),
		},
		Unsupported: []types.LinterName{"gohumanize"},
	}

	findings, err := EvaluationToFindings(evaluation)
	if err != nil {
		t.Fatalf("EvaluationToFindings failed: %v", err)
	}

	unsupported := findRule(findings, RuleIDUnsupportedRecommendation)
	if unsupported == nil {
		t.Fatalf("EvaluationToFindings() = %v, want one %s finding", findings, RuleIDUnsupportedRecommendation)
	}

	if unsupported.Severity != finding.SeverityWarning {
		t.Errorf("severity = %s, want warning (an actionable gap, not an FYI)", unsupported.Severity)
	}

	wantParts := []string{
		"gohumanize", "golangci-lint v2.14.0", "does not include it", "no enable entry was written",
	}

	for _, want := range wantParts {
		if !strings.Contains(unsupported.Message, want) {
			t.Errorf("message %q missing %q", unsupported.Message, want)
		}
	}

	if !strings.Contains(unsupported.Suggestion, "module plugin") {
		t.Errorf("suggestion %q must name the module-plugin adoption path", unsupported.Suggestion)
	}
}

func TestEvaluationToFindingsUnsupportedMessageWithoutToolVersion(t *testing.T) {
	evaluation := &linter.Evaluation{
		ConfigPath:  ".golangci.yml",
		Unsupported: []types.LinterName{"gohumanize"},
	}

	findings, err := EvaluationToFindings(evaluation)
	if err != nil {
		t.Fatalf("EvaluationToFindings failed: %v", err)
	}

	unsupported := findRule(findings, RuleIDUnsupportedRecommendation)
	if unsupported == nil {
		t.Fatalf("no %s finding in %v", RuleIDUnsupportedRecommendation, findings)
	}

	if !strings.Contains(unsupported.Message, "the detected golangci-lint") {
		t.Errorf("message %q must fall back to the generic tool reference", unsupported.Message)
	}
}

func TestEvaluationToFindingsWithoutUnsupportedIsEmpty(t *testing.T) {
	findings, err := EvaluationToFindings(&linter.Evaluation{ConfigPath: ".golangci.yml"})
	if err != nil {
		t.Fatalf("EvaluationToFindings failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("EvaluationToFindings() = %v, want no findings for a converged evaluation", findings)
	}
}

func findRule(findings []finding.Finding, ruleID string) *finding.Finding {
	for i := range findings {
		if findings[i].Rule == finding.RuleName(ruleID) {
			return &findings[i]
		}
	}

	return nil
}
