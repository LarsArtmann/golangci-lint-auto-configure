package linter

// TODO: Consider using a transaction pattern for config changes (all or nothing)
// TODO: Extract duplicate linter detection into a separate validation step
// TODO: Add rollback mechanism for failed config saves
// TODO: Consider using immutable config copies for safer modifications

// TODO: Move duplicate deprecated linter handling logic into shared functions

// TODO: Reduce cognitive complexity of extracting helper functions

// TODO: Review type models for better architecture

// TODO: Consider using well-established libraries for easier maintenance

import (
	"context"
	"fmt"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader *config.Loader
	analyzer     *Analyzer
	logger       *log.Logger
}
// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer *Analyzer) *Fixer {
	return &Fixer{
		configLoader: config.NewLoader(logger),
		analyzer:     analyzer,
		logger:       logger,
	}
}
// FixConfig fixes the golangci-lint configuration by enabling recommended linters
func (f *Fixer) FixConfig(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (*types.MigrationResult, error) {
	result := f.FixConfigResult(ctx, configPath, priority, dryRun)
	return result.Get()
}
// fixStats tracks the statistics of fixes applied
type fixStats struct {
	deprecationFixes int
	enableFixes       int
	formatterFixes   int
	redundantFixes  int
}
// FixConfigResult fixes the config and returns a Result type for railway-oriented programming
func (f *Fixer) FixConfigResult(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) types.MigrationResultType {
	f.logger.Infof("Loading configuration: %s", configPath)
	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError("failed to load config", configPath, err))
	}
	// Check for deprecated linters before analysis
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	if f.hasDeprecatedLinters(enabledLinters) {
		// Pre-fix deprecated linters before analysis
		if err := f.preFixDeprecatedLinters(cfg, configPath, dryRun); err != nil {
            return types.ErrMigration(apperrors.NewAnalysisError("failed to pre-fix deprecated linters", configPath, err))
        }
        // In dry-run mode with deprecated linters, skip analysis
        if dryRun && f.hasDeprecatedLinters(enabledLinters) {
                f.logger.Infof("Dry-run with deprecated linters - skipping analysis (run without --dry-run to fix)")
                return f.calculateDryRunResultWithDeprecated(cfg, priority)
        }
        f.logger.Infof("Analyzing configuration...")
        analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
        if err != nil {
                return types.ErrMigration(apperrors.NewAnalysisError("failed to analyze config", configPath, err))
        }
        // Process linters and formatters
        stats := f.processLintersAndFormatters(cfg, analysis, priority, dryRun)
        // Handle dry-run mode
        if dryRun {
                f.logger.Infof("[DRY-RUN] Would apply %d fixes", stats.total())
                return types.OkMigration(&types.MigrationResult{
                        Success:      true,
                        FixesApplied: stats.total(),
                        Message:      fmt.Sprintf("Would apply %d fixes (dry-run mode)", stats.total()),
                })
        }
        if stats.total() == 0 {
                return types.OkMigration(&types.MigrationResult{
                        Success:      true,
                        FixesApplied: 0,
                        Message:      "No fixes to apply",
                })
        }
        f.logger.Infof("Applying %d fixes...", stats.total())
        // Apply changes to config
        if err := f.applyConfigChanges(cfg, stats, configPath); err != nil {
                return types.ErrMigration(apperrors.NewAnalysisError("failed to save config", configPath, err))
        }
        return types.OkMigration(&types.MigrationResult{
                Success:      true,
                FixesApplied: stats.total(),
                Message: fmt.Sprintf(
                        "Successfully applied %d fixes (%d linters, %d formatters, %d deprecated, %d redundant)",
                        stats.total(),
                        stats.enableFixes,
                        stats.formatterFixes,
                        stats.deprecationFixes,
                        stats.redundantFixes,
                ),
        })
}
// hasDeprecatedLinters checks if any enabled linters are deprecated
func (f *Fixer) hasDeprecatedLinters(enabledLinters []string) bool {
        for _, linter := range enabledLinters {
                if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
                        return true
                }
        }
        return false
}
// processLintersAndFormatters handles all linter and formatter processing
func (f *Fixer) processLintersAndFormatters(
        cfg *types.Config,
        analysis *types.ConfigAnalysis,
        priority types.LinterPriority,
        dryRun bool,
) fixStats {
        stats := fixStats{}}
        enabledLinters := f.configLoader.GetLintersEnabled(cfg)
        disabledLinters := f.configLoader.GetLintersDisabled(cfg)
        // Track all linters to ensure uniqueness
        linterSet := make(map[string]bool)
        for _, linter := range enabledLinters {
                linterSet[linter] = true
        }
        // Replace deprecated linters
        stats.deprecationFixes = f.replaceDeprecatedLintersInSet(linterSet, enabledLinters, dryRun)
        // Handle formatters
        formatterSet := make(map[string]bool)
        for _, formatter := range cfg.Formatters.Enable {
                formatterSet[formatter] = true
        }
        stats.formatterFixes = f.handleFormatters(analysis, formatterSet, dryRun)
        // Handle redundant linters
        stats.redundantFixes = f.handleRedundantLinters(linterSet, formatterSet, dryRun)
        // Apply linter recommendations
        stats.enableFixes = f.applyLinterRecommendations(
                analysis.LinterRecommendations,
                linterSet,
                disabledLinters,
                priority,
                dryRun,
        )
        // Update config with final sets (only in non-dry-run)
        if !dryRun {
                f.updateConfigSets(cfg, linterSet, formatterSet)
        }
        return stats
}
// replaceDeprecatedLintersInSet replaces deprecated linters in the linter set
func (f *Fixer) replaceDeprecatedLintersInSet(
        linterSet map[string]bool,
        enabledLinters []string,
        dryRun bool,
) int {
        deprecationFixes := 0
        for _, linter := range enabledLinters {
                replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]
                if !isDeprecated {
                        continue
                }
                deprecationFixes++
                delete(linterSet, linter)
                replacementName := string(replacement.Replacement)
                if linterSet[replacementName] {
                        f.logDeprecatedLinterRemoval(linter, replacementName, dryRun)
                        continue
                }
                f.logDeprecatedLinterReplacement(linter, replacement, dryRun)
                if !dryRun {
                        linterSet[replacementName] = true
                }
        }
        return deprecationFixes
}
// logDeprecatedLinterReplacement logs the replacement of a deprecated linter
func (f *Fixer) logDeprecatedLinterReplacement(
        linter string,
        replacement types.LinterReplacement,
        dryRun bool,
) {
        prefix := "[DRY-RUN] Would "
        if dryRun {
                f.logger.Infof("%sreplace deprecated linter: %s -> %s (%s)",
                        prefix, linter, replacement.Replacement, replacement.Reason)
        } else {
                f.logger.Infof("Replacing deprecated linter: %s -> %s (%s)",
                        linter, replacement.Replacement, replacement.Reason)
        }
}
// logDeprecatedLinterRemoval logs the removal of a deprecated linter (replacement already exists)
func (f *Fixer) logDeprecatedLinterRemoval(linter, replacementName string, dryRun bool) {
        if dryRun {
                f.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)", linter, replacementName)
        } else {
                f.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacementName)
        }
}
// handleFormatters handles formatter recommendations
func (f *Fixer) handleFormatters(
        analysis *types.ConfigAnalysis,
        formatterSet map[string]bool,
        dryRun bool,
) int {
        shouldEnableGolines := false
        for _, rec := range analysis.FormatterRecommendations {
                if rec.Name == "golines" && rec.Priority == types.FormatterPriorityHigh {
                        shouldEnableGolines = true
                        break
                }
        }
        if !shouldEnableGolines || formatterSet["golines"] {
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
// handleRedundantLinters removes redundant linters when corresponding formatters are enabled
func (f *Fixer) handleRedundantLinters(
        linterSet map[string]bool,
        formatterSet map[string]bool,
        dryRun bool,
) int {
        redundantFixes := 0
        golinesWillBeEnabled := formatterSet["golines"]
        for linterName, range constants.RedundantLinters {
                if !linterSet[string(linterName)] {
                        continue
                }
                if linterName != "lll" || !golinesWillBeEnabled {
                        continue
                }
                redundantFixes++
                f.logRedundantLinterRemoval(linterName, constants.RedundantLinters[linterName], dryRun)
                if !dryRun {
                        delete(linterSet, string(linterName))
                }
        }
        return redundantFixes
}
// logRedundantLinterRemoval logs the removal of a redundant linter
func (f *Fixer) logRedundantLinterRemoval(linterName types.LinterName, reason string, dryRun bool) {
        if dryRun {
                f.logger.Debugf("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, reason)
        } else {
                f.logger.Debugf("Removing redundant linter: %s (%s)", linterName, reason)
        }
}
// applyLinterRecommendations applies linter recommendations to the linter set
func (f *Fixer) applyLinterRecommendations(
        recommendations []types.LinterRecommendation,
        linterSet map[string]bool,
        disabledLinters []string,
        priority types.LinterPriority,
        dryRun bool,
) int {
        enableFixes := 0
        for _, rec := range recommendations {
                if rec.Priority > priority {
                        continue
                }
                lintName := rec.Name.String()
                // Check if this linter is deprecated and replace it with its successor
                if replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(lintName)]; isDeprecated {
                        lintName = string(replacement.Replacement)
                }
                if linterSet[lintName] || contains(disabledLinters, lintName) {
                        continue
                }
                enableFixes++
                f.logLinterEnable(lintName, rec.Reason, dryRun)
                if !dryRun {
                        linterSet[lintName] = true
                }
        }
        return enableFixes
}
// logLinterEnable logs the enabling of a linter
func (f *Fixer) logLinterEnable(lintName, reason string, dryRun bool) {
        if dryRun {
                f.logger.Debugf("[DRY-RUN] Would enable: %s (%s)", lintName, reason)
        } else {
                f.logger.Debugf("Enabling: %s (%s)", lintName, reason)
        }
}
// applyConfigChanges applies the final changes to the config and saves
func (f *Fixer) applyConfigChanges(
        cfg *types.Config,
        stats fixStats,
        configPath string,
) error {
        // Convert final linter set to sorted slice
        enabledLinters := make([]string, 0, len(stats.linterSet))
        for linter := range stats.linterSet {
                enabledLinters = append(enabledLinters, linter)
        }
        cfg.Linters.Enable = enabledLinters
        cfg.Linters.Disable = []string{}
        // Convert formatter set to slice
        if len(stats.formatterSet) > 0 {
                enabledFormatters := make([]string, 0, len(stats.formatterSet))
                for formatter := range stats.formatterSet {
                        enabledFormatters = append(enabledFormatters, formatter)
                }
                cfg.Formatters.Enable = enabledFormatters
        }
        f.logger.Infof("Saving configuration...")
        return f.configLoader.SaveConfig(cfg, configPath)
}
// updateConfigSets is a helper that updates the linter and formatter sets
// It must be called with the final sets after processing
func updateConfigSets(cfg *types.Config, linterSet, formatterSet map[string]bool) {
        cfg.Linters.Enable = keysToSortedSlice(linterSet)
        cfg.Linters.Disable = []string{}
        if len(formatterSet) > 0 {
                cfg.Formatters.Enable = keysToSortedSlice(formatterSet)
        }
}
// keysToSortedSlice converts a map keys to a sorted slice
func keysToSortedSlice(m map[string]bool) []string {
        result := make([]string, 0, len(m))
        for k := range m {
                result = append(result, k)
        }
        return result
}
// total returns the total number of fixes
func (s fixStats) total() int {
        return s.deprecationFixes + s.enableFixes + s.formatterFixes + s.redundantFixes
}
// contains checks if a slice contains an item
func contains(slice []string, item string) bool {
        return slices.Contains(slice, item)
}
// preFixDeprecatedLinters replaces deprecated linters in the config before analysis
// This is necessary because golangci-lint linters command will fail if the config
// contains deprecated/removed linters (even in the disable list)
func (f *Fixer) preFixDeprecatedLinters(cfg *types.Config, configPath string, dryRun bool) error {
        enabledLinters := f.configLoader.GetLintersEnabled(cfg)
        disabledLinters := f.configLoader.GetLintersDisabled(cfg)
        var deprecatedFound []string
        linterSet := make(map[string]bool)
        for _, linter := range enabledLinters {
                if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
                        deprecatedFound = append(deprecatedFound, linter)
                        continue
                }
                linterSet[linter] = true
        }
        disabledSet := make(map[string]bool)
        for _, linter := range disabledLinters {
                if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
                        deprecatedFound = append(deprecatedFound, linter+" (disabled)")
                        continue
                }
                disabledSet[linter] = true
        }
        if len(deprecatedFound) == 0 {
                return nil
        }
        if dryRun {
                f.logger.Infof("[DRY-RUN] Would pre-fix %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)
                return nil
        }
        f.logger.Infof("Pre-fixing %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)
        fixedEnabled := keysToSortedSlice(linterSet)
        fixedDisabled := keysToSortedSlice(disabledSet)
        cfg.Linters.Enable = fixedEnabled
        cfg.Linters.Disable = fixedDisabled
        // Save the fixed config so that golangci-lint linters command will work
        err := f.configLoader.SaveConfig(cfg, configPath)
        if err != nil {
                return fmt.Errorf("failed to save pre-fixed config: %w", err)
        }
        return nil
}
// calculateDryRunResultWithDeprecated calculates the dry-run result when deprecated linters are present
// Since the config has deprecated linters, we can't run golangci-lint linters for analysis,
// so we just report what would be fixed regarding deprecated linters
func (f *Fixer) calculateDryRunResultWithDeprecated(
        cfg *types.Config,
        priority types.LinterPriority,
) types.MigrationResultType {
        enabledLinters := f.configLoader.GetLintersEnabled(cfg)
        deprecationFixes := 0
        linterSet := make(map[string]bool)
        for _, linter := range enabledLinters {
                replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]
                if isDeprecated {
                        deprecationFixes++
                        delete(linterSet, linter)
                        replacementName := string(replacement.Replacement)
                        if !linterSet[replacementName] {
                                f.logger.Infof(
                                        "[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)",
                                        linter,
                                        replacement.Replacement,
                                        replacement.Reason,
                                )
                                linterSet[replacementName] = true
                        } else {
                                f.logger.Infof(
                                        "[DRY-RUN] Would remove deprecated %s (keeping existing %s)",
                                        linter,
                                        replacement.Replacement,
                                )
                        }
                } else {
                        linterSet[linter] = true
                }
        }
        f.logger.Infof("[DRY-RUN] Would apply %d fixes", deprecationFixes)
        return types.OkMigration(&types.MigrationResult{
                Success:      true,
                FixesApplied: deprecationFixes,
                Message: fmt.Sprintf(
                        "Would apply %d fixes (dry-run mode, skipped analysis due to deprecated linters)",
                        deprecationFixes,
                ),
        })
}
