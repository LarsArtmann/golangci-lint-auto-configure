package cli

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
)

func newCapturingLogger() (*log.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}

	return log.NewWithOptions(buf, log.Options{Level: log.InfoLevel}), buf
}

func TestRunListPresets_NoError(t *testing.T) {
	logger, _ := newCapturingLogger()

	err := runListPresets(logger)
	if err != nil {
		t.Errorf("runListPresets() returned unexpected error: %v", err)
	}
}

func TestRunListPresets_ContainsAllPresets(t *testing.T) {
	logger, buf := newCapturingLogger()

	err := runListPresets(logger)
	if err != nil {
		t.Fatalf("runListPresets() error = %v", err)
	}

	output := buf.String()

	for name := range constants.PresetDescriptions {
		if !strings.Contains(output, name) {
			t.Errorf("runListPresets() output missing preset %q in:\n%s", name, output)
		}
	}
}

func TestRunListPresets_AlphabeticalOrder(t *testing.T) {
	logger, buf := newCapturingLogger()

	err := runListPresets(logger)
	if err != nil {
		t.Fatalf("runListPresets() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	presetLines := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "linters") || strings.Contains(trimmed, "formatters") {
			continue
		}

		presetLines = append(presetLines, trimmed)
	}

	if len(presetLines) < 2 {
		return
	}

	for i := 1; i < len(presetLines); i++ {
		if presetLines[i-1] > presetLines[i] {
			t.Errorf(
				"runListPresets() presets not in alphabetical order: %q before %q",
				presetLines[i-1],
				presetLines[i],
			)
		}
	}
}

func TestRunListPresets_ShowsLinterCount(t *testing.T) {
	logger, buf := newCapturingLogger()

	err := runListPresets(logger)
	if err != nil {
		t.Fatalf("runListPresets() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "linters") {
		t.Errorf("runListPresets() output should mention 'linters' count:\n%s", output)
	}

	if !strings.Contains(output, "formatters") {
		t.Errorf("runListPresets() output should mention 'formatters' count for format preset:\n%s", output)
	}
}
