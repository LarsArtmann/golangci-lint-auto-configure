package linter

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// configSnapshot captures the enable, disable, settings-key, and formatter state
// of a config so the fixer can diff what changed across a run and record it in
// the audit ledger.
type configSnapshot struct {
	enable           types.Set[string]
	disable          types.Set[string]
	settingsKeys     types.Set[string]
	formatterEnable  types.Set[string]
	formatterDisable types.Set[string]
}

// snapshotLinterState captures the current linter and formatter state of the config.
func snapshotLinterState(cfg *types.Config) configSnapshot {
	settingsKeys := types.NewSet[string]()
	for name := range cfg.Linters.Settings {
		settingsKeys.Add(name)
	}

	return configSnapshot{
		enable:           types.NewSet(cfg.Linters.Enable...),
		disable:          types.NewSet(cfg.Linters.Disable...),
		settingsKeys:     settingsKeys,
		formatterEnable:  types.NewSet(cfg.Formatters.Enable...),
		formatterDisable: types.NewSet(cfg.Formatters.Disable...),
	}
}

// recordConfigChanges diffs the before/after config state and records every net
// change to the audit ledger. Only called when changes were persisted (not dry-run).
func (f *Fixer) recordConfigChanges(before configSnapshot, cfg *types.Config) {
	after := snapshotLinterState(cfg)

	f.recordLinterChanges(before, after)
	f.recordFormatterChanges(before, after)
}

func (f *Fixer) recordLinterChanges(before, after configSnapshot) {
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

func (f *Fixer) recordFormatterChanges(before, after configSnapshot) {
	for _, formatter := range types.ToSortedSlice(after.formatterEnable.Difference(before.formatterEnable)) {
		f.ledger.Record(audit.ActionFormatterAddedToEnable, formatter, "recommended formatter")
	}

	for _, formatter := range types.ToSortedSlice(before.formatterEnable.Difference(after.formatterEnable)) {
		f.ledger.Record(audit.ActionFormatterRemovedFromEnable, formatter, "")
	}
}
