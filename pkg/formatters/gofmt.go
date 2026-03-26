// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

import (
	"fmt"
	"os/exec"
)

// GoFmtFormatter implements gofmt formatter.
type GoFmtFormatter struct{}

// NewGoFmt creates a new gofmt formatter.
func NewGoFmt() *GoFmtFormatter {
	return &GoFmtFormatter{}
}

// Name returns the name of the formatter.
func (f *GoFmtFormatter) Name() string {
	return "gofmt"
}

// Type returns the formatter type.
func (f *GoFmtFormatter) Type() FormatterType {
	return FormatterGoFmt
}

// Available checks if gofmt is available (always true - part of Go toolchain).
func (f *GoFmtFormatter) Available() bool {
	return lookPath("gofmt")
}

// Run executes gofmt in the specified directory with the given mode.
func (f *GoFmtFormatter) Run(mode FormatterMode, dir string) (*FormatterResult, error) {
	flags, err := getFlags(mode)
	if err != nil {
		return nil, fmt.Errorf("invalid mode: %w", err)
	}

	if mode == ModeFix {
		checkCmd := exec.Command("gofmt", "-l", dir)
		checkOutput, _ := checkCmd.Output()
		filesToFix := parseOutput(string(checkOutput))

		if len(filesToFix) > 0 {
			fixCmd := exec.Command("gofmt", "-w", dir)
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

	cmd := exec.Command("gofmt", flags, dir)
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
