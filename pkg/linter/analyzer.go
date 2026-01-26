package linter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"golang.org/x/mod/semver"
)

// Analyzer analyzes golangci-lint configurations and provides recommendations
type Analyzer struct {
	golangciLintPath string
	logger           *log.Logger
}

// NewAnalyzer creates a new linter analyzer
func NewAnalyzer(logger *log.Logger) *Analyzer {
	return &Analyzer{
		golangciLintPath: "",
		logger:           logger,
	}
}

// golangciLintOutput represents JSON output from golangci-lint linters/formatters commands
type golangciLintOutput struct {
	Enabled  []types.LinterInfo `json:"Enabled"`
	Disabled []types.LinterInfo `json:"Disabled"`
}

// golangciLintVersion represents JSON output from `golangci-lint version --json`
type golangciLintVersion struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

// FindBinary finds the golangci-lint binary in PATH
func (a *Analyzer) FindBinary() error {
	path, err := exec.LookPath("golangci-lint")
	if err != nil {
		return errors.NewAnalysisError("golangci-lint not found in PATH", "", err)
	}
	a.golangciLintPath = path
	return nil
}

// CheckVersion verifies golangci-lint is at least v2.8.0
func (a *Analyzer) CheckVersion() error {
	// Try JSON output first (more reliable)
	cmd := exec.Command(a.golangciLintPath, "version", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to text parsing if --json not supported
		return a.checkVersionText()
	}
	
	// Parse JSON output
	var versionInfo golangciLintVersion
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		// JSON parsing failed, fall back to text parsing
		a.logger.Debugf("Failed to parse JSON version output, falling back to text: %v", err)
		return a.checkVersionText()
	}
	
	if versionInfo.Version == "" {
		return errors.NewAnalysisError("could not parse golangci-lint version from JSON", "", fmt.Errorf("output: %s", string(output)))
	}
	
	version := versionInfo.Version
	
	// Ensure version has 'v' prefix for semver
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	
	// Validate semver format
	if !semver.IsValid(version) {
		return errors.NewAnalysisError("invalid golangci-lint version format", "", fmt.Errorf("version: %s", version))
	}
	
	// Compare with minimum required version (v2.8.0)
	minVersion := "v2.8.0"
	if semver.Compare(version, minVersion) < 0 {
		return errors.NewAnalysisError(
			fmt.Sprintf("golangci-lint version %s is too old", version),
			"",
			fmt.Errorf("minimum required version is %s. Please upgrade: https://golangci-lint.run/usage/install/", minVersion),
		)
	}
	
	a.logger.Debugf("golangci-lint version %s (>= %s) ✓", version, minVersion)
	return nil
}

// checkVersionText is a fallback that parses text output from golangci-lint --version
// Used when --json flag is not available or fails
func (a *Analyzer) checkVersionText() error {
	cmd := exec.Command(a.golangciLintPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.NewAnalysisError("failed to check golangci-lint version", "", err)
	}
	
	// Parse version from output (format: "golangci-lint has version 2.8.0 built with...")
	outputStr := string(output)
	version := a.parseVersionText(outputStr)
	if version == "" {
		return errors.NewAnalysisError("could not parse golangci-lint version from output", "", fmt.Errorf("output: %s", outputStr))
	}
	
	// Ensure version has 'v' prefix for semver
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	
	// Validate semver format
	if !semver.IsValid(version) {
		return errors.NewAnalysisError("invalid golangci-lint version format", "", fmt.Errorf("version: %s", version))
	}
	
	// Compare with minimum required version (v2.8.0)
	minVersion := "v2.8.0"
	if semver.Compare(version, minVersion) < 0 {
		return errors.NewAnalysisError(
			fmt.Sprintf("golangci-lint version %s is too old", version),
			"",
			fmt.Errorf("minimum required version is %s. Please upgrade: https://golangci-lint.run/usage/install/", minVersion),
		)
	}
	
	a.logger.Debugf("golangci-lint version %s (>= %s) ✓", version, minVersion)
	return nil
}

// parseVersionText extracts version number from text output
// Format: "golangci-lint has version 2.8.0 built with..."
func (a *Analyzer) parseVersionText(output string) string {
	parts := strings.Fields(output)
	for i, part := range parts {
		if part == "version" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// AnalyzeConfig analyzes the current golangci-lint configuration
func (a *Analyzer) AnalyzeConfig(configPath string) (*types.ConfigAnalysis, error) {
	if err := a.FindBinary(); err != nil {
		return nil, err
	}
	
	// Check version meets minimum requirement
	if err := a.CheckVersion(); err != nil {
		return nil, err
	}

	output, err := a.runLintersCommand()
	if err != nil {
		return nil, errors.NewAnalysisError("failed to run golangci-lint linters", "", err)
	}

	var jsonOutput golangciLintOutput
	if err := json.Unmarshal(output, &jsonOutput); err != nil {
		return nil, errors.NewAnalysisError("failed to parse golangci-lint JSON output", "", err)
	}

	analysis := &types.ConfigAnalysis{
		ConfigPath:      configPath,
		EnabledLinters:  jsonOutput.Enabled,
		DisabledLinters: jsonOutput.Disabled,
		Recommendations: a.categorizeLinters(jsonOutput.Disabled),
	}

	a.calculateRecommendationCounts(analysis)

	return analysis, nil
}

// runLintersCommand runs `golangci-lint linters` and returns JSON output
func (a *Analyzer) runLintersCommand() ([]byte, error) {
	cmd := exec.Command(a.golangciLintPath, "linters", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		a.logger.Debugf("golangci-lint linters command failed: %v", err)
		a.logger.Debugf("Output: %s", string(output))
		return output, errors.NewAnalysisError("golangci-lint linters command failed", "", err)
	}
	return output, nil
}

// categorizeLinters categorizes disabled linters by priority
func (a *Analyzer) categorizeLinters(disabledLinters []types.LinterInfo) []types.LinterRecommendation {
	var recommendations []types.LinterRecommendation

	for _, linter := range disabledLinters {
		rec := types.LinterRecommendation{
			Name:   types.LinterName(linter.Name),
			Reason: a.getLinterReason(linter.Name),
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

// getLinterReason returns the human-readable reason for a linter recommendation
func (a *Analyzer) getLinterReason(name string) string {
	if reason, ok := constants.LinterReasons[types.LinterName(name)]; ok {
		return reason
	}
	return "Linter is disabled but may be useful"
}

// calculateRecommendationCounts calculates counts by priority level
func (a *Analyzer) calculateRecommendationCounts(analysis *types.ConfigAnalysis) {
	for _, rec := range analysis.Recommendations {
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

// GetLintersByPriority returns recommendations filtered by priority
func (a *Analyzer) GetLintersByPriority(recommendations []types.LinterRecommendation, priority types.LinterPriority) []types.LinterRecommendation {
	var filtered []types.LinterRecommendation
	for _, rec := range recommendations {
		if rec.Priority == priority {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

// FormatRecommendations formats recommendations as human-readable output
func (a *Analyzer) FormatRecommendations(analysis *types.ConfigAnalysis) string {
	var builder strings.Builder

	critical := a.GetLintersByPriority(analysis.Recommendations, types.LinterPriorityCritical)
	highValue := a.GetLintersByPriority(analysis.Recommendations, types.LinterPriorityHigh)
	mediumValue := a.GetLintersByPriority(analysis.Recommendations, types.LinterPriorityMedium)
	optional := a.GetLintersByPriority(analysis.Recommendations, types.LinterPriorityOptional)

	if len(critical) > 0 {
		builder.WriteString(fmt.Sprintf("🚨 %d CRITICAL linter(s) are disabled (should ALWAYS be enabled):\n", len(critical)))
		for _, rec := range critical {
			builder.WriteString(fmt.Sprintf("  - %s: %s\n", rec.Name, rec.Reason))
		}
		builder.WriteString("\n")
	}

	if len(highValue) > 0 {
		builder.WriteString(fmt.Sprintf("⚠️  %d HIGH VALUE linter(s) are disabled (recommended for most projects):\n", len(highValue)))
		for _, rec := range highValue {
			builder.WriteString(fmt.Sprintf("  - %s: %s\n", rec.Name, rec.Reason))
		}
		builder.WriteString("\n")
	}

	if len(mediumValue) > 0 {
		builder.WriteString(fmt.Sprintf("ℹ️  %d MEDIUM VALUE linter(s) are disabled (optional but recommended):\n", len(mediumValue)))
		for _, rec := range mediumValue {
			builder.WriteString(fmt.Sprintf("  - %s: %s\n", rec.Name, rec.Reason))
		}
		builder.WriteString("\n")
	}

	if len(optional) > 0 {
		builder.WriteString(fmt.Sprintf("💡 %d OPTIONAL linter(s) are disabled (for niche use cases):\n", len(optional)))
		for _, rec := range optional {
			builder.WriteString(fmt.Sprintf("  - %s: %s\n", rec.Name, rec.Reason))
		}
	}

	return builder.String()
}

// GetSummary returns a brief summary of recommendations
func (a *Analyzer) GetSummary(analysis *types.ConfigAnalysis) string {
	var parts []string

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
		len(analysis.Recommendations),
		strings.Join(parts, ", "),
	)
}
