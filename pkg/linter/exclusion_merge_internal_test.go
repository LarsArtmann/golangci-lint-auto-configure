package linter

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// newTestUpdater builds a configUpdater with a capturing logger.
func newTestUpdater(showMergedRules bool) (*configUpdater, *bytes.Buffer) {
	buf := &bytes.Buffer{}

	level := log.ErrorLevel
	if showMergedRules {
		level = log.InfoLevel
	}

	logger := log.NewWithOptions(buf, log.Options{Level: level})

	updater := newConfigUpdater(logger, func(context.Context) string { return "" }, "")
	updater.setShowMergedRules(showMergedRules)

	return updater, buf
}

// partialExclusionConfig builds a config whose exclusion rules carry only a
// strict subset of one default rule's linters, forcing a RuleKey merge.
func partialExclusionConfig(t *testing.T) *types.Config {
	t.Helper()

	if len(constants.DefaultExclusionRules) == 0 {
		t.Fatal("no default exclusion rules to merge against")
	}

	defaultRule := constants.DefaultExclusionRules[0]
	if len(defaultRule.Linters) < 2 {
		t.Fatalf("first default rule has %d linters; need >= 2 for a partial merge", len(defaultRule.Linters))
	}

	cfg := &types.Config{}
	cfg.Linters.Exclusions.Rules = []types.ExclusionRuleConfig{
		{
			Path:    defaultRule.Path,
			Text:    defaultRule.Text,
			Source:  defaultRule.Source,
			Linters: []string{defaultRule.Linters[0]},
		},
	}

	return cfg
}

func TestUpdateExclusionRulesCollectsMerges(t *testing.T) {
	updater, buf := newTestUpdater(true)
	cfg := partialExclusionConfig(t)

	defaultRule := constants.DefaultExclusionRules[0]

	changed := updater.updateExclusionRules(cfg)
	if changed == 0 {
		t.Fatal("expected at least one change from the partial exclusion rule")
	}

	merges := updater.drainExclusionMerges()
	if len(merges) != 1 {
		t.Fatalf("expected exactly 1 collected merge, got %d", len(merges))
	}

	merge := merges[0]
	if merge.RuleKey != defaultRule.RuleKey() {
		t.Fatalf("merge RuleKey = %q, want %q", merge.RuleKey, defaultRule.RuleKey())
	}

	wantAdded := defaultRule.Linters[1:]
	if strings.Join(merge.AddedLinters, ",") != strings.Join(wantAdded, ",") {
		t.Fatalf("added linters = %v, want %v", merge.AddedLinters, wantAdded)
	}

	for _, linter := range defaultRule.Linters {
		found := false

		for _, rule := range cfg.Linters.Exclusions.Rules {
			for _, existing := range rule.Linters {
				if existing == linter {
					found = true
				}
			}
		}

		if !found {
			t.Fatalf("linter %q missing from merged exclusion rules", linter)
		}
	}

	if !strings.Contains(buf.String(), defaultRule.RuleKey()) {
		t.Fatalf("--show-merged-rules output missing RuleKey; got: %s", buf.String())
	}

	if again := updater.drainExclusionMerges(); len(again) != 0 {
		t.Fatalf("drain must clear the merge list, got %d entries", len(again))
	}
}

func TestUpdateExclusionRulesSilentWithoutFlag(t *testing.T) {
	updater, buf := newTestUpdater(false)
	cfg := partialExclusionConfig(t)

	if updater.updateExclusionRules(cfg) == 0 {
		t.Fatal("expected merge to happen")
	}

	if len(updater.drainExclusionMerges()) != 1 {
		t.Fatal("merges are collected even when the flag is off")
	}

	if buf.String() != "" {
		t.Fatalf("no merge logging expected without --show-merged-rules; got: %s", buf.String())
	}
}

func TestDrainExclusionMergesEmpty(t *testing.T) {
	updater, _ := newTestUpdater(true)

	if merges := updater.drainExclusionMerges(); len(merges) != 0 {
		t.Fatalf("fresh updater must have no merges, got %d", len(merges))
	}
}
