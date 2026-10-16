package finding

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("deviationTags", func() {
	DescribeTable("keeps only convention-valid tags",
		func(element string, want []string) {
			tags := deviationTags(element)

			got := make([]string, 0, len(tags))
			for _, tag := range tags {
				got = append(got, string(tag))
			}

			Expect(got).To(Equal(want))
		},
		Entry("plain linter name keeps its tag", "gosec", []string{"gosec"}),
		Entry("underscored name normalizes", "exhaustruct_v5", []string{"exhaustruct-v5"}),
		Entry(
			"rename element carries no tag instead of crashing finding construction",
			"exhaustruct->exhaustruct_v5",
			[]string{},
		),
		Entry("path pattern carries no tag", `_templ\.go$`, []string{}),
		Entry("empty element carries no tag", "", []string{}),
	)
})

func TestDeviationTags(t *testing.T) {
	t.Parallel()

	if tags := deviationTags("exhaustruct->exhaustruct_v5"); len(tags) != 0 {
		t.Fatalf("rename element must carry no tag, got %v", tags)
	}

	if tags := deviationTags("gosec"); len(tags) != 1 || string(tags[0]) != "gosec" {
		t.Fatalf("plain linter name must keep its tag, got %v", tags)
	}
}
