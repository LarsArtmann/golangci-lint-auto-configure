package linter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	golangciLintGoVersion := f.analyzer.DetectedBuildGoVersion()

	f.warnIfGoModExceedsBinary(configPath, golangciLintGoVersion)

	if !runGoExceedsBinary(cfg.Run.Go, golangciLintGoVersion) {
		return nil
	}

	configGoVersion, _ := types.ParseGoVersion(cfg.Run.Go)

	if dryRun {
		return dryRunRescueError(configGoVersion, golangciLintGoVersion, configPath)
	}

	return f.applyRunGoRescue(cfg, configPath, golangciLintGoVersion)
}

// runGoExceedsBinary reports whether the config's run.go is a parsable Go
// version strictly newer than the version golangci-lint was built with.
// Empty or unparsable values never exceed (golangci-lint owns those errors).
func runGoExceedsBinary(runGo string, golangciLintGoVersion types.GoVersion) bool {
	if runGo == "" || golangciLintGoVersion.IsZero() {
		return false
	}

	configGoVersion, err := types.ParseGoVersion(runGo)
	if err != nil {
		return false
	}

	return configGoVersion.Compare(golangciLintGoVersion) > 0
}

// warnIfGoModExceedsBinary warns when the project's go.mod go-directive targets
// a newer Go than the installed golangci-lint was built with. The go directive
// is the project's declared language version and is never rewritten — the only
// fix is upgrading golangci-lint — so this stays a warning in every mode
// (including --check/--dry-run), never an error and never a mutation.
// Best-effort: a missing or unparsable go.mod is silent.
func (f *Fixer) warnIfGoModExceedsBinary(configPath string, golangciLintGoVersion types.GoVersion) {
	directive := readGoModGoDirective(configPath)
	if directive.IsZero() || golangciLintGoVersion.IsZero() {
		return
	}

	if directive.Compare(golangciLintGoVersion) <= 0 {
		return
	}

	f.logger.Warnf(
		"go.mod declares go %s, newer than the Go used to build golangci-lint (%s); "+
			"golangci-lint may fail to analyze packages using newer syntax — upgrade golangci-lint "+
			"(go.mod is never rewritten by this tool)",
		directive, golangciLintGoVersion,
	)
}

// readGoModGoDirective returns the go-directive version from the go.mod next to
// the config, normalized to major.minor, or the zero value when the file is
// absent, unreadable, or carries no parsable directive. Directives inside
// parenthesized blocks (require/exclude/replace) are ignored — the go directive
// is a top-level statement.
func readGoModGoDirective(configPath string) types.GoVersion {
	goModPath := filepath.Join(filepath.Dir(configPath), "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return types.GoVersion{}
	}

	defer func() { _ = file.Close() }()

	return scanGoDirective(bufio.NewScanner(file))
}

func scanGoDirective(scanner *bufio.Scanner) types.GoVersion {
	inBlock := false

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "(":
			inBlock = true
		case ")":
			inBlock = false
		case "go":
			if !inBlock && len(fields) >= 2 {
				if parsed, err := types.ParseGoVersion(fields[1]); err == nil {
					return parsed
				}
			}
		}
	}

	return types.GoVersion{}
}

// dryRunRescueError explains, without touching the file, why the config is
// unloadable and how to proceed.
func dryRunRescueError(configGoVersion, golangciLintGoVersion types.GoVersion, configPath string) error {
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
func (f *Fixer) applyRunGoRescue(cfg *types.Config, configPath string, golangciLintGoVersion types.GoVersion) error {
	f.logger.Warnf(
		"run.go %s is newer than the Go used to build golangci-lint (%s); "+
			"golangci-lint cannot load this config at all — repairing run.go to %s",
		cfg.Run.Go, golangciLintGoVersion, golangciLintGoVersion,
	)

	previous := cfg.Run.Go
	cfg.Run.Go = golangciLintGoVersion.String()

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.rescue_run_go",
			"saving rescued config %s", configPath)
	}

	f.ledger.Record(audit.ActionRescuedRunGo, "run.go",
		fmt.Sprintf("%s -> %s (golangci-lint built with go%s)",
			previous, golangciLintGoVersion, golangciLintGoVersion))

	return nil
}
