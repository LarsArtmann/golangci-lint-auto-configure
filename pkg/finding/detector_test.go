package finding

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func newDeviationTestFixer(logger *log.Logger) *linter.Fixer {
	return linter.NewFixer(logger, linter.NewAnalyzer(logger), config.NewLoader(logger))
}

func TestNewDeviationDetector(t *testing.T) {
	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})

	detector := NewDeviationDetector(newDeviationTestFixer(logger), ".golangci.yml", types.LinterPriorityOptional)
	if detector == nil {
		t.Fatal("NewDeviationDetector returned nil")
	}

	if detector.Name() != "config-deviation" {
		t.Errorf("expected name 'config-deviation', got %q", detector.Name())
	}
}

func TestDeviationDetectorDetect(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".golangci.yml")
	configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
	detector := NewDeviationDetector(newDeviationTestFixer(logger), configPath, types.LinterPriorityOptional)

	findings, err := detector.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("expected at least one finding for minimal config, got 0")
	}

	for _, f := range findings {
		if f.Position.File != finding.FilePath(configPath) {
			t.Errorf("finding file = %q, want %q", f.Position.File, configPath)
		}
	}
}

func TestDeviationDetectorDetectPropagatesPlannerErrors(t *testing.T) {
	detector := NewDeviationDetector(failingPlanner{}, "missing.yml", types.LinterPriorityOptional)

	if _, err := detector.Detect(context.Background()); err == nil {
		t.Fatal("expected error from failing planner, got nil")
	}
}

type failingPlanner struct{}

func (failingPlanner) PlanDeviations(context.Context, string, types.LinterPriority) ([]linter.Change, error) {
	return nil, errors.New("planner boom")
}

func TestDeviationsToFindings(t *testing.T) {
	changes := []linter.Change{
		{
			Path: linter.PathLintersEnable, Element: "gosec", Kind: linter.ChangeAdd,
			To:     "gosec",
			Reason: "Inspects source code for security problems",
		},
		{
			Path: linter.PathLintersDisable, Element: "ireturn", Kind: linter.ChangeRemove,
			From:   "ireturn",
			Reason: "linter is enabled; a linter cannot be both enabled and disabled",
		},
		{
			Path: "run.timeout", Kind: linter.ChangeUpdate,
			From: "1s", To: "5m",
			Reason: "invalid duration",
		},
	}

	findings, err := DeviationsToFindings(changes, ".golangci.yml")
	if err != nil {
		t.Fatalf("DeviationsToFindings failed: %v", err)
	}

	if len(findings) != len(changes) {
		t.Fatalf("expected %d findings, got %d", len(changes), len(findings))
	}

	enable := findings[0]
	if enable.Rule != RuleIDDeviationLinterEnable {
		t.Errorf("enable finding rule = %q, want %q", enable.Rule, RuleIDDeviationLinterEnable)
	}

	if enable.Severity != finding.SeverityInfo {
		t.Errorf("enable finding severity = %q, want info", enable.Severity)
	}

	disable := findings[1]
	if disable.Rule != RuleIDDeviationLinterDisable {
		t.Errorf("disable finding rule = %q, want %q", disable.Rule, RuleIDDeviationLinterDisable)
	}

	if disable.Severity != finding.SeverityWarning {
		t.Errorf("disable finding severity = %q, want warning", disable.Severity)
	}

	timeout := findings[2]
	if timeout.Rule != RuleIDDeviationDefaults {
		t.Errorf("timeout finding rule = %q, want %q", timeout.Rule, RuleIDDeviationDefaults)
	}
}

func TestDeviationsToFindingsEmpty(t *testing.T) {
	findings, err := DeviationsToFindings(nil, ".golangci.yml")
	if err != nil {
		t.Fatalf("DeviationsToFindings failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty plan, got %d", len(findings))
	}
}
