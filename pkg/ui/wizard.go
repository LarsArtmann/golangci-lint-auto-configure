package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// WizardOption represents an option in the wizard.
type WizardOption struct {
	LinterName string
	Priority   int
	Reason     string
	Selected   bool
}

// WizardResult represents the result of the wizard.
type WizardResult struct {
	SelectedLinters []string
	Priority        types.LinterPriority
	Confirmed       bool
}

// RunLinterSelectionWizard runs an interactive TUI wizard for linter selection.
func RunLinterSelectionWizard(analysis *types.ConfigAnalysis) (*WizardResult, error) {
	var result WizardResult
	var selectedLinters []string

	// Create a map to track selections
	selections := make(map[string]bool)

	// Create huh group for linters by priority
	var groups []*huh.Group

	// Priority labels
	priorityLabels := []string{"Critical", "High Priority", "Medium Priority", "Optional"}
	priorityDescriptions := []string{
		"Security and correctness issues",
		"Quality and maintainability",
		"Style and consistency",
		"May be too strict for most projects",
	}

	// Group linters by priority
	for p := 0; p <= 3; p++ {
		var options []huh.Option[string]
		for _, rec := range analysis.LinterRecommendations {
			if int(rec.Priority) != p {
				continue
			}

			// Check if already enabled
			enabled := false
			for _, e := range analysis.EnabledLinters {
				if e.Name.String() == rec.Name.String() {
					enabled = true
					break
				}
			}

			if enabled {
				// Already enabled, skip
				continue
			}

			desc := rec.Reason
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}

			option := huh.NewOption(
				fmt.Sprintf("%s - %s", rec.Name.String(), desc),
				rec.Name.String(),
			)
			options = append(options, option)
		}

		if len(options) == 0 {
			continue
		}

		// Create multi-select for this priority group
		multiSelect := huh.NewMultiSelect[string]().
			Title(priorityLabels[p]).
			Description(priorityDescriptions[p]).
			Options(options...).
			WithMini(true)

		group := huh.NewGroup(multiSelect)
		groups = append(groups, group)
	}

	if len(groups) == 0 {
		return &WizardResult{
			SelectedLinters: []string{},
			Confirmed:       true,
		}, nil
	}

	// Add priority selection
	var selectedPriority types.LinterPriority = types.LinterPriorityHigh
	prioritySelect := huh.NewSelect[types.LinterPriority]().
		Title("Select minimum priority level").
		Description("Which priority of linters should be enabled?").
		Options(
			huh.NewOption("Critical only - Security and correctness", types.LinterPriorityCritical),
			huh.NewOption("High and above - Quality and maintainability (Recommended)", types.LinterPriorityHigh),
			huh.NewOption("Medium and above - Style and consistency", types.LinterPriorityMedium),
			huh.NewOption("All - Including optional linters", types.LinterPriorityOptional),
		).
		Value(&selectedPriority)

	priorityGroup := huh.NewGroup(prioritySelect)
	groups = append([]*huh.Group{priorityGroup}, groups...)

	// Add confirmation
	var confirmed bool
	confirmSelect := huh.NewConfirm().
		Title("Apply these changes?").
		Description("This will modify your .golangci.yml configuration file.").
		Value(&confirmed)

	confirmGroup := huh.NewGroup(confirmSelect)
	groups = append(groups, confirmGroup)

	// Run the form
	form := huh.NewForm(groups...)

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run wizard: %w", err)
	}

	if !confirmed {
		return nil, nil
	}

	// Collect selected linters from all multi-selects
	for _, group := range groups[1 : len(groups)-1] { // Skip priority and confirm groups
		for _, field := range group.Fields() {
			if ms, ok := field.(*huh.MultiSelect[string]); ok {
				for _, v := range ms.GetValue() {
					selections[v] = true
				}
			}
		}
	}

	for linter := range selections {
		selectedLinters = append(selectedLinters, linter)
	}

	return &WizardResult{
		SelectedLinters: selectedLinters,
		Priority:        selectedPriority,
		Confirmed:       confirmed,
	}, nil
}

// RunPriorityWizard runs a simple priority selection wizard.
func RunPriorityWizard() (types.LinterPriority, error) {
	var priority types.LinterPriority

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[types.LinterPriority]().
				Title("Select priority level").
				Description("Choose which linters to enable based on priority").
				Options(
					huh.NewOption("🔴 Critical - Security and correctness only", types.LinterPriorityCritical),
					huh.NewOption("🟠 High - Quality and maintainability (Recommended)", types.LinterPriorityHigh),
					huh.NewOption("🟡 Medium - Style and consistency", types.LinterPriorityMedium),
					huh.NewOption("🔵 All - Including optional linters", types.LinterPriorityOptional),
				).
				Value(&priority),
		),
	)

	err := form.Run()
	if err != nil {
		return types.LinterPriorityHigh, fmt.Errorf("failed to run priority wizard: %w", err)
	}

	return priority, nil
}

// RunConfirmWizard runs a confirmation wizard.
func RunConfirmWizard(title, description string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Value(&confirmed),
		),
	)

	err := form.Run()
	if err != nil {
		return false, fmt.Errorf("failed to run confirm wizard: %w", err)
	}

	return confirmed, nil
}
