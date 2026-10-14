package linter

import (
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
func ReplaceDeprecatedLinters(observed *types.Config, tool types.GolangciVersion, neverEnable func(types.LinterName) bool) (*types.Config, []Change) {
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
// renamed, names without one are dropped. Pure.
func RenameDeprecatedInExclusionRules(observed *types.Config) (*types.Config, []Change) {
	var changes []Change

	normalized := observed.Clone()

	for i := range normalized.Linters.Exclusions.Rules {
		rule := &normalized.Linters.Exclusions.Rules[i]

		for _, entry := range append([]string{}, rule.Linters...) {
			record, known := constants.PolicyRecord(types.LinterName(entry))
			if !known || record.SupersessionReason == "" {
				continue
			}

			if record.SupersededBy != "" {
				rule.Linters = replaceStringEntry(rule.Linters, entry, string(record.SupersededBy))

				changes = append(changes, Change{
					Path: "linters.exclusions.rules",
					Element: entry + "->" + string(record.SupersededBy),
					Kind:    ChangeUpdate,
					Reason:  record.SupersessionReason,
				})
			} else {
				rule.Linters = replaceStringEntry(rule.Linters, entry, "")

				changes = append(changes, Change{
					Path: "linters.exclusions.rules", Element: entry, Kind: ChangeRemove,
					From:   entry,
					Reason: record.SupersessionReason,
				})
			}
		}
	}

	if len(changes) == 0 {
		return observed, nil
	}

	return normalized, changes
}

func deprecatedIn(enabled []types.LinterName) []types.LinterName {
	var deprecated []types.LinterName

	for _, name := range enabled {
		if record, known := constants.PolicyRecord(name); known && record.SupersessionReason != "" {
			deprecated = append(deprecated, name)
		}
	}

	sort.Slice(deprecated, func(i, j int) bool { return deprecated[i] < deprecated[j] })

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
		cfg.Linters.Settings[string(to)] = settings.Clone()
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
	for _, name := range names {
		if name == want {
			return true
		}
	}

	return false
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
