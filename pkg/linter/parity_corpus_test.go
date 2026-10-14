package linter

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/mod/semver"
)

// TestParityCorpusConverge is the byte-level regression gate for the configure
// engine. Each fixture's saved config bytes are compared against the committed
// goldens: the config bytes are the contract 160 consumer projects depend on.
// Result/audit summaries are internal shape and are NOT compared here (the
// audit mapping has its own specs).
//
// The run refuses when the environment's golangci-lint is older than a
// fixture's requiresToolAtLeast floor: output would legitimately differ, and
// silently regenerating goldens from a different tool version would corrupt
// the contract.
func TestParityCorpusConverge(t *testing.T) {
	for _, fixture := range LoadParityCorpus(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			guardToolVersion(t, fixture)

			configBytes := RunParityFixture(t, fixture)

			goldenPath := filepath.Join(fixture.Dir, "golden", "config.yml")

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden config (regenerate with UPDATE_PARITY=1): %v", err)
			}

			if string(want) != string(configBytes) {
				if dump := os.Getenv("CONVERGE_DUMP"); dump != "" {
					dumpDir := filepath.Join(os.TempDir(), "converge-dump", fixture.Name)
					_ = os.MkdirAll(dumpDir, 0o755)
					_ = os.WriteFile(filepath.Join(dumpDir, "want.yml"), want, 0o644)
					_ = os.WriteFile(filepath.Join(dumpDir, "got.yml"), configBytes, 0o644)
				}

				t.Errorf("converge engine drifted from golden for %s (%d want bytes, %d got bytes)",
					fixture.Name, len(want), len(configBytes))
			}
		})
	}
}

// guardToolVersion refuses to run the fixture when the environment's
// golangci-lint binary is older than the manifest's floor.
func guardToolVersion(t *testing.T, fixture ParityFixture) {
	t.Helper()

	if fixture.Manifest.RequiresToolAtLeast == "" {
		return
	}

	detected := detectCorpusToolVersion(t)

	if detected == "" {
		t.Fatalf(
			"fixture %s requires golangci-lint >= %s but no version could be detected (is golangci-lint installed?)",
			fixture.Name,
			fixture.Manifest.RequiresToolAtLeast,
		)
	}

	if semver.Compare(detected, fixture.Manifest.RequiresToolAtLeast) < 0 {
		t.Fatalf(
			"fixture %s requires golangci-lint >= %s, found %s — run on a compatible binary or consciously regenerate goldens",
			fixture.Name,
			fixture.Manifest.RequiresToolAtLeast,
			detected,
		)
	}
}

// detectCorpusToolVersion shells out to the real binary once per test binary.
var detectCorpusToolVersion = func(t *testing.T) string {
	t.Helper()

	logger := NewTestLogger()

	analyzer := NewAnalyzer(logger)
	if err := analyzer.FindBinary(t.Context()); err != nil {
		return ""
	}

	if err := analyzer.CheckVersion(t.Context()); err != nil {
		return ""
	}

	return analyzer.GetDetectedVersion()
}
