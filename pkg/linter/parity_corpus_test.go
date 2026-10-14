package linter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"golang.org/x/mod/semver"
)

// TestParityCorpus is the byte-level regression gate for the configure engine
// (plan phase P0). Each fixture's saved config bytes, audit entries, and
// result summary are compared against the committed goldens. Set
// UPDATE_PARITY=1 to regenerate goldens, then review the diff before
// committing — the goldens are the contract later refactor phases are
// judged against.
//
// The run refuses when the environment's golangci-lint is older than a
// fixture's requiresToolAtLeast floor: output would legitimately differ, and
// silently regenerating goldens from a different tool version would corrupt
// the contract.
func TestParityCorpus(t *testing.T) {
	update := os.Getenv("UPDATE_PARITY") == "1"

	for _, fixture := range LoadParityCorpus(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			guardToolVersion(t, fixture)

			outcome := RunParityFixture(t, fixture)

			goldenDir := filepath.Join(fixture.Dir, "golden")

			assertGoldenBytes(t, update, goldenDir, "config.yml", outcome.ConfigBytes)
			assertGoldenBytes(t, update, goldenDir, "audit.jsonl", outcome.AuditJSONL)
			assertGoldenResult(t, update, goldenDir, outcome.Result)
		})
	}
}

// TestParityCorpusIdempotent runs every fixture a second time against the
// first run's output: a converged config must re-converge to byte-identical
// output. This is the property the declarative engine makes structural; until
// then it pins the imperative engine's existing guarantee.
func TestParityCorpusIdempotent(t *testing.T) {
	for _, fixture := range LoadParityCorpus(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			guardToolVersion(t, fixture)

			first := RunParityFixture(t, fixture)

			if first.Result.FixesApplied == 0 {
				t.Skipf("fixture %s: no changes on first run, idempotency trivially holds", fixture.Name)
			}

			second := RunParityFixture(t, fixture)

			if string(first.ConfigBytes) != string(second.ConfigBytes) {
				t.Errorf("fixture %s not idempotent: second run changed config bytes\n--- first ---\n%s\n--- second ---\n%s",
					fixture.Name, first.ConfigBytes, second.ConfigBytes)
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
		t.Fatalf("fixture %s requires golangci-lint >= %s but no version could be detected (is golangci-lint installed?)",
			fixture.Name, fixture.Manifest.RequiresToolAtLeast)
	}

	if semver.Compare(detected, fixture.Manifest.RequiresToolAtLeast) < 0 {
		t.Fatalf("fixture %s requires golangci-lint >= %s, found %s — run on a compatible binary or consciously regenerate goldens",
			fixture.Name, fixture.Manifest.RequiresToolAtLeast, detected)
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

func assertGoldenBytes(t *testing.T, update bool, goldenDir, name string, actual []byte) {
	t.Helper()

	goldenPath := filepath.Join(goldenDir, name)

	if update {
		if len(actual) == 0 && name == "audit.jsonl" {
			_ = os.Remove(goldenPath)

			return
		}

		if err := os.MkdirAll(goldenDir, 0o755); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}

		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", name, err)
		}

		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) && name == "audit.jsonl" && len(actual) == 0 {
			return
		}

		t.Fatalf("read golden %s (regenerate with UPDATE_PARITY=1): %v", name, err)
	}

	if string(want) != string(actual) {
		t.Errorf("golden mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", name, want, actual)
	}
}

func assertGoldenResult(t *testing.T, update bool, goldenDir string, result *types.MigrationResult) {
	t.Helper()

	if result == nil {
		t.Fatal("nil MigrationResult")

		return
	}

	payload, err := json.MarshalIndent(struct {
		FixesApplied int
		Message      string
		DryRun       bool
	}{
		FixesApplied: result.FixesApplied,
		Message:      result.Message,
		DryRun:       result.DryRun,
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}

	payload = append(payload, '\n')

	assertGoldenBytes(t, update, goldenDir, "result.json", payload)
}
