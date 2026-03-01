package linter

// Package linter provides analysis capabilities for golangci-lint configurations.
//
// This file contains the core Analyzer type and main analysis logic.
// Related functionality has been split into separate files:
//   - version_checker.go: Version checking functionality
//

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Analyzer analyzes golangci-lint configurations and provides recommendations.
type Analyzer struct {
	golangciLintPath string
	logger           *log.Logger
}

// NewAnalyzer creates a new linter analyzer.
func NewAnalyzer(logger *log.Logger) *Analyzer {
	return &Analyzer{
		golangciLintPath: "",
		logger:           logger,
	}
}

// golangciLintOutput represents JSON output from golangci-lint linters command.
type golangciLintOutput struct {
	Enabled  []types.LinterInfo `json:"Enabled"`
	Disabled []types.LinterInfo `json:"Disabled"`
}

// golangciLintFormattersOutput represents JSON output from golangci-lint formatters command.
type golangciLintFormattersOutput struct {
	Enabled  []types.FormatterInfo `json:"Enabled"`
	Disabled []types.FormatterInfo `json:"Disabled"`
}

// FindBinary finds the golangci-lint binary in PATH.
func (a *Analyzer) FindBinary() error {
	path, err := exec.LookPath("golangci-lint")
	if err != nil {
		return errors.NewAnalysisError("golangci-lint not found in PATH", "", err)
	}

	a.golangciLintPath = path

	return nil
}

// AnalyzeConfig analyzes the current golangci-lint configuration.
func (a *Analyzer) AnalyzeConfig(configPath string) (*types.ConfigAnalysis, error) {
	if err := a.FindBinary(); err != nil {
		return nil, err
	}

	// Check version meets minimum requirement
	if err := a.CheckVersion(); err != nil {
		return nil, err
	}

	// Analyze linters
	lintOutput, err := a.runLintersCommand(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to run golangci-lint linters", "", err)
	}

	var jsonLinterOutput golangciLintOutput
	if err := json.Unmarshal(lintOutput, &jsonLinterOutput); err != nil {
		return nil, errors.NewAnalysisError("failed to parse golangci-lint linters JSON output", "", err)
	}

	// Analyze formatters
	formatOutput, err := a.runFormattersCommand(configPath)
	if err != nil {
		// Formatters command may not exist in older versions, log but don't fail
		a.logger.Debugf("Formatters analysis skipped: %v", err)

		formatOutput = []byte(`{"Enabled": [], "Disabled": []}`) // Empty output
	}

	var jsonFormatOutput golangciLintFormattersOutput
	if err := json.Unmarshal(formatOutput, &jsonFormatOutput); err != nil {
		// Don't fail if formatters JSON parsing fails
		a.logger.Debugf("Failed to parse formatters JSON, skipping: %v", err)

		jsonFormatOutput = golangciLintFormattersOutput{
			Enabled:  []types.FormatterInfo{},
			Disabled: []types.FormatterInfo{},
		}
	}

	analysis := &types.ConfigAnalysis{
		ConfigPath:               configPath,
		EnabledLinters:           jsonLinterOutput.Enabled,
		DisabledLinters:          jsonLinterOutput.Disabled,
		EnabledFormatters:        jsonFormatOutput.Enabled,
		DisabledFormatters:       jsonFormatOutput.Disabled,
		LinterRecommendations:    a.categorizeLinters(jsonLinterOutput.Disabled),
		FormatterRecommendations: a.categorizeFormatters(jsonFormatOutput.Disabled),
	}

	a.calculateDeprecatedLinters(analysis)
	a.calculateRecommendationCounts(analysis)

	return analysis, nil
}

// runLintersCommand runs `golangci-lint linters` and returns JSON output.
func (a *Analyzer) runLintersCommand(configPath string) ([]byte, error) {
	cmd := exec.Command(a.golangciLintPath, "linters", "--config", configPath, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		a.logger.Debugf("golangci-lint linters command failed: %v", err)
		a.logger.Debugf("Output: %s", string(output))

		return output, errors.NewAnalysisError("golangci-lint linters command failed", "", err)
	}

	return output, nil
}

// runFormattersCommand runs `golangci-lint formatters` and returns JSON output.
func (a *Analyzer) runFormattersCommand(configPath string) ([]byte, error) {
	cmd := exec.Command(a.golangciLintPath, "formatters", "--config", configPath, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Command may not exist in older golangci-lint versions
		return nil, fmt.Errorf("formatters command not available: %w", err)
	}

	return output, nil
}

// categorizeLinters categorizes disabled linters by priority.
func (a *Analyzer) categorizeLinters(disabledLinters []types.LinterInfo) []types.LinterRecommendation {
	var recommendations []types.LinterRecommendation

	for _, linter := range disabledLinters {
		// Skip deprecated linters - they shouldn't be recommended
		if linter.Deprecated {
			a.logger.Debugf("Skipping deprecated linter in analysis: %s", linter.Name)

			continue
		}

		rec := types.LinterRecommendation{
			Name:   linter.Name,
			Reason: a.getLinterReason(string(linter.Name)),
		}

		// Get priority from constants, default to Optional if not found
		if priority, ok := constants.LinterPriorities[rec.Name]; ok {
			rec.Priority = priority
		} else {
			rec.Priority = types.LinterPriorityOptional
		}

		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// categorizeFormatters categorizes disabled formatters by priority.
func (a *Analyzer) categorizeFormatters(disabledFormatters []types.FormatterInfo) []types.FormatterRecommendation {
	var recommendations []types.FormatterRecommendation

	for _, formatter := range disabledFormatters {
		name := types.FormatterName(formatter.Name)
		rec := types.FormatterRecommendation{
			Name:   name,
			Reason: a.getFormatterReason(formatter.Name),
		}

		// Get priority from constants, default to Low if not found
		if priority, ok := constants.FormatterPriorities[name]; ok {
			rec.Priority = priority
		} else {
			rec.Priority = types.FormatterPriorityLow
		}

		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// getFormatterReason returns the human-readable reason for a formatter recommendation.
func (a *Analyzer) getFormatterReason(name string) string {
	if reason, ok := constants.FormatterReasons[types.FormatterName(name)]; ok {
		return reason
	}

	return "Formatter is disabled but may be useful"
}

// getLinterReason returns the human-readable reason for a linter recommendation.
func (a *Analyzer) getLinterReason(name string) string {
	if reason, ok := constants.LinterReasons[types.LinterName(name)]; ok {
		return reason
	}

	return "Linter is disabled but may be useful"
}

// calculateDeprecatedLinters finds deprecated linters that are enabled.
func (a *Analyzer) calculateDeprecatedLinters(analysis *types.ConfigAnalysis) {
	for _, linter := range analysis.EnabledLinters {
		if linter.Deprecated {
			analysis.DeprecatedLinters = append(analysis.DeprecatedLinters, linter)
			analysis.DeprecatedCount++
		}
	}
}

// calculateRecommendationCounts calculates counts by priority level.
func (a *Analyzer) calculateRecommendationCounts(analysis *types.ConfigAnalysis) {
	for _, rec := range analysis.LinterRecommendations {
		switch rec.Priority {
		case types.LinterPriorityCritical:
			analysis.CriticalCount++
		case types.LinterPriorityHigh:
			analysis.HighValueCount++
		case types.LinterPriorityMedium:
			analysis.MediumValueCount++
		case types.LinterPriorityOptional:
			analysis.OptionalCount++
		}
	}
}

// GetLintersByPriority returns recommendations filtered by priority.
func (a *Analyzer) GetLintersByPriority(
	recommendations []types.LinterRecommendation,
	priority types.LinterPriority,
) []types.LinterRecommendation {
	var filtered []types.LinterRecommendation

	for _, rec := range recommendations {
		if rec.Priority == priority {
			filtered = append(filtered, rec)
		}
	}

	return filtered
}

// FormatRecommendations formats recommendations as human-readable output.
func (a *Analyzer) FormatRecommendations(analysis *types.ConfigAnalysis) string {
	var builder strings.Builder

	// Show deprecated linters first (most important to address)
	if len(analysis.DeprecatedLinters) > 0 {
		fmt.Fprintf(&builder, "⚠️  %d DEPRECATED linter(s) are enabled (should be migrated):\n",
			len(analysis.DeprecatedLinters))

		for _, linter := range analysis.DeprecatedLinters {
			// Check if there's a replacement
			if replacement, ok := constants.DeprecatedLinters[types.LinterName(linter.Name)]; ok {
				fmt.Fprintf(&builder, "  - %s: Use %s instead (%s)\n",
					linter.Name,
					replacement.Replacement,
					linter.Description)
			} else {
				fmt.Fprintf(&builder, "  - %s: %s (no replacement specified)\n", linter.Name, linter.Description)
			}
		}

		builder.WriteString("\n")
	}

	critical := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityCritical)
	highValue := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityHigh)
	mediumValue := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityMedium)
	optional := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityOptional)

	if len(critical) > 0 {
		fmt.Fprintf(&builder, "🚨 %d CRITICAL linter(s) are disabled (should ALWAYS be enabled):\n", len(critical))

		for _, rec := range critical {
			fmt.Fprintf(&builder, "  - %s: %s\n", rec.Name, rec.Reason)
		}

		builder.WriteString("\n")
	}

	if len(highValue) > 0 {
		fmt.Fprintf(
			&builder,
			"⚠️  %d HIGH VALUE linter(s) are disabled (recommended for most projects):\n",
			len(highValue),
		)

		for _, rec := range highValue {
			fmt.Fprintf(&builder, "  - %s: %s\n", rec.Name, rec.Reason)
		}

		builder.WriteString("\n")
	}

	if len(mediumValue) > 0 {
		fmt.Fprintf(
			&builder,
			"ℹ️  %d MEDIUM VALUE linter(s) are disabled (optional but recommended):\n",
			len(mediumValue),
		)

		for _, rec := range mediumValue {
			fmt.Fprintf(&builder, "  - %s: %s\n", rec.Name, rec.Reason)
		}

		builder.WriteString("\n")
	}

	if len(optional) > 0 {
		fmt.Fprintf(&builder, "💡 %d OPTIONAL linter(s) are disabled (for niche use cases):\n", len(optional))

		for _, rec := range optional {
			fmt.Fprintf(&builder, "  - %s: %s\n", rec.Name, rec.Reason)
		}
	}

	return builder.String()
}

// GetSummary returns a brief summary of recommendations.
func (a *Analyzer) GetSummary(analysis *types.ConfigAnalysis) string {
	var parts []string

	if analysis.DeprecatedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d DEPRECATED", analysis.DeprecatedCount))
	}

	if analysis.CriticalCount > 0 {
		parts = append(parts, fmt.Sprintf("%d CRITICAL", analysis.CriticalCount))
	}

	if analysis.HighValueCount > 0 {
		parts = append(parts, fmt.Sprintf("%d HIGH", analysis.HighValueCount))
	}

	if analysis.MediumValueCount > 0 {
		parts = append(parts, fmt.Sprintf("%d MEDIUM", analysis.MediumValueCount))
	}

	if analysis.OptionalCount > 0 {
		parts = append(parts, fmt.Sprintf("%d OPTIONAL", analysis.OptionalCount))
	}

	if len(parts) == 0 {
		return "All linters enabled - no recommendations"
	}

	return fmt.Sprintf("Found %d disabled linters: %s (see details above)",
		len(analysis.LinterRecommendations),
		strings.Join(parts, ", "),
	)
}
