package constants

import (
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// RedundancyKind is the closed union of redundancy relations: which list the
// superseded entry is dropped from when the superseder is enabled.
type RedundancyKind uint8

const (
	// RedundancyLinter marks a linter superseded by an enabled formatter
	// (lll dies with golines: the formatter fixes what the linter reports).
	RedundancyLinter RedundancyKind = iota
	// RedundancyFormatter marks a formatter superseded by a stricter
	// formatter (gofmt dies with gofumpt).
	RedundancyFormatter
)

// String renders the kind for audit reasons and specs.
func (k RedundancyKind) String() string {
	if k == RedundancyLinter {
		return "linter"
	}

	return "formatter"
}

// RedundancyRelation is one typed supersession edge: when Superseder is (or
// will be) enabled, the engine drops Superseded from the list the kind
// names. Relations live HERE, not hardcoded in the engine — adding a
// relation is a data edit, and the integrity specs make dangling targets a
// test failure instead of a silent no-op.
type RedundancyRelation struct {
	// Superseded is the entry dropped from the enable list.
	Superseded types.LinterName
	// Superseder is the entry whose presence makes Superseded redundant.
	Superseder types.FormatterName
	// Kind selects the list Superseded is dropped from.
	Kind RedundancyKind
	// Reason is the human-readable audit/report explanation.
	Reason string
}

// RedundancyRelations is the curated relation table. The names repeat the
// policy tables' literals by design (they are the relation's subject).
//
//nolint:goconst // curated name literals, duplicates across tables are intentional
var RedundancyRelations = []RedundancyRelation{
	{
		Superseded: "lll",
		Superseder: "golines",
		Kind:       RedundancyLinter,
		Reason:     "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
	},
	{
		Superseded: "gofmt",
		Superseder: "gofumpt",
		Kind:       RedundancyFormatter,
		Reason:     "redundant when gofumpt is enabled (gofumpt is the stricter superset)",
	},
}

// RedundanciesFor returns every relation that supersedes the given entry.
func RedundanciesFor(superseded types.LinterName) []RedundancyRelation {
	var found []RedundancyRelation

	for _, relation := range RedundancyRelations {
		if relation.Superseded == superseded {
			found = append(found, relation)
		}
	}

	return found
}

// ValidateProblem returns a human-readable integrity violation, or "" when
// the relation is sound. Superseded entries must exist in the linter table
// (except formatter-on-formatter relations, which must exist in the
// formatter table), and superseders must exist in the formatter table. The
// init-time check panics on any problem — curated data, not a runtime
// condition.
func (r RedundancyRelation) ValidateProblem(
	linters map[types.LinterName]bool,
	formatters map[types.FormatterName]bool,
) string {
	if !formatters[r.Superseder] {
		return fmt.Sprintf("redundancy relation %q superseded by unknown formatter %q", r.Superseded, r.Superseder)
	}

	if r.Kind == RedundancyFormatter {
		if !formatters[types.FormatterName(r.Superseded)] {
			return fmt.Sprintf("formatter redundancy %q not in the formatter table", r.Superseded)
		}

		return ""
	}

	if !linters[r.Superseded] {
		return fmt.Sprintf("linter redundancy %q not in the linter table", r.Superseded)
	}

	return ""
}
