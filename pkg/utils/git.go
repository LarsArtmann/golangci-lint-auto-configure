// Package utils provides shared utility functions for the golangci-lint-auto-configure tool.
// It includes Git repository detection utilities.
package utils

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
)

// GitCheckTimeout is the default timeout for git operations.
const GitCheckTimeout = 5 * time.Second

func gitOutput(ctx context.Context, dir string, args ...string) string {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(bytes.TrimSpace(output))
}

// gitIsInsideWorkTree runs `git rev-parse --is-inside-work-tree` in dir and returns the trimmed output.
// Returns the empty string when git is unavailable or the directory is not inside a git working tree.
func gitIsInsideWorkTree(ctx context.Context, dir string) string {
	return gitOutput(ctx, dir, "rev-parse", "--is-inside-work-tree")
}

// IsGitRepo checks if the specified directory is inside a git repository.
// Returns true if inside a git repo, false otherwise.
func IsGitRepo(ctx context.Context, dir string) bool {
	return gitIsInsideWorkTree(ctx, dir) == "true"
}

// CheckGitRepo verifies that the specified directory is inside a git repository.
// Returns apperrors.ErrNotGitRepository or apperrors.ErrNotInGitWorkingTree if not in a git repo.
func CheckGitRepo(ctx context.Context, dir string) error {
	result := gitIsInsideWorkTree(ctx, dir)

	switch result {
	case "":
		return errorfamily.WrapRejectionf(apperrors.ErrNotGitRepository, "git.not_repository",
			"not a git repository (dir=%s)", dir)
	case "true":
		return nil
	default:
		return errorfamily.WrapRejectionf(apperrors.ErrNotInGitWorkingTree, "git.not_work_tree",
			"not in git working tree (dir=%s)", dir)
	}
}

// GitHead returns the current commit hash of the repo at dir, or "" when git is
// unavailable or dir is not inside a git work tree.
func GitHead(ctx context.Context, dir string) string {
	return gitOutput(ctx, dir, "rev-parse", "HEAD")
}

// CheckGitRepoWithTimeout verifies git repository with a default timeout.
// This is a convenience function that creates a context with GitCheckTimeout.
func CheckGitRepoWithTimeout(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), GitCheckTimeout)
	defer cancel()

	return CheckGitRepo(ctx, dir)
}
