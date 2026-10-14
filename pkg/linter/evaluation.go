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
	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, analysisError("load config", priority, true, configPath, "", err)
	}

	version := f.detectVersion(ctx)

	if f.analyzer != nil {
		if err := f.verifyAnalyzerContract(ctx); err != nil {
			return nil, err
		}
	}

	if err := f.rescueOverspecifiedRunGo(cfg, configPath, true); err != nil {
		return nil, analysisError("rescue run.go", priority, true, configPath, version, err)
	}

	f.loadPolicy(configPath)

	overlay := f.sidecarOverlay(cfg)

	env, envErr := CollectEnv(ctx, f.analyzer, f.goVersionProvider, configPath)
	if envErr != nil {
		return nil, analysisError("collect environment", priority, true, configPath, version, envErr)
	}

	result, err := Converge(ctx, cfg, env, overlay, f.convergeOptions(priority, true))
	if err != nil {
		return nil, analysisError("converge config", priority, true, configPath, version, err)
	}

	userDisabled := disableMinusEnable(cfg)

	return &Evaluation{
		ConfigPath: configPath,
		Env:        env,
		Desired:    renderDesired(env, f.convergeOptions(priority, true), cfg, userDisabled),
		Observed:   cfg,
		Deviations: append(append([]Change{}, result.NormalizerChanges...), result.Plan...),
		Suppressed: result.Suppressed,
		Converged:  result.Converged,
	}, nil
}
