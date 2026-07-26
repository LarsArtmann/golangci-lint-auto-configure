package diff_test

import (
	"strings"
	"testing"

	diffpkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

var baseV2 = &types.Config{
	Version: "2",
	Linters: types.LintersConfig{Enable: []types.LinterName{"errcheck"}},
}

func TestDiffer_Compare(t *testing.T) {
	differ := diffpkg.NewDiffer()

	for _, tc := range compareTests {
		t.Run(tc.name, func(t *testing.T) {
			changes := differ.Compare(tc.old, tc.new)
			if len(changes) != tc.wantChanges {
				t.Errorf("got %d changes, want %d", len(changes), tc.wantChanges)
			}
		})
	}
}

var compareTests = []struct {
	name        string
	old, new    *types.Config
	wantChanges int
}{
	{
		"version change",
		newTestConfigV1("errcheck"),
		baseV2,
		1,
	},
	{
		"linter added",
		baseV2,
		newTestConfig("2", "", []string{"errcheck", "gosec"}),
		1,
	},
	{
		"linter removed",
		newTestConfig("2", "", []string{"errcheck", "gosec"}),
		baseV2,
		1,
	},
	{
		"multiple changes",
		newTestConfig("1", "5m", []string{"errcheck"}),
		newTestConfig("2", "10m", []string{"gosec"}),
		4,
	},
	{"no changes", baseV2, baseV2, 0},
}

func newTestConfig(version, timeout string, linters []string) *types.Config {
	return &types.Config{
		Version: types.Version(version),
		Run:     types.RunConfig{Timeout: timeout},
		Linters: types.LintersConfig{Enable: linters},
	}
}

func newTestConfigV1(linters string) *types.Config {
	return newTestConfig("1", "", []string{linters})
}

func TestDiffer_FormatChanges(t *testing.T) {
	differ := diffpkg.NewDiffer()

	for _, tc := range formatTests {
		t.Run(tc.name, func(t *testing.T) {
			got := differ.FormatChanges(tc.changes)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("FormatChanges() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

var formatTests = []struct {
	name    string
	changes []diffpkg.Change
	want    []string
}{
	{"no changes", []diffpkg.Change{}, []string{"No changes detected"}},
	{
		"added linter",
		[]diffpkg.Change{
			{
				Type:        diffpkg.ChangeTypeAdded,
				Path:        "linters.enable.gosec",
				Description: "Enabled linter: gosec",
			},
		},
		[]string{"1 added", "+ Enabled linter: gosec"},
	},
	{
		"removed linter",
		[]diffpkg.Change{
			{
				Type:        diffpkg.ChangeTypeRemoved,
				Path:        "linters.enable.errcheck",
				Description: "Disabled linter: errcheck",
			},
		},
		[]string{"1 removed", "- Disabled linter: errcheck"},
	},
	{
		"modified timeout",
		[]diffpkg.Change{
			{
				Type:        diffpkg.ChangeTypeModified,
				Path:        "run.timeout",
				Description: "Timeout changed from 5m to 10m",
			},
		},
		[]string{"1 modified", "~ Timeout changed from 5m to 10m"},
	},
}

func TestDiffer_GetSummary(t *testing.T) {
	differ := diffpkg.NewDiffer()

	for _, tc := range summaryTests {
		t.Run(tc.name, func(t *testing.T) {
			got := differ.GetSummary(tc.changes)
			if got != tc.want {
				t.Errorf("GetSummary() = %q, want %q", got, tc.want)
			}
		})
	}
}

var summaryTests = []struct {
	name    string
	changes []diffpkg.Change
	want    string
}{
	{"no changes", []diffpkg.Change{}, "No changes"},
	{
		"mixed changes",
		[]diffpkg.Change{
			{Type: diffpkg.ChangeTypeAdded},
			{Type: diffpkg.ChangeTypeAdded},
			{Type: diffpkg.ChangeTypeRemoved},
			{Type: diffpkg.ChangeTypeModified},
		},
		"2 added, 1 removed, 1 modified",
	},
	{
		"only added",
		[]diffpkg.Change{{Type: diffpkg.ChangeTypeAdded}, {Type: diffpkg.ChangeTypeAdded}},
		"2 added",
	},
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

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.changeType.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}
