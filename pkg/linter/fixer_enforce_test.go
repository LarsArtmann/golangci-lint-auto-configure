package linter

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// recordedEnforceAction captures a single audit ledger Record call.
type recordedEnforceAction struct {
	action audit.Action
	linter string
	reason string
}

// enforceRecorder is a test double for audit.Recorder that records every call.
type enforceRecorder struct {
	actions           []recordedEnforceAction
	previouslyEnabled map[string]bool
}

func (r *enforceRecorder) Record(action audit.Action, linter, reason string) {
	r.actions = append(r.actions, recordedEnforceAction{action: action, linter: linter, reason: reason})
}

// PreviouslyAutoEnabled implements ledgerReader so cycle-detection tests
// can inject a fake "previously auto-enabled" set.
func (r *enforceRecorder) PreviouslyAutoEnabled() map[string]bool {
	return r.previouslyEnabled
}

func (r *enforceRecorder) hasReEnable(linter string) bool {
	for _, a := range r.actions {
		if a.action == audit.ActionReEnabled && a.linter == linter {
			return true
		}
	}

	return false
}

func (r *enforceRecorder) hasSuppressedReEnable(linter string) bool {
	for _, a := range r.actions {
		if a.action == audit.ActionSuppressedReEnable && a.linter == linter {
			return true
		}
	}

	return false
}

// newEnforceFixer builds a Fixer with only the fields the policy-enforcement
// code path touches (logger + ledger). configLoader/analyzer stay nil —
// enforceDisableReasons/loadPolicy/tryReEnableLinter never reach them.
func newEnforceFixer() *Fixer {
	return &Fixer{
		logger: NewTestLogger(),
		ledger: &enforceRecorder{},
	}
}

func (f *Fixer) recorder() *enforceRecorder {
	if r, ok := f.ledger.(*enforceRecorder); ok {
		return r
	}

	return &enforceRecorder{}
}

func writeSidecar(t *testing.T, dir, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, policy.SidecarFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
}

func TestIsToolLevelManaged(t *testing.T) {
	tests := []struct {
		name   string
		linter types.LinterName
		want   bool
	}{
		{"funcorder is tool-level managed (forcibly disabled)", "funcorder", true},
		{"noinlineerr is tool-level managed (forcibly disabled)", "noinlineerr", true},
		{"depguard is tool-level managed (forcibly disabled)", "depguard", true},
		{"exhaustruct is tool-level managed (never-auto-enable)", "exhaustruct", true},
		{"errcheck is not tool-level managed", "errcheck", false},
		{"gofmt is not tool-level managed", "gofmt", false},
		{"unknown linter is not tool-level managed", "does-not-exist", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isToolLevelManaged(tt.linter); got != tt.want {
				t.Errorf("isToolLevelManaged(%q) = %v, want %v", tt.linter, got, tt.want)
			}
		})
	}
}

func TestLoadPolicy_NoSidecar(t *testing.T) {
	f := newEnforceFixer()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	f.loadPolicy(configPath)

	if f.pol != nil {
		t.Errorf("expected nil policy when no sidecar exists, got %+v", f.pol)
	}
}

func TestLoadPolicy_WithSidecar(t *testing.T) {
	f := newEnforceFixer()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	writeSidecar(t, dir, "disabled:\n  gofmt:\n    reason: prefer golines\n    category: convention\n")

	f.loadPolicy(configPath)

	if f.pol == nil {
		t.Fatal("expected non-nil policy when sidecar exists")
	}

	if !f.pol.IsJustified("gofmt") {
		t.Error("expected gofmt to be justified by sidecar")
	}

	if f.pol.IsJustified("errcheck") {
		t.Error("errcheck should not be justified")
	}
}

func TestLoadPolicy_MalformedSidecar(t *testing.T) {
	f := newEnforceFixer()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	writeSidecar(t, dir, "}{not valid yaml ][\n")

	f.loadPolicy(configPath)

	if f.pol != nil {
		t.Errorf("expected nil policy on malformed sidecar (error warned, not fatal), got %+v", f.pol)
	}
}

func TestEnforceDisableReasons_NoPolicy(t *testing.T) {
	f := newEnforceFixer()
	// f.pol is nil by default — enforcement must be a no-op.

	cfg := &types.Config{Linters: types.LintersConfig{Disable: []types.LinterName{"errcheck", "gofmt"}}}

	if count := f.enforceDisableReasons(cfg); count != 0 {
		t.Fatalf("expected 0 re-enables with no policy, got %d", count)
	}

	if len(cfg.Linters.Disable) != 2 {
		t.Errorf("disable list should be unchanged, got %v", cfg.Linters.Disable)
	}

	if len(cfg.Linters.Enable) != 0 {
		t.Errorf("enable list should be empty, got %v", cfg.Linters.Enable)
	}
}

func TestEnforceDisableReasons_ReEnablesUnjustified(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{Disabled: map[types.LinterName]policy.DisableJustification{
		"gofmt": {Reason: "prefer golines", Category: policy.CategoryConvention},
	}}

	cfg := &types.Config{Linters: types.LintersConfig{
		Disable: []types.LinterName{"gofmt", "errcheck"},
	}}

	count := f.enforceDisableReasons(cfg)

	if count != 1 {
		t.Fatalf("expected 1 re-enable (errcheck), got %d", count)
	}

	if !sliceHas(cfg.Linters.Enable, "errcheck") {
		t.Errorf("errcheck should be re-enabled; enable=%v", cfg.Linters.Enable)
	}

	if sliceHas(cfg.Linters.Disable, "errcheck") {
		t.Errorf("errcheck should be removed from disable; disable=%v", cfg.Linters.Disable)
	}

	if !sliceHas(cfg.Linters.Disable, "gofmt") {
		t.Errorf("gofmt should remain disabled (justified); disable=%v", cfg.Linters.Disable)
	}

	if !f.recorder().hasReEnable("errcheck") {
		t.Errorf("expected audit ActionReEnabled for errcheck; recorded=%v", f.recorder().actions)
	}
}

func TestEnforceDisableReasons_ToolLevelExempt(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{} // present but no justifications

	cfg := &types.Config{Linters: types.LintersConfig{
		Disable: []types.LinterName{"funcorder", "errcheck"},
	}}

	count := f.enforceDisableReasons(cfg)

	// funcorder is tool-level managed (forcibly disabled) → exempt. Only errcheck re-enabled.
	if count != 1 {
		t.Fatalf("expected 1 re-enable (errcheck only), got %d", count)
	}

	if !sliceHas(cfg.Linters.Disable, "funcorder") {
		t.Errorf("funcorder must stay disabled (tool-level); disable=%v", cfg.Linters.Disable)
	}

	if f.recorder().hasReEnable("funcorder") {
		t.Error("funcorder must never be recorded as re-enabled")
	}
}

func TestEnforceDisableReasons_NeverAutoEnableExempt(t *testing.T) {
	f := newEnforceFixer()
	// Sidecar is present but justifies nothing: anti-gaming enforcement would
	// re-enable every unjustified disable. NeverAutoEnable linters must be exempt
	// so the tool cannot be tricked into force-enabling them via a sidecar gap.
	f.pol = &policy.Policy{}

	cfg := &types.Config{Linters: types.LintersConfig{
		Disable: []types.LinterName{"exhaustruct", "errcheck"},
	}}

	count := f.enforceDisableReasons(cfg)

	// exhaustruct is tool-level managed (never-auto-enable) → exempt. Only errcheck re-enabled.
	if count != 1 {
		t.Fatalf("expected 1 re-enable (errcheck only), got %d", count)
	}

	if !sliceHas(cfg.Linters.Disable, "exhaustruct") {
		t.Errorf("exhaustruct must stay disabled (tool-level managed); disable=%v", cfg.Linters.Disable)
	}

	if sliceHas(cfg.Linters.Enable, "exhaustruct") {
		t.Errorf("exhaustruct must never be force-enabled by sidecar enforcement; enable=%v", cfg.Linters.Enable)
	}

	if f.recorder().hasReEnable("exhaustruct") {
		t.Error("exhaustruct must never be recorded as re-enabled")
	}
}

func TestEnforceDisableReasons_NeverEnableOverridesUnjustified(t *testing.T) {
	f := newEnforceFixer()
	// godoclint is in never-enable AND disabled without justification.
	// Without the never-enable check, anti-gaming enforcement would re-enable it.
	// never-enable must take priority: the linter stays disabled.
	f.pol = &policy.Policy{
		NeverEnable: map[types.LinterName]policy.DisableJustification{
			"godoclint": {Reason: "incompatible with templ", Category: policy.CategoryConvention},
		},
	}

	cfg := &types.Config{Linters: types.LintersConfig{
		Disable: []types.LinterName{"godoclint", "errcheck"},
	}}

	count := f.enforceDisableReasons(cfg)

	// Only errcheck is re-enabled; godoclint is protected by never-enable.
	if count != 1 {
		t.Fatalf("expected 1 re-enable (errcheck only), got %d", count)
	}

	if !sliceHas(cfg.Linters.Disable, "godoclint") {
		t.Errorf("godoclint must stay disabled (never-enable); disable=%v", cfg.Linters.Disable)
	}

	if sliceHas(cfg.Linters.Enable, "godoclint") {
		t.Errorf("godoclint must never be force-enabled (never-enable); enable=%v", cfg.Linters.Enable)
	}

	if !sliceHas(cfg.Linters.Enable, "errcheck") {
		t.Errorf("errcheck should be re-enabled (unjustified, not never-enable); enable=%v", cfg.Linters.Enable)
	}

	if f.recorder().hasReEnable("godoclint") {
		t.Error("godoclint must never be recorded as re-enabled (never-enable)")
	}
}

func TestEnforceDisableReasons_AllJustified(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{Disabled: map[types.LinterName]policy.DisableJustification{
		"gofmt":    {Reason: "prefer golines", Category: policy.CategoryConvention},
		"errcheck": {Reason: "false positives", Category: policy.CategoryFalsePositives},
	}}

	cfg := &types.Config{Linters: types.LintersConfig{
		Disable: []types.LinterName{"gofmt", "errcheck"},
	}}

	if count := f.enforceDisableReasons(cfg); count != 0 {
		t.Fatalf("expected 0 re-enables when all justified, got %d", count)
	}

	if len(cfg.Linters.Disable) != 2 {
		t.Errorf("disable list should be unchanged; got %v", cfg.Linters.Disable)
	}
}

func TestEnforceDisableReasons_EmptyDisable(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{}

	cfg := &types.Config{}

	if count := f.enforceDisableReasons(cfg); count != 0 {
		t.Fatalf("expected 0 re-enables with empty disable, got %d", count)
	}
}

func TestTryReEnableLinter(t *testing.T) {
	tests := []struct {
		name            string
		linter          string
		pol             *policy.Policy
		wantReEnable    bool
		wantInEnable    bool
		wantInDisable   bool
		wantReEnableRec bool
		wantNoAuditRecs bool
	}{
		{
			name:            "re-enables unjustified non-tool linter",
			linter:          "errcheck",
			pol:             &policy.Policy{},
			wantReEnable:    true,
			wantInEnable:    true,
			wantInDisable:   false,
			wantReEnableRec: true,
		},
		{
			name:            "keeps tool-level managed linter",
			linter:          "funcorder",
			pol:             &policy.Policy{},
			wantReEnable:    false,
			wantInEnable:    false,
			wantInDisable:   true,
			wantNoAuditRecs: true,
		},
		{
			name:         "keeps justified linter",
			linter:       "gofmt",
			pol: &policy.Policy{Disabled: map[types.LinterName]policy.DisableJustification{
				"gofmt": {Reason: "prefer golines", Category: policy.CategoryConvention},
			}},
			wantReEnable:  false,
			wantInEnable:  false,
			wantInDisable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEnforceFixer()
			f.pol = tt.pol

			enable := types.NewSet[types.LinterName]()
			disable := types.NewSet[types.LinterName](types.LinterName(tt.linter))

			got := f.tryReEnableLinter(types.LinterName(tt.linter), enable, disable)
			assertReEnableResult(t, got, tt.wantReEnable, enable, disable, tt.linter,
				tt.wantInEnable, tt.wantInDisable)
			assertReEnableAudit(t, f, tt.linter, tt.wantReEnableRec, tt.wantNoAuditRecs)
		})
	}
}

func assertReEnableResult(t *testing.T, got, wantReEnable bool, enable, disable types.Set[types.LinterName],
	linter string, wantInEnable, wantInDisable bool,
) {
	t.Helper()

	if got != wantReEnable {
		t.Fatalf("tryReEnableLinter() = %v, want %v", got, wantReEnable)
	}

	if enable.Contains(types.LinterName(linter)) != wantInEnable {
		t.Errorf("linter %s in enable: got %v, want %v", linter, wantInEnable, !wantInEnable)
	}

	if disable.Contains(types.LinterName(linter)) != wantInDisable {
		t.Errorf("linter %s in disable: got %v, want %v", linter, wantInDisable, !wantInDisable)
	}
}

func assertReEnableAudit(t *testing.T, f *Fixer, linter string, wantRec, wantNoRecs bool) {
	t.Helper()

	if wantNoRecs && len(f.recorder().actions) != 0 {
		t.Errorf("no audit record expected; got %v", f.recorder().actions)
	}

	if wantRec && !f.recorder().hasReEnable(linter) {
		t.Errorf("expected audit record for re-enabled %s", linter)
	}
}

func TestTryReEnableLinter_NeverEnable(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{
		NeverEnable: map[types.LinterName]policy.DisableJustification{
			"godoclint": {Reason: "incompatible with templ", Category: policy.CategoryConvention},
		},
	}

	enable := types.NewSet[types.LinterName]()
	disable := types.NewSet[types.LinterName]("godoclint")

	if f.tryReEnableLinter("godoclint", enable, disable) {
		t.Fatal("expected false for never-enable linter")
	}

	if enable.Contains("godoclint") {
		t.Error("never-enable linter must not be added to enable")
	}

	if !disable.Contains("godoclint") {
		t.Error("never-enable linter must remain in disable")
	}

	if len(f.recorder().actions) != 0 {
		t.Errorf("no audit record expected; got %v", f.recorder().actions)
	}
}

func sliceHas(slice []types.LinterName, want types.LinterName) bool {
	return slices.Contains(slice, want)
}

func TestEnableRecommendedLinters_NeverEnableSidecar(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{
		NeverEnable: map[types.LinterName]policy.DisableJustification{
			"godoclint": {Reason: "incompatible with templ", Category: policy.CategoryConvention},
		},
	}

	linterSet := types.NewSet[types.LinterName]()
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "godoclint", Priority: types.LinterPriorityCritical, Reason: "doc linting"},
			{Name: "errcheck", Priority: types.LinterPriorityCritical, Reason: "error checking"},
		},
	}

	count := f.enableRecommendedLinters(linterSet, nil, analysis, types.LinterPriorityCritical, false)

	if count != 1 {
		t.Fatalf("expected 1 linter enabled (errcheck only), got %d", count)
	}

	if linterSet.Contains("godoclint") {
		t.Error("godoclint must NOT be enabled (listed in neverEnable sidecar)")
	}

	if !linterSet.Contains("errcheck") {
		t.Error("errcheck should be enabled (not in neverEnable)")
	}
}

func TestEnableRecommendedLinters_CycleDetection(t *testing.T) {
	recorder := &enforceRecorder{
		previouslyEnabled: map[string]bool{"godoclint": true},
	}
	f := &Fixer{
		logger: NewTestLogger(),
		ledger: recorder,
		reader: recorder,
	}

	linterSet := types.NewSet[types.LinterName]()
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "godoclint", Priority: types.LinterPriorityCritical, Reason: "doc linting"},
			{Name: "errcheck", Priority: types.LinterPriorityCritical, Reason: "error checking"},
		},
	}

	count := f.enableRecommendedLinters(linterSet, nil, analysis, types.LinterPriorityCritical, false)

	if count != 1 {
		t.Fatalf("expected 1 linter enabled (errcheck only), got %d", count)
	}

	if linterSet.Contains("godoclint") {
		t.Error("godoclint must NOT be re-enabled (regression loop detected)")
	}

	if !linterSet.Contains("errcheck") {
		t.Error("errcheck should be enabled (no cycle)")
	}

	if !recorder.hasSuppressedReEnable("godoclint") {
		t.Error("expected ActionSuppressedReEnable audit record for godoclint")
	}
}

func TestEnableRecommendedLinters_CycleDetectionDryRun(t *testing.T) {
	recorder := &enforceRecorder{
		previouslyEnabled: map[string]bool{"godoclint": true},
	}
	f := &Fixer{
		logger: NewTestLogger(),
		ledger: recorder,
		reader: recorder,
	}

	linterSet := types.NewSet[types.LinterName]()
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "godoclint", Priority: types.LinterPriorityCritical, Reason: "doc linting"},
		},
	}

	count := f.enableRecommendedLinters(linterSet, nil, analysis, types.LinterPriorityCritical, true)

	if count != 0 {
		t.Fatalf("expected 0 linters enabled in dry-run (godoclint suppressed), got %d", count)
	}

	if recorder.hasSuppressedReEnable("godoclint") {
		t.Error("dry-run must NOT record ActionSuppressedReEnable (no actual change)")
	}
}

func TestEnableRecommendedLinters_NoReaderNoCycleDetection(t *testing.T) {
	f := newEnforceFixer()

	linterSet := types.NewSet[types.LinterName]()
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "godoclint", Priority: types.LinterPriorityCritical, Reason: "doc linting"},
		},
	}

	count := f.enableRecommendedLinters(linterSet, nil, analysis, types.LinterPriorityCritical, false)

	if count != 1 {
		t.Fatalf("expected 1 linter enabled (no reader → no cycle detection), got %d", count)
	}

	if !linterSet.Contains("godoclint") {
		t.Error("godoclint should be enabled when no ledger reader is available")
	}
}

func TestEnableRecommendedLinters_DisabledNotAffectedByCycle(t *testing.T) {
	recorder := &enforceRecorder{
		previouslyEnabled: map[string]bool{"errcheck": true},
	}
	f := &Fixer{
		logger: NewTestLogger(),
		ledger: recorder,
		reader: recorder,
	}

	linterSet := types.NewSet[types.LinterName]()
	analysis := &types.ConfigAnalysis{
		LinterRecommendations: []types.LinterRecommendation{
			{Name: "errcheck", Priority: types.LinterPriorityCritical, Reason: "error checking"},
		},
	}

	disabled := []types.LinterName{"errcheck"}
	count := f.enableRecommendedLinters(linterSet, disabled, analysis, types.LinterPriorityCritical, false)

	if count != 0 {
		t.Fatalf("expected 0 (errcheck in disable list, skipped before cycle check), got %d", count)
	}

	if recorder.hasSuppressedReEnable("errcheck") {
		t.Error("disabled linters must not trigger cycle detection (they are already in disable)")
	}
}
