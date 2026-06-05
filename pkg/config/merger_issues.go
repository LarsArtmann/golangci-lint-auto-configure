package config

// mergeIssuesConfig merges issues configurations.
func (cm *Merger) mergeIssuesConfig(primary, secondary *IssuesConfig) int {
	changes := 0

	changes += mergeIssuesNumericFields(primary, secondary)
	changes += mergeIssuesStringFields(primary, secondary)
	changes += mergeIssuesBoolFields(primary, secondary)

	return changes
}

// Zero is a valid value (disable), so we check if primary hasn't been explicitly set.
func mergeIssuesNumericFields(primary, secondary *IssuesConfig) int {
	changes := 0

	if primary.MaxIssuesPerLinter == 0 && secondary.MaxIssuesPerLinter != 0 {
		primary.MaxIssuesPerLinter = secondary.MaxIssuesPerLinter
		changes++
	}

	if primary.MaxSameIssues == 0 && secondary.MaxSameIssues != 0 {
		primary.MaxSameIssues = secondary.MaxSameIssues
		changes++
	}

	return changes
}

func mergeIssuesStringFields(primary, secondary *IssuesConfig) int {
	changes := 0

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

	return changes
}

func mergeIssuesBoolFields(primary, secondary *IssuesConfig) int {
	changes := 0

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
		primary.UniqByLine = secondary.UniqByLine
		changes++
	}

	return changes
}
