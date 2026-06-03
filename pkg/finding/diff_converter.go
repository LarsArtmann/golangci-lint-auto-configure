package finding

import (
	"fmt"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
)

// ChangesToFindings converts diff.Change slice to Findings.
func ChangesToFindings(changes []diff.Change, configPath string) ([]finding.Finding, error) {
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

		f, err := buildFinding(builder)
		if err != nil {
			return nil, fmt.Errorf("build finding for change %s: %w", change.Path, err)
		}

		result = append(result, f)
	}

	return result, nil
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

	found, err := buildFinding(finding.NewBuilder(
		"config-fix",
		toolName,
		fmt.Sprintf("%s: %d fixes applied", message, fixesApplied),
		finding.SeverityInfo,
		pos,
	).
		WithCategory(finding.CategoryConfiguration).
		WithFixStrategy(finding.FixStrategySuggest).
		WithSuggestion(fmt.Sprintf("%d fixes applied", fixesApplied)))
	if err != nil {
		return nil, fmt.Errorf("build migration result finding: %w", err)
	}

	return []finding.Finding{found}, nil
}
