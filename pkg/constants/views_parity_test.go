package constants_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// This file pins the legacy linter-management maps as byte-level golden JSON.
// The declarative policy table (linter_policy_table.go) must reproduce these
// maps EXACTLY through its compat views: the goldens are the parity contract
// between the old hand-maintained maps and the table, captured before the
// table existed. Regenerate with UPDATE_VIEWS=1 only when a curated-data
// change is intentional — then the table changes too, never only one side.

const viewsGoldenDir = "testdata/views-golden"

func TestLegacyViewsGolden(t *testing.T) {
	update := os.Getenv("UPDATE_VIEWS") == "1"

	type views struct {
		name string
		// any is the legacy map serialized to its canonical JSON form
		any func() any
	}

	all := []views{
		{"linter_priorities", func() any { return canonicalJSONMap(t, constants.LinterPriorities) }},
		{"linter_reasons", func() any { return canonicalJSONMap(t, constants.LinterReasons) }},
		{"deprecated_linters", func() any { return canonicalJSONMap(t, constants.DeprecatedLinters) }},
		{"disabled_linters", func() any { return canonicalJSONMap(t, constants.DisabledLinters) }},
		{"never_auto_enable_linters", func() any { return canonicalJSONMap(t, constants.NeverAutoEnableLinters) }},
		{"pragmatic_noise_linters", func() any { return canonicalJSONMap(t, constants.PragmaticNoiseLinters) }},
		{"linter_min_versions", func() any { return canonicalJSONMap(t, constants.LinterMinVersions) }},
		{"project_specific_linters", func() any { return canonicalJSONMap(t, constants.ProjectSpecificLinters) }},
		{"default_linter_settings", func() any { return canonicalJSONMap(t, linterSettingsToMap(t, constants.DefaultLinterSettings)) }},
		{"formatter_priorities", func() any { return canonicalJSONMap(t, constants.FormatterPriorities) }},
		{"formatter_reasons", func() any { return canonicalJSONMap(t, constants.FormatterReasons) }},
		{"formatter_info", func() any { return canonicalJSONMap(t, constants.FormatterInfo) }},
		{"default_formatter_settings", func() any { return canonicalJSONMap(t, formatterSettingsToMap(t, constants.DefaultFormatterSettings)) }},
	}

	for _, view := range all {
		t.Run(view.name, func(t *testing.T) {
			payload, err := json.MarshalIndent(view.any(), "", "  ")
			if err != nil {
				t.Fatalf("marshal %s: %v", view.name, err)
			}

			payload = append(payload, '\n')

			goldenPath := filepath.Join(viewsGoldenDir, view.name+".json")

			if update {
				if err := os.MkdirAll(viewsGoldenDir, 0o755); err != nil {
					t.Fatalf("mkdir golden dir: %v", err)
				}

				if err := os.WriteFile(goldenPath, payload, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}

				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s (regenerate with UPDATE_VIEWS=1): %v", goldenPath, err)
			}

			if string(want) != string(payload) {
				t.Errorf("legacy map %s drifted from golden\n--- want ---\n%s\n--- got ---\n%s",
					view.name, want, payload)
			}
		})
	}
}

// canonicalJSONMap wraps each value through JSON round-trips so both sides of
// the comparison are plain, key-sorted JSON values.
func canonicalJSONMap[K ~string, V any](t *testing.T, m map[K]V) map[string]any {
	t.Helper()

	out := make(map[string]any, len(m))
	for key, value := range m {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal value for %s: %v", key, err)
		}

		var generic any
		if err := json.Unmarshal(raw, &generic); err != nil {
			t.Fatalf("unmarshal value for %s: %v", key, err)
		}

		out[string(key)] = generic
	}

	return out
}

func linterSettingsToMap(t *testing.T, m map[types.LinterName]constants.SettingsConverter) map[string]any {
	t.Helper()

	out := make(map[string]any, len(m))
	for name, settings := range m {
		out[string(name)] = settings.ToMap()
	}

	return out
}

func formatterSettingsToMap(t *testing.T, m map[types.FormatterName]constants.SettingsConverter) map[string]any {
	t.Helper()

	out := make(map[string]any, len(m))
	for name, settings := range m {
		out[string(name)] = settings.ToMap()
	}

	return out
}

