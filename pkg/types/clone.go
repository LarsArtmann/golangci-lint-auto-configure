package types

func cloneSlice[T any](src []T) []T {
	if src == nil {
		return nil
	}

	cp := make([]T, len(src))
	copy(cp, src)

	return cp
}

func cloneAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	cp := make(map[string]any, len(m))
	for k, v := range m {
		cp[k] = deepCloneAny(v)
	}

	return cp
}

func deepCloneAny(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return cloneAnyMap(val)
	case []any:
		return deepCloneSlice(val)
	default:
		return v
	}
}

func deepCloneSlice(s []any) []any {
	if s == nil {
		return nil
	}

	cp := make([]any, len(s))
	for i, v := range s {
		cp[i] = deepCloneAny(v)
	}

	return cp
}

func cloneExclusionRules(src []ExclusionRuleConfig) []ExclusionRuleConfig {
	if src == nil {
		return nil
	}

	cp := make([]ExclusionRuleConfig, len(src))
	for i, rule := range src {
		cp[i] = ExclusionRuleConfig{
			Path:       rule.Path,
			PathExcept: rule.PathExcept,
			Text:       rule.Text,
			Source:     rule.Source,
			Linters:    cloneSlice(rule.Linters),
		}
	}

	return cp
}

// Clone returns a deep copy of the Config.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	return &Config{
		Version:           c.Version,
		Run:               c.Run.Clone(),
		Output:            c.Output.Clone(),
		Linters:           c.Linters.Clone(),
		Formatters:        c.Formatters.Clone(),
		Issues:            c.Issues.Clone(),
		LintersSettingsV1: cloneAnyMap(c.LintersSettingsV1),
	}
}

// Clone returns a deep copy of RunConfig.
func (r *RunConfig) Clone() RunConfig {
	return RunConfig{
		Timeout:              r.Timeout,
		Go:                   r.Go,
		BuildTags:            cloneSlice(r.BuildTags),
		ModulesDownloadMode:  r.ModulesDownloadMode,
		AllowParallelRunners: r.AllowParallelRunners,
		AllowSerialRunners:   r.AllowSerialRunners,
		IssuesExitCode:       r.IssuesExitCode,
		Tests:                r.Tests,
		Concurrency:          r.Concurrency,
		RelativePathMode:     r.RelativePathMode,
	}
}

// Clone returns a deep copy of OutputConfig.
func (o *OutputConfig) Clone() OutputConfig {
	return OutputConfig{
		Formats:    cloneAnyMap(o.Formats),
		PathPrefix: o.PathPrefix,
		PathMode:   o.PathMode,
		SortOrder:  cloneSlice(o.SortOrder),
		ShowStats:  o.ShowStats,
	}
}

// Clone returns a deep copy of LintersConfig.
func (l *LintersConfig) Clone() LintersConfig {
	return LintersConfig{
		Enable:     cloneSlice(l.Enable),
		Disable:    cloneSlice(l.Disable),
		Default:    l.Default,
		Settings:   cloneAnyMap(l.Settings),
		Exclusions: l.Exclusions.Clone(),
	}
}

// Clone returns a deep copy of LintersExclusionsConfig.
func (e *LintersExclusionsConfig) Clone() LintersExclusionsConfig {
	return LintersExclusionsConfig{
		Generated:   e.Generated,
		WarnUnused:  e.WarnUnused,
		Presets:     cloneSlice(e.Presets),
		Rules:       cloneExclusionRules(e.Rules),
		Paths:       cloneSlice(e.Paths),
		PathsExcept: cloneSlice(e.PathsExcept),
	}
}

// Clone returns a deep copy of IssuesConfig.
func (i *IssuesConfig) Clone() IssuesConfig {
	return IssuesConfig{
		MaxIssuesPerLinter: i.MaxIssuesPerLinter,
		MaxSameIssues:      i.MaxSameIssues,
		NewFromRev:         i.NewFromRev,
		NewFromPatch:       i.NewFromPatch,
		New:                i.New,
		NewFromMergeBase:   i.NewFromMergeBase,
		WholeFiles:         i.WholeFiles,
		Fix:                i.Fix,
		UniqByLine:         i.UniqByLine,
	}
}

// Clone returns a deep copy of FormattersConfig.
func (f *FormattersConfig) Clone() FormattersConfig {
	return FormattersConfig{
		Enable:     cloneSlice(f.Enable),
		Disable:    cloneSlice(f.Disable),
		Settings:   cloneAnyMap(f.Settings),
		Exclusions: f.Exclusions.Clone(),
	}
}

// Clone returns a deep copy of FormattersExclusionsConfig.
func (f *FormattersExclusionsConfig) Clone() FormattersExclusionsConfig {
	return FormattersExclusionsConfig{
		Generated:  f.Generated,
		WarnUnused: f.WarnUnused,
		Paths:      cloneSlice(f.Paths),
	}
}
