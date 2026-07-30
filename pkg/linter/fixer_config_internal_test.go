package linter

import (
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
