package linter

import (
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

// EnableCoreFormatters enables the core formatters: gci, gofumpt, goimports.
func (fm *FormatterManager) EnableCoreFormatters(formatterSet types.Set[string], dryRun bool) int {
	coreFormatters := []string{"gci", "gofumpt", "goimports"}
	count := 0

	for _, formatter := range coreFormatters {
		if formatterSet.Contains(formatter) {
			continue
		}

		count++

		if dryRun {
			fm.logger.Debugf("[DRY-RUN] Would enable formatter: %s", formatter)
		} else {
			fm.logger.Debugf("Enabling formatter: %s", formatter)
			formatterSet.Add(formatter)
		}
	}

	return count
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

	if dryRun {
		fm.logger.Debugf("[DRY-RUN] Would enable formatter: golines (formats code and fixes long lines)")
	} else {
		fm.logger.Debugf("Enabling formatter: golines (formats code and fixes long lines)")

		formatterSet.Add("golines")
	}

	return 1
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

	if dryRun {
		fm.logger.Debugf("[DRY-RUN] Would enable formatter: swaggo (detected swaggo usage in project)")
	} else {
		fm.logger.Debugf("Enabling formatter: swaggo (detected swaggo usage in project)")

		formatterSet.Add("swaggo")
	}

	return 1
}

// RemoveRedundantGofmt removes gofmt when gofumpt is enabled (gofumpt is a superset).
func (fm *FormatterManager) RemoveRedundantGofmt(formatterSet types.Set[string], dryRun bool) int {
	if !formatterSet.Contains("gofumpt") || !formatterSet.Contains("gofmt") {
		return 0
	}

	if dryRun {
		fm.logger.Debugf("[DRY-RUN] Would remove redundant formatter: gofmt (gofumpt is enabled and is a superset)")
	} else {
		fm.logger.Debugf("Removing redundant formatter: gofmt (gofumpt is enabled and is a superset)")
		formatterSet.Delete("gofmt")
	}

	return 1
}

// RemoveRedundantLinters removes linters that are superseded by enabled formatters.
func (fm *FormatterManager) RemoveRedundantLinters(
	linterSet types.Set[string],
	formatterSet types.Set[string],
	dryRun bool,
) int {
	count := 0

	for linterName, mapping := range constants.RedundantLinters {
		if !linterSet.Contains(string(linterName)) {
			continue
		}

		if !formatterSet.Contains(string(mapping.Formatter)) {
			continue
		}

		count++

		if dryRun {
			fm.logger.Debugf("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, mapping.Reason)
		} else {
			fm.logger.Debugf("Removing redundant linter: %s (%s)", linterName, mapping.Reason)
			linterSet.Delete(string(linterName))
		}
	}

	return count
}

// ToOrderedSlice converts formatter set to ordered slice.
// Order: gci → goimports → gofumpt → golines → swaggo → others (sorted).
func (fm *FormatterManager) ToOrderedSlice(set types.Set[string]) []string {
	order := []string{"gci", "goimports", "gofumpt", "golines", "swaggo"}

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
