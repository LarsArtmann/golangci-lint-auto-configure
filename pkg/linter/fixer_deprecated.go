package linter

import (
	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// deprecatedLinterHandler handles deprecated linter replacements.
type deprecatedLinterHandler struct {
	logger *log.Logger
}

// newDeprecatedLinterHandler creates a new deprecated linter handler.
func newDeprecatedLinterHandler(logger *log.Logger) *deprecatedLinterHandler {
	return &deprecatedLinterHandler{logger: logger}
}

// replaceLinters replaces deprecated linters with their successors in the linter set.
func (h *deprecatedLinterHandler) replaceLinters(
	linterSet types.Set[string],
	enabledLinters []string,
	dryRun bool,
	counts *fixCounts,
) types.Set[string] {
	for _, linter := range enabledLinters {
		replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]
		if !isDeprecated {
			continue
		}

		counts.deprecation++

		linterSet.Delete(linter)

		if linterSet.Contains(string(replacement.Replacement)) {
			h.logKeep(linter, replacement.Replacement, dryRun)

			continue
		}

		h.logReplace(linter, replacement, dryRun)

		if !dryRun {
			linterSet.Add(string(replacement.Replacement))
		}
	}

	return linterSet
}

func (h *deprecatedLinterHandler) logKeep(linter string, replacement types.LinterName, dryRun bool) {
	if dryRun {
		h.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)", linter, replacement)
	} else {
		h.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacement)
	}
}

func (h *deprecatedLinterHandler) logReplace(linter string, replacement types.LinterReplacement, dryRun bool) {
	if dryRun {
		h.logger.Infof(
			"[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)",
			linter,
			replacement.Replacement,
			replacement.Reason,
		)
	} else {
		h.logger.Infof(
			"Replacing deprecated linter: %s -> %s (%s)",
			linter,
			replacement.Replacement,
			replacement.Reason,
		)
	}
}

// hasDeprecatedLinters checks if any of the enabled linters are deprecated.
func hasDeprecatedLinters(enabledLinters []string) bool {
	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			return true
		}
	}

	return false
}

// resolveLinterName resolves a linter name, replacing deprecated linters with their successors.
func resolveLinterName(name types.LinterName) string {
	if replacement, isDeprecated := constants.DeprecatedLinters[name]; isDeprecated {
		return string(replacement.Replacement)
	}

	return name.String()
}
