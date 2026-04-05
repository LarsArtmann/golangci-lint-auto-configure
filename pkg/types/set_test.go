package types_test

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Set Suite")
}

var _ = Describe("Set", func() {
	It("should create an empty set", func() {
		s := types.NewSet[string]()

		Expect(s.Len()).To(Equal(0))
	})

	It("should create a set from items", func() {
		s := types.NewSet("a", "b", "c")

		Expect(s.Len()).To(Equal(3))
		Expect(s.Contains("a")).To(BeTrue())
		Expect(s.Contains("b")).To(BeTrue())
		Expect(s.Contains("c")).To(BeTrue())
	})

	It("should deduplicate items on creation", func() {
		s := types.NewSet("a", "b", "a")

		Expect(s.Len()).To(Equal(2))
	})

	It("should add and delete items", func() {
		s := types.NewSet[string]()
		s.Add("x")

		Expect(s.Contains("x")).To(BeTrue())

		s.Delete("x")

		Expect(s.Contains("x")).To(BeFalse())
	})

	It("should convert to unsorted slice", func() {
		s := types.NewSet("a", "b", "c")
		slice := s.ToSlice()

		Expect(slice).To(HaveLen(3))
		Expect(slice).To(ContainElements("a", "b", "c"))
	})

	It("should convert to sorted slice", func() {
		s := types.NewSet("c", "a", "b")
		slice := types.ToSortedSlice(s)

		Expect(slice).To(Equal([]string{"a", "b", "c"}))
	})

	It("should work with int type", func() {
		s := types.NewSet(3, 1, 2)
		slice := types.ToSortedSlice(s)

		Expect(slice).To(Equal([]int{1, 2, 3}))
	})
})
