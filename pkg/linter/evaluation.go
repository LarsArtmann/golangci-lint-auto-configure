package linter

import (
	"context"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Evaluation is the complete report input of one convergence assessment: the
// environment facts, the desired state they render, the observed config, and
// the deviations between them. One type in, every report out — the HTML,
// JSON, SARIF, and finding outputs are projections of the same evaluation,
// so they can never disagree about the facts.
type Evaluation struct {
	// ConfigPath locates the observed config on disk.
	ConfigPath string
	// Env carries the collected environment facts.
	Env Env
	// Desired is the rendered desired configuration.
	Desired DesiredConfig
	// Observed is the config as it exists on disk (normalized shape).
	Observed *types.Config
	// Deviations are the changes that move observed toward desired
	// (normalizer repairs + plan), in canonical order.
	Deviations []Change
	// Suppressed lists linters the never-enable section protected from
	// re-enable (audit-only events, no config change).
	Suppressed []types.LinterName
	// Converged reports whether the observed config already matched.
	Converged bool
}

// Evaluate runs the full convergence assessment in report-only mode: collect,
// normalize (in memory), render, plan. Nothing is persisted and no audit
// entries are written.
func (f *Fixer) Evaluate(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
) (*Evaluation, error) {
	cfg, env, result, err := f.evaluateConvergence(ctx, configPath, priority)
	if err != nil {
		return nil, err
	}

	userDisabled := disableMinusEnable(cfg)
	options := f.convergeOptions(priority, true)

	return &Evaluation{
		ConfigPath: configPath,
		Env:        env,
		Desired:    renderDesired(env, options, cfg, userDisabled),
		Observed:   cfg,
		Deviations: append(append([]Change{}, result.NormalizerChanges...), result.Plan...),
		Suppressed: result.Suppressed,
		Converged:  result.Converged,
	}, nil
}

// evaluateConvergence runs the report-only pipeline: load, verify, rescue,
// collect, converge — returning config, env, and convergence result.
func (f *Fixer) evaluateConvergence(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
) (*types.Config, Env, *ConvergeResult, error) {
	cfg, err := f.loadForEvaluation(ctx, configPath, priority)
	if err != nil {
		return nil, Env{}, nil, err
	}

	env, err := f.collectEvaluationEnv(ctx, configPath, priority)
	if err != nil {
		return nil, Env{}, nil, err
	}

	result, err := Converge(ctx, cfg, env, f.sidecarOverlay(cfg), f.convergeOptions(priority, true))
	if err != nil {
		return nil, Env{}, nil, analysisError("converge config", priority, true, configPath, "", err)
	}

	return cfg, env, result, nil
}

// loadForEvaluation loads the config and runs the pre-analysis preparation.
func (f *Fixer) loadForEvaluation(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
) (*types.Config, error) {
	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, analysisError("load config", priority, true, configPath, "", err)
	}

	if err := f.prepareEvaluation(ctx, cfg, configPath, priority); err != nil {
		return nil, err
	}

	return cfg, nil
}

// collectEvaluationEnv gathers the full environment for the report-only run.
func (f *Fixer) collectEvaluationEnv(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
) (Env, error) {
	env, envErr := CollectEnv(ctx, f.analyzer, f.goVersionProvider, configPath)
	if envErr != nil {
		return Env{}, analysisError("collect environment", priority, true, configPath, f.detectVersion(ctx), envErr)
	}

	return env, nil
}

// prepareEvaluation runs the analyzer contract and the run.go rescue; a
// config the binary cannot load fails here instead of degrading silently.
func (f *Fixer) prepareEvaluation(
	ctx context.Context,
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
) error {
	version := f.detectVersion(ctx)

	if f.analyzer != nil {
		if err := f.verifyAnalyzerContract(ctx); err != nil {
			return err
		}
	}

	if err := f.rescueOverspecifiedRunGo(cfg, configPath, true); err != nil {
		return analysisError("rescue run.go", priority, true, configPath, version, err)
	}

	f.loadPolicy(configPath)

	return nil
}
