// Package finding converts golangci-lint-auto-configure domain types
// to go-finding's unified Finding model for SARIF, JSON, and pipeline integration.
package finding

import (
	"fmt"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const toolName = "golangci-lint-auto-configure"

// buildFinding is a helper that builds a Finding or panics on invalid state.
// All callers construct valid findings by design.
func buildFinding(b *finding.Builder) finding.Finding {
	f, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("finding builder error: %v", err))
	}

	return f
}

// PriorityToSeverity maps LinterPriority to finding.Severity.
func PriorityToSeverity(p types.LinterPriority) finding.Severity {
	switch p {
	case types.LinterPriorityCritical:
		return finding.SeverityCritical
	case types.LinterPriorityHigh:
		return finding.SeverityError
	case types.LinterPriorityMedium:
		return finding.SeverityWarning
	case types.LinterPriorityOptional:
		return finding.SeverityInfo
	default:
		return finding.SeverityInfo
	}
}

// FormatterPriorityToSeverity maps FormatterPriority to finding.Severity.
func FormatterPriorityToSeverity(p types.FormatterPriority) finding.Severity {
	switch p {
	case types.FormatterPriorityHigh:
		return finding.SeverityError
	case types.FormatterPriorityMedium:
		return finding.SeverityWarning
	case types.FormatterPriorityLow:
		return finding.SeverityInfo
	default:
		return finding.SeverityInfo
	}
}

// RecommendationsToFindings converts LinterRecommendations to Findings.
func RecommendationsToFindings(
	recommendations []types.LinterRecommendation,
	configPath string,
) []finding.Finding {
	result := make([]finding.Finding, 0, len(recommendations))

	for _, rec := range recommendations {
		pos := finding.Position{File: configPath}
		found := buildFinding(finding.NewBuilder(
			"missing-linter",
			toolName,
			fmt.Sprintf("Linter %s is disabled: %s", rec.Name, rec.Reason),
			PriorityToSeverity(rec.Priority),
			pos,
		).
			WithTags(finding.Tag(string(rec.Name))).
			WithCategory(LinterNameToCategory(rec.Name)).
			WithFixStrategy(finding.FixStrategyDirect).
			WithSuggestion(fmt.Sprintf("Enable %s in linters.enable section", rec.Name)))

		result = append(result, found)
	}

	return result
}

// FormatterRecommendationsToFindings converts FormatterRecommendations to Findings.
func FormatterRecommendationsToFindings(
	recommendations []types.FormatterRecommendation,
	configPath string,
) []finding.Finding {
	result := make([]finding.Finding, 0, len(recommendations))

	for _, rec := range recommendations {
		pos := finding.Position{File: configPath}
		found := buildFinding(finding.NewBuilder(
			"missing-formatter",
			toolName,
			fmt.Sprintf("Formatter %s is disabled: %s", rec.Name, rec.Reason),
			FormatterPriorityToSeverity(rec.Priority),
			pos,
		).
			WithTags(finding.Tag(string(rec.Name))).
			WithCategory(finding.CategoryStyle).
			WithFixStrategy(finding.FixStrategyDirect).
			WithSuggestion(fmt.Sprintf("Enable %s in formatters.enable section", rec.Name)))

		result = append(result, found)
	}

	return result
}

// DeprecatedLintersToFindings converts deprecated linter info to Findings.
func DeprecatedLintersToFindings(
	linters []types.LinterInfo,
	configPath string,
) []finding.Finding {
	result := make([]finding.Finding, 0, len(linters))

	for _, linter := range linters {
		pos := finding.Position{File: configPath}
		replacement := "no replacement specified"

		if repl, ok := constants.DeprecatedLinters[linter.Name]; ok {
			replacement = fmt.Sprintf("use %s instead (%s)", repl.Replacement, repl.Reason)
		}

		found := buildFinding(finding.NewBuilder(
			"deprecated-linter",
			toolName,
			fmt.Sprintf("Deprecated linter %s is enabled: %s", linter.Name, replacement),
			finding.SeverityWarning,
			pos,
		).
			WithTags(finding.Tag(string(linter.Name))).
			WithCategory(finding.CategoryMigration).
			WithFixStrategy(finding.FixStrategyDirect).
			WithSuggestion(replacement))

		result = append(result, found)
	}

	return result
}

// ValidationErrorsToFindings converts ValidationErrors to Findings.
func ValidationErrorsToFindings(
	errors []types.ValidationError,
	configPath string,
) []finding.Finding {
	result := make([]finding.Finding, 0, len(errors))

	for _, verr := range errors {
		pos := finding.Position{File: configPath, Line: verr.Line}
		found := buildFinding(finding.NewBuilder(
			"validation-error",
			toolName,
			verr.Message,
			finding.SeverityError,
			pos,
		).
			WithTags(finding.Tag(verr.Field)).
			WithCategory(finding.CategoryConfiguration).
			WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(fmt.Sprintf("Fix field %s: %s", verr.Field, verr.Message)))

		result = append(result, found)
	}

	return result
}

// ErrorsToFindings converts generic errors to Findings.
func ErrorsToFindings(errors []error, configPath string) []finding.Finding {
	result := make([]finding.Finding, 0, len(errors))

	for _, err := range errors {
		pos := finding.Position{File: configPath}
		found := buildFinding(finding.NewBuilder(
			"validation-error",
			toolName,
			err.Error(),
			finding.SeverityError,
			pos,
		).
			WithCategory(finding.CategoryConfiguration))

		result = append(result, found)
	}

	return result
}

// AnalysisToReport converts a full ConfigAnalysis to a finding.Report.
func AnalysisToReport(analysis *types.ConfigAnalysis, version string) *finding.Report {
	report := finding.NewReport(finding.ToolInfo{
		Name:    toolName,
		Version: version,
	})

	report.AddFindings(RecommendationsToFindings(analysis.LinterRecommendations, analysis.ConfigPath))
	report.AddFindings(FormatterRecommendationsToFindings(analysis.FormatterRecommendations, analysis.ConfigPath))
	report.AddFindings(DeprecatedLintersToFindings(analysis.DeprecatedLinters, analysis.ConfigPath))
	report.ComputeSummary()

	return report
}

// AnalysisToSARIF converts a ConfigAnalysis directly to SARIF JSON.
func AnalysisToSARIF(analysis *types.ConfigAnalysis, version string) ([]byte, error) {
	report := AnalysisToReport(analysis, version)

	sarif, err := report.ToSARIF()
	if err != nil {
		return nil, fmt.Errorf("generate SARIF: %w", err)
	}

	return sarif, nil
}
