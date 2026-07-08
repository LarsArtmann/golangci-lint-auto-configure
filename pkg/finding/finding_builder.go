package finding

import (
	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
)

// buildFinding is a helper that builds a Finding from a Builder, returning an error
// instead of panicking on invalid builder state.
func buildFinding(b *finding.Builder) (finding.Finding, error) {
	found, err := b.Build()
	if err != nil {
		return finding.Finding{}, errorfamily.WrapCorruption(err, "finding.builder",
			"finding builder error")
	}

	return found, nil
}

// configPosition creates a finding.Position for a config-level finding.
// go-finding v1.1.0 requires Line > 0; config-level findings use Line: 1.
func configPosition(path string, line int) finding.Position {
	if line <= 0 {
		line = 1
	}

	return finding.Position{File: finding.FilePath(path), Line: line}
}

type configFindingParams struct {
	RuleID     string
	Message    string
	Severity   finding.Severity
	Position   finding.Position
	Category   finding.Category
	Tags       []finding.Tag
	Suggestion string
}

func configFinding(params configFindingParams) (finding.Finding, error) {
	builder := finding.NewBuilder(
		finding.RuleName(params.RuleID),
		finding.ToolName(constants.ToolName),
		params.Message,
		params.Severity,
		params.Position,
	)

	if len(params.Tags) > 0 {
		builder = builder.WithTags(params.Tags...)
	}

	if params.Category != "" {
		builder = builder.WithCategory(params.Category)
	}

	builder = builder.WithFixStrategy(finding.FixStrategySuggest)

	if params.Suggestion != "" {
		builder = builder.WithSuggestion(params.Suggestion)
	}

	found, err := builder.Build()
	if err != nil {
		return finding.Finding{}, errorfamily.WrapCorruptionf(err, "finding.build",
			"build finding for %s", params.RuleID)
	}

	return found, nil
}
