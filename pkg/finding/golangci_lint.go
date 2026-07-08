package finding

import (
	"encoding/json"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
)

// GolangciLintIssue represents a single issue from golangci-lint JSON output.
type GolangciLintIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
	Severity string `json:"Severity"`
}

// golangciLintOutput represents the top-level JSON structure from golangci-lint.
type golangciLintOutput struct {
	Issues []GolangciLintIssue `json:"Issues"`
}

// ParseGolangciLintJSON converts golangci-lint JSON output to Findings.
func ParseGolangciLintJSON(data []byte) ([]finding.Finding, error) {
	var output golangciLintOutput

	err := json.Unmarshal(data, &output)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "golangci_lint.parse_json",
			"parse golangci-lint JSON")
	}

	findings := make([]finding.Finding, 0, len(output.Issues))

	for _, issue := range output.Issues {
		converted, err := issueToFinding(issue)
		if err != nil {
			return nil, errorfamily.WrapCorruptionf(err, "golangci_lint.convert_issue",
				"convert issue from %s", issue.FromLinter)
		}

		findings = append(findings, converted)
	}

	return findings, nil
}

func issueToFinding(issue GolangciLintIssue) (finding.Finding, error) {
	severity := golangciLintSeverityToFinding(issue.Severity)
	category := linterNameToCategory(issue.FromLinter)
	pos := finding.Position{
		File:   finding.FilePath(issue.Pos.Filename),
		Line:   issue.Pos.Line,
		Column: issue.Pos.Column,
	}

	return buildFinding(finding.NewBuilder(
		finding.RuleName(issue.FromLinter),
		finding.ToolName(constants.GolangciLintBinaryName),
		issue.Text,
		severity,
		pos,
	).
		WithCategory(category).
		WithTags(linterTag(issue.FromLinter)))
}

// golangciLintSeverityToFinding maps golangci-lint severity strings to finding.Severity.
func golangciLintSeverityToFinding(severity string) finding.Severity {
	switch strings.ToLower(severity) {
	case "error", "critical":
		return finding.SeverityError
	case "warning", "warn":
		return finding.SeverityWarning
	case "info", "information":
		return finding.SeverityInfo
	default:
		return finding.SeverityWarning
	}
}

// linterNameToCategory maps golangci-lint linter names to finding categories.
func linterNameToCategory(linterName string) finding.Category {
	return LinterToCategory(linterName)
}
