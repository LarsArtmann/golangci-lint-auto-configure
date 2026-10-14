package linter

import (
	"os"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// newQuietFixer builds a Fixer whose logs stay out of test output. Only the
// logger matters: loadPolicy and the pure replacement normalizers never touch
// the loader/analyzer.
func newQuietFixer() *Fixer {
	return &Fixer{
		logger: log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel}),
		ledger: audit.NoopRecorder{},
	}
}

// exhaustructReplacementInput builds a config carrying the deprecated
// exhaustruct predecessor (with a settings block) so the replacement
// normalizer has something to migrate.
func exhaustructReplacementInput(t *testing.T) *types.Config {
	t.Helper()

	deprecated := types.LinterName("exhaustruct")

	record, known := constants.PolicyRecord(deprecated)
	if !known || record.SupersessionReason == "" {
		t.Fatalf("%q has no supersession record; pick a real predecessor", deprecated)
	}

	cfg := &types.Config{}
	cfg.Linters.Enable = []types.LinterName{deprecated}
	cfg.Linters.Settings = map[string]any{
		string(deprecated): map[string]any{"custom": true},
	}

	return cfg
}

// TestReplacementRespectsNeverEnable pins the M32 guard on the reconciler:
// when the sidecar never-enable section lists the REPLACEMENT, the deprecated
// predecessor is still removed but the replacement is NOT added and its
// settings are NOT migrated.
func TestReplacementRespectsNeverEnable(t *testing.T) {
	deprecated := types.LinterName("exhaustruct")
	replacement := types.LinterName("exhaustruct_v5")

	cfg := exhaustructReplacementInput(t)

	normalized, changes := ReplaceDeprecatedLinters(
		cfg, types.GolangciVersion{}, func(name types.LinterName) bool { return name == replacement },
	)

	if len(changes) == 0 {
		t.Fatal("the predecessor removal must be reported as a change")
	}

	for _, name := range normalized.Linters.Enable {
		if name == replacement {
			t.Fatal("replacement was added despite never-enable — documented contract violated")
		}

		if name == deprecated {
			t.Fatal("deprecated predecessor must still be removed")
		}
	}

	if _, migrated := normalized.Linters.Settings[string(replacement)]; migrated {
		t.Fatal("settings must not be migrated to a never-enabled replacement")
	}
}

// TestReplacementProceedsWithoutNeverEnable pins the baseline: with no
// never-enable predicate, the replacement path swaps the predecessor for its
// successor and migrates the settings block.
func TestReplacementProceedsWithoutNeverEnable(t *testing.T) {
	deprecated := types.LinterName("exhaustruct")
	replacement := types.LinterName("exhaustruct_v5")

	cfg := exhaustructReplacementInput(t)

	normalized, changes := ReplaceDeprecatedLinters(cfg, types.GolangciVersion{}, nil)

	found := false

	for _, name := range normalized.Linters.Enable {
		if name == replacement {
			found = true
		}

		if name == deprecated {
			t.Fatal("deprecated predecessor must be removed on swap")
		}
	}

	if !found {
		t.Fatal("replacement must be added when not never-enabled")
	}

	if _, migrated := normalized.Linters.Settings[string(replacement)]; !migrated {
		t.Fatal("settings must migrate to the replacement")
	}

	if len(changes) == 0 {
		t.Fatal("the swap must be reported as changes")
	}
}

// TestReplacementVersionGate pins the version gate: when the installed tool
// predates the supersession version, the predecessor is dropped WITHOUT the
// successor and the change reason says why.
func TestReplacementVersionGate(t *testing.T) {
	deprecated := types.LinterName("exhaustruct")
	replacement := types.LinterName("exhaustruct_v5")

	cfg := exhaustructReplacementInput(t)

	tool := types.MustParseGolangciVersion("v2.12.0")

	normalized, changes := ReplaceDeprecatedLinters(cfg, tool, nil)

	for _, name := range normalized.Linters.Enable {
		if name == replacement {
			t.Fatal("replacement must not be added when the tool predates the supersession")
		}
	}

	gated := false

	for _, change := range changes {
		if change.Element == string(deprecated) && change.Kind == ChangeRemove {
			gated = true
		}
	}

	if !gated {
		t.Fatalf("expected a removal change for %s, got %+v", deprecated, changes)
	}
}

// TestPragmaticCompositionWithNeverEnable pins the --pragmatic × never-enable
// composition: the pragmatic-noise set and NeverAutoEnableLinters must stay
// disjoint, and the never-enable predicate fires purely on the sidecar —
// independent of pragmatic filtering.
func TestPragmaticCompositionWithNeverEnable(t *testing.T) {
	for pragmaticName := range constants.PragmaticNoiseLinters {
		if _, alsoNeverAuto := constants.NeverAutoEnableLinters[pragmaticName]; alsoNeverAuto {
			t.Fatalf(
				"%s is both pragmatic-noise and never-auto-enable — the sets must stay disjoint",
				pragmaticName,
			)
		}
	}

	deprecated := types.LinterName("exhaustruct")
	replacement := types.LinterName("exhaustruct_v5")

	cfg := exhaustructReplacementInput(t)

	normalized, _ := ReplaceDeprecatedLinters(
		cfg, types.GolangciVersion{}, func(types.LinterName) bool { return true },
	)

	for _, name := range normalized.Linters.Enable {
		if name == replacement {
			t.Fatal("never-enable must hold under pragmatic composition too")
		}
	}
}
