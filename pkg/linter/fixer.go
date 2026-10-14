package linter

import (
	"context"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// fixerConfigLoader is the narrowest interface the Fixer needs: load, save,
// and inspect linter state. This is a subset of types.ConfigLoader.
type fixerConfigLoader interface {
	types.ConfigReader
	types.ConfigWriter
	types.ConfigInspector
}

// Fixer fixes golangci-lint configurations by converging the observed config
// onto the declarative desired state (ADR-017).
type Fixer struct {
	configLoader      fixerConfigLoader
	analyzer          types.LinterAnalyzer
	logger            *log.Logger
	ledger            audit.Recorder
	pol               *policy.Policy
	goVersionProvider GoVersionProvider
	forceSettings     bool
	showMergedRules   bool
	pragmatic         bool
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer types.LinterAnalyzer, configLoader fixerConfigLoader) *Fixer {
	return &Fixer{
		configLoader:      configLoader,
		analyzer:          analyzer,
		logger:            logger,
		goVersionProvider: func(context.Context) string { return "" },
		ledger:            audit.NoopRecorder{},
		pol:               nil,
		forceSettings:     false,
		showMergedRules:   false,
		pragmatic:         false,
	}
}

// SetLedger sets the audit recorder used to log every config change the fixer makes.
// If recorder is nil, the fixer falls back to a NoopRecorder (no recording).
func (f *Fixer) SetLedger(recorder audit.Recorder) {
	if recorder == nil {
		recorder = audit.NoopRecorder{}
	}

	f.ledger = recorder
}

// SetGoVersionProvider sets the function used to detect the local Go version.
// If provider is nil, the fixer uses a no-op that returns an empty string (skips go-version updates).
func (f *Fixer) SetGoVersionProvider(provider GoVersionProvider) {
	if provider == nil {
		provider = func(context.Context) string { return "" }
	}

	f.goVersionProvider = provider
}

// SetForceSettings controls whether injectDefaultSettings overwrites existing
// linter settings (true) or only fills missing ones (false, the default).
func (f *Fixer) SetForceSettings(force bool) {
	f.forceSettings = force
}

// SetShowMergedRules controls whether exclusion-rule merges are logged at
// info level (the --show-merged-rules affordance). The audit-ledger records
// are independent of this flag.
func (f *Fixer) SetShowMergedRules(show bool) {
	f.showMergedRules = show
}

// SetPragmatic drops the pragmatic-noise linters from the converge render.
func (f *Fixer) SetPragmatic(enabled bool) {
	f.pragmatic = enabled
}

// PlanDeviations returns the full change set (normalizer repairs + plan) the
// converge engine would apply to the config. Report-only: nothing is
// persisted and no audit entries are written. This is the single source both
// lint mode and apply mode derive from, so findings can never drift from
// what apply actually does.
func (f *Fixer) PlanDeviations(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
) ([]Change, error) {
	evaluation, err := f.Evaluate(ctx, configPath, priority)
	if err != nil {
		return nil, err
	}

	return evaluation.Deviations, nil
}

// FixConfig converges the golangci-lint configuration onto the desired state.
func (f *Fixer) FixConfig(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (*types.MigrationResult, error) {
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return migrationError("load config", priority, dryRun, configPath, "", err)
	}

	return f.convergeFixConfig(ctx, cfg, configPath, priority, dryRun)
}

// detectVersion tries to detect the golangci-lint version early for version-gated replacements.
// Returns empty string if detection fails (version-gating is skipped, all replacements applied).
func (f *Fixer) detectVersion(ctx context.Context) string {
	if f.analyzer == nil {
		return ""
	}

	version := f.analyzer.GetDetectedVersion()
	if version != "" {
		return version
	}

	if err := f.analyzer.FindBinary(ctx); err != nil {
		f.logger.Debugf("Version detection skipped: %v", err)

		return ""
	}

	if err := f.analyzer.CheckVersion(ctx); err != nil {
		f.logger.Debugf("Version check failed: %v", err)

		return ""
	}

	return f.analyzer.GetDetectedVersion()
}
