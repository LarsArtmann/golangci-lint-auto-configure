package constants_test

import (
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Preset parity: the query-derived preset lists must reproduce the hardcoded
// lists they replaced (frozen here as the legacy truth). If a query result
// legitimately changes, change the legacy expectation CONSCIOUSLY in this
// file and in FEATURES.md in the same commit.
var _ = Describe("Preset queries parity", func() {
	var legacyPresetLinters map[string][]types.LinterName

	BeforeEach(func() {
		legacyPresetLinters = map[string][]types.LinterName{
			"minimal": {
				"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
			},
			"standard": {
				"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
				"gocritic", "unused", "copyloopvar",
			},
			"strict": {
				"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
				"gocritic", "unused", "copyloopvar",
				"gocyclo", "funlen", "cyclop", "gocognit",
				"nestif", "maintidx",
				"dupl",
				"goconst", "misspell",
				"nolintlint", "godot", "godox",
			},
			"security":    {"gosec"},
			"performance": {"ineffassign", "prealloc", "unconvert", "perfsprint"},
			"reference": {
				"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
				"loggercheck", "errchkjson", "musttag", "sloglint", "nilerr", "noctx",
				"wrapcheck", "errorlint", "prealloc", "unconvert",
				"gocyclo", "funlen", "cyclop", "gocognit", "maintidx",
				"exhaustive", "goconst", "misspell",
				"revive", "nolintlint", "forcetypeassert", "gocritic", "unused",
				"bodyclose", "contextcheck", "dupl", "durationcheck", "errname",
				"gochecknoglobals", "gochecknoinits", "gosmopolitan", "interfacebloat",
				"nestif", "nilnil", "nakedret", "predeclared", "reassign",
				"rowserrcheck", "spancheck", "sqlclosecheck", "testifylint",
				"thelper", "unparam", "wastedassign", "copyloopvar",
				"ginkgolinter", "gochecksumtype", "intrange", "mirror",
				"perfsprint", "protogetter", "usetesting", "recvcheck", "nilnesserr",
				"zerologlint", "paralleltest",
			},
			"format": {
				"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
			},
			"house": {
				"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
			},
		}
	})

	It("derives every preset exactly as the old hardcoded lists", func() {
		Expect(constants.PresetLinters).To(HaveLen(len(legacyPresetLinters)))

		for preset, legacy := range legacyPresetLinters {
			derived, ok := constants.PresetLinters[preset]
			Expect(ok).To(BeTrue(), "preset %q missing", preset)

			derivedSorted := slices.Clone(derived)
			slices.Sort(derivedSorted)

			legacySorted := slices.Clone(legacy)
			slices.Sort(legacySorted)

			Expect(derivedSorted).To(Equal(legacySorted),
				"preset %q drifted from the legacy hardcoded list", preset)
		}
	})

	It("keeps every preset name valid", func() {
		for _, name := range []string{"minimal", "standard", "strict", "security", "performance", "reference", "format", "house"} {
			_, ok := constants.PresetLinters[name]
			Expect(ok).To(BeTrue(), "preset %q missing from PresetLinters", name)

			Expect(constants.ValidPresets).To(ContainSubstring(name))
		}
	})

	It("derives descriptions whose counts match the resolved sets", func() {
		Expect(constants.PresetDescriptions["minimal"]).
			To(ContainSubstring("(5 linters)"))
		Expect(constants.PresetDescriptions["standard"]).
			To(ContainSubstring("(8 linters)"))
		Expect(constants.PresetDescriptions["strict"]).
			To(ContainSubstring("(20 linters)"))
		Expect(constants.PresetDescriptions["reference"]).
			To(ContainSubstring("(61 linters"))
		Expect(constants.PresetDescriptions["format"]).
			To(ContainSubstring("(5 linters, 3 formatters)"))
		Expect(constants.PresetDescriptions["house"]).
			To(ContainSubstring("(3 formatters"))
	})

	It("resolves format and house formatter sets from the core stack", func() {
		for _, preset := range []string{"format", "house"} {
			formatters, ok := constants.PresetFormatters[preset]
			Expect(ok).To(BeTrue(), "preset %q missing from PresetFormatters", preset)
			Expect(formatters).To(Equal(constants.CoreFormatters))
		}
	})
})
