// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Validator is an interface for configuration validation.
type Validator interface {
	ValidateConfig(m *Migrator) error
}

// DefaultValidator uses golangci-lint for validation.
type DefaultValidator struct{}

func (v DefaultValidator) ValidateConfig(migrator *Migrator) error {
	golangciLintPath, err := exec.LookPath("golangci-lint")
	if err != nil {
		return fmt.Errorf("golangci-lint not found: %w", err)
	}

	//nolint:mnd // 30 seconds is a reasonable timeout for config validation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, golangciLintPath, "config", "verify")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("config validation failed: %w\nOutput: %s", err, stderr.String())
	}

	if migrator.verbose {
		cliPrintln("Configuration is valid")
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
		return ErrMockValidationFailed
	}

	//nolint:err113 // Test helper that needs dynamic error message
	return fmt.Errorf("%s", v.ErrorMessage)
}
