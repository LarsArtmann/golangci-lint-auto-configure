package linter

import (
	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"golang.org/x/mod/semver"
)

// deprecatedLinterHandler handles deprecated linter replacements.
type deprecatedLinterHandler struct {
	logger  *log.Logger
	version string // detected golangci-lint version
	// neverEnable reports whether the sidecar policy forbids adding the given
	// linter to enable. When it fires, the deprecated predecessor is still
	// removed (deprecation cleanup) but the replacement is NOT added — the
	// documented never-enable contract ("never added to enable, period")
	// outranks deprecation migration. NeverAutoEnableLinters deliberately
	// does NOT apply here: replacing a linter the user manually enabled is a
	// migration of their explicit choice, not an auto-enable.
	neverEnable func(types.LinterName) bool
}

// newDeprecatedLinterHandler creates a new deprecated linter handler.
func newDeprecatedLinterHandler(logger *log.Logger, version string) *deprecatedLinterHandler {
	return &deprecatedLinterHandler{logger: logger, version: version, neverEnable: nil}
}

// replaceLinters replaces deprecated linters with their successors in the linter set.
// Returns the updated set and the number of replacements applied.
func (h *deprecatedLinterHandler) replaceLinters(
	linterSet types.Set[types.LinterName],
	enabledLinters []types.LinterName,
	dryRun bool,
	cfg *types.Config,
) (types.Set[types.LinterName], int) {
	count := 0

	for _, linter := range enabledLinters {
		count += h.replaceOne(linterSet, linter, dryRun, cfg)
	}

	return linterSet, count
}

func (h *deprecatedLinterHandler) replaceOne(
	linterSet types.Set[types.LinterName],
	linter types.LinterName,
	dryRun bool,
	cfg *types.Config,
) int {
	replacement, isDeprecated := constants.DeprecatedLinters[linter]
	if !isDeprecated {
		return 0
	}

	if !h.replacementAvailable(replacement) {
		h.logSkip(linter, replacement)

		return 0
	}

	return h.applyReplacement(linterSet, linter, replacement, dryRun, cfg)
}

func (h *deprecatedLinterHandler) applyReplacement(
	linterSet types.Set[types.LinterName],
	linter types.LinterName,
	replacement types.LinterReplacement,
	dryRun bool,
	cfg *types.Config,
) int {
	linterSet.Delete(linter)

	if replacement.Replacement == "" {
		h.logRemove(linter, replacement, dryRun)

		return 1
	}

	if linterSet.Contains(replacement.Replacement) {
		h.logKeep(linter, replacement.Replacement, dryRun)

		return 1
	}

	if h.blockedByNeverEnable(replacement.Replacement) {
		h.logNeverEnableSkip(linter, replacement, dryRun)

		return 1
	}

	h.addReplacement(linterSet, linter, replacement, dryRun, cfg)

	return 1
}

// addReplacement logs the swap, then adds the replacement linter and migrates
// its settings outside dry-run mode.
func (h *deprecatedLinterHandler) addReplacement(
	linterSet types.Set[types.LinterName],
	linter types.LinterName,
	replacement types.LinterReplacement,
	dryRun bool,
	cfg *types.Config,
) {
	h.logReplace(linter, replacement, dryRun)

	if dryRun {
		return
	}

	linterSet.Add(replacement.Replacement)
	h.migrateSettings(cfg, linter, replacement.Replacement)
}

func (h *deprecatedLinterHandler) logSkip(linter types.LinterName, replacement types.LinterReplacement) {
	h.logger.Debugf(
		"Skipping deprecation replacement %s -> %s: installed golangci-lint %s < %s",
		linter, replacement.Replacement, h.version, replacement.MinVersion,
	)
}

// replacementAvailable checks if the replacement linter exists in the installed golangci-lint version.
func (h *deprecatedLinterHandler) replacementAvailable(replacement types.LinterReplacement) bool {
	return isReplacementAvailable(replacement, h.version)
}

// migrateSettings moves linter settings from the deprecated name to the replacement name.
func (h *deprecatedLinterHandler) migrateSettings(cfg *types.Config, oldName, newName types.LinterName) {
	if cfg.Linters.Settings == nil {
		return
	}

	oldKey := string(oldName)
	newKey := string(newName)

	settings, exists := cfg.Linters.Settings[oldKey]
	if !exists {
		return
	}

	h.logger.Debugf("Migrating settings from %s to %s", oldName, newName)

	cfg.Linters.Settings[newKey] = settings
	delete(cfg.Linters.Settings, oldKey)
}

// exclusionLinterMigration records one dead linter name found inside an
// exclusion rule: renamed to a successor (To set) or dropped outright (To
// empty — the linter was removed from golangci-lint with no replacement).
type exclusionLinterMigration struct {
	From types.LinterName
	To   types.LinterName
}

// exclusionCleanup aggregates what the exclusion-rule sweep changed in one
// run: deprecated names rewritten to their replacements, dead names dropped
// outright, and how many exclusion rules were touched.
type exclusionCleanup struct {
	renames map[types.LinterName]types.LinterName
	removed map[types.LinterName]bool
	rules   int
}

func newExclusionCleanup() exclusionCleanup {
	return exclusionCleanup{
		renames: make(map[types.LinterName]types.LinterName),
		removed: make(map[types.LinterName]bool),
		rules:   0,
	}
}

func (c *exclusionCleanup) absorb(migrations []exclusionLinterMigration) {
	for _, migration := range migrations {
		if migration.To == "" {
			c.removed[migration.From] = true
		} else {
			c.renames[migration.From] = migration.To
		}
	}
}

// cleanupExclusionLinters rewrites deprecated linter names inside
// linters.exclusions.rules[].linters. The enable-set swap and the pre-flight
// sweep never touch exclusion rules, so a migrated config keeps carrying the
// dead predecessor there — golangci-lint rejects unknown linter names in
// exclusion rules, which hard-breaks config loading once upstream deletes the
// name. The sweep runs on every fixer pass so post-migration configs heal too,
// and is version-gated like the enable-set swap: a name is only rewritten once
// its successor exists in the installed golangci-lint. Dry-run leaves the
// config untouched and only reports counts.
//
// Removal semantics for linters deleted without a replacement: dropping the
// name keeps path-scoped rules intact, but a rule carrying a Text or Source
// pattern scoped purely to the dead linter is dropped as well — left behind
// with an empty linter list, that pattern would suppress matching findings
// from every linter.
func (h *deprecatedLinterHandler) cleanupExclusionLinters(cfg *types.Config, dryRun bool) exclusionCleanup {
	cleanup := newExclusionCleanup()

	if cfg == nil || len(cfg.Linters.Exclusions.Rules) == 0 {
		return cleanup
	}

	kept := h.sweepExclusionRules(cfg.Linters.Exclusions.Rules, &cleanup)

	if !dryRun {
		cfg.Linters.Exclusions.Rules = kept
	}

	return cleanup
}

// sweepExclusionRules sweeps every exclusion rule, tracks the migrations in
// cleanup, and returns the rules to keep (text-scoped rules whose linter list
// the sweep emptied are dropped entirely).
func (h *deprecatedLinterHandler) sweepExclusionRules(
	rules []types.ExclusionRuleConfig,
	cleanup *exclusionCleanup,
) []types.ExclusionRuleConfig {
	kept := make([]types.ExclusionRuleConfig, 0, len(rules))

	for _, rule := range rules {
		swept := h.sweepRuleLinters(rule)

		cleanup.absorb(swept.migrations)

		if !swept.changed {
			kept = append(kept, rule)

			continue
		}

		cleanup.rules++

		if swept.dropRule {
			continue
		}

		rule.Linters = swept.linters
		kept = append(kept, rule)
	}

	return kept
}

// exclusionRuleSweep is the per-rule result of sweeping one exclusion rule.
type exclusionRuleSweep struct {
	linters    []string
	migrations []exclusionLinterMigration
	changed    bool
	dropRule   bool
}

func newExclusionRuleSweep(capacity int) exclusionRuleSweep {
	return exclusionRuleSweep{
		linters:    make([]string, 0, capacity),
		migrations: nil,
		changed:    false,
		dropRule:   false,
	}
}

func (s *exclusionRuleSweep) keep(name string, seen map[string]bool) {
	if seen[name] {
		return
	}

	seen[name] = true
	s.linters = append(s.linters, name)
}

func (s *exclusionRuleSweep) drop(from types.LinterName) {
	s.migrations = append(s.migrations, exclusionLinterMigration{From: from, To: ""})
}

func (s *exclusionRuleSweep) rename(from, to types.LinterName, seen map[string]bool) {
	s.migrations = append(s.migrations, exclusionLinterMigration{From: from, To: to})
	s.keep(string(to), seen)
}

// sweepRuleLinters rewrites one rule's linter list: deprecated names with an
// available successor become that successor (deduped), removed names vanish,
// and a rule whose Text/Source scope exists only for a removed linter is
// dropped entirely.
func (h *deprecatedLinterHandler) sweepRuleLinters(rule types.ExclusionRuleConfig) exclusionRuleSweep {
	sweep := newExclusionRuleSweep(len(rule.Linters))
	seen := make(map[string]bool, len(rule.Linters))

	for _, name := range rule.Linters {
		linterName := types.LinterName(name)

		replacement, isDeprecated := constants.DeprecatedLinters[linterName]
		if !isDeprecated || !replacementAvailable(replacement, h.version) {
			sweep.keep(name, seen)

			continue
		}

		sweep.changed = true

		if replacement.Replacement == "" {
			sweep.drop(linterName)

			continue
		}

		sweep.rename(linterName, replacement.Replacement, seen)
	}

	sweep.dropRule = sweep.changed && len(sweep.linters) == 0 && (rule.Text != "" || rule.Source != "")

	return sweep
}

func (h *deprecatedLinterHandler) logKeep(linter types.LinterName, replacement types.LinterName, dryRun bool) {
	if dryRun {
		h.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)", linter, replacement)
	} else {
		h.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacement)
	}
}

func dryRunActionPrefix(dryRun bool) string {
	if dryRun {
		return "[DRY-RUN] Would "
	}

	return ""
}

func (h *deprecatedLinterHandler) logRemove(linter types.LinterName, replacement types.LinterReplacement, dryRun bool) {
	h.logger.Infof("%sremove deprecated linter: %s (%s)", dryRunActionPrefix(dryRun), linter, replacement.Reason)
}

func (h *deprecatedLinterHandler) logReplace(
	linter types.LinterName,
	replacement types.LinterReplacement,
	dryRun bool,
) {
	h.logger.Infof(
		"%sreplace deprecated linter: %s -> %s (%s)",
		dryRunActionPrefix(dryRun),
		linter,
		replacement.Replacement,
		replacement.Reason,
	)
}

// hasDeprecatedLinters checks if any of the enabled linters are deprecated
// and have replacements available for the installed golangci-lint version.
func hasDeprecatedLinters(enabledLinters []types.LinterName, version string) bool {
	for _, linter := range enabledLinters {
		replacement, isDeprecated := constants.DeprecatedLinters[linter]
		if !isDeprecated {
			continue
		}

		if replacementAvailable(replacement, version) {
			return true
		}
	}

	return false
}

// replacementAvailable checks if a replacement is available for the given golangci-lint version.
func replacementAvailable(replacement types.LinterReplacement, version string) bool {
	return isReplacementAvailable(replacement, version)
}

// isReplacementAvailable is the shared implementation for replacement availability checks.
func isReplacementAvailable(replacement types.LinterReplacement, version string) bool {
	if replacement.MinVersion == "" {
		return true
	}

	if version == "" {
		return true
	}

	return semver.Compare(version, replacement.MinVersion) >= 0
}

// resolveLinterName resolves a linter name, replacing deprecated linters with their successors.
func resolveLinterName(name types.LinterName) types.LinterName {
	if replacement, isDeprecated := constants.DeprecatedLinters[name]; isDeprecated {
		return replacement.Replacement
	}

	return name
}

// blockedByNeverEnable reports whether the sidecar policy forbids adding the
// replacement linter to enable.
func (h *deprecatedLinterHandler) blockedByNeverEnable(replacement types.LinterName) bool {
	return h.neverEnable != nil && h.neverEnable(replacement)
}

func (h *deprecatedLinterHandler) logNeverEnableSkip(
	linter types.LinterName, replacement types.LinterReplacement, dryRun bool,
) {
	h.logger.Infof(
		"%sremove deprecated linter: %s — replacement %s is under never-enable, NOT added",
		dryRunActionPrefix(dryRun), linter, replacement.Replacement,
	)
}
