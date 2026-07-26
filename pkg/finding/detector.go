package finding

import (
	"context"

	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ConfigAnalyzer is the narrowest interface the detector needs from the linter analyzer.
// This decouples pkg/finding from pkg/linter.
type ConfigAnalyzer interface {
	AnalyzeConfig(ctx context.Context, configPath string) (*types.ConfigAnalysis, error)
}

const initialFindingsCapacity = 3

// ConfigAnalysisDetector implements pipeline.Detector for golangci-lint config analysis.
// It detects missing linters, deprecated linters, and missing formatters.
type ConfigAnalysisDetector struct {
	analyzer   ConfigAnalyzer
	configPath string
	version    string
	priority   types.LinterPriority
}

// NewConfigAnalysisDetector creates a detector that analyzes golangci-lint configuration.
// The detector reports findings for all recommendations regardless of priority
// (equivalent to LinterPriorityOptional). Use WithPriority to align detection
// with a repairer's priority threshold, preventing detect→repair loops where
// the detector reports findings the repairer intentionally won't fix.
func NewConfigAnalysisDetector(analyzer ConfigAnalyzer, configPath, version string) *ConfigAnalysisDetector {
	return &ConfigAnalysisDetector{
		analyzer:   analyzer,
		configPath: configPath,
		version:    version,
		priority:   types.LinterPriorityOptional,
	}
}

// WithPriority sets the maximum priority for linter recommendations reported
// as findings. Recommendations with Priority > maxPriority are suppressed,
// aligning detection with a repairer that only fixes recommendations at or
// below the same threshold. This prevents detect→repair loops where findings
// the repairer intentionally skips reappear on every re-detect.
// Returns the detector for chaining.
func (d *ConfigAnalysisDetector) WithPriority(maxPriority types.LinterPriority) *ConfigAnalysisDetector {
	d.priority = maxPriority

	return d
}

// Name returns the detector name.
func (d *ConfigAnalysisDetector) Name() string {
	return "config-analysis"
}

// Detect runs config analysis and returns findings for missing/deprecated linters.
func (d *ConfigAnalysisDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	analysis, err := d.analyzer.AnalyzeConfig(ctx, d.configPath)
	if err != nil {
		return nil, apperrors.WrapClassifiedf(err, "detector.analyze",
			"analyze config %s", d.configPath)
	}

	findings := make([]finding.Finding, 0, initialFindingsCapacity)

	filteredLinterRecs := filterLinterRecommendationsByPriority(analysis.LinterRecommendations, d.priority)

	err = appendDetectorFindings(&findings, analysis.ConfigPath, "detector.convert_recommendations",
		RecommendationsToFindings, filteredLinterRecs)
	if err != nil {
		return nil, err
	}

	err = appendDetectorFindings(&findings, analysis.ConfigPath, "detector.convert_formatters",
		FormatterRecommendationsToFindings, analysis.FormatterRecommendations)
	if err != nil {
		return nil, err
	}

	err = appendDetectorFindings(&findings, analysis.ConfigPath, "detector.convert_deprecated",
		DeprecatedLintersToFindings, analysis.DeprecatedLinters)
	if err != nil {
		return nil, err
	}

	return findings, nil
}

// convertFunc is a generic converter that transforms domain data into Findings.
type convertFunc[T any] func([]T, string) ([]finding.Finding, error)

// appendDetectorFindings converts domain data to Findings and appends them.
// On error, wraps with the given code as Corruption.
func appendDetectorFindings[T any](
	out *[]finding.Finding, configPath, code string,
	convert convertFunc[T], items []T,
) error {
	recs, err := convert(items, configPath)
	if err != nil {
		return errorfamily.WrapCorruption(err, code, "convert findings")
	}

	*out = append(*out, recs...)

	return nil
}

// DetectWithValidation also includes validation errors as findings.
func (d *ConfigAnalysisDetector) DetectWithValidation(
	ctx context.Context,
	validationErrors []types.ValidationError,
) ([]finding.Finding, error) {
	findings, err := d.Detect(ctx)
	if err != nil {
		return nil, err
	}

	valRecs, err := ValidationErrorsToFindings(validationErrors, d.configPath)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "detector.convert_validation",
			"convert validation errors")
	}

	findings = append(findings, valRecs...)

	return findings, nil
}
