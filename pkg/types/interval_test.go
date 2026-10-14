package types_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interval", func() {
	tool := types.MustParseGolangciVersion

	Describe("unbounded zero-value semantics", func() {
		It("matches everything when both bounds are zero", func() {
			var anyVersion types.Interval[types.GolangciVersion]
			Expect(anyVersion.IsUnbounded()).To(BeTrue())
			Expect(anyVersion.Contains(tool("v0.1.0"))).To(BeTrue())
			Expect(anyVersion.Contains(tool("v99.0.0"))).To(BeTrue())
			Expect(anyVersion.String()).To(Equal("any"))
		})

		It("treats a zero Min as no lower bound", func() {
			interval := types.Interval[types.GolangciVersion]{Max: tool("v2.12.0")}
			Expect(interval.Contains(tool("v1.0.0"))).To(BeTrue())
			Expect(interval.Contains(tool("v2.12.0"))).To(BeTrue())
			Expect(interval.Contains(tool("v2.12.1"))).To(BeFalse())
			Expect(interval.String()).To(Equal("<=v2.12.0"))
		})

		It("treats a zero Max as no upper bound", func() {
			interval := types.Interval[types.GolangciVersion]{Min: tool("v2.12.0")}
			Expect(interval.Contains(tool("v2.11.9"))).To(BeFalse())
			Expect(interval.Contains(tool("v2.12.0"))).To(BeTrue())
			Expect(interval.Contains(tool("v99.0.0"))).To(BeTrue())
			Expect(interval.String()).To(Equal(">=v2.12.0"))
		})
	})

	Describe("closed bounds", func() {
		It("contains the bounds themselves", func() {
			interval := types.NewInterval(tool("v2.12.0"), tool("v2.13.2"))
			Expect(interval.Contains(tool("v2.11.9"))).To(BeFalse())
			Expect(interval.Contains(tool("v2.12.0"))).To(BeTrue())
			Expect(interval.Contains(tool("v2.13.2"))).To(BeTrue())
			Expect(interval.Contains(tool("v2.13.3"))).To(BeFalse())
			Expect(interval.String()).To(Equal(">=v2.12.0,<=v2.13.2"))
		})
	})

	Describe("inversion", func() {
		It("NewInterval panics on min > max", func() {
			Expect(func() {
				types.NewInterval(tool("v2.14.0"), tool("v2.12.0"))
			}).To(Panic())
		})

		It("IsEmpty detects a struct-literal inversion", func() {
			inverted := types.Interval[types.GolangciVersion]{Min: tool("v2.14.0"), Max: tool("v2.12.0")}
			Expect(inverted.IsEmpty()).To(BeTrue())
			Expect(inverted.Contains(tool("v2.13.0"))).To(BeFalse())
		})
	})

	Describe("cross-axis misuse", func() {
		It("is a compile error to mix the two version axes in one interval", func() {
			// Interval is generic over ONE ordered type. The following does
			// not compile:
			//
			//   _ := types.NewInterval(
			//       types.MustParseGolangciVersion("v2.12.0"),
			//       types.MustParseGoVersion("1.27"), // compile error
			//   )
			//
			// Pinned by review; Go's type system enforces it structurally.
			Succeed()
		})
	})

	Describe("GoVersion intervals", func() {
		It("bounds language relevance", func() {
			copyloopvar := types.Interval[types.GoVersion]{Min: types.MustParseGoVersion("1.22")}
			Expect(copyloopvar.Contains(types.MustParseGoVersion("1.21"))).To(BeFalse())
			Expect(copyloopvar.Contains(types.MustParseGoVersion("1.22"))).To(BeTrue())
			Expect(copyloopvar.Contains(types.MustParseGoVersion("1.27"))).To(BeTrue())
		})
	})

	Describe("ParseInterval round-trip", func() {
		DescribeTable("ParseInterval(String()) equals the original",
			func(original types.Interval[types.GolangciVersion]) {
				parsed, err := types.ParseInterval(original.String(), types.ParseGolangciVersion)
				Expect(err).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(original))
			},
			Entry("unbounded", types.Interval[types.GolangciVersion]{}),
			Entry("max only", types.Interval[types.GolangciVersion]{Max: tool("v2.12.0")}),
			Entry("min only", types.Interval[types.GolangciVersion]{Min: tool("v2.12.0")}),
			Entry("both bounds", types.NewInterval(tool("v2.12.0"), tool("v2.14.0"))),
		)

		DescribeTable("rejecting invalid forms",
			func(input string) {
				_, err := types.ParseInterval[types.GolangciVersion](input, types.ParseGolangciVersion)
				Expect(err).To(HaveOccurred(), "input %q should not parse", input)
			},
			Entry("empty", ""),
			Entry("garbage bound", ">=not-a-version"),
			Entry("bad operator", "=v2.12.0"),
			Entry("trailing comma", ">=v2.12.0,"),
		)
	})

	Describe("MustValidate", func() {
		It("panics on inverted struct literals", func() {
			inverted := types.Interval[types.GolangciVersion]{Min: tool("v2.14.0"), Max: tool("v2.12.0")}
			Expect(inverted.MustValidate).To(Panic())
		})

		It("accepts valid and unbounded intervals", func() {
			Expect(func() {
				types.NewInterval(tool("v2.12.0"), tool("v2.14.0")).MustValidate()

				var unbounded types.Interval[types.GolangciVersion]
				unbounded.MustValidate()
			}).NotTo(Panic())
		})
	})
})
