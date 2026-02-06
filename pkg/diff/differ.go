package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Change represents a single change between two configs
type Change struct {
	Type        ChangeType
	Path        string
	OldValue    string
	NewValue    string
	Description string
}

// ChangeType indicates the type of change
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

// Differ compares two configurations and returns the differences
type Differ struct{}

// NewDiffer creates a new config differ
func NewDiffer() *Differ {
	return &Differ{}
}

// Compare compares two configs and returns the changes
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
			OldValue:    fmt.Sprintf("%v", old.Tests),
			NewValue:    fmt.Sprintf("%v", new.Tests),
			Description: fmt.Sprintf("Tests changed from %v to %v", old.Tests, new.Tests),
		})
	}

	return changes
}

func (d *Differ) compareLinters(old, new types.LintersConfig) []Change {
	var changes []Change

	// Compare enabled linters
	oldEnabled := make(map[string]bool)
	for _, l := range old.Enable {
		oldEnabled[l] = true
	}

	newEnabled := make(map[string]bool)
	for _, l := range new.Enable {
		newEnabled[l] = true
	}

	// Find added linters
	for l := range newEnabled {
		if !oldEnabled[l] {
			changes = append(changes, Change{
				Type:        ChangeTypeAdded,
				Path:        fmt.Sprintf("linters.enable.%s", l),
				OldValue:    "",
				NewValue:    l,
				Description: fmt.Sprintf("Enabled linter: %s", l),
			})
		}
	}

	// Find removed linters
	for l := range oldEnabled {
		if !newEnabled[l] {
			changes = append(changes, Change{
				Type:        ChangeTypeRemoved,
				Path:        fmt.Sprintf("linters.enable.%s", l),
				OldValue:    l,
				NewValue:    "",
				Description: fmt.Sprintf("Disabled linter: %s", l),
			})
		}
	}

	return changes
}

func (d *Differ) compareFormatters(old, new types.FormattersConfig) []Change {
	var changes []Change

	oldEnabled := make(map[string]bool)
	for _, f := range old.Enable {
		oldEnabled[f] = true
	}

	newEnabled := make(map[string]bool)
	for _, f := range new.Enable {
		newEnabled[f] = true
	}

	// Find added formatters
	for f := range newEnabled {
		if !oldEnabled[f] {
			changes = append(changes, Change{
				Type:        ChangeTypeAdded,
				Path:        fmt.Sprintf("formatters.enable.%s", f),
				OldValue:    "",
				NewValue:    f,
				Description: fmt.Sprintf("Enabled formatter: %s", f),
			})
		}
	}

	// Find removed formatters
	for f := range oldEnabled {
		if !newEnabled[f] {
			changes = append(changes, Change{
				Type:        ChangeTypeRemoved,
				Path:        fmt.Sprintf("formatters.enable.%s", f),
				OldValue:    f,
				NewValue:    "",
				Description: fmt.Sprintf("Disabled formatter: %s", f),
			})
		}
	}

	return changes
}

// FormatChanges formats changes as a human-readable string
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

	sb.WriteString(fmt.Sprintf("Changes: %d added, %d removed, %d modified\n\n", added, removed, modified))

	// Sort changes by path
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	for _, c := range changes {
		switch c.Type {
		case ChangeTypeAdded:
			sb.WriteString(fmt.Sprintf("+ %s\n", c.Description))
		case ChangeTypeRemoved:
			sb.WriteString(fmt.Sprintf("- %s\n", c.Description))
		case ChangeTypeModified:
			sb.WriteString(fmt.Sprintf("~ %s\n", c.Description))
		}
	}

	return sb.String()
}

// GetSummary returns a brief summary of changes
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
