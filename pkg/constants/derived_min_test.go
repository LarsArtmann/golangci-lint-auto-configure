package constants_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The derived tool minimum replaces the hand-maintained constant (plan T13):
// it must equal the value the e2e pin matrix proved (v2.12.0, the goconst
// ignore-tests floor), and adding a newer curated setting must move it.
var _ = Describe("DerivedMinToolVersion", func() {
	It("derives the proven v2.12.0 floor from the table", func() {
		Expect(constants.DerivedMinToolVersion()).
			To(Equal(types.MustParseGolangciVersion("v2.12.0")))
	})

	It("backs the exported constant", func() {
		Expect(constants.MinGolangCILintVersion).
			To(Equal(constants.DerivedMinToolVersion().String()))
	})

	It("ignores settings on never-rendered tiers", func() {
		// exhaustruct_v5 carries v2.13.0 settings but is NeverAutoEnable —
		// its settings never render, so they must not constrain the minimum.
		record, known := constants.PolicyRecord("exhaustruct_v5")
		Expect(known).To(BeTrue())
		Expect(record.Settings[0].Since).To(Equal(types.MustParseGolangciVersion("v2.13.0")))

		Expect(constants.DerivedMinToolVersion()).
			To(Equal(types.MustParseGolangciVersion("v2.12.0")))
	})
})
