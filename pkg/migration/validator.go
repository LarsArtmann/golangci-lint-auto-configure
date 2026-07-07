// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
)

// Validator is an interface for configuration validation.
type Validator interface {
	ValidateConfig(m *Migrator) error
}

// ValidationTimeout is the timeout for config validation.
const ValidationTimeout = 30 * time.Second

// DefaultValidator uses golangci-lint for validation.
type DefaultValidator struct{}

func (v DefaultValidator) ValidateConfig(migrator *Migrator) error {
	golangciLintPath, err := exec.LookPath(constants.GolangciLintBinaryName)
	if err != nil {
		return fmt.Errorf("golangci-lint not found: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ValidationTimeout)
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
