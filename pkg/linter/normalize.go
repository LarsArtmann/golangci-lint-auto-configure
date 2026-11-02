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

// resolveSupersessionChain walks a linter's supersession chain to the
// terminal replacement applicable to the installed tool, so a chained
// deprecation (exhaustivestruct -> exhaustruct -> exhaustruct_v5) converges
// in ONE configure run instead of oscillating across two. Every hop is
// gated by its own never-enable policy and SupersessionSince, mirroring
// blockedReplacementSuffix. Returns the terminal record and whether the
// swap may proceed: false when the chain dead-ends — no successor, an
// intermediate that is itself superseded further but blocked, or a
// successor that would itself be swapped away. The caller then drops the
// linter without a replacement, the same outcome a single deprecated
// linter gets on a tool that predates its successor.
func resolveSupersessionChain(
	record types.LinterPolicy,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
) (types.LinterPolicy, bool) {
	for record.SupersededBy != "" {
		if suffix := blockedReplacementSuffix(record, tool, neverEnable); suffix != "" {
			return record, false
		}

		next, known := constants.PolicyRecord(record.SupersededBy)
		if !known {
			return record, false
		}

		record = next

		if record.SupersessionReason == "" {
			return record, true
		}
	}

	return record, false
}

// replaceDeprecatedLinter migrates one enabled deprecated linter: dropped
// outright without a (reachable) successor, or swapped (with settings and
// exclusion migrations) to the terminal successor of its supersession
// chain otherwise. Blocked hops (never-enable policy, successor gated on a
// newer tool) drop the linter WITHOUT any replacement.
func replaceDeprecatedLinter(
	normalized *types.Config,
	name types.LinterName,
	record types.LinterPolicy,
	tool types.GolangciVersion,
	neverEnable func(types.LinterName) bool,
) []Change {
	terminal, swappable := resolveSupersessionChain(record, tool, neverEnable)

	removeChange := Change{
		Path: PathLintersEnable, Element: string(name), Kind: ChangeRemove,
		From: string(name), Reason: record.SupersessionReason,
	}

	if !swappable {
		if suffix := blockedReplacementSuffix(terminal, tool, neverEnable); suffix != "" {
			removeChange.Reason += suffix
		}

		normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)

		return []Change{removeChange}
	}

	return swapDeprecatedLinter(normalized, name, terminal, removeChange)
}

// swapDeprecatedLinter rewrites the enable entry to the terminal successor
// and migrates its settings block and exclusion-rule memberships. When
// several deprecated linters share one successor (deadcode, varcheck,
// structcheck all land on unused), the successor is added once and the
// later predecessors are dropped — the change list stays truthful about it.
func swapDeprecatedLinter(
	normalized *types.Config,
	name types.LinterName,
	record types.LinterPolicy,
	removeChange Change,
) []Change {
	successor := record.Name
	successorExisted := slices.Contains(normalized.Linters.Enable, successor)
	normalized.Linters.Enable = replaceNameUnique(normalized.Linters.Enable, name, successor)

	changes := make([]Change, 0, typicalReplacementChanges)
	changes = append(changes, removeChange)

	if !successorExisted {
		changes = append(changes,
			Change{
				Path: PathLintersEnable, Element: string(successor), Kind: ChangeAdd,
				To: string(successor), Reason: removeChange.Reason,
			},
		)
	}

	changes = append(changes, migrateDeprecatedSettings(normalized, name, successor)...)
	changes = append(changes, replaceExclusionLinters(normalized, name, successor)...)

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

func partitionUnknown(
	names []types.LinterName, known map[types.LinterName]bool,
) ([]types.LinterName, []types.LinterName) {
	var kept, removed []types.LinterName

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
// exclusion rules (not only enabled ones): names whose supersession chain
// reaches a successor available on the installed tool are renamed to the
// terminal successor (a successor already listed dedupes the predecessor
// entry away), and names without a reachable successor are dropped — the
// entry would reference a linter the binary cannot name. Rules whose
// linter list the sweep emptied are dropped entirely: golangci-lint v2
// rejects single-criterion rules at load ("at least 2 of text, source,
// path, linters"), so a survived path-only rule would break the config the
// very next lint run. Pure.
func RenameDeprecatedInExclusionRules(
	observed *types.Config,
	tool types.GolangciVersion,
) (*types.Config, []Change) {
	normalized := observed.Clone()

	var changes []Change

	rules := normalized.Linters.Exclusions.Rules
	kept := make([]types.ExclusionRuleConfig, 0, len(rules))

	for _, rule := range rules {
		swept, ruleChanged, ruleChanges := sweepExclusionRuleLinters(rule, tool)
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
// rules emptied by the sweep are dropped entirely — a rule referencing no
// linter suppresses nothing, and golangci-lint v2's schema validation
// rejects single-criterion rules at load, so keeping the shell would break
// the config instead of just doing nothing.
func keepSweptRule(
	rule types.ExclusionRuleConfig,
	swept []string,
	ruleChanged bool,
) (types.ExclusionRuleConfig, bool) {
	if !ruleChanged {
		return rule, true
	}

	if len(swept) == 0 {
		return rule, false
	}

	rule.Linters = swept

	return rule, true
}

// sweepExclusionRuleLinters rewrites one rule's linter list: deprecated names
// with a successor become that successor (deduped against what the rule
// already lists), names without one vanish. The returned list is the swept
// linters; ruleChanged reports whether any deprecated name was touched.
func sweepExclusionRuleLinters(rule types.ExclusionRuleConfig, tool types.GolangciVersion) ([]string, bool, []Change) {
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

		var swept Change

		kept, swept = sweepDeprecatedEntry(entry, record, tool, kept, seen)
		changes = append(changes, swept)
	}

	return kept, changed, changes
}

// sweepDeprecatedEntry resolves one supersession-record entry inside a rule:
// renamed to the terminal successor when the chain is usable on the installed
// tool, dropped when it is not.
func sweepDeprecatedEntry(
	entry string,
	record types.LinterPolicy,
	tool types.GolangciVersion,
	kept []string,
	seen map[string]bool,
) ([]string, Change) {
	terminal, swappable := resolveSupersessionChain(record, tool, nil)
	if !swappable {
		return kept, sweptDropChange(entry, record)
	}

	return sweepRename(entry, terminal.Name, record.SupersessionReason, kept, seen)
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
	successor types.LinterName,
	reason string,
	kept []string,
	seen map[string]bool,
) ([]string, Change) {
	name := string(successor)

	if !seen[name] {
		seen[name] = true
		kept = append(kept, name)
	}

	return kept, Change{
		Path:    PathLintersExclusionRules,
		Element: entry + "->" + name,
		Kind:    ChangeUpdate,
		Reason:  reason,
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

// replaceNameUnique rewrites every from entry to to, except when to is
// already listed elsewhere — then from is simply dropped, so shared
// successors (deadcode, varcheck, structcheck all land on unused) can never
// produce duplicate enable entries.
func replaceNameUnique(names []types.LinterName, from, successor types.LinterName) []types.LinterName {
	out := names[:0:0]
	successorListed := slices.Contains(names, successor)

	for _, candidate := range names {
		if candidate == from && successorListed {
			continue
		}

		if candidate == from {
			out = append(out, successor)

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
