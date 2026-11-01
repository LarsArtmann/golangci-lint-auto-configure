package linter

import (
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// enableNames flattens the config's enable list to plain strings, sorted.
func enableNames(cfg *types.Config) []string {
	names := make([]string, 0, len(cfg.Linters.Enable))
	for _, name := range cfg.Linters.Enable {
		names = append(names, string(name))
	}

	slices.Sort(names)

	return names
}

// countName counts one linter's occurrences in the enable list.
func countName(cfg *types.Config, name string) int {
	count := 0

	for _, candidate := range cfg.Linters.Enable {
		if string(candidate) == name {
			count++
		}
	}

	return count
}

// TestChainCollapsesInOneRun pins the one-round convergence of chained
// supersessions: exhaustivestruct was renamed to exhaustruct (v1.46.0), and
// exhaustruct itself was deprecated for exhaustruct_v5 (v2.13.0). A config
// carrying exhaustivestruct must land on exhaustruct_v5 in ONE Replace run
// on a binary that carries the v5 linter — a two-step settle would make
// `configure --check` report drift one run after every repair.
func TestChainCollapsesInOneRun(t *testing.T) {
	cfg := &types.Config{}
	cfg.Linters.Enable = []types.LinterName{"exhaustivestruct"}

	normalized, changes := ReplaceDeprecatedLinters(
		cfg,
		types.MustParseGolangciVersion("v2.14.0"),
		nil,
	)

	names := enableNames(normalized)

	if !slices.Equal(names, []string{"exhaustruct_v5"}) {
		t.Fatalf("enable = %v, want [exhaustruct_v5] (chain must collapse in one run)", names)
	}

	var removed, added bool

	for _, change := range changes {
		switch change.Kind {
		case ChangeRemove:
			removed = removed || change.Element == "exhaustivestruct"
		case ChangeAdd:
			added = added || change.Element == "exhaustruct_v5"
		case ChangeUpdate:
			t.Fatalf("unexpected update change %+v", change)
		}
	}

	if !removed || !added {
		t.Fatalf("changes must report the removal and the addition (removed=%t added=%t)", removed, added)
	}
}

// TestChainBlockedMidWayDropsWithoutReplacement pins the stable outcome when
// the chain dead-ends on a tool-gated hop: on a pre-v2.13 binary the
// exhaustivestruct chain cannot reach exhaustruct_v5, and leaving the
// intermediate exhaustruct behind would oscillate (added on run 1, dropped
// on run 2) — so the linter is dropped without any replacement, the same
// outcome a directly-enabled exhaustruct gets on that binary.
func TestChainBlockedMidWayDropsWithoutReplacement(t *testing.T) {
	cfg := &types.Config{}
	cfg.Linters.Enable = []types.LinterName{"exhaustivestruct"}

	normalized, changes := ReplaceDeprecatedLinters(
		cfg,
		types.MustParseGolangciVersion("v2.12.0"),
		nil,
	)

	if names := enableNames(normalized); len(names) != 0 {
		t.Fatalf("enable = %v, want empty (blocked chain must not leave a deprecated intermediate)", names)
	}

	if len(changes) != 1 || changes[0].Kind != ChangeRemove || changes[0].Element != "exhaustivestruct" {
		t.Fatalf("changes = %+v, want exactly one exhaustivestruct removal", changes)
	}

	if !strings.Contains(changes[0].Reason, "needs v2.13.0") {
		t.Fatalf("removal reason %q must explain the blocked replacement", changes[0].Reason)
	}
}

// TestChainRespectsNeverEnableOnTerminal pins the never-enable guard along
// the chain: with the terminal successor under never-enable, the predecessor
// is dropped without any replacement (identical to the single-hop contract).
func TestChainRespectsNeverEnableOnTerminal(t *testing.T) {
	cfg := &types.Config{}
	cfg.Linters.Enable = []types.LinterName{"exhaustivestruct"}

	normalized, _ := ReplaceDeprecatedLinters(
		cfg,
		types.MustParseGolangciVersion("v2.14.0"),
		func(name types.LinterName) bool { return name == "exhaustruct_v5" },
	)

	if names := enableNames(normalized); len(names) != 0 {
		t.Fatalf("enable = %v, want empty (never-enabled terminal must not be added)", names)
	}
}

// TestSharedSuccessorAddedOnce pins truthful change accounting when several
// deprecated linters share one successor: deadcode, varcheck, and structcheck
// all land on unused — exactly one unused entry, one addition, three removals.
func TestSharedSuccessorAddedOnce(t *testing.T) {
	cfg := &types.Config{}
	cfg.Linters.Enable = []types.LinterName{"deadcode", "varcheck", "structcheck"}

	normalized, changes := ReplaceDeprecatedLinters(
		cfg,
		types.MustParseGolangciVersion("v2.14.0"),
		nil,
	)

	if count := countName(normalized, "unused"); count != 1 {
		t.Fatalf("unused appears %d times, want exactly one shared-successor entry", count)
	}

	if names := enableNames(normalized); !slices.Equal(names, []string{"unused"}) {
		t.Fatalf("enable = %v, want [unused]", names)
	}

	adds := 0

	for _, change := range changes {
		if change.Kind == ChangeAdd {
			if change.Element != "unused" {
				t.Fatalf("unexpected addition %+v", change)
			}

			adds++
		}
	}

	if adds != 1 {
		t.Fatalf("%d addition changes for unused, want 1 (the change list must stay truthful)", adds)
	}
}

// TestRemovedWithoutReplacementStaysDropped pins the no-replacement records:
// execinquery, interfacer, and nosnakecase were removed in v2 with no
// replacement prescribed upstream — the sweep drops them without inventing
// a successor (a fabricated staticcheck/revive enable would silently change
// the user's lint surface).
func TestRemovedWithoutReplacementStaysDropped(t *testing.T) {
	cfg := &types.Config{}
	cfg.Linters.Enable = []types.LinterName{"execinquery", "interfacer", "nosnakecase", "ifshort"}

	normalized, changes := ReplaceDeprecatedLinters(
		cfg,
		types.MustParseGolangciVersion("v2.14.0"),
		nil,
	)

	if names := enableNames(normalized); len(names) != 0 {
		t.Fatalf("enable = %v, want empty (removed linters must not gain fabricated successors)", names)
	}

	if len(changes) != 4 {
		t.Fatalf("%d changes, want exactly 4 removals", len(changes))
	}
}

// TestExclusionChainRenamesToTerminal pins the exclusion-rule sweep's chain
// behavior: an exhaustivestruct entry renames straight to exhaustruct_v5 in
// one run (not to the deprecated intermediate), and drops on a binary that
// cannot carry the terminal.
func TestExclusionChainRenamesToTerminal(t *testing.T) {
	newCfg := func() *types.Config {
		cfg := &types.Config{}
		cfg.Linters.Exclusions.Rules = []types.ExclusionRuleConfig{
			{Path: `_test\.go`, Linters: []string{"exhaustivestruct"}},
		}

		return cfg
	}

	renamed, changes := RenameDeprecatedInExclusionRules(newCfg(), types.MustParseGolangciVersion("v2.14.0"))
	rule := renamed.Linters.Exclusions.Rules[0]

	if !slices.Equal(rule.Linters, []string{"exhaustruct_v5"}) {
		t.Fatalf("exclusion linters = %v, want [exhaustruct_v5] (terminal rename)", rule.Linters)
	}

	if len(changes) != 1 || changes[0].Kind != ChangeUpdate {
		t.Fatalf("changes = %+v, want one update", changes)
	}

	dropped, _ := RenameDeprecatedInExclusionRules(newCfg(), types.MustParseGolangciVersion("v2.12.0"))
	if rules := dropped.Linters.Exclusions.Rules; len(rules) != 0 {
		t.Fatalf("rules = %+v, want none (blocked chain must drop the entry, not leave an invalid rule)", rules)
	}
}
