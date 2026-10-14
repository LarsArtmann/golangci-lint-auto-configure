package linter

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ParityCorpusDir is the on-disk home of the parity corpus fixtures.
var ParityCorpusDir = filepath.Join("testdata", "parity-corpus")

// ParityManifest carries the environment assumptions and fixer flags for one
// corpus fixture. Every field is part of the parity contract: changing one is
// a behavior change and must regenerate the fixture's goldens.
// Field names are the JSON keys (PascalCase, report-type family).
type ParityManifest struct {
	Name                string
	Description         string
	Source              string
	Priority            string
	DryRun              bool
	ForceSettings       bool
	Pragmatic           bool
	LocalGoVersion      string
	RequiresToolAtLeast string
}

// ParityFixture is one corpus entry: inputs plus the parsed manifest.
type ParityFixture struct {
	Name         string
	Dir          string
	Manifest     ParityManifest
	ConfigInput  []byte
	SidecarInput []byte
	HasSidecar   bool
}

// parityRunContext is fixed so ledger bytes stay deterministic across machines.
var parityRunContext = audit.RunContext{
	RunID:    "parity-corpus-run",
	RepoHash: "paritycorpus",
	RepoPath: "/parity/fixture/repo",
}

// LoadParityCorpus reads every fixture directory (input files + manifest).
// Directories without a manifest.json are ignored so scratch work is safe.
func LoadParityCorpus(t *testing.T) []ParityFixture {
	t.Helper()

	entries, err := os.ReadDir(ParityCorpusDir)
	if err != nil {
		t.Fatalf("read parity corpus dir: %v", err)
	}

	var fixtures []ParityFixture

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(ParityCorpusDir, entry.Name())

		manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		if err != nil {
			continue
		}

		var manifest ParityManifest
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			t.Fatalf("fixture %s: parse manifest.json: %v", entry.Name(), err)
		}

		if manifest.Name == "" {
			manifest.Name = entry.Name()
		}

		configInput, err := os.ReadFile(filepath.Join(dir, "input.golangci.yml"))
		if err != nil {
			t.Fatalf("fixture %s: read input.golangci.yml: %v", entry.Name(), err)
		}

		fixture := ParityFixture{
			Name:        manifest.Name,
			Dir:         dir,
			Manifest:    manifest,
			ConfigInput: configInput,
		}

		sidecar, sidecarErr := os.ReadFile(filepath.Join(dir, "input.golangci-lint-auto-configure.yml"))
		if sidecarErr == nil {
			fixture.SidecarInput = sidecar
			fixture.HasSidecar = true
		}

		fixtures = append(fixtures, fixture)
	}

	if len(fixtures) == 0 {
		t.Fatal("parity corpus is empty")
	}

	return fixtures
}

// ParityOutcome is everything one fixture run produces that the goldens pin:
// saved config bytes, scrubbed audit ledger bytes, and the result summary.
type ParityOutcome struct {
	ConfigBytes []byte
	AuditJSONL  []byte
	Result      *types.MigrationResult
}

// RunParityFixture materializes the fixture in a fresh temp directory, runs
// the full FixConfig pipeline with a real analyzer and a real (temp) audit
// ledger, and returns the byte-level outcome.
func RunParityFixture(t *testing.T, fixture ParityFixture) ParityOutcome {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	if err := os.WriteFile(configPath, fixture.ConfigInput, 0o644); err != nil {
		t.Fatalf("write fixture config: %v", err)
	}

	if fixture.HasSidecar {
		sidecarPath := filepath.Join(dir, ".golangci-lint-auto-configure.yml")
		if err := os.WriteFile(sidecarPath, fixture.SidecarInput, 0o644); err != nil {
			t.Fatalf("write fixture sidecar: %v", err)
		}
	}

	logger := NewTestLogger()
	analyzer := NewAnalyzer(logger)
	analyzer.SetPragmatic(fixture.Manifest.Pragmatic)

	priority, err := types.ParseLinterPriority(fixture.Manifest.Priority)
	if err != nil {
		t.Fatalf("fixture %s: %v", fixture.Name, err)
	}

	ledgerPath := filepath.Join(dir, "audit.jsonl")
	ledger := audit.NewLedger(logger, parityRunContext, ledgerPath)

	fixer := NewFixer(logger, analyzer, config.NewLoader(logger))
	fixer.SetLedger(ledger)
	fixer.SetGoVersionProvider(func(context.Context) string { return fixture.Manifest.LocalGoVersion })
	fixer.SetForceSettings(fixture.Manifest.ForceSettings)

	result, err := fixer.FixConfig(context.Background(), configPath, priority, fixture.Manifest.DryRun)
	if err != nil {
		t.Fatalf("fixture %s: FixConfig failed: %v", fixture.Name, err)
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("fixture %s: read saved config: %v", fixture.Name, err)
	}

	return ParityOutcome{
		ConfigBytes: configBytes,
		AuditJSONL:  scrubAuditLedger(t, fixture.Name, ledgerPath),
		Result:      result,
	}
}

// scrubAuditLedger normalizes the volatile per-run fields of the audit JSONL
// (timestamp, run_id, repo_hash, repo_path, git_head) to fixed placeholders so
// the action sequence stays byte-comparable across runs and machines.
func scrubAuditLedger(t *testing.T, fixtureName, ledgerPath string) []byte {
	t.Helper()

	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		t.Fatalf("fixture %s: read audit ledger: %v", fixtureName, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var scrubbed strings.Builder

	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		var entry audit.Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("fixture %s: parse audit entry: %v", fixtureName, err)
		}

		entry.Timestamp = time.Time{}
		entry.RunID = "parity-corpus-run"
		entry.RepoHash = "paritycorpus"
		entry.RepoPath = "/parity/fixture/repo"
		entry.GitHead = ""

		lineBytes, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("fixture %s: marshal scrubbed audit entry: %v", fixtureName, err)
		}

		scrubbed.Write(lineBytes)
		scrubbed.WriteByte('\n')
	}

	return []byte(scrubbed.String())
}
