package linter

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
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

	record, known := constants.PolicyRecord(linter.Name)

	if known && a.skipByPolicy(record) {
		return true
	}

	if a.isLinterRedundant(linter, formatterSet) {
		return true
	}

	return false
}

// skipByPolicy applies the table-driven gates of one known record: management
// tiers, the tool-version gate, and the tech gate.
func (a *Analyzer) skipByPolicy(record types.LinterPolicy) bool {
	switch {
	case record.Tier.Kind == types.TierDisabled:
		a.logger.Debugf("Skipping disabled linter in analysis: %s (%s)", record.Name, record.Tier.Rationale)

		return true
	case record.Tier.Kind == types.TierNeverAutoEnable:
		a.logger.Debugf("Skipping never-auto-enable linter: %s (%s)", record.Name, record.Tier.Rationale)

		return true
	case record.Tier.Kind == types.TierPragmaticNoise && a.pragmatic:
		a.logger.Debugf("Skipping noise linter (--pragmatic): %s (%s)", record.Name, record.Tier.Rationale)

		return true
	case a.isBelowMinVersion(record):
		return true
	case a.isTechGated(record):
		return true
	}

	return false
}

// isBelowMinVersion reports whether the installed golangci-lint predates the
// record's ValidTool lower bound (the linter does not exist there yet).
func (a *Analyzer) isBelowMinVersion(record types.LinterPolicy) bool {
	if record.ValidTool.Min.IsZero() || a.detectedVersion.IsZero() || a.detectedVersion.AtLeast(record.ValidTool.Min) {
		return false
	}

	a.logger.Debugf(
		"Skipping linter %s: requires golangci-lint %s (have %s)",
		record.Name, record.ValidTool.Min, a.detectedVersion,
	)

	return true
}

// isTechGated reports whether the record requires a technology the project
// does not use.
func (a *Analyzer) isTechGated(record types.LinterPolicy) bool {
	if record.Gate.Kind != types.GateRequiresTech || a.hasTechnology(record.Gate.Tech) {
		return false
	}

	a.logger.Debugf("Skipping project-specific linter: %s (project does not use %s)", record.Name, record.Gate.Tech)

	return true
}

func (a *Analyzer) isLinterRedundant(linter types.LinterInfo, formatterSet types.Set[types.FormatterName]) bool {
	for _, relation := range constants.RedundanciesFor(linter.Name) {
		if !formatterSet.Contains(relation.Superseder) {
			continue
		}

		a.logger.Debugf("Skipping redundant linter in analysis: %s (%s)", linter.Name, relation.Reason)

		return true
	}

	return false
}

func (a *Analyzer) makeLinterRecommendation(linter types.LinterInfo) types.LinterRecommendation {
	rec := types.LinterRecommendation{
		Name:   linter.Name,
		Reason: a.getLinterReason(linter.Name),

		Priority: types.LinterPriorityOptional,
	}
	if record, known := constants.PolicyRecord(linter.Name); known && record.Priority != types.LinterPriorityOptional {
		rec.Priority = record.Priority
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

		Priority: types.FormatterPriorityLow,
	}
	if record, known := constants.FormatterPolicyRecord(name); known {
		rec.Priority = record.Priority
	}

	return rec
}

// shouldSkipFormatter returns true if a disabled formatter should not be
// recommended because it is redundant with an enabled formatter, because
// the project does not use the formatter's target technology, or because
// the fixer will not auto-enable it (manual-only).
func (a *Analyzer) shouldSkipFormatter(
	name types.FormatterName,
	enabledSet types.Set[types.FormatterName],
) bool {
	record, known := constants.FormatterPolicyRecord(name)
	if !known {
		return !constants.IsRecommendable(name)
	}

	if record.SupersededBy != "" && enabledSet.Contains(record.SupersededBy) {
		a.logger.Debugf("Skipping redundant formatter: %s (superseded by %s)", name, record.SupersededBy)

		return true
	}

	if record.Gate.Kind == types.GateRequiresTech && !a.hasTechnology(record.Gate.Tech) {
		a.logger.Debugf("Skipping project-specific formatter: %s (project does not use %s)", name, record.Gate.Tech)

		return true
	}

	if !constants.IsRecommendable(name) {
		a.logger.Debugf("Skipping manual-only formatter: %s (fixer will not auto-enable it)", name)

		return true
	}

	return false
}

// hasTechnology checks if the project uses a given technology (swaggo,
// clickhouse, arangodb, go-humanize, etc.) by delegating to the detection package.
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
	case "go-humanize":
		return detector.HasGoHumanize()
	default:
		return true
	}
}

// getFormatterReason returns the human-readable reason for a formatter recommendation.
func (a *Analyzer) getFormatterReason(name types.FormatterName) string {
	if record, known := constants.FormatterPolicyRecord(name); known && record.Rationale != "" {
		return record.Rationale
	}

	return "Formatter is disabled but may be useful"
}

// getLinterReason returns the human-readable reason for a linter recommendation.
func (a *Analyzer) getLinterReason(name types.LinterName) string {
	if record, known := constants.PolicyRecord(name); known && record.Rationale != "" {
		return record.Rationale
	}

	return "Linter is disabled but may be useful"
}

// calculateDeprecatedLinters finds deprecated linters that are enabled.
func (a *Analyzer) calculateDeprecatedLinters(analysis *types.ConfigAnalysis) {
	for _, linter := range analysis.EnabledLinters {
		if linter.Deprecated {
			analysis.DeprecatedLinters = append(analysis.DeprecatedLinters, linter)
		}
	}
}
