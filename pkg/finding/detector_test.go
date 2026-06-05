package finding

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestNewConfigAnalysisDetector(t *testing.T) {
	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
	analyzer := linter.NewAnalyzer(logger)

	detector := NewConfigAnalysisDetector(analyzer, ".golangci.yml", "v1.0.0")
	if detector == nil {
		t.Fatal("NewConfigAnalysisDetector returned nil")
	}

	if detector.Name() != "config-analysis" {
		t.Errorf("expected name 'config-analysis', got %q", detector.Name())
	}
}

func TestConfigAnalysisDetectorDetect(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".golangci.yml")
	configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
	analyzer := linter.NewAnalyzer(logger)
	detector := NewConfigAnalysisDetector(analyzer, configPath, "v1.0.0")

	findings, err := detector.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(findings) == 0 {
		t.Error("expected at least one finding for minimal config, got 0")
	}
}

func TestConfigAnalysisDetectorDetectWithValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".golangci.yml")
	configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`

	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
	analyzer := linter.NewAnalyzer(logger)
	detector := NewConfigAnalysisDetector(analyzer, configPath, "v1.0.0")

	validationErrors := []types.ValidationError{
		{Field: "version", Message: "invalid version"},
	}

	findings, err := detector.DetectWithValidation(context.Background(), validationErrors)
	if err != nil {
		t.Fatalf("DetectWithValidation failed: %v", err)
	}

	if len(findings) == 0 {
		t.Error("expected at least one finding, got 0")
	}
}
