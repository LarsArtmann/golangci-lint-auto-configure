package linter

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestRemoveUnknownLinters(t *testing.T) {
	known := map[types.LinterName]bool{
		"errcheck": true,
		"gosec":    true,
	}

	tests := []struct {
		name            string
		known           map[types.LinterName]bool
		enable          []types.LinterName
		disable         []types.LinterName
		wantEnable      []types.LinterName
		wantDisable     []types.LinterName
		wantChangeCount int
		wantNoopPointer bool
	}{
		{
			name:            "removes unknown names from enable and disable",
			known:           known,
			enable:          []types.LinterName{"errcheck", "gohumanize", "gosec"},
			disable:         []types.LinterName{"totallyfakeplugin", "gosec"},
			wantEnable:      []types.LinterName{"errcheck", "gosec"},
			wantDisable:     []types.LinterName{"gosec"},
			wantChangeCount: 2,
		},
		{
			name:            "no-op when every name is known",
			known:           known,
			enable:          []types.LinterName{"errcheck"},
			disable:         []types.LinterName{"gosec"},
			wantEnable:      []types.LinterName{"errcheck"},
			wantDisable:     []types.LinterName{"gosec"},
			wantChangeCount: 0,
			wantNoopPointer: true,
		},
		{
			name:            "no-op when the binary could not be probed",
			known:           nil,
			enable:          []types.LinterName{"gohumanize"},
			wantEnable:      []types.LinterName{"gohumanize"},
			wantChangeCount: 0,
			wantNoopPointer: true,
		},
		{
			name:            "no-op for an empty known set",
			known:           map[types.LinterName]bool{},
			enable:          []types.LinterName{"gohumanize"},
			wantEnable:      []types.LinterName{"gohumanize"},
			wantChangeCount: 0,
			wantNoopPointer: true,
		},
		{
			name:            "unknown disable-only entry is removed (config load rejects both lists)",
			known:           known,
			disable:         []types.LinterName{"gohumanize"},
			wantDisable:     []types.LinterName{},
			wantChangeCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observed := &types.Config{}
			observed.Linters.Enable = tt.enable
			observed.Linters.Disable = tt.disable

			normalized, changes := RemoveUnknownLinters(observed, tt.known)

			if len(changes) != tt.wantChangeCount {
				t.Fatalf("RemoveUnknownLinters() changes = %d, want %d (%v)", len(changes), tt.wantChangeCount, changes)
			}

			if tt.wantNoopPointer {
				if normalized != observed {
					t.Errorf("RemoveUnknownLinters() returned a clone for a no-op run, want the original")
				}

				return
			}

			if !equalLinterNames(normalized.Linters.Enable, tt.wantEnable) {
				t.Errorf("enable = %v, want %v", normalized.Linters.Enable, tt.wantEnable)
			}

			if !equalLinterNames(normalized.Linters.Disable, tt.wantDisable) {
				t.Errorf("disable = %v, want %v", normalized.Linters.Disable, tt.wantDisable)
			}

			for _, change := range changes {
				if change.Kind != ChangeRemove {
					t.Errorf("change kind = %s, want remove", change.Kind)
				}

				if change.Path != PathLintersEnable && change.Path != PathLintersDisable {
					t.Errorf("change path = %s, want a linter list path", change.Path)
				}

				if change.Reason == "" {
					t.Errorf("change for %s carries no reason", change.Element)
				}
			}
		})
	}
}

func equalLinterNames(got, want []types.LinterName) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}
