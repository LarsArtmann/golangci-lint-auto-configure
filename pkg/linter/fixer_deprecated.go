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
}

// newDeprecatedLinterHandler creates a new deprecated linter handler.
func newDeprecatedLinterHandler(logger *log.Logger, version string) *deprecatedLinterHandler {
	return &deprecatedLinterHandler{logger: logger, version: version}
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

	h.logReplace(linter, replacement, dryRun)

	if !dryRun {
		linterSet.Add(replacement.Replacement)
		h.migrateSettings(cfg, linter, replacement.Replacement)
	}

	return 1
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

func (h *deprecatedLinterHandler) logReplace(linter types.LinterName, replacement types.LinterReplacement, dryRun bool) {
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
func resolveLinterName(name types.LinterName) string {
	if replacement, isDeprecated := constants.DeprecatedLinters[name]; isDeprecated {
		return string(replacement.Replacement)
	}

	return name.String()
}
