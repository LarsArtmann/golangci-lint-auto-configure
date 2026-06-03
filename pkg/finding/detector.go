package finding

import (
	"context"
	"fmt"

	finding "github.com/larsartmann/go-finding"
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
		return nil, fmt.Errorf("analyze config %s: %w", d.configPath, err)
	}

	findings := make([]finding.Finding, 0, initialFindingsCapacity)

	recs, err := RecommendationsToFindings(analysis.LinterRecommendations, analysis.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("convert recommendations: %w", err)
	}

	findings = append(findings, recs...)

	fmtRecs, err := FormatterRecommendationsToFindings(analysis.FormatterRecommendations, analysis.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("convert formatter recommendations: %w", err)
	}

	findings = append(findings, fmtRecs...)

	depRecs, err := DeprecatedLintersToFindings(analysis.DeprecatedLinters, analysis.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("convert deprecated linters: %w", err)
	}

	findings = append(findings, depRecs...)

	return findings, nil
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
		return nil, fmt.Errorf("convert validation errors: %w", err)
	}

	findings = append(findings, valRecs...)

	return findings, nil
}
