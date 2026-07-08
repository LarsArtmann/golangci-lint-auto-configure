package finding

import (
	"context"

	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const initialFindingsCapacity = 3

// ConfigAnalysisDetector implements pipeline.Detector for golangci-lint config analysis.
// It detects missing linters, deprecated linters, and missing formatters.
type ConfigAnalysisDetector struct {
	analyzer   *linter.Analyzer
	configPath string
	version    string
}

// NewConfigAnalysisDetector creates a detector that analyzes golangci-lint configuration.
func NewConfigAnalysisDetector(analyzer *linter.Analyzer, configPath, version string) *ConfigAnalysisDetector {
	return &ConfigAnalysisDetector{
		analyzer:   analyzer,
		configPath: configPath,
		version:    version,
	}
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

	err = appendDetectorFindings(&findings, analysis.ConfigPath, "detector.convert_recommendations",
		RecommendationsToFindings, analysis.LinterRecommendations)
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
