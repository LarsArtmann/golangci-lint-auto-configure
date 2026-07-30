package linter

import (
	"slices"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestInjectDefaultSettings_PreservesExisting(t *testing.T) {
	cfg := &types.Config{
		Linters: types.LintersConfig{
			Enable: []types.LinterName{"gocritic"},
			Settings: map[string]any{
				"gocritic": map[string]any{
					"disabled-checks": []any{"hugeParam", "rangeValCopy"},
				},
			},
		},
	}

	injected := injectDefaultSettings(cfg, []types.LinterName{"gocritic"}, false)
	if injected != 0 {
		t.Fatalf("expected 0 injections (settings exist), got %d", injected)
	}

	settings, _ := types.AsSettingsMap(cfg.Linters.Settings["gocritic"])

	disabled, ok := settings["disabled-checks"]
	if !ok {
		t.Fatal("expected disabled-checks to be preserved")
	}

	checks := disabled.([]any)
	if len(checks) != 2 || checks[0] != "hugeParam" {
		t.Errorf("expected user settings preserved, got %v", checks)
	}
}

func TestInjectDefaultSettings_ForceOverwrites(t *testing.T) {
	cfg := &types.Config{
		Linters: types.LintersConfig{
			Enable: []types.LinterName{"gocritic"},
			Settings: map[string]any{
				"gocritic": map[string]any{
					"disabled-checks": []any{"hugeParam", "rangeValCopy"},
				},
			},
		},
	}

	injected := injectDefaultSettings(cfg, []types.LinterName{"gocritic"}, true)
	if injected != 1 {
		t.Fatalf("expected 1 injection (forced overwrite), got %d", injected)
	}

	settings, _ := types.AsSettingsMap(cfg.Linters.Settings["gocritic"])

	disabled, ok := settings["disabled-checks"]
	if !ok {
		t.Fatal("expected disabled-checks to exist after force")
	}

	checks := disabled.([]any)
	if len(checks) != 1 || checks[0] != "ifElseChain" {
		t.Errorf("expected default settings (ifElseChain), got %v", checks)
	}
}

func TestInjectDefaultSettings_DefaultInjectsMissing(t *testing.T) {
	cfg := &types.Config{
		Linters: types.LintersConfig{
			Enable: []types.LinterName{"gocritic"},
		},
	}

	injected := injectDefaultSettings(cfg, []types.LinterName{"gocritic"}, false)
	if injected != 1 {
		t.Fatalf("expected 1 injection (no existing settings), got %d", injected)
	}

	settings, _ := types.AsSettingsMap(cfg.Linters.Settings["gocritic"])
	if _, ok := settings["disabled-checks"]; !ok {
		t.Error("expected default disabled-checks to be injected")
	}
}

func TestMergeExclusionLinters(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		defaults []string
		want     []string
	}{
		{
			name:     "empty existing returns defaults",
			existing: nil,
			defaults: []string{"gosec", "errcheck"},
			want:     []string{"gosec", "errcheck"},
		},
		{
			name:     "empty defaults returns existing",
			existing: []string{"gosec"},
			defaults: nil,
			want:     []string{"gosec"},
		},
		{
			name:     "both empty returns empty",
			existing: nil,
			defaults: nil,
			want:     []string{},
		},
		{
			name:     "union with no overlap",
			existing: []string{"gosec", "errcheck"},
			defaults: []string{"wrapcheck", "funlen"},
			want:     []string{"gosec", "errcheck", "wrapcheck", "funlen"},
		},
		{
			name:     "union with overlap deduplicates",
			existing: []string{"gosec", "errcheck", "wrapcheck"},
			defaults: []string{"errcheck", "funlen"},
			want:     []string{"gosec", "errcheck", "wrapcheck", "funlen"},
		},
		{
			name:     "existing order preserved",
			existing: []string{"zzz", "aaa", "mmm"},
			defaults: []string{"bbb"},
			want:     []string{"zzz", "aaa", "mmm", "bbb"},
		},
		{
			name:     "duplicates within existing are removed",
			existing: []string{"gosec", "gosec", "errcheck"},
			defaults: []string{},
			want:     []string{"gosec", "errcheck"},
		},
		{
			name:     "duplicates within defaults are removed",
			existing: []string{"gosec"},
			defaults: []string{"errcheck", "errcheck", "funlen"},
			want:     []string{"gosec", "errcheck", "funlen"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeExclusionLinters(tt.existing, tt.defaults)
			if len(got) != len(tt.want) {
				t.Fatalf("mergeExclusionLinters() = %v (len %d), want %v (len %d)",
					got, len(got), tt.want, len(tt.want))
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("mergeExclusionLinters()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestInjectDefaultFormatterSettings_ForceOverwrites(t *testing.T) {
	cfg := &types.Config{
		Formatters: types.FormattersConfig{
			Enable: []types.FormatterName{"golines"},
			Settings: map[string]any{
				"golines": map[string]any{"tab-len": 99},
			},
		},
	}

	injected := injectDefaultFormatterSettings(cfg, []types.FormatterName{"golines"}, true)
	if injected != 1 {
		t.Fatalf("expected 1 forced injection, got %d", injected)
	}

	settings, _ := types.AsSettingsMap(cfg.Formatters.Settings["golines"])
	tabLen, ok := settings["tab-len"]
	if !ok {
		t.Fatal("expected tab-len to exist after force")
	}

	if tabLen == 99 {
		t.Error("expected forced overwrite to replace stale value 99")
	}
}

func TestInjectDefaultFormatterSettings_PreservesExisting(t *testing.T) {
	cfg := &types.Config{
		Formatters: types.FormattersConfig{
			Enable: []types.FormatterName{"golines"},
			Settings: map[string]any{
				"golines": map[string]any{"tab-len": 4},
			},
		},
	}

	injected := injectDefaultFormatterSettings(cfg, []types.FormatterName{"golines"}, false)
	if injected != 0 {
		t.Fatalf("expected 0 injections (settings exist), got %d", injected)
	}

	if !slices.Contains(cfg.Formatters.Enable, "golines") {
		t.Error("golines should still be enabled")
	}
}
