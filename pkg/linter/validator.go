package linter

import (
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// ValidationResult holds the result of validating linters and formatters.
type ValidationResult struct {
	DeprecatedLinters []DeprecatedLinterCheck
	RedundantLinters  []RedundantLinterCheck
	FormattersToAdd   []string
}

// DeprecatedLinterCheck represents a deprecated linter that needs replacement.
type DeprecatedLinterCheck struct {
	Name           string
	Replacement    types.LinterReplacement
	AlreadyPresent bool
}

// RedundantLinterCheck represents a redundant linter that should be removed.
type RedundantLinterCheck struct {
	Name   string
	Reason string
}

// Validator provides functionality to validate golangci-lint configurations.
type Validator struct{}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateLinters checks for deprecated and redundant linters.
func (v *Validator) ValidateLinters(
	enabledLinters []string,
	_ *types.ConfigAnalysis,
	formatterSet map[string]bool,
	shouldEnableGolines bool,
	dryRun bool,
) ValidationResult {
	result := ValidationResult{
		DeprecatedLinters: make([]DeprecatedLinterCheck, 0),
		RedundantLinters:  make([]RedundantLinterCheck, 0),
		FormattersToAdd:   make([]string, 0),
	}

	linterSet := make(map[string]bool)
	for _, linter := range enabledLinters {
		linterSet[linter] = true
	}

	// Check for deprecated linters
	for _, linter := range enabledLinters {
		if replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			result.DeprecatedLinters = append(result.DeprecatedLinters, DeprecatedLinterCheck{
				Name:           linter,
				Replacement:    replacement,
				AlreadyPresent: linterSet[string(replacement.Replacement)],
			})
		}
	}

	// Check for redundant linters when formatters are enabled
	for linterName, reason := range constants.RedundantLinters {
		if linterSet[string(linterName)] {
			golinesWillBeEnabled := formatterSet["golines"] || (shouldEnableGolines && dryRun)

			if linterName == "lll" && golinesWillBeEnabled {
				result.RedundantLinters = append(result.RedundantLinters, RedundantLinterCheck{
					Name:   string(linterName),
					Reason: reason,
				})
			}
		}
	}

	return result
}

// ShouldEnableGolines checks if golines formatter should be enabled.
func (v *Validator) ShouldEnableGolines(analysis *types.ConfigAnalysis) bool {
	for _, rec := range analysis.FormatterRecommendations {
		if rec.Name == "golines" && rec.Priority == types.FormatterPriorityHigh {
			return true
		}
	}

	return false
}
