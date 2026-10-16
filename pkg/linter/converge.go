package linter

import (
	"context"
	"strconv"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
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
	// Unsupported lists recommendable linters the detected binary cannot
	// name (plugin linters). Nothing was written for them; the fixer logs a
	// warning and finding-based reports carry one finding each.
	Unsupported []types.LinterName
	// Merged is the converged configuration (nil for dry runs).
	Merged *types.Config
	// Converged reports whether the observed config already matched.
	Converged bool
}

// NormalizeConfig runs every pre-analysis normalizer: config-shape repairs
// (version, duration, typecheck, tool-disabled moves), bad settings keys,
// deprecation replacement (version-gated, never-enable aware), unknown-linter
// removal (capability-gated), and exclusion rule renames. Pure: clone in,
// changes out. The result must be PERSISTED before golangci-lint can list
// linters for it.
func NormalizeConfig(
	observed *types.Config,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
	binaryKnown map[types.LinterName]bool,
) (*types.Config, []Change) {
	working, changes := NormalizeConfigShape(observed)

	working, badKeyChanges := NormalizeBadSettingsKeys(working)
	changes = append(changes, badKeyChanges...)

	working, replacementChanges := ReplaceDeprecatedLinters(working, tool, neverEnable)
	changes = append(changes, replacementChanges...)

	working, unknownChanges := RemoveUnknownLinters(working, binaryKnown)
	changes = append(changes, unknownChanges...)

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
	var neverEnable func(types.LinterName) bool
	if sidecar != nil {
		neverEnable = sidecar.neverEnable
	}

	working, normalizerChanges := convergeNormalizers(observed, env, neverEnable)

	// The disable list minus the enable list is user intent; the renderer
	// must not recommend those linters back (enable/disable conflicts are
	// resolved by the plan instead).
	userDisabled := disableMinusEnable(working)

	working, redundancyChanges := removeConvergeRedundancies(working, env, options, userDisabled)
	normalizerChanges = append(normalizerChanges, redundancyChanges...)

	plan, suppressed, unsupported, planErr := planConvergence(working, env, sidecar, options, userDisabled)
	if planErr != nil {
		return nil, planErr
	}

	return newConvergeResult(env, working, plan, normalizerChanges, suppressed, unsupported, options), nil
}

// newConvergeResult assembles the run outcome; the merged config is
// materialized only for real (non-dry-run) runs.
func newConvergeResult(
	env Env,
	working *types.Config,
	plan []Change,
	normalizerChanges []Change,
	suppressed []types.LinterName,
	unsupported []types.LinterName,
	options ConvergeOptions,
) *ConvergeResult {
	var merged *types.Config
	if !options.DryRun {
		merged = ApplyPlan(working, plan)
	}

	return &ConvergeResult{
		Env:               env,
		Plan:              plan,
		NormalizerChanges: normalizerChanges,
		Suppressed:        suppressed,
		Unsupported:       unsupported,
		Merged:            merged,
		Converged:         len(plan) == 0 && len(normalizerChanges) == 0,
	}
}

// convergeNormalizers runs the pre-render normalizer chain: bad settings
// keys, deprecation replacement (version-gated, never-enable aware),
// unknown-linter removal (capability-gated), and exclusion rule renames. The
// config-shape repairs live in NormalizeConfig, which the fixer runs in its
// persistence phase.
func convergeNormalizers(
	observed *types.Config,
	env Env,
	neverEnable func(types.LinterName) bool,
) (*types.Config, []Change) {
	working, changes := NormalizeBadSettingsKeys(observed)

	working, replacementChanges := ReplaceDeprecatedLinters(working, env.Tool, neverEnable)
	changes = append(changes, replacementChanges...)

	working, unknownChanges := RemoveUnknownLinters(working, env.BinaryKnown)
	changes = append(changes, unknownChanges...)

	working, renameChanges := RenameDeprecatedInExclusionRules(working)

	return working, append(changes, renameChanges...)
}

// removeConvergeRedundancies drops linters and formatters superseded by the
// desired formatter stack (lll with golines, gofmt with gofumpt). It runs
// just before rendering: redundancy removal needs the desired formatters,
// which need the normalized world.
func removeConvergeRedundancies(
	working *types.Config,
	env Env,
	options ConvergeOptions,
	userDisabled []types.LinterName,
) (*types.Config, []Change) {
	preDesired := renderDesired(env, options, working, userDisabled)

	return RemoveRedundancies(working, preDesired.EnableFormatters)
}

// planConvergence renders the desired state, plans the move, and appends the
// closing overlays: anti-gaming enforcement, enable/disable conflict
// resolution, and orphaned-settings pruning. The returned plan is validated.
func planConvergence(
	working *types.Config,
	env Env,
	sidecar *sidecarOverlay,
	options ConvergeOptions,
	userDisabled []types.LinterName,
) ([]Change, []types.LinterName, []types.LinterName, error) {
	desired := renderDesired(env, options, working, userDisabled)

	strategies := StrategiesFor(MergeOptions{
		ForceSettings: options.ForceSettings,
		Pragmatic:     options.Pragmatic,
	})

	plan := PlanDesired(working, desired, strategies)

	overlayChanges, suppressed := planDisableOverlay(working, sidecar, env)
	plan = append(plan, overlayChanges...)
	plan = append(plan, closingPlans(working, desired, overlayChanges)...)

	sortChanges(plan)

	if err := ValidatePlan(plan); err != nil {
		return nil, nil, nil, err
	}

	return plan, suppressed, desired.Unsupported, nil
}

// renderDesired renders the desired configuration with the run's user-intent
// options and the working config's enable/disable state.
func renderDesired(
	env Env,
	options ConvergeOptions,
	working *types.Config,
	userDisabled []types.LinterName,
) DesiredConfig {
	return Desired(env,
		WithPragmatic(options.Pragmatic),
		WithMinPriority(options.Priority),
		WithUserEnabled(working.Linters.Enable),
		WithUserDisabled(userDisabled),
	)
}

// closingPlans appends the two enable-set-dependent plans. The final enable
// set decides both: overlay adds count as enabled (they ARE the enforcement
// result).
func closingPlans(working *types.Config, desired DesiredConfig, overlayChanges []Change) []Change {
	finalEnable := finalEnableSet(working, desired, overlayChanges)

	changes := planEnableDisableConflicts(working, finalEnable, overlayChanges)
	changes = append(changes, planOrphanSettings(working, finalEnable)...)

	return changes
}

// finalEnableSet unions the observed enable list, the rendered enable set,
// and the overlay's enable adds.
func finalEnableSet(
	working *types.Config,
	desired DesiredConfig,
	overlayChanges []Change,
) types.Set[types.LinterName] {
	finalEnable := types.NewSet(working.Linters.Enable...)

	for _, name := range desired.EnableLinters {
		finalEnable.Add(name)
	}

	for _, change := range overlayChanges {
		if change.Path == PathLintersEnable && change.Kind == ChangeAdd {
			finalEnable.Add(types.LinterName(change.Element))
		}
	}

	return finalEnable
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

	for _, name := range observed.Linters.Disable {
		if !overlay.unjustified[name] {
			continue
		}

		if overlay.neverEnable != nil && overlay.neverEnable(name) {
			suppressed = append(suppressed, name)

			continue
		}

		changes = append(changes, overlayReEnableChanges(name, env)...)
	}

	return changes, suppressed
}

// overlayReEnableChanges plans one unjustified disable's undo: the enable
// add (skipped when the binary already enables the linter by default) plus
// the disable removal.
func overlayReEnableChanges(name types.LinterName, env Env) []Change {
	var changes []Change

	if !env.AnalyzerEnabled[name] {
		changes = append(changes, Change{
			Path: PathLintersEnable, Element: string(name), Kind: ChangeAdd,
			To:     string(name),
			Reason: "unjustified disable (no entry in sidecar)",
		})
	}

	changes = append(changes, Change{
		Path: PathLintersDisable, Element: string(name), Kind: ChangeRemove,
		From: string(name), Reason: "unjustified disable re-enabled",
	})

	return changes
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
		if change.Path == PathLintersDisable && change.Kind == ChangeRemove {
			alreadyRemoved[types.LinterName(change.Element)] = true
		}
	}

	var changes []Change

	for _, name := range observed.Linters.Disable {
		if alreadyRemoved[name] || !finalEnable.Contains(name) {
			continue
		}

		changes = append(changes, Change{
			Path: PathLintersDisable, Element: string(name), Kind: ChangeRemove,
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
			Path: PathLintersSettings, Element: name, Kind: ChangeRemove,
			Reason: "orphaned settings (linter not in the enable set)",
		})
	}

	return changes
}

// Report renders a human-readable plan summary.
func (r *ConvergeResult) Report() string {
	total := len(r.NormalizerChanges) + len(r.Plan)

	if total == 0 {
		return "config already converged"
	}

	return "would apply " + strconv.Itoa(total) + " changes (" + strconv.Itoa(len(r.Plan)) + " planned, " +
		strconv.Itoa(len(r.NormalizerChanges)) + " normalized)"
}

// AuditActionsFor converts a convergence result into audit ledger actions.
// Audit = diff, always truthful: every mutation lands as the Change it is,
// and the suppressed re-enables (never-enable blocks) land as the events
// they are.
func AuditActionsFor(result *ConvergeResult) []auditAction {
	actions := make(
		[]auditAction, 0,
		len(result.NormalizerChanges)+len(result.Plan)+len(result.Suppressed),
	)

	for _, change := range result.NormalizerChanges {
		actions = append(actions, auditActionFromChange(change))
	}

	for _, change := range result.Plan {
		actions = append(actions, auditActionFromChange(change))
	}

	for _, name := range result.Suppressed {
		actions = append(actions, auditAction{
			Linter: string(name),
			Action: suppressedReEnableAction,
			Reason: suppressedReEnableReason,
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

// Audit action vocabulary shared by the mapping table and the special
// cases.
const (
	actionUpdated               = "updated"
	actionAddedToEnable         = "added-to-enable"
	actionReEnabled             = "re-enabled"
	actionExclusionRenamed      = "exclusion-linter-renamed"
	actionExclusionRemoved      = "exclusion-linter-removed"
	actionExclusionRuleMerged   = "exclusion-rule-merged"
	suppressedReEnableAction    = "suppressed-re-enable"
	suppressedReEnableReason    = "never-enable sidecar entry blocks enforcement re-enable"
	unjustifiedDisableReasonPfx = "unjustified"
)

// auditActionMap keys "path|kind" to the ledger action. Rows outside the
// table fall through to the exclusion-aware fallback and then "updated".
//
//nolint:gochecknoglobals // immutable audit vocabulary table
var auditActionMap = map[string]string{
	PathLintersEnable + "|" + string(ChangeAdd):            actionAddedToEnable,
	PathLintersEnable + "|" + string(ChangeRemove):         "removed-from-enable",
	PathLintersDisable + "|" + string(ChangeAdd):           "moved-to-disable",
	PathLintersDisable + "|" + string(ChangeRemove):        "removed-from-disable",
	PathFormattersEnable + "|" + string(ChangeAdd):         "formatter-added-to-enable",
	PathFormattersEnable + "|" + string(ChangeRemove):      "formatter-removed-from-enable",
	PathLintersSettings + "|" + string(ChangeRemove):       "pruned-settings",
	PathLintersExclusionRules + "|" + string(ChangeRemove): actionExclusionRemoved,
}

// auditActionFromChange maps one planned change to its audit vocabulary.
// Exclusion-rule updates resolve by element shape ("pred->succ" is a
// rename, a RuleKey is a merge), and the anti-gaming overlay's adds are
// re-enables, not recommendations (the reason prefix tells them apart).
func auditActionFromChange(change Change) auditAction {
	action := auditActionMap[change.Path+"|"+string(change.Kind)]
	linter := change.Element
	reason := change.Reason

	if action == "" {
		action, linter = auditExclusionUpdateAction(change)
	}

	if action == actionAddedToEnable && strings.HasPrefix(change.Reason, unjustifiedDisableReasonPfx) {
		action = actionReEnabled
	}

	if action == actionExclusionRuleMerged {
		reason = exclusionMergeReason(change)
	}

	return auditAction{Linter: linter, Action: action, Reason: reason}
}

// auditExclusionUpdateAction maps exclusion-rule ChangeUpdates: a rename
// names the predecessor in the ledger, a merge names the RuleKey.
func auditExclusionUpdateAction(change Change) (string, string) {
	if change.Path != PathLintersExclusionRules || change.Kind != ChangeUpdate {
		return actionUpdated, change.Element
	}

	if from, _, isRename := strings.Cut(change.Element, "->"); isRename {
		return actionExclusionRenamed, from
	}

	return actionExclusionRuleMerged, change.Element
}

// exclusionMergeReason names the linters the rule merge actually added, so
// the ledger answers "what did the tool union into my rule?".
func exclusionMergeReason(change Change) string {
	reason := change.Reason

	from, fromOK := change.From.([]string)
	union, unionOK := change.To.([]string)

	if !fromOK || !unionOK {
		return reason
	}

	existing := make(map[string]bool, len(from))
	for _, entry := range from {
		existing[entry] = true
	}

	var added []string

	for _, entry := range union {
		if !existing[entry] {
			added = append(added, entry)
		}
	}

	if len(added) == 0 {
		return reason
	}

	return reason + " (added: " + strings.Join(added, ", ") + ")"
}
