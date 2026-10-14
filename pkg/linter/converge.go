package linter

import (
	"context"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Engine names the configuration strategy the fixer drives.
type Engine string

const (
	// EngineConverge is the declarative reconciler (collect → normalize →
	// render → plan → apply) and the DEFAULT engine since the corpus flip:
	// TestParityCorpusConverge proves byte-identical output across the
	// parity corpus.
	EngineConverge Engine = "converge"
	// EngineLegacy is the historical imperative mutator pipeline, kept as an
	// explicit opt-out during the migration window.
	EngineLegacy Engine = "legacy"
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
	// Suppressed lists linters the anti-gaming overlay would have
	// re-enabled but the sidecar's never-enable section protected. They are
	// audit-only events: no config change accompanies them.
	Suppressed []types.LinterName
	// Merged is the converged configuration (nil for dry runs).
	Merged *types.Config
	// Converged reports whether the observed config already matched.
	Converged bool
}

// NormalizeConfig runs every pre-analysis normalizer: config-shape repairs
// (version, duration, typecheck, tool-disabled moves), bad settings keys,
// deprecation replacement (version-gated, never-enable aware), and exclusion
// rule renames. Pure: clone in, changes out. The result must be PERSISTED
// before golangci-lint can list linters for it.
func NormalizeConfig(
	observed *types.Config,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
) (*types.Config, []Change) {
	working, changes := NormalizeConfigShape(observed)

	working, badKeyChanges := NormalizeBadSettingsKeys(working)
	changes = append(changes, badKeyChanges...)

	working, replacementChanges := ReplaceDeprecatedLinters(working, tool, neverEnable)
	changes = append(changes, replacementChanges...)

	working, ruleRenameChanges := RenameDeprecatedInExclusionRules(working)
	changes = append(changes, ruleRenameChanges...)

	if len(changes) == 0 {
		return observed, nil
	}

	return working, changes
}

// Converge computes the desired state for env, plans the move from observed,
// and returns the merged config. Pure with respect to disk: nothing here
// reads or writes files — the fixer owns I/O and audit emission. env must
// already carry the analyzer's enabled-linter set for the (normalized)
// config on disk.
func Converge(
	ctx context.Context,
	observed *types.Config,
	env Env,
	sidecar *sidecarOverlay,
	options ConvergeOptions,
) (*ConvergeResult, error) {
	working, normalizerChanges := NormalizeBadSettingsKeys(observed)

	var neverEnable func(types.LinterName) bool
	if sidecar != nil {
		neverEnable = sidecar.neverEnable
	}

	working, replacementChanges := ReplaceDeprecatedLinters(working, env.Tool, neverEnable)
	normalizerChanges = append(normalizerChanges, replacementChanges...)

	working, ruleRenameChanges := RenameDeprecatedInExclusionRules(working)
	normalizerChanges = append(normalizerChanges, ruleRenameChanges...)

	// The disable list minus the enable list is user intent; the renderer
	// must not recommend those linters back (enable/disable conflicts are
	// resolved by the plan below instead).
	userDisabled := disableMinusEnable(working)

	// Redundancy removal needs the desired formatter stack, which needs the
	// normalized world; run it just before rendering.
	preDesired := Desired(env,
		WithPragmatic(options.Pragmatic),
		WithMinPriority(options.Priority),
		WithUserEnabled(working.Linters.Enable),
		WithUserDisabled(userDisabled),
	)
	working, redundancyChanges := RemoveRedundancies(working, preDesired.EnableFormatters)
	normalizerChanges = append(normalizerChanges, redundancyChanges...)

	desired := Desired(env,
		WithPragmatic(options.Pragmatic),
		WithMinPriority(options.Priority),
		WithUserEnabled(working.Linters.Enable),
		WithUserDisabled(userDisabled),
	)

	strategies := StrategiesFor(MergeOptions{ForceSettings: options.ForceSettings})

	plan := PlanDesired(working, desired, strategies)

	overlayChanges, suppressed := planDisableOverlay(working, sidecar, env)
	plan = append(plan, overlayChanges...)

	// The final enable set decides the two closing plans: enable/disable
	// conflict resolution and orphaned-settings pruning. Overlay adds count
	// as enabled (the enforcement result).
	finalEnable := types.NewSet(working.Linters.Enable...)
	for _, name := range desired.EnableLinters {
		finalEnable.Add(name)
	}
	for _, change := range overlayChanges {
		if change.Path == "linters.enable" && change.Kind == ChangeAdd {
			finalEnable.Add(types.LinterName(change.Element))
		}
	}

	plan = append(plan, planEnableDisableConflicts(working, finalEnable, overlayChanges)...)
	plan = append(plan, planOrphanSettings(working, finalEnable)...)
	sortChanges(plan)

	if err := ValidatePlan(plan); err != nil {
		return nil, err
	}

	converged := len(plan) == 0 && len(normalizerChanges) == 0

	result := &ConvergeResult{
		Env:               env,
		Plan:              plan,
		NormalizerChanges: normalizerChanges,
		Suppressed:        suppressed,
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
// changes: every unjustified disable is removed from the disable list and
// re-enabled UNLESS the binary already enables it by default (writing a
// default-enabled linter would churn every config). The never-enable section
// blocks the re-enable outright and reports the linter as suppressed (an
// audit-only event, mirroring the legacy enforcement ledger entry). Without
// a sidecar, disables are respected.
func planDisableOverlay(
	observed *types.Config,
	overlay *sidecarOverlay,
	env Env,
) ([]Change, []types.LinterName) {
	if overlay == nil || !overlay.hasSidecarDoc || len(observed.Linters.Disable) == 0 {
		return nil, nil
	}

	var changes []Change
	var suppressed []types.LinterName

	disable := append([]types.LinterName{}, observed.Linters.Disable...)

	for _, name := range disable {
		if !overlay.unjustified[name] {
			continue
		}

		if overlay.neverEnable != nil && overlay.neverEnable(name) {
			suppressed = append(suppressed, name)

			continue
		}

		if !env.AnalyzerEnabled[name] {
			changes = append(changes, Change{
				Path: "linters.enable", Element: string(name), Kind: ChangeAdd,
				To:     string(name),
				Reason: "unjustified disable (no entry in sidecar)",
			})
		}

		changes = append(changes, Change{
			Path: "linters.disable", Element: string(name), Kind: ChangeRemove,
			From: string(name), Reason: "unjustified disable re-enabled",
		})
	}

	return changes, suppressed
}

// disableMinusEnable returns the config's disable list minus its enable
// list: a linter in both is a conflict the plan resolves (enable wins), not
// a user disable.
func disableMinusEnable(cfg *types.Config) []types.LinterName {
	enabled := types.NewSet(cfg.Linters.Enable...)

	var disabled []types.LinterName

	for _, name := range cfg.Linters.Disable {
		if !enabled.Contains(name) {
			disabled = append(disabled, name)
		}
	}

	return disabled
}

// planEnableDisableConflicts removes every disable entry whose linter ends
// up enabled: a linter cannot be both enabled and disabled, and the legacy
// engine resolved the contradiction in favor of enable.
func planEnableDisableConflicts(
	observed *types.Config,
	finalEnable types.Set[types.LinterName],
	overlayChanges []Change,
) []Change {
	alreadyRemoved := make(map[types.LinterName]bool, len(overlayChanges))

	for _, change := range overlayChanges {
		if change.Path == "linters.disable" && change.Kind == ChangeRemove {
			alreadyRemoved[types.LinterName(change.Element)] = true
		}
	}

	var changes []Change

	for _, name := range observed.Linters.Disable {
		if alreadyRemoved[name] || !finalEnable.Contains(name) {
			continue
		}

		changes = append(changes, Change{
			Path: "linters.disable", Element: string(name), Kind: ChangeRemove,
			From:   string(name),
			Reason: "linter is enabled; a linter cannot be both enabled and disabled",
		})
	}

	return changes
}

// planOrphanSettings removes settings blocks for linters outside the final
// enable set: disabled and unenabled linters carry dead weight settings that
// imply activity where there is none (the legacy prune passes, as one plan).
func planOrphanSettings(observed *types.Config, finalEnable types.Set[types.LinterName]) []Change {
	if len(observed.Linters.Settings) == 0 {
		return nil
	}

	var changes []Change

	for _, name := range sortedSettingNames(observed.Linters.Settings) {
		if finalEnable.Contains(types.LinterName(name)) {
			continue
		}

		changes = append(changes, Change{
			Path: "linters.settings", Element: name, Kind: ChangeRemove,
			Reason: "orphaned settings (linter not in the enable set)",
		})
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
// Audit = diff, always truthful: every mutation lands as the Change it is,
// and the suppressed re-enables (never-enable blocks) land as the events
// they are.
func AuditActionsFor(result *ConvergeResult) []auditAction {
	var actions []auditAction

	for _, change := range result.NormalizerChanges {
		actions = append(actions, auditActionFromChange(change))
	}

	for _, change := range result.Plan {
		actions = append(actions, auditActionFromChange(change))
	}

	for _, name := range result.Suppressed {
		actions = append(actions, auditAction{
			Linter: string(name),
			Action: "suppressed-re-enable",
			Reason: "never-enable sidecar entry blocks enforcement re-enable",
		})
	}

	return actions
}

// auditAction is the ledger-shaped projection of one change.
type auditAction struct {
	Linter string
	Action string
	Reason string
}

// auditActionFromChange maps one planned change to its audit vocabulary. The
// linter name for exclusion-rule renames is the predecessor (the "from" side
// of the "pred->succ" element).
func auditActionFromChange(change Change) auditAction {
	action := "updated"

	switch {
	case change.Path == "linters.enable" && change.Kind == ChangeAdd:
		action = "added-to-enable"

		// The anti-gaming overlay's adds are re-enables of user-disabled
		// linters, not recommendations; the reason prefix distinguishes them
		// from ordinary renderer adds.
		if strings.HasPrefix(change.Reason, "unjustified") {
			action = "re-enabled"
		}
	case change.Path == "linters.enable" && change.Kind == ChangeRemove:
		action = "removed-from-enable"
	case change.Path == "linters.disable" && change.Kind == ChangeAdd:
		action = "moved-to-disable"
	case change.Path == "linters.disable" && change.Kind == ChangeRemove:
		action = "removed-from-disable"
	case change.Path == "formatters.enable" && change.Kind == ChangeAdd:
		action = "formatter-added-to-enable"
	case change.Path == "formatters.enable" && change.Kind == ChangeRemove:
		action = "formatter-removed-from-enable"
	case change.Path == "linters.settings" && change.Kind == ChangeRemove:
		action = "pruned-settings"
	case change.Path == "linters.exclusions.rules" && change.Kind == ChangeUpdate:
		if _, _, isRename := strings.Cut(change.Element, "->"); isRename {
			action = "exclusion-linter-renamed"
		} else {
			action = "exclusion-rule-merged"
		}
	case change.Path == "linters.exclusions.rules" && change.Kind == ChangeRemove:
		action = "exclusion-linter-removed"
	}

	linter := change.Element
	reason := change.Reason

	if action == "exclusion-linter-renamed" {
		if from, _, found := strings.Cut(change.Element, "->"); found {
			linter = from
		}
	}

	if action == "exclusion-rule-merged" {
		reason = exclusionMergeReason(change)
	}

	return auditAction{
		Linter: linter,
		Action: action,
		Reason: reason,
	}
}

// exclusionMergeReason names the linters the rule merge actually added, so
// the ledger answers "what did the tool union into my rule?".
func exclusionMergeReason(change Change) string {
	reason := change.Reason

	from, fromOK := change.From.([]string)
	to, toOK := change.To.([]string)

	if !fromOK || !toOK {
		return reason
	}

	existing := make(map[string]bool, len(from))
	for _, entry := range from {
		existing[entry] = true
	}

	var added []string

	for _, entry := range to {
		if !existing[entry] {
			added = append(added, entry)
		}
	}

	if len(added) == 0 {
		return reason
	}

	return reason + " (added: " + strings.Join(added, ", ") + ")"
}

var _ = constants.MinGolangCILintVersion
