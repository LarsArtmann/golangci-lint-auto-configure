package config

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// mergeRunConfig merges run configurations. Uses multiple conditionals for config fields.
func (cm *Merger) mergeRunConfig(primary, secondary *types.RunConfig) int {
	changes := 0

	changes += mergeRunStringFields(primary, secondary)
	changes += mergeRunBoolFields(primary, secondary)
	changes += mergeRunNumericFields(primary, secondary)

	return changes
}

func mergeRunStringFields(primary, secondary *types.RunConfig) int {
	changes := 0

	if primary.Timeout == "" && secondary.Timeout != "" {
		primary.Timeout = secondary.Timeout
		changes++
	}

	if primary.Go == "" && secondary.Go != "" {
		primary.Go = secondary.Go
		changes++
	}

	if len(primary.BuildTags) == 0 && len(secondary.BuildTags) > 0 {
		primary.BuildTags = secondary.BuildTags
		changes++
	}

	if primary.ModulesDownloadMode == "" && secondary.ModulesDownloadMode != "" {
		primary.ModulesDownloadMode = secondary.ModulesDownloadMode
		changes++
	}

	if primary.RelativePathMode == "" && secondary.RelativePathMode != "" {
		primary.RelativePathMode = secondary.RelativePathMode
		changes++
	}

	return changes
}

func mergeRunBoolFields(primary, secondary *types.RunConfig) int {
	changes := 0

	if !primary.AllowParallelRunners && secondary.AllowParallelRunners {
		primary.AllowParallelRunners = secondary.AllowParallelRunners
		changes++
	}

	if !primary.AllowSerialRunners && secondary.AllowSerialRunners {
		primary.AllowSerialRunners = secondary.AllowSerialRunners
		changes++
	}

	if !primary.Tests && secondary.Tests {
		primary.Tests = secondary.Tests
		changes++
	}

	return changes
}

func mergeRunNumericFields(primary, secondary *types.RunConfig) int {
	changes := 0

	if primary.IssuesExitCode == 0 && secondary.IssuesExitCode != 0 {
		primary.IssuesExitCode = secondary.IssuesExitCode
		changes++
	}

	if primary.Concurrency == 0 && secondary.Concurrency != 0 {
		primary.Concurrency = secondary.Concurrency
		changes++
	}

	return changes
}
