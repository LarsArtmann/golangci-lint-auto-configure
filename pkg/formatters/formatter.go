// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Formatter represents a code formatter.
type Formatter interface {
	Name() string
	Type() FormatterType
	Available() bool
	Run(mode FormatterMode, dir string) (*FormatterResult, error)
}

// FormatterResult contains the result of running a formatter.
type FormatterResult struct {
	Error    error
	Output   string
	Stderr   string
	Files    []string
	Type     FormatterType
	ExitCode int
}

var (
	// ErrFormatterNotFound is returned when a formatter tool is not installed.
	ErrFormatterNotFound = errors.New("formatter tool not found")

	// ErrInvalidMode is returned when an invalid mode is specified.
	ErrInvalidMode = errors.New("invalid formatter mode")
)

// Success returns true if the formatter ran successfully.
func (r *FormatterResult) Success() bool {
	return r.Error == nil && r.ExitCode == 0
}

// HasChanges returns true if files were modified or changes were detected.
func (r *FormatterResult) HasChanges() bool {
	return len(r.Files) > 0
}

// Summary returns a human-readable summary of the result.
func (r *FormatterResult) Summary() string {
	if !r.Success() {
		return fmt.Sprintf("%s: failed (exit code %d, error: %v)", r.Type, r.ExitCode, r.Error)
	}

	if !r.HasChanges() {
		return fmt.Sprintf("%s: no changes needed", r.Type)
	}

	return fmt.Sprintf("%s: %d file(s) affected", r.Type, len(r.Files))
}

// String returns the formatter type as a string.
func (t FormatterType) String() string {
	switch t {
	case FormatterGoFmt:
		return "gofmt"
	case FormatterGoImports:
		return "goimports"
	case FormatterGoFumpt:
		return "gofumpt"
	default:
		return "unknown"
	}
}

// String returns the mode as a string.
func (m FormatterMode) String() string {
	switch m {
	case ModeCheck:
		return "check"
	case ModeFix:
		return "fix"
	case ModeDiff:
		return "diff"
	default:
		return "unknown"
	}
}

// getFlags returns the appropriate command-line flags for a mode.
func getFlags(mode FormatterMode) (string, error) {
	switch mode {
	case ModeCheck:
		return "-l", nil
	case ModeFix:
		return "-w", nil
	case ModeDiff:
		return "-d", nil
	default:
		return "", fmt.Errorf("%w: %v", ErrInvalidMode, mode)
	}
}

// lookPath checks if a command is available in PATH.
func lookPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// parseOutput parses formatter output to extract affected files.
func parseOutput(output string) []string {
	if output == "" {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			files = append(files, line)
		}
	}

	return files
}
