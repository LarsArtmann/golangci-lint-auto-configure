package linter

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// linterSnapshot captures the enable, disable, and settings-key state of a config
// so the fixer can diff what changed across a run and record it in the audit ledger.
type linterSnapshot struct {
	enable       types.Set[string]
	disable      types.Set[string]
	settingsKeys types.Set[string]
}

// snapshotLinterState captures the current linter enable/disable/settings-key state.
func snapshotLinterState(cfg *types.Config) linterSnapshot {
	settingsKeys := types.NewSet[string]()
	for name := range cfg.Linters.Settings {
		settingsKeys.Add(name)
	}

	return linterSnapshot{
		enable:       types.NewSet(cfg.Linters.Enable...),
		disable:      types.NewSet(cfg.Linters.Disable...),
		settingsKeys: settingsKeys,
	}
}

// recordConfigChanges diffs the before/after linter state and records every net
// change to the audit ledger. Only called when changes were persisted (not dry-run).
func (f *Fixer) recordConfigChanges(before linterSnapshot, cfg *types.Config) {
	after := snapshotLinterState(cfg)

	for _, linter := range types.ToSortedSlice(after.enable.Difference(before.enable)) {
		reason := constants.LinterReasons[types.LinterName(linter)]
		f.ledger.Record(audit.ActionAddedToEnable, linter, reason)
	}

	for _, linter := range types.ToSortedSlice(before.enable.Difference(after.enable)) {
		f.ledger.Record(audit.ActionRemovedFromEnable, linter, "")
	}

	for _, linter := range types.ToSortedSlice(after.disable.Difference(before.disable)) {
		reason := constants.DisabledLinters[types.LinterName(linter)]
		f.ledger.Record(audit.ActionMovedToDisable, linter, reason)
	}

	for _, linter := range types.ToSortedSlice(before.disable.Difference(after.disable)) {
		f.ledger.Record(audit.ActionRemovedFromDisable, linter, "")
	}

	for _, linter := range types.ToSortedSlice(before.settingsKeys.Difference(after.settingsKeys)) {
		f.ledger.Record(audit.ActionPrunedSettings, linter, "linter disabled")
	}
}
