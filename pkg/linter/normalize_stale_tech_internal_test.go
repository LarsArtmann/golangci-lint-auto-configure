package linter

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestRemoveStaleTechGatedLinters(t *testing.T) {
	tests := []struct {
		name        string
		enable      []types.LinterName
		disable     []types.LinterName
		tech        map[string]bool
		wantEnable  []types.LinterName
		wantDisable []types.LinterName
		wantChanges int
	}{
		{
			name:        "tech definitively absent retracts the gated linter",
			enable:      []types.LinterName{"arangolint", "errcheck"},
			tech:        map[string]bool{"arangodb": false},
			wantEnable:  []types.LinterName{"errcheck"},
			wantChanges: 1,
		},
		{
			name:        "tech present keeps the linter",
			enable:      []types.LinterName{"arangolint"},
			tech:        map[string]bool{"arangodb": true},
			wantEnable:  []types.LinterName{"arangolint"},
			wantChanges: 0,
		},
		{
			name:        "tech undetected fails open (kept)",
			enable:      []types.LinterName{"arangolint"},
			tech:        map[string]bool{"clickhouse": false},
			wantEnable:  []types.LinterName{"arangolint"},
			wantChanges: 0,
		},
		{
			name:        "empty tech map is a no-op",
			enable:      []types.LinterName{"arangolint"},
			tech:        nil,
			wantEnable:  []types.LinterName{"arangolint"},
			wantChanges: 0,
		},
		{
			name:        "disable entries are never retracted",
			enable:      []types.LinterName{"arangolint"},
			disable:     []types.LinterName{"clickhouselint"},
			tech:        map[string]bool{"arangodb": false, "clickhouse": false},
			wantEnable:  nil,
			wantDisable: []types.LinterName{"clickhouselint"},
			wantChanges: 1,
		},
		{
			name:        "ungated linters are untouched",
			enable:      []types.LinterName{"errcheck", "gosec"},
			tech:        map[string]bool{"arangodb": false},
			wantEnable:  []types.LinterName{"errcheck", "gosec"},
			wantChanges: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observed := &types.Config{Linters: types.LintersConfig{
				Enable:  tt.enable,
				Disable: tt.disable,
			}}

			normalized, changes := RemoveStaleTechGatedLinters(observed, tt.tech)

			if len(changes) != tt.wantChanges {
				t.Fatalf("changes: got %d, want %d (%+v)", len(changes), tt.wantChanges, changes)
			}

			if got := normalized.Linters.Enable; !equalLinterNames(got, tt.wantEnable) {
				t.Errorf("enable: got %v, want %v", got, tt.wantEnable)
			}

			if got := normalized.Linters.Disable; !equalLinterNames(got, tt.wantDisable) {
				t.Errorf("disable: got %v, want %v", got, tt.wantDisable)
			}

			for _, change := range changes {
				if change.Path != PathLintersEnable || change.Kind != ChangeRemove {
					t.Errorf("change %+v must be an enable-list removal", change)
				}
			}
		})
	}
}
