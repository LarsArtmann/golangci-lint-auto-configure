package diff

import (
	"strings"
	"testing"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

func TestDiffer_Compare(t *testing.T) {
	differ := NewDiffer()

	tests := []struct {
		name        string
		old         *types.Config
		new         *types.Config
		wantChanges int
		description string
	}{
		{
			name: "version change",
			old: &types.Config{
				Version: "1",
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			new: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			wantChanges: 1,
			description: "Should detect version change",
		},
		{
			name: "linter added",
			old: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			new: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck", "gosec"}},
			},
			wantChanges: 1,
			description: "Should detect added linter",
		},
		{
			name: "linter removed",
			old: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck", "gosec"}},
			},
			new: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			wantChanges: 1,
			description: "Should detect removed linter",
		},
		{
			name: "multiple changes",
			old: &types.Config{
				Version: "1",
				Run:     types.RunConfig{Timeout: "5m"},
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			new: &types.Config{
				Version: "2",
				Run:     types.RunConfig{Timeout: "10m"},
				Linters: types.LintersConfig{Enable: []string{"gosec"}},
			},
			wantChanges: 4,
			description: "Should detect multiple changes (version, timeout, errcheck removed, gosec added)",
		},
		{
			name: "no changes",
			old: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			new: &types.Config{
				Version: "2",
				Linters: types.LintersConfig{Enable: []string{"errcheck"}},
			},
			wantChanges: 0,
			description: "Should detect no changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := differ.Compare(tt.old, tt.new)
			if len(changes) != tt.wantChanges {
				t.Errorf("Compare() returned %d changes, want %d - %s", len(changes), tt.wantChanges, tt.description)
			}
		})
	}
}

func TestDiffer_FormatChanges(t *testing.T) {
	differ := NewDiffer()

	tests := []struct {
		name    string
		changes []Change
		want    []string // substrings that should be present
	}{
		{
			name:    "no changes",
			changes: []Change{},
			want:    []string{"No changes detected"},
		},
		{
			name: "added linter",
			changes: []Change{
				{Type: ChangeTypeAdded, Path: "linters.enable.gosec", Description: "Enabled linter: gosec"},
			},
			want: []string{"1 added", "+ Enabled linter: gosec"},
		},
		{
			name: "removed linter",
			changes: []Change{
				{Type: ChangeTypeRemoved, Path: "linters.enable.errcheck", Description: "Disabled linter: errcheck"},
			},
			want: []string{"1 removed", "- Disabled linter: errcheck"},
		},
		{
			name: "modified timeout",
			changes: []Change{
				{Type: ChangeTypeModified, Path: "run.timeout", Description: "Timeout changed from 5m to 10m"},
			},
			want: []string{"1 modified", "~ Timeout changed from 5m to 10m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := differ.FormatChanges(tt.changes)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("FormatChanges() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestDiffer_GetSummary(t *testing.T) {
	differ := NewDiffer()

	tests := []struct {
		name    string
		changes []Change
		want    string
	}{
		{
			name:    "no changes",
			changes: []Change{},
			want:    "No changes",
		},
		{
			name: "mixed changes",
			changes: []Change{
				{Type: ChangeTypeAdded},
				{Type: ChangeTypeAdded},
				{Type: ChangeTypeRemoved},
				{Type: ChangeTypeModified},
			},
			want: "2 added, 1 removed, 1 modified",
		},
		{
			name: "only added",
			changes: []Change{
				{Type: ChangeTypeAdded},
				{Type: ChangeTypeAdded},
			},
			want: "2 added",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := differ.GetSummary(tt.changes)
			if got != tt.want {
				t.Errorf("GetSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChangeType_String(t *testing.T) {
	tests := []struct {
		changeType ChangeType
		want       string
	}{
		{ChangeTypeAdded, "ADDED"},
		{ChangeTypeRemoved, "REMOVED"},
		{ChangeTypeModified, "MODIFIED"},
		{ChangeType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.changeType.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
