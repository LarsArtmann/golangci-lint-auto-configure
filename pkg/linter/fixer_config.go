package linter

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/gogenfilter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// GoVersionProvider returns the locally installed Go version.
type GoVersionProvider func(ctx context.Context) string

// ApplyGeneratedExclusions injects default exclusion paths (templ, vendor) and
// scans the project for auto-generated Go files to add additional exclusion
// patterns. This is the public entry point used by both the fixer flow and the
// preset flow. Uses gogenfilter for two-phase detection (filename first,
// content second).
func ApplyGeneratedExclusions(logger *log.Logger, cfg *types.Config, configPath string) int {
	projectDir := filepath.Dir(configPath)

	if cfg.Linters.Exclusions.Generated == "" {
		logger.Infof("Setting linters.exclusions.generated to \"lax\"")

		cfg.Linters.Exclusions.Generated = "lax"
	}

	if cfg.Formatters.Exclusions.Generated == "" {
		logger.Infof("Setting formatters.exclusions.generated to \"lax\"")

		cfg.Formatters.Exclusions.Generated = "lax"
	}

	totalAdded := 0

	linterDefaults := mergeExclusionPaths(
		&cfg.Linters.Exclusions.Paths, constants.DefaultLinterExclusionPaths, logger, "linters",
	)
	formatterDefaults := mergeExclusionPaths(
		&cfg.Formatters.Exclusions.Paths, constants.DefaultFormatterExclusionPaths, logger, "formatters",
	)
	totalAdded += linterDefaults + formatterDefaults

	result, err := gogenfilter.ScanProject(os.DirFS(projectDir))
	if err != nil {
		logger.Debugf("Generated file scan skipped: %v", err)

		return totalAdded
	}

	if len(result.Exclusions) > 0 {
		logger.Infof(
			"Found %d generated file types (%d files scanned, %d generated): %v",
			len(result.Generators), result.ScannedFiles, result.GeneratedFiles, result.Generators,
		)

		newPaths := gogenfilter.ExclusionPaths(result.Exclusions)

		totalAdded += mergeExclusionPaths(&cfg.Linters.Exclusions.Paths, newPaths, logger, "linters")
		totalAdded += mergeExclusionPaths(&cfg.Formatters.Exclusions.Paths, newPaths, logger, "formatters")
	}

	return totalAdded
}

func mergeExclusionPaths(existing *[]string, newPaths []string, logger *log.Logger, section string) int {
	if len(newPaths) == 0 {
		return 0
	}

	merged := gogenfilter.MergeExclusionPaths(*existing, newPaths)
	added := len(merged) - len(*existing)

	if added > 0 {
		logger.Infof("Adding %d generated file exclusions to %s: %v", added, section, newPaths)

		*existing = merged
	}

	return added
}
