// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

import (
	"errors"
	"fmt"
)

// Runner executes multiple formatters on a codebase.
type Runner struct {
	Dir        string
	Formatters []Formatter
	Mode       FormatterMode
	Verbose    bool
}

// NewRunner creates a new formatter runner with default formatters.
func NewRunner() *Runner {
	formatters := []Formatter{
		NewGoFmt(),
		NewGoImports(),
		NewGoFumpt(),
	}

	return &Runner{
		Formatters: formatters,
		Mode:       ModeFix,
		Dir:        ".",
		Verbose:    false,
	}
}

// SetMode sets the execution mode.
func (r *Runner) SetMode(mode FormatterMode) *Runner {
	r.Mode = mode
	return r
}

// SetDir sets the directory to run formatters on.
func (r *Runner) SetDir(dir string) *Runner {
	r.Dir = dir
	return r
}

// SetVerbose enables or disables verbose logging.
func (r *Runner) SetVerbose(verbose bool) *Runner {
	r.Verbose = verbose
	return r
}

// SetFormatters sets custom formatters to run.
func (r *Runner) SetFormatters(formatters []Formatter) *Runner {
	r.Formatters = formatters
	return r
}

// Run executes all configured formatters in order.
func (r *Runner) Run() ([]*FormatterResult, error) {
	if r.Dir == "" {
		return nil, errors.New("directory not set")
	}

	results := make([]*FormatterResult, 0, len(r.Formatters))
	skippedCount := 0

	for _, formatter := range r.Formatters {
		if !formatter.Available() {
			result := &FormatterResult{
				Type:     formatter.Type(),
				Files:    []string{},
				ExitCode: 0,
			}

			if r.Verbose && r.Mode == ModeFix {
				fmt.Printf("[SKIP] %s: not installed\n", formatter.Type())
			}

			skippedCount++
			results = append(results, result)
			continue
		}

		if r.Verbose {
			fmt.Printf("[RUN] %s: %s mode on %s\n", formatter.Type(), r.Mode, r.Dir)
		}

		result, err := formatter.Run(r.Mode, r.Dir)
		if err != nil {
			results = append(results, result)
			return results, fmt.Errorf("formatter %s failed: %w", formatter.Type(), err)
		}

		results = append(results, result)

		if r.Verbose {
			if result.HasChanges() {
				fmt.Printf("[OK] %s: %d file(s) affected\n", formatter.Type(), len(result.Files))
			} else {
				fmt.Printf("[OK] %s: no changes needed\n", formatter.Type())
			}
		}
	}

	if r.Verbose && skippedCount > 0 {
		fmt.Printf("[INFO] Skipped %d unavailable formatter(s)\n", skippedCount)
	}

	return results, nil
}

// GetSummary returns a summary of all formatter results.
func (r *Runner) GetSummary(results []*FormatterResult) string {
	totalFiles := 0
	successCount := 0
	failCount := 0
	changesCount := 0

	for _, result := range results {
		if result.Success() {
			successCount++
		} else {
			failCount++
		}

		if result.HasChanges() {
			changesCount++
			totalFiles += len(result.Files)
		}
	}

	if failCount > 0 {
		return fmt.Sprintf(
			"Formatters: %d success, %d failed, %d with changes",
			successCount,
			failCount,
			changesCount,
		)
	}

	if changesCount == 0 {
		return fmt.Sprintf("Formatters: %d success, no changes needed", successCount)
	}

	return fmt.Sprintf(
		"Formatters: %d success, %d with changes, %d total files affected",
		successCount,
		changesCount,
		totalFiles,
	)
}
