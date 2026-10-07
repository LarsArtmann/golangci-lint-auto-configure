package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"go.yaml.in/yaml/v3"
)

// TestFixConfigRecordsExclusionMergeLedger proves the e2e contract: a config
// carrying a partial default exclusion rule gets the merge recorded in the
// audit ledger as exclusion-rule-merged (non-dry runs only).
func TestFixConfigRecordsExclusionMergeLedger(t *testing.T) {
	dir := t.TempDir()

	defaultRule := constants.DefaultExclusionRules[0]

	cfg := &types.Config{}
	cfg.Version = "2"
	cfg.Linters.Exclusions.Rules = []types.ExclusionRuleConfig{
		{
			Path:    defaultRule.Path,
			Text:    defaultRule.Text,
			Source:  defaultRule.Source,
			Linters: []string{defaultRule.Linters[0]},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal fixture config: %v", err)
	}

	configPath := filepath.Join(dir, ".golangci.yml")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write fixture config: %v", err)
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})

	ledgerPath := filepath.Join(dir, "audit", "audit.jsonl")

	fixer := NewFixer(logger, NewAnalyzer(logger), config.NewLoader(logger))
	fixer.SetLedger(audit.NewLedger(logger, audit.RunContext{
		RunID:    "exclusion-merge-test",
		RepoHash: "exclusionmergetest",
		RepoPath: dir,
	}, ledgerPath))

	priority, err := types.ParseLinterPriority("optional")
	if err != nil {
		t.Fatalf("parse priority: %v", err)
	}

	if _, err := fixer.FixConfig(t.Context(), configPath, priority, false); err != nil {
		t.Fatalf("FixConfig: %v", err)
	}

	entries, err := audit.ReadAll(ledgerPath)
	if err != nil {
		t.Fatalf("read audit ledger: %v", err)
	}

	for _, entry := range entries {
		if entry.Action != audit.ActionExclusionRuleMerged {
			continue
		}

		if entry.Linter == defaultRule.RuleKey() && strings.Contains(entry.Reason, defaultRule.Linters[1]) {
			return // found the merge record with RuleKey + added linters
		}
	}

	t.Fatalf("no %s entry with RuleKey %s and added linters; got %d entries",
		audit.ActionExclusionRuleMerged, defaultRule.RuleKey(), len(entries))
}
