// Package finding converts golangci-lint-auto-configure domain types
// to go-finding's unified Finding model for SARIF, JSON, and pipeline integration.
package finding

import (
	"errors"
	"fmt"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const (
	RuleIDMissingLinter    = "missing-linter"
	RuleIDMissingFormatter = "missing-formatter"
	RuleIDDeprecatedLinter = "deprecated-linter"
	RuleIDValidationError  = "validation-error"
	RuleIDGenericError     = "generic-error"
	RuleIDConfigFix        = "config-fix"
)

func linterTag(name string) finding.Tag {
	replacer := strings.NewReplacer("_", "-", ".", "-", "/", "-")

	return finding.Tag(replacer.Replace(name))
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

// filterLinterRecommendationsByPriority returns only recommendations whose
// Priority is at or below maxPriority. This aligns detection with the repairer's
// enableRecommendedLinters filter (which skips rec.Priority > threshold),
// preventing detect→repair loops where unfixable findings persist forever.
func filterLinterRecommendationsByPriority(
	recommendations []types.LinterRecommendation,
	maxPriority types.LinterPriority,
) []types.LinterRecommendation {
	if maxPriority >= types.LinterPriorityOptional {
		return recommendations
	}

	filtered := make([]types.LinterRecommendation, 0, len(recommendations))

	for _, rec := range recommendations {
		if rec.Priority <= maxPriority {
			filtered = append(filtered, rec)
		}
	}

	return filtered
}

// RecommendationsToFindings converts LinterRecommendations to Findings.
func RecommendationsToFindings(
	recommendations []types.LinterRecommendation,
	configPath string,
) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(recommendations))

	for _, rec := range recommendations {
		found, err := configFinding(configFindingParams{
			RuleID:     RuleIDMissingLinter,
			Message:    fmt.Sprintf("Linter %s is disabled: %s", rec.Name, rec.Reason),
			Severity:   PriorityToSeverity(rec.Priority),
			Position:   configPosition(configPath, 0),
			Category:   LinterNameToCategory(rec.Name),
			Tags:       []finding.Tag{linterTag(string(rec.Name))},
			Suggestion: fmt.Sprintf("Enable %s in linters.enable section", rec.Name),
		})
		if err != nil {
			return nil, errorfamily.WrapCorruptionf(err, "converter.linter_finding",
				"build finding for linter %s", rec.Name)
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
		found, err := configFinding(configFindingParams{
			RuleID:     RuleIDMissingFormatter,
			Message:    fmt.Sprintf("Formatter %s is disabled: %s", rec.Name, rec.Reason),
			Severity:   FormatterPriorityToSeverity(rec.Priority),
			Position:   configPosition(configPath, 0),
			Category:   finding.CategoryStyle,
			Tags:       []finding.Tag{linterTag(string(rec.Name))},
			Suggestion: fmt.Sprintf("Enable %s in formatters.enable section", rec.Name),
		})
		if err != nil {
			return nil, errorfamily.WrapCorruptionf(err, "converter.formatter_finding",
				"build finding for formatter %s", rec.Name)
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
		f, err := deprecatedLinterFinding(linter, configPath)
		if err != nil {
			return nil, err
		}

		result = append(result, f)
	}

	return result, nil
}

func deprecatedLinterFinding(linter types.LinterInfo, configPath string) (finding.Finding, error) {
	replacement := "no replacement specified"

	if repl, ok := constants.DeprecatedLinters[linter.Name]; ok {
		replacement = fmt.Sprintf("use %s instead (%s)", repl.Replacement, repl.Reason)
	}

	return configFinding(configFindingParams{
		RuleID:     RuleIDDeprecatedLinter,
		Message:    fmt.Sprintf("Deprecated linter %s is enabled: %s", linter.Name, replacement),
		Severity:   finding.SeverityWarning,
		Position:   configPosition(configPath, 0),
		Category:   finding.CategoryMigration,
		Tags:       []finding.Tag{linterTag(string(linter.Name))},
		Suggestion: replacement,
	})
}

// ValidationErrorsToFindings converts ValidationErrors to Findings.
func ValidationErrorsToFindings(
	errors []types.ValidationError,
	configPath string,
) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(errors))

	for _, verr := range errors {
		found, err := configFinding(configFindingParams{
			RuleID:     RuleIDValidationError,
			Message:    verr.Message,
			Severity:   finding.SeverityError,
			Position:   configPosition(configPath, verr.Line),
			Category:   finding.CategoryConfiguration,
			Tags:       []finding.Tag{linterTag(verr.Field)},
			Suggestion: fmt.Sprintf("Fix field %s: %s", verr.Field, verr.Message),
		})
		if err != nil {
			return nil, errorfamily.WrapCorruptionf(err, "converter.validation_finding",
				"build finding for validation error %s", verr.Field)
		}

		result = append(result, found)
	}

	return result, nil
}

// ErrorsToFindings converts generic errors to Findings.
func ErrorsToFindings(errors []error, configPath string) ([]finding.Finding, error) {
	result := make([]finding.Finding, 0, len(errors))

	for _, err := range errors {
		found, buildErr := configFinding(configFindingParams{
			RuleID:     RuleIDGenericError,
			Message:    err.Error(),
			Severity:   finding.SeverityError,
			Position:   configPosition(configPath, 0),
			Category:   finding.CategoryConfiguration,
			Tags:       []finding.Tag{},
			Suggestion: "",
		})
		if buildErr != nil {
			return nil, errorfamily.WrapCorruptionf(buildErr, "converter.error_finding",
				"build finding for error %q", err.Error())
		}

		result = append(result, found)
	}

	return result, nil
}

// AnalysisToReport converts a full ConfigAnalysis to a finding.Report.
func AnalysisToReport(analysis *types.ConfigAnalysis, version string) (*finding.Report, error) {
	report := finding.NewReport(finding.ToolInfo{
		Name:    constants.ToolName,
		Version: version,
	})

	errs := collectAnalysisFindings(report, analysis)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	report.ComputeSummary()

	return report, nil
}

// collectAnalysisFindings adds all analysis findings to the report and returns any errors.
func collectAnalysisFindings(report *finding.Report, analysis *types.ConfigAnalysis) []error {
	var errs []error

	recs, err := RecommendationsToFindings(analysis.LinterRecommendations, analysis.ConfigPath)
	errs = appendFindingBatch(report, errs, recs, err,
		"converter.recommendations", "convert recommendations")

	fmtRecs, err := FormatterRecommendationsToFindings(analysis.FormatterRecommendations, analysis.ConfigPath)
	errs = appendFindingBatch(report, errs, fmtRecs, err,
		"converter.formatter_recommendations", "convert formatter recommendations")

	depRecs, err := DeprecatedLintersToFindings(analysis.DeprecatedLinters, analysis.ConfigPath)
	errs = appendFindingBatch(report, errs, depRecs, err,
		"converter.deprecated", "convert deprecated linters")

	return errs
}

// appendFindingBatch adds findings to the report and collects any conversion
// error, wrapped with the given code and message.
func appendFindingBatch(
	report *finding.Report,
	errs []error,
	findings []finding.Finding,
	err error,
	code, msg string,
) []error {
	if err != nil {
		return append(errs, errorfamily.WrapCorruption(err, code, msg))
	}

	report.AddFindings(findings)

	return errs
}

// AnalysisToSARIF converts a ConfigAnalysis directly to SARIF JSON.
func AnalysisToSARIF(analysis *types.ConfigAnalysis, version string) ([]byte, error) {
	report, err := AnalysisToReport(analysis, version)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "converter.build_report", "build report")
	}

	sarif, err := report.ToSARIF()
	if err != nil {
		return nil, errorfamily.WrapCorruptionf(err, "converter.sarif",
			"generate SARIF (version=%s)", version)
	}

	return sarif, nil
}
