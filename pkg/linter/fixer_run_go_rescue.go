package linter

import (
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// rescueOverspecifiedRunGo repairs a config whose run.go targets a Go version
// newer than the one the installed golangci-lint binary was built with.
// golangci-lint refuses to load such a config at all ("the Go language version
// used to build golangci-lint is lower than the targeted Go version"), so the
// analysis phase would fail before the fixer ever gets a chance to heal the
// config. This runs before analysis: in normal mode it rewrites run.go to the
// binary's build Go version (audited as ActionRescuedRunGo); in dry-run mode
// the file must stay untouched, so a classified error explains the way out.
func (f *Fixer) rescueOverspecifiedRunGo(cfg *types.Config, configPath string, dryRun bool) error {
	if f.analyzer == nil {
		return nil
	}

	golangciLintGoVersion := f.analyzer.GetDetectedGoVersion()
	if !runGoExceedsBinary(cfg.Run.Go, golangciLintGoVersion) {
		return nil
	}

	configGoVersion, _ := types.NormalizeGoMajorMinor(cfg.Run.Go)

	if dryRun {
		return dryRunRescueError(configGoVersion, golangciLintGoVersion, configPath)
	}

	return f.applyRunGoRescue(cfg, configPath, golangciLintGoVersion)
}

// runGoExceedsBinary reports whether the config's run.go is a parsable Go
// version strictly newer than the version golangci-lint was built with.
// Empty or unparsable values never exceed (golangci-lint owns those errors).
func runGoExceedsBinary(runGo, golangciLintGoVersion string) bool {
	if runGo == "" || golangciLintGoVersion == "" {
		return false
	}

	configGoVersion, ok := types.NormalizeGoMajorMinor(runGo)

	return ok && types.CompareGoMajorMinor(configGoVersion, golangciLintGoVersion) > 0
}

// dryRunRescueError explains, without touching the file, why the config is
// unloadable and how to proceed.
func dryRunRescueError(configGoVersion, golangciLintGoVersion, configPath string) error {
	return apperrors.NewAnalysisError(
		fmt.Sprintf("run.go %s is newer than the Go used to build golangci-lint (%s)",
			configGoVersion, golangciLintGoVersion),
		configPath,
		errorfamily.WrapRejectionf(apperrors.ErrRunGoNewerThanBinary, "config.run_go.newer_than_binary",
			"run.go %s vs binary built with go%s (dry run: not repaired)",
			configGoVersion, golangciLintGoVersion),
	)
}

// applyRunGoRescue rewrites run.go to the binary's build Go version, persists
// the config, and records the mutation in the audit ledger.
func (f *Fixer) applyRunGoRescue(cfg *types.Config, configPath, golangciLintGoVersion string) error {
	f.logger.Warnf(
		"run.go %s is newer than the Go used to build golangci-lint (%s); "+
			"golangci-lint cannot load this config at all — repairing run.go to %s",
		cfg.Run.Go, golangciLintGoVersion, golangciLintGoVersion,
	)

	previous := cfg.Run.Go
	cfg.Run.Go = golangciLintGoVersion

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.rescue_run_go",
			"saving rescued config %s", configPath)
	}

	f.ledger.Record(audit.ActionRescuedRunGo, "run.go",
		fmt.Sprintf("%s -> %s (golangci-lint built with go%s)",
			previous, golangciLintGoVersion, golangciLintGoVersion))

	return nil
}
