package linter

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"golang.org/x/mod/semver"
)

func formatterNameSet(formatters []types.FormatterInfo) types.Set[types.FormatterName] {
	formatterSet := types.NewSet[types.FormatterName]()
	for _, formatter := range formatters {
		formatterSet.Add(formatter.Name)
	}

	return formatterSet
}

// CategorizeLinters categorizes disabled linters by priority.
// exported for testing.
func (a *Analyzer) CategorizeLinters(
	disabledLinters []types.LinterInfo,
	enabledFormatters []types.FormatterInfo,
) []types.LinterRecommendation {
	enabledFormatterSet := formatterNameSet(enabledFormatters)

	var recommendations []types.LinterRecommendation

	for _, linter := range disabledLinters {
		if a.shouldSkipLinter(linter, enabledFormatterSet) {
			continue
		}

		recommendations = append(recommendations, a.makeLinterRecommendation(linter))
	}

	return recommendations
}

func (a *Analyzer) shouldSkipLinter(linter types.LinterInfo, formatterSet types.Set[types.FormatterName]) bool {
	if linter.Deprecated {
		a.logger.Debugf("Skipping deprecated linter in analysis: %s", linter.Name)

		return true
	}

	if reason, ok := constants.DisabledLinters[linter.Name]; ok {
		a.logger.Debugf("Skipping disabled linter in analysis: %s (%s)", linter.Name, reason)

		return true
	}

	if a.isNeverAutoEnable(linter) {
		return true
	}

	if a.isPragmaticNoise(linter) || a.isLinterBelowMinVersion(linter) {
		return true
	}

	if a.isLinterRedundant(linter, formatterSet) || a.isLinterProjectSpecific(linter) {
		return true
	}

	return false
}

func (a *Analyzer) isNeverAutoEnable(linter types.LinterInfo) bool {
	if reason, ok := constants.NeverAutoEnableLinters[linter.Name]; ok {
		a.logger.Debugf("Skipping never-auto-enable linter: %s (%s)", linter.Name, reason)

		return true
	}

	return false
}

func (a *Analyzer) isPragmaticNoise(linter types.LinterInfo) bool {
	if !a.pragmatic {
		return false
	}

	if reason, ok := constants.PragmaticNoiseLinters[linter.Name]; ok {
		a.logger.Debugf("Skipping noise linter (--pragmatic): %s (%s)", linter.Name, reason)

		return true
	}

	return false
}

func (a *Analyzer) isLinterRedundant(linter types.LinterInfo, formatterSet types.Set[types.FormatterName]) bool {
	mapping, isRedundant := constants.RedundantLinters[linter.Name]
	if !isRedundant || !formatterSet.Contains(mapping.Formatter) {
		return false
	}

	a.logger.Debugf("Skipping redundant linter in analysis: %s (%s)", linter.Name, mapping.Reason)

	return true
}

func (a *Analyzer) isLinterProjectSpecific(linter types.LinterInfo) bool {
	tech, isSpecific := constants.ProjectSpecificLinters[linter.Name]
	if !isSpecific || a.hasTechnology(tech) {
		return false
	}

	a.logger.Debugf("Skipping project-specific linter: %s (project does not use %s)", linter.Name, tech)

	return true
}

func (a *Analyzer) isLinterBelowMinVersion(linter types.LinterInfo) bool {
	minVer, hasMin := constants.LinterMinVersions[linter.Name]
	if !hasMin {
		return false
	}

	if a.detectedVersion != "" && semver.Compare(a.detectedVersion, minVer) < 0 {
		a.logger.Debugf(
			"Skipping linter %s: requires golangci-lint %s (have %s)",
			linter.Name, minVer, a.detectedVersion,
		)

		return true
	}

	return false
}

func (a *Analyzer) makeLinterRecommendation(linter types.LinterInfo) types.LinterRecommendation {
	rec := types.LinterRecommendation{
		Name:   linter.Name,
		Reason: a.getLinterReason(linter.Name),
	}

	if priority, ok := constants.LinterPriorities[linter.Name]; ok {
		rec.Priority = priority
	} else {
		rec.Priority = types.LinterPriorityOptional
	}

	return rec
}

// CategorizeFormatters categorizes disabled formatters by priority, skipping
// formatters that are redundant (superseded by an enabled formatter) or
// project-specific (only useful when the project uses the corresponding technology).
// exported for testing.
func (a *Analyzer) CategorizeFormatters(
	disabledFormatters []types.FormatterInfo,
	enabledFormatters []types.FormatterInfo,
) []types.FormatterRecommendation {
	enabledSet := formatterNameSet(enabledFormatters)

	recommendations := make([]types.FormatterRecommendation, 0, len(disabledFormatters))

	for _, formatter := range disabledFormatters {
		if a.shouldSkipFormatter(formatter.Name, enabledSet) {
			continue
		}

		recommendations = append(recommendations, a.makeFormatterRecommendation(formatter.Name))
	}

	return recommendations
}

func (a *Analyzer) makeFormatterRecommendation(name types.FormatterName) types.FormatterRecommendation {
	rec := types.FormatterRecommendation{
		Name:   name,
		Reason: a.getFormatterReason(name),
	}

	if priority, ok := constants.FormatterPriorities[name]; ok {
		rec.Priority = priority
	} else {
		rec.Priority = types.FormatterPriorityLow
	}

	return rec
}

// shouldSkipFormatter returns true if a disabled formatter should not be
// recommended because it is redundant with an enabled formatter or because
// the project does not use the formatter's target technology.
func (a *Analyzer) shouldSkipFormatter(
	name types.FormatterName,
	enabledSet types.Set[types.FormatterName],
) bool {
	if superset, isRedundant := constants.RedundantFormatters[name]; isRedundant {
		if enabledSet.Contains(superset) {
			a.logger.Debugf("Skipping redundant formatter: %s (superseded by %s)", name, superset)

			return true
		}
	}

	if tech, isSpecific := constants.ProjectSpecificFormatters[name]; isSpecific {
		if !a.hasTechnology(tech) {
			a.logger.Debugf("Skipping project-specific formatter: %s (project does not use %s)", name, tech)

			return true
		}
	}

	return false
}

// hasTechnology checks if the project uses a given technology (swaggo,
// clickhouse, arangodb, etc.) by delegating to the detection package.
// Returns true when the technology cannot be determined (fail-open)
// to avoid suppressing valid recommendations.
func (a *Analyzer) hasTechnology(tech string) bool {
	if a.projectRoot == "" {
		return true
	}

	detector := detection.NewDetector(a.projectRoot)

	switch tech {
	case "swaggo":
		result, err := detector.HasSwaggo()
		if err != nil {
			return true
		}

		return result
	case "clickhouse":
		return detector.HasClickHouse()
	case "arangodb":
		return detector.HasArangoDB()
	default:
		return true
	}
}

// getFormatterReason returns the human-readable reason for a formatter recommendation.
func (a *Analyzer) getFormatterReason(name types.FormatterName) string {
	if reason, ok := constants.FormatterReasons[name]; ok {
		return reason
	}

	return "Formatter is disabled but may be useful"
}

// getLinterReason returns the human-readable reason for a linter recommendation.
func (a *Analyzer) getLinterReason(name types.LinterName) string {
	if reason, ok := constants.LinterReasons[name]; ok {
		return reason
	}

	return "Linter is disabled but may be useful"
}

// calculateDeprecatedLinters finds deprecated linters that are enabled.
func (a *Analyzer) calculateDeprecatedLinters(analysis *types.ConfigAnalysis) {
	for _, linter := range analysis.EnabledLinters {
		if linter.Deprecated {
			analysis.DeprecatedLinters = append(analysis.DeprecatedLinters, linter)
			analysis.DeprecatedCount++
		}
	}
}

// calculateRecommendationCounts calculates counts by priority level.
func (a *Analyzer) calculateRecommendationCounts(analysis *types.ConfigAnalysis) {
	for _, rec := range analysis.LinterRecommendations {
		switch rec.Priority {
		case types.LinterPriorityCritical:
			analysis.CriticalCount++
		case types.LinterPriorityHigh:
			analysis.HighValueCount++
		case types.LinterPriorityMedium:
			analysis.MediumValueCount++
		case types.LinterPriorityOptional:
			analysis.OptionalCount++
		}
	}
}
