package linter

import (
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// DesiredConfig is the desired state the policy table renders for one
// environment: the complete linter/formatter sets, their settings, and the
// structural defaults. Pure data — rendering is deterministic and contains no
// I/O, so the same (policy, env) pair always renders identically.
type DesiredConfig struct {
	// EnableLinters is the desired linters.enable set (sorted).
	EnableLinters []types.LinterName
	// EnableFormatters is the desired formatters.enable set (canonical order).
	EnableFormatters []types.FormatterName
	// LinterSettings maps each desired linter to its valid settings blocks.
	LinterSettings map[types.LinterName][]types.VersionedSettings
	// FormatterSettings maps each desired formatter to its settings blocks.
	FormatterSettings map[types.FormatterName][]types.VersionedSettings
	// LinterExclusionPaths are the static default exclusion paths.
	LinterExclusionPaths []string
	// FormatterExclusionPaths are the static formatter exclusion paths.
	FormatterExclusionPaths []string
	// GeneratedExclusionPaths are the env-detected generated-file patterns.
	GeneratedExclusionPaths []string
	// ExclusionRules are the default test-file exclusion rules.
	ExclusionRules []types.ExclusionRuleConfig
	// RunDefaults are the run-section defaults the tool enforces.
	RunDefaults RunDefaults
	// IssuesDefaults are the issues-section defaults.
	IssuesDefaults IssuesDefaults
}

// RunDefaults mirrors the run-section values the tool injects.
type RunDefaults struct {
	Timeout             string
	Go                  string // major.minor, "" when the local Go is unknown
	BuildTags           []string
	AllowParallelRunner bool
	AllowSerialRunner   bool
	IssuesExitCode      int
}

// IssuesDefaults mirrors the issues-section values the tool injects.
type IssuesDefaults struct {
	MaxIssuesPerLinter int
	MaxSameIssues      int
}

// DesiredOption customizes the render. Options are the renderer's only
// inputs beyond (policy, env) — they carry user intent (pragmatic opt-out,
// priority threshold), never environment facts.
type DesiredOption func(*desiredRequest)

type desiredRequest struct {
	pragmatic         bool
	minPriority       types.LinterPriority
	tool              types.GolangciVersion
	localGo           types.GoVersion
	projectGo         types.GoVersion
	envTech           func(string) bool
	enabledFormatters []types.FormatterName
}

// WithPragmatic drops the pragmatic-noise tier from the enable set.
func WithPragmatic(pragmatic bool) DesiredOption {
	return func(r *desiredRequest) { r.pragmatic = pragmatic }
}

// WithMinPriority only recommends linters at or above the given priority
// (the configure command's priority threshold).
func WithMinPriority(priority types.LinterPriority) DesiredOption {
	return func(r *desiredRequest) { r.minPriority = priority }
}

// WithEnvTool pins the tool version for validity gating (defaults to env.Tool).
func WithEnvTool(tool types.GolangciVersion) DesiredOption {
	return func(r *desiredRequest) { r.tool = tool }
}

// WithLocalGo pins the local Go version for run.go rendering.
func WithLocalGo(version types.GoVersion) DesiredOption {
	return func(r *desiredRequest) { r.localGo = version }
}

// Desired renders the desired configuration for the environment. This is the
// pure heart of the reconciler: policy (the table) + environment (facts) in,
// desired state out. No I/O, no mutation of package data, fully deterministic.
func Desired(env Env, options ...DesiredOption) DesiredConfig {
	request := desiredRequest{
		pragmatic:         false,
		minPriority:       types.LinterPriorityOptional,
		tool:              env.Tool,
		localGo:           env.LocalGo,
		projectGo:         env.ProjectGo,
		envTech:           env.UsesTech,
		enabledFormatters: []types.FormatterName{},
	}

	for _, apply := range options {
		apply(&request)
	}

	desired := DesiredConfig{
		EnableLinters:           renderEnableSet(request),
		EnableFormatters:        renderFormatterSet(request),
		LinterSettings:          renderLinterSettings(request),
		FormatterSettings:       renderFormatterSettings(request),
		LinterExclusionPaths:    append([]string{}, constants.DefaultLinterExclusionPaths...),
		FormatterExclusionPaths: append([]string{}, constants.DefaultFormatterExclusionPaths...),
		GeneratedExclusionPaths: append([]string{}, env.Generated...),
		ExclusionRules:          append([]types.ExclusionRuleConfig{}, constants.DefaultExclusionRules...),
		RunDefaults:             renderRunDefaults(env, request),
		IssuesDefaults: IssuesDefaults{
			MaxIssuesPerLinter: constants.DefaultMaxIssuesPerLinter,
			MaxSameIssues:      constants.DefaultMaxSameIssues,
		},
	}

	return desired
}

// renderEnableSet computes the desired linters.enable: every recommendable
// record whose validity interval contains the tool version, whose language
// interval contains the project's Go, whose gate passes, and which the tier
// and priority filters admit.
func renderEnableSet(request desiredRequest) []types.LinterName {
	var enabled []types.LinterName

	for _, record := range constants.LinterPolicies {
		if !shouldRender(record, request) {
			continue
		}

		enabled = append(enabled, record.Name)
	}

	sort.Slice(enabled, func(i, j int) bool { return enabled[i] < enabled[j] })

	return enabled
}

// shouldRender evaluates one record against the environment. Zero tool
// version fails validity checks open (an unknown tool gates nothing) — the
// same conservative behavior the categorizer applies today.
func shouldRender(record types.LinterPolicy, request desiredRequest) bool {
	switch record.Tier.Kind {
	case types.TierDisabled:
		return false
	case types.TierNeverAutoEnable:
		return false
	case types.TierPragmaticNoise:
		if request.pragmatic {
			return false
		}
	case types.TierRecommendable:
	}

	if record.Priority > request.minPriority {
		return false
	}

	if !record.ValidTool.IsUnbounded() && !record.ValidTool.Contains(request.tool) {
		return false
	}

	if !record.Lang.IsUnbounded() && !record.Lang.Contains(projectGoOf(request)) {
		return false
	}

	if record.Gate.Kind == types.GateRequiresTech {
		if !techAvailable(request, record.Gate.Tech) {
			return false
		}
	}

	return true
}

func techAvailable(request desiredRequest, tech string) bool {
	return request.envTech == nil || request.envTech(tech)
}

func projectGoOf(request desiredRequest) types.GoVersion {
	return request.projectGo
}

func renderFormatterSet(request desiredRequest) []types.FormatterName {
	ordered := make([]types.FormatterName, 0, len(constants.CoreFormatters))

	for _, name := range constants.FormatterOrder {
		record, known := constants.FormatterPolicyRecord(name)
		if !known {
			continue
		}

		if !record.Core {
			continue
		}

		if record.Gate.Kind == types.GateRequiresTech && !techAvailable(request, record.Gate.Tech) {
			continue
		}

		ordered = append(ordered, name)
	}

	return ordered
}

// renderLinterSettings collects the version-valid settings blocks of every
// rendered linter. This is the data min-version DERIVES from: the maximum
// Since over rendered blocks is the oldest tool that can load the config.
func renderLinterSettings(request desiredRequest) map[types.LinterName][]types.VersionedSettings {
	settings := make(map[types.LinterName][]types.VersionedSettings)

	for _, name := range renderEnableSet(request) {
		record, known := constants.PolicyRecord(name)
		if !known || len(record.Settings) == 0 {
			continue
		}

		valid := make([]types.VersionedSettings, 0, len(record.Settings))

		for _, block := range record.Settings {
			if !block.Since.IsZero() && !request.tool.IsZero() && block.Since.Compare(request.tool) > 0 {
				continue
			}

			valid = append(valid, block)
		}

		if len(valid) > 0 {
			settings[name] = valid
		}
	}

	return settings
}

func renderFormatterSettings(request desiredRequest) map[types.FormatterName][]types.VersionedSettings {
	settings := make(map[types.FormatterName][]types.VersionedSettings)

	for _, name := range renderFormatterSet(request) {
		record, known := constants.FormatterPolicyRecord(name)
		if !known || len(record.Settings) == 0 {
			continue
		}

		settings[name] = append([]types.VersionedSettings{}, record.Settings...)
	}

	return settings
}

func renderRunDefaults(env Env, request desiredRequest) RunDefaults {
	defaults := RunDefaults{
		Timeout:             constants.DefaultTimeout,
		Go:                  request.localGo.String(),
		BuildTags:           constants.GoExperimentTagsFor(env.LocalGo.String()),
		AllowParallelRunner: true,
		AllowSerialRunner:   true,
		IssuesExitCode:      1,
	}

	// The run.go cap: a local Go newer than the binary's build Go is capped
	// (golangci-lint refuses newer run.go values outright).
	if !request.localGo.IsZero() && !env.ToolBuildGo.IsZero() && request.localGo.Compare(env.ToolBuildGo) > 0 {
		defaults.Go = env.ToolBuildGo.String()
	}

	sort.Strings(defaults.BuildTags)

	return defaults
}

// DerivedMinToolVersion returns the oldest golangci-lint version that can
// LOAD a config rendered with these settings: the maximum Since over every
// rendered VersionedSettings block. Zero means "no constraint" (no rendered
// block carries a version floor).
func (d DesiredConfig) DerivedMinToolVersion() types.GolangciVersion {
	var minimum types.GolangciVersion

	for _, blocks := range d.LinterSettings {
		for _, block := range blocks {
			if block.Since.IsZero() {
				continue
			}

			if minimum.IsZero() || block.Since.Compare(minimum) > 0 {
				minimum = block.Since
			}
		}
	}

	return minimum
}
