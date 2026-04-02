package linter

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader types.ConfigLoader
	analyzer     types.LinterAnalyzer
	logger       *log.Logger
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer types.LinterAnalyzer, configLoader types.ConfigLoader) *Fixer {
	return &Fixer{
		configLoader: configLoader,
		analyzer:     analyzer,
		logger:       logger,
	}
}

// FixConfig fixes the golangci-lint configuration by enabling recommended linters.
func (f *Fixer) FixConfig(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (*types.MigrationResult, error) {
	result := f.FixConfigResult(ctx, configPath, priority, dryRun)

	return result.Get()
}

// FixConfigResult fixes the config and returns a Result type for railway-oriented programming.
func (f *Fixer) FixConfigResult(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) types.MigrationResultType {
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return types.ErrMigration(analysisError("load config", priority, dryRun, configPath, err))
	}

	originalEnabled := f.configLoader.GetLintersEnabled(cfg)

	hasInvalid, err := f.runPreFlightChecks(cfg, configPath, priority, dryRun)
	if err != nil {
		return types.ErrMigration(err)
	}

	if dryRun {
		if result, shouldReturn := f.checkDryRunEarlyReturns(
			cfg,
			hasInvalid,
			hasDeprecatedLinters(originalEnabled),
		); shouldReturn {
			return result
		}
	}

	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return types.ErrMigration(analysisError("analyze config", priority, dryRun, configPath, err))
	}

	return f.applyLintersFix(ctx, cfg, analysis, configPath, priority, dryRun, originalEnabled)
}

// checkDryRunEarlyReturns checks if we should early-return in dry-run mode.
func (f *Fixer) checkDryRunEarlyReturns(
	cfg *types.Config,
	hasInvalidDurations bool,
	deprecatedPresent bool,
) (types.MigrationResultType, bool) {
	if hasInvalidDurations {
		return f.calculateDryRunResultWithInvalidDurations(cfg), true
	}

	if deprecatedPresent {
		f.logger.Infof("Dry-run with deprecated linters - skipping analysis (run without --dry-run to fix)")

		return f.calculateDryRunResultWithDeprecated(cfg), true
	}

	return types.OkMigration(nil), false
}

func analysisError(action string, priority types.LinterPriority, dryRun bool, configPath string, err error) error {
	return apperrors.NewAnalysisError(
		fmt.Sprintf("failed to %s (priority=%d, dryRun=%t)", action, priority, dryRun),
		configPath, err,
	)
}

// runPreFlightChecks runs all pre-flight fixes and returns whether invalid durations were found.
func (f *Fixer) runPreFlightChecks(
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (bool, error) {
	checks := []struct {
		run func() (bool, error)
		msg string
	}{
		{func() (bool, error) {
			h, e := f.preFixInvalidDurations(cfg, configPath, dryRun)

			return h, e
		}, "invalid durations"},
		{func() (bool, error) {
			return false, f.preFixVersion(cfg, configPath, dryRun)
		}, "version field"},
		{func() (bool, error) {
			return false, f.preFixDeprecatedLinters(cfg, configPath, dryRun)
		}, "deprecated linters"},
		{func() (bool, error) {
			_, e := f.preFixTypecheck(cfg, configPath, dryRun)

			return false, e
		}, "typecheck"},
	}

	var hasInvalid bool

	for _, check := range checks {
		found, err := check.run()
		if found {
			hasInvalid = true
		}

		if err != nil {
			return hasInvalid, analysisError("pre-fix "+check.msg, priority, dryRun, configPath, err)
		}
	}

	return hasInvalid, nil
}

// fixCounts tracks the number of fixes applied by category.
type fixCounts struct {
	deprecation int
	enable      int
	formatter   int
	redundant   int
}

func (c fixCounts) total() int {
	return c.deprecation + c.enable + c.formatter + c.redundant
}

// applyLintersFix processes linter recommendations, applies fixes, and saves the config.
func (f *Fixer) applyLintersFix(
	ctx context.Context,
	cfg *types.Config,
	analysis *types.ConfigAnalysis,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
) types.MigrationResultType {
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)
	linterSet := buildLinterSet(cfg.Linters.Enable)

	counts := fixCounts{deprecation: 0, enable: 0, formatter: 0, redundant: 0}

	linterSet = f.replaceDeprecatedLinters(linterSet, originalEnabled, dryRun, &counts)

	formatterSet := buildLinterSet(cfg.Formatters.Enable)
	counts.formatter += f.enableCoreFormatters(formatterSet, dryRun)
	counts.formatter += f.enableGolinesFormatter(formatterSet, analysis, dryRun)
	counts.formatter += f.enableSwaggoFormatter(formatterSet, configPath, dryRun)
	counts.redundant += f.removeRedundantLinters(linterSet, formatterSet, dryRun)
	counts.redundant += f.removeRedundantGofmt(formatterSet, dryRun)

	counts.enable = f.enableRecommendedLinters(linterSet, disabledLinters, analysis, priority, dryRun)

	if dryRun {
		return f.dryRunResult(counts)
	}

	if counts.total() == 0 {
		return noFixesResult()
	}

	return f.applyAndSave(ctx, cfg, linterSet, formatterSet, configPath, priority, dryRun, counts)
}

func (f *Fixer) dryRunResult(counts fixCounts) types.MigrationResultType {
	f.logger.Infof("[DRY-RUN] Would apply %d fixes", counts.total())

	return types.OkMigration(&types.MigrationResult{
		FixesApplied: counts.total(),
		Message:      fmt.Sprintf("Would apply %d fixes (dry-run mode)", counts.total()),
		NextSteps: []string{
			"Run without --dry-run to apply these fixes",
			"Then run 'golangci-lint run --fix' to auto-fix code issues",
		},
	})
}

func noFixesResult() types.MigrationResultType {
	return types.OkMigration(&types.MigrationResult{
		FixesApplied: 0,
		Message:      "No fixes to apply",
		NextSteps: []string{
			"Your configuration is already up to date",
			"Run 'golangci-lint run' to check for code issues",
		},
	})
}

func (f *Fixer) applyAndSave(
	ctx context.Context,
	cfg *types.Config,
	linterSet, formatterSet map[string]bool,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	counts fixCounts,
) types.MigrationResultType {
	f.logger.Infof("Applying %d fixes...", counts.total())

	f.updateConfigFromSets(cfg, linterSet, formatterSet)
	f.updateGoVersion(ctx, cfg)
	f.updateRunnerSettings(cfg)
	f.updateBuildTags(cfg)

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return types.ErrMigration(analysisError("save config", priority, dryRun, configPath, err))
	}

	return types.OkMigration(&types.MigrationResult{
		FixesApplied: counts.total(),
		Message: fmt.Sprintf(
			"Successfully applied %d fixes (%d linters, %d formatters, %d deprecated, %d redundant)",
			counts.total(), counts.enable, counts.formatter, counts.deprecation, counts.redundant,
		),
		NextSteps: []string{
			"Run 'golangci-lint run --fix' to auto-fix code issues found by the newly enabled linters",
			"Run 'golangci-lint run' to see remaining issues that require manual fixes",
		},
	})
}

func (f *Fixer) updateGoVersion(ctx context.Context, cfg *types.Config) {
	goVersion := config.GetLocalGoVersion(ctx)
	if goVersion == "" {
		return
	}

	if cfg.Run.Go != goVersion {
		f.logger.Infof("Setting run.go to local version: %q -> %q", cfg.Run.Go, goVersion)
		cfg.Run.Go = goVersion
	}
}

func (f *Fixer) updateRunnerSettings(cfg *types.Config) {
	if !cfg.Run.AllowParallelRunners {
		f.logger.Infof("Enabling allow-parallel-runners: %v -> true", cfg.Run.AllowParallelRunners)
		cfg.Run.AllowParallelRunners = true
	}

	if !cfg.Run.AllowSerialRunners {
		f.logger.Infof("Enabling allow-serial-runners: %v -> true", cfg.Run.AllowSerialRunners)
		cfg.Run.AllowSerialRunners = true
	}
}

// goExperimentTags are build tags for GOEXPERIMENT features that affect user code.
// See: https://go.dev/src/internal/goexperiment/flags.go
var goExperimentTags = []string{
	"goexperiment.jsonv2",               // Enables json/v2 package
	"goexperiment.simd",                 // Enables simd package and intrinsics
	"goexperiment.goroutineleakprofile", // Enables goroutine leak profiling
}

func (f *Fixer) updateBuildTags(cfg *types.Config) {
	existingTags := make(map[string]bool)
	for _, tag := range cfg.Run.BuildTags {
		existingTags[tag] = true
	}

	for _, tag := range goExperimentTags {
		if !existingTags[tag] {
			f.logger.Infof("Adding build tag: %s", tag)
			cfg.Run.BuildTags = append(cfg.Run.BuildTags, tag)
			existingTags[tag] = true
		}
	}

	// Sort and deduplicate
	cfg.Run.BuildTags = sortAndDeduplicate(cfg.Run.BuildTags)
}

func sortAndDeduplicate(tags []string) []string {
	if len(tags) == 0 {
		return tags
	}

	seen := make(map[string]bool)
	unique := make([]string, 0, len(tags))

	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			unique = append(unique, tag)
		}
	}

	slices.Sort(unique)

	return unique
}

// replaceDeprecatedLinters replaces deprecated linters with their successors in the linter set.
func (f *Fixer) replaceDeprecatedLinters(
	linterSet map[string]bool,
	enabledLinters []string,
	dryRun bool,
	counts *fixCounts,
) map[string]bool {
	for _, linter := range enabledLinters {
		replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]
		if !isDeprecated {
			continue
		}

		counts.deprecation++

		delete(linterSet, linter)

		if linterSet[string(replacement.Replacement)] {
			f.logDeprecatedKeep(linter, replacement.Replacement, dryRun)

			continue
		}

		f.logDeprecatedReplace(linter, replacement, dryRun)

		if !dryRun {
			linterSet[string(replacement.Replacement)] = true
		}
	}

	return linterSet
}

func (f *Fixer) logDeprecatedKeep(linter string, replacement types.LinterName, dryRun bool) {
	if dryRun {
		f.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)", linter, replacement)
	} else {
		f.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacement)
	}
}

func (f *Fixer) logDeprecatedReplace(linter string, replacement types.LinterReplacement, dryRun bool) {
	if dryRun {
		f.logger.Infof(
			"[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)",
			linter,
			replacement.Replacement,
			replacement.Reason,
		)
	} else {
		f.logger.Infof(
			"Replacing deprecated linter: %s -> %s (%s)",
			linter,
			replacement.Replacement,
			replacement.Reason,
		)
	}
}

// enableGolinesFormatter enables the golines formatter if recommended at high priority.
func (f *Fixer) enableGolinesFormatter(
	formatterSet map[string]bool,
	analysis *types.ConfigAnalysis,
	dryRun bool,
) int {
	shouldEnable := false

	for _, rec := range analysis.FormatterRecommendations {
		if rec.Name == "golines" && rec.Priority == types.FormatterPriorityHigh {
			shouldEnable = true

			break
		}
	}

	if !shouldEnable || formatterSet["golines"] {
		return 0
	}

	if dryRun {
		f.logger.Debugf("[DRY-RUN] Would enable formatter: golines (formats code and fixes long lines)")
	} else {
		f.logger.Debugf("Enabling formatter: golines (formats code and fixes long lines)")

		formatterSet["golines"] = true
	}

	return 1
}

// enableCoreFormatters enables the core formatters: gci, gofumpt, goimports.
func (f *Fixer) enableCoreFormatters(formatterSet map[string]bool, dryRun bool) int {
	coreFormatters := []string{"gci", "gofumpt", "goimports"}
	count := 0

	for _, formatter := range coreFormatters {
		if formatterSet[formatter] {
			continue
		}

		count++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would enable formatter: %s", formatter)
		} else {
			f.logger.Debugf("Enabling formatter: %s", formatter)
			formatterSet[formatter] = true
		}
	}

	return count
}

// removeRedundantGofmt removes gofmt when gofumpt is enabled (gofumpt is a superset).
func (f *Fixer) removeRedundantGofmt(formatterSet map[string]bool, dryRun bool) int {
	if !formatterSet["gofumpt"] || !formatterSet["gofmt"] {
		return 0
	}

	if dryRun {
		f.logger.Debugf("[DRY-RUN] Would remove redundant formatter: gofmt (gofumpt is enabled and is a superset)")
	} else {
		f.logger.Debugf("Removing redundant formatter: gofmt (gofumpt is enabled and is a superset)")
		delete(formatterSet, "gofmt")
	}

	return 1
}

// enableSwaggoFormatter enables the swaggo formatter if swaggo is detected in the project.
func (f *Fixer) enableSwaggoFormatter(formatterSet map[string]bool, configPath string, dryRun bool) int {
	if formatterSet["swaggo"] {
		return 0
	}

	// Detect if swaggo is used in the project
	rootDir := filepath.Dir(configPath)
	detector := detection.NewDetector(rootDir)

	if !detector.HasSwaggo() {
		return 0
	}

	if dryRun {
		f.logger.Debugf("[DRY-RUN] Would enable formatter: swaggo (detected swaggo usage in project)")
	} else {
		f.logger.Debugf("Enabling formatter: swaggo (detected swaggo usage in project)")
		formatterSet["swaggo"] = true
	}

	return 1
}

// removeRedundantLinters removes linters that are superseded by enabled formatters.
func (f *Fixer) removeRedundantLinters(
	linterSet map[string]bool,
	formatterSet map[string]bool,
	dryRun bool,
) int {
	count := 0

	for linterName, mapping := range constants.RedundantLinters {
		if !linterSet[string(linterName)] {
			continue
		}

		if !formatterSet[string(mapping.Formatter)] {
			continue
		}

		count++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, mapping.Reason)
		} else {
			f.logger.Debugf("Removing redundant linter: %s (%s)", linterName, mapping.Reason)
			delete(linterSet, string(linterName))
		}
	}

	return count
}

// enableRecommendedLinters enables recommended linters that aren't already enabled or explicitly disabled.
func (f *Fixer) enableRecommendedLinters(
	linterSet map[string]bool,
	disabledLinters []string,
	analysis *types.ConfigAnalysis,
	priority types.LinterPriority,
	dryRun bool,
) int {
	count := 0

	for _, rec := range analysis.LinterRecommendations {
		if rec.Priority > priority {
			continue
		}

		lintName := resolveLinterName(rec.Name)

		if linterSet[lintName] || contains(disabledLinters, lintName) {
			continue
		}

		count++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would enable: %s (%s)", lintName, rec.Reason)
		} else {
			f.logger.Debugf("Enabling: %s (%s)", lintName, rec.Reason)
			linterSet[lintName] = true
		}
	}

	return count
}

func resolveLinterName(name types.LinterName) string {
	if replacement, isDeprecated := constants.DeprecatedLinters[name]; isDeprecated {
		return string(replacement.Replacement)
	}

	return name.String()
}

// updateConfigFromSets applies the linter and formatter sets back to the config struct.
func (f *Fixer) updateConfigFromSets(
	cfg *types.Config,
	linterSet map[string]bool,
	formatterSet map[string]bool,
) {
	enabledLinters := setToSortedSlice(linterSet)

	disabledLintersList := make([]string, 0)
	enabledLinters = slices.DeleteFunc(enabledLinters, func(linter string) bool {
		if _, isDisabled := constants.DisabledLinters[types.LinterName(linter)]; isDisabled {
			disabledLintersList = append(disabledLintersList, linter)

			return true
		}

		return false
	})

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = disabledLintersList

	if len(formatterSet) > 0 {
		cfg.Formatters.Enable = formattersToOrderedSlice(formatterSet)
	}
}

// formattersToOrderedSlice converts formatter set to ordered slice.
// Order: gci → goimports → gofumpt → golines → swaggo → others (sorted)
func formattersToOrderedSlice(set map[string]bool) []string {
	// Define explicit order
	order := []string{"gci", "goimports", "gofumpt", "golines", "swaggo"}

	result := make([]string, 0, len(set))
	remaining := make([]string, 0)

	// First pass: add formatters in explicit order
	for _, name := range order {
		if set[name] {
			result = append(result, name)
		}
	}

	// Second pass: add any remaining formatters (sorted alphabetically)
	for name := range set {
		if !slices.Contains(order, name) {
			remaining = append(remaining, name)
		}
	}

	slices.Sort(remaining)

	return append(result, remaining...)
}

func buildLinterSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))

	for _, item := range items {
		set[item] = true
	}

	return set
}

func setToSortedSlice(set map[string]bool) []string {
	result := make([]string, 0, len(set))

	for item := range set {
		result = append(result, item)
	}

	slices.Sort(result)

	return result
}

func hasDeprecatedLinters(enabledLinters []string) bool {
	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			return true
		}
	}

	return false
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
