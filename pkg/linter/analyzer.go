package linter

// Package linter provides analysis capabilities for golangci-lint configurations.
//
// This file contains the core Analyzer type and main analysis logic.
// Related functionality has been split into separate files:
//   - version_checker.go: Version checking functionality
//

import (
	"context"
	"encoding/json/v2"
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
	projectRoot      string
}

// NewAnalyzer creates a new linter analyzer.
func NewAnalyzer(logger *log.Logger) *Analyzer {
	return &Analyzer{
		golangciLintPath: "",
		logger:           logger,
		detectedVersion:  "",
		projectRoot:      "",
	}
}

// GetDetectedVersion returns the detected golangci-lint version, or empty string if not yet checked.
func (a *Analyzer) GetDetectedVersion() string {
	return a.detectedVersion
}

// golangciLinterEntry matches the JSON wire format of golangci-lint's linterHelp struct.
// golangci-lint uses lowercase JSON keys for linter fields but capitalized keys for the
// Enabled/Disabled wrapper. This is decoupled from types.LinterInfo (a Report type with
// PascalCase json) to avoid conflicting tag requirements.
type golangciLinterEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Groups      []string `json:"groups,omitempty"`
	Fast        bool     `json:"fast,omitempty"`
	AutoFix     bool     `json:"autoFix,omitempty"`
	Deprecated  bool     `json:"deprecated"`
	Since       string   `json:"since"`
	OriginalURL string   `json:"originalURL"`
}

func (e golangciLinterEntry) toLinterInfo() types.LinterInfo {
	return types.LinterInfo{
		Name:        types.LinterName(e.Name),
		Description: e.Description,
		Groups:      e.Groups,
		Fast:        e.Fast,
		AutoFix:     e.AutoFix,
		Deprecated:  e.Deprecated,
		Since:       e.Since,
		OriginalURL: e.OriginalURL,
	}
}

type golangciLintOutput struct {
	Enabled  []golangciLinterEntry `json:"Enabled"`
	Disabled []golangciLinterEntry `json:"Disabled"`
}

// golangciFormatterEntry matches the JSON wire format of golangci-lint's formatterHelp struct.
type golangciFormatterEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AutoFix     bool   `json:"autoFix,omitempty"`
}

func (e golangciFormatterEntry) toFormatterInfo() types.FormatterInfo {
	return types.FormatterInfo{
		Name:        types.FormatterName(e.Name),
		Description: e.Description,
		AutoFix:     e.AutoFix,
	}
}

// golangciLintFormattersOutput represents JSON output from golangci-lint formatters command.
type golangciLintFormattersOutput struct {
	Enabled  []golangciFormatterEntry `json:"Enabled"`
	Disabled []golangciFormatterEntry `json:"Disabled"`
}

// FindBinary finds the golangci-lint binary in PATH.
// It warns if multiple golangci-lint binaries are found in different PATH entries.
func (a *Analyzer) FindBinary(_ context.Context) error {
	path, err := exec.LookPath(constants.GolangciLintBinaryName)
	if err != nil {
		return apperrors.NewAnalysisError("golangci-lint not found in PATH", "", err)
	}

	a.golangciLintPath = path

	if allPaths := lookupAll(constants.GolangciLintBinaryName); len(allPaths) > 1 {
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
	a.projectRoot = filepath.Dir(configPath)

	enabledLinters := convertLinters(linterOutput.Enabled)
	disabledLinters := convertLinters(linterOutput.Disabled)
	enabledFormatters := convertFormatters(formatterOutput.Enabled)
	disabledFormatters := convertFormatters(formatterOutput.Disabled)

	analysis := &types.ConfigAnalysis{
		ConfigPath:               configPath,
		EnabledLinters:           enabledLinters,
		DisabledLinters:          disabledLinters,
		EnabledFormatters:        enabledFormatters,
		DisabledFormatters:       disabledFormatters,
		LinterRecommendations:    a.CategorizeLinters(disabledLinters, enabledFormatters),
		FormatterRecommendations: a.CategorizeFormatters(disabledFormatters, enabledFormatters),
	}

	a.calculateDeprecatedLinters(analysis)
	a.calculateRecommendationCounts(analysis)

	return analysis
}

func convertLinters(entries []golangciLinterEntry) []types.LinterInfo {
	result := make([]types.LinterInfo, 0, len(entries))
	for _, e := range entries {
		result = append(result, e.toLinterInfo())
	}

	return result
}

func convertFormatters(entries []golangciFormatterEntry) []types.FormatterInfo {
	result := make([]types.FormatterInfo, 0, len(entries))
	for _, e := range entries {
		result = append(result, e.toFormatterInfo())
	}

	return result
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
			Enabled:  []golangciFormatterEntry{},
			Disabled: []golangciFormatterEntry{},
		}
	}

	return &output
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
