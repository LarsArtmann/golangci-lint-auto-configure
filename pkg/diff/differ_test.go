package diff_test

import (
	"strings"
	"testing"

	diffpkg "github.com/larsartmann/golangcli-linter-auto-configure/pkg/diff"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

func TestDiffer_Compare(t *testing.T) {
	differ := diffpkg.NewDiffer()

	baseConfigV2WithErrcheck := &types.Config{
		Version: "2",
		Linters: types.LintersConfig{Enable: []string{"errcheck"}},
	}

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
			new:         baseConfigV2WithErrcheck,
			wantChanges: 1,
			description: "Should detect version change",
		},
		{
			name: "linter added",
			old:  baseConfigV2WithErrcheck,
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
			new:         baseConfigV2WithErrcheck,
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
			name:        "no changes",
			old:         baseConfigV2WithErrcheck,
			new:         baseConfigV2WithErrcheck,
			wantChanges: 0,
			description: "Should detect no changes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			changes := differ.Compare(testCase.old, testCase.new)
			if len(changes) != testCase.wantChanges {
				t.Errorf("Compare() returned %d changes, want %d - %s", len(changes), testCase.wantChanges, testCase.description)
			}
		})
	}
}

func TestDiffer_FormatChanges(t *testing.T) {
	differ := diffpkg.NewDiffer()

	tests := []struct {
		name    string
		changes []diffpkg.Change
		want    []string // substrings that should be present
	}{
		{
			name:    "no changes",
			changes: []diffpkg.Change{},
			want:    []string{"No changes detected"},
		},
		{
			name: "added linter",
			changes: []diffpkg.Change{
				{Type: diffpkg.ChangeTypeAdded, Path: "linters.enable.gosec", Description: "Enabled linter: gosec"},
			},
			want: []string{"1 added", "+ Enabled linter: gosec"},
		},
		{
			name: "removed linter",
			changes: []diffpkg.Change{
				{Type: diffpkg.ChangeTypeRemoved, Path: "linters.enable.errcheck", Description: "Disabled linter: errcheck"},
			},
			want: []string{"1 removed", "- Disabled linter: errcheck"},
		},
		{
			name: "modified timeout",
			changes: []diffpkg.Change{
				{Type: diffpkg.ChangeTypeModified, Path: "run.timeout", Description: "Timeout changed from 5m to 10m"},
			},
			want: []string{"1 modified", "~ Timeout changed from 5m to 10m"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := differ.FormatChanges(testCase.changes)
			for _, wantStr := range testCase.want {
				if !strings.Contains(got, wantStr) {
					t.Errorf("FormatChanges() = %q, should contain %q", got, wantStr)
				}
			}
		})
	}
}

func TestDiffer_GetSummary(t *testing.T) {
	differ := diffpkg.NewDiffer()

	tests := []struct {
		name    string
		changes []diffpkg.Change
		want    string
	}{
		{
			name:    "no changes",
			changes: []diffpkg.Change{},
			want:    "No changes",
		},
		{
			name: "mixed changes",
			changes: []diffpkg.Change{
				{Type: diffpkg.ChangeTypeAdded},
				{Type: diffpkg.ChangeTypeAdded},
				{Type: diffpkg.ChangeTypeRemoved},
				{Type: diffpkg.ChangeTypeModified},
			},
			want: "2 added, 1 removed, 1 modified",
		},
		{
			name: "only added",
			changes: []diffpkg.Change{
				{Type: diffpkg.ChangeTypeAdded},
				{Type: diffpkg.ChangeTypeAdded},
			},
			want: "2 added",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := differ.GetSummary(testCase.changes)
			if got != testCase.want {
				t.Errorf("GetSummary() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestChangeType_String(t *testing.T) {
	tests := []struct {
		changeType diffpkg.ChangeType
		want       string
	}{
		{diffpkg.ChangeTypeAdded, "ADDED"},
		{diffpkg.ChangeTypeRemoved, "REMOVED"},
		{diffpkg.ChangeTypeModified, "MODIFIED"},
		{diffpkg.ChangeType(99), "UNKNOWN"},
	}

	for _, testCase := range tests {
		t.Run(testCase.want, func(t *testing.T) {
			if got := testCase.changeType.String(); got != testCase.want {
				t.Errorf("String() = %q, want %q", got, testCase.want)
			}
		})
	}
}
