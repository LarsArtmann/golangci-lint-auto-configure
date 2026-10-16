package linter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// fakeLintersJSON is the wire shape `golangci-lint linters --json` emits:
// capitalized wrapper keys, lowercase field keys.
const fakeLintersJSON = `{
  "Enabled": [
    {"name": "errcheck", "description": "default", "fast": false, "deprecated": false},
    {"name": "govet", "description": "default", "fast": true, "deprecated": false}
  ],
  "Disabled": [
    {"name": "gosec", "description": "known", "fast": false, "deprecated": false},
    {"name": "gohumanize", "description": "plugin", "fast": false, "deprecated": false}
  ]
}`

// writeProbeBinary materializes an executable stand-in for golangci-lint that
// emits fixture JSON and appends one line per invocation to countFile, so
// specs can prove the probe is cached.
func writeProbeBinary(t *testing.T, dir, countFile string) string {
	t.Helper()

	script := "#!/bin/sh\n"
	if countFile != "" {
		script += "echo invoked >> " + countFile + "\n"
	}

	script += "cat <<'JSONEOF'\n" + fakeLintersJSON + "\nJSONEOF\n"

	path := filepath.Join(dir, "golangci-lint")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	return path
}

func TestProbeKnownLinters(t *testing.T) {
	dir := t.TempDir()
	binary := writeProbeBinary(t, dir, "")

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = binary

	known, err := analyzer.ProbeKnownLinters(context.Background())
	if err != nil {
		t.Fatalf("ProbeKnownLinters failed: %v", err)
	}

	for _, name := range []string{"errcheck", "govet", "gosec", "gohumanize"} {
		if !known[types.LinterName(name)] {
			t.Errorf("ProbeKnownLinters() missing %s (enabled and disabled must both count as known)", name)
		}
	}

	defaults := analyzer.DefaultEnabledLinters(context.Background())
	if len(defaults) != 2 || defaults[0] != "errcheck" || defaults[1] != "govet" {
		t.Errorf("DefaultEnabledLinters() = %v, want [errcheck govet] (the probe's Enabled list)", defaults)
	}
}

func TestProbeKnownLintersCachesPerAnalyzer(t *testing.T) {
	dir := t.TempDir()
	countFile := filepath.Join(dir, "invocations")
	binary := writeProbeBinary(t, dir, countFile)

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = binary

	ctx := context.Background()

	if _, err := analyzer.ProbeKnownLinters(ctx); err != nil {
		t.Fatalf("first probe failed: %v", err)
	}

	if _, err := analyzer.ProbeKnownLinters(ctx); err != nil {
		t.Fatalf("second probe failed: %v", err)
	}

	counted, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("read invocation count: %v", err)
	}

	if invocations := strings.Count(string(counted), "invoked"); invocations != 1 {
		t.Errorf("probe ran %d times, want 1 (the result must be cached)", invocations)
	}
}

func TestProbeKnownLintersMissingBinaryFailsOpen(t *testing.T) {
	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = filepath.Join(t.TempDir(), "does-not-exist")

	known, err := analyzer.ProbeKnownLinters(context.Background())
	if err == nil {
		t.Fatal("ProbeKnownLinters() error = nil, want failure for a missing binary")
	}

	if known != nil {
		t.Errorf("ProbeKnownLinters() = %v, want nil on failure (callers must fail open)", known)
	}
}

func TestProbeKnownLintersGarbageOutputFailsOpen(t *testing.T) {
	dir := t.TempDir()

	script := "#!/bin/sh\necho 'not json'\n"
	path := filepath.Join(dir, "golangci-lint")

	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = path

	known, err := analyzer.ProbeKnownLinters(context.Background())
	if err == nil {
		t.Fatal("ProbeKnownLinters() error = nil, want failure for unparseable output")
	}

	if known != nil {
		t.Errorf("ProbeKnownLinters() = %v, want nil on failure (callers must fail open)", known)
	}
}
