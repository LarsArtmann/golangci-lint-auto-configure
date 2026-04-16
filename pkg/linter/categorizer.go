package linter

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// CategorizeLinters categorizes disabled linters by priority.
// exported for testing.
func (a *Analyzer) CategorizeLinters(
	disabledLinters []types.LinterInfo,
	enabledFormatters []types.FormatterInfo,
) []types.LinterRecommendation {
	enabledFormatterSet := types.NewSet[types.FormatterName]()
	for _, formatter := range enabledFormatters {
		enabledFormatterSet.Add(formatter.Name)
	}

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

	if constants.DisabledLinters.Contains(linter.Name) {
		a.logger.Debugf("Skipping explicitly disabled linter in analysis: %s", linter.Name)

		return true
	}

	if mapping, isRedundant := constants.RedundantLinters[linter.Name]; isRedundant {
		if formatterSet.Contains(mapping.Formatter) {
			a.logger.Debugf("Skipping redundant linter in analysis: %s (%s)", linter.Name, mapping.Reason)

			return true
		}
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

// categorizeFormatters categorizes disabled formatters by priority.
func (a *Analyzer) categorizeFormatters(disabledFormatters []types.FormatterInfo) []types.FormatterRecommendation {
	recommendations := make([]types.FormatterRecommendation, 0, len(disabledFormatters))

	for _, formatter := range disabledFormatters {
		name := formatter.Name
		rec := types.FormatterRecommendation{
			Name:   name,
			Reason: a.getFormatterReason(name),
		}

		// Get priority from constants, default to Low if not found
		if priority, ok := constants.FormatterPriorities[name]; ok {
			rec.Priority = priority
		} else {
			rec.Priority = types.FormatterPriorityLow
		}

		recommendations = append(recommendations, rec)
	}

	return recommendations
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
