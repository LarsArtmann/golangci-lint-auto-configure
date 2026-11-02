// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"slices"
	"strings"
)

// normalizeLocalPrefixes converts local-prefixes from string to array format.
func normalizeLocalPrefixes(settings map[string]any) int {
	if localPrefixes, ok := settings["local-prefixes"]; ok {
		switch v := localPrefixes.(type) {
		case string:
			if v != "" {
				settings["local-prefixes"] = []string{v}

				return 1
			}

			delete(settings, "local-prefixes")

			return 1
		case nil:
			delete(settings, "local-prefixes")

			return 1
		}
	}

	return 0
}

// migrateLinterSettings applies all linter-specific settings migrations.
// Each step is a self-contained helper so the sequence stays auditable.
func migrateLinterSettings(settings map[string]any, rules *MigrationRules) int {
	fixes := 0

	fixes += removeDeprecatedLinterSettings(settings, rules)
	fixes += renameLegacyLinterSettings(settings, rules)
	fixes += removeSettingsForLintersWithoutSettings(settings, rules)
	fixes += removeDeprecatedGocriticCheckers(settings, rules)

	// goimports.local-prefixes: v1 allowed a bare string, v2 wants a list
	if goimports, ok := settings["goimports"].(map[string]any); ok {
		fixes += normalizeLocalPrefixes(goimports)
	}

	fixes += dropNullGosecExcludes(settings)
	fixes += migrateForbidigoPatternKey(settings)
	fixes += migrateModernizeDisableValues(settings, rules)
	fixes += migrateSloglintKeyNamingCase(settings, rules)

	// predeclared.ignore became a list in v2 (v1 allowed a bare string)
	if predeclared, ok := settings["predeclared"].(map[string]any); ok {
		fixes += normalizeStringToSlice(predeclared, "ignore")
	}

	fixes += migrateSloglintContextOnly(settings)
	fixes += migrateGovetCheckShadowing(settings)
	fixes += migrateGocriticRuleguardPaths(settings)

	return fixes
}

// removeDeprecatedLinterSettings deletes properties that no longer exist in
// the v2 schema, dropping the linter's block entirely when it empties out.
func removeDeprecatedLinterSettings(settings map[string]any, rules *MigrationRules) int {
	fixes := 0

	for linterName, propsToRemove := range rules.RemovedLinterSettings {
		if linterSettings, ok := settings[linterName].(map[string]any); ok {
			for _, prop := range propsToRemove {
				if _, exists := linterSettings[prop]; exists {
					delete(linterSettings, prop)

					fixes++
				}
			}

			if len(linterSettings) == 0 {
				delete(settings, linterName)
			}
		}
	}

	return fixes
}

// renameLegacyLinterSettings renames v1 keys whose v2 successor has identical
// semantics; an already-present successor key wins and the legacy value drops.
func renameLegacyLinterSettings(settings map[string]any, rules *MigrationRules) int {
	fixes := 0

	for linterName, renames := range rules.RenamedLinterSettings {
		linterSettings, ok := settings[linterName].(map[string]any)
		if !ok {
			continue
		}

		for oldKey, newKey := range renames {
			value, exists := linterSettings[oldKey]
			if !exists {
				continue
			}

			delete(linterSettings, oldKey)

			if _, hasNew := linterSettings[newKey]; !hasNew {
				linterSettings[newKey] = value
			}

			fixes++
		}

		if len(linterSettings) == 0 {
			delete(settings, linterName)
		}
	}

	return fixes
}

// removeSettingsForLintersWithoutSettings drops whole settings blocks for
// linters that expose no settings at all in v2.
func removeSettingsForLintersWithoutSettings(settings map[string]any, rules *MigrationRules) int {
	fixes := 0

	for linterName := range rules.LintersWithoutSettings {
		if _, exists := settings[linterName]; exists {
			delete(settings, linterName)

			fixes++
		}
	}

	return fixes
}

// removeDeprecatedGocriticCheckers deletes v1 gocritic checker settings that
// v2 no longer supports, removing the settings key when it empties out.
func removeDeprecatedGocriticCheckers(settings map[string]any, rules *MigrationRules) int {
	fixes := 0

	if gocritic, ok := settings["gocritic"].(map[string]any); ok {
		if gocriticSettings, ok := gocritic["settings"].(map[string]any); ok {
			for _, setting := range rules.GocriticSettingsToRemove {
				if _, exists := gocriticSettings[setting]; exists {
					delete(gocriticSettings, setting)

					fixes++
				}
			}

			if len(gocriticSettings) == 0 {
				delete(gocritic, "settings")
			}
		}
	}

	return fixes
}

// dropNullGosecExcludes removes a gosec.excludes key that only carries null.
func dropNullGosecExcludes(settings map[string]any) int {
	gosec, ok := settings["gosec"].(map[string]any)
	if !ok {
		return 0
	}

	if excludes, exists := gosec["excludes"]; exists && excludes == nil {
		delete(gosec, "excludes")

		return 1
	}

	return 0
}

// migrateForbidigoPatternKey renames the forbidigo.forbid item key p to
// pattern, the v2 name for the same matcher field.
func migrateForbidigoPatternKey(settings map[string]any) int {
	fixes := 0

	forbidigo, ok := settings["forbidigo"].(map[string]any)
	if !ok {
		return 0
	}

	forbid, ok := forbidigo["forbid"].([]any)
	if !ok {
		return 0
	}

	for i, item := range forbid {
		if itemMap, ok := item.(map[string]any); ok {
			if p, exists := itemMap["p"]; exists {
				itemMap["pattern"] = p
				delete(itemMap, "p")

				fixes++
			}
		}

		forbid[i] = item
	}

	return fixes
}

// migrateModernizeDisableValues maps v1 modernize.disable enum values onto
// their v2 spellings, passing unknown values through untouched.
func migrateModernizeDisableValues(settings map[string]any, rules *MigrationRules) int {
	modernize, ok := settings["modernize"].(map[string]any)
	if !ok {
		return 0
	}

	disable, ok := modernize["disable"].([]any)
	if !ok {
		return 0
	}

	mapped, changed := mapModernizeDisableValues(disable, rules)
	if changed > 0 {
		modernize["disable"] = mapped
	}

	return changed
}

// mapModernizeDisableValues rewrites one disable list, reporting the mapped
// list and how many enum values changed; non-string entries drop out, exactly
// as the v1→v2 carry always behaved.
func mapModernizeDisableValues(disable []any, rules *MigrationRules) ([]any, int) {
	mapped := make([]any, 0, len(disable))

	changed := 0

	for _, item := range disable {
		if str, isString := item.(string); isString {
			value, exists := rules.MapModernizeDisable(str)
			if !exists {
				value = str
			} else if value != str {
				changed++
			}

			mapped = append(mapped, value)
		}
	}

	return mapped, changed
}

// migrateSloglintKeyNamingCase maps the v1 sloglint.key-naming-case enum
// onto v2 spellings; unmappable values are dropped (v2 rejects them).
func migrateSloglintKeyNamingCase(settings map[string]any, rules *MigrationRules) int {
	sloglint, ok := settings["sloglint"].(map[string]any)
	if !ok {
		return 0
	}

	keyNamingCase, ok := sloglint["key-naming-case"].(string)
	if !ok {
		return 0
	}

	if mapped, exists := rules.MapSloglintKeyNamingCase(keyNamingCase); exists {
		if mapped != keyNamingCase {
			sloglint["key-naming-case"] = mapped

			return 1
		}

		return 0
	}

	delete(sloglint, "key-naming-case")

	return 1
}

// migrateSloglintContextOnly converts the v1 sloglint.context-only boolean
// into the v2 sloglint.context enum ("all"); an explicit context key wins.
func migrateSloglintContextOnly(settings map[string]any) int {
	sloglint, ok := settings["sloglint"].(map[string]any)
	if !ok {
		return 0
	}

	v, exists := sloglint["context-only"]
	if !exists {
		return 0
	}

	if truthy, isBool := v.(bool); isBool && truthy {
		if _, hasContext := sloglint["context"]; !hasContext {
			sloglint["context"] = "all"
		}
	}

	delete(sloglint, "context-only")

	return 1
}

// migrateGovetCheckShadowing replaces the removed govet.check-shadowing flag
// with the "shadow" analyzer in govet.enable, deduplicating an existing entry.
func migrateGovetCheckShadowing(settings map[string]any) int {
	govet, ok := settings["govet"].(map[string]any)
	if !ok {
		return 0
	}

	v, exists := govet["check-shadowing"]
	if !exists {
		return 0
	}

	if truthy, isBool := v.(bool); isBool && truthy {
		enable, _ := govet["enable"].([]any)
		if !slices.Contains(enable, "shadow") {
			govet["enable"] = append(enable, "shadow")
		}
	}

	delete(govet, "check-shadowing")

	return 1
}

// migrateGocriticRuleguardPaths rewrites the ${configDir} variable in gocritic
// ruleguard paths to its v2 name ${base-path}.
func migrateGocriticRuleguardPaths(settings map[string]any) int {
	gocritic, ok := settings["gocritic"].(map[string]any)
	if !ok {
		return 0
	}

	gocriticSettings, ok := gocritic["settings"].(map[string]any)
	if !ok {
		return 0
	}

	ruleguard, ok := gocriticSettings["ruleguard"].(map[string]any)
	if !ok {
		return 0
	}

	rules, ok := ruleguard["rules"].(string)
	if !ok || !strings.Contains(rules, "${configDir}") {
		return 0
	}

	ruleguard["rules"] = strings.ReplaceAll(rules, "${configDir}", "${base-path}")

	return 1
}

// normalizeStringToSlice converts a bare string value at key into a one-element
// string list, mirroring v2 settings keys that became arrays (predeclared.ignore).
func normalizeStringToSlice(settings map[string]any, key string) int {
	value, ok := settings[key]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case string:
		if v != "" {
			settings[key] = []string{v}

			return 1
		}

		delete(settings, key)

		return 1
	case nil:
		delete(settings, key)

		return 1
	}

	return 0
}

// migrateFormattersSettings applies migrations to formatters.settings.
func migrateFormattersSettings(formatters map[string]any) int {
	fixes := 0

	// 1. gci: remove skip-generated
	if gci, ok := formatters["gci"].(map[string]any); ok {
		if _, exists := gci["skip-generated"]; exists {
			delete(gci, "skip-generated")

			fixes++
		}
	}

	// 2. goimports: convert local-prefixes from string to array
	if goimports, ok := formatters["goimports"].(map[string]any); ok {
		fixes += normalizeLocalPrefixes(goimports)
	}

	return fixes
}
