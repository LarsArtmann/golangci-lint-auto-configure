package linter

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
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
				"golines": map[string]any{"max-len": 99},
			},
		},
	}

	injected := injectDefaultFormatterSettings(cfg, []types.FormatterName{"golines"}, true)
	if injected != 1 {
		t.Fatalf("expected 1 forced injection, got %d", injected)
	}

	settings, _ := types.AsSettingsMap(cfg.Formatters.Settings["golines"])

	maxLen, ok := settings["max-len"]
	if !ok {
		t.Fatal("expected max-len to exist after force")
	}

	if maxLen == 99 {
		t.Error("expected forced overwrite to replace stale value 99 with default 120")
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

func TestUpdateGeneratedExclusions_Idempotent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	writeFixerFixture := func(relPath, content string) {
		t.Helper()

		fullPath := filepath.Join(dir, relPath)

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeFixerFixture("page_templ.go", "// Code generated by templ - DO NOT EDIT.\npackage main\n")
	writeFixerFixture(
		"graph/generated.go",
		"// Code generated by github.com/99designs/gqlgen. DO NOT EDIT.\npackage graph",
	)
	writeFixerFixture(
		"graph/models.go",
		"// Code generated by github.com/99designs/gqlgen. DO NOT EDIT.\npackage graph",
	)
	writeFixerFixture(
		"internal/gen/zz_generated.go",
		"// Code generated by mytool - DO NOT EDIT\npackage gen",
	)
	writeFixerFixture("internal/gen/handwritten.go", "package gen\n\nfunc handwritten() {}")

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
	cu := newConfigUpdater(logger, nil)

	cfg := &types.Config{}

	firstAdded := cu.updateGeneratedExclusions(cfg, configPath)
	if firstAdded == 0 {
		t.Fatal("expected exclusions to be added on the first run")
	}

	firstLinter := slices.Clone(cfg.Linters.Exclusions.Paths)
	firstFormatter := slices.Clone(cfg.Formatters.Exclusions.Paths)

	secondAdded := cu.updateGeneratedExclusions(cfg, configPath)
	if secondAdded != 0 {
		t.Fatalf(
			"expected 0 new exclusions on the second run (idempotence), got %d: linters=%v formatters=%v",
			secondAdded, cfg.Linters.Exclusions.Paths, cfg.Formatters.Exclusions.Paths,
		)
	}

	if !slices.Equal(firstLinter, cfg.Linters.Exclusions.Paths) {
		t.Errorf("linter paths changed between runs: first=%v second=%v", firstLinter, cfg.Linters.Exclusions.Paths)
	}

	if !slices.Equal(firstFormatter, cfg.Formatters.Exclusions.Paths) {
		t.Errorf(
			"formatter paths changed between runs: first=%v second=%v",
			firstFormatter, cfg.Formatters.Exclusions.Paths,
		)
	}

	// Default-injected paths (e.g. "vendor/") are deliberately unanchored
	// constants; the anchoring rule guards only scan-derived directory patterns,
	// which previously blanket-excluded hand-written neighbor directories.
	defaults := map[string]bool{}
	for _, p := range constants.DefaultLinterExclusionPaths {
		defaults[p] = true
	}

	for _, p := range constants.DefaultFormatterExclusionPaths {
		defaults[p] = true
	}

	for _, paths := range [][]string{cfg.Linters.Exclusions.Paths, cfg.Formatters.Exclusions.Paths} {
		for _, p := range paths {
			if defaults[p] {
				continue
			}

			unanchoredDirPattern := strings.HasSuffix(p, "/") && !strings.HasPrefix(p, "^")
			if unanchoredDirPattern {
				t.Errorf("unanchored directory exclusion %q: directory patterns must be ^-anchored", p)
			}
		}
	}
}

func TestPruneUnenabledLinterSettings(t *testing.T) {
	t.Run("removes settings for linters neither enabled nor disabled", func(t *testing.T) {
		cfg := &types.Config{
			Linters: types.LintersConfig{
				Enable: []types.LinterName{"goconst"},
				Settings: map[string]any{
					"goconst": map[string]any{"min-len": 4},
					"lll":     map[string]any{"line-length": 120},
				},
			},
		}

		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		pruned := pruneUnenabledLinterSettings(cfg, cfg.Linters.Enable, logger)

		if pruned != 1 {
			t.Fatalf("expected 1 pruned block, got %d", pruned)
		}
		if _, exists := cfg.Linters.Settings["lll"]; exists {
			t.Fatal("unenabled linter settings block must be pruned")
		}
		if _, exists := cfg.Linters.Settings["goconst"]; !exists {
			t.Fatal("enabled linter settings must be kept")
		}
	})

	t.Run("keeps settings for tool-level disabled linters", func(t *testing.T) {
		cfg := &types.Config{
			Linters: types.LintersConfig{
				Enable:   []types.LinterName{"goconst"},
				Settings: map[string]any{"funcorder": map[string]any{"sort-methods": true}},
			},
		}

		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		pruned := pruneUnenabledLinterSettings(cfg, cfg.Linters.Enable, logger)

		if pruned != 0 {
			t.Fatalf("tool-level disabled linter settings must be kept, pruned %d", pruned)
		}
	})

	t.Run("no-op on empty settings", func(t *testing.T) {
		cfg := &types.Config{Linters: types.LintersConfig{Enable: []types.LinterName{"goconst"}}}

		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		if pruned := pruneUnenabledLinterSettings(cfg, cfg.Linters.Enable, logger); pruned != 0 {
			t.Fatalf("expected 0 pruned, got %d", pruned)
		}
	})
}
