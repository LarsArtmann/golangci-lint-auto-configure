// Package main generates the schema-verification fixture consumed by the CI
// schema-compat gate: a golangci-lint config containing EVERY default the fixer
// injects (linter/formatter settings from DefaultLinterSettings and
// DefaultFormatterSettings, run/output/issues normalizations, and default
// exclusion rules and paths).
//
// Usage:
//
//	go run ./cmd/generate-schema-fixture -output=pkg/constants/testdata/schema-fixture/.golangci.yml
//
// The fixture is serialized through the tool's own SaveConfig path, so the gate
// verifies the exact bytes users receive. CI runs the generator with
// `git diff --exit-code` (drift guard: new settings cannot skip the gate) and
// `golangci-lint config verify` (schema gate: every injected key must be valid).
//
//nolint:all // dev tool — assembles a fixture from constants, no error taxonomy needed
package main

import (
	"charm.land/log/v2"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func main() {
	output := flag.String(
		"output",
		"pkg/constants/testdata/schema-fixture/.golangci.yml",
		"path of the generated fixture config",
	)
	flag.Parse()

	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.WarnLevel})

	cfg := buildFixture()

	if err := os.MkdirAll(filepath.Dir(*output), 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create fixture directory: %v\n", err)

		os.Exit(1)
	}

	if err := config.NewLoader(logger).SaveConfig(cfg, *output); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write fixture: %v\n", err)

		os.Exit(1)
	}

	fmt.Printf("wrote schema fixture to %s (%d linter settings, %d formatter settings)\n",
		*output, len(cfg.Linters.Settings), len(cfg.Formatters.Settings))
}

func buildFixture() *types.Config {
	linterNames := make([]types.LinterName, 0, len(constants.DefaultLinterSettings))
	settings := make(map[string]any, len(constants.DefaultLinterSettings))

	for name, converter := range constants.DefaultLinterSettings {
		linterNames = append(linterNames, name)
		settings[string(name)] = converter.ToMap()
	}

	sort.Slice(linterNames, func(i, j int) bool { return linterNames[i] < linterNames[j] })

	formatterNames := make([]types.FormatterName, 0, len(constants.DefaultFormatterSettings))
	formatterSettings := make(map[string]any, len(constants.DefaultFormatterSettings))

	for name, converter := range constants.DefaultFormatterSettings {
		formatterNames = append(formatterNames, name)
		formatterSettings[string(name)] = converter.ToMap()
	}

	sort.Slice(formatterNames, func(i, j int) bool { return formatterNames[i] < formatterNames[j] })

	buildTags := constants.GoExperimentTags()
	sort.Strings(buildTags)

	return &types.Config{
		Version: "2",
		Run: types.RunConfig{
			Timeout:              constants.DefaultTimeout,
			BuildTags:            buildTags,
			AllowParallelRunners: true,
			AllowSerialRunners:   true,
			IssuesExitCode:       1,
		},
		Output: types.OutputConfig{Formats: map[string]any{}},
		Linters: types.LintersConfig{
			Default:  "none",
			Enable:   linterNames,
			Settings: settings,
			Exclusions: types.LintersExclusionsConfig{
				Generated: "lax",
				Rules:     constants.DefaultExclusionRules,
				Paths:     constants.DefaultLinterExclusionPaths,
			},
		},
		Formatters: types.FormattersConfig{
			Enable:   formatterNames,
			Settings: formatterSettings,
			Exclusions: types.FormattersExclusionsConfig{
				Generated: "lax",
				Paths:     constants.DefaultFormatterExclusionPaths,
			},
		},
		Issues: types.IssuesConfig{
			MaxIssuesPerLinter: constants.DefaultMaxIssuesPerLinter,
			MaxSameIssues:      constants.DefaultMaxSameIssues,
		},
	}
}
