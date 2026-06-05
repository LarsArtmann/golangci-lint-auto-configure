package linter

// Package linter provides analysis capabilities for golangci-lint configurations.
//
// This file contains the core Analyzer type and main analysis logic.
// Related functionality has been split into separate files:
//   - version_checker.go: Version checking functionality
//

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"golang.org/x/sync/errgroup"
)

// Analyzer analyzes golangci-lint configurations and provides recommendations.
type Analyzer struct {
	golangciLintPath string
	logger           *log.Logger
	detectedVersion  string
}

// NewAnalyzer creates a new linter analyzer.
func NewAnalyzer(logger *log.Logger) *Analyzer {
	return &Analyzer{
		golangciLintPath: "",
		logger:           logger,
		detectedVersion:  "",
	}
}

// GetDetectedVersion returns the detected golangci-lint version, or empty string if not yet checked.
func (a *Analyzer) GetDetectedVersion() string {
	return a.detectedVersion
}

type golangciLintOutput struct {
	Enabled  []types.LinterInfo `json:"enabled"`
	Disabled []types.LinterInfo `json:"disabled"`
}

// golangciLintFormattersOutput represents JSON output from golangci-lint formatters command.
type golangciLintFormattersOutput struct {
	Enabled  []types.FormatterInfo `json:"enabled"`
	Disabled []types.FormatterInfo `json:"disabled"`
}

// FindBinary finds the golangci-lint binary in PATH.
// It warns if multiple golangci-lint binaries are found in different PATH entries.
func (a *Analyzer) FindBinary(_ context.Context) error {
	path, err := exec.LookPath("golangci-lint")
	if err != nil {
		return apperrors.NewAnalysisError("golangci-lint not found in PATH", "", err)
	}

	a.golangciLintPath = path

	if allPaths := lookupAll("golangci-lint"); len(allPaths) > 1 {
		a.logger.Warnf(
			"Multiple golangci-lint binaries found in PATH (%s); "+
				"using %s — consider removing duplicates to avoid ambiguity",
			strings.Join(allPaths, ", "), path,
		)
	}

	return nil
}

// AnalyzeConfig analyzes the current golangci-lint configuration.
func (a *Analyzer) AnalyzeConfig(ctx context.Context, configPath string) (*types.ConfigAnalysis, error) {
	if err := a.FindBinary(ctx); err != nil {
		return nil, err
	}

	if err := a.CheckVersion(ctx); err != nil {
		return nil, err
	}

	linterOutput, formatterOutput, err := a.parseConfigOutputs(ctx, configPath)
	if err != nil {
		return nil, err
	}

	return a.buildAnalysis(configPath, linterOutput, formatterOutput), nil
}

// parseConfigOutputs runs linter and formatter parsing in parallel using errgroup.
func (a *Analyzer) parseConfigOutputs(
	ctx context.Context,
	configPath string,
) (*golangciLintOutput, *golangciLintFormattersOutput, error) {
	errGroup, ctx := errgroup.WithContext(ctx)

	var (
		linterOutput *golangciLintOutput
		linterErr    error
	)

	errGroup.Go(func() error {
		linterOutput, linterErr = a.parseLintersOutput(ctx, configPath)

		return linterErr
	})

	var formatterOutput *golangciLintFormattersOutput

	errGroup.Go(func() error {
		formatterOutput = a.parseFormattersOutput(ctx, configPath)

		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return nil, nil, err
	}

	return linterOutput, formatterOutput, nil
}

// buildAnalysis constructs the ConfigAnalysis from parsed outputs.
func (a *Analyzer) buildAnalysis(
	configPath string,
	linterOutput *golangciLintOutput,
	formatterOutput *golangciLintFormattersOutput,
) *types.ConfigAnalysis {
	analysis := &types.ConfigAnalysis{
		ConfigPath:               configPath,
		EnabledLinters:           linterOutput.Enabled,
		DisabledLinters:          linterOutput.Disabled,
		EnabledFormatters:        formatterOutput.Enabled,
		DisabledFormatters:       formatterOutput.Disabled,
		LinterRecommendations:    a.CategorizeLinters(linterOutput.Disabled, formatterOutput.Enabled),
		FormatterRecommendations: a.categorizeFormatters(formatterOutput.Disabled),
	}

	a.calculateDeprecatedLinters(analysis)
	a.calculateRecommendationCounts(analysis)

	return analysis
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

	a.formatDeprecatedSection(&builder, analysis)

	critical := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityCritical)
	highValue := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityHigh)
	mediumValue := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityMedium)
	optional := a.GetLintersByPriority(analysis.LinterRecommendations, types.LinterPriorityOptional)

	a.formatPrioritySection(&builder, critical, "🚨", "CRITICAL", "should ALWAYS be enabled")
	a.formatPrioritySection(&builder, highValue, "⚠️ ", "HIGH VALUE", "recommended for most projects")
	a.formatPrioritySection(&builder, mediumValue, "ℹ️ ", "MEDIUM VALUE", "optional but recommended")
	a.formatPrioritySection(&builder, optional, "💡", "OPTIONAL", "for niche use cases")

	return builder.String()
}

// GetSummary returns a brief summary of recommendations.
func (a *Analyzer) GetSummary(analysis *types.ConfigAnalysis) string {
	parts := summaryParts(analysis)

	if len(parts) == 0 {
		return "All linters enabled - no recommendations"
	}

	return fmt.Sprintf(
		"Found %d disabled linters: %s (see details above)",
		len(analysis.LinterRecommendations),
		strings.Join(parts, ", "),
	)
}

func summaryParts(analysis *types.ConfigAnalysis) []string {
	counts := []struct {
		n int
		l string
	}{
		{analysis.DeprecatedCount, "DEPRECATED"},
		{analysis.CriticalCount, "CRITICAL"},
		{analysis.HighValueCount, "HIGH"},
		{analysis.MediumValueCount, "MEDIUM"},
		{analysis.OptionalCount, "OPTIONAL"},
	}

	var parts []string

	for _, c := range counts {
		if c.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c.n, c.l))
		}
	}

	return parts
}

func (a *Analyzer) parseLintersOutput(ctx context.Context, configPath string) (*golangciLintOutput, error) {
	lintOutput, err := a.runLintersCommand(ctx, configPath)
	if err != nil {
		return nil, apperrors.NewAnalysisError("failed to run golangci-lint linters", "", err)
	}

	var output golangciLintOutput
	if err := json.Unmarshal(lintOutput, &output); err != nil {
		return nil, apperrors.NewAnalysisError("failed to parse golangci-lint linters JSON output", "", err)
	}

	return &output, nil
}

func (a *Analyzer) parseFormattersOutput(ctx context.Context, configPath string) *golangciLintFormattersOutput {
	formatOutput, err := a.runFormattersCommand(ctx, configPath)
	if err != nil {
		a.logger.Debugf("Formatters analysis skipped: %v", err)

		formatOutput = []byte(`{"Enabled": [], "Disabled": []}`)
	}

	var output golangciLintFormattersOutput
	if err := json.Unmarshal(formatOutput, &output); err != nil {
		a.logger.Debugf("Failed to parse formatters JSON, skipping: %v", err)

		output = golangciLintFormattersOutput{
			Enabled:  []types.FormatterInfo{},
			Disabled: []types.FormatterInfo{},
		}
	}

	return &output
}

func (a *Analyzer) formatDeprecatedSection(builder *strings.Builder, analysis *types.ConfigAnalysis) {
	if len(analysis.DeprecatedLinters) == 0 {
		return
	}

	fmt.Fprintf(builder, "⚠️  %d DEPRECATED linter(s) are enabled (should be migrated):\n",
		len(analysis.DeprecatedLinters))

	for _, linter := range analysis.DeprecatedLinters {
		if replacement, ok := constants.DeprecatedLinters[linter.Name]; ok {
			fmt.Fprintf(builder, "  - %s: Use %s instead (%s)\n",
				linter.Name, replacement.Replacement, linter.Description)
		} else {
			fmt.Fprintf(builder, "  - %s: %s (no replacement specified)\n", linter.Name, linter.Description)
		}
	}

	builder.WriteString("\n")
}

func (a *Analyzer) formatPrioritySection(
	builder *strings.Builder,
	recommendations []types.LinterRecommendation,
	icon, label, subtitle string,
) {
	if len(recommendations) == 0 {
		return
	}

	fmt.Fprintf(builder, "%s %d %s linter(s) are disabled (%s):\n", icon, len(recommendations), label, subtitle)

	for _, rec := range recommendations {
		fmt.Fprintf(builder, "  - %s: %s\n", rec.Name, rec.Reason)
	}

	builder.WriteString("\n")
}

// lookupAll searches every directory in PATH for the named executable
// and returns all matching absolute paths in order.
func lookupAll(name string) []string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil
	}

	var found []string

	for _, dir := range filepath.SplitList(pathEnv) {
		candidate := filepath.Join(dir, name)
		if info, err := os.Stat(candidate); err == nil && //nolint:gosec // PATH entries are trusted
			!info.IsDir() && isExecutable(info) {
			if abs, err := filepath.Abs(candidate); err == nil {
				found = append(found, abs)
			}
		}
	}

	return found
}

// isExecutable checks whether the file mode has any execute bit set.
func isExecutable(info os.FileInfo) bool {
	return info.Mode().Perm()&0o111 != 0
}
