package config

// mergeRunConfig merges run configurations. Uses multiple conditionals for config fields.
func (cm *Merger) mergeRunConfig(primary, secondary *RunConfig) int {
	changes := 0

	// Only merge if primary has zero values and secondary has values
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

	if !primary.AllowParallelRunners && secondary.AllowParallelRunners {
		primary.AllowParallelRunners = secondary.AllowParallelRunners
		changes++
	}

	if !primary.AllowSerialRunners && secondary.AllowSerialRunners {
		primary.AllowSerialRunners = secondary.AllowSerialRunners
		changes++
	}

	if primary.IssuesExitCode == 0 && secondary.IssuesExitCode != 0 {
		primary.IssuesExitCode = secondary.IssuesExitCode
		changes++
	}

	if !primary.Tests && secondary.Tests {
		primary.Tests = secondary.Tests
		changes++
	}

	if primary.Concurrency == 0 && secondary.Concurrency != 0 {
		primary.Concurrency = secondary.Concurrency
		changes++
	}

	if primary.RelativePathMode == "" && secondary.RelativePathMode != "" {
		primary.RelativePathMode = secondary.RelativePathMode
		changes++
	}

	return changes
}
