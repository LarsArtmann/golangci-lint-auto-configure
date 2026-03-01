package linter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"golang.org/x/mod/semver"
)

// golangciLintVersion represents JSON output from `golangci-lint version --json`.
type golangciLintVersion struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

// CheckVersion verifies golangci-lint is at least the minimum required version.
func (a *Analyzer) CheckVersion() error {
	// Try JSON output first (more reliable)
	cmd := exec.Command(a.golangciLintPath, "version", "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to text parsing if --json not supported
		return a.checkVersionText()
	}

	// Parse JSON output
	var versionInfo golangciLintVersion
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		// JSON parsing failed, fall back to text parsing
		a.logger.Debugf("Failed to parse JSON version output, falling back to text: %v", err)

		return a.checkVersionText()
	}

	if versionInfo.Version == "" {
		return errors.NewAnalysisError(
			"could not parse golangci-lint version from JSON",
			"",
			fmt.Errorf("output: %s", string(output)),
		)
	}

	version := versionInfo.Version

	// Ensure version has 'v' prefix for semver
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Validate semver format
	if !semver.IsValid(version) {
		return errors.NewAnalysisError("invalid golangci-lint version format", "", fmt.Errorf("version: %s", version))
	}

	// Compare with minimum required version
	minVersion := constants.MinGolangCILintVersion
	if semver.Compare(version, minVersion) < 0 {
		return errors.NewAnalysisError(
			fmt.Sprintf("golangci-lint version %s is too old", version),
			"",
			fmt.Errorf(
				"minimum required version is %s. Please upgrade: https://golangci-lint.run/usage/install/",
				minVersion,
			),
		)
	}

	a.logger.Debugf("golangci-lint version %s (>= %s) ✓", version, minVersion)

	return nil
}

// checkVersionText is a fallback that parses text output from golangci-lint --version
// Used when --json flag is not available or fails.
func (a *Analyzer) checkVersionText() error {
	cmd := exec.Command(a.golangciLintPath, "--version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.NewAnalysisError("failed to check golangci-lint version", "", err)
	}

	// Parse version from text output (format: "golangci-lint has version X.Y.Z built with...")
	outputStr := string(output)

	version := a.parseVersionText(outputStr)

	if version == "" {
		return errors.NewAnalysisError(
			"could not parse golangci-lint version from output",
			"",
			fmt.Errorf("output: %s", outputStr),
		)
	}

	// Ensure version has 'v' prefix for semver
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Validate semver format
	if !semver.IsValid(version) {
		return errors.NewAnalysisError("invalid golangci-lint version format", "", fmt.Errorf("version: %s", version))
	}

	// Compare with minimum required version
	minVersion := constants.MinGolangCILintVersion
	if semver.Compare(version, minVersion) < 0 {
		return errors.NewAnalysisError(
			fmt.Sprintf("golangci-lint version %s is too old", version),
			"",
			fmt.Errorf(
				"minimum required version is %s. Please upgrade: https://golangci-lint.run/usage/install/",
				minVersion,
			),
		)
	}

	a.logger.Debugf("golangci-lint version %s (>= %s) ✓", version, minVersion)

	return nil
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
