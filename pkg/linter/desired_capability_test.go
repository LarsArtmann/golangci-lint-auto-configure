package linter

import (
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The capability gate: a linter the detected golangci-lint binary cannot
// name (a module plugin on a stock binary) must never be rendered — an
// enable entry the binary rejects makes the whole config unloadable
// ("unknown linters"). These specs pin the gate, the fail-open behavior,
// and the Unsupported report.
var _ = Describe("Desired renderer capability gate", func() {
	contains := func(names []types.LinterName, want string) bool {
		return slices.Contains(names, types.LinterName(want))
	}

	// stockEnv mirrors a probed stock binary: every bundled linter known,
	// gohumanize (a module plugin) absent.
	stockEnv := func() Env {
		known := make(map[types.LinterName]bool, len(constants.LinterPolicies))
		for _, record := range constants.LinterPolicies {
			if record.Name == "gohumanize" {
				continue
			}

			known[record.Name] = true
		}

		return Env{
			Tool:        types.MustParseGolangciVersion("v2.14.0"),
			ToolBuildGo: types.MustParseGoVersion("1.27"),
			LocalGo:     types.MustParseGoVersion("1.27"),
			Tech:        map[string]bool{"go-humanize": true},
			BinaryKnown: known,
		}
	}

	It("withholds unknown-to-binary linters from the enable set", func() {
		desired := Desired(stockEnv())

		Expect(contains(desired.EnableLinters, "gohumanize")).
			To(BeFalse(), "a stock binary rejects gohumanize; rendering it breaks the config")
		Expect(contains(desired.EnableLinters, "errcheck")).
			To(BeTrue(), "known linters render unaffected")
	})

	It("reports withheld recommendations as unsupported", func() {
		desired := Desired(stockEnv())

		Expect(desired.Unsupported).To(Equal([]types.LinterName{"gohumanize"}))
	})

	It("fails open when the binary could not be probed", func() {
		env := stockEnv()
		env.BinaryKnown = nil

		desired := Desired(env)

		Expect(contains(desired.EnableLinters, "gohumanize")).
			To(BeTrue(), "an unknown binary gates nothing")
		Expect(desired.Unsupported).To(BeEmpty())
	})

	It("keeps gating on tech detection before the capability gate", func() {
		env := stockEnv()
		env.Tech["go-humanize"] = false

		desired := Desired(env)

		Expect(contains(desired.EnableLinters, "gohumanize")).To(BeFalse())
		Expect(desired.Unsupported).To(BeEmpty(),
			"a tech-gated miss is not an unsupported recommendation")
	})

	It("does not report user-disabled or user-enabled linters as unsupported", func() {
		disabled := Desired(stockEnv(), WithUserDisabled([]types.LinterName{"gohumanize"}))
		Expect(disabled.Unsupported).To(BeEmpty())

		enabled := Desired(stockEnv(), WithUserEnabled([]types.LinterName{"gohumanize"}))
		Expect(enabled.Unsupported).To(BeEmpty(),
			"a manual enable of an unknown linter is the heal normalizer's job, not a recommendation")
	})

	It("withholds settings for unknown-to-binary linters", func() {
		env := stockEnv()
		delete(env.BinaryKnown, "cyclop")

		desired := Desired(env)

		Expect(desired.LinterSettings).NotTo(HaveKey(types.LinterName("cyclop")),
			"settings for an unknown linter ride along with an unloadable config")
		Expect(desired.LinterSettings).To(HaveKey(types.LinterName("goconst")),
			"known linters keep their curated settings")
		Expect(desired.Unsupported).To(ContainElement(types.LinterName("cyclop")))
	})
})
