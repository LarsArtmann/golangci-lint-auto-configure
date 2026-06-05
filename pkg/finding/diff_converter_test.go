package finding

import (
	"testing"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
)

func TestChangeSeverity(t *testing.T) {
	tests := []struct {
		input    diff.ChangeType
		expected finding.Severity
	}{
		{diff.ChangeTypeAdded, finding.SeverityInfo},
		{diff.ChangeTypeRemoved, finding.SeverityWarning},
		{diff.ChangeTypeModified, finding.SeverityInfo},
		{diff.ChangeType(99), finding.SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := changeSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("changeSeverity(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestChangeRule(t *testing.T) {
	tests := []struct {
		input    diff.ChangeType
		expected string
	}{
		{diff.ChangeTypeAdded, "config-added"},
		{diff.ChangeTypeRemoved, "config-removed"},
		{diff.ChangeTypeModified, "config-modified"},
		{diff.ChangeType(99), "config-changed"},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := changeRule(tt.input)
			if got != tt.expected {
				t.Errorf("changeRule(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestChangesToFindings(t *testing.T) {
	changes := []diff.Change{
		{Type: diff.ChangeTypeAdded, Path: "linters.enable", Description: "Added gosec"},
		{Type: diff.ChangeTypeRemoved, Path: "linters.disable", Description: "Removed typecheck"},
		{
			Type:        diff.ChangeTypeModified,
			Path:        "run.timeout",
			OldValue:    "5m",
			NewValue:    "10m",
			Description: "Changed timeout",
		},
	}

	findings, err := ChangesToFindings(changes, ".golangci.yml")
	if err != nil {
		t.Fatalf("ChangesToFindings failed: %v", err)
	}

	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	assertFinding(
		t,
		findings[0],
		"config-added",
		"",
		finding.SeverityInfo,
		finding.CategoryConfiguration,
	)

	assertFinding(
		t,
		findings[1],
		"config-removed",
		"",
		finding.SeverityWarning,
		finding.CategoryConfiguration,
	)

	assertFinding(
		t,
		findings[2],
		"config-modified",
		"",
		finding.SeverityInfo,
		finding.CategoryConfiguration,
	)

	for _, f := range findings {
		if f.Position.File != ".golangci.yml" {
			t.Errorf("expected file .golangci.yml, got %s", f.Position.File)
		}
	}
}

func TestChangesToFindingsEmpty(t *testing.T) {
	findings, err := ChangesToFindings(nil, ".golangci.yml")
	if err != nil {
		t.Fatalf("ChangesToFindings failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for nil input, got %d", len(findings))
	}
}

func TestMigrationResultToFindings(t *testing.T) {
	findings, err := MigrationResultToFindings("Migration completed", 3, nil, ".golangci.yml")
	if err != nil {
		t.Fatalf("MigrationResultToFindings failed: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	assertFinding(
		t,
		findings[0],
		RuleIDConfigFix,
		"",
		finding.SeverityInfo,
		finding.CategoryConfiguration,
	)

	if findings[0].Position.File != ".golangci.yml" {
		t.Errorf("expected file .golangci.yml, got %s", findings[0].Position.File)
	}
}

func TestMigrationResultToFindingsZeroFixes(t *testing.T) {
	findings, err := MigrationResultToFindings("No changes needed", 0, nil, ".golangci.yml")
	if err != nil {
		t.Fatalf("MigrationResultToFindings failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for zero fixes, got %d", len(findings))
	}
}
