package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func silentLogger() *log.Logger {
	return log.NewWithOptions(nil, log.Options{Level: log.FatalLevel})
}

func TestEffectiveDryRunForCheckDiff(t *testing.T) {
	cfg := &types.Config{Version: "2"}

	tests := []struct {
		name        string
		isDryRun    bool
		check       bool
		showDiff    bool
		originalCfg *types.Config
		want        bool
	}{
		{
			name:        "check+diff+cfg forces non-dry-run",
			isDryRun:    true,
			check:       true,
			showDiff:    true,
			originalCfg: cfg,
			want:        false,
		},
		{
			name:        "check without diff stays dry-run",
			isDryRun:    true,
			check:       true,
			showDiff:    false,
			originalCfg: cfg,
			want:        true,
		},
		{
			name:        "check+diff but nil cfg stays dry-run",
			isDryRun:    true,
			check:       true,
			showDiff:    true,
			originalCfg: nil,
			want:        true,
		},
		{
			name:        "no check, no diff, isDryRun true",
			isDryRun:    true,
			check:       false,
			showDiff:    false,
			originalCfg: nil,
			want:        true,
		},
		{
			name:        "no check, no diff, isDryRun false",
			isDryRun:    false,
			check:       false,
			showDiff:    false,
			originalCfg: nil,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveDryRunForCheckDiff(tt.isDryRun, tt.check, tt.showDiff, tt.originalCfg)
			if got != tt.want {
				t.Errorf("effectiveDryRunForCheckDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolvePresetNoDetect(t *testing.T) {
	logger := silentLogger()

	t.Run("returns preset as-is when detect is false", func(t *testing.T) {
		got := resolvePreset("strict", false, logger)
		if got != "strict" {
			t.Errorf("expected 'strict', got %q", got)
		}
	})

	t.Run("returns empty when preset empty and no detect", func(t *testing.T) {
		got := resolvePreset("", false, logger)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

func TestApplyCheckDiffNoOp(t *testing.T) {
	logger := silentLogger()

	t.Run("shouldDiff false is no-op", func(t *testing.T) {
		applyCheckDiff(false, false, nil, nil, "", logger)
	})

	t.Run("nil originalCfg is no-op", func(t *testing.T) {
		applyCheckDiff(true, false, nil, nil, "", logger)
	})
}

func TestRunFmtUnlessDryDryRun(t *testing.T) {
	t.Run("dry-run does nothing", func(t *testing.T) {
		runFmtUnlessDry(context.Background(), silentLogger(), nil, "", true)
	})
}

func TestCloneConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(silentLogger())

	clone := cloneConfig(loader, configPath, silentLogger())
	if clone == nil {
		t.Fatal("expected non-nil config")
	}

	if clone.Version != "2" {
		t.Errorf("expected version '2', got %q", clone.Version)
	}

	clone.Version = "modified"

	loadedAgain := cloneConfig(loader, configPath, silentLogger())
	if loadedAgain.Version != "2" {
		t.Error("cloneConfig does not return independent copies")
	}
}

func TestLogNextStepsEmptyNoOp(t *testing.T) {
	logNextSteps(silentLogger(), nil)
	logNextSteps(silentLogger(), []string{})
}

func TestLogNextStepsWithSteps(t *testing.T) {
	logNextSteps(silentLogger(), []string{"step 1", "step 2"})
}

func TestRestoreOriginalConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	loader := config.NewLoader(silentLogger())

	original := cloneConfig(loader, configPath, silentLogger())
	original.Version = "modified-v2"

	restoreOriginalConfig(loader, original, configPath, silentLogger())

	restored := cloneConfig(loader, configPath, silentLogger())
	if restored.Version != "modified-v2" {
		t.Errorf("expected restored version 'modified-v2', got %q", restored.Version)
	}
}
