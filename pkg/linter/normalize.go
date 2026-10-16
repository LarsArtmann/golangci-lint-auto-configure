package linter

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Normalizers are pure functions observed -> (normalized, changes). They
// repair states that make a config UNLOADABLE or that this tool itself
// historically broke; everything else is merge/strategy territory. Counting
// disappears: the returned changes ARE the count, the audit trail, and the
// dry-run report at once.

// typicalReplacementChanges caps the plan of one deprecated-linter swap:
// enable removal, successor add, settings migration, exclusion sweep.
const typicalReplacementChanges = 4

// NormalizeBadSettingsKeys renames schema-invalid settings keys this tool
// once emitted (constants.KnownBadSettingsKeys). Pure: returns a normalized
// CLONE and the changes; the input is never mutated. When both the bad and
// the good key exist, the good key wins and the bad key is dropped.
func NormalizeBadSettingsKeys(observed *types.Config) (*types.Config, []Change) {
	if len(observed.Linters.Settings) == 0 || len(constants.KnownBadSettingsKeys) == 0 {
		return observed, nil
	}

	normalized := observed.Clone()

	var changes []Change

	for _, linterName := range sortedSettingNames(constants.KnownBadSettingsKeys) {
		raw, exists := normalized.Linters.Settings[linterName]
		if !exists {
			continue
		}

		settings, ok := types.AsSettingsMap(raw)
		if !ok {
			continue
		}

		changes = append(changes, renameBadSettingsKeys(settings, linterName)...)
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// renameBadSettingsKeys rewrites the schema-invalid keys of one linter's
// settings block in place (the good key wins; the bad key is dropped).
func renameBadSettingsKeys(settings types.SettingsMap, linterName string) []Change {
	keyMap := constants.KnownBadSettingsKeys[linterName]

	changes := make([]Change, 0, len(keyMap))

	for _, badKey := range sortedKeys(keyMap) {
		badValue, hasBad := settings[badKey]
		if !hasBad {
			continue
		}

		delete(settings, badKey)
		changes = append(changes, badKeyChange(settings, linterName, badKey, keyMap[badKey], badValue))
	}

	return changes
}

// badKeyChange rewrites one schema-invalid key: renamed to the good key
// when absent, dropped when the good key is already present.
func badKeyChange(
	settings types.SettingsMap,
	linterName, badKey, goodKey string,
	badValue any,
) Change {
	if _, hasGood := settings[goodKey]; hasGood {
		return Change{
			Path: PathLintersSettings + "." + linterName, Element: badKey,
			Kind: ChangeRemove, From: badValue,
			Reason: "schema-invalid key",
		}
	}

	settings[goodKey] = badValue

	return Change{
		Path: PathLintersSettings + "." + linterName, Element: badKey + "->" + goodKey,
		Kind: ChangeUpdate, From: badValue, To: badValue,
		Reason: "schema-invalid key renamed",
	}
}

func sortedSettingNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// ReplaceDeprecatedLinters rewrites enabled deprecated linters to their
// successors (policy-table supersession), migrating settings blocks and
// exclusion-rule memberships. Pure: clone in, changes out. Replacement is
// skipped when the successor is under a never-enable policy.
func ReplaceDeprecatedLinters(
	observed *types.Config,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
) (*types.Config, []Change) {
	deprecated := deprecatedIn(observed.Linters.Enable)
	if len(deprecated) == 0 {
		return observed, nil
	}

	normalized := observed.Clone()

	var changes []Change

	for _, name := range deprecated {
		record, known := constants.PolicyRecord(name)
		if !known || record.SupersessionReason == "" {
			continue
		}

		changes = append(changes, replaceDeprecatedLinter(normalized, name, record, tool, neverEnable)...)
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// replaceDeprecatedLinter migrates one enabled deprecated linter: dropped
// outright without a successor, dropped WITHOUT the successor under a
// never-enable policy, dropped when the installed tool predates the
// successor, or swapped (with settings and exclusion migrations) otherwise.
func replaceDeprecatedLinter(
	normalized *types.Config,
	name types.LinterName,
	record types.LinterPolicy,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
) []Change {
	removeChange := Change{
		Path: PathLintersEnable, Element: string(name), Kind: ChangeRemove,
		From: string(name), Reason: record.SupersessionReason,
	}

	if record.SupersededBy == "" {
		normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)

		return []Change{removeChange}
	}

	if suffix := blockedReplacementSuffix(record, tool, neverEnable); suffix != "" {
		normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)
		removeChange.Reason += suffix

		return []Change{removeChange}
	}

	normalized.Linters.Enable = replaceName(normalized.Linters.Enable, name, record.SupersededBy)

	return swapDeprecatedLinter(normalized, name, record, removeChange)
}

// swapDeprecatedLinter rewrites the enable entry to the successor and
// migrates its settings block and exclusion-rule memberships.
func swapDeprecatedLinter(
	normalized *types.Config,
	name types.LinterName,
	record types.LinterPolicy,
	removeChange Change,
) []Change {
	normalized.Linters.Enable = replaceName(normalized.Linters.Enable, name, record.SupersededBy)

	changes := make([]Change, 0, typicalReplacementChanges)
	changes = append(changes,
		removeChange,
		Change{
			Path: PathLintersEnable, Element: string(record.SupersededBy), Kind: ChangeAdd,
			To: string(record.SupersededBy), Reason: record.SupersessionReason,
		},
	)

	changes = append(changes, migrateDeprecatedSettings(normalized, name, record.SupersededBy)...)
	changes = append(changes, replaceExclusionLinters(normalized, name, record.SupersededBy)...)

	return changes
}

// blockedReplacementSuffix reports why the deprecated linter must be dropped
// WITHOUT its successor (never-enable policy, or the successor does not
// exist in the installed tool version); empty means the swap may proceed.
func blockedReplacementSuffix(
	record types.LinterPolicy,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
) string {
	if neverEnable != nil && neverEnable(record.SupersededBy) {
		return " (replacement under never-enable, NOT added)"
	}

	if !tool.IsZero() && !record.SupersessionSince.IsZero() && !tool.AtLeast(record.SupersessionSince) {
		return " (replacement needs " + record.SupersessionSince.String() + ")"
	}

	return ""
}

// NormalizeConfigShape repairs the structural facts golangci-lint refuses to
// load: a missing/wrong config version, an invalid run timeout, and the
// removed typecheck meta-linter, plus the tool-disabled linter moves. Pure.
func NormalizeConfigShape(observed *types.Config) (*types.Config, []Change) {
	normalized := observed.Clone()

	changes := repairConfigScalars(normalized)
	changes = append(changes, removeTypecheckEntries(normalized)...)

	// Tool-level disabled linters (funcorder, noinlineerr): forcibly moved to
	// the disable list and removed from enable on every run.
	changes = append(changes, moveToolDisabledLinters(normalized)...)

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// needsDurationFix checks if the timeout value needs to be fixed.
// Returns true and a reason if fixing is needed, false otherwise.
func needsDurationFix(timeout string) (bool, string) {
	if timeout == "" {
		return true, "run.timeout is empty"
	}

	_, err := time.ParseDuration(timeout)
	if err != nil {
		return true, fmt.Sprintf("run.timeout %q is invalid", timeout)
	}

	return false, ""
}

// repairConfigScalars fixes the scalar fields golangci-lint refuses to load:
// a missing/wrong config version and an invalid run timeout.
func repairConfigScalars(normalized *types.Config) []Change {
	var changes []Change

	if normalized.Version != "2" {
		changes = append(changes, Change{
			Path: "version", Kind: ChangeUpdate,
			From: normalized.Version, To: "2",
			Reason: "golangci-lint v2 config schema",
		})

		normalized.Version = "2"
	}

	if durationInvalid, durationReason := needsDurationFix(
		normalized.Run.Timeout,
	); durationInvalid &&
		normalized.Run.Timeout != "" {
		changes = append(changes, Change{
			Path: PathRunTimeout, Kind: ChangeUpdate,
			From:   normalized.Run.Timeout,
			To:     constants.DefaultTimeout,
			Reason: durationReason,
		})

		normalized.Run.Timeout = constants.DefaultTimeout
	}

	return changes
}

// removeTypecheckEntries strips the removed typecheck meta-linter from both
// membership lists (govet covers it).
func removeTypecheckEntries(normalized *types.Config) []Change {
	var changes []Change

	for _, list := range []struct {
		path  string
		field *[]types.LinterName
	}{
		{PathLintersEnable, &normalized.Linters.Enable},
		{PathLintersDisable, &normalized.Linters.Disable},
	} {
		if containsLinterName(toStrings(*list.field), "typecheck") {
			*list.field = removeName(*list.field, "typecheck")

			changes = append(changes, Change{
				Path: list.path, Element: "typecheck", Kind: ChangeRemove,
				From:   "typecheck",
				Reason: "typecheck is a removed meta-linter (govet covers it)",
			})
		}
	}

	return changes
}

// moveToolDisabledLinters relocates every tool-level disabled linter from
// the enable list to the disable list (a tool-level decision, not user
// choice), in place.
func moveToolDisabledLinters(normalized *types.Config) []Change {
	var changes []Change

	for _, name := range constants.LintersInTier(types.TierDisabled) {
		if !containsLinterName(toStrings(normalized.Linters.Enable), string(name)) {
			continue
		}

		changes = append(changes, moveToolDisabledLinter(normalized, name)...)
	}

	return changes
}

// moveToolDisabledLinter relocates one tool-level disabled linter: out of
// the enable list, into the disable list (idempotent).
func moveToolDisabledLinter(normalized *types.Config, name types.LinterName) []Change {
	normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)

	reason := "tool-level disabled"
	if record, known := constants.PolicyRecord(name); known && record.Tier.Rationale != "" {
		reason = record.Tier.Rationale
	}

	changes := []Change{
		{
			Path: PathLintersEnable, Element: string(name), Kind: ChangeRemove,
			From: string(name), Reason: reason,
		},
		{
			Path: PathLintersDisable, Element: string(name), Kind: ChangeAdd,
			To: string(name), Reason: reason,
		},
	}

	if !containsLinterName(toStrings(normalized.Linters.Disable), string(name)) {
		normalized.Linters.Disable = append(normalized.Linters.Disable, name)
	}

	return changes
}

// PruneOrphanSettings removes settings blocks for linters that are neither
// enabled nor disabled (dead weight; the legacy prune pass). Pure.
func PruneOrphanSettings(observed *types.Config) (*types.Config, []Change) {
	if len(observed.Linters.Settings) == 0 {
		return observed, nil
	}

	enabled := types.NewSet(observed.Linters.Enable...)
	disabled := types.NewSet(observed.Linters.Disable...)

	changes := make([]Change, 0, len(observed.Linters.Settings))

	normalized := observed.Clone()

	for _, name := range sortedSettingNames(observed.Linters.Settings) {
		if enabled.Contains(types.LinterName(name)) || disabled.Contains(types.LinterName(name)) {
			continue
		}

		delete(normalized.Linters.Settings, name)

		changes = append(changes, Change{
			Path: PathLintersSettings, Element: name, Kind: ChangeRemove,
			Reason: "orphaned settings block (linter neither enabled nor disabled)",
		})
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// RemoveUnknownLinters drops linters.enable and linters.disable entries the
// detected golangci-lint binary cannot name. Either entry fails config LOAD
// ("unknown linters"), so the whole file is dead weight until the name is
// gone — this normalizer heals configs poisoned by plugin-linter writes and
// lets the rest of the pipeline analyze reality. A nil/empty known set (no
// binary, failed probe) is a no-op: an unknown binary gates nothing. Pure.
func RemoveUnknownLinters(observed *types.Config, known map[types.LinterName]bool) (*types.Config, []Change) {
	if len(known) == 0 {
		return observed, nil
	}

	normalized := observed.Clone()

	changes := make([]Change, 0, 2)

	kept, removed := partitionUnknown(normalized.Linters.Enable, known)
	if len(removed) > 0 {
		normalized.Linters.Enable = kept

		changes = append(changes, unknownLinterChanges(removed, PathLintersEnable)...)
	}

	kept, removed = partitionUnknown(normalized.Linters.Disable, known)
	if len(removed) > 0 {
		normalized.Linters.Disable = kept

		changes = append(changes, unknownLinterChanges(removed, PathLintersDisable)...)
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// unknownLinterChangeReason explains why an unknown name must go: the entry
// is not an opinion, it is a load failure.
const unknownLinterChangeReason = "not supported by the detected golangci-lint binary; " +
	"the config cannot be loaded while it is listed (remove it, or run a binary " +
	"built with this linter as a module plugin)"

func unknownLinterChanges(removed []types.LinterName, path string) []Change {
	changes := make([]Change, 0, len(removed))
	for _, name := range removed {
		changes = append(changes, Change{
			Path: path, Element: string(name), Kind: ChangeRemove,
			From:   string(name),
			Reason: unknownLinterChangeReason,
		})
	}

	return changes
}

func partitionUnknown(names []types.LinterName, known map[types.LinterName]bool) (kept, removed []types.LinterName) {
	for _, name := range names {
		if known[name] {
			kept = append(kept, name)

			continue
		}

		removed = append(removed, name)
	}

	return kept, removed
}

// RemoveRedundancies drops linters and formatters superseded by an enabled
// counterpart (the RedundancyRelations table: lll dies with golines, gofmt
// dies with gofumpt). Pure.
func RemoveRedundancies(observed *types.Config, desiredFormatters []types.FormatterName) (*types.Config, []Change) {
	normalized := observed.Clone()

	changes := make([]Change, 0, 2)

	formatters := types.NewSet(desiredFormatters...)
	for _, name := range normalized.Formatters.Enable {
		formatters.Add(name)
	}

	for _, relation := range constants.RedundancyRelations {
		if !formatters.Contains(relation.Superseder) {
			continue
		}

		changes = append(changes, removeRedundantEntry(normalized, relation)...)
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// removeRedundantEntry drops one superseded entry from the list its relation
// kind names, reporting the removal as a change.
func removeRedundantEntry(normalized *types.Config, relation constants.RedundancyRelation) []Change {
	if relation.Kind == constants.RedundancyLinter {
		if !containsLinterName(toStrings(normalized.Linters.Enable), string(relation.Superseded)) {
			return nil
		}

		normalized.Linters.Enable = removeName(normalized.Linters.Enable, relation.Superseded)

		return []Change{{
			Path: PathLintersEnable, Element: string(relation.Superseded), Kind: ChangeRemove,
			From:   string(relation.Superseded),
			Reason: relation.Reason,
		}}
	}

	superseded := types.FormatterName(relation.Superseded)
	if !containsFormatterName(normalized.Formatters.Enable, superseded) {
		return nil
	}

	normalized.Formatters.Enable = removeFormatterName(normalized.Formatters.Enable, superseded)

	return []Change{{
		Path: PathFormattersEnable, Element: string(superseded), Kind: ChangeRemove,
		From:   string(superseded),
		Reason: relation.Reason,
	}}
}

func removeFormatterName(names []types.FormatterName, name types.FormatterName) []types.FormatterName {
	out := names[:0:0]

	for _, candidate := range names {
		if candidate != name {
			out = append(out, candidate)
		}
	}

	return out
}

func toStrings(names []types.LinterName) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, string(name))
	}

	return out
}

// RenameDeprecatedInExclusionRules rewrites EVERY deprecated linter name in
// exclusion rules (not only enabled ones): names with a successor are
// renamed (a successor already listed dedupes the predecessor entry away),
// names without one are dropped. Rules scoped by Text or Source whose linter
// list the sweep emptied are dropped entirely — left behind, the pattern
// would suppress matching findings from every linter (the legacy removal
// contract; path-scoped rules survive with an emptied list). Pure.
func RenameDeprecatedInExclusionRules(observed *types.Config) (*types.Config, []Change) {
	normalized := observed.Clone()

	var changes []Change

	rules := normalized.Linters.Exclusions.Rules
	kept := make([]types.ExclusionRuleConfig, 0, len(rules))

	for _, rule := range rules {
		swept, ruleChanged, ruleChanges := sweepExclusionRuleLinters(rule)
		changes = append(changes, ruleChanges...)

		if keptRule, survives := keepSweptRule(rule, swept, ruleChanged); survives {
			kept = append(kept, keptRule)
		}
	}

	if len(changes) == 0 {
		return observed, nil
	}

	normalized.Linters.Exclusions.Rules = kept

	return normalized, changes
}

// keepSweptRule decides a swept rule's fate: untouched rules pass through,
// rules emptied by the sweep whose scope is Text or Source are dropped
// entirely (the pattern would suppress matching findings from every linter),
// and every other swept rule survives with the rewritten linter list.
func keepSweptRule(
	rule types.ExclusionRuleConfig,
	swept []string,
	ruleChanged bool,
) (types.ExclusionRuleConfig, bool) {
	if !ruleChanged {
		return rule, true
	}

	if len(swept) == 0 && (rule.Text != "" || rule.Source != "") {
		return rule, false
	}

	rule.Linters = swept

	return rule, true
}

// sweepExclusionRuleLinters rewrites one rule's linter list: deprecated names
// with a successor become that successor (deduped against what the rule
// already lists), names without one vanish. The returned list is the swept
// linters; ruleChanged reports whether any deprecated name was touched.
func sweepExclusionRuleLinters(rule types.ExclusionRuleConfig) ([]string, bool, []Change) {
	kept := make([]string, 0, len(rule.Linters))
	seen := make(map[string]bool, len(rule.Linters))

	var changes []Change

	changed := false

	for _, entry := range rule.Linters {
		record, known := constants.PolicyRecord(types.LinterName(entry))
		if !known || record.SupersessionReason == "" {
			kept = keepSweptEntry(entry, kept, seen)

			continue
		}

		changed = true

		if record.SupersededBy == "" {
			changes = append(changes, sweptDropChange(entry, record))

			continue
		}

		var rename Change

		kept, rename = sweepRename(entry, record, kept, seen)
		changes = append(changes, rename)
	}

	return kept, changed, changes
}

// keepSweptEntry keeps a non-deprecated entry, deduped against what the
// rule already lists.
func keepSweptEntry(entry string, kept []string, seen map[string]bool) []string {
	if !seen[entry] {
		seen[entry] = true
		kept = append(kept, entry)
	}

	return kept
}

// sweptDropChange plans the removal of a deprecated name without a
// successor from one rule's linter list.
func sweptDropChange(entry string, record types.LinterPolicy) Change {
	return Change{
		Path: PathLintersExclusionRules, Element: entry, Kind: ChangeRemove,
		From:   entry,
		Reason: record.SupersessionReason,
	}
}

// sweepRename plans one rename to the successor, deduped against what the
// rule already lists, and returns the extended kept list.
func sweepRename(
	entry string,
	record types.LinterPolicy,
	kept []string,
	seen map[string]bool,
) ([]string, Change) {
	successor := string(record.SupersededBy)

	if !seen[successor] {
		seen[successor] = true
		kept = append(kept, successor)
	}

	return kept, Change{
		Path:    PathLintersExclusionRules,
		Element: entry + "->" + successor,
		Kind:    ChangeUpdate,
		Reason:  record.SupersessionReason,
	}
}

func deprecatedIn(enabled []types.LinterName) []types.LinterName {
	var deprecated []types.LinterName

	for _, name := range enabled {
		if record, known := constants.PolicyRecord(name); known && record.SupersessionReason != "" {
			deprecated = append(deprecated, name)
		}
	}

	slices.Sort(deprecated)

	return deprecated
}

func removeName(names []types.LinterName, name types.LinterName) []types.LinterName {
	out := names[:0:0]

	for _, candidate := range names {
		if candidate != name {
			out = append(out, candidate)
		}
	}

	return out
}

func replaceName(names []types.LinterName, from, to types.LinterName) []types.LinterName {
	out := names[:0:0]

	for _, candidate := range names {
		if candidate == from {
			out = append(out, to)

			continue
		}

		out = append(out, candidate)
	}

	return out
}

// migrateDeprecatedSettings moves a deprecated linter's settings block to
// its successor and renames KnownBadSettingsKeys entries for the successor.
func migrateDeprecatedSettings(cfg *types.Config, from, successor types.LinterName) []Change {
	raw, exists := cfg.Linters.Settings[string(from)]
	if !exists {
		return nil
	}

	settings, ok := types.AsSettingsMap(raw)
	if !ok || settings.IsEmpty() {
		delete(cfg.Linters.Settings, string(from))

		return []Change{{
			Path: "linters.settings." + string(from), Kind: ChangeRemove,
			Reason: "deprecated linter settings block removed (empty)",
		}}
	}

	delete(cfg.Linters.Settings, string(from))

	if _, hasSuccessor := cfg.Linters.Settings[string(successor)]; !hasSuccessor {
		// Store the RAW map type: SettingsMap is a defined type and would
		// fail every map[string]any assertion downstream.
		cfg.Linters.Settings[string(successor)] = map[string]any(settings.Clone())
	}

	return []Change{{
		Path: PathLintersSettings, Element: string(from) + "->" + string(successor),
		Kind: ChangeUpdate, From: settings.Clone(), To: settings.Clone(),
		Reason: "settings block migrated to successor",
	}}
}

// replaceExclusionLinters rewrites a deprecated linter name inside
// linters.exclusions.rules[].linters; names without a successor are dropped.
func replaceExclusionLinters(cfg *types.Config, from, successor types.LinterName) []Change {
	var changes []Change

	for i := range cfg.Linters.Exclusions.Rules {
		rule := &cfg.Linters.Exclusions.Rules[i]

		if !containsLinterName(rule.Linters, string(from)) {
			continue
		}

		rule.Linters = replaceStringEntry(rule.Linters, string(from), string(successor))

		changes = append(changes, Change{
			Path: PathLintersExclusionRules, Element: string(from) + "->" + string(successor),
			Kind: ChangeUpdate, Reason: "exclusion rule follows the replacement",
		})
	}

	return changes
}

func containsLinterName(names []string, want string) bool {
	return slices.Contains(names, want)
}

func replaceStringEntry(names []string, from, target string) []string {
	out := names[:0:0]

	for _, name := range names {
		if name == from {
			if target != "" {
				out = append(out, target)
			}

			continue
		}

		out = append(out, name)
	}

	return out
}
