package linter

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestReplacementAvailable(t *testing.T) {
	tests := []struct {
		name        string
		replacement types.LinterReplacement
		version     string
		want        bool
	}{
		{
			name: "no min version always available",
			replacement: types.LinterReplacement{
				Replacement: "wsl_v5",
				Reason:      "test",
			},
			version: "v2.10.1",
			want:    true,
		},
		{
			name: "empty version allows all",
			replacement: types.LinterReplacement{
				Replacement: "gomodguard_v2",
				Reason:      "test",
				MinVersion:  "v2.12.0",
			},
			version: "",
			want:    true,
		},
		{
			name: "version meets min version",
			replacement: types.LinterReplacement{
				Replacement: "gomodguard_v2",
				Reason:      "test",
				MinVersion:  "v2.12.0",
			},
			version: "v2.12.0",
			want:    true,
		},
		{
			name: "version exceeds min version",
			replacement: types.LinterReplacement{
				Replacement: "gomodguard_v2",
				Reason:      "test",
				MinVersion:  "v2.12.0",
			},
			version: "v2.12.2",
			want:    true,
		},
		{
			name: "version below min version",
			replacement: types.LinterReplacement{
				Replacement: "gomodguard_v2",
				Reason:      "test",
				MinVersion:  "v2.12.0",
			},
			version: "v2.11.0",
			want:    false,
		},
		{
			name: "version well below min version",
			replacement: types.LinterReplacement{
				Replacement: "gomodguard_v2",
				Reason:      "test",
				MinVersion:  "v2.12.0",
			},
			version: "v2.10.1",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replacementAvailable(tt.replacement, tt.version)
			if got != tt.want {
				t.Errorf("replacementAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasDeprecatedLinters_VersionGated(t *testing.T) {
	linters := []types.LinterName{"gomodguard", "gosec"}

	if !hasDeprecatedLinters(linters, "v2.12.0") {
		t.Error("expected gomodguard to be deprecated for v2.12.0+")
	}

	if hasDeprecatedLinters(linters, "v2.10.1") {
		t.Error("expected gomodguard to NOT be deprecated for v2.10.1")
	}
}

// TestCleanupExclusionLinters_VersionGated pins the version gate of the
// exclusion-rule sweep: a name is only rewritten once its successor exists in
// the installed golangci-lint (exhaustruct needs v2.13.0), while ungated
// renames (wsl -> wsl_v5) apply at any version.
func TestCleanupExclusionLinters_VersionGated(t *testing.T) {
	cfg := &types.Config{}
	cfg.Linters.Exclusions.Rules = []types.ExclusionRuleConfig{
		{Linters: []string{"exhaustruct", "gosec"}, Text: "is missing fields"},
		{Linters: []string{"wsl"}, Path: `_test\.go`},
	}

	handler := newDeprecatedLinterHandler(testQuietLogger(), "v2.12.0")
	cleanup := handler.cleanupExclusionLinters(cfg, false)

	if cleanup.rules != 1 {
		t.Fatalf("expected exactly one rule to change (the wsl rule), got %d", cleanup.rules)
	}

	exhaustructRule := cfg.Linters.Exclusions.Rules[0]
	if got := exhaustructRule.Linters; len(got) != 2 || got[0] != "exhaustruct" || got[1] != "gosec" {
		t.Errorf("exhaustruct must stay untouched below its min version, got %v", got)
	}

	wslRule := cfg.Linters.Exclusions.Rules[1]
	if got := wslRule.Linters; len(got) != 1 || got[0] != "wsl_v5" {
		t.Errorf("ungated rename must apply at any version, got %v", got)
	}
}
