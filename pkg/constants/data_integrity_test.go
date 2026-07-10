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

var _ = Describe("DisabledLinters", func() {
	It("should not have entries in LinterPriorities", func() {
		for linter := range constants.DisabledLinters {
			_, exists := constants.LinterPriorities[linter]
			Expect(exists).
				To(BeFalse(), "DisabledLinters contains %q which also has a LinterPriorities entry — disabled linters must never be recommended", linter)
		}
	})

	It("should not have entries in LinterReasons", func() {
		for linter := range constants.DisabledLinters {
			_, exists := constants.LinterReasons[linter]
			Expect(exists).
				To(BeFalse(), "DisabledLinters contains %q which also has a LinterReasons entry — disabled linters must never be recommended", linter)
		}
	})

	It("should have a non-empty reason for every entry", func() {
		for linter, reason := range constants.DisabledLinters {
			Expect(reason).
				ToNot(BeEmpty(), "DisabledLinters contains %q with an empty reason — every disabled linter must explain why it is disabled", linter)
		}
	})
})

var _ = Describe("LinterPriorities and LinterReasons consistency", func() {
	It("should have exactly the same keys in LinterPriorities and LinterReasons", func() {
		for linter := range constants.LinterPriorities {
			_, exists := constants.LinterReasons[linter]
			Expect(exists).
				To(BeTrue(), "LinterPriorities contains %q which is missing from LinterReasons", linter)
		}

		for linter := range constants.LinterReasons {
			_, exists := constants.LinterPriorities[linter]
			Expect(exists).
				To(BeTrue(), "LinterReasons contains %q which is missing from LinterPriorities", linter)
		}
	})
})

var _ = Describe("Linter and formatter separation", func() {
	It("should not have any formatter names in LinterPriorities", func() {
		for formatter := range constants.FormatterInfo {
			_, exists := constants.LinterPriorities[types.LinterName(formatter)]
			Expect(exists).
				To(BeFalse(), "Formatter %q is incorrectly listed in LinterPriorities — formatters must not appear in linter maps", formatter)
		}
	})

	It("should not have any formatter names in LinterReasons", func() {
		for formatter := range constants.FormatterInfo {
			_, exists := constants.LinterReasons[types.LinterName(formatter)]
			Expect(exists).
				To(BeFalse(), "Formatter %q is incorrectly listed in LinterReasons — formatters must not appear in linter maps", formatter)
		}
	})
})

var _ = Describe("DeprecatedLinters", func() {
	It("should have a non-empty reason for every entry", func() {
		for linter, replacement := range constants.DeprecatedLinters {
			Expect(replacement.Reason).
				ToNot(BeEmpty(), "DeprecatedLinters contains %q with an empty reason", linter)
		}
	})

	It("should have non-empty replacement targets that exist in LinterPriorities", func() {
		for linter, replacement := range constants.DeprecatedLinters {
			if replacement.Replacement == "" {
				continue
			}

			_, exists := constants.LinterPriorities[replacement.Replacement]
			Expect(exists).
				To(BeTrue(), "DeprecatedLinters[%q] replacement %q is missing from LinterPriorities", linter, replacement.Replacement)
		}
	})
})

var _ = Describe("FormatterPriorities and FormatterReasons consistency", func() {
	It("should have exactly the same keys in FormatterPriorities and FormatterReasons", func() {
		for formatter := range constants.FormatterPriorities {
			_, exists := constants.FormatterReasons[formatter]
			Expect(exists).
				To(BeTrue(), "FormatterPriorities contains %q which is missing from FormatterReasons", formatter)
		}

		for formatter := range constants.FormatterReasons {
			_, exists := constants.FormatterPriorities[formatter]
			Expect(exists).
				To(BeTrue(), "FormatterReasons contains %q which is missing from FormatterPriorities", formatter)
		}
	})
})
