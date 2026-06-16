package finding

import (
	"fmt"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
)

// ChangesToFindings converts diff.Change slice to Findings.
func ChangesToFindings(changes []diff.Change, configPath string) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(changes))

	for _, change := range changes {
		f, err := changeFinding(change, configPath)
		if err != nil {
			return nil, err
		}

		result = append(result, f)
	}

	return result, nil
}

func changeFinding(change diff.Change, configPath string) (finding.Finding, error) {
	severity := changeSeverity(change.Type)
	rule := changeRule(change.Type)

	builder := finding.NewBuilder(
		rule,
		constants.ToolName,
		change.Description,
		severity,
		finding.Position{File: configPath},
	).
		WithCategory(finding.CategoryConfiguration).
		WithSuggestion(change.Description)

	if change.OldValue != "" {
		builder = builder.WithBeforeCode(change.OldValue)
	}

	if change.NewValue != "" {
		builder = builder.WithAfterCode(change.NewValue)
	}

	f, err := buildFinding(builder)
	if err != nil {
		return finding.Finding{}, fmt.Errorf("build finding for change %s: %w", change.Path, err)
	}

	return f, nil
}

func changeSeverity(t diff.ChangeType) finding.Severity {
	switch t {
	case diff.ChangeTypeAdded:
		return finding.SeverityInfo
	case diff.ChangeTypeRemoved:
		return finding.SeverityWarning
	case diff.ChangeTypeModified:
		return finding.SeverityInfo
	default:
		return finding.SeverityInfo
	}
}

func changeRule(t diff.ChangeType) string {
	switch t {
	case diff.ChangeTypeAdded:
		return "config-added"
	case diff.ChangeTypeRemoved:
		return "config-removed"
	case diff.ChangeTypeModified:
		return "config-modified"
	default:
		return "config-changed"
	}
}

// MigrationResultToFindings converts a MigrationResult to Findings.
func MigrationResultToFindings(
	message string,
	fixesApplied int,
	_ []string,
	configPath string,
) ([]finding.Finding, error) {
	if fixesApplied == 0 {
		return nil, nil
	}

	pos := finding.Position{File: configPath}

	found, err := configFinding(configFindingParams{
		RuleID:     RuleIDConfigFix,
		Message:    fmt.Sprintf("%s: %d fixes applied", message, fixesApplied),
		Severity:   finding.SeverityInfo,
		Position:   pos,
		Category:   finding.CategoryConfiguration,
		Tags:       []finding.Tag{},
		Suggestion: fmt.Sprintf("%d fixes applied", fixesApplied),
	})
	if err != nil {
		return nil, fmt.Errorf("build migration result finding: %w", err)
	}

	return []finding.Finding{found}, nil
}
