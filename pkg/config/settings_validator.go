package config

import (
	"fmt"
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ValidateSettingsKeys checks linter settings keys against known linters and
// their expected settings schema. Returns a list of warnings (not errors) for
// unrecognized linter names or unknown settings sub-keys. Warnings are soft —
// they don't block config loading, since newer golangci-lint versions may add
// new settings that this tool doesn't know about yet.
func ValidateSettingsKeys(config *types.Config) []string {
	var warnings []string

	for linterName, rawSettings := range config.Linters.Settings {
		warning := validateSingleLinterSettings(linterName, rawSettings)
		warnings = append(warnings, warning...)
	}

	slices.Sort(warnings)

	return warnings
}

func validateSingleLinterSettings(linterName string, rawSettings any) []string {
	_, isKnown := constants.LinterPriorities[types.LinterName(linterName)]
	if !isKnown {
		return []string{
			fmt.Sprintf(
				"settings for %q: linter not in known priorities (may be a new linter)",
				linterName,
			),
		}
	}

	settingsMap, ok := rawSettings.(map[string]any)
	if !ok {
		return nil
	}

	defaultSettings, hasDefaults := constants.DefaultLinterSettings[types.LinterName(linterName)]
	if !hasDefaults {
		return nil
	}

	return findUnknownSettingsKeys(linterName, settingsMap, defaultSettings.ToMap())
}

func findUnknownSettingsKeys(linterName string, userKeys, expectedKeys map[string]any) []string {
	var unknown []string

	for key := range userKeys {
		if _, ok := expectedKeys[key]; !ok {
			unknown = append(unknown, key)
		}
	}

	slices.Sort(unknown)

	validKeys := sortedKeys(expectedKeys)
	warnings := make([]string, 0, len(unknown))

	for _, key := range unknown {
		warnings = append(warnings, fmt.Sprintf(
			"settings for %q: unknown key %q (valid: %v)",
			linterName, key, validKeys,
		))
	}

	return warnings
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	return keys
}
