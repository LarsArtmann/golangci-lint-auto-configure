package linter

import (
	"slices"
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Normalizers are pure functions observed -> (normalized, changes). They
// repair states that make a config UNLOADABLE or that this tool itself
// historically broke; everything else is merge/strategy territory. Counting
// disappears: the returned changes ARE the count, the audit trail, and the
// dry-run report at once.

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

		keyMap := constants.KnownBadSettingsKeys[linterName]

		for _, badKey := range sortedKeys(keyMap) {
			badValue, hasBad := settings[badKey]
			if !hasBad {
				continue
			}

			goodKey := keyMap[badKey]

			delete(settings, badKey)

			change := Change{
				Path: "linters.settings." + linterName, Element: badKey,
				Kind: ChangeRemove, From: badValue,
				Reason: "schema-invalid key",
			}

			if _, hasGood := settings[goodKey]; !hasGood {
				settings[goodKey] = badValue

				change = Change{
					Path: "linters.settings." + linterName, Element: badKey + "->" + goodKey,
					Kind: ChangeUpdate, From: badValue, To: badValue,
					Reason: "schema-invalid key renamed",
				}
			}

			changes = append(changes, change)
		}
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
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

		removeChange := Change{
			Path: "linters.enable", Element: string(name), Kind: ChangeRemove,
			From: string(name), Reason: record.SupersessionReason,
		}

		if record.SupersededBy == "" {
			normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)

			changes = append(changes, removeChange)

			continue
		}

		if neverEnable != nil && neverEnable(record.SupersededBy) {
			normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)
			removeChange.Reason += " (replacement under never-enable, NOT added)"

			changes = append(changes, removeChange)

			continue
		}

		if tool.IsZero() || record.SupersessionSince.IsZero() || tool.AtLeast(record.SupersessionSince) {
			normalized.Linters.Enable = replaceName(normalized.Linters.Enable, name, record.SupersededBy)

			changes = append(changes, removeChange)
			changes = append(changes, Change{
				Path: "linters.enable", Element: string(record.SupersededBy), Kind: ChangeAdd,
				To: string(record.SupersededBy), Reason: record.SupersessionReason,
			})

			changes = append(changes, migrateDeprecatedSettings(normalized, name, record.SupersededBy)...)
			changes = append(changes, replaceExclusionLinters(normalized, name, record.SupersededBy)...)
		} else {
			// Replacement does not exist in the installed version: drop the
			// deprecated linter without adding its successor.
			normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)

			removeChange.Reason += " (replacement needs " + record.SupersessionSince.String() + ")"

			changes = append(changes, removeChange)
		}
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// NormalizeConfigShape repairs the structural facts golangci-lint refuses to
// load: a missing/wrong config version, an invalid run timeout, and the
// removed typecheck meta-linter, plus the tool-disabled linter moves. Pure.
func NormalizeConfigShape(observed *types.Config) (*types.Config, []Change) {
	normalized := observed.Clone()

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
			Path: "run.timeout", Kind: ChangeUpdate,
			From:   normalized.Run.Timeout,
			To:     constants.DefaultTimeout,
			Reason: durationReason,
		})

		normalized.Run.Timeout = constants.DefaultTimeout
	}

	for _, list := range []struct {
		path  string
		field *[]types.LinterName
	}{
		{"linters.enable", &normalized.Linters.Enable},
		{"linters.disable", &normalized.Linters.Disable},
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

	// Tool-level disabled linters (funcorder, noinlineerr): forcibly moved to
	// linters.disable and removed from enable on every run.
	for _, name := range constants.LintersInTier(types.TierDisabled) {
		if !containsLinterName(toStrings(normalized.Linters.Enable), string(name)) {
			continue
		}

		normalized.Linters.Enable = removeName(normalized.Linters.Enable, name)

		reason := "tool-level disabled"
		if record, known := constants.PolicyRecord(name); known && record.Tier.Rationale != "" {
			reason = record.Tier.Rationale
		}

		changes = append(changes,
			Change{
				Path: "linters.enable", Element: string(name), Kind: ChangeRemove,
				From: string(name), Reason: reason,
			},
			Change{
				Path: "linters.disable", Element: string(name), Kind: ChangeAdd,
				To: string(name), Reason: reason,
			},
		)

		if !containsLinterName(toStrings(normalized.Linters.Disable), string(name)) {
			normalized.Linters.Disable = append(normalized.Linters.Disable, name)
		}
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// PruneOrphanSettings removes settings blocks for linters that are neither
// enabled nor disabled (dead weight; the legacy prune pass). Pure.
func PruneOrphanSettings(observed *types.Config) (*types.Config, []Change) {
	if len(observed.Linters.Settings) == 0 {
		return observed, nil
	}

	enabled := types.NewSet(observed.Linters.Enable...)
	disabled := types.NewSet(observed.Linters.Disable...)

	var changes []Change

	normalized := observed.Clone()

	for _, name := range sortedSettingNames(observed.Linters.Settings) {
		if enabled.Contains(types.LinterName(name)) || disabled.Contains(types.LinterName(name)) {
			continue
		}

		delete(normalized.Linters.Settings, name)

		changes = append(changes, Change{
			Path: "linters.settings", Element: name, Kind: ChangeRemove,
			Reason: "orphaned settings block (linter neither enabled nor disabled)",
		})
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

// RemoveRedundancies drops linters and formatters superseded by an enabled
// counterpart: lll dies with golines, gofmt dies with gofumpt. Pure.
func RemoveRedundancies(observed *types.Config, desiredFormatters []types.FormatterName) (*types.Config, []Change) {
	normalized := observed.Clone()

	var changes []Change

	formatters := types.NewSet(desiredFormatters...)
	for _, name := range normalized.Formatters.Enable {
		formatters.Add(name)
	}

	if formatters.Contains("golines") && containsLinterName(toStrings(normalized.Linters.Enable), "lll") {
		normalized.Linters.Enable = removeName(normalized.Linters.Enable, "lll")

		changes = append(changes, Change{
			Path: "linters.enable", Element: "lll", Kind: ChangeRemove,
			From:   "lll",
			Reason: "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
		})
	}

	if formatters.Contains("gofumpt") && formatters.Contains("gofmt") {
		normalized.Formatters.Enable = removeFormatterName(normalized.Formatters.Enable, "gofmt")

		changes = append(changes, Change{
			Path: "formatters.enable", Element: "gofmt", Kind: ChangeRemove,
			From:   "gofmt",
			Reason: "redundant when gofumpt is enabled (gofumpt is the stricter superset)",
		})
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
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

		if !ruleChanged {
			kept = append(kept, rule)

			continue
		}

		if len(swept) == 0 && (rule.Text != "" || rule.Source != "") {
			continue
		}

		rule.Linters = swept
		kept = append(kept, rule)
	}

	if len(changes) == 0 {
		return observed, nil
	}

	normalized.Linters.Exclusions.Rules = kept

	return normalized, changes
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
			if !seen[entry] {
				seen[entry] = true
				kept = append(kept, entry)
			}

			continue
		}

		changed = true

		if record.SupersededBy == "" {
			changes = append(changes, Change{
				Path: "linters.exclusions.rules", Element: entry, Kind: ChangeRemove,
				From:   entry,
				Reason: record.SupersessionReason,
			})

			continue
		}

		successor := string(record.SupersededBy)

		changes = append(changes, Change{
			Path:    "linters.exclusions.rules",
			Element: entry + "->" + successor,
			Kind:    ChangeUpdate,
			Reason:  record.SupersessionReason,
		})

		if !seen[successor] {
			seen[successor] = true
			kept = append(kept, successor)
		}
	}

	return kept, changed, changes
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
func migrateDeprecatedSettings(cfg *types.Config, from, to types.LinterName) []Change {
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

	if _, hasSuccessor := cfg.Linters.Settings[string(to)]; !hasSuccessor {
		// Store the RAW map type: SettingsMap is a defined type and would
		// fail every map[string]any assertion downstream.
		cfg.Linters.Settings[string(to)] = map[string]any(settings.Clone())
	}

	return []Change{{
		Path: "linters.settings", Element: string(from) + "->" + string(to),
		Kind: ChangeUpdate, From: settings.Clone(), To: settings.Clone(),
		Reason: "settings block migrated to successor",
	}}
}

// replaceExclusionLinters rewrites a deprecated linter name inside
// linters.exclusions.rules[].linters; names without a successor are dropped.
func replaceExclusionLinters(cfg *types.Config, from, to types.LinterName) []Change {
	var changes []Change

	for i := range cfg.Linters.Exclusions.Rules {
		rule := &cfg.Linters.Exclusions.Rules[i]

		if !containsLinterName(rule.Linters, string(from)) {
			continue
		}

		rule.Linters = replaceStringEntry(rule.Linters, string(from), string(to))

		changes = append(changes, Change{
			Path: "linters.exclusions.rules", Element: string(from) + "->" + string(to),
			Kind: ChangeUpdate, Reason: "exclusion rule follows the replacement",
		})
	}

	return changes
}

func containsLinterName(names []string, want string) bool {
	return slices.Contains(names, want)
}

func replaceStringEntry(names []string, from, to string) []string {
	out := names[:0:0]

	for _, name := range names {
		if name == from {
			if to != "" {
				out = append(out, to)
			}

			continue
		}

		out = append(out, name)
	}

	return out
}
