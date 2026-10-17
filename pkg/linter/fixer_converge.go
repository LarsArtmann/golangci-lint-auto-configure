package linter

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// loadPolicy reads the disable-reason sidecar file (if it exists) alongside the
// config file. When the sidecar is present, the fixer projects it into the
// plan overlay: unjustified disables are re-enabled and never-enable linters
// are never added to enable. When the sidecar is absent, all disables are
// respected as before.
func (f *Fixer) loadPolicy(configPath string) {
	sidecarPath := filepath.Join(filepath.Dir(configPath), policy.SidecarFileName)

	pol, err := policy.Load(sidecarPath)
	if err != nil {
		f.logger.Warnf("⚠️  Failed to load policy file %s: %v", sidecarPath, err)

		f.pol = nil

		return
	}

	if pol != nil {
		f.logger.Infof("📋 Loaded policy from %s (%d justified disables, %d never-enable)",
			sidecarPath, len(pol.Disabled), len(pol.NeverEnable))
	}

	f.pol = pol
}

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

	// Analyzer contract: the reconciler needs the binary and a parsable
	// version. Failures carry the legacy classified exit codes
	// (Infrastructure 69 for a missing binary, Corruption 65 for garbage
	// version output) instead of degrading silently.
	if f.analyzer != nil {
		if err := f.verifyAnalyzerContract(ctx); err != nil {
			return nil, err
		}
	}

	if err := f.rescueOverspecifiedRunGo(cfg, configPath, dryRun); err != nil {
		return migrationError("rescue run.go", priority, dryRun, configPath, version, err)
	}

	f.loadPolicy(configPath)

	overlay := f.sidecarOverlay(cfg)

	working, normalizeChanges, phaseOneErr := f.convergePhaseOne(
		ctx, cfg, configPath, overlay, priority, dryRun, version,
	)
	if phaseOneErr != nil {
		return nil, phaseOneErr
	}

	return f.convergePhaseTwo(ctx, working, configPath, overlay, priority, dryRun, version, normalizeChanges)
}

// verifyAnalyzerContract proves the analyzer can run: the binary exists and
// its version parses. The classified failure carries the legacy exit codes
// (Infrastructure 69, Corruption 65).
func (f *Fixer) verifyAnalyzerContract(ctx context.Context) error {
	if err := f.analyzer.FindBinary(ctx); err != nil {
		return analysisError("find golangci-lint binary", 0, false, "", "", err)
	}

	if err := f.analyzer.CheckVersion(ctx); err != nil {
		return analysisError("check golangci-lint version", 0, false, "", "", err)
	}

	return nil
}

// convergePhaseOne normalizes WITHOUT analysis (golangci-lint cannot list
// linters for a config still naming removed ones) and persists the
// normalized state so phase two analyzes reality. Returns the working
// config for phase two.
func (f *Fixer) convergePhaseOne(
	ctx context.Context,
	cfg *types.Config,
	configPath string,
	overlay *sidecarOverlay,
	priority types.LinterPriority,
	dryRun bool,
	version string,
) (*types.Config, []Change, error) {
	var neverEnable func(types.LinterName) bool
	if overlay != nil {
		neverEnable = overlay.neverEnable
	}

	baseEnv, envErr := CollectEnvBase(ctx, f.analyzer, f.goVersionProvider, configPath)
	if envErr != nil {
		return nil, nil, analysisError("collect environment", priority, dryRun, configPath, version, envErr)
	}

	normalized, normalizeChanges := NormalizeConfig(cfg, baseEnv.Tool, neverEnable, baseEnv.BinaryKnown)
	if len(normalizeChanges) > 0 && !dryRun {
		if err := f.configLoader.SaveConfig(normalized, configPath); err != nil {
			return nil, nil, analysisError("save normalized config", priority, dryRun, configPath, version, err)
		}

		return normalized, normalizeChanges, nil
	}

	return cfg, normalizeChanges, nil
}

// convergePhaseTwo collects the full environment (with the analyzer run),
// converges, persists the result, and records the audit trail.
func (f *Fixer) convergePhaseTwo(
	ctx context.Context,
	cfg *types.Config,
	configPath string,
	overlay *sidecarOverlay,
	priority types.LinterPriority,
	dryRun bool,
	version string,
	normalizeChanges []Change,
) (*types.MigrationResult, error) {
	env, envErr := CollectEnv(ctx, f.analyzer, f.goVersionProvider, configPath)
	if envErr != nil {
		return migrationError("collect environment", priority, dryRun, configPath, version, envErr)
	}

	result, err := Converge(ctx, cfg, env, overlay, f.convergeOptions(priority, dryRun))
	if err != nil {
		return migrationError("converge config", priority, dryRun, configPath, version, err)
	}

	return f.finalizeConvergence(ctx, configPath, result, normalizeChanges, priority, dryRun, version)
}

// finalizeConvergence projects the convergence onto reports, logs, audit, and
// persistence: normalizations from phase one are attached (never
// double-counted), unsupported recommendations are surfaced, and non-dry-run
// results are saved with a post-apply load check.
func (f *Fixer) finalizeConvergence(
	ctx context.Context,
	configPath string,
	result *ConvergeResult,
	normalizeChanges []Change,
	priority types.LinterPriority,
	dryRun bool,
	version string,
) (*types.MigrationResult, error) {
	f.logMergedExclusionRules(result)
	f.warnUnsupportedRecommendations(result)

	// Normalizations already persisted in phase one must not be
	// double-counted. Phase one and Converge OVERLAP rather than replace
	// each other: both derive the four shared config passes, Converge alone
	// adds the redundancy pass. Union by change identity keeps every repair
	// exactly once; the overwrite this replaces silently discarded
	// Converge-only repairs (redundancy removals), short-circuiting apply
	// with "already converged" while the report-only path still planned the
	// same drift — a detect/apply divergence (2026-10-17 BuildFlow incident).
	result.NormalizerChanges = mergeNormalizerChanges(normalizeChanges, result.NormalizerChanges)

	total := len(result.NormalizerChanges) + len(result.Plan)

	if outcome := convergenceOutcome(total, result, dryRun); outcome != nil {
		return outcome, nil
	}

	return f.persistConvergence(ctx, configPath, result, total, priority, dryRun, version)
}

// mergeNormalizerChanges unions the phase-one normalizations with the
// in-memory normalizations Converge derived, deduplicating by change
// identity. Both sides derive the shared passes identically, so equal keys
// are the same repair seen twice; a key on only one side is that side's
// exclusive repair (config-shape repairs only phase one runs, the redundancy
// pass only Converge runs). Pure.
func mergeNormalizerChanges(persisted, derived []Change) []Change {
	merged := make([]Change, 0, len(persisted)+len(derived))
	seen := make(map[string]bool, len(persisted)+len(derived))

	for _, change := range persisted {
		seen[change.Key()] = true
		merged = append(merged, change)
	}

	for _, change := range derived {
		if seen[change.Key()] {
			continue
		}

		seen[change.Key()] = true
		merged = append(merged, change)
	}

	return merged
}

// persistConvergence saves the merged config and records the audit trail;
// it is the only non-dry-run path past the no-op outcome.
func (f *Fixer) persistConvergence(
	ctx context.Context,
	configPath string,
	result *ConvergeResult,
	total int,
	priority types.LinterPriority,
	dryRun bool,
	version string,
) (*types.MigrationResult, error) {
	if err := f.configLoader.SaveConfig(result.Merged, configPath); err != nil {
		return migrationError("save converged config", priority, dryRun, configPath, version, err)
	}

	f.recordConvergeAudit(result)
	f.verifyWrittenConfig(ctx, configPath)

	return &types.MigrationResult{
		FixesApplied: total,
		Message:      convergeMessage(total, result),
	}, nil
}

// warnUnsupportedRecommendations surfaces the recommendations withheld
// because the detected binary cannot name them (plugin linters on stock
// binaries): the project qualifies, the binary does not — silence here would
// read as "the tool sees nothing to add".
func (f *Fixer) warnUnsupportedRecommendations(result *ConvergeResult) {
	for _, name := range result.Unsupported {
		tool := "the detected golangci-lint"
		if !result.Env.Tool.IsZero() {
			tool = fmt.Sprintf("golangci-lint %s", result.Env.Tool)
		}

		f.logger.Warnf(
			"⚠️  %s is recommended for this project but %s does not include it; no enable entry was written "+
				"(build a custom golangci-lint binary with the linter as a module plugin to adopt it)",
			name, tool,
		)
	}
}

// configLoadVerifier is the post-apply tripwire surface of the concrete
// *Analyzer (duck-typed like the probe so test doubles skip it).
type configLoadVerifier interface {
	VerifyConfigLoads(ctx context.Context, configPath string) error
}

// verifyWrittenConfig proves the saved config loads with the detected
// binary. Best-effort by design: analyzers without the verifier skip it, and
// a failed verification warns loudly instead of failing the run (the write
// already happened; the check mode is the gating surface that repeats it).
func (f *Fixer) verifyWrittenConfig(ctx context.Context, configPath string) {
	verifier, ok := f.analyzer.(configLoadVerifier)
	if !ok {
		return
	}

	if err := verifier.VerifyConfigLoads(ctx, configPath); err != nil {
		f.logger.Warnf("⚠️  golangci-lint cannot load the config that was just written (%s):", configPath)
		f.logger.Warnf("   %v", err)
		f.logger.Warnf("   Re-run with --check and inspect the last change; this write breaks linting until repaired")
	}
}

// convergeOptions bundles the fixer's user-intent flips for one run.
func (f *Fixer) convergeOptions(priority types.LinterPriority, dryRun bool) ConvergeOptions {
	return ConvergeOptions{
		Priority:      priority,
		DryRun:        dryRun,
		ForceSettings: f.forceSettings,
		Pragmatic:     f.pragmatic,
	}
}

// convergenceOutcome returns the early MigrationResult for dry runs and
// no-op runs; nil means "persist the merged config".
func convergenceOutcome(total int, result *ConvergeResult, dryRun bool) *types.MigrationResult {
	if dryRun {
		return &types.MigrationResult{
			FixesApplied: total,
			Message:      dryRunMessage(total, result),
			NextSteps:    []string{"Run without --dry-run to apply the plan"},
			DryRun:       true,
		}
	}

	if total == 0 {
		return &types.MigrationResult{
			Message: "Config already converged - no changes needed",
		}
	}

	return nil
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

// logMergedExclusionRules surfaces the exclusion-rule unions at info level
// when --show-merged-rules is set. The audit ledger records the merges
// regardless of this flag.
func (f *Fixer) logMergedExclusionRules(result *ConvergeResult) {
	if !f.showMergedRules {
		return
	}

	for _, change := range result.Plan {
		if change.Path != PathLintersExclusionRules || change.Kind != ChangeUpdate {
			continue
		}

		from, fromOK := change.From.([]string)
		if !fromOK {
			continue
		}

		to, toOK := change.To.([]string)
		if !toOK {
			continue
		}

		added := addedStrings(from, to)
		if len(added) == 0 {
			continue
		}

		f.logger.Infof("\U0001f500 merged default exclusion rule %s: added linters %v", change.Element, added)
	}
}

// addedStrings returns the entries of after that are not in before, preserving order.
func addedStrings(before, after []string) []string {
	set := make(map[string]bool, len(before))
	for _, entry := range before {
		set[entry] = true
	}

	added := make([]string, 0, len(after))

	for _, entry := range after {
		if !set[entry] {
			added = append(added, entry)
		}
	}

	return added
}

// recordConvergeAudit emits one audit entry per applied change. Audit = diff.
func (f *Fixer) recordConvergeAudit(result *ConvergeResult) {
	for _, action := range AuditActionsFor(result) {
		f.ledger.Record(ledgerActionFor(action.Action), action.Linter, action.Reason)
	}
}

// ledgerActionMap maps the engine's audit-action vocabulary onto the
// ledger's Action union. Unknown actions fall back to "updated".
//
//nolint:gochecknoglobals // immutable vocabulary table
var ledgerActionMap = map[string]audit.Action{
	"added-to-enable":               audit.ActionAddedToEnable,
	"formatter-added-to-enable":     audit.ActionFormatterAddedToEnable,
	"formatter-removed-from-enable": audit.ActionFormatterRemovedFromEnable,
	"removed-from-enable":           audit.ActionRemovedFromEnable,
	"moved-to-disable":              audit.ActionMovedToDisable,
	"removed-from-disable":          audit.ActionRemovedFromDisable,
	"pruned-settings":               audit.ActionPrunedSettings,
	"re-enabled":                    audit.ActionReEnabled,
	"suppressed-re-enable":          audit.ActionSuppressedReEnable,
	"exclusion-linter-renamed":      audit.ActionExclusionLinterRenamed,
	"exclusion-linter-removed":      audit.ActionExclusionLinterRemoved,
	"exclusion-rule-merged":         audit.ActionExclusionRuleMerged,
}

func ledgerActionFor(action string) audit.Action {
	if mapped, known := ledgerActionMap[action]; known {
		return mapped
	}

	return audit.Action("updated")
}
