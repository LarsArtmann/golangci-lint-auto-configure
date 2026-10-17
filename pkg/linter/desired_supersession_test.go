package linter

import (
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the one-round self-consistency of the desired state: a
// linter superseded by a desired formatter must not be recommended in the
// same render. Planning both produced a config the next detect reported as
// stale — a permanent detect/repair convergence loop (2026-10-17 BuildFlow
// incident, golines supersedes lll).
var _ = Describe("Desired supersession consistency", func() {
	basicEnv := func() Env {
		return Env{
			Tool:        types.MustParseGolangciVersion("v2.14.0"),
			ToolBuildGo: types.MustParseGoVersion("1.27"),
			LocalGo:     types.MustParseGoVersion("1.27"),
			Tech:        map[string]bool{},
		}
	}

	It("never recommends a linter a desired formatter supersedes", func() {
		desired := Desired(basicEnv())

		Expect(desired.EnableFormatters).To(ContainElement(types.FormatterName("golines")))
		Expect(slices.Contains(desired.EnableLinters, types.LinterName("lll"))).
			To(BeFalse(), "lll is superseded by golines and must not be planned for enable")
	})

	It("keeps unsuperseded linters in the enable set", func() {
		desired := Desired(basicEnv())

		Expect(desired.EnableLinters).To(ContainElement(types.LinterName("errcheck")))
	})

	It("keeps the render sorted after supersession filtering", func() {
		desired := Desired(basicEnv())

		Expect(slices.IsSorted(desired.EnableLinters)).To(BeTrue())
	})
})
