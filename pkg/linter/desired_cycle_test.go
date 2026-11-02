package linter

import (
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the declarative regression-loop guard: the renderer
// declines linters the audit ledger observed being auto-enabled and then
// removed by the user, and reports each decline so the fixer can offer the
// durable never-enable sidecar entry. A manual re-enable wins.
var _ = Describe("Desired cycle suppression", func() {
	basicEnv := func() Env {
		return Env{
			Tool:        types.MustParseGolangciVersion("v2.14.0"),
			ToolBuildGo: types.MustParseGoVersion("1.27"),
			LocalGo:     types.MustParseGoVersion("1.27"),
			Tech:        map[string]bool{},
		}
	}

	It("declines a previously-auto-enabled linter and reports the decline", func() {
		desired := Desired(basicEnv(),
			WithCycleSuppressed([]types.LinterName{"gosec"}),
		)

		Expect(slices.Contains(desired.EnableLinters, types.LinterName("gosec"))).
			To(BeFalse(), "cycle-suppressed linter must not be recommended")
		Expect(desired.CycleSuppressed).To(ContainElement(types.LinterName("gosec")))
	})

	It("lets a manual re-enable win over the suppression", func() {
		desired := Desired(basicEnv(),
			WithCycleSuppressed([]types.LinterName{"gosec"}),
			WithUserEnabled([]types.LinterName{"gosec"}),
		)

		Expect(desired.EnableLinters).To(ContainElement(types.LinterName("gosec")))
		Expect(desired.CycleSuppressed).NotTo(ContainElement(types.LinterName("gosec")))
	})

	It("does not report declines for linters that were not recommendable anyway", func() {
		env := basicEnv()
		// Explicitly detected-absent: a missing tech key fails open, so the
		// empty map would render arangolint. False is the real "no driver".
		env.Tech = map[string]bool{"arangodb": false}
		desired := Desired(env,
			WithCycleSuppressed([]types.LinterName{"arangolint"}),
		)

		// arangolint is tech-gated and the project carries no ArangoDB
		// driver: shouldRender already withheld it, so the decline list
		// stays empty.
		Expect(desired.CycleSuppressed).To(BeEmpty())
	})

	It("keeps both output lists sorted", func() {
		desired := Desired(basicEnv(),
			WithCycleSuppressed([]types.LinterName{"gosec", "mnd", "errcheck"}),
		)

		Expect(slices.IsSorted(desired.EnableLinters)).To(BeTrue())
		Expect(slices.IsSorted(desired.CycleSuppressed)).To(BeTrue())
	})
})
