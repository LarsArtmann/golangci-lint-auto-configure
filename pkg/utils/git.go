// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

// Package utils provides shared utility functions for the application.
package utils

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

// Git operation errors.
var (
	// ErrNotGitRepository indicates the current directory is not in a git repository.
	ErrNotGitRepository = errors.New("not in a git repository (use git init or clone a repo first)")

	// ErrNotInGitWorkingTree indicates the current directory is not inside a git working tree.
	ErrNotInGitWorkingTree = errors.New("not inside git working tree")
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
// Returns ErrNotGitRepository or ErrNotInGitWorkingTree if not in a git repo.
func CheckGitRepo(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return ErrNotGitRepository
	}

	// Git returns "true" with a newline when inside a work tree
	if len(output) == 0 || output[0] != 't' {
		return ErrNotInGitWorkingTree
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
