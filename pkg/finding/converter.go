// Package finding converts golangci-lint-auto-configure domain types
// to go-finding's unified Finding model for SARIF, JSON, and pipeline integration.
package finding

import (
	"errors"
	"fmt"
	"strings"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const toolName = "golangci-lint-auto-configure"

const (
	RuleIDMissingLinter    = "missing-linter"
	RuleIDMissingFormatter = "missing-formatter"
	RuleIDDeprecatedLinter = "deprecated-linter"
	RuleIDValidationError  = "validation-error"
	RuleIDGenericError     = "generic-error"
	RuleIDConfigFix        = "config-fix"
)

var linterTagReplacer = strings.NewReplacer("_", "-", ".", "-", "/", "-")

func linterTag(name string) finding.Tag {
	return finding.Tag(linterTagReplacer.Replace(name))
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
		found, err := configFinding(configFindingParams{
			RuleID:     RuleIDMissingLinter,
			Message:    fmt.Sprintf("Linter %s is disabled: %s", rec.Name, rec.Reason),
			Severity:   PriorityToSeverity(rec.Priority),
			Position:   finding.Position{File: configPath},
			Category:   LinterNameToCategory(rec.Name),
			Tags:       []finding.Tag{linterTag(string(rec.Name))},
			Suggestion: fmt.Sprintf("Enable %s in linters.enable section", rec.Name),
		})
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
		found, err := configFinding(configFindingParams{
			RuleID:     RuleIDMissingFormatter,
			Message:    fmt.Sprintf("Formatter %s is disabled: %s", rec.Name, rec.Reason),
			Severity:   FormatterPriorityToSeverity(rec.Priority),
			Position:   finding.Position{File: configPath},
			Category:   finding.CategoryStyle,
			Tags:       []finding.Tag{linterTag(string(rec.Name))},
			Suggestion: fmt.Sprintf("Enable %s in formatters.enable section", rec.Name),
		})
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
		Position:   finding.Position{File: configPath},
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
			Position:   finding.Position{File: configPath, Line: verr.Line},
			Category:   finding.CategoryConfiguration,
			Tags:       []finding.Tag{linterTag(verr.Field)},
			Suggestion: fmt.Sprintf("Fix field %s: %s", verr.Field, verr.Message),
		})
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
		found, buildErr := configFinding(configFindingParams{
			RuleID:     RuleIDGenericError,
			Message:    err.Error(),
			Severity:   finding.SeverityError,
			Position:   finding.Position{File: configPath},
			Category:   finding.CategoryConfiguration,
			Tags:       []finding.Tag{},
			Suggestion: "",
		})
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

	var errs []error

	recs, err := RecommendationsToFindings(analysis.LinterRecommendations, analysis.ConfigPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("convert recommendations: %w", err))
	}

	report.AddFindings(recs)

	fmtRecs, err := FormatterRecommendationsToFindings(analysis.FormatterRecommendations, analysis.ConfigPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("convert formatter recommendations: %w", err))
	}

	report.AddFindings(fmtRecs)

	depRecs, err := DeprecatedLintersToFindings(analysis.DeprecatedLinters, analysis.ConfigPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("convert deprecated linters: %w", err))
	}

	report.AddFindings(depRecs)

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

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
