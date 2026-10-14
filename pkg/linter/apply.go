package linter

import (
	"slices"
	"sort"

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
	case "linters.enable":
		if change.Kind == ChangeAdd {
			cfg.Linters.Enable = append(cfg.Linters.Enable, types.LinterName(change.Element))
		}
	case "linters.disable":
		if change.Kind == ChangeRemove {
			cfg.Linters.Disable = removeLinterName(cfg.Linters.Disable, types.LinterName(change.Element))
		}
	case "linters.settings":
		applySettingsChange(&cfg.Linters.Settings, change)
	case "linters.exclusions.paths":
		applyStringList(&cfg.Linters.Exclusions.Paths, change)
	case "linters.exclusions.generated":
		if change.Kind == ChangeAdd {
			cfg.Linters.Exclusions.Generated = toString(change.To)
		}
	}
}

func applyFormatterChanges(cfg *types.Config, change Change) {
	switch change.Path {
	case "formatters.enable":
		if change.Kind == ChangeAdd {
			cfg.Formatters.Enable = append(cfg.Formatters.Enable, types.FormatterName(change.Element))
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
	case "run.timeout":
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

	slices.Sort(cfg.Formatters.Enable)

	if cfg.Run.BuildTags != nil {
		sort.Strings(cfg.Run.BuildTags)
	}

	if cfg.Linters.Exclusions.Paths != nil {
		sort.Strings(cfg.Linters.Exclusions.Paths)
	}

	if cfg.Formatters.Exclusions.Paths != nil {
		sort.Strings(cfg.Formatters.Exclusions.Paths)
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

	*target = append(*target, change.Element)
}
