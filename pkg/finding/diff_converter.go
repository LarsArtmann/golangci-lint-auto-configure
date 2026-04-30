package finding

import (
	"fmt"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
)

// ChangesToFindings converts diff.Change slice to Findings.
func ChangesToFindings(changes []diff.Change, configPath string) []finding.Finding {
	result := make([]finding.Finding, 0, len(changes))

	for _, change := range changes {
		severity := changeSeverity(change.Type)
		rule := changeRule(change.Type)

		pos := finding.Position{File: configPath}
		builder := finding.NewBuilder(
			rule,
			toolName,
			change.Description,
			severity,
			pos,
		).
			WithCategory(finding.CategoryConfiguration).
			WithSuggestion(change.Description)

		if change.OldValue != "" {
			builder = builder.WithBeforeCode(change.OldValue)
		}

		if change.NewValue != "" {
			builder = builder.WithAfterCode(change.NewValue)
		}

		result = append(result, buildFinding(builder))
	}

	return result
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
	nextSteps []string,
	configPath string,
) []finding.Finding {
	if fixesApplied == 0 {
		return nil
	}

	pos := finding.Position{File: configPath}
	f := buildFinding(finding.NewBuilder(
		"config-fix",
		toolName,
		fmt.Sprintf("%s: %d fixes applied", message, fixesApplied),
		finding.SeverityInfo,
		pos,
	).
		WithCategory(finding.CategoryConfiguration).
		WithFixStrategy(finding.FixStrategyDirect))

	return []finding.Finding{f}
}
