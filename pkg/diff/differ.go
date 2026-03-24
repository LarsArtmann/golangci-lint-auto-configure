package diff

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
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

	if old.Timeout != newConfig.Timeout {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.timeout",
			OldValue:    old.Timeout,
			NewValue:    newConfig.Timeout,
			Description: fmt.Sprintf("Timeout changed from %s to %s", old.Timeout, newConfig.Timeout),
		})
	}

	if old.Go != newConfig.Go {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.go",
			OldValue:    old.Go,
			NewValue:    newConfig.Go,
			Description: fmt.Sprintf("Go version changed from %s to %s", old.Go, newConfig.Go),
		})
	}

	if old.Tests != newConfig.Tests {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.tests",
			OldValue:    strconv.FormatBool(old.Tests),
			NewValue:    strconv.FormatBool(newConfig.Tests),
			Description: fmt.Sprintf("Tests changed from %v to %v", old.Tests, newConfig.Tests),
		})
	}

	return changes
}

func (d *Differ) compareEnabled(oldEnable, newEnable []string, pathPrefix, entityName string) []Change {
	var changes []Change

	oldEnabled := make(map[string]bool)
	for _, item := range oldEnable {
		oldEnabled[item] = true
	}

	newEnabled := make(map[string]bool)
	for _, item := range newEnable {
		newEnabled[item] = true
	}

	// Find added items
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

	// Find removed items
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

	var builder strings.Builder

	added := 0
	removed := 0
	modified := 0

	// Group by type
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

	fmt.Fprintf(&builder, "Changes: %d added, %d removed, %d modified\n\n", added, removed, modified)

	// Sort changes by path
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	for _, change := range changes {
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

	added := 0
	removed := 0
	modified := 0

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

	return strings.Join(parts, ", ")
}
