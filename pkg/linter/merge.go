package linter

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ChangeKind is the closed union of config mutations a plan can contain.
type ChangeKind string

const (
	// ChangeAdd introduces a value that was absent.
	ChangeAdd ChangeKind = "add"
	// ChangeUpdate rewrites an existing value.
	ChangeUpdate ChangeKind = "update"
	// ChangeRemove drops a value.
	ChangeRemove ChangeKind = "remove"
)

// Change is one planned config mutation with a stable identity. The Key
// (Path + element) is what makes plans comparable across runs and drivers:
// lint mode reports changes, apply mode applies them, and both speak about
// the SAME objects — detect/repair drift is structurally impossible.
type Change struct {
	// Path is the config location (PathLintersEnable, "run.go",
	// "linters.settings.goconst").
	Path string
	// Element identifies the entry within the path ("goconst" for list
	// membership, "min-len" for settings keys, "" for scalar paths).
	Element string
	// Kind is the mutation kind.
	Kind ChangeKind
	// From is the observed value (nil for adds).
	From any
	// To is the desired value (nil for removes).
	To any
	// Reason explains WHY (rationale from the policy record or rule).
	Reason string
}

// Key returns the stable identity of this change: path + element.
func (c Change) Key() string {
	if c.Element == "" {
		return c.Path
	}

	return c.Path + "." + c.Element
}

// String renders the change for logs and plan output.
func (c Change) String() string {
	switch c.Kind {
	case ChangeAdd:
		return fmt.Sprintf("%s: add %s = %v", c.Key(), c.Path, c.To)
	case ChangeUpdate:
		return fmt.Sprintf("%s: update %s: %v -> %v", c.Key(), c.Path, c.From, c.To)
	case ChangeRemove:
		return fmt.Sprintf("%s: remove from %s (was %v)", c.Key(), c.Path, c.From)
	default:
		return fmt.Sprintf("%s: %s", c.Key(), c.Kind)
	}
}

// ConflictError reports a plan that contains two different outcomes for the
// same stable key — a writer bug, never a user-facing condition.
type ConflictError struct {
	Key    string
	First  Change
	Second Change
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("plan conflict on %s: %q vs %q", e.Key, e.First.String(), e.Second.String())
}

// ValidatePlan detects conflicts: two changes with the same stable key but
// different kind/outcome. A conflict is a programming error in the strategy
// table, surfaced as a classified rejection.
func ValidatePlan(changes []Change) error {
	seen := make(map[string]Change, len(changes))

	for _, change := range changes {
		previous, exists := seen[change.Key()]
		if !exists {
			seen[change.Key()] = change

			continue
		}

		if previous.Kind != change.Kind || !fmtEqual(previous.To, change.To) {
			return &ConflictError{Key: change.Key(), First: previous, Second: change}
		}
	}

	return nil
}

func fmtEqual(a, b any) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

// MergeStrategy decides who wins on a config path when desired and observed
// disagree. Strategies are DATA (a path-keyed table), not code scattered
// through mutator passes — flipping --force-settings is a table rewrite.
type MergeStrategy string

// Config path keys shared by the strategy table, the planners, the applier,
// and the audit projection. One constant per path the engine touches.
const (
	PathLintersEnable         = "linters.enable"
	PathLintersDisable        = "linters.disable"
	PathLintersSettings       = "linters.settings"
	PathLintersExclusionRules = "linters.exclusions.rules"
	PathFormattersEnable      = "formatters.enable"
	PathRunTimeout            = "run.timeout"
)

const (
	// StrategyEnsurePresent adds desired entries to a list, keeping every
	// user entry (the enable-list semantics: recommend, never strip).
	StrategyEnsurePresent MergeStrategy = "ensure-present"
	// StrategyUnion merges desired entries into a list/set, deduplicating
	// (exclusion paths: user patterns plus defaults).
	StrategyUnion MergeStrategy = "union"
	// StrategyPreferUser fills only missing values; existing user settings
	// are never overwritten (the default settings semantics, idempotency
	// guard included).
	StrategyPreferUser MergeStrategy = "prefer-user"
	// StrategyPreferDesired overwrites the observed value with the desired
	// one (the --force-settings flip).
	StrategyPreferDesired MergeStrategy = "prefer-desired"
	// StrategyEnsureDefault sets a scalar when absent or zero (issues-exit-code).
	StrategyEnsureDefault MergeStrategy = "ensure-default"
	// StrategyCapGoVersion sets run.go to min(local, binary build Go).
	StrategyCapGoVersion MergeStrategy = "cap-go-version"
	// StrategyReplace removes-and-replaces a list wholesale (disable lists
	// under anti-gaming enforcement).
	StrategyReplace MergeStrategy = "replace"
)

// Plan capacity hints: a full default render plans roughly 60-80 changes on
// an empty config; section-level planners are an order of magnitude smaller.
const (
	typicalPlanSize    = 64
	typicalSectionSize = 8
)

// DefaultMergeStrategies is the path-keyed strategy table for the converge
// engine. One row per config path the engine touches.
//
//nolint:gochecknoglobals // immutable strategy table, read-only after init
var DefaultMergeStrategies = map[string]MergeStrategy{
	PathLintersEnable:          StrategyEnsurePresent,
	PathLintersSettings:        StrategyPreferUser,
	PathFormattersEnable:       StrategyEnsurePresent,
	"formatters.settings":      StrategyPreferUser,
	"linters.exclusions.paths": StrategyUnion,
	"formatters.exclusions.paths": StrategyUnion,
	PathLintersExclusionRules:  StrategyUnion,
	"run.go":                          StrategyCapGoVersion,
	PathRunTimeout:                     StrategyEnsureDefault,
	"run.allow-parallel-runners":      StrategyEnsureDefault,
	"run.allow-serial-runners":        StrategyEnsureDefault,
	"run.issues-exit-code":            StrategyEnsureDefault,
	"run.build-tags":                  StrategyUnion,
	"issues.max-issues-per-linter":    StrategyEnsureDefault,
	"issues.max-same-issues":          StrategyEnsureDefault,
	"linters.exclusions.generated":    StrategyEnsureDefault,
	"formatters.exclusions.generated": StrategyEnsureDefault,
}

// MergeOptions carries the user-intent flips for one convergence run.
type MergeOptions struct {
	// ForceSettings flips every settings path from PreferUser to
	// PreferDesired (the --force-settings escape hatch).
	ForceSettings bool
	// Pragmatic is absorbed by the renderer; kept here for plan symmetry.
	Pragmatic bool
}

// StrategiesFor returns the effective strategy table for the options:
// ForceSettings rewrites the settings rows to PreferDesired. This is the
// WHOLE implementation of --force-settings — a table flip, not a flag
// threaded through twelve mutators.
func StrategiesFor(options MergeOptions) map[string]MergeStrategy {
	strategies := make(map[string]MergeStrategy, len(DefaultMergeStrategies))
	maps.Copy(strategies, DefaultMergeStrategies)

	if options.ForceSettings {
		strategies[PathLintersSettings] = StrategyPreferDesired
		strategies["formatters.settings"] = StrategyPreferDesired
	}

	return strategies
}

// PlanDesired computes the changes that move observed toward desired. Pure:
// no mutation of either input, deterministic order (paths sorted, then keys).
func PlanDesired(observed *types.Config, desired DesiredConfig, strategies map[string]MergeStrategy) []Change {
	changes := make([]Change, 0, typicalPlanSize)

	changes = append(changes, planEnableList(observed, desired)...)
	changes = append(changes, planFormatters(observed, desired)...)
	changes = append(changes, planSettings(observed, desired, strategies)...)
	changes = append(changes, planExclusions(observed, desired, strategies)...)
	changes = append(changes, planRun(observed, desired, strategies)...)
	changes = append(changes, planIssues(observed, desired)...)

	sortChanges(changes)

	return changes
}

// planPathOrder is the canonical section order of a plan (stable across
// runs; generation order within a section is preserved — element order
// inside a section is part of the byte contract, e.g. default exclusion
// rules). The rank map is derived so the order reads as data, not as magic
// indexes.
//
//nolint:gochecknoglobals // derived once at init, read-only afterwards
var planPathRank = buildPlanPathRank()

func buildPlanPathRank() map[string]int {
	order := []string{
		PathLintersEnable,
		PathLintersDisable,
		PathLintersSettings,
		"linters.exclusions.generated",
		"linters.exclusions.paths",
		PathLintersExclusionRules,
		PathFormattersEnable,
		"formatters.settings",
		"formatters.exclusions.generated",
		"formatters.exclusions.paths",
		"run.go",
		PathRunTimeout,
		"run.allow-parallel-runners",
		"run.allow-serial-runners",
		"run.issues-exit-code",
		"run.build-tags",
		"issues.max-issues-per-linter",
		"issues.max-same-issues",
	}

	rank := make(map[string]int, len(order))
	for index, path := range order {
		rank[path] = index
	}

	return rank
}

func sortChanges(changes []Change) {
	sort.SliceStable(changes, func(i, j int) bool {
		return planPathRank[changes[i].Path] < planPathRank[changes[j].Path]
	})
}

func planEnableList(observed *types.Config, desired DesiredConfig) []Change {
	existing := types.NewSet(observed.Linters.Enable...)

	var changes []Change

	for _, name := range desired.EnableLinters {
		if existing.Contains(name) {
			continue
		}

		record, known := constants.PolicyRecord(name)

		reason := "recommended"
		if known && record.Rationale != "" {
			reason = record.Rationale
		}

		changes = append(changes, Change{
			Path: PathLintersEnable, Element: string(name),
			Kind: ChangeAdd, To: string(name), Reason: reason,
		})
	}

	return changes
}

func planFormatters(observed *types.Config, desired DesiredConfig) []Change {
	existing := types.NewSet(observed.Formatters.Enable...)

	var changes []Change

	for _, name := range desired.EnableFormatters {
		if existing.Contains(name) {
			continue
		}

		reason := "core formatter"
		if record, known := constants.FormatterPolicyRecord(name); known && record.Rationale != "" {
			reason = record.Rationale
		}

		changes = append(changes, Change{
			Path: PathFormattersEnable, Element: string(name),
			Kind: ChangeAdd, To: string(name), Reason: reason,
		})
	}

	return changes
}

func planSettings(
	observed *types.Config,
	desired DesiredConfig,
	strategies map[string]MergeStrategy,
) []Change {
	changes := make([]Change, 0, typicalSectionSize)

	changes = append(changes, planSettingsSection(
		PathLintersSettings, observed.Linters.Settings, desiredSettingsToMaps(desired.LinterSettings), strategies,
	)...)
	changes = append(changes, planSettingsSection(
		"formatters.settings",
		observed.Formatters.Settings,
		desiredSettingsToMaps(desired.FormatterSettings),
		strategies,
	)...)

	return changes
}

func desiredSettingsToMaps[T ~string](m map[T][]types.VersionedSettings) map[string]map[string]any {
	out := make(map[string]map[string]any, len(m))
	for name, blocks := range m {
		if len(blocks) == 0 {
			continue
		}

		out[string(name)] = blocks[0].Settings.ToMap()
	}

	return out
}

func planSettingsSection(
	path string,
	observed map[string]any,
	desired map[string]map[string]any,
	strategies map[string]MergeStrategy,
) []Change {
	strategy, ok := strategies[path]
	if !ok {
		strategy = StrategyPreferUser
	}

	changes := make([]Change, 0, len(desired))

	for _, name := range sortedKeys(desired) {
		if change, planned := planSettingsBlock(path, name, observed[name], desired[name], strategy); planned {
			changes = append(changes, change)
		}
	}

	return changes
}

func planSettingsBlock(
	path, name string,
	observedRaw any,
	desiredBlock map[string]any,
	strategy MergeStrategy,
) (Change, bool) {
	if !exists(observedRaw) {
		return missingBlock(path, name, desiredBlock), true
	}

	observedBlock, isMap := types.AsSettingsMap(observedRaw)
	if !isMap || observedBlock.IsEmpty() {
		return Change{
			Path: path, Element: name, Kind: ChangeAdd,
			From: observedRaw, To: desiredBlock, Reason: reasonForStrategy(strategy),
		}, true
	}

	if strategy == StrategyPreferUser {
		return Change{}, false // existing non-empty settings are never touched
	}

	if strategy == StrategyPreferDesired && !settingsEqual(observedBlock, desiredBlock) {
		return Change{
			Path: path, Element: name, Kind: ChangeUpdate,
			From: observedBlock.Clone(), To: desiredBlock, Reason: "force-settings",
		}, true
	}

	return Change{}, false
}

func exists(v any) bool {
	return v != nil
}

func missingBlock(path, name string, desiredBlock map[string]any) Change {
	return Change{
		Path: path, Element: name, Kind: ChangeAdd,
		From: nil, To: desiredBlock, Reason: "safe default settings",
	}
}

func reasonForStrategy(strategy MergeStrategy) string {
	if strategy == StrategyPreferDesired {
		return "force-settings"
	}

	return "safe default settings"
}

func settingsEqual(a types.SettingsMap, b map[string]any) bool {
	return fmtEqual(a.Clone(), b)
}

func planExclusions(
	observed *types.Config,
	desired DesiredConfig,
	strategies map[string]MergeStrategy,
) []Change {
	changes := make([]Change, 0, typicalSectionSize)

	changes = append(changes, planPathList(
		"linters.exclusions.paths", observed.Linters.Exclusions.Paths, desired.LinterExclusionPaths, strategies,
	)...)
	changes = append(changes, planPathList(
		"formatters.exclusions.paths",
		observed.Formatters.Exclusions.Paths,
		desired.FormatterExclusionPaths,
		strategies,
	)...)
	changes = append(changes, planPathList(
		"generated.exclusions", observed.Linters.Exclusions.Paths, desired.GeneratedExclusionPaths, strategies,
	)...)
	changes = append(changes, planDefaultExclusionRules(observed, desired)...)

	return changes
}

// planDefaultExclusionRules merges the default test-file exclusion rules into
// the observed config by RuleKey: a matching observed rule absorbs the
// default's linters (union), a non-matching default is appended. This is the
// declarative twin of the legacy exclusion merge (plan T05 / gotcha on
// exclusion-rule dedup).
func planDefaultExclusionRules(observed *types.Config, desired DesiredConfig) []Change {
	var changes []Change

	// Legacy-exact matching: later rules with the same RuleKey shadow earlier
	// ones (the legacy merge builds a map while iterating forward).
	existingByKey := make(map[string]int, len(observed.Linters.Exclusions.Rules))
	for i := range observed.Linters.Exclusions.Rules {
		existingByKey[observed.Linters.Exclusions.Rules[i].RuleKey()] = i
	}

	for _, defaultRule := range desired.ExclusionRules {
		key := defaultRule.RuleKey()

		idx, exists := existingByKey[key]
		if !exists {
			changes = append(changes, Change{
				Path: PathLintersExclusionRules, Element: key, Kind: ChangeAdd,
				To: cloneRule(defaultRule), Reason: "default exclusion rule",
			})

			continue
		}

		existingLinters := observed.Linters.Exclusions.Rules[idx].Linters

		union, added := unionStrings(existingLinters, defaultRule.Linters)
		if !added {
			continue
		}

		changes = append(changes, Change{
			Path: PathLintersExclusionRules, Element: key, Kind: ChangeUpdate,
			From: append([]string{}, existingLinters...), To: union,
			Reason: "default exclusion rule unioned with user rule",
		})
	}

	return changes
}

func unionStrings(existing, additions []string) ([]string, bool) {
	set := make(map[string]bool, len(existing))
	for _, entry := range existing {
		set[entry] = true
	}

	union := append([]string{}, existing...)
	added := false

	for _, entry := range additions {
		if !set[entry] {
			union = append(union, entry)
			added = true
		}
	}

	return union, added
}

func cloneRule(rule types.ExclusionRuleConfig) types.ExclusionRuleConfig {
	clone := rule
	clone.Linters = append([]string{}, rule.Linters...)

	return clone
}

func planPathList(path string, observed []string, desired []string, strategies map[string]MergeStrategy) []Change {
	if len(desired) == 0 {
		return nil
	}

	if strategies[path] != StrategyUnion && strategies[path] != StrategyEnsurePresent {
		return nil
	}

	existing := types.NewSet(observed...)

	changes := make([]Change, 0, len(desired))

	for _, entry := range desired {
		if existing.Contains(entry) {
			continue
		}

		changes = append(changes, Change{
			Path: path, Element: entry, Kind: ChangeAdd,
			To: entry, Reason: "default exclusion",
		})
	}

	return changes
}

func planRun(observed *types.Config, desired DesiredConfig, strategies map[string]MergeStrategy) []Change {
	changes := make([]Change, 0, typicalSectionSize)

	if strategies["run.go"] == StrategyCapGoVersion {
		if change, changed := capRunGo(observed.Run.Go, desired.RunDefaults.Go); changed {
			changes = append(changes, change)
		}
	}

	if observed.Run.Timeout == "" || observed.Run.Timeout == "0s" {
		changes = append(changes, Change{
			Path: PathRunTimeout, Kind: ChangeAdd,
			To: desired.RunDefaults.Timeout, Reason: "default timeout",
		})
	}

	changes = append(changes, planRunnerFlags(observed)...)

	if observed.Run.IssuesExitCode == 0 && desired.RunDefaults.IssuesExitCode != 0 {
		changes = append(changes, Change{
			Path: "run.issues-exit-code", Kind: ChangeAdd,
			To: desired.RunDefaults.IssuesExitCode, Reason: "lint failures must break CI",
		})
	}

	tagChanges := planPathList("run.build-tags", observed.Run.BuildTags, desired.RunDefaults.BuildTags, strategies)
	changes = append(changes, tagChanges...)

	return changes
}

func planRunnerFlags(observed *types.Config) []Change {
	var changes []Change

	if !observed.Run.AllowParallelRunners {
		changes = append(changes, Change{
			Path: "run.allow-parallel-runners", Kind: ChangeUpdate,
			From: false, To: true, Reason: "parallel runners",
		})
	}

	if !observed.Run.AllowSerialRunners {
		changes = append(changes, Change{
			Path: "run.allow-serial-runners", Kind: ChangeUpdate,
			From: false, To: true, Reason: "serial runners",
		})
	}

	return changes
}

func capRunGo(observed, desired string) (Change, bool) {
	if desired == "" || observed == desired {
		return Change{}, false
	}

	reason := "track local Go (capped at binary build Go)"

	if observed != "" {
		if observedParsed, observedErr := types.ParseGoVersion(observed); observedErr == nil {
			if desiredParsed, desiredErr := types.ParseGoVersion(
				desired,
			); desiredErr == nil &&
				observedParsed.Compare(desiredParsed) == 0 {
				reason = "normalize to major.minor"
			}
		}
	}

	return Change{
		Path: "run.go", Kind: ChangeUpdate,
		From: observed, To: desired, Reason: reason,
	}, true
}

func planIssues(observed *types.Config, desired DesiredConfig) []Change {
	var changes []Change

	if observed.Issues.MaxIssuesPerLinter == 0 && desired.IssuesDefaults.MaxIssuesPerLinter != 0 {
		changes = append(changes, Change{
			Path: "issues.max-issues-per-linter", Kind: ChangeAdd,
			To: desired.IssuesDefaults.MaxIssuesPerLinter, Reason: "prevent silent truncation",
		})
	}

	if observed.Issues.MaxSameIssues == 0 && desired.IssuesDefaults.MaxSameIssues != 0 {
		changes = append(changes, Change{
			Path: "issues.max-same-issues", Kind: ChangeAdd,
			To: desired.IssuesDefaults.MaxSameIssues, Reason: "prevent silent truncation",
		})
	}

	changes = append(changes, planGeneratedDefaults(observed)...)

	return changes
}

func planGeneratedDefaults(observed *types.Config) []Change {
	var changes []Change

	if observed.Linters.Exclusions.Generated == "" {
		changes = append(changes, Change{
			Path: "linters.exclusions.generated", Kind: ChangeAdd,
			To: "lax", Reason: "generated-file exclusions",
		})
	}

	if observed.Formatters.Exclusions.Generated == "" {
		changes = append(changes, Change{
			Path: "formatters.exclusions.generated", Kind: ChangeAdd,
			To: "lax", Reason: "generated-file exclusions",
		})
	}

	return changes
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

var _ = strings.TrimSpace
