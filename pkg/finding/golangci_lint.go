package finding

import (
	"encoding/json"
	"fmt"
	"strings"

	finding "github.com/larsartmann/go-finding"
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
		return nil, fmt.Errorf("parse golangci-lint JSON: %w", err)
	}

	findings := make([]finding.Finding, 0, len(output.Issues))

	for _, issue := range output.Issues {
		f, err := issueToFinding(issue)
		if err != nil {
			return nil, fmt.Errorf("convert issue from %s: %w", issue.FromLinter, err)
		}

		findings = append(findings, f)
	}

	return findings, nil
}

func issueToFinding(issue GolangciLintIssue) (finding.Finding, error) {
	severity := golangciLintSeverityToFinding(issue.Severity)
	category := linterNameToCategory(issue.FromLinter)
	pos := finding.Position{
		File:   issue.Pos.Filename,
		Line:   issue.Pos.Line,
		Column: issue.Pos.Column,
	}

	return buildFinding(finding.NewBuilder(
		issue.FromLinter,
		"golangci-lint",
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
