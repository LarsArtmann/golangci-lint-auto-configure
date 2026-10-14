package linter

import (
	"slices"
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ApplyPlan materializes a plan onto a CLONE of the observed config. Pure:
// the observed input is never mutated, and applying the empty plan returns a
// clone equal to the input (idempotency by construction).
func ApplyPlan(observed *types.Config, changes []Change) *types.Config {
	merged := observed.Clone()

	for _, change := range changes {
		applyChange(merged, change)
	}

	sortEnableLists(merged)

	return merged
}

func applyChange(cfg *types.Config, change Change) {
	applyLinterChanges(cfg, change)
	applyFormatterChanges(cfg, change)
	applyRunChanges(cfg, change)
	applyIssueChanges(cfg, change)
}

func applyLinterChanges(cfg *types.Config, change Change) {
	switch change.Path {
	case PathLintersEnable:
		switch change.Kind {
		case ChangeAdd:
			if !containsLinterName(toStrings(cfg.Linters.Enable), change.Element) {
				cfg.Linters.Enable = append(cfg.Linters.Enable, types.LinterName(change.Element))
			}
		case ChangeRemove:
			cfg.Linters.Enable = removeLinterName(cfg.Linters.Enable, types.LinterName(change.Element))
		}
	case PathLintersDisable:
		if change.Kind == ChangeRemove {
			cfg.Linters.Disable = removeLinterName(cfg.Linters.Disable, types.LinterName(change.Element))
		}
	case PathLintersSettings:
		applySettingsChange(&cfg.Linters.Settings, change)
	case "linters.exclusions.paths":
		applyStringList(&cfg.Linters.Exclusions.Paths, change)
	case PathLintersExclusionRules:
		applyExclusionRule(&cfg.Linters.Exclusions.Rules, change)
	case "linters.exclusions.generated":
		if change.Kind == ChangeAdd {
			cfg.Linters.Exclusions.Generated = toString(change.To)
		}
	}
}

func applyFormatterChanges(cfg *types.Config, change Change) {
	switch change.Path {
	case PathFormattersEnable:
		switch change.Kind {
		case ChangeAdd:
			if !containsLinterName(toStrings(cfg.Linters.Enable), change.Element) &&
				!containsFormatterName(cfg.Formatters.Enable, types.FormatterName(change.Element)) {
				cfg.Formatters.Enable = append(cfg.Formatters.Enable, types.FormatterName(change.Element))
			}
		case ChangeRemove:
			cfg.Formatters.Enable = removeFormatterEntry(cfg.Formatters.Enable, types.FormatterName(change.Element))
		}
	case "formatters.settings":
		applySettingsChange(&cfg.Formatters.Settings, change)
	case "formatters.exclusions.paths":
		applyStringList(&cfg.Formatters.Exclusions.Paths, change)
	case "formatters.exclusions.generated":
		if change.Kind == ChangeAdd {
			cfg.Formatters.Exclusions.Generated = toString(change.To)
		}
	}
}

func applyRunChanges(cfg *types.Config, change Change) {
	switch change.Path {
	case "run.go":
		if change.Kind == ChangeUpdate {
			cfg.Run.Go = toString(change.To)
		}
	case PathRunTimeout:
		if change.Kind == ChangeAdd {
			cfg.Run.Timeout = toString(change.To)
		}
	case "run.allow-parallel-runners":
		if change.Kind == ChangeUpdate {
			cfg.Run.AllowParallelRunners = true
		}
	case "run.allow-serial-runners":
		if change.Kind == ChangeUpdate {
			cfg.Run.AllowSerialRunners = true
		}
	case "run.issues-exit-code":
		if change.Kind == ChangeAdd {
			cfg.Run.IssuesExitCode = toInt(change.To)
		}
	case "run.build-tags":
		applyStringList(&cfg.Run.BuildTags, change)
	}
}

func applyIssueChanges(cfg *types.Config, change Change) {
	switch change.Path {
	case "issues.max-issues-per-linter":
		if change.Kind == ChangeAdd {
			cfg.Issues.MaxIssuesPerLinter = toInt(change.To)
		}
	case "issues.max-same-issues":
		if change.Kind == ChangeAdd {
			cfg.Issues.MaxSameIssues = toInt(change.To)
		}
	}
}

func sortEnableLists(cfg *types.Config) {
	slices.Sort(cfg.Linters.Enable)

	// Formatters render in the canonical FormatterOrder (unknown names
	// follow, sorted) — the legacy ToOrderedSlice contract. Exclusion paths
	// are NOT sorted here: they keep the user's order unless the merge added
	// entries (applyStringList sorts on add, matching the legacy merge).
	ordered := make([]types.FormatterName, 0, len(cfg.Formatters.Enable))

	for _, name := range constants.FormatterOrder {
		if slices.Contains(cfg.Formatters.Enable, name) {
			ordered = append(ordered, name)
		}
	}

	remaining := make([]types.FormatterName, 0)

	for _, name := range cfg.Formatters.Enable {
		if !slices.Contains(constants.FormatterOrder, name) {
			remaining = append(remaining, name)
		}
	}

	slices.Sort(remaining)

	cfg.Formatters.Enable = append(ordered, remaining...)

	if cfg.Run.BuildTags != nil {
		sort.Strings(cfg.Run.BuildTags)
	}
}

func toString(v any) string {
	if s, isString := v.(string); isString {
		return s
	}

	return ""
}

func toInt(v any) int {
	if i, isInt := v.(int); isInt {
		return i
	}

	return 0
}

func applyExclusionRule(rules *[]types.ExclusionRuleConfig, change Change) {
	switch change.Kind {
	case ChangeAdd:
		if rule, ok := change.To.(types.ExclusionRuleConfig); ok {
			*rules = append(*rules, rule)
		}
	case ChangeUpdate:
		if union, ok := change.To.([]string); ok {
			// Legacy-exact: the union lands on the LAST rule with the key
			// (later rules shadow earlier ones).
			for i := len(*rules) - 1; i >= 0; i-- {
				if (*rules)[i].RuleKey() == change.Element {
					(*rules)[i].Linters = union

					break
				}
			}
		}
	case ChangeRemove:
		for i := range *rules {
			if (*rules)[i].RuleKey() == change.Element {
				*rules = append((*rules)[:i], (*rules)[i+1:]...)

				return
			}
		}
	}
}

func containsFormatterName(names []types.FormatterName, want types.FormatterName) bool {
	return slices.Contains(names, want)
}

func removeFormatterEntry(names []types.FormatterName, name types.FormatterName) []types.FormatterName {
	out := names[:0:0]

	for _, candidate := range names {
		if candidate != name {
			out = append(out, candidate)
		}
	}

	return out
}

func removeLinterName(names []types.LinterName, name types.LinterName) []types.LinterName {
	out := names[:0:0]

	for _, candidate := range names {
		if candidate != name {
			out = append(out, candidate)
		}
	}

	return out
}

func applySettingsChange(section *map[string]any, change Change) {
	if *section == nil {
		*section = map[string]any{}
	}

	switch change.Kind {
	case ChangeAdd:
		(*section)[change.Element] = change.To
	case ChangeUpdate:
		(*section)[change.Element] = change.To
	case ChangeRemove:
		delete(*section, change.Element)
	}
}

func applyStringList(target *[]string, change Change) {
	if change.Kind != ChangeAdd {
		return
	}

	if slices.Contains(*target, change.Element) {
		return
	}

	// Legacy merge semantics (gogenfilter.MergeExclusionPaths): the merged
	// list is sorted whenever anything was added; untouched lists keep the
	// user's order.
	*target = append(*target, change.Element)
	slices.Sort(*target)
}
