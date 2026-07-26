package linter

import (
	"path/filepath"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// loadPolicy reads the disable-reason sidecar file (if it exists) alongside the
// config file. When the sidecar is present, the fixer enforces that every
// user-disabled linter has a justification entry — unjustified disables are
// re-enabled. When the sidecar is absent, all disables are respected as before.
func (f *Fixer) loadPolicy(configPath string) {
	sidecarPath := filepath.Join(filepath.Dir(configPath), policy.SidecarFileName)

	pol, err := policy.Load(sidecarPath)
	if err != nil {
		f.logger.Warnf("⚠️  Failed to load policy file %s: %v", sidecarPath, err)

		f.pol = nil

		return
	}

	if pol != nil {
		f.logger.Infof("📋 Loaded disable-reason policy from %s (%d justified disables)",
			sidecarPath, len(pol.Disabled))
	}

	f.pol = pol
}

// enforceDisableReasons re-enables linters that are in the disable list without
// a justification in the policy sidecar. Tool-level disabled linters (defined in
// constants.DisabledLinters) are always exempt — they are disabled by this tool,
// not by user choice.
func (f *Fixer) enforceDisableReasons(cfg *types.Config) int {
	if f.pol == nil {
		return 0
	}

	enableSet := types.NewSet(cfg.Linters.Enable...)
	disableSet := types.NewSet(cfg.Linters.Disable...)

	reEnabledCount := 0

	for _, linter := range cfg.Linters.Disable {
		if f.tryReEnableLinter(linter, enableSet, disableSet) {
			reEnabledCount++
		}
	}

	if reEnabledCount > 0 {
		cfg.Linters.Enable = types.ToSortedSlice(enableSet)
		cfg.Linters.Disable = types.ToSortedSlice(disableSet)
	}

	return reEnabledCount
}

// tryReEnableLinter checks whether the given linter should be re-enabled (it is
// disabled without justification) and, if so, moves it from the disable set to
// the enable set and records the action in the audit ledger.
func (f *Fixer) tryReEnableLinter(linter types.LinterName, enableSet, disableSet types.Set[types.LinterName]) bool {
	if isToolLevelDisabled(linter) {
		return false
	}

	if f.pol.IsJustified(string(linter)) {
		return false
	}

	disableSet.Delete(linter)
	enableSet.Add(linter)

	f.logger.Infof("📋 Re-enabling %s: disabled without a justification in %s",
		linter, policy.SidecarFileName)
	f.ledger.Record(audit.ActionReEnabled, string(linter),
		"unjustified disable (no entry in sidecar)")

	return true
}

// isToolLevelDisabled reports whether the linter is in the tool's hardcoded
// DisabledLinters set (funcorder, noinlineerr, depguard, etc.).
func isToolLevelDisabled(linter types.LinterName) bool {
	_, ok := constants.DisabledLinters[linter]

	return ok
}
