package constants_test

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExperiments(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Experiments Suite")
}

var _ = Describe("GoExperiments", func() {
	It("should have all required fields populated", func() {
		for _, exp := range constants.GoExperiments {
			Expect(exp.Tag).NotTo(BeEmpty())
			Expect(exp.Package).NotTo(BeEmpty())
			Expect(exp.Description).NotTo(BeEmpty())
			Expect(exp.Tag).To(HavePrefix("goexperiment."))
		}
	})

	It("should have unique tags", func() {
		tags := constants.GoExperimentTags()
		unique := types.NewSet[string]()

		for _, tag := range tags {
			Expect(unique.Contains(tag)).To(BeFalse(), "duplicate tag: %s", tag)
			unique.Add(tag)
		}
	})

	It("should have unique packages", func() {
		pkgs := types.NewSet[string]()

		for _, exp := range constants.GoExperiments {
			Expect(pkgs.Contains(exp.Package)).To(BeFalse(), "duplicate package: %s", exp.Package)
			pkgs.Add(exp.Package)
		}
	})

	It("should include all expected experiments", func() {
		tags := constants.GoExperimentTags()

		expected := []string{
			"goexperiment.arenas",
			"goexperiment.goroutineleakprofile",
			"goexperiment.jsonv2",
			"goexperiment.runtimesecret",
			"goexperiment.simd",
		}

		for _, exp := range expected {
			Expect(tags).To(ContainElement(exp))
		}

		Expect(tags).To(HaveLen(len(expected)))
	})
})

var _ = Describe("GoExperimentTags", func() {
	It("should return tags matching experiment data", func() {
		tags := constants.GoExperimentTags()
		Expect(tags).To(HaveLen(len(constants.GoExperiments)))

		tagSet := types.NewSet(tags...)

		for _, exp := range constants.GoExperiments {
			Expect(tagSet).To(HaveKey(exp.Tag))
		}
	})
})

var _ = Describe("GoExperiment type", func() {
	It("should store and retrieve fields", func() {
		exp := types.GoExperiment{
			Tag:         "goexperiment.test",
			Package:     "test",
			Description: "test experiment",
		}

		Expect(exp.Tag).To(Equal("goexperiment.test"))
		Expect(exp.Package).To(Equal("test"))
		Expect(exp.Description).To(Equal("test experiment"))
	})
})
