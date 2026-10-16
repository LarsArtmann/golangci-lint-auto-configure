package linter

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/gogenfilter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Env is the complete set of environment facts the desired-state renderer
// needs. Every fact is collected once, at one boundary, with typed shapes —
// the renderer downstream is pure and takes no I/O.
//
// Zero values mean "unknown": a zero Tool disables version gating, a zero
// LocalGo skips run.go updates, an empty Tech set fails gates open (matching
// the analyzer's fail-open behavior).
type Env struct {
	// Tool is the detected golangci-lint version; zero when undetectable.
	Tool types.GolangciVersion
	// ToolBuildGo is the Go version the golangci-lint binary was built with
	// (run.go cap); zero when unknown.
	ToolBuildGo types.GoVersion
	// LocalGo is the local toolchain version (major.minor); zero when
	// unparsable.
	LocalGo types.GoVersion
	// ProjectGo is the project's go.mod go-directive; zero when absent.
	ProjectGo types.GoVersion
	// Tech maps technology keys (from policy Gates) to detection results.
	// Missing keys fail open in ShouldRecommend.
	Tech map[string]bool
	// Generated holds the auto-detected generated-file exclusion patterns
	// (gogenfilter results).
	Generated []string
	// AnalyzerEnabled lists the linters golangci-lint reports as enabled for
	// this config (config enables plus the binary's default set). The
	// renderer skips these: re-enabling an already-enabled linter would
	// churn every config.
	AnalyzerEnabled map[types.LinterName]bool
	// BinaryKnown lists every linter name the detected golangci-lint binary
	// can name (Enabled ∪ Disabled of `golangci-lint linters`). Plugin
	// linters are absent from stock binaries, and an enable entry naming one
	// makes the whole config unloadable. Empty means "binary unknown" —
	// capability gating fails open.
	BinaryKnown map[types.LinterName]bool
}

// UsesTech reports the project's technology detection result, failing open
// for unknown keys (an undetectable technology must not suppress
// recommendations).
func (e Env) UsesTech(tech string) bool {
	result, known := e.Tech[tech]
	if !known {
		return true
	}

	return result
}

// techKeys returns every technology key the policy tables gate on.
func techKeys() []string {
	keys := make(map[string]bool)

	for _, record := range constants.LinterPolicies {
		if record.Gate.Kind == types.GateRequiresTech {
			keys[record.Gate.Tech] = true
		}
	}

	for _, record := range constants.FormatterPolicies {
		if record.Gate.Kind == types.GateRequiresTech {
			keys[record.Gate.Tech] = true
		}
	}

	unique := make([]string, 0, len(keys))
	for key := range keys {
		unique = append(unique, key)
	}

	sort.Strings(unique)

	return unique
}

// CollectEnv gathers every environment fact the renderer needs. Best-effort
// by design: detection failures degrade to zero values / fail-open tech
// results (with the same semantics the analyzer applies today), because a
// partially-known environment must produce a conservative config, not an
// error. The returned error is reserved for a fundamentally unusable
// environment (nil today — every failure mode has a defined degradation).
func CollectEnv(
	ctx context.Context,
	analyzer types.LinterAnalyzer,
	goVersionProvider GoVersionProvider,
	configPath string,
) (Env, error) {
	env, err := CollectEnvBase(ctx, analyzer, goVersionProvider, configPath)
	if err != nil {
		return env, err
	}

	env.AnalyzerEnabled = collectAnalyzerEnabled(ctx, analyzer, configPath)

	return env, nil
}

// CollectEnvBase gathers every environment fact EXCEPT the analyzer's
// linter-list run. The converge driver uses it to normalize a config BEFORE
// analysis (the analyzer cannot list linters for a config that still names
// removed ones) and attaches the analysis afterwards.
func CollectEnvBase(
	ctx context.Context,
	analyzer types.LinterAnalyzer,
	goVersionProvider GoVersionProvider,
	configPath string,
) (Env, error) {
	env := Env{Tech: map[string]bool{}}

	if analyzer != nil {
		if err := analyzer.FindBinary(ctx); err == nil {
			_ = analyzer.CheckVersion(ctx) // classify: failure leaves Tool zero
		}

		env.Tool = detectedToolVersion(analyzer)
		env.ToolBuildGo = analyzer.DetectedBuildGoVersion()
	}

	env.LocalGo = collectLocalGo(goVersionProvider, ctx)
	env.ProjectGo = readGoModGoDirective(configPath)
	env.Tech = collectTech(configPath)
	env.Generated = collectGeneratedExclusions(configPath)
	env.BinaryKnown = collectBinaryKnown(ctx, analyzer)

	return env, nil
}

// linterProber is the capability surface of the concrete *Analyzer: which
// linters the detected binary can name, and which it enables by default.
// Duck-typed (like the audit ledger reader) so test doubles that lack the
// probe fail open instead of implementing the full analyzer surface.
type linterProber interface {
	ProbeKnownLinters(ctx context.Context) (map[types.LinterName]bool, error)
	DefaultEnabledLinters(ctx context.Context) []types.LinterName
}

// collectAnalyzerEnabled records which linters the binary reports as enabled
// (config enables plus its default set). Best-effort: on failure the set
// falls back to the probe's default-enabled linters (a config naming an
// unknown linter cannot be analyzed, but the binary's defaults still prevent
// bogus re-declaration churn), and on probe failure stays empty.
func collectAnalyzerEnabled(
	ctx context.Context,
	analyzer types.LinterAnalyzer,
	configPath string,
) map[types.LinterName]bool {
	if analyzer == nil {
		return nil
	}

	if enabled := analysisEnabledSet(analyzer, ctx, configPath); enabled != nil {
		return enabled
	}

	return probeDefaultEnabledSet(analyzer, ctx)
}

// analysisEnabledSet records the linters a successful analysis reports as
// enabled (config enables plus the binary's defaults); nil when the config
// could not be analyzed (e.g. it names a linter the binary does not carry).
func analysisEnabledSet(
	analyzer types.LinterAnalyzer, ctx context.Context, configPath string,
) map[types.LinterName]bool {
	analysis, err := analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil || analysis == nil {
		return nil
	}

	enabled := make(map[types.LinterName]bool, len(analysis.EnabledLinters))
	for _, info := range analysis.EnabledLinters {
		enabled[info.Name] = true
	}

	return enabled
}

// probeDefaultEnabledSet falls back to the binary's default-enabled linters
// (the same probe run's Enabled list): a config the binary cannot load is
// about to be repaired, and the defaults are the right churn guard for that
// render. Nil when the analyzer lacks the probe or the probe failed.
func probeDefaultEnabledSet(analyzer types.LinterAnalyzer, ctx context.Context) map[types.LinterName]bool {
	prober, ok := analyzer.(linterProber)
	if !ok {
		return nil
	}

	defaults := prober.DefaultEnabledLinters(ctx)
	if len(defaults) == 0 {
		return nil
	}

	enabled := make(map[types.LinterName]bool, len(defaults))
	for _, name := range defaults {
		enabled[name] = true
	}

	return enabled
}

// collectBinaryKnown runs the capability probe: every linter name the
// detected binary can name. Best-effort: analyzers without the probe and
// failed probes yield nil, and the renderer fails open.
func collectBinaryKnown(ctx context.Context, analyzer types.LinterAnalyzer) map[types.LinterName]bool {
	if analyzer == nil {
		return nil
	}

	prober, ok := analyzer.(linterProber)
	if !ok {
		return nil
	}

	known, err := prober.ProbeKnownLinters(ctx)
	if err != nil {
		return nil
	}

	return known
}

func detectedToolVersion(analyzer types.LinterAnalyzer) types.GolangciVersion {
	raw := analyzer.GetDetectedVersion()
	if raw == "" {
		return types.GolangciVersion{}
	}

	parsed, err := types.ParseGolangciVersion(raw)
	if err != nil {
		return types.GolangciVersion{}
	}

	return parsed
}

func collectLocalGo(provider GoVersionProvider, ctx context.Context) types.GoVersion {
	if provider == nil {
		return types.GoVersion{}
	}

	raw := provider(ctx)
	if raw == "" {
		return types.GoVersion{}
	}

	parsed, err := types.ParseGoVersion(raw)
	if err != nil {
		return types.GoVersion{}
	}

	return parsed
}

// collectTech runs every policy-gated technology detection, failing open per
// key (detection errors count as "uses it", matching the analyzer's
// hasTechnology contract so valid recommendations are never suppressed).
func collectTech(configPath string) map[string]bool {
	projectRoot := filepathDir(configPath)

	results := make(map[string]bool, len(techKeys()))

	detector := detection.NewDetector(projectRoot)

	for _, tech := range techKeys() {
		results[tech] = detectTechnology(detector, tech)
	}

	return results
}

func detectTechnology(detector *detection.Detector, tech string) bool {
	switch tech {
	case "swaggo":
		result, err := detector.HasSwaggo()
		if err != nil {
			return true
		}

		return result
	case "clickhouse":
		return detector.HasClickHouse()
	case "arangodb":
		return detector.HasArangoDB()
	case "go-humanize":
		return detector.HasGoHumanize()
	default:
		return true
	}
}

// collectGeneratedExclusions scans the project for generated files. Best
// effort: a failed scan yields no extra patterns (the static defaults still
// apply in the renderer).
func collectGeneratedExclusions(configPath string) []string {
	result, err := gogenfilter.ScanProject(os.DirFS(filepathDir(configPath)))
	if err != nil {
		return nil
	}

	patterns := make([]string, 0, len(result.Exclusions))
	for _, exclusion := range result.Exclusions {
		patterns = append(patterns, exclusion.Path)
	}

	sort.Strings(patterns)

	return patterns
}

// filepathDir returns the directory portion of path ("." when there is none).
func filepathDir(path string) string {
	if idx := strings.LastIndex(path, "/"); idx > 0 {
		return path[:idx]
	}

	return "."
}
