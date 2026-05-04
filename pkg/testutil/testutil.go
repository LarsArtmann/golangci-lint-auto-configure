// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides common test utilities for the project.
package testutil

import (
	"os"

	"charm.land/log/v2"
)

// TempFileConfig is a helper for creating temporary config files in tests.
type TempFileConfig struct {
	Path    string
	Content string
}

// WriteConfigFile writes a config file with the given content.
// Returns the file path and an error if writing fails.
// Callers should typically use defer os.Remove(path) to clean up.
func WriteConfigFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

// NewTestLogger creates a logger for testing with error level.
func NewTestLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
}
