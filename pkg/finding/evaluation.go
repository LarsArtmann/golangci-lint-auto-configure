package finding

import (
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
)

// EvaluationToFindings converts one evaluation into findings: one per
// deviation, plus a suppressed-re-enable finding per never-enable block.
func EvaluationToFindings(evaluation *linter.Evaluation) ([]finding.Finding, error) {
	findings, err := DeviationsToFindings(evaluation.Deviations, evaluation.ConfigPath)
	if err != nil {
		return nil, err
	}

	for _, name := range evaluation.Suppressed {
		suppressed, buildErr := configFinding(configFindingParams{
			RuleID:   RuleIDSuppressedReEnable,
			Message:  fmt.Sprintf("Re-enabling %s was suppressed by a never-enable sidecar entry", name),
			Severity: finding.SeverityInfo,
			Position: configPosition(evaluation.ConfigPath, 0),
			Category: finding.CategoryConfiguration,
			Tags:     deviationTags(string(name)),
			Suggestion: "Keep the linter in the sidecar's never-enable section if the " +
				"disable is intentional; delete the entry to let enforcement re-enable it",
		})
		if buildErr != nil {
			return nil, errorfamily.WrapCorruptionf(buildErr, "evaluation.suppressed_finding",
				"build finding for suppressed linter %s", name)
		}

		findings = append(findings, suppressed)
	}

	return findings, nil
}

// RuleIDSuppressedReEnable reports a never-enable suppression (no config
// change accompanies it).
const RuleIDSuppressedReEnable = "suppressed-re-enable"

// EvaluationToReport converts one evaluation to a finding.Report.
func EvaluationToReport(evaluation *linter.Evaluation, version string) (*finding.Report, error) {
	report := finding.NewReport(finding.ToolInfo{
		Name:    constants.ToolName,
		Version: version,
	})

	findings, err := EvaluationToFindings(evaluation)
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "evaluation.build_report", "build evaluation report")
	}

	report.AddFindings(findings)
	report.ComputeSummary()

	return report, nil
}

// EvaluationToSARIF converts one evaluation directly to SARIF JSON.
func EvaluationToSARIF(evaluation *linter.Evaluation, version string) ([]byte, error) {
	report, err := EvaluationToReport(evaluation, version)
	if err != nil {
		return nil, err
	}

	sarif, err := report.ToSARIF()
	if err != nil {
		return nil, errorfamily.WrapCorruptionf(err, "evaluation.sarif",
			"generate SARIF (version=%s)", version)
	}

	return sarif, nil
}
