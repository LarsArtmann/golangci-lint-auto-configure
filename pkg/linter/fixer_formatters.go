package linter

import (
	"fmt"
	"path/filepath"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// FormatterManager handles formatter-related operations.
type FormatterManager struct {
	logger *log.Logger
}

// NewFormatterManager creates a new formatter manager.
func NewFormatterManager(logger *log.Logger) *FormatterManager {
	return &FormatterManager{logger: logger}
}

// EnableCoreFormatters enables the core formatters: gci, goimports, gofumpt, golines.
func (fm *FormatterManager) EnableCoreFormatters(formatterSet types.Set[string], dryRun bool) int {
	coreFormatters := constants.CoreFormatters
	count := 0

	for _, formatter := range coreFormatters {
		if formatterSet.Contains(formatter) {
			continue
		}

		count++

		fm.logFormatterChange(formatter, "enabling", dryRun)

		if !dryRun {
			formatterSet.Add(formatter)
		}
	}

	return count
}

func (fm *FormatterManager) logFormatterChange(name, action string, dryRun bool) {
	fm.logChange(name, "formatter", action, "", dryRun)
}

// EnableGolinesFormatter enables the golines formatter if recommended at high priority.
func (fm *FormatterManager) EnableGolinesFormatter(
	formatterSet types.Set[string],
	analysis *types.ConfigAnalysis,
	dryRun bool,
) int {
	shouldEnable := false

	for _, rec := range analysis.FormatterRecommendations {
		if rec.Name == "golines" && rec.Priority == types.FormatterPriorityHigh {
			shouldEnable = true

			break
		}
	}

	if !shouldEnable || formatterSet.Contains("golines") {
		return 0
	}

	return fm.addFormatter(formatterSet, "golines", "formats code and fixes long lines", dryRun)
}

func (fm *FormatterManager) addFormatter(set types.Set[string], name, reason string, dryRun bool) int {
	fm.logChange(name, "formatter", "enabling", reason, dryRun)

	if !dryRun {
		set.Add(name)
	}

	return 1
}

func (fm *FormatterManager) logChange(name, entityType, action, reason string, dryRun bool) {
	if reason != "" {
		reason = fmt.Sprintf(" (%s)", reason)
	}

	if dryRun {
		fm.logger.Debugf("[DRY-RUN] Would %s %s: %s%s", action, entityType, name, reason)
	} else {
		fm.logger.Debugf("%s %s: %s%s", action, entityType, name, reason)
	}
}

// EnableSwaggoFormatter enables the swaggo formatter if swaggo is detected in the project.
func (fm *FormatterManager) EnableSwaggoFormatter(
	formatterSet types.Set[string],
	configPath string,
	dryRun bool,
) int {
	if formatterSet.Contains("swaggo") {
		return 0
	}

	if !fm.projectUsesSwaggo(configPath) {
		return 0
	}

	return fm.addFormatter(formatterSet, "swaggo", "detected swaggo usage in project", dryRun)
}

// RemoveRedundantGofmt removes gofmt when gofumpt is enabled (gofumpt is a superset).
func (fm *FormatterManager) RemoveRedundantGofmt(formatterSet types.Set[string], dryRun bool) int {
	if !formatterSet.Contains("gofumpt") || !formatterSet.Contains("gofmt") {
		return 0
	}

	fm.logFormatterChange("gofmt", "removing redundant", dryRun)

	if !dryRun {
		formatterSet.Delete("gofmt")
	}

	return 1
}

// RemoveRedundantLinters removes linters that are superseded by enabled formatters.
func (fm *FormatterManager) RemoveRedundantLinters(
	linterSet types.Set[types.LinterName],
	formatterSet types.Set[string],
	dryRun bool,
) int {
	count := 0

	for linterName, mapping := range constants.RedundantLinters {
		if !linterSet.Contains(linterName) {
			continue
		}

		if !formatterSet.Contains(string(mapping.Formatter)) {
			continue
		}

		count++

		fm.logLinterChange(string(linterName), "removing redundant", mapping.Reason, dryRun)

		if !dryRun {
			linterSet.Delete(linterName)
		}
	}

	return count
}

func (fm *FormatterManager) logLinterChange(name, action, reason string, dryRun bool) {
	fm.logChange(name, "linter", action, reason, dryRun)
}

// ToOrderedSlice converts formatter set to ordered slice.
// Order: gci → goimports → gofumpt → golines → swaggo → others (sorted).
func (fm *FormatterManager) ToOrderedSlice(set types.Set[string]) []string {
	order := constants.FormatterOrder

	result := make([]string, 0, set.Len())
	remaining := make([]string, 0)

	for _, name := range order {
		if set.Contains(name) {
			result = append(result, name)
		}
	}

	for name := range set {
		if !slices.Contains(order, name) {
			remaining = append(remaining, name)
		}
	}

	slices.Sort(remaining)

	return append(result, remaining...)
}

func (fm *FormatterManager) projectUsesSwaggo(configPath string) bool {
	rootDir := filepath.Dir(configPath)
	detector := detection.NewDetector(rootDir)

	hasSwaggo, err := detector.HasSwaggo()
	if err != nil {
		fm.logger.Debugf("Error detecting swaggo: %v", err)

		return false
	}

	return hasSwaggo
}
