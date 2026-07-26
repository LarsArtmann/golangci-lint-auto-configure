package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCoverageCheck(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Coverage Check Suite")
}

var _ = Describe("parseTotalPercentage", func() {
	DescribeTable(
		"parses the total line from go tool cover output",
		func(output string, expected float64) {
			percent, err := parseTotalPercentage(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(percent).To(Equal(expected))
		},
		Entry("realistic multi-line output", sampleCoverOutput, 72.5),
		Entry("total only", "total:\t\t(stats)\t72.5%\n", 72.5),
		Entry("exactly 100 percent", "total:\t\t(stats)\t100.0%\n", 100.0),
		Entry("exactly 0 percent", "total:\t\t(stats)\t0.0%\n", 0.0),
		Entry("high precision decimal", "total:\t\t(stats)\t85.392%\n", 85.392),
		Entry("trailing newline absent", "total:\t\t(stats)\t42.0%", 42.0),
	)

	DescribeTable(
		"returns an error for invalid output",
		func(output string, expectedErr string) {
			_, err := parseTotalPercentage(output)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErr))
		},
		Entry("no total line at all", "github.com/foo/bar/main.go:10: SomeFunc\t100.0%\n", "no total line"),
		Entry("empty output", "", "no total line"),
		Entry("total line with no percentage field", "total:", "unexpected total line format"),
	)

	It("handles output with only per-file lines (no total)", func() {
		output := "github.com/foo/bar/a.go:10: FuncA\t75.0%\ngithub.com/foo/bar/b.go:5: FuncB\t80.0%\n"
		_, err := parseTotalPercentage(output)
		Expect(err).To(MatchError(ContainSubstring("no total line")))
	})
})

const sampleCoverOutput = `github.com/foo/bar/main.go:10:		Init		100.0%
github.com/foo/bar/config.go:15:	Load		80.0%
github.com/foo/bar/utils.go:20:	Helper		50.0%
total:			(stats)	72.5%
`
