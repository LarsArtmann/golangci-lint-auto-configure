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
		set := types.NewSet[string]()
		set.Add("x")

		Expect(set.Contains("x")).To(BeTrue())

		set.Delete("x")

		Expect(set.Contains("x")).To(BeFalse())
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

	It("should report IsEmpty correctly", func() {
		empty := types.NewSet[string]()
		Expect(empty.IsEmpty()).To(BeTrue())

		nonEmpty := types.NewSet("a")
		Expect(nonEmpty.IsEmpty()).To(BeFalse())
	})

	It("should clone a set", func() {
		original := types.NewSet("a", "b")
		cloned := original.Clone()

		Expect(cloned.Len()).To(Equal(2))
		Expect(cloned.Contains("a")).To(BeTrue())
		Expect(cloned.Contains("b")).To(BeTrue())

		cloned.Delete("a")
		Expect(original.Contains("a")).To(BeTrue())
	})

	It("should union two sets", func() {
		set1 := types.NewSet("a", "b")
		set2 := types.NewSet("b", "c")
		merged := set1.Union(set2)

		Expect(merged.Len()).To(Equal(3))
		Expect(merged.Contains("a")).To(BeTrue())
		Expect(merged.Contains("b")).To(BeTrue())
		Expect(merged.Contains("c")).To(BeTrue())
	})

	Context("Set Operations", func() {
		It("should compute difference", func() {
			set1 := types.NewSet("a", "b", "c")
			set2 := types.NewSet("b", "c")
			diff := set1.Difference(set2)

			Expect(diff.Len()).To(Equal(1))
			Expect(diff.Contains("a")).To(BeTrue())
			Expect(diff.Contains("b")).To(BeFalse())
		})

		It("should compute intersection", func() {
			set1 := types.NewSet("a", "b", "c")
			set2 := types.NewSet("b", "c", "d")
			intersection := set1.Intersect(set2)

			Expect(intersection.Len()).To(Equal(2))
			Expect(intersection.Contains("b")).To(BeTrue())
			Expect(intersection.Contains("c")).To(BeTrue())
			Expect(intersection.Contains("a")).To(BeFalse())
		})

		It("should compare equality", func() {
			set1 := types.NewSet("a", "b", "c")
			set2 := types.NewSet("a", "b", "c")
			set3 := types.NewSet("a", "b")

			Expect(set1.Equal(set2)).To(BeTrue())
			Expect(set1.Equal(set3)).To(BeFalse())
		})

		It("should check subset", func() {
			empty := types.NewSet[string]()
			small := types.NewSet("a")
			medium := types.NewSet("a", "b")
			large := types.NewSet("a", "b", "c")

			Expect(empty.IsSubset(large)).To(BeTrue())
			Expect(small.IsSubset(medium)).To(BeTrue())
			Expect(medium.IsSubset(medium)).To(BeTrue())
			Expect(large.IsSubset(small)).To(BeFalse())
		})

		It("should check superset", func() {
			empty := types.NewSet[string]()
			small := types.NewSet("a")
			medium := types.NewSet("a", "b")
			large := types.NewSet("a", "b", "c")

			Expect(large.IsSuperset(empty)).To(BeTrue())
			Expect(large.IsSuperset(small)).To(BeTrue())
			Expect(medium.IsSuperset(medium)).To(BeTrue())
			Expect(small.IsSuperset(large)).To(BeFalse())
		})

		It("should check proper subset", func() {
			small := types.NewSet("a")
			medium := types.NewSet("a", "b")

			Expect(small.IsProperSubset(medium)).To(BeTrue())
			Expect(medium.IsProperSubset(medium)).To(BeFalse())
			Expect(medium.IsProperSubset(small)).To(BeFalse())
		})

		It("should check proper superset", func() {
			small := types.NewSet("a")
			medium := types.NewSet("a", "b")

			Expect(medium.IsProperSuperset(small)).To(BeTrue())
			Expect(medium.IsProperSuperset(medium)).To(BeFalse())
			Expect(small.IsProperSuperset(medium)).To(BeFalse())
		})
	})

	Context("Edge Cases", func() {
		It("should handle difference with empty set", func() {
			set1 := types.NewSet("a", "b")
			empty := types.NewSet[string]()

			Expect(set1.Difference(empty)).To(Equal(set1))
			Expect(empty.Difference(set1)).To(BeEmpty())
		})

		It("should handle intersection with empty set", func() {
			set1 := types.NewSet("a", "b")
			empty := types.NewSet[string]()

			Expect(set1.Intersect(empty)).To(BeEmpty())
			Expect(empty.Intersect(set1)).To(BeEmpty())
		})

		It("should handle union with empty set", func() {
			set1 := types.NewSet("a", "b")
			empty := types.NewSet[string]()
			union := set1.Union(empty)

			Expect(union).To(Equal(set1))
		})
	})
})
