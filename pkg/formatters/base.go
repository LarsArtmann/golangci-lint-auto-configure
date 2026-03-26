// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

import (
	"errors"
	"fmt"
	"os/exec"
)

// baseFormatter provides common functionality for external formatter tools.
type baseFormatter struct {
	toolErr error
	name    string
	tool    string
}

// newBaseFormatter creates a new base formatter.
func newBaseFormatter(name, tool string, toolErr error) baseFormatter {
	return baseFormatter{
		name:    name,
		tool:    tool,
		toolErr: toolErr,
	}
}

// Name returns the name of the formatter.
func (f *baseFormatter) Name() string {
	return f.name
}

// Type returns the formatter type - returns invalid value by default.
func (f *baseFormatter) Type() FormatterType {
	return FormatterType(-1)
}

// Available checks if the formatter tool is installed.
func (f *baseFormatter) Available() bool {
	return lookPath(f.tool)
}

// Run executes the formatter in the specified directory with the given mode.
func (f *baseFormatter) Run(mode FormatterMode, dir string) (*FormatterResult, error) {
	if !f.Available() {
		return &FormatterResult{
			Type:     f.Type(),
			Files:    []string{},
			ExitCode: 1,
			Error:    fmt.Errorf("%w: %s", f.toolErr, f.tool),
		}, f.toolErr
	}

	flags, err := getFlags(mode)
	if err != nil {
		return nil, fmt.Errorf("invalid mode: %w", err)
	}

	if mode == ModeFix {
		checkCmd := exec.Command(f.tool, "-l", dir)
		checkOutput, err := checkCmd.Output()
		filesToFix := parseOutput(string(checkOutput))

		exitError := &exec.ExitError{}

		if err != nil && !errors.As(err, &exitError) {
			return &FormatterResult{
				Type:     f.Type(),
				Files:    []string{},
				ExitCode: 1,
				Error:    err,
			}, err
		}

		if len(filesToFix) > 0 {
			fixCmd := exec.Command(f.tool, "-w", dir)
			err = fixCmd.Run()
			if err != nil {
				return &FormatterResult{
					Type:     f.Type(),
					Files:    filesToFix,
					ExitCode: 1,
					Error:    err,
				}, err
			}

			return &FormatterResult{
				Type:     f.Type(),
				Files:    filesToFix,
				ExitCode: 0,
			}, nil
		}

		return &FormatterResult{
			Type:     f.Type(),
			Files:    []string{},
			ExitCode: 0,
		}, nil
	}

	cmd := exec.Command(f.tool, flags, dir)
	stdoutBytes, err := cmd.Output()
	if err != nil {
		return handleExecError(f.Type(), stdoutBytes, err)
	}

	return &FormatterResult{
		Type:     f.Type(),
		Files:    parseOutput(string(stdoutBytes)),
		ExitCode: 0,
		Output:   string(stdoutBytes),
	}, nil
}

func handleExecError(formatterType FormatterType, stdoutBytes []byte, err error) (*FormatterResult, error) {
	exitError := &exec.ExitError{}
	if errors.As(err, &exitError) {
		return &FormatterResult{
			Type:     formatterType,
			Files:    parseOutput(string(stdoutBytes)),
			ExitCode: exitError.ExitCode(),
			Output:   string(stdoutBytes),
			Stderr:   string(exitError.Stderr),
		}, nil
	}

	return &FormatterResult{
		Type:     formatterType,
		Files:    []string{},
		ExitCode: 1,
		Error:    err,
	}, err
}
