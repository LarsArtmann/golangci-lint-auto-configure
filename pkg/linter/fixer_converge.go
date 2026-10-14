package linter

import (
	"context"
	"strconv"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// convergeFixConfig drives the declarative reconciler end to end:
// rescue (unloadable run.go) -> converge (collect -> normalize -> render ->
// plan -> apply) -> save -> audit. I/O and audit live HERE and only here; the
// pipeline itself (Converge) is pure.
func (f *Fixer) convergeFixConfig(
	ctx context.Context,
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (*types.MigrationResult, error) {
	version := f.detectVersion(ctx)

	if err := f.rescueOverspecifiedRunGo(cfg, configPath, dryRun); err != nil {
		return migrationError("rescue run.go", priority, dryRun, configPath, version, err)
	}

	f.loadPolicy(configPath)

	overlay := f.sidecarOverlay(cfg)

	var neverEnable func(types.LinterName) bool
	if overlay != nil {
		neverEnable = overlay.neverEnable
	}

	// Phase 1: normalize WITHOUT analysis (golangci-lint cannot list linters
	// for a config still naming removed ones), persisting the normalized
	// state so phase 2 analyzes reality.
	baseEnv, envErr := CollectEnvBase(ctx, f.analyzer, f.goVersionProvider, configPath)
	if envErr != nil {
		return migrationError("collect environment", priority, dryRun, configPath, version, envErr)
	}

	normalized, normalizeChanges := NormalizeConfig(cfg, baseEnv.Tool, neverEnable)
	if len(normalizeChanges) > 0 && !dryRun {
		if err := f.configLoader.SaveConfig(normalized, configPath); err != nil {
			return migrationError("save normalized config", priority, dryRun, configPath, version, err)
		}

		cfg = normalized
	}

	// Phase 2: full environment (with the analyzer run) and the plan.
	env, envErr := CollectEnv(ctx, f.analyzer, f.goVersionProvider, configPath)
	if envErr != nil {
		return migrationError("collect environment", priority, dryRun, configPath, version, envErr)
	}

	result, err := Converge(ctx, cfg, env, overlay, ConvergeOptions{
		Priority:      priority,
		DryRun:        dryRun,
		ForceSettings: f.forceSettings,
		Pragmatic:     f.pragmatic,
	})
	if err != nil {
		return migrationError("converge config", priority, dryRun, configPath, version, err)
	}

	// Normalizations already persisted in phase 1 must not be double-counted.
	result.NormalizerChanges = normalizeChanges

	total := len(result.NormalizerChanges) + len(result.Plan)

	if dryRun {
		return &types.MigrationResult{
			FixesApplied: total,
			Message:      dryRunMessage(total, result),
			NextSteps:    []string{"Run without --dry-run to apply the plan"},
			DryRun:       true,
		}, nil
	}

	if total == 0 {
		return &types.MigrationResult{
			Message: "Config already converged - no changes needed",
		}, nil
	}

	if err := f.configLoader.SaveConfig(result.Merged, configPath); err != nil {
		return migrationError("save converged config", priority, dryRun, configPath, version, err)
	}

	f.recordConvergeAudit(result)

	return &types.MigrationResult{
		FixesApplied: total,
		Message:      convergeMessage(total, result),
	}, nil
}

func dryRunMessage(total int, result *ConvergeResult) string {
	if total == 0 {
		return "Dry run: config already converged - no changes needed"
	}

	return "Dry run: " + result.Report()
}

func convergeMessage(total int, result *ConvergeResult) string {
	var message strings.Builder

	message.WriteString("Converged config: applied ")
	message.WriteString(strconv.Itoa(total))
	message.WriteString(" changes (")
	message.WriteString(strconv.Itoa(len(result.NormalizerChanges)))
	message.WriteString(" normalized, ")
	message.WriteString(strconv.Itoa(len(result.Plan)))
	message.WriteString(" planned)")

	return message.String()
}

// sidecarOverlay projects the loaded policy into the plan-level overlay: the
// never-enable set and the unjustified-disable map (every disable entry that
// lacks a sidecar justification when a sidecar is present).
func (f *Fixer) sidecarOverlay(cfg *types.Config) *sidecarOverlay {
	if f.pol == nil {
		return nil
	}

	overlay := &sidecarOverlay{
		neverEnable:   f.pol.IsNeverEnable,
		unjustified:   map[types.LinterName]bool{},
		hasSidecarDoc: true,
	}

	for _, name := range f.configLoader.GetLintersDisabled(cfg) {
		if !f.pol.IsJustified(name) {
			overlay.unjustified[name] = true
		}
	}

	return overlay
}

// recordConvergeAudit emits one audit entry per applied change. Audit = diff.
func (f *Fixer) recordConvergeAudit(result *ConvergeResult) {
	for _, action := range AuditActionsFor(result) {
		f.ledger.Record(ledgerActionFor(action.Action), action.Linter, action.Reason)
	}
}

func ledgerActionFor(action string) audit.Action {
	switch action {
	case "added-to-enable":
		return audit.ActionAddedToEnable
	case "formatter-added-to-enable":
		return audit.ActionFormatterAddedToEnable
	case "removed-from-disable":
		return audit.ActionRemovedFromDisable
	case "re-enabled":
		return audit.ActionReEnabled
	default:
		return audit.Action("updated")
	}
}
