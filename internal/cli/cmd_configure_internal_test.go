package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"go.yaml.in/yaml/v3"
)

// mockPresetConfigLoader is a mock implementation of presetConfigLoader for testing.
type mockPresetConfigLoader struct {
	loadFunc  func(path string) (*types.Config, error)
	saveFunc  func(config *types.Config, path string) error
	loadErr   error
	saveErr   error
	savedCfg  *types.Config
	savedPath string
}

func (m *mockPresetConfigLoader) LoadConfig(path string) (*types.Config, error) {
	if m.loadFunc != nil {
		return m.loadFunc(path)
	}

	if m.loadErr != nil {
		return nil, m.loadErr
	}

	return &types.Config{
		Version: "2",
		Linters: types.LintersConfig{
			Enable:  []string{"errcheck"},
			Disable: []string{},
		},
	}, nil
}

func (m *mockPresetConfigLoader) SaveConfig(config *types.Config, path string) error {
	m.savedCfg = config
	m.savedPath = path

	if m.saveFunc != nil {
		return m.saveFunc(config, path)
	}

	return m.saveErr
}

// newTestLogger creates a logger for tests with error-level output.
func newTestLogger() *log.Logger {
	return log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})
}

func TestParsePriorityParam(t *testing.T) {
	validTests := []struct {
		input    string
		expected types.LinterPriority
	}{
		{"critical", types.LinterPriorityCritical},
		{"high", types.LinterPriorityHigh},
		{"medium", types.LinterPriorityMedium},
		{"optional", types.LinterPriorityOptional},
	}

	for _, tt := range validTests {
		t.Run("valid/"+tt.input, func(t *testing.T) {
			result, err := ParsePriorityParam(tt.input)
			if err != nil {
				t.Fatalf("ParsePriorityParam(%q) returned unexpected error: %v", tt.input, err)
			}

			if result != tt.expected {
				t.Errorf("ParsePriorityParam(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}

	invalidTests := []string{"unknown", "", "CRITICAL", "MEDIUM"}

	for _, input := range invalidTests {
		t.Run("invalid/"+input, func(t *testing.T) {
			_, err := ParsePriorityParam(input)
			if err == nil {
				t.Errorf("ParsePriorityParam(%q) expected error, got nil", input)
			}
		})
	}
}

func TestParsePriorityParam_AllPriorities(t *testing.T) {
	priorities := map[string]types.LinterPriority{
		"critical": types.LinterPriorityCritical,
		"high":     types.LinterPriorityHigh,
		"medium":   types.LinterPriorityMedium,
		"optional": types.LinterPriorityOptional,
	}

	for input, expected := range priorities {
		result, err := ParsePriorityParam(input)
		if err != nil {
			t.Fatalf("ParsePriorityParam(%q) returned unexpected error: %v", input, err)
		}

		if result != expected {
			t.Errorf("ParsePriorityParam(%q) = %v, want %v", input, result, expected)
		}
	}
}

func TestApplyPreset_ValidPreset(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := newTestLogger()

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "minimal", false)
	if err != nil {
		t.Errorf("applyPreset() error = %v, want nil", err)
	}

	// Verify config was saved
	if mock.savedCfg == nil {
		t.Error("applyPreset() did not save config")
	}

	// Verify minimal preset has linters
	expectedLinters := constants.PresetLinters["minimal"]
	if len(mock.savedCfg.Linters.Enable) != len(expectedLinters) {
		t.Errorf(
			"applyPreset() enabled %d linters, want %d",
			len(mock.savedCfg.Linters.Enable),
			len(expectedLinters),
		)
	}
}

func TestApplyPreset_DryRun(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := newTestLogger()

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "minimal", true)
	if err != nil {
		t.Errorf("applyPreset() dry-run error = %v, want nil", err)
	}

	// Verify config was NOT saved in dry-run mode
	if mock.savedCfg != nil {
		t.Error("applyPreset() dry-run should not save config")
	}
}

func TestApplyPreset_UnknownPreset(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := newTestLogger()

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "nonexistent", false)
	if err == nil {
		t.Error("applyPreset() expected error for unknown preset, got nil")
	}
}

func TestApplyPreset_LoadError(t *testing.T) {
	mock := &mockPresetConfigLoader{
		loadErr: errors.New("failed to load config"),
	}
	logger := newTestLogger()

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "minimal", false)
	if err == nil {
		t.Error("applyPreset() expected error when load fails, got nil")
	}
}

func TestApplyPreset_SaveError(t *testing.T) {
	mock := &mockPresetConfigLoader{
		saveErr: errors.New("failed to save config"),
	}
	logger := newTestLogger()

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "minimal", false)
	if err == nil {
		t.Error("applyPreset() expected error when save fails, got nil")
	}
}

func TestApplyPreset_AllPresets(t *testing.T) {
	presets := []string{"minimal", "standard", "strict", "security", "performance", "reference", "format"}

	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			mock := &mockPresetConfigLoader{}
			logger := newTestLogger()

			err := applyPreset(
				context.Background(),
				logger,
				mock,
				"/test/config.yml",
				preset,
				false,
			)
			if err != nil {
				t.Errorf("applyPreset(%q) error = %v, want nil", preset, err)
			}

			// Verify linters were set
			if len(mock.savedCfg.Linters.Enable) == 0 {
				t.Errorf("applyPreset(%q) enabled no linters", preset)
			}
		})
	}
}

func TestApplyPreset_FormatEnablesFormatters(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := newTestLogger()

	err := applyPreset(
		context.Background(),
		logger,
		mock,
		"/test/config.yml",
		"format",
		false,
	)
	if err != nil {
		t.Fatalf("applyPreset(\"format\") error = %v, want nil", err)
	}

	if mock.savedCfg == nil {
		t.Fatal("applyPreset(\"format\") did not save config")
	}

	expectedFormatters := []string{"gci", "gofumpt", "goimports"}
	if len(mock.savedCfg.Formatters.Enable) != len(expectedFormatters) {
		t.Fatalf("expected %d formatters, got %d", len(expectedFormatters), len(mock.savedCfg.Formatters.Enable))
	}

	formatterSet := make(map[string]bool)
	for _, f := range mock.savedCfg.Formatters.Enable {
		formatterSet[f] = true
	}

	for _, expected := range expectedFormatters {
		if !formatterSet[expected] {
			t.Errorf("expected formatter %q not found in enabled formatters", expected)
		}
	}
}

// mockWriter implements io.Writer for charmbracelet log.
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestHandleCheckMode(t *testing.T) {
	t.Run("returns nil when check is false", func(t *testing.T) {
		err := handleCheckMode(false, &types.MigrationResult{FixesApplied: 5}, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("returns nil when check true but no fixes", func(t *testing.T) {
		err := handleCheckMode(true, &types.MigrationResult{FixesApplied: 0}, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("returns ErrChangesNeeded when check true and fixes applied", func(t *testing.T) {
		logger := newTestLogger()

		err := handleCheckMode(true, &types.MigrationResult{FixesApplied: 3}, logger)
		if !errors.Is(err, apperrors.ErrChangesNeeded) {
			t.Errorf("expected ErrChangesNeeded, got %v", err)
		}
	})
}

func TestCaptureOriginalConfig(t *testing.T) {
	t.Run("returns nil when shouldClone is false", func(t *testing.T) {
		result := captureOriginalConfig(false, nil, "", nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}

func TestDisplayFixResult(t *testing.T) {
	t.Run("does not panic", func(t *testing.T) {
		displayFixResult("test.yml", &types.MigrationResult{
			FixesApplied: 2,
			Message:      "test",
		})
	})
}

func TestConvertNames(t *testing.T) {
	input := []types.LinterName{"gosec", "errcheck"}
	result := convertNames(input)

	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}

	if result[0] != "gosec" {
		t.Errorf("expected gosec, got %s", result[0])
	}
}

func TestApplyPreset_FormatYAMLIntegration(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := newTestLogger()

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "format", false)
	if err != nil {
		t.Fatalf("applyPreset(\"format\") error = %v, want nil", err)
	}

	if mock.savedCfg == nil {
		t.Fatal("applyPreset(\"format\") did not save config")
	}

	data, err := yaml.Marshal(mock.savedCfg)
	if err != nil {
		t.Fatalf("failed to marshal saved config: %v", err)
	}

	yamlStr := string(data)
	for _, expected := range []string{"gci", "goimports", "gofumpt"} {
		if !strings.Contains(yamlStr, expected) {
			t.Errorf("YAML output missing formatter %q in:\n%s", expected, yamlStr)
		}
	}
}
