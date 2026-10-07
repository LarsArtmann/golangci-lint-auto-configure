package linter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// stubRescueAnalyzer is a minimal types.LinterAnalyzer double that only
// reports a fixed golangci-lint build Go version.
type stubRescueAnalyzer struct {
	golangciLintGoVersion string
}

func (s stubRescueAnalyzer) AnalyzeConfig(
	_ context.Context, _ string,
) (*types.ConfigAnalysis, error) {
	return nil, errors.New("not implemented")
}

func (s stubRescueAnalyzer) FindBinary(_ context.Context) error { return nil }

func (s stubRescueAnalyzer) CheckVersion(_ context.Context) error { return nil }

func (s stubRescueAnalyzer) GetDetectedVersion() string { return "v2.13.2" }

func (s stubRescueAnalyzer) GetDetectedGoVersion() string { return s.golangciLintGoVersion }

func (s stubRescueAnalyzer) GetSummary(_ *types.ConfigAnalysis) string { return "" }

func (s stubRescueAnalyzer) GetLintersByPriority(
	recommendations []types.LinterRecommendation, _ types.LinterPriority,
) []types.LinterRecommendation {
	return recommendations
}

type rescueRecorder struct {
	actions []recordedEnforceAction
}

func (r *rescueRecorder) Record(action audit.Action, linter, reason string) {
	r.actions = append(r.actions, recordedEnforceAction{action: action, linter: linter, reason: reason})
}

func (r *rescueRecorder) hasRescuedRunGo() bool {
	for _, a := range r.actions {
		if a.action == audit.ActionRescuedRunGo {
			return true
		}
	}

	return false
}

const brokenRescueConfig = `version: "2"
run:
  timeout: 5m
  go: "1.99"
linters:
  default: standard
`

func newRescueFixer(t *testing.T, binaryGoVersion string, loader *config.Loader) *Fixer {
	t.Helper()

	return &Fixer{
		configLoader:      loader,
		analyzer:          stubRescueAnalyzer{golangciLintGoVersion: binaryGoVersion},
		logger:            log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel}),
		ledger:            &rescueRecorder{},
		goVersionProvider: func(context.Context) string { return "" },
		formatterManager:  NewFormatterManager(log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})),
		forceSettings:     false,
		showMergedRules:   false,
	}
}

func writeBrokenRescueConfig(t *testing.T) (string, *config.Loader) {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte(brokenRescueConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})

	return configPath, config.NewLoader(logger)
}

func TestRescueOverspecifiedRunGoRepairs(t *testing.T) {
	configPath, loader := writeBrokenRescueConfig(t)
	fixer := newRescueFixer(t, "1.27", loader)
	recorder := fixer.ledger.(*rescueRecorder)

	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := fixer.rescueOverspecifiedRunGo(cfg, configPath, false); err != nil {
		t.Fatalf("expected repair to succeed, got %v", err)
	}

	if cfg.Run.Go != "1.27" {
		t.Fatalf("in-memory run.go = %q, want %q", cfg.Run.Go, "1.27")
	}

	reloaded, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if reloaded.Run.Go != "1.27" {
		t.Fatalf("on-disk run.go = %q, want %q", reloaded.Run.Go, "1.27")
	}

	if !recorder.hasRescuedRunGo() {
		t.Fatal("expected an ActionRescuedRunGo audit entry")
	}
}

func TestRescueOverspecifiedRunGoDryRun(t *testing.T) {
	configPath, loader := writeBrokenRescueConfig(t)
	fixer := newRescueFixer(t, "1.27", loader)
	recorder := fixer.ledger.(*rescueRecorder)

	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	err = fixer.rescueOverspecifiedRunGo(cfg, configPath, true)
	if !errors.Is(err, apperrors.ErrRunGoNewerThanBinary) {
		t.Fatalf("expected ErrRunGoNewerThanBinary, got %v", err)
	}

	if code := errorfamily.Code(err); code != "config.run_go.newer_than_binary" {
		t.Fatalf("expected classified code config.run_go.newer_than_binary, got %q", code)
	}

	if cfg.Run.Go != "1.99" {
		t.Fatalf("dry run must not modify config, run.go = %q", cfg.Run.Go)
	}

	reloaded, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if reloaded.Run.Go != "1.99" {
		t.Fatalf("dry run must not modify file, run.go = %q", reloaded.Run.Go)
	}

	if recorder.hasRescuedRunGo() {
		t.Fatal("dry runs must not record audit entries")
	}
}

func TestRescueOverspecifiedRunGoNoOps(t *testing.T) {
	cases := []struct {
		name       string
		runGo      string
		binaryGo   string
		expectNoOp bool
	}{
		{name: "run.go equals binary", runGo: "1.27", binaryGo: "1.27", expectNoOp: true},
		{name: "run.go older than binary", runGo: "1.24", binaryGo: "1.27", expectNoOp: true},
		{name: "run.go empty", runGo: "", binaryGo: "1.27", expectNoOp: true},
		{name: "binary go version unknown", runGo: "1.99", binaryGo: "", expectNoOp: true},
		{name: "unparsable run.go left to golangci-lint", runGo: "banana", binaryGo: "1.27", expectNoOp: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			configPath, loader := writeBrokenRescueConfig(t)
			fixer := newRescueFixer(t, testCase.binaryGo, loader)
			recorder := fixer.ledger.(*rescueRecorder)

			cfg := &types.Config{}
			cfg.Run.Go = testCase.runGo

			if err := fixer.rescueOverspecifiedRunGo(cfg, configPath, false); err != nil {
				t.Fatalf("expected no-op, got error %v", err)
			}

			if cfg.Run.Go != testCase.runGo {
				t.Fatalf("run.go changed %q -> %q in a no-op case", testCase.runGo, cfg.Run.Go)
			}

			if recorder.hasRescuedRunGo() {
				t.Fatal("no-op case must not record audit entries")
			}
		})
	}
}
