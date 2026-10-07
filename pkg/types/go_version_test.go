package types_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NormalizeGoMajorMinor", func() {
	DescribeTable("reducing Go version strings to major.minor",
		func(input, want string, wantOK bool) {
			got, ok := types.NormalizeGoMajorMinor(input)
			Expect(ok).To(Equal(wantOK))
			Expect(got).To(Equal(want))
		},
		Entry("go-prefixed patch version", "go1.27.1", "1.27", true),
		Entry("bare patch version", "1.27.1", "1.27", true),
		Entry("zero patch version is a no-op", "1.27.0", "1.27", true),
		Entry("major.minor only", "1.27", "1.27", true),
		Entry("older version", "1.26.7", "1.26", true),
		Entry("empty", "", "", false),
		Entry("devel keyword", "devel", "", false),
		Entry("devel pseudo version", "go1.28-0f9a9bc", "", false),
		Entry("non-numeric minor", "1.x", "", false),
		Entry("major only", "1", "", false),
	)
})

var _ = Describe("CompareGoMajorMinor", func() {
	DescribeTable("comparing Go versions at major.minor granularity",
		func(a, b string, want int) {
			Expect(types.CompareGoMajorMinor(a, b)).To(Equal(want))
		},
		Entry("newer minor", "1.27", "1.26", 1),
		Entry("older minor", "1.26", "1.27", -1),
		Entry("equal", "1.27", "1.27", 0),
		Entry("equal ignoring patch", "1.27.0", "go1.27.9", 0),
		Entry("numeric minor not lexicographic", "1.9", "1.10", -1),
		Entry("newer major", "2.0", "1.99", 1),
		Entry("unparsable compares equal", "devel", "1.27", 0),
	)
})
