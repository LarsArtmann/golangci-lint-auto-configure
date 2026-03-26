// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

// Package utils provides shared utility functions for the application.
package utils

import (
	"context"
	"os/exec"
	"time"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
)

// GitCheckTimeout is the default timeout for git operations.
const GitCheckTimeout = 5 * time.Second

// IsGitRepo checks if the specified directory is inside a git repository.
// Returns true if inside a git repo, false otherwise.
func IsGitRepo(ctx context.Context, dir string) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Git returns "true" with a newline when inside a work tree
	return len(output) > 0 && output[0] == 't'
}

// CheckGitRepo verifies that the specified directory is inside a git repository.
// Returns apperrors.ErrNotGitRepository or apperrors.ErrNotInGitWorkingTree if not in a git repo.
func CheckGitRepo(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return apperrors.ErrNotGitRepository
	}

	// Git returns "true" with a newline when inside a work tree
	if len(output) == 0 || output[0] != 't' {
		return apperrors.ErrNotInGitWorkingTree
	}

	return nil
}

// CheckGitRepoWithTimeout verifies git repository with a default timeout.
// This is a convenience function that creates a context with GitCheckTimeout.
func CheckGitRepoWithTimeout(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), GitCheckTimeout)
	defer cancel()

	return CheckGitRepo(ctx, dir)
}
