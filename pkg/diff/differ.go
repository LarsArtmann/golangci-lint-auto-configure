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
func (d *Differ) Compare(old, new *types.Config) []Change {
	var changes []Change

	// Compare version
	if old.Version != new.Version {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "version",
			OldValue:    old.Version,
			NewValue:    new.Version,
			Description: fmt.Sprintf("Version changed from %s to %s", old.Version, new.Version),
		})
	}

	// Compare run settings
	changes = append(changes, d.compareRunSettings(old.Run, new.Run)...)

	// Compare linters
	changes = append(changes, d.compareLinters(old.Linters, new.Linters)...)

	// Compare formatters
	changes = append(changes, d.compareFormatters(old.Formatters, new.Formatters)...)

	return changes
}

func (d *Differ) compareRunSettings(old, new types.RunConfig) []Change {
	var changes []Change

	if old.Timeout != new.Timeout {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.timeout",
			OldValue:    old.Timeout,
			NewValue:    new.Timeout,
			Description: fmt.Sprintf("Timeout changed from %s to %s", old.Timeout, new.Timeout),
		})
	}

	if old.Go != new.Go {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.go",
			OldValue:    old.Go,
			NewValue:    new.Go,
			Description: fmt.Sprintf("Go version changed from %s to %s", old.Go, new.Go),
		})
	}

	if old.Tests != new.Tests {
		changes = append(changes, Change{
			Type:        ChangeTypeModified,
			Path:        "run.tests",
			OldValue:    strconv.FormatBool(old.Tests),
			NewValue:    strconv.FormatBool(new.Tests),
			Description: fmt.Sprintf("Tests changed from %v to %v", old.Tests, new.Tests),
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

func (d *Differ) compareLinters(old, new types.LintersConfig) []Change {
	return d.compareEnabled(old.Enable, new.Enable, "linters", "linter")
}

func (d *Differ) compareFormatters(old, new types.FormattersConfig) []Change {
	return d.compareEnabled(old.Enable, new.Enable, "formatters", "formatter")
}

// FormatChanges formats changes as a human-readable string.
func (d *Differ) FormatChanges(changes []Change) string {
	if len(changes) == 0 {
		return "No changes detected"
	}

	var sb strings.Builder

	added := 0
	removed := 0
	modified := 0

	// Group by type
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

	fmt.Fprintf(&sb, "Changes: %d added, %d removed, %d modified\n\n", added, removed, modified)

	// Sort changes by path
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	for _, c := range changes {
		switch c.Type {
		case ChangeTypeAdded:
			fmt.Fprintf(&sb, "+ %s\n", c.Description)
		case ChangeTypeRemoved:
			fmt.Fprintf(&sb, "- %s\n", c.Description)
		case ChangeTypeModified:
			fmt.Fprintf(&sb, "~ %s\n", c.Description)
		}
	}

	return sb.String()
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
