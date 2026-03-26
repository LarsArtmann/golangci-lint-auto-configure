// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

// Validator is an interface for configuration validation.
type Validator interface {
	ValidateConfig(m *Migrator) error
}

// DefaultValidator uses golangci-lint for validation.
type DefaultValidator struct{}

func (v DefaultValidator) ValidateConfig(m *Migrator) error {
	golangciLintPath, err := exec.LookPath("golangci-lint")
	if err != nil {
		return fmt.Errorf("golangci-lint not found: %w", err)
	}

	cmd := exec.Command(golangciLintPath, "config", "verify")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("config validation failed: %w\nOutput: %s", err, stderr.String())
	}

	if m.verbose {
		fmt.Println("Configuration is valid")
	}

	return nil
}

// MockValidator is a mock validator for testing.
type MockValidator struct{}

func (v MockValidator) ValidateConfig(_ *Migrator) error {
	return nil
}

// FailingValidator is a mock validator that always fails.
type FailingValidator struct {
	ErrorMessage string
}

func (v FailingValidator) ValidateConfig(_ *Migrator) error {
	if v.ErrorMessage == "" {
		return errors.New("mock validation failed")
	}
	return fmt.Errorf("%s", v.ErrorMessage)
}
