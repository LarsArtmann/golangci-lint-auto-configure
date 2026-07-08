package finding

import (
	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// FindingsToLSP converts Findings to LSP Diagnostics.
func FindingsToLSP(findings []finding.Finding) []finding.LSPDiagnostic {
	diagnostics := make([]finding.LSPDiagnostic, 0, len(findings))

	for idx := range findings {
		diagnostics = append(diagnostics, findings[idx].ToLSP())
	}

	return diagnostics
}

// FilterByPriority filters findings by linter priority using the priority mapping.
func FilterByPriority(findings []finding.Finding, priority types.LinterPriority) []finding.Finding {
	severity := PriorityToSeverity(priority)

	return finding.Filter(findings, finding.BySeverityAtLeast(severity))
}

// MergeReports merges multiple finding Reports with deduplication.
func MergeReports(reports []*finding.Report) *finding.Report {
	return finding.Combine(reports, finding.WithDeduplication(true))
}

// AnalysisFindingsByFile groups analysis findings by file path.
func AnalysisFindingsByFile(
	analysis *types.ConfigAnalysis,
	version string,
) (map[finding.FilePath][]finding.Finding, error) {
	report, err := AnalysisToReport(analysis, version)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "report.build_by_file", "build report")
	}

	return finding.GroupByFile(report.FindingsSnapshot()), nil
}

// AnalysisFindingsByCategory groups analysis findings by category.
func AnalysisFindingsByCategory(
	analysis *types.ConfigAnalysis, version string,
) (map[finding.Category][]finding.Finding, error) {
	report, err := AnalysisToReport(analysis, version)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "report.build_by_category", "build report")
	}

	return finding.GroupByCategory(report.FindingsSnapshot()), nil
}

// SeverityFromHealthSeverity converts a types.HealthSeverity to the corresponding
// go-finding Severity. Warnings map to error severity (the highest non-critical).
func SeverityFromHealthSeverity(sev types.HealthSeverity) finding.Severity {
	switch sev {
	case types.HealthSeverityCritical:
		return finding.SeverityCritical
	case types.HealthSeverityWarning:
		return finding.SeverityError
	case types.HealthSeverityInfo:
		return finding.SeverityInfo
	default:
		return finding.SeverityWarning
	}
}

// AutoFixableFindings returns only findings with FixStrategyDirect.
func AutoFixableFindings(analysis *types.ConfigAnalysis, version string) ([]finding.Finding, error) {
	report, err := AnalysisToReport(analysis, version)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "report.build_autofixable", "build report")
	}

	return report.ByFixStrategy(finding.FixStrategyDirect), nil
}
