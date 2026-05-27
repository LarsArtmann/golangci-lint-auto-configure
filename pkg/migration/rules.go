// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// MigrationRules contains all the rules and mappings needed for configuration migration.
//

type MigrationRules struct {
	ValidVersions                 types.Set[string]
	RemovedLinterSettings         map[string][]string
	ModernizeDisableMappings      map[string]string
	SloglintKeyNamingCaseMappings map[string]string
	GocriticSettingsToRemove      []string
	LintersWithoutSettings        types.Set[string]
}

// DefaultRules returns a MigrationRules populated with the default golangci-lint v2 rules.
func DefaultRules() *MigrationRules {
	return &MigrationRules{
		ValidVersions:                 validVersions(),
		RemovedLinterSettings:         removedLinterSettings(),
		ModernizeDisableMappings:      modernizeDisableMappings(),
		SloglintKeyNamingCaseMappings: sloglintKeyNamingCaseMappings(),
		GocriticSettingsToRemove:      gocriticSettingsToRemove(),
		LintersWithoutSettings:        lintersWithoutSettings(),
	}
}

func validVersions() types.Set[string] {
	return types.NewSet("2", "2.8", "2.8.0", "2.9", "2.9.0", "2.10", "2.10.0", "2.10.1")
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
		"wrapcheck":      {"ignoreSigRegexps", "ignoreSigs", "ignorePackageGlobs"},
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
	return r.ValidVersions.Contains(version)
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
