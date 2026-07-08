// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
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
		return apperrors.WrapClassified(err, "migration.find_binary",
			"golangci-lint not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), ValidationTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, golangciLintPath, "config", "verify")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return errorfamily.WrapRejection(err, "migration.validate_config",
			"config validation failed").
			WithContext("output", stderr.String())
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
