package finding

import (
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Rule IDs for desired-state deviations. One rule per config path the
// reconciler plans against: a finding's rule answers "what kind of drift",
// its message answers "what exactly", its suggestion answers "why".
const (
	RuleIDDeviationLinterEnable   = "deviation-linter-enable"
	RuleIDDeviationLinterDisable  = "deviation-linter-disable"
	RuleIDDeviationLinterSettings = "deviation-linter-settings"
	RuleIDDeviationFormatter      = "deviation-formatter-enable"
	RuleIDDeviationExclusionRules = "deviation-exclusion-rules"
	RuleIDDeviationDefaults       = "deviation-config-defaults"
)

// DeviationsToFindings converts the reconciler's planned changes into
// findings. Findings = diff: every change the converge engine would apply IS
// one finding, built from the same Change objects the apply path executes —
// detect/repair drift is structurally impossible because lint mode and apply
// mode speak about identical objects.
func DeviationsToFindings(changes []linter.Change, configPath string) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0, len(changes))

	for _, change := range changes {
		f, err := deviationFinding(change, configPath)
		if err != nil {
			return nil, err
		}

		findings = append(findings, f)
	}

	return findings, nil
}

func deviationFinding(change linter.Change, configPath string) (finding.Finding, error) {
	found, err := configFinding(configFindingParams{
		RuleID:     deviationRuleID(change.Path),
		Message:    deviationMessage(change),
		Severity:   deviationSeverity(change.Kind),
		Position:   configPosition(configPath, 0),
		Category:   deviationCategory(change),
		Tags:       deviationTags(change.Element),
		Suggestion: change.Reason,
		BeforeCode: nonEmptyValue(change.From),
		AfterCode:  nonEmptyValue(change.To),
	})
	if err != nil {
		return finding.Finding{}, errorfamily.WrapCorruptionf(err, "deviation.finding",
			"build finding for change %s", change.Key())
	}

	return found, nil
}

// deviationRuleID maps the config path to the finding rule: the rule names
// the drift site, not the mutation kind (the message carries that).
func deviationRuleID(path string) string {
	switch path {
	case linter.PathLintersEnable:
		return RuleIDDeviationLinterEnable
	case linter.PathLintersDisable:
		return RuleIDDeviationLinterDisable
	case linter.PathFormattersEnable:
		return RuleIDDeviationFormatter
	case linter.PathLintersExclusionRules:
		return RuleIDDeviationExclusionRules
	case linter.PathLintersSettings:
		return RuleIDDeviationLinterSettings
	default:
		if hasPathPrefix(path, linter.PathLintersSettings+".") {
			return RuleIDDeviationLinterSettings
		}

		return RuleIDDeviationDefaults
	}
}

func hasPathPrefix(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}

// deviationSeverity maps the mutation kind: additions are recommendations
// (info), updates and removals mean an existing value is wrong (warning).
func deviationSeverity(kind linter.ChangeKind) finding.Severity {
	if kind == linter.ChangeAdd {
		return finding.SeverityInfo
	}

	return finding.SeverityWarning
}

func deviationMessage(change linter.Change) string {
	subject := change.Element
	if subject == "" {
		subject = "value"
	}

	switch change.Kind {
	case linter.ChangeAdd:
		return fmt.Sprintf("%s is missing from %s", subject, change.Path)
	case linter.ChangeUpdate:
		return fmt.Sprintf("%s in %s deviates from the desired value", subject, change.Path)
	case linter.ChangeRemove:
		return fmt.Sprintf("%s in %s is stale and should be removed", subject, change.Path)
	default:
		return fmt.Sprintf("%s in %s deviates from the desired state", subject, change.Path)
	}
}

func deviationCategory(change linter.Change) finding.Category {
	if change.Path == linter.PathFormattersEnable {
		return finding.CategoryStyle
	}

	if name, ok := linterName(change); ok {
		return LinterNameToCategory(name)
	}

	return finding.CategoryConfiguration
}

// linterName extracts the linter identity of a change: the element for list
// memberships, the first path segment after linters.settings for settings
// blocks, nothing for scalar paths.
func linterName(change linter.Change) (types.LinterName, bool) {
	if change.Element == "" {
		return "", false
	}

	if hasPathPrefix(change.Path, linter.PathLintersSettings+".") {
		return types.LinterName(change.Element), true
	}

	switch change.Path {
	case linter.PathLintersEnable, linter.PathLintersDisable, linter.PathLintersSettings:
		return types.LinterName(change.Element), true
	default:
		return "", false
	}
}

func deviationTags(element string) []finding.Tag {
	tag := linterTag(element)
	if !tagSafe(string(tag)) {
		return []finding.Tag{}
	}

	return []finding.Tag{tag}
}

// tagSafe reports whether the tag satisfies go-finding's lowercase-hyphenated
// convention (first char a-z, remaining chars a-z0-9-). Callers pass the
// TRANSFORMED tag (linterTag output), not the raw config element: rename
// elements like `exhaustruct->exhaustruct_v5` normalize to
// `exhaustruct->exhaustruct-v5`, whose `>` is invalid — such elements carry no
// tag instead of crashing finding construction.
func tagSafe(tag string) bool {
	if tag == "" {
		return false
	}

	for i := 0; i < len(tag); i++ {
		c := tag[i]

		switch {
		case c >= 'a' && c <= 'z':
		case i > 0 && (c >= '0' && c <= '9' || c == '-'):
		default:
			return false
		}
	}

	return true
}

func nonEmptyValue(v any) string {
	if v == nil {
		return ""
	}

	if s, isString := v.(string); isString {
		return s
	}

	return fmt.Sprint(v)
}
