package linter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// stubCapabilityAnalyzer is a minimal types.LinterAnalyzer double that can
// additionally carry the probe surface, so the duck-typed env collection can
// be exercised without a real golangci-lint binary.
type stubCapabilityAnalyzer struct {
	analyzeErr      error
	probeKnown      map[types.LinterName]bool
	probeErr        error
	probeDefaults   []types.LinterName
	implementsProbe bool
}

func (s stubCapabilityAnalyzer) AnalyzeConfig(
	_ context.Context, _ string,
) (*types.ConfigAnalysis, error) {
	if s.analyzeErr != nil {
		return nil, s.analyzeErr
	}

	return &types.ConfigAnalysis{
		EnabledLinters: []types.LinterInfo{{Name: "gosec"}},
	}, nil
}

func (s stubCapabilityAnalyzer) FindBinary(_ context.Context) error { return nil }

func (s stubCapabilityAnalyzer) CheckVersion(_ context.Context) error { return nil }

func (s stubCapabilityAnalyzer) GetDetectedVersion() string { return "v2.14.0" }

func (s stubCapabilityAnalyzer) GetDetectedGoVersion() string { return "1.27" }

func (s stubCapabilityAnalyzer) DetectedBuildGoVersion() types.GoVersion { return types.GoVersion{} }

func (s stubCapabilityAnalyzer) GetSummary(_ *types.ConfigAnalysis) string { return "" }

func (s stubCapabilityAnalyzer) GetLintersByPriority(
	recommendations []types.LinterRecommendation, _ types.LinterPriority,
) []types.LinterRecommendation {
	return recommendations
}

func (s stubCapabilityAnalyzer) ProbeKnownLinters(
	_ context.Context,
) (map[types.LinterName]bool, error) {
	if s.probeErr != nil {
		return nil, s.probeErr
	}

	return s.probeKnown, nil
}

func (s stubCapabilityAnalyzer) DefaultEnabledLinters(_ context.Context) []types.LinterName {
	return s.probeDefaults
}

// The interface assertion doubles as the compile-time check that the stub
// satisfies the analyzer contract used by the env collector.
var _ types.LinterAnalyzer = stubCapabilityAnalyzer{}

func TestCollectEnvBasePopulatesBinaryKnown(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	probing := stubCapabilityAnalyzer{
		probeKnown: map[types.LinterName]bool{"errcheck": true, "govet": true},
		probeDefaults: []types.LinterName{
			"errcheck",
		},
		implementsProbe: true,
	}

	env, err := CollectEnvBase(context.Background(), probing, nil, configPath)
	if err != nil {
		t.Fatalf("CollectEnvBase failed: %v", err)
	}

	if !env.BinaryKnown["errcheck"] || env.BinaryKnown["gosec"] {
		t.Errorf("BinaryKnown = %v, want the probe vocabulary", env.BinaryKnown)
	}
}

func TestCollectEnvBaseFailsOpenWithoutProbe(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	plain := stubCapabilityAnalyzer{implementsProbe: false}

	env, err := CollectEnvBase(context.Background(), plain, nil, configPath)
	if err != nil {
		t.Fatalf("CollectEnvBase failed: %v", err)
	}

	if env.BinaryKnown != nil {
		t.Errorf("BinaryKnown = %v, want nil for an analyzer without the probe (fail open)", env.BinaryKnown)
	}
}

func TestCollectEnvBaseFailsOpenOnProbeError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	failing := stubCapabilityAnalyzer{
		probeErr:        errors.New("binary exploded"),
		implementsProbe: true,
	}

	env, err := CollectEnvBase(context.Background(), failing, nil, configPath)
	if err != nil {
		t.Fatalf("CollectEnvBase failed: %v", err)
	}

	if env.BinaryKnown != nil {
		t.Errorf("BinaryKnown = %v, want nil on probe failure (fail open)", env.BinaryKnown)
	}
}

func TestCollectAnalyzerEnabledFallsBackToProbeDefaults(t *testing.T) {
	unanalyzable := stubCapabilityAnalyzer{
		analyzeErr: errors.New("unknown linters: 'gohumanize'"),
		probeKnown: map[types.LinterName]bool{"errcheck": true},
		probeDefaults: []types.LinterName{
			"errcheck",
			"govet",
		},
		implementsProbe: true,
	}

	enabled := collectAnalyzerEnabled(
		context.Background(), unanalyzable, filepath.Join(t.TempDir(), ".golangci.yml"))

	if !enabled["errcheck"] || !enabled["govet"] || enabled["gosec"] {
		t.Errorf("AnalyzerEnabled = %v, want the probe's default-enabled set", enabled)
	}
}

func TestCollectAnalyzerEnabledPrefersAnalysis(t *testing.T) {
	analyzable := stubCapabilityAnalyzer{
		probeDefaults: []types.LinterName{"errcheck"},
	}

	enabled := collectAnalyzerEnabled(
		context.Background(), analyzable, filepath.Join(t.TempDir(), ".golangci.yml"))

	if !enabled["gosec"] || enabled["errcheck"] {
		t.Errorf("AnalyzerEnabled = %v, want the analysis set (gosec), not the defaults", enabled)
	}
}
