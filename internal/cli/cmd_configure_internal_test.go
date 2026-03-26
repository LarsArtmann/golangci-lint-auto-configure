package cli

import (
	"context"
	"errors"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
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

func TestParsePriorityParam(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected types.LinterPriority
	}{
		{
			name:     "critical priority",
			input:    "critical",
			expected: types.LinterPriorityCritical,
		},
		{
			name:     "high priority",
			input:    "high",
			expected: types.LinterPriorityHigh,
		},
		{
			name:     "medium priority",
			input:    "medium",
			expected: types.LinterPriorityMedium,
		},
		{
			name:     "optional priority",
			input:    "optional",
			expected: types.LinterPriorityOptional,
		},
		{
			name:     "unknown priority defaults to high",
			input:    "unknown",
			expected: types.LinterPriorityHigh,
		},
		{
			name:     "empty string defaults to high",
			input:    "",
			expected: types.LinterPriorityHigh,
		},
		{
			name:     "case insensitive - CRITICAL not matched",
			input:    "CRITICAL",
			expected: types.LinterPriorityHigh, // Not matched, defaults to high
		},
		{
			name:     "case insensitive - MEDIUM not matched",
			input:    "MEDIUM",
			expected: types.LinterPriorityHigh, // Not matched, defaults to high
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePriorityParam(tt.input)
			if result != tt.expected {
				t.Errorf("parsePriorityParam(%q) = %v, want %v", tt.input, result, tt.expected)
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
		result := ParsePriorityParam(input)
		if result != expected {
			t.Errorf("parsePriorityParam(%q) = %v, want %v", input, result, expected)
		}
	}
}

func TestApplyPreset_ValidPreset(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})

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
		t.Errorf("applyPreset() enabled %d linters, want %d", len(mock.savedCfg.Linters.Enable), len(expectedLinters))
	}
}

func TestApplyPreset_DryRun(t *testing.T) {
	mock := &mockPresetConfigLoader{}
	logger := log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})

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
	logger := log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "nonexistent", false)

	if err == nil {
		t.Error("applyPreset() expected error for unknown preset, got nil")
	}
}

func TestApplyPreset_LoadError(t *testing.T) {
	mock := &mockPresetConfigLoader{
		loadErr: errors.New("failed to load config"),
	}
	logger := log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "minimal", false)

	if err == nil {
		t.Error("applyPreset() expected error when load fails, got nil")
	}
}

func TestApplyPreset_SaveError(t *testing.T) {
	mock := &mockPresetConfigLoader{
		saveErr: errors.New("failed to save config"),
	}
	logger := log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})

	err := applyPreset(context.Background(), logger, mock, "/test/config.yml", "minimal", false)

	if err == nil {
		t.Error("applyPreset() expected error when save fails, got nil")
	}
}

func TestApplyPreset_AllPresets(t *testing.T) {
	presets := []string{"minimal", "standard", "strict", "security", "performance"}

	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			mock := &mockPresetConfigLoader{}
			logger := log.NewWithOptions(&mockWriter{}, log.Options{Level: log.ErrorLevel})

			err := applyPreset(context.Background(), logger, mock, "/test/config.yml", preset, false)
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

// mockWriter implements io.Writer for charmbracelet log.
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
