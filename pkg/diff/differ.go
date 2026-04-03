package diff

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Change represents a single change between two configs.
type Change struct {
	Type        ChangeType
	Path        string
	OldValue    string
	NewValue    string
	Description string
}

// ChangeType indicates the type of change.
type ChangeType int

const (
	ChangeTypeAdded ChangeType = iota
	ChangeTypeRemoved
	ChangeTypeModified
)

func (c ChangeType) String() string {
	switch c {
	case ChangeTypeAdded:
		return "ADDED"
	case ChangeTypeRemoved:
		return "REMOVED"
	case ChangeTypeModified:
		return "MODIFIED"
	default:
		return "UNKNOWN"
	}
}

// Differ compares two configurations and returns the differences.
type Differ struct{}

// NewDiffer creates a new config differ.
func NewDiffer() *Differ {
	return &Differ{}
}

// Compare compares two configs and returns the changes.
func (d *Differ) Compare(old, newConfig *types.Config) []Change {
	var changes []Change

	// Compare version
	if old.Version != newConfig.Version {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "version",
			OldValue:    old.Version,
			NewValue:    newConfig.Version,
			Description: fmt.Sprintf("Version changed from %s to %s", old.Version, newConfig.Version),
		})
	}

	// Compare run settings
	changes = append(changes, d.compareRunSettings(old.Run, newConfig.Run)...)

	// Compare linters
	changes = append(changes, d.compareLinters(old.Linters, newConfig.Linters)...)

	// Compare formatters
	changes = append(changes, d.compareFormatters(old.Formatters, newConfig.Formatters)...)

	return changes
}

func (d *Differ) compareRunSettings(old, newConfig types.RunConfig) []Change {
	var changes []Change

	changes = d.addChangeIfDifferent(changes, old.Timeout, newConfig.Timeout, "run.timeout",
		"Timeout", func(o, n string) string { return fmt.Sprintf("Timeout changed from %s to %s", o, n) })
	changes = d.addChangeIfDifferent(changes, old.Go, newConfig.Go, "run.go",
		"Go version", func(o, n string) string { return fmt.Sprintf("Go version changed from %s to %s", o, n) })
	changes = d.addTestChangeIfDifferent(changes, old.Tests, newConfig.Tests)

	return changes
}

func (d *Differ) addChangeIfDifferent(
	changes []Change,
	oldVal, newVal string,
	path, label string,
	formatFunc func(string, string) string,
) []Change {
	if oldVal != newVal {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        path,
			OldValue:    oldVal,
			NewValue:    newVal,
			Description: formatFunc(oldVal, newVal),
		})
	}

	return changes
}

func (d *Differ) addTestChangeIfDifferent(changes []Change, oldTests, newTests bool) []Change {
	if oldTests != newTests {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.tests",
			OldValue:    strconv.FormatBool(oldTests),
			NewValue:    strconv.FormatBool(newTests),
			Description: fmt.Sprintf("Tests changed from %v to %v", oldTests, newTests),
		})
	}

	return changes
}

func (d *Differ) compareEnabled(oldEnable, newEnable []string, pathPrefix, entityName string) []Change {
	oldEnabled := makeStringSet(oldEnable)
	newEnabled := makeStringSet(newEnable)

	changes := d.findAddedItems(oldEnabled, newEnabled, pathPrefix, entityName)
	changes = d.findRemovedItems(oldEnabled, newEnabled, pathPrefix, entityName, changes)

	return changes
}

func makeStringSet(items []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range items {
		set[item] = true
	}

	return set
}

func (d *Differ) findAddedItems(
	oldEnabled, newEnabled map[string]bool,
	pathPrefix, entityName string,
) []Change {
	var changes []Change

	for item := range newEnabled {
		if !oldEnabled[item] {
			changes = append(changes, Change{
				Type:        ChangeTypeAdded,
				Path:        fmt.Sprintf("%s.enable.%s", pathPrefix, item),
				OldValue:    "",
				NewValue:    item,
				Description: fmt.Sprintf("Enabled %s: %s", entityName, item),
			})
		}
	}

	return changes
}

func (d *Differ) findRemovedItems(
	oldEnabled, newEnabled map[string]bool,
	pathPrefix, entityName string,
	changes []Change,
) []Change {
	for item := range oldEnabled {
		if !newEnabled[item] {
			changes = append(changes, Change{
				Type:        ChangeTypeRemoved,
				Path:        fmt.Sprintf("%s.enable.%s", pathPrefix, item),
				OldValue:    item,
				NewValue:    "",
				Description: fmt.Sprintf("Disabled %s: %s", entityName, item),
			})
		}
	}

	return changes
}

func (d *Differ) compareLinters(old, newCfg types.LintersConfig) []Change {
	return d.compareEnabled(old.Enable, newCfg.Enable, "linters", "linter")
}

func (d *Differ) compareFormatters(old, newCfg types.FormattersConfig) []Change {
	return d.compareEnabled(old.Enable, newCfg.Enable, "formatters", "formatter")
}

// FormatChanges formats changes as a human-readable string.
func (d *Differ) FormatChanges(changes []Change) string {
	if len(changes) == 0 {
		return "No changes detected"
	}

	added, removed, modified := countChangesByType(changes)
	builder := formatChangeHeader(added, removed, modified)
	sortedChanges := sortChangesByPath(changes)

	return formatChangeDetails(builder.String(), sortedChanges)
}

func countChangesByType(changes []Change) (added, removed, modified int) {
	for _, change := range changes {
		switch change.Type {
		case ChangeTypeAdded:
			added++
		case ChangeTypeRemoved:
			removed++
		case ChangeTypeModified:
			modified++
		}
	}

	return added, removed, modified
}

func formatChangeHeader(added, removed, modified int) *strings.Builder {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Changes: %d added, %d removed, %d modified\n\n", added, removed, modified)

	return &builder
}

func sortChangesByPath(changes []Change) []Change {
	sorted := make([]Change, len(changes))
	copy(sorted, changes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	return sorted
}

func formatChangeDetails(builderStr string, sortedChanges []Change) string {
	var builder strings.Builder
	builder.WriteString(builderStr)

	for _, change := range sortedChanges {
		switch change.Type {
		case ChangeTypeAdded:
			fmt.Fprintf(&builder, "+ %s\n", change.Description)
		case ChangeTypeRemoved:
			fmt.Fprintf(&builder, "- %s\n", change.Description)
		case ChangeTypeModified:
			fmt.Fprintf(&builder, "~ %s\n", change.Description)
		}
	}

	return builder.String()
}

// GetSummary returns a brief summary of changes.
func (d *Differ) GetSummary(changes []Change) string {
	if len(changes) == 0 {
		return "No changes"
	}

	added, removed, modified := countChangeTypes(changes)
	parts := buildSummaryParts(added, removed, modified)

	return strings.Join(parts, ", ")
}

func countChangeTypes(changes []Change) (added, removed, modified int) {
	for _, c := range changes {
		switch c.Type {
		case ChangeTypeAdded:
			added++
		case ChangeTypeRemoved:
			removed++
		case ChangeTypeModified:
			modified++
		}
	}

	return added, removed, modified
}

func buildSummaryParts(added, removed, modified int) []string {
	parts := []string{}
	if added > 0 {
		parts = append(parts, fmt.Sprintf("%d added", added))
	}

	if removed > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", removed))
	}

	if modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", modified))
	}

	return parts
}
