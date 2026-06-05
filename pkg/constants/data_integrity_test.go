package constants_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/semver"
)

var _ = Describe("LinterMinVersions", func() {
	It("should have valid semver for every entry", func() {
		for linter, version := range constants.LinterMinVersions {
			Expect(semver.IsValid(version)).
				To(BeTrue(), "LinterMinVersions[%s] = %q is not valid semver", linter, version)
		}
	})

	It("should reference linters that exist in LinterPriorities", func() {
		for linter := range constants.LinterMinVersions {
			_, exists := constants.LinterPriorities[linter]
			Expect(exists).
				To(BeTrue(), "LinterMinVersions references %q which is missing from LinterPriorities", linter)
		}
	})
})

var _ = Describe("Reference preset", func() {
	It("should only contain linters present in LinterPriorities", func() {
		referenceLinters, ok := constants.PresetLinters["reference"]
		Expect(ok).To(BeTrue(), "reference preset must exist")

		for _, linter := range referenceLinters {
			_, exists := constants.LinterPriorities[linter]
			Expect(exists).
				To(BeTrue(), "reference preset contains %q which is missing from LinterPriorities", linter)
		}
	})

	It("should contain all critical and high priority linters", func() {
		referenceLinters := constants.PresetLinters["reference"]
		referenceSet := make(map[string]struct{}, len(referenceLinters))

		for _, l := range referenceLinters {
			referenceSet[string(l)] = struct{}{}
		}

		for linter, priority := range constants.LinterPriorities {
			if priority == types.LinterPriorityCritical || priority == types.LinterPriorityHigh {
				_, exists := referenceSet[string(linter)]
				Expect(exists).
					To(BeTrue(), "LinterPriorities[%s] (priority=%s) is missing from reference preset", linter, priority)
			}
		}
	})
})
