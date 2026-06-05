// Package finding converts golangci-lint-auto-configure domain types
// to go-finding's unified Finding model for SARIF, JSON, and pipeline integration.
package finding

import (
	"fmt"
	"strings"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const toolName = "golangci-lint-auto-configure"

// linterTag sanitizes a string into a valid go-finding Tag.
// Tags must be lowercase-hyphenated (first char must be a letter).
func linterTag(name string) finding.Tag {
	r := strings.NewReplacer("_", "-", ".", "-", "/", "-")

	return finding.Tag(r.Replace(name))
}

// buildFinding is a helper that builds a Finding from a Builder, returning an error
// instead of panicking on invalid builder state.
func buildFinding(b *finding.Builder) (finding.Finding, error) {
	f, err := b.Build()
	if err != nil {
		return finding.Finding{}, fmt.Errorf("finding builder error: %w", err)
	}

	return f, nil
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
) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(recommendations))

	for _, rec := range recommendations {
		pos := finding.Position{File: configPath}

		found, err := buildFinding(finding.NewBuilder(
			"missing-linter",
			toolName,
			fmt.Sprintf("Linter %s is disabled: %s", rec.Name, rec.Reason),
			PriorityToSeverity(rec.Priority),
			pos,
		).
			WithTags(linterTag(string(rec.Name))).
			WithCategory(LinterNameToCategory(rec.Name)).
			WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(fmt.Sprintf("Enable %s in linters.enable section", rec.Name)))
		if err != nil {
			return nil, fmt.Errorf("build finding for linter %s: %w", rec.Name, err)
		}

		result = append(result, found)
	}

	return result, nil
}

// FormatterRecommendationsToFindings converts FormatterRecommendations to Findings.
func FormatterRecommendationsToFindings(
	recommendations []types.FormatterRecommendation,
	configPath string,
) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(recommendations))

	for _, rec := range recommendations {
		pos := finding.Position{File: configPath}

		found, err := buildFinding(finding.NewBuilder(
			"missing-formatter",
			toolName,
			fmt.Sprintf("Formatter %s is disabled: %s", rec.Name, rec.Reason),
			FormatterPriorityToSeverity(rec.Priority),
			pos,
		).
			WithTags(linterTag(string(rec.Name))).
			WithCategory(finding.CategoryStyle).
			WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(fmt.Sprintf("Enable %s in formatters.enable section", rec.Name)))
		if err != nil {
			return nil, fmt.Errorf("build finding for formatter %s: %w", rec.Name, err)
		}

		result = append(result, found)
	}

	return result, nil
}

// DeprecatedLintersToFindings converts deprecated linter info to Findings.
func DeprecatedLintersToFindings(
	linters []types.LinterInfo,
	configPath string,
) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(linters))

	for _, linter := range linters {
		found, err := deprecatedLinterFinding(linter, configPath)
		if err != nil {
			return nil, err
		}

		result = append(result, found)
	}

	return result, nil
}

func deprecatedLinterFinding(linter types.LinterInfo, configPath string) (finding.Finding, error) {
	replacement := "no replacement specified"

	if repl, ok := constants.DeprecatedLinters[linter.Name]; ok {
		replacement = fmt.Sprintf("use %s instead (%s)", repl.Replacement, repl.Reason)
	}

	found, err := buildFinding(finding.NewBuilder(
		"deprecated-linter",
		toolName,
		fmt.Sprintf("Deprecated linter %s is enabled: %s", linter.Name, replacement),
		finding.SeverityWarning,
		finding.Position{File: configPath},
	).
		WithTags(linterTag(string(linter.Name))).
		WithCategory(finding.CategoryMigration).
		WithFixStrategy(finding.FixStrategySuggest).
		WithSuggestion(replacement))
	if err != nil {
		return finding.Finding{}, fmt.Errorf("build finding for deprecated linter %s: %w", linter.Name, err)
	}

	return found, nil
}

// ValidationErrorsToFindings converts ValidationErrors to Findings.
func ValidationErrorsToFindings(
	errors []types.ValidationError,
	configPath string,
) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(errors))

	for _, verr := range errors {
		pos := finding.Position{File: configPath, Line: verr.Line}

		found, err := buildFinding(finding.NewBuilder(
			"validation-error",
			toolName,
			verr.Message,
			finding.SeverityError,
			pos,
		).
			WithTags(linterTag(verr.Field)).
			WithCategory(finding.CategoryConfiguration).
			WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(fmt.Sprintf("Fix field %s: %s", verr.Field, verr.Message)))
		if err != nil {
			return nil, fmt.Errorf("build finding for validation error %s: %w", verr.Field, err)
		}

		result = append(result, found)
	}

	return result, nil
}

// ErrorsToFindings converts generic errors to Findings.
func ErrorsToFindings(errors []error, configPath string) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(errors))

	for _, err := range errors {
		pos := finding.Position{File: configPath}

		found, buildErr := buildFinding(finding.NewBuilder(
			"validation-error",
			toolName,
			err.Error(),
			finding.SeverityError,
			pos,
		).
			WithCategory(finding.CategoryConfiguration))
		if buildErr != nil {
			return nil, fmt.Errorf("build finding for error %q: %w", err.Error(), buildErr)
		}

		result = append(result, found)
	}

	return result, nil
}

// AnalysisToReport converts a full ConfigAnalysis to a finding.Report.
func AnalysisToReport(analysis *types.ConfigAnalysis, version string) (*finding.Report, error) {
	report := finding.NewReport(finding.ToolInfo{
		Name:    toolName,
		Version: version,
	})

	recs, err := RecommendationsToFindings(analysis.LinterRecommendations, analysis.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("convert recommendations: %w", err)
	}

	report.AddFindings(recs)

	fmtRecs, err := FormatterRecommendationsToFindings(analysis.FormatterRecommendations, analysis.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("convert formatter recommendations: %w", err)
	}

	report.AddFindings(fmtRecs)

	depRecs, err := DeprecatedLintersToFindings(analysis.DeprecatedLinters, analysis.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("convert deprecated linters: %w", err)
	}

	report.AddFindings(depRecs)
	report.ComputeSummary()

	return report, nil
}

// AnalysisToSARIF converts a ConfigAnalysis directly to SARIF JSON.
func AnalysisToSARIF(analysis *types.ConfigAnalysis, version string) ([]byte, error) {
	report, err := AnalysisToReport(analysis, version)
	if err != nil {
		return nil, fmt.Errorf("build report: %w", err)
	}

	sarif, err := report.ToSARIF()
	if err != nil {
		return nil, fmt.Errorf("generate SARIF (version=%s): %w", version, err)
	}

	return sarif, nil
}
