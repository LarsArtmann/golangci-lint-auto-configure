package linter

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ProbeKnownLinters returns every linter name the detected golangci-lint
// binary can name: the Enabled ∪ Disabled vocabulary of a config-independent
// `golangci-lint linters` run. Plugin linters (e.g. gohumanize, a golangci-lint
// module plugin) are absent from stock binaries — no version interval can
// express that, only the binary itself can. The probe result is cached for the
// Analyzer's lifetime; an error (missing binary, unparseable output) returns a
// nil set and callers must fail open.
func (a *Analyzer) ProbeKnownLinters(ctx context.Context) (map[types.LinterName]bool, error) {
	if a.knownLinters != nil {
		return a.knownLinters, nil
	}

	if a.golangciLintPath == "" {
		if err := a.FindBinary(ctx); err != nil {
			return nil, err
		}
	}

	known, defaults, err := a.runLintersProbe(ctx)
	if err != nil {
		return nil, err
	}

	a.knownLinters = known
	a.defaultEnabled = defaults

	return known, nil
}

// runLintersProbe shells out once and splits the wire output into the binary's
// vocabulary (enabled ∪ disabled) and its default-enabled set (enabled).
func (a *Analyzer) runLintersProbe(
	ctx context.Context,
) (map[types.LinterName]bool, []types.LinterName, error) {
	output, err := a.runProbeCommand(ctx)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := parseLintersProbe(output)
	if err != nil {
		return nil, nil, err
	}

	vocabulary := probeVocabulary(parsed)
	defaults := probeDefaultEnabled(parsed)

	// Defense-in-depth against an inherited config: a default set covering
	// most of the vocabulary is not "default-enabled" but "everything"
	// (observed 2026-11-01 when an ambient /tmp/.golangci.yml leaked into a
	// non-hermetic probe; the initial "v2.15.0 lists everything" attribution
	// was falsified 2026-11-02 — v2.15.0's bare-dir output matches v2.14's).
	// The only consumer is the degraded-analysis fallback in
	// collectAnalyzerEnabled, where a poisoned set would silently suppress
	// EVERY recommendation; an implausible set fails open to nil instead.
	if len(defaults)*2 > len(vocabulary) {
		defaults = nil
	}

	return vocabulary, defaults, nil
}

// parseLintersProbe decodes the wire format and rejects empty listings (a
// binary that names no linters is broken; treating that as "known nothing"
// would silently strip every linter from the render).
func parseLintersProbe(output []byte) (golangciLintOutput, error) {
	var parsed golangciLintOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return golangciLintOutput{}, apperrors.NewAnalysisError(
			"failed to parse golangci-lint linters JSON output", "", err)
	}

	if len(parsed.Enabled)+len(parsed.Disabled) == 0 {
		return golangciLintOutput{}, apperrors.NewAnalysisError(
			"golangci-lint linters probe returned no linters", "", nil)
	}

	return parsed, nil
}

func probeVocabulary(parsed golangciLintOutput) map[types.LinterName]bool {
	known := make(map[types.LinterName]bool, len(parsed.Enabled)+len(parsed.Disabled))
	for _, entry := range parsed.Enabled {
		known[types.LinterName(entry.Name)] = true
	}

	for _, entry := range parsed.Disabled {
		known[types.LinterName(entry.Name)] = true
	}

	return known
}

func probeDefaultEnabled(parsed golangciLintOutput) []types.LinterName {
	defaults := make([]types.LinterName, 0, len(parsed.Enabled))
	for _, entry := range parsed.Enabled {
		defaults = append(defaults, types.LinterName(entry.Name))
	}

	return defaults
}

// DefaultEnabledLinters returns the linters the detected binary enables with
// no config (the Enabled list of the same probe run). Nil until the probe
// has run successfully — callers treat nil as "unknown".
func (a *Analyzer) DefaultEnabledLinters(_ context.Context) []types.LinterName {
	return a.defaultEnabled
}

// runProbeCommand runs `golangci-lint linters --json` pinned to a minimal
// explicit config in a fresh temp dir. A bare temp cwd is NOT hermetic:
// golangci-lint discovers configs by walking UP from the working
// directory, so an ambient config in any parent (observed 2026-11-01:
// /tmp/.golangci.yml enabling 108 of 115 linters) leaks into the probe and
// poisons both the default-enabled set and, via load failures, the
// vocabulary. The explicit --config pins the listing to the binary's true
// defaults regardless of where the probe runs. The config stays minimal
// (version only) so the probe never depends on the user's config — a
// config naming an unknown linter fails to load, exactly the condition
// the probe must diagnose.
// probeConfigFilePerm keeps the throwaway probe config owner-only; nothing
// else on the machine needs to read it.
const probeConfigFilePerm os.FileMode = 0o600

func (a *Analyzer) runProbeCommand(ctx context.Context) ([]byte, error) {
	empty, err := os.MkdirTemp("", "golangci-lint-auto-configure-probe-")
	if err != nil {
		return nil, apperrors.WrapClassified(err, "probe.tempdir", "create probe working dir")
	}

	defer os.RemoveAll(empty)

	minimal := filepath.Join(empty, "probe.yml")
	if err := os.WriteFile(minimal, []byte("version: \"2\"\n"), probeConfigFilePerm); err != nil {
		return nil, apperrors.WrapClassified(err, "probe.minimal-config", "write minimal probe config")
	}

	cmd := a.executeCommand(ctx, "linters", "--json", "--config", minimal)
	cmd.Dir = empty

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, a.formatCommandError("linters probe", output, err)
	}

	return output, nil
}

// VerifyConfigLoads proves the written config is loadable by the detected
// binary: the one check that catches unknown-linter entries (`golangci-lint
// config verify` validates the schema, which accepts any linter string, but
// config LOAD rejects unknown names). It is the post-apply tripwire: a fix
// that breaks every future lint run must never pass silently.
func (a *Analyzer) VerifyConfigLoads(ctx context.Context, configPath string) error {
	if _, err := a.runLintersCommand(ctx, configPath); err != nil {
		return err
	}

	return nil
}
