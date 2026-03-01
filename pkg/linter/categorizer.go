package linter

import (
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// categorizeLinters categorizes disabled linters by priority.
func (a *Analyzer) categorizeLinters(disabledLinters []types.LinterInfo) []types.LinterRecommendation {
	var recommendations []types.LinterRecommendation

	for _, linter := range disabledLinters {
		// Skip deprecated linters - they shouldn't be recommended
		if linter.Deprecated {
			a.logger.Debugf("Skipping deprecated linter in analysis: %s", linter.Name)

			continue
		}

		rec := types.LinterRecommendation{
			Name:   linter.Name,
			Reason: a.getLinterReason(string(linter.Name)),
		}

		// Get priority from constants, default to Optional if not found
		if priority, ok := constants.LinterPriorities[rec.Name]; ok {
			rec.Priority = priority
		} else {
			rec.Priority = types.LinterPriorityOptional
		}

		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// categorizeFormatters categorizes disabled formatters by priority.
func (a *Analyzer) categorizeFormatters(disabledFormatters []types.FormatterInfo) []types.FormatterRecommendation {
	var recommendations []types.FormatterRecommendation

	for _, formatter := range disabledFormatters {
		name := types.FormatterName(formatter.Name)
		rec := types.FormatterRecommendation{
			Name:   name,
			Reason: a.getFormatterReason(formatter.Name),
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
func (a *Analyzer) getFormatterReason(name string) string {
	if reason, ok := constants.FormatterReasons[types.FormatterName(name)]; ok {
		return reason
	}

	return "Formatter is disabled but may be useful"
}

// getLinterReason returns the human-readable reason for a linter recommendation.
func (a *Analyzer) getLinterReason(name string) string {
	if reason, ok := constants.LinterReasons[types.LinterName(name)]; ok {
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
