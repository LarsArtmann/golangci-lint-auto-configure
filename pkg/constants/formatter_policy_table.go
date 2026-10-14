package constants

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// FormatterPolicies is the formatter policy table: one record per known
// formatter. Name repetition across records (supersession + redundancy
// references) is curated data, not extractable strings.
//
// of rendering, not of a single formatter); core-stack membership, priority,
// redundancy, tech gates, and settings live here.
//
// gci is the motivating non-core record: it left CoreFormatters on
// 2026-10-11 (too many projects rejected its import-group reordering) but
// stays a valid manual choice, so it keeps its table record and its
// FormatterOrder slot.
//
//nolint:goconst // curated table literals, names repeat by design The canonical render order lives in FormatterOrder (a property
var FormatterPolicies = []types.FormatterPolicy{
	{
		Name:        "gci",
		Priority:    types.FormatterPriorityMedium,
		Rationale:   "Organizes import statements with custom section rules",
		Description: "Organizes import statements with additional rules",
		AutoFix:     true,
	},
	{
		Name:         "gofmt",
		Priority:     types.FormatterPriorityMedium,
		Rationale:    "Standard Go code formatting (consider gofumpt for stricter formatting)",
		Description:  "Standard Go code formatting",
		AutoFix:      true,
		SupersededBy: "gofumpt",
	},
	{
		Name:        "gofumpt",
		Priority:    types.FormatterPriorityHigh,
		Rationale:   "Enhanced Go formatting with stricter rules than gofmt",
		Description: "Enhanced Go formatting with stricter rules",
		AutoFix:     true,
		Core:        true,
	},
	{
		Name:        "goimports",
		Priority:    types.FormatterPriorityMedium,
		Rationale:   "Formats code and automatically manages import statements",
		Description: "Formats code and manages import statements",
		AutoFix:     true,
		Core:        true,
	},
	{
		Name:        "golines",
		Priority:    types.FormatterPriorityHigh,
		Rationale:   "Formats code and fixes long lines by breaking them",
		Description: "Formats code and fixes long lines",
		AutoFix:     true,
		Core:        true,
		Settings: []types.VersionedSettings{
			{Settings: GolinesFormatterSettings{MaxLen: 120}}, //nolint:mnd // intentional default line length
		},
	},
	{
		Name:        "swaggo",
		Priority:    types.FormatterPriorityLow,
		Rationale:   "Formats Swagger/OpenAPI documentation comments",
		Description: "Formats Swagger/OpenAPI documentation comments",
		AutoFix:     true,
		Gate:        types.RequiresTech("swaggo"),
	},
}

// FormatterPolicyRecord returns the table record for a formatter, reporting
// whether the formatter is known to the table.
func FormatterPolicyRecord(name types.FormatterName) (types.FormatterPolicy, bool) {
	for _, record := range FormatterPolicies {
		if record.Name == name {
			return record, true
		}
	}

	return types.FormatterPolicy{}, false
}

// CoreFormattersInOrder returns the core-stack formatters in canonical render
// order (FormatterOrder filtered to Core records). The renderer and the
// presets both consume this instead of a stored slice.
func CoreFormattersInOrder() []types.FormatterName {
	var names []types.FormatterName

	for _, name := range FormatterOrder {
		if record, known := FormatterPolicyRecord(name); known && record.Core {
			names = append(names, name)
		}
	}

	return names
}
