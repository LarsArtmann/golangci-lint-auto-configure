package constants_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs assert the invariants of the policy table itself. Many of the
// old map-level data-integrity specs were deleted when the table landed:
// with one record per linter and one Tier field per record, tier overlap and
// map-key divergence are structurally impossible, and the init() in
// linter_policy_views.go panics on invalid records before any test runs.

var _ = Describe("Policy table", func() {
	It("has unique names (structural)", func() {
		names := make(map[types.LinterName]bool, len(constants.LinterPolicies))
		for _, record := range constants.LinterPolicies {
			Expect(names[record.Name]).
				To(BeFalse(), "duplicate policy record for %q", record.Name)
			names[record.Name] = true
		}
	})

	It("every record passes Validate", func() {
		for _, record := range constants.LinterPolicies {
			Expect(record.Validate()).To(Succeed(), "record %q", record.Name)
		}
	})

	It("never puts a disabled-tier linter into the recommendation views", func() {
		for linter := range constants.DisabledLinters {
			_, hasPriority := constants.LinterPriorities[linter]
			Expect(hasPriority).
				To(BeFalse(), "disabled linter %q leaked into LinterPriorities", linter)

			_, hasReason := constants.LinterReasons[linter]
			Expect(hasReason).
				To(BeFalse(), "disabled linter %q leaked into LinterReasons", linter)
		}
	})

	It("gives never-auto-enable linters priority and rationale (manual enables need report data)", func() {
		for linter := range constants.NeverAutoEnableLinters {
			_, hasPriority := constants.LinterPriorities[linter]
			Expect(hasPriority).
				To(BeTrue(), "never-auto-enable linter %q missing from LinterPriorities", linter)

			_, hasReason := constants.LinterReasons[linter]
			Expect(hasReason).
				To(BeTrue(), "never-auto-enable linter %q missing from LinterReasons", linter)
		}
	})

	It("keeps supersession targets inside the table (referential integrity)", func() {
		known := constants.LinterPolicyView()

		for _, record := range constants.LinterPolicies {
			if record.SupersededBy == "" {
				continue
			}

			_, exists := known[record.SupersededBy]
			Expect(exists).
				To(BeTrue(), "%q supersedes unknown linter %q", record.Name, record.SupersededBy)
		}
	})

	It("agrees with itself on gated replacements (deprecation since == replacement min version)", func() {
		known := constants.LinterPolicyView()

		for _, record := range constants.LinterPolicies {
			if record.SupersededBy == "" || record.SupersessionSince.IsZero() {
				continue
			}

			replacement, exists := known[record.SupersededBy]
			Expect(exists).To(BeTrue())

			Expect(replacement.ValidTool.Min).
				To(Equal(record.SupersessionSince),
					"%q says replacement %q exists since %s, but the replacement's own ValidTool.Min disagrees",
					record.Name, record.SupersededBy, record.SupersessionSince)
		}
	})

	It("carries friction evidence for the measured high-friction linters", func() {
		for _, name := range []types.LinterName{
			"exhaustruct_v5", "gochecknoglobals", "wrapcheck", "ireturn", "funlen",
		} {
			record, known := constants.PolicyRecord(name)
			Expect(known).To(BeTrue(), "%q missing from the table", name)
			Expect(record.Friction).
				ToNot(BeNil(), "%q has a documented friction ratio but no Evidence", name)
			Expect(record.Friction.Ratio).To(BeNumerically(">", 0))
		}
	})

	It("maps every linter golangci-lint v2.14.0 reports as deprecated", func() {
		for _, deprecatedInV2 := range []types.LinterName{"wsl", "gomodguard", "exhaustruct"} {
			record, known := constants.PolicyRecord(deprecatedInV2)
			Expect(known).To(BeTrue(), "%q missing from the table", deprecatedInV2)
			Expect(record.SupersessionReason).
				ToNot(BeEmpty(), "%q is deprecated upstream but carries no supersession", deprecatedInV2)
		}
	})

	It("gives gohumanize the plugin gate", func() {
		record, known := constants.PolicyRecord("gohumanize")
		Expect(known).To(BeTrue())
		Expect(record.Gate).To(Equal(types.RequiresTech("go-humanize")))
		Expect(record.Rationale).To(ContainSubstring("module plugin"))
	})
})

var _ = Describe("Formatter policy table", func() {
	It("has unique names and valid records", func() {
		names := make(map[types.FormatterName]bool, len(constants.FormatterPolicies))
		for _, record := range constants.FormatterPolicies {
			Expect(record.Validate()).To(Succeed(), "formatter record %q", record.Name)
			Expect(names[record.Name]).To(BeFalse(), "duplicate formatter record %q", record.Name)
			names[record.Name] = true
		}
	})

	It("every FormatterOrder entry references a known record", func() {
		for _, name := range constants.FormatterOrder {
			_, known := constants.FormatterPolicyRecord(name)
			Expect(known).To(BeTrue(), "FormatterOrder references unknown formatter %q", name)
		}
	})

	It("marks exactly the validated core stack as Core, in canonical order", func() {
		Expect(constants.CoreFormatters).To(Equal([]types.FormatterName{"goimports", "gofumpt", "golines"}))
	})

	It("keeps gci as a manual-only choice (known, ordered, not core)", func() {
		_, known := constants.FormatterPolicyRecord("gci")
		Expect(known).To(BeTrue())

		Expect(constants.FormatterOrder).To(ContainElement(types.FormatterName("gci")))
		Expect(constants.CoreFormatters).NotTo(ContainElement(types.FormatterName("gci")),
			"gci was removed from the default stack 2026-10-11")
	})
})
