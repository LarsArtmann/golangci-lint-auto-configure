package config

// mergeIssuesConfig merges issues configurations.
func (cm *Merger) mergeIssuesConfig(primary, secondary *IssuesConfig) int {
	changes := 0

	// Note: 0 is a valid value (disable), so we check if primary hasn't been explicitly set
	// We use a heuristic: if both MaxIssuesPerLinter and MaxSameIssues are 0, assume unset
	if primary.MaxIssuesPerLinter == 0 && secondary.MaxIssuesPerLinter != 0 {
		primary.MaxIssuesPerLinter = secondary.MaxIssuesPerLinter
		changes++
	}

	if primary.MaxSameIssues == 0 && secondary.MaxSameIssues != 0 {
		primary.MaxSameIssues = secondary.MaxSameIssues
		changes++
	}

	if primary.NewFromRev == "" && secondary.NewFromRev != "" {
		primary.NewFromRev = secondary.NewFromRev
		changes++
	}

	if primary.NewFromPatch == "" && secondary.NewFromPatch != "" {
		primary.NewFromPatch = secondary.NewFromPatch
		changes++
	}

	if primary.NewFromMergeBase == "" && secondary.NewFromMergeBase != "" {
		primary.NewFromMergeBase = secondary.NewFromMergeBase
		changes++
	}

	if !primary.New && secondary.New {
		primary.New = secondary.New
		changes++
	}

	if !primary.WholeFiles && secondary.WholeFiles {
		primary.WholeFiles = secondary.WholeFiles
		changes++
	}

	if !primary.Fix && secondary.Fix {
		primary.Fix = secondary.Fix
		changes++
	}

	if !primary.UniqByLine && secondary.UniqByLine {
		// Only set if secondary is true (default is usually true)
		primary.UniqByLine = secondary.UniqByLine
		changes++
	}

	return changes
}
