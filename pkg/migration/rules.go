// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// MigrationRules contains all the rules and mappings needed for configuration migration.
//

type MigrationRules struct {
	RemovedLinterSettings map[string][]string
	// RenamedLinterSettings maps linter name -> old settings key -> v2 key,
	// for renames whose semantics are unchanged (wrapcheck's v1 camelCase
	// keys became kebab-case in v2 with identical meaning). Renaming keeps
	// the user's intent; the old RemovedLinterSettings drop silently lost it.
	RenamedLinterSettings         map[string]map[string]string
	ModernizeDisableMappings      map[string]string
	SloglintKeyNamingCaseMappings map[string]string
	GocriticSettingsToRemove      []string
	LintersWithoutSettings        types.Set[string]
}

// DefaultRules returns a MigrationRules populated with the default golangci-lint v2 rules.
func DefaultRules() *MigrationRules {
	return &MigrationRules{
		RemovedLinterSettings:         removedLinterSettings(),
		RenamedLinterSettings:         renamedLinterSettings(),
		ModernizeDisableMappings:      modernizeDisableMappings(),
		SloglintKeyNamingCaseMappings: sloglintKeyNamingCaseMappings(),
		GocriticSettingsToRemove:      gocriticSettingsToRemove(),
		LintersWithoutSettings:        lintersWithoutSettings(),
	}
}

func removedLinterSettings() map[string][]string {
	return map[string][]string{
		"cyclop":         {"skip-tests"},
		"exhaustive":     {"check-generated"},
		"fatcontext":     {"ignore-len"},
		"gci":            {"skip-generated"},
		"gochecksumtype": {"exhaustive"},
		"gomodguard":     {"local-replace-directives"},
		"nolintlint":     {"allow-leading-space"},
		"tagliatelle":    {"use-field-name"},
		"unused":         {"check-exported"},
	}
}

// renamedLinterSettings covers the v1 keys the migration guide renames with
// unchanged semantics. wrapcheck's camelCase keys all became kebab-case;
// ignoreInterfaceRegexps was missing from earlier versions of this table,
// which silently dropped that part of a user's wrapcheck configuration.
func renamedLinterSettings() map[string]map[string]string {
	return map[string]map[string]string{
		"wrapcheck": {
			"ignoreSigs":             "ignore-sigs",
			"ignoreSigRegexps":       "ignore-sig-regexps",
			"ignorePackageGlobs":     "ignore-package-globs",
			"ignoreInterfaceRegexps": "ignore-interface-regexps",
		},
	}
}

func gocriticSettingsToRemove() []string {
	return []string{
		"appendAssign",
		"dogsled",
		"errcheck",
		"hexLiteral",
		"paramTypeCombine",
		"unimport",
	}
}

func lintersWithoutSettings() types.Set[string] {
	return types.NewSet(
		"containedctx",
		"contextcheck",
		"gochecknoglobals",
		"gochecknoinits",
		"intrange",
		"mirror",
		"nilnesserr",
		"stylecheck",
		"testableexamples",
		"wastedassign",
		"zerologlint",
	)
}

func modernizeDisableMappings() map[string]string {
	return map[string]string{
		"fmtappendf":      "fmtappendf",
		"forvar":          "forvar",
		"mapsloop":        "mapsloop",
		"mapsdotdelete":   "mapsdotdelete",
		"minmax":          "minmax",
		"newexpr":         "newexpr",
		"omitzero":        "omitzero",
		"plusbuild":       "plusbuild",
		"rangeint":        "rangeint",
		"reflecttypefor":  "reflecttypefor",
		"slicesappend":    "slicesappend",
		"slicesclone":     "slicesclone",
		"slicescompact":   "slicescompact",
		"slicescontains":  "slicescontains",
		"slicesdelete":    "slicesdelete",
		"slicesgrows":     "slicesgrows",
		"slicesloop":      "mapsloop",
		"slicessort":      "slicessort",
		"stditer":         "stditerators",
		"stringscompare":  "stringscompare",
		"typesforversion": "typesforversion",
		"waitgroup":       "waitgroup",
	}
}

func sloglintKeyNamingCaseMappings() map[string]string {
	return map[string]string{
		"snake":      "snake",
		"kebab":      "kebab",
		"camel":      "camel",
		"pascal":     "pascal",
		"camelCase":  "camel",
		"PascalCase": "pascal",
		"snake_case": "snake",
		"kebab-case": "kebab",
	}
}

// IsValidVersion returns true if the given version string is valid for v2.
func (r *MigrationRules) IsValidVersion(version string) bool {
	return version == "2" || strings.HasPrefix(version, "2.")
}

// GetDeprecatedProperties returns the deprecated properties for a given linter.
func (r *MigrationRules) GetDeprecatedProperties(linterName string) []string {
	return r.RemovedLinterSettings[linterName]
}

// IsLinterWithoutSettings returns true if the linter doesn't support settings.
func (r *MigrationRules) IsLinterWithoutSettings(linterName string) bool {
	return r.LintersWithoutSettings.Contains(linterName)
}

// MapModernizeDisable maps old modernize disable values to new valid values.
func (r *MigrationRules) MapModernizeDisable(oldValue string) (string, bool) {
	mapped, exists := r.ModernizeDisableMappings[oldValue]

	return mapped, exists
}

// MapSloglintKeyNamingCase maps old sloglint key-naming-case values to valid ones.
func (r *MigrationRules) MapSloglintKeyNamingCase(oldValue string) (string, bool) {
	mapped, exists := r.SloglintKeyNamingCaseMappings[oldValue]

	return mapped, exists
}
