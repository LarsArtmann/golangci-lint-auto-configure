package linter

import (
	"context"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Engine names the configuration strategy the fixer drives.
type Engine string

const (
	// EngineLegacy is the historical imperative mutator pipeline.
	EngineLegacy Engine = "legacy"
	// EngineConverge is the declarative reconciler: collect → normalize →
	// render → plan → apply. Opt-in until the parity corpus proves byte
	// equivalence (plan phase P5, flip in T20).
	EngineConverge Engine = "converge"
)

// ConvergeOptions carries the user intent for one convergence run.
type ConvergeOptions struct {
	Priority      types.LinterPriority
	DryRun        bool
	ForceSettings bool
	Pragmatic     bool
}

// ConvergeResult reports what one convergence run decided.
type ConvergeResult struct {
	// Env is the collected environment (echoed for reports).
	Env Env
	// Plan is the full ordered change set (empty when converged).
	Plan []Change
	// NormalizerChanges are repairs applied before planning (bad keys,
	// deprecation replacement).
	NormalizerChanges []Change
	// Merged is the converged configuration (nil for dry runs).
	Merged *types.Config
	// Converged reports whether the observed config already matched.
	Converged bool
}

// Converge computes the desired state for env, plans the move from observed,
// and returns the merged config. Pure with respect to disk: nothing here
// reads or writes files — the fixer owns I/O and audit emission.
func Converge(
	ctx context.Context,
	observed *types.Config,
	analyzer types.LinterAnalyzer,
	goVersionProvider GoVersionProvider,
	configPath string,
	sidecar *sidecarOverlay,
	options ConvergeOptions,
) (*ConvergeResult, error) {
	env, err := CollectEnv(ctx, analyzer, goVersionProvider, configPath)
	if err != nil {
		return nil, err
	}

	working, normalizerChanges := NormalizeBadSettingsKeys(observed)

	working, replacementChanges := ReplaceDeprecatedLinters(working, env.Tool, sidecar.neverEnable)
	normalizerChanges = append(normalizerChanges, replacementChanges...)

	desired := Desired(env,
		WithPragmatic(options.Pragmatic),
		WithMinPriority(options.Priority),
	)

	strategies := StrategiesFor(MergeOptions{ForceSettings: options.ForceSettings})

	plan := PlanDesired(working, desired, strategies)
	plan = append(plan, planDisableOverlay(working, sidecar)...)
	sortChanges(plan)

	if err := ValidatePlan(plan); err != nil {
		return nil, err
	}

	converged := len(plan) == 0 && len(normalizerChanges) == 0

	result := &ConvergeResult{
		Env:               env,
		Plan:              plan,
		NormalizerChanges: normalizerChanges,
		Converged:         converged,
	}

	if !options.DryRun {
		merged := ApplyPlan(working, plan)
		result.Merged = merged
	}

	return result, nil
}

// sidecarOverlay is the declared-divergence layer: the sidecar policy's
// never-enable set and the anti-gaming disable enforcement, expressed as
// filters on the plan instead of hidden checks inside mutators.
type sidecarOverlay struct {
	neverEnable   func(types.LinterName) bool
	unjustified   map[types.LinterName]bool
	hasSidecarDoc bool
}

// planDisableOverlay converts the sidecar's anti-gaming enforcement into
// changes: every unjustified disable is re-enabled and removed from the
// disable list. Without a sidecar, disables are respected.
func planDisableOverlay(observed *types.Config, overlay *sidecarOverlay) []Change {
	if overlay == nil || !overlay.hasSidecarDoc || len(observed.Linters.Disable) == 0 {
		return nil
	}

	var changes []Change

	disable := append([]types.LinterName{}, observed.Linters.Disable...)

	for _, name := range disable {
		if !overlay.unjustified[name] {
			continue
		}

		if overlay.neverEnable != nil && overlay.neverEnable(name) {
			continue
		}

		changes = append(changes,
			Change{
				Path: "linters.enable", Element: string(name), Kind: ChangeAdd,
				To:     string(name),
				Reason: "unjustified disable (no entry in sidecar)",
			},
			Change{
				Path: "linters.disable", Element: string(name), Kind: ChangeRemove,
				From: string(name), Reason: "unjustified disable re-enabled",
			},
		)
	}

	return changes
}

// ConvergedReport renders a human-readable plan summary.
func (r *ConvergeResult) Report() string {
	total := len(r.NormalizerChanges) + len(r.Plan)

	if total == 0 {
		return "config already converged"
	}

	return "would apply " + itoa(total) + " changes (" + itoa(len(r.Plan)) + " planned, " +
		itoa(len(r.NormalizerChanges)) + " normalized)"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	digits := ""

	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}

	return digits
}

// AuditActionsFor converts a convergence result into audit ledger actions.
// Audit = diff, always truthful: every mutation lands as the Change it is.
func AuditActionsFor(result *ConvergeResult) []auditAction {
	var actions []auditAction

	for _, change := range result.NormalizerChanges {
		actions = append(actions, auditActionFromChange(change, true))
	}

	for _, change := range result.Plan {
		actions = append(actions, auditActionFromChange(change, false))
	}

	return actions
}

// auditAction is the ledger-shaped projection of one change.
type auditAction struct {
	Linter string
	Action string
	Reason string
}

func auditActionFromChange(change Change, normalized bool) auditAction {
	action := map[ChangeKind]string{
		ChangeAdd:    "added-to-enable",
		ChangeUpdate: "updated",
		ChangeRemove: "removed",
	}[change.Kind]

	switch change.Path {
	case "formatters.enable":
		action = "formatter-added-to-enable"
	case "linters.enable":
		if normalized {
			action = "re-enabled"
		}
	case "linters.disable":
		action = "removed-from-disable"
	}

	return auditAction{
		Linter: change.Element,
		Action: action,
		Reason: change.Reason,
	}
}

var _ = constants.MinGolangCILintVersion
