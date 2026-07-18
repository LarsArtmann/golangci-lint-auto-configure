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

var _ = Describe("DefaultLinterSettings", func() {
	It("should produce non-empty ToMap output for every entry", func() {
		for linter, settings := range constants.DefaultLinterSettings {
			m := settings.ToMap()
			Expect(m).
				ToNot(BeEmpty(), "DefaultLinterSettings[%q].ToMap() returned empty map", linter)
		}
	})

	It("should have cyclop max-complexity as int in ToMap output", func() {
		m := constants.DefaultLinterSettings["cyclop"].ToMap()
		maxComplexity, ok := m["max-complexity"]
		Expect(ok).To(BeTrue(), "cyclop settings missing max-complexity key")
		Expect(maxComplexity).To(Equal(12))
	})
})

var _ = Describe("DefaultFormatterSettings", func() {
	It("should produce non-empty ToMap output for every entry", func() {
		for formatter, settings := range constants.DefaultFormatterSettings {
			m := settings.ToMap()
			Expect(m).
				ToNot(BeEmpty(), "DefaultFormatterSettings[%q].ToMap() returned empty map", formatter)
		}
	})

	It("should have golines max-len as int in ToMap output", func() {
		m := constants.DefaultFormatterSettings["golines"].ToMap()
		maxLen, ok := m["max-len"]
		Expect(ok).To(BeTrue(), "golines settings missing max-len key")
		Expect(maxLen).To(Equal(120))
	})
})

var _ = Describe("Format preset", func() {
	It("should exist in PresetLinters", func() {
		_, ok := constants.PresetLinters["format"]
		Expect(ok).To(BeTrue(), "format preset missing from PresetLinters")
	})

	It("should exist in PresetFormatters with core formatters", func() {
		formatters, ok := constants.PresetFormatters["format"]
		Expect(ok).To(BeTrue(), "format preset missing from PresetFormatters")
		Expect(formatters).To(ContainElement(types.FormatterName("gci")))
		Expect(formatters).To(ContainElement(types.FormatterName("gofumpt")))
		Expect(formatters).To(ContainElement(types.FormatterName("goimports")))
	})

	It("should exist in PresetDescriptions", func() {
		desc, ok := constants.PresetDescriptions["format"]
		Expect(ok).To(BeTrue(), "format preset missing from PresetDescriptions")
		Expect(desc).ToNot(BeEmpty())
	})

	It("should be listed in ValidPresets", func() {
		Expect(constants.ValidPresets).To(ContainSubstring("format"))
	})

	It("should have formatters in canonical order matching FormatterOrder", func() {
		formatters := constants.PresetFormatters["format"]
		Expect(formatters).To(HaveLen(3))

		for i, f := range formatters {
			expected := constants.FormatterOrder[i]
			Expect(string(f)).To(Equal(expected),
				"format preset formatter at index %d: got %s, want %s", i, f, expected)
		}
	})
})

var _ = Describe("DefaultLinterSettings ToMap equivalence", func() {
	It("ireturn should produce correct allow list", func() {
		m := constants.DefaultLinterSettings["ireturn"].ToMap()
		allow, ok := m["allow"]
		Expect(ok).To(BeTrue(), "ireturn settings missing allow key")
		allowSlice, ok := allow.([]any)
		Expect(ok).To(BeTrue(), "ireturn allow is not []any")
		Expect(allowSlice).To(HaveLen(5))
	})

	It("gocritic should produce disabled-checks list", func() {
		m := constants.DefaultLinterSettings["gocritic"].ToMap()
		checks, ok := m["disabled-checks"]
		Expect(ok).To(BeTrue(), "gocritic settings missing disabled-checks key")
		checksSlice, ok := checks.([]any)
		Expect(ok).To(BeTrue(), "gocritic disabled-checks is not []any")
		Expect(checksSlice).To(ContainElement(ContainSubstring("ifElseChain")))
	})

	It("exhaustruct should produce exclude list", func() {
		m := constants.DefaultLinterSettings["exhaustruct"].ToMap()
		exclude, ok := m["exclude"]
		Expect(ok).To(BeTrue(), "exhaustruct settings missing exclude key")
		excludeSlice, ok := exclude.([]any)
		Expect(ok).To(BeTrue(), "exhaustruct exclude is not []any")
		Expect(excludeSlice).To(ContainElement(ContainSubstring("os/exec.Cmd")))
	})

	It("revive should produce rules with disabled entries", func() {
		m := constants.DefaultLinterSettings["revive"].ToMap()
		rules, ok := m["rules"]
		Expect(ok).To(BeTrue(), "revive settings missing rules key")
		rulesSlice, ok := rules.([]any)
		Expect(ok).To(BeTrue(), "revive rules is not []any")
		Expect(rulesSlice).To(HaveLen(2))
	})

	It("varnamelen should produce ignore flags and names", func() {
		m := constants.DefaultLinterSettings["varnamelen"].ToMap()
		Expect(m["ignore-map-index-ok"]).To(BeTrue())
		Expect(m["ignore-type-assert-ok"]).To(BeTrue())
		names, ok := m["ignore-names"]
		Expect(ok).To(BeTrue())
		namesSlice, ok := names.([]any)
		Expect(ok).To(BeTrue())
		Expect(namesSlice).To(HaveLen(11))
	})

	It("gomoddirectives should produce replace-local bool", func() {
		m := constants.DefaultLinterSettings["gomoddirectives"].ToMap()
		Expect(m["replace-local"]).To(BeTrue())
	})

	It("ginkgolinter should produce forbid flags", func() {
		m := constants.DefaultLinterSettings["ginkgolinter"].ToMap()
		Expect(m["forbid-focus-container"]).To(BeTrue())
		Expect(m["forbid-spec-pollution"]).To(BeTrue())
	})

	It("testifylint should produce enable-all and disable list", func() {
		m := constants.DefaultLinterSettings["testifylint"].ToMap()
		Expect(m["enable-all"]).To(BeTrue())
		disable, ok := m["disable"]
		Expect(ok).To(BeTrue())
		disableSlice, ok := disable.([]any)
		Expect(ok).To(BeTrue())
		Expect(disableSlice).To(ContainElement(ContainSubstring("go-require")))
	})

	It("makezero should produce always bool", func() {
		m := constants.DefaultLinterSettings["makezero"].ToMap()
		Expect(m["always"]).To(BeTrue())
	})
})
