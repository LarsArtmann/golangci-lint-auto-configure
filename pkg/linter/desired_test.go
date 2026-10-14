package linter

import (
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the pure desired-state renderer: policy (the table) plus
// environment in, desired state out, no I/O. They are the golden render
// contract the P5 converge engine is built on.
var _ = Describe("Desired renderer", func() {
	basicEnv := func() Env {
		return Env{
			Tool:        types.MustParseGolangciVersion("v2.14.0"),
			ToolBuildGo: types.MustParseGoVersion("1.27"),
			LocalGo:     types.MustParseGoVersion("1.27"),
			Tech:        map[string]bool{},
		}
	}

	contains := func(names []types.LinterName, want string) bool {
		return slices.Contains(names, types.LinterName(want))
	}

	It("is deterministic: the same env renders identically", func() {
		first := Desired(basicEnv())
		second := Desired(basicEnv())

		Expect(first).To(Equal(second))
	})

	It("renders the core recommendable linters", func() {
		desired := Desired(basicEnv())

		for _, linter := range []string{"errcheck", "staticcheck", "govet", "gosec", "wrapcheck", "goconst"} {
			Expect(contains(desired.EnableLinters, linter)).
				To(BeTrue(), "%s must be in the default enable set", linter)
		}
	})

	It("never renders disabled or never-auto-enable linters", func() {
		desired := Desired(basicEnv())

		for _, linter := range []string{"funcorder", "noinlineerr", "depguard", "exhaustruct_v5"} {
			Expect(contains(desired.EnableLinters, linter)).
				To(BeFalse(), "%s must never be auto-rendered", linter)
		}
	})

	It("renders pragmatic-noise linters by default and drops them with --pragmatic", func() {
		defaultRender := Desired(basicEnv())
		for _, linter := range constants.LintersInTier(types.TierPragmaticNoise) {
			Expect(contains(defaultRender.EnableLinters, string(linter))).
				To(BeTrue(), "pragmatic-noise linter %s must render by default", linter)
		}

		pragmatic := Desired(basicEnv(), WithPragmatic(true))
		for _, linter := range constants.LintersInTier(types.TierPragmaticNoise) {
			Expect(contains(pragmatic.EnableLinters, string(linter))).
				To(BeFalse(), "pragmatic-noise linter %s must drop under --pragmatic", linter)
		}
	})

	It("respects the priority threshold", func() {
		highOnly := Desired(basicEnv(), WithMinPriority(types.LinterPriorityHigh))

		Expect(contains(highOnly.EnableLinters, "errcheck")).To(BeTrue()) // Critical
		Expect(contains(highOnly.EnableLinters, "wrapcheck")).To(BeTrue()) // High
		Expect(contains(highOnly.EnableLinters, "godot")).To(BeFalse())    // Medium
	})

	It("gates tech-gated linters on detection, failing open on unknown tech", func() {
		excluded := basicEnv()
		excluded.Tech["clickhouse"] = false
		Expect(contains(Desired(excluded).EnableLinters, "clickhouselint")).To(BeFalse())

		included := basicEnv()
		included.Tech["clickhouse"] = true
		Expect(contains(Desired(included).EnableLinters, "clickhouselint")).To(BeTrue())

		unknown := basicEnv()
		delete(unknown.Tech, "go-humanize")
		Expect(contains(Desired(unknown).EnableLinters, "gohumanize")).To(BeTrue())
	})

	It("validity intervals gate on the tool version", func() {
		old := basicEnv()
		old.Tool = types.MustParseGolangciVersion("v2.12.0")
		Expect(contains(Desired(old).EnableLinters, "exhaustruct_v5")).To(BeFalse())

		// exhaustruct_v5 is never-auto-enable, so use its ValidTool bound
		// through a record that IS renderable: gomodguard_v2 does not exist
		// below v2.12.0.
		below := basicEnv()
		below.Tool = types.MustParseGolangciVersion("v2.11.0")
		Expect(contains(Desired(below).EnableLinters, "gomodguard_v2")).To(BeFalse())

		at := basicEnv()
		at.Tool = types.MustParseGolangciVersion("v2.12.0")
		Expect(contains(Desired(at).EnableLinters, "gomodguard_v2")).To(BeTrue())
	})

	It("language intervals gate on the project Go version", func() {
		old := basicEnv()
		old.ProjectGo = types.MustParseGoVersion("1.21")
		Expect(contains(Desired(old).EnableLinters, "copyloopvar")).To(BeFalse())

		reached := basicEnv()
		reached.ProjectGo = types.MustParseGoVersion("1.22")
		Expect(contains(Desired(reached).EnableLinters, "copyloopvar")).To(BeTrue())
	})

	It("renders the core formatter stack in canonical order", func() {
		desired := Desired(basicEnv())

		Expect(desired.EnableFormatters).To(Equal([]types.FormatterName{"goimports", "gofumpt", "golines"}))
	})

	It("gates swaggo on tech detection", func() {
		without := basicEnv()
		without.Tech["swaggo"] = false

		// swaggo is not core, so it never renders in the default set; the
		// gate matters for the record's recommendability, covered by the
		// formatter table specs. Here we pin that core rendering is stable.
		Expect(Desired(without).EnableFormatters).To(Equal(constants.CoreFormatters))
	})

	It("carries the curated settings of rendered linters", func() {
		desired := Desired(basicEnv())

		Expect(desired.LinterSettings).To(HaveKey(types.LinterName("goconst")))
		Expect(desired.LinterSettings).To(HaveKey(types.LinterName("errcheck")))
		Expect(desired.LinterSettings).NotTo(HaveKey(types.LinterName("govet")))

		blocks := desired.LinterSettings[types.LinterName("goconst")]
		Expect(blocks).To(HaveLen(1))
		Expect(blocks[0].Since).To(Equal(types.MustParseGolangciVersion("v2.12.0")))
		rendered := blocks[0].Settings.ToMap()
		Expect(rendered).To(HaveKey("ignore-tests"))
	})

	It("hides settings blocks newer than the tool", func() {
		old := basicEnv()
		old.Tool = types.MustParseGolangciVersion("v2.11.0")

		desired := Desired(old)
		Expect(desired.LinterSettings).NotTo(HaveKey(types.LinterName("goconst")))
	})

	It("caps run.go at the binary's build Go", func() {
		newer := basicEnv()
		newer.LocalGo = types.MustParseGoVersion("1.28")

		desired := Desired(newer)
		Expect(desired.RunDefaults.Go).To(Equal("1.27"))
		Expect(desired.RunDefaults.IssuesExitCode).To(Equal(1))
		Expect(desired.RunDefaults.AllowParallelRunner).To(BeTrue())
	})

	It("leaves run.go empty when the local Go is unknown", func() {
		unknown := basicEnv()
		unknown.LocalGo = types.GoVersion{}

		desired := Desired(unknown)
		Expect(desired.RunDefaults.Go).To(Equal(""))
	})

	It("includes static and detected exclusion paths", func() {
		env := basicEnv()
		env.Generated = []string{`_gen\.go$`, `zz_generated\.go$`}

		desired := Desired(env)
		Expect(desired.LinterExclusionPaths).To(ContainElements("_templ\\.go$", "vendor/"))
		Expect(desired.GeneratedExclusionPaths).To(Equal([]string{"_gen\\.go$", "zz_generated\\.go$"}))
		Expect(desired.ExclusionRules).To(HaveLen(len(constants.DefaultExclusionRules)))
	})

	It("derives the tool minimum from rendered settings", func() {
		desired := Desired(basicEnv())

		// goconst's ignore-tests key is the binding constraint; the derived
		// value must equal the hand-maintained constant it replaces.
		Expect(desired.DerivedMinToolVersion()).
			To(Equal(types.MustParseGolangciVersion(constants.MinGolangCILintVersion)))
	})

	It("derives no minimum when no settings render", func() {
		old := basicEnv()
		old.Tool = types.MustParseGolangciVersion("v2.11.0")

		desired := Desired(old, WithMinPriority(types.LinterPriorityHigh))
		Expect(desired.DerivedMinToolVersion().IsZero()).To(BeTrue())
	})
})
