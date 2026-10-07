package linter

import (
	"os"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// applyReplacementInput builds a linter set + config carrying the deprecated
// predecessor so the handler migrates it.
func applyReplacementInput(
	t *testing.T,
	deprecated, replacement types.LinterName,
) (types.Set[types.LinterName], *types.Config, []types.LinterName) {
	t.Helper()

	if _, isDeprecated := constants.DeprecatedLinters[deprecated]; !isDeprecated {
		t.Fatalf("%q is not in DeprecatedLinters; pick a real predecessor", deprecated)
	}

	linterSet := types.NewSet(deprecated)
	enabled := []types.LinterName{deprecated}

	cfg := &types.Config{}
	cfg.Linters.Settings = map[string]any{
		string(deprecated): map[string]any{"custom": true},
	}

	return linterSet, cfg, enabled
}

// TestReplacementRespectsNeverEnable pins the M32 guard: when the sidecar
// never-enable section lists the REPLACEMENT, the deprecated predecessor is
// still removed but the replacement is NOT added and its settings are NOT
// migrated.
func TestReplacementRespectsNeverEnable(t *testing.T) {
	deprecated := types.LinterName("exhaustruct")
	replacement := types.LinterName("exhaustruct_v5")

	linterSet, cfg, enabled := applyReplacementInput(t, deprecated, replacement)

	handler := newDeprecatedLinterHandler(testQuietLogger(), "v2.14.0")
	handler.neverEnable = func(name types.LinterName) bool { return name == replacement }

	count := 0
	for _, linter := range enabled {
		count += handler.replaceOne(linterSet, linter, false, cfg)
	}

	if count != 1 {
		t.Fatalf("replacement must count as one action, got %d", count)
	}

	if linterSet.Contains(replacement) {
		t.Fatal("replacement was added despite never-enable — documented contract violated")
	}

	if linterSet.Contains(deprecated) {
		t.Fatal("deprecated predecessor must still be removed")
	}

	if _, migrated := cfg.Linters.Settings[string(replacement)]; migrated {
		t.Fatal("settings must not be migrated to a never-enabled replacement")
	}
}

// TestReplacementProceedsWithoutNeverEnable pins the baseline: with no
// never-enable predicate, the replacement path works exactly as before.
func TestReplacementProceedsWithoutNeverEnable(t *testing.T) {
	deprecated := types.LinterName("exhaustruct")
	replacement := types.LinterName("exhaustruct_v5")

	linterSet, cfg, enabled := applyReplacementInput(t, deprecated, replacement)

	handler := newDeprecatedLinterHandler(testQuietLogger(), "v2.14.0")

	for _, linter := range enabled {
		handler.replaceOne(linterSet, linter, false, cfg)
	}

	if !linterSet.Contains(replacement) {
		t.Fatal("replacement must be added when not never-enabled")
	}

	if _, migrated := cfg.Linters.Settings[string(replacement)]; !migrated {
		t.Fatal("settings must migrate to the replacement")
	}
}

// TestPragmaticCompositionWithNeverEnable pins the --pragmatic × never-enable
// composition: shouldSkipLinter (pragmatic) and never-enable are independent
// gates — a pragmatic-skipped linter's absence must not be overridden by
// replacement logic, and a never-enabled replacement stays blocked regardless
// of pragmatic mode.
func TestPragmaticCompositionWithNeverEnable(t *testing.T) {
	// exhaustruct_v5 is the replacement target AND in NeverAutoEnableLinters;
	// pragmatic drops gochecknoglobals etc. from recommendations, but the
	// never-enable predicate applies at the replacement layer independent of
	// pragmatic filtering.
	for pragmaticName := range constants.PragmaticNoiseLinters {
		if _, alsoNeverAuto := constants.NeverAutoEnableLinters[pragmaticName]; alsoNeverAuto {
			t.Fatalf(
				"%s is both pragmatic-noise and never-auto-enable — the sets must stay disjoint",
				pragmaticName,
			)
		}
	}

	// The guard predicate is orthogonal: it fires purely on the sidecar.
	handler := newDeprecatedLinterHandler(testQuietLogger(), "v2.14.0")
	handler.neverEnable = func(types.LinterName) bool { return true }

	linterSet, cfg, enabled := applyReplacementInput(t, "exhaustruct", "exhaustruct_v5")

	for _, linter := range enabled {
		handler.replaceOne(linterSet, linter, false, cfg)
	}

	if linterSet.Contains("exhaustruct_v5") {
		t.Fatal("never-enable must hold in pragmatic composition too")
	}
}

// testQuietLogger discards output.
func testQuietLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
}
