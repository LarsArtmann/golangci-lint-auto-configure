package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
)

func writeSidecar(t *testing.T, dir, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, policy.SidecarFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
}

func writePolicyConfig(t *testing.T, dir string) string {
	t.Helper()

	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, []byte("version: \"2\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return configPath
}

func TestLoadPolicy_NoSidecar(t *testing.T) {
	f := newQuietFixer()
	dir := t.TempDir()
	configPath := writePolicyConfig(t, dir)

	f.loadPolicy(configPath)

	if f.pol != nil {
		t.Errorf("expected nil policy when no sidecar exists, got %+v", f.pol)
	}
}

func TestLoadPolicy_WithSidecar(t *testing.T) {
	f := newQuietFixer()
	dir := t.TempDir()
	configPath := writePolicyConfig(t, dir)

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
	f := newQuietFixer()
	dir := t.TempDir()
	configPath := writePolicyConfig(t, dir)

	writeSidecar(t, dir, "}{not valid yaml ][\n")

	f.loadPolicy(configPath)

	if f.pol != nil {
		t.Errorf("expected nil policy on malformed sidecar (error warned, not fatal), got %+v", f.pol)
	}
}
