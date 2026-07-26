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
