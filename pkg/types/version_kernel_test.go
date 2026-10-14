package types_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GolangciVersion", func() {
	Describe("ParseGolangciVersion", func() {
		DescribeTable("accepting valid inputs and canonicalizing",
			func(input, want string) {
				got, err := types.ParseGolangciVersion(input)
				Expect(err).NotTo(HaveOccurred())
				Expect(got.String()).To(Equal(want))
				Expect(got.IsZero()).To(BeFalse())
			},
			Entry("canonical v-prefixed", "v2.14.0", "v2.14.0"),
			Entry("bare version gets v prefix", "2.14.0", "v2.14.0"),
			Entry("prerelease", "v2.14.0-rc1", "v2.14.0-rc1"),
			Entry("whitespace trimmed", "  v2.12.0  ", "v2.12.0"),
		)

		DescribeTable("rejecting invalid input",
			func(input string) {
				_, err := types.ParseGolangciVersion(input)
				Expect(err).To(HaveOccurred(), "input %q should not parse", input)
			},
			Entry("empty", ""),
			Entry("whitespace only", "   "),
			Entry("garbage", "abc"),
			Entry("partial semver", "2.x"),
			Entry("bare major", "2"),
		)
	})

	Describe("zero value", func() {
		It("is the zero value of the struct", func() {
			var zero types.GolangciVersion
			Expect(zero.IsZero()).To(BeTrue())
			Expect(zero.String()).To(Equal(""))
		})

		It("is unreachable from Parse", func() {
			_, err := types.ParseGolangciVersion("")
			Expect(err).To(HaveOccurred())
		})

		It("sorts before every parsed version", func() {
			var zero types.GolangciVersion

			v := types.MustParseGolangciVersion("v2.0.0")
			Expect(zero.Compare(v)).To(Equal(-1))
			Expect(v.Compare(zero)).To(Equal(1))

			var otherZero types.GolangciVersion
			Expect(zero.Compare(otherZero)).To(Equal(0))
		})
	})

	Describe("Compare", func() {
		type versionCase struct {
			a, b string
			want int
		}

		DescribeTable("ordering semantically",
			func(c versionCase) {
				a := types.MustParseGolangciVersion(c.a)
				b := types.MustParseGolangciVersion(c.b)
				Expect(a.Compare(b)).To(Equal(c.want))
			},
			Entry("equal", versionCase{"v2.14.0", "v2.14.0", 0}),
			Entry("patch difference", versionCase{"v2.14.0", "v2.14.1", -1}),
			Entry("minor difference", versionCase{"v2.13.9", "v2.14.0", -1}),
			Entry("major difference", versionCase{"v1.99.0", "v2.0.0", -1}),
			Entry("numeric not lexicographic", versionCase{"v2.9.0", "v2.10.0", -1}),
			Entry("symmetry", versionCase{a: "v2.14.1", b: "v2.14.0", want: 1}),
		)
	})

	Describe("AtLeast", func() {
		It("is true when equal or newer", func() {
			v := types.MustParseGolangciVersion("v2.12.0")
			Expect(v.AtLeast(types.MustParseGolangciVersion("v2.12.0"))).To(BeTrue())
			Expect(v.AtLeast(types.MustParseGolangciVersion("v2.11.9"))).To(BeTrue())
			Expect(v.AtLeast(types.MustParseGolangciVersion("v2.12.1"))).To(BeFalse())
		})

		It("is false for a zero receiver (unknown satisfies no minimum)", func() {
			var zero types.GolangciVersion
			Expect(zero.AtLeast(types.MustParseGolangciVersion("v0.0.1"))).To(BeFalse())
		})
	})

	Describe("cross-axis misuse", func() {
		It("is a compile error to compare the two version axes", func() {
			// The following does not compile (different types, no implicit
			// conversion):
			//
			//   tool := types.MustParseGolangciVersion("v2.14.0")
			//   lang := types.MustParseGoVersion("1.27")
			//   tool.Compare(lang) // compile error
			//
			// Pinned by review; Go's type system enforces it structurally.
			Succeed()
		})
	})
})

var _ = Describe("GoVersion", func() {
	Describe("ParseGoVersion", func() {
		DescribeTable("accepting valid inputs at major.minor granularity",
			func(input, want string) {
				got, err := types.ParseGoVersion(input)
				Expect(err).NotTo(HaveOccurred())
				Expect(got.String()).To(Equal(want))
			},
			Entry("go-prefixed patch", "go1.27.2", "1.27"),
			Entry("bare patch", "1.27.1", "1.27"),
			Entry("major.minor", "1.27", "1.27"),
			Entry("go-prefixed major.minor", "go1.26", "1.26"),
		)

		DescribeTable("rejecting invalid input",
			func(input string) {
				_, err := types.ParseGoVersion(input)
				Expect(err).To(HaveOccurred(), "input %q should not parse", input)
			},
			Entry("empty", ""),
			Entry("devel", "devel"),
			Entry("pseudo version", "go1.28-0f9a9bc"),
			Entry("non-numeric minor", "1.x"),
			Entry("major only", "1"),
			Entry("major zero (zero value reserved)", "0.42"),
		)
	})

	Describe("zero value", func() {
		It("sorts before every parsed version", func() {
			var zero types.GoVersion

			v := types.MustParseGoVersion("1.20")
			Expect(zero.Compare(v)).To(Equal(-1))
			Expect(v.Compare(zero)).To(Equal(1))
		})

		It("is never AtLeast anything", func() {
			var zero types.GoVersion
			Expect(zero.AtLeast(types.MustParseGoVersion("1.1"))).To(BeFalse())
		})
	})

	Describe("Compare", func() {
		DescribeTable("ordering numerically",
			func(a, b string, want int) {
				Expect(types.MustParseGoVersion(a).Compare(types.MustParseGoVersion(b))).To(Equal(want))
			},
			Entry("equal", "1.27", "1.27", 0),
			Entry("equal ignoring patch", "1.27.0", "1.27.9", 0),
			Entry("numeric minor", "1.9", "1.10", -1),
			Entry("major dominates", "2.0", "1.99", 1),
			Entry("older minor", "1.26", "1.27", -1),
		)
	})

	Describe("AtLeast", func() {
		It("implements the run.go cap comparison", func() {
			binary := types.MustParseGoVersion("1.27")
			Expect(types.MustParseGoVersion("1.28").AtLeast(binary)).To(BeTrue())
			Expect(types.MustParseGoVersion("1.27").AtLeast(binary)).To(BeTrue())
			Expect(types.MustParseGoVersion("1.26").AtLeast(binary)).To(BeFalse())
		})
	})
})

var _ = Describe("deprecated string helpers", func() {
	It("NormalizeGoMajorMinor forwards to ParseGoVersion", func() {
		got, ok := types.NormalizeGoMajorMinor("go1.27.2")
		Expect(ok).To(BeTrue())
		Expect(got).To(Equal("1.27"))

		_, ok = types.NormalizeGoMajorMinor("devel")
		Expect(ok).To(BeFalse())
	})

	It("CompareGoMajorMinor forwards to GoVersion.Compare", func() {
		Expect(types.CompareGoMajorMinor("1.9", "1.10")).To(Equal(-1))
		Expect(types.CompareGoMajorMinor("devel", "1.27")).To(Equal(0))
	})
})
