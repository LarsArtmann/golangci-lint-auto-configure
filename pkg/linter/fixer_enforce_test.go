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
	actions []recordedEnforceAction
}

func (r *enforceRecorder) Record(action audit.Action, linter, reason string) {
	r.actions = append(r.actions, recordedEnforceAction{action: action, linter: linter, reason: reason})
}

func (r *enforceRecorder) hasReEnable(linter string) bool {
	for _, a := range r.actions {
		if a.action == audit.ActionReEnabled && a.linter == linter {
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

func TestIsToolLevelDisabled(t *testing.T) {
	tests := []struct {
		name   string
		linter types.LinterName
		want   bool
	}{
		{"funcorder is tool-level disabled", "funcorder", true},
		{"noinlineerr is tool-level disabled", "noinlineerr", true},
		{"depguard is tool-level disabled", "depguard", true},
		{"errcheck is not tool-level disabled", "errcheck", false},
		{"gofmt is not tool-level disabled", "gofmt", false},
		{"unknown linter is not tool-level disabled", "does-not-exist", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isToolLevelDisabled(tt.linter); got != tt.want {
				t.Errorf("isToolLevelDisabled(%q) = %v, want %v", tt.linter, got, tt.want)
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
	f.pol = &policy.Policy{Disabled: map[string]policy.DisableJustification{
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

	// funcorder is tool-level disabled → exempt. Only errcheck re-enabled.
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

func TestEnforceDisableReasons_AllJustified(t *testing.T) {
	f := newEnforceFixer()
	f.pol = &policy.Policy{Disabled: map[string]policy.DisableJustification{
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
	t.Run("re-enables unjustified non-tool linter", func(t *testing.T) {
		f := newEnforceFixer()
		f.pol = &policy.Policy{} // no justifications

		enable := types.NewSet[types.LinterName]()
		disable := types.NewSet("errcheck")

		if !f.tryReEnableLinter("errcheck", enable, disable) {
			t.Fatal("expected tryReEnableLinter to re-enable unjustified linter")
		}

		if !enable.Contains("errcheck") {
			t.Error("errcheck should be moved into the enable set")
		}

		if disable.Contains("errcheck") {
			t.Error("errcheck should be removed from the disable set")
		}

		if !f.recorder().hasReEnable("errcheck") {
			t.Error("expected audit record for re-enabled errcheck")
		}
	})

	t.Run("keeps tool-level disabled linter", func(t *testing.T) {
		f := newEnforceFixer()
		f.pol = &policy.Policy{}

		enable := types.NewSet[types.LinterName]()
		disable := types.NewSet("funcorder")

		if f.tryReEnableLinter("funcorder", enable, disable) {
			t.Fatal("expected false for tool-level disabled linter")
		}

		if enable.Contains("funcorder") {
			t.Error("funcorder must not be added to enable")
		}

		if !disable.Contains("funcorder") {
			t.Error("funcorder must remain in disable")
		}

		if len(f.recorder().actions) != 0 {
			t.Errorf("no audit record expected; got %v", f.recorder().actions)
		}
	})

	t.Run("keeps justified linter", func(t *testing.T) {
		f := newEnforceFixer()
		f.pol = &policy.Policy{Disabled: map[string]policy.DisableJustification{
			"gofmt": {Reason: "prefer golines", Category: policy.CategoryConvention},
		}}

		enable := types.NewSet[types.LinterName]()
		disable := types.NewSet("gofmt")

		if f.tryReEnableLinter("gofmt", enable, disable) {
			t.Fatal("expected false for justified linter")
		}

		if enable.Contains("gofmt") {
			t.Error("justified linter must not be added to enable")
		}

		if !disable.Contains("gofmt") {
			t.Error("justified linter must remain in disable")
		}
	})
}

func sliceHas(slice []types.LinterName, want types.LinterName) bool {
	return slices.Contains(slice, want)
}
