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
			OldValue:    string(old.Version),
			NewValue:    string(newConfig.Version),
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
		func(o, n string) string { return formatChangeMessage("Timeout", o, n) })
	changes = d.addChangeIfDifferent(changes, old.Go, newConfig.Go, "run.go",
		func(o, n string) string { return formatChangeMessage("Go version", o, n) })
	changes = d.addTestChangeIfDifferent(changes, old.Tests, newConfig.Tests)

	return changes
}

func (d *Differ) addChangeIfDifferent(
	changes []Change,
	oldVal, newVal string,
	path string,
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

func formatChangeMessage(fieldName, oldVal, newVal string) string {
	return fmt.Sprintf("%s changed from %s to %s", fieldName, oldVal, newVal)
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

func (d *Differ) compareListChanges(oldItems, newItems []string, pathPrefix, entityName, subKey string) []Change {
	oldSet := types.NewSet(oldItems...)
	newSet := types.NewSet(newItems...)

	var changes []Change

	for item := range newSet {
		if !oldSet.Contains(item) {
			changes = append(changes, d.makeAddedChange(pathPrefix, subKey, entityName, item))
		}
	}

	for item := range oldSet {
		if !newSet.Contains(item) {
			changes = append(changes, d.makeRemovedChange(pathPrefix, subKey, entityName, item))
		}
	}

	return changes
}

func changeAction(subKey, enabledAction, disabledAction string) string {
	if subKey == "disable" {
		return disabledAction
	}

	return enabledAction
}

func linterNamesToStrings(names []types.LinterName) []string {
	result := make([]string, len(names))

	for i, n := range names {
		result[i] = string(n)
	}

	return result
}

func (d *Differ) makeAddedChange(pathPrefix, subKey, entityName, item string) Change {
	return Change{
		Type:        ChangeTypeAdded,
		Path:        fmt.Sprintf("%s.%s.%s", pathPrefix, subKey, item),
		OldValue:    "",
		NewValue:    item,
		Description: fmt.Sprintf("%s %s: %s", changeAction(subKey, "Enabled", "Disabled"), entityName, item),
	}
}

func (d *Differ) makeRemovedChange(pathPrefix, subKey, entityName, item string) Change {
	return Change{
		Type:        ChangeTypeRemoved,
		Path:        fmt.Sprintf("%s.%s.%s", pathPrefix, subKey, item),
		OldValue:    item,
		NewValue:    "",
		Description: fmt.Sprintf("%s %s: %s", changeAction(subKey, "Disabled", "Re-enabled"), entityName, item),
	}
}

func (d *Differ) compareLinters(old, newCfg types.LintersConfig) []Change {
	return d.compareEnableDisable(
		linterNamesToStrings(old.Enable), linterNamesToStrings(old.Disable),
		linterNamesToStrings(newCfg.Enable), linterNamesToStrings(newCfg.Disable),
		"linters", "linter",
	)
}

func (d *Differ) compareFormatters(old, newCfg types.FormattersConfig) []Change {
	return d.compareEnableDisable(old.Enable, old.Disable, newCfg.Enable, newCfg.Disable, "formatters", "formatter")
}

func (d *Differ) compareEnableDisable(
	oldEnable, oldDisable, newEnable, newDisable []string,
	pathPrefix, entityName string,
) []Change {
	changes := make([]Change, 0, len(oldEnable)+len(oldDisable)+len(newEnable)+len(newDisable))
	changes = append(changes, d.compareListChanges(oldEnable, newEnable, pathPrefix, entityName, "enable")...)
	changes = append(changes, d.compareListChanges(oldDisable, newDisable, pathPrefix, entityName, "disable")...)

	return changes
}

func changeCounts(changes []Change) (int, int, int, bool) {
	if len(changes) == 0 {
		return 0, 0, 0, false
	}

	added, removed, modified := countChangesByType(changes)

	return added, removed, modified, true
}

func changesOrEmpty(changes []Change, emptyMessage string) (int, int, int, string) {
	added, removed, modified, hasChanges := changeCounts(changes)
	if !hasChanges {
		return 0, 0, 0, emptyMessage
	}

	return added, removed, modified, ""
}

func changesOrMessage(changes []Change, emptyMessage string) (int, int, int, *string) {
	added, removed, modified, emptyResult := changesOrEmpty(changes, emptyMessage)
	if emptyResult != "" {
		return 0, 0, 0, &emptyResult
	}

	return added, removed, modified, nil
}

// FormatChanges formats changes as a human-readable string.
func (d *Differ) FormatChanges(changes []Change) string {
	added, removed, modified, emptyResult := changesOrMessage(changes, "No changes detected")
	if emptyResult != nil {
		return *emptyResult
	}

	builder := formatChangeHeader(added, removed, modified)
	sortedChanges := sortChangesByPath(changes)

	return formatChangeDetails(builder.String(), sortedChanges)
}

func countChangesByType(changes []Change) (int, int, int) {
	var added, removed, modified int

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
	sorted := make([]Change, 0, len(changes))
	sorted = append(sorted, changes...)
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
	added, removed, modified, emptyResult := changesOrMessage(changes, "No changes")
	if emptyResult != nil {
		return *emptyResult
	}

	parts := buildSummaryParts(added, removed, modified)

	return strings.Join(parts, ", ")
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
