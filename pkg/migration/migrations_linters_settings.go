// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

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
//
//nolint:gocognit,nestif,gocyclo,varnamelen,cyclop,funlen // Complex migration logic with many linter-specific cases
func migrateLinterSettings(settings map[string]any, rules *MigrationRules) int {
	fixes := 0

	// 1. Remove deprecated properties from known linters
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

	// 2. Remove settings for linters that no longer support them
	for _, linterName := range rules.LintersWithoutSettings {
		if _, exists := settings[linterName]; exists {
			delete(settings, linterName)

			fixes++
		}
	}

	// 3. Handle gocritic.settings - remove old checker settings
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

	// 4. Convert goimports.local-prefixes from string to array
	if goimports, ok := settings["goimports"].(map[string]any); ok {
		fixes += normalizeLocalPrefixes(goimports)
	}

	// 5. Convert gosec.excludes from null to array (or remove)
	if gosec, ok := settings["gosec"].(map[string]any); ok {
		if excludes, exists := gosec["excludes"]; exists {
			if excludes == nil {
				delete(gosec, "excludes")

				fixes++
			}
		}
	}

	// 6. Migrate forbidigo.forbid field rename: p -> pattern
	if forbidigo, ok := settings["forbidigo"].(map[string]any); ok {
		if forbid, ok := forbidigo["forbid"].([]any); ok {
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
		}
	}

	// 7. Fix modernize.disable enum values
	if modernize, ok := settings["modernize"].(map[string]any); ok {
		if disable, ok := modernize["disable"].([]any); ok {
			var newDisable []any

			for _, item := range disable {
				if str, ok := item.(string); ok {
					if mapped, exists := rules.MapModernizeDisable(str); exists {
						if mapped != str {
							fixes++
						}

						newDisable = append(newDisable, mapped)
					} else {
						newDisable = append(newDisable, str)
					}
				}
			}

			if fixes > 0 {
				modernize["disable"] = newDisable
			}
		}
	}

	// 8. Fix sloglint.key-naming-case enum values
	if sloglint, ok := settings["sloglint"].(map[string]any); ok {
		if keyNamingCase, ok := sloglint["key-naming-case"].(string); ok {
			if mapped, exists := rules.MapSloglintKeyNamingCase(keyNamingCase); exists {
				if mapped != keyNamingCase {
					sloglint["key-naming-case"] = mapped
					fixes++
				}
			} else {
				delete(sloglint, "key-naming-case")

				fixes++
			}
		}
	}

	return fixes
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
