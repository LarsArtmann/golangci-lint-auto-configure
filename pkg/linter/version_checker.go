package linter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"golang.org/x/mod/semver"
)

// golangciLintVersion represents JSON output from `golangci-lint version --json`.
type golangciLintVersion struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

// validateVersion checks if the version meets minimum requirements.
// Returns error if version is too old, nil if valid.
func (a *Analyzer) validateVersion(version string) error {
	version = strings.TrimPrefix(version, "v")
	version = "v" + version

	if !semver.IsValid(version) {
		return apperrors.NewAnalysisError(
			"invalid golangci-lint version format",
			"",
			errorfamily.WrapCorruptionf(apperrors.ErrInvalidVersionFormat, "version.invalid_format",
				"version %s", version),
		)
	}

	minVersion := constants.MinGolangCILintVersion
	if semver.Compare(version, minVersion) < 0 {
		return apperrors.NewAnalysisError(
			fmt.Sprintf("golangci-lint version %s is too old", version),
			"",
			errorfamily.WrapRejectionf(apperrors.ErrVersionTooOld, "version.too_old",
				"minimum required version is %s. Please upgrade: https://golangci-lint.run/usage/install/",
				minVersion),
		)
	}

	a.logger.Debugf("golangci-lint version %s (>= %s) ✓", version, minVersion)
	a.detectedVersion = version
	a.warnUnexpectedVersion(version)

	return nil
}

// warnUnexpectedVersion warns when the detected version differs from the tested version.
func (a *Analyzer) warnUnexpectedVersion(version string) {
	if version == constants.ExpectedGolangCILintVersion {
		return
	}

	a.logger.Warnf(
		"golangci-lint version %s detected but %s is recommended — "+
			"unexpected versions may have behavioral differences",
		version, constants.ExpectedGolangCILintVersion,
	)
}

// CheckVersion verifies golangci-lint is at least the minimum required version.
func (a *Analyzer) CheckVersion(ctx context.Context) error {
	// Try JSON output first (more reliable)
	output, err := a.runVersionCommandWithRetry(ctx, "version", "--json")
	if err != nil {
		// Fallback to text parsing if --json not supported or parallel running
		return a.checkVersionText(ctx)
	}

	// Parse JSON output
	var versionInfo golangciLintVersion
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		// JSON parsing failed, fall back to text parsing
		a.logger.Debugf("Failed to parse JSON version output, falling back to text: %v", err)

		return a.checkVersionText(ctx)
	}

	if versionInfo.Version == "" {
		return apperrors.NewAnalysisError(
			"could not parse golangci-lint version from JSON",
			"",
			errorfamily.WrapCorruptionf(apperrors.ErrVersionParse, "version.parse_json",
				"output: %s", string(output)),
		)
	}

	return a.validateVersion(versionInfo.Version)
}

// runVersionCommandWithRetry runs a version command with retry for parallel running errors.
func (a *Analyzer) runVersionCommandWithRetry(ctx context.Context, args ...string) ([]byte, error) {
	output, err := a.runWithRetry(ctx, "version check", a.commandOperation(ctx, args...))
	if err != nil {
		return output, apperrors.NewAnalysisError(
			"version check command failed",
			"",
			apperrors.WrapClassifiedf(err, "version.command_failed",
				"command %v failed", args),
		)
	}

	return output, nil
}

// checkVersionText is a fallback that parses text output from golangci-lint --version
// Used when --json flag is not available or fails.
func (a *Analyzer) checkVersionText(ctx context.Context) error {
	output, err := a.runVersionCommandWithRetry(ctx, "--version")
	if err != nil {
		return apperrors.NewAnalysisError("failed to check golangci-lint version", "", err)
	}

	// Parse version from text output (format: "golangci-lint has version X.Y.Z built with...")
	outputStr := string(output)

	version := a.parseVersionText(outputStr)

	if version == "" {
		return apperrors.NewAnalysisError(
			"could not parse golangci-lint version from output",
			"",
			errorfamily.WrapCorruptionf(apperrors.ErrVersionParse, "version.parse_text",
				"output: %s", outputStr),
		)
	}

	return a.validateVersion(version)
}

// parseVersionText extracts version number from text output
// Format: "golangci-lint has version X.Y.Z built with...".
func (a *Analyzer) parseVersionText(output string) string {
	parts := strings.Fields(output)
	for i, part := range parts {
		if part == "version" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}
