package linter

import (
	"context"
	"os"
	"path/filepath"
	"sort"
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

// inheritedConfigLintersJSON mimics a probe that inherited an AMBIENT
// config from a parent directory (observed 2026-11-01: /tmp/.golangci.yml
// enabling 108 of 115 linters leaked into the temp-cwd probe; the initial
// attribution to a v2.15.0 output change was falsified 2026-11-02): the
// Enabled list covers nearly the whole vocabulary, so it is not
// "default-enabled" but "everything".
const inheritedConfigLintersJSON = `{
  "Enabled": [
    {"name": "errcheck", "description": "d", "fast": false, "deprecated": false},
    {"name": "govet", "description": "d", "fast": true, "deprecated": false},
    {"name": "gosec", "description": "d", "fast": false, "deprecated": false},
    {"name": "bodyclose", "description": "d", "fast": false, "deprecated": false},
    {"name": "wsl_v5", "description": "d", "fast": false, "deprecated": false}
  ],
  "Disabled": [
    {"name": "gohumanize", "description": "plugin", "fast": false, "deprecated": false}
  ]
}`

func TestProbeDefaultEnabledImplausibleSetFailsOpen(t *testing.T) {
	dir := t.TempDir()

	script := "#!/bin/sh\ncat <<'JSONEOF'\n" + inheritedConfigLintersJSON + "\nJSONEOF\n"

	path := filepath.Join(dir, "golangci-lint")

	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = path

	known, err := analyzer.ProbeKnownLinters(context.Background())
	if err != nil {
		t.Fatalf("ProbeKnownLinters failed: %v", err)
	}

	if len(known) != 6 {
		t.Fatalf("vocabulary = %d names, want 6 (enabled and disabled both count)", len(known))
	}

	if defaults := analyzer.DefaultEnabledLinters(context.Background()); defaults != nil {
		t.Errorf("DefaultEnabledLinters() = %v, want nil: a default set covering most of "+
			"the vocabulary is inherited-config contamination (or a genuinely "+
			"list-everything binary), not the default set; "+
			"consuming it would silently suppress every recommendation", defaults)
	}
}

// defaultEraLintersJSON pins the TRUE default-enabled set. Verified against
// clean-parent probes on 2026-11-02: golangci-lint v2.14 AND v2.15 behave
// identically here (the earlier "v2.15 lists everything" reading was ambient
// /tmp/.golangci.yml contamination, not an output change), so this one
// fixture covers both binary eras. Enabled is the documented 5-linter
// default set; the disabled remainder keeps defaults ≤ half the vocabulary
// so the plausibility guard accepts it.
const defaultEraLintersJSON = `{
  "Enabled": [
    {"name": "errcheck", "description": "d", "fast": false, "deprecated": false},
    {"name": "govet", "description": "d", "fast": true, "deprecated": false},
    {"name": "ineffassign", "description": "d", "fast": true, "deprecated": false},
    {"name": "staticcheck", "description": "d", "fast": false, "deprecated": false},
    {"name": "unused", "description": "d", "fast": false, "deprecated": false}
  ],
  "Disabled": [
    {"name": "gosec", "description": "d", "fast": false, "deprecated": false},
    {"name": "bodyclose", "description": "d", "fast": false, "deprecated": false},
    {"name": "wsl_v5", "description": "d", "fast": false, "deprecated": false},
    {"name": "ginkgolinter", "description": "d", "fast": false, "deprecated": false},
    {"name": "nilnesserr", "description": "d", "fast": false, "deprecated": false},
    {"name": "usestdlibvars", "description": "d", "fast": true, "deprecated": false}
  ]
}`

// The plausibility guard must ACCEPT the genuine default set: v2.14 and
// v2.15 binaries both report these 5 enabled linters under a hermetic probe,
// and the degraded-analysis fallback consumes exactly this set.
func TestProbeDefaultEnabledSetMatchesBothEras(t *testing.T) {
	dir := t.TempDir()

	script := "#!/bin/sh\ncat <<'JSONEOF'\n" + defaultEraLintersJSON + "\nJSONEOF\n"

	path := filepath.Join(dir, "golangci-lint")

	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = path

	known, err := analyzer.ProbeKnownLinters(context.Background())
	if err != nil {
		t.Fatalf("ProbeKnownLinters failed: %v", err)
	}

	if len(known) != 11 {
		t.Fatalf("vocabulary = %d names, want 11 (enabled and disabled both count)", len(known))
	}

	defaults := analyzer.DefaultEnabledLinters(context.Background())
	if defaults == nil {
		t.Fatal("DefaultEnabledLinters() = nil, want the 5-linter default set: the " +
			"plausibility guard must not trip on a genuine default-era listing")
	}

	want := []string{"errcheck", "govet", "ineffassign", "staticcheck", "unused"}
	if len(defaults) != len(want) {
		t.Fatalf("default set = %v, want %v", defaults, want)
	}

	got := make([]string, 0, len(defaults))
	for _, name := range defaults {
		got = append(got, string(name))
	}

	sort.Strings(got)
	sort.Strings(want)

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("default set = %v, want %v", got, want)
		}
	}
}

// The probe must pin itself to a minimal explicit --config: a bare temp
// cwd is not hermetic because golangci-lint walks UP from the working
// directory to discover configs, so an ambient config in any parent
// (observed: /tmp/.golangci.yml) leaks into the listing.
func TestProbePinsMinimalConfig(t *testing.T) {
	dir := t.TempDir()

	argsPath := filepath.Join(dir, "args.txt")
	configPath := filepath.Join(dir, "captured-config.yml")

	script := "#!/bin/sh\n" +
		"echo \"$@\" > \"" + argsPath + "\"\n" +
		"while [ $# -gt 0 ]; do case \"$1\" in --config) shift; cp \"$1\" \"" + configPath + "\";; esac; shift; done\n" +
		"cat <<'JSONEOF'\n" + inheritedConfigLintersJSON + "\nJSONEOF\n"

	path := filepath.Join(dir, "golangci-lint")

	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = path

	if _, err := analyzer.ProbeKnownLinters(context.Background()); err != nil {
		t.Fatalf("ProbeKnownLinters failed: %v", err)
	}

	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("fake binary captured no args: %v", err)
	}

	if !strings.Contains(string(args), "--config") {
		t.Errorf("probe args = %q, want an explicit --config pinning the listing", args)
	}

	captured, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("fake binary captured no --config value: %v", err)
	}

	if got := strings.TrimSpace(string(captured)); got != `version: "2"` {
		t.Errorf("probe config = %q, want a minimal version-only config", got)
	}
}
