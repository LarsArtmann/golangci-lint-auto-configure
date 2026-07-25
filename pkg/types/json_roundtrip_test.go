package types_test

import (
	"encoding/json/v2"
	"errors"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON Round-Trip Serialization", func() {
	Describe("Report types with all fields populated", func() {
		It("round-trips LinterInfo", func() {
			original := types.LinterInfo{
				Name:        "gosec",
				Description: "Security checks",
				Groups:      []string{"security", "bugs"},
				Fast:        true,
				AutoFix:     true,
				Deprecated:  true,
				Since:       "v1.0.0",
				OriginalURL: "https://example.com/gosec",
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.LinterInfo
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips LinterInfo with zero-value optional fields", func() {
			original := types.LinterInfo{
				Name:        "govet",
				Description: "Vet checks",
				Since:       "v1.0.0",
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.LinterInfo
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips FormatterInfo", func() {
			original := types.FormatterInfo{
				Name:        "gofmt",
				Description: "Formats Go code",
				AutoFix:     true,
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.FormatterInfo
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips LinterRecommendation", func() {
			original := types.LinterRecommendation{
				Name:     "errcheck",
				Priority: types.LinterPriorityCritical,
				Reason:   "Unchecked errors cause bugs",
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.LinterRecommendation
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips FormatterRecommendation", func() {
			original := types.FormatterRecommendation{
				Name:     "gofumpt",
				Priority: types.FormatterPriorityHigh,
				Reason:   "Stricter gofmt",
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.FormatterRecommendation
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips LinterReplacement", func() {
			original := types.LinterReplacement{
				Replacement: "wsl_v5",
				Reason:      "wsl deprecated since v2.2.0",
				MinVersion:  "v2.2.0",
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.LinterReplacement
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips ValidationError with Line", func() {
			original := types.ValidationError{
				Field:   "linters.enable",
				Message: "unknown linter: foobar",
				Line:    42,
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.ValidationError
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips ValidationError without Line", func() {
			original := types.ValidationError{
				Field:   "linters.disable",
				Message: "empty list",
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.ValidationError
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips MigrationResult", func() {
			original := types.MigrationResult{
				FixesApplied: 5,
				Message:      "migration complete",
				NextSteps:    []string{"review changes", "run golangci-lint"},
				DryRun:       true,
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.MigrationResult
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})

		It("round-trips ValidationResult", func() {
			original := types.ValidationResult{
				Valid: false,
				Errors: []types.ValidationError{
					{Field: "version", Message: "too old"},
					{Field: "linters.enable", Message: "duplicate", Line: 10},
				},
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.ValidationResult
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt).To(Equal(original))
		})
	})

	Describe("ConfigAnalysis with nested types", func() {
		It("round-trips the complete analysis object", func() {
			original := types.ConfigAnalysis{
				ConfigPath:      ".golangci.yml",
				EnabledLinters:  []types.LinterInfo{{Name: "govet", Description: "Vet", Since: "v1.0.0"}},
				DisabledLinters: []types.LinterInfo{{Name: "gosec", Description: "Security", Since: "v1.0.0"}},
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "gosec", Priority: types.LinterPriorityCritical, Reason: "Security"},
				},
				CriticalCount:    1,
				HighValueCount:   2,
				MediumValueCount: 3,
				OptionalCount:    4,
				DeprecatedCount:  0,
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			var rt types.ConfigAnalysis
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt.ConfigPath).To(Equal(original.ConfigPath))
			Expect(rt.EnabledLinters).To(Equal(original.EnabledLinters))
			Expect(rt.DisabledLinters).To(Equal(original.DisabledLinters))
			Expect(rt.LinterRecommendations).To(Equal(original.LinterRecommendations))
			Expect(rt.CriticalCount).To(Equal(1))
			Expect(rt.HighValueCount).To(Equal(2))
			Expect(rt.MediumValueCount).To(Equal(3))
			Expect(rt.OptionalCount).To(Equal(4))
		})
	})

	Describe("MigrationResult Error field exclusion", func() {
		It("does not serialize the Error field", func() {
			original := types.MigrationResult{
				FixesApplied: 1,
				Message:      "done",
				Error:        errors.New("something went wrong"),
			}
			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			var rt types.MigrationResult
			Expect(json.Unmarshal(data, &rt)).To(Succeed())
			Expect(rt.Error).To(BeNil())
		})
	})

	Describe("Priority enum round-trips as integer", func() {
		DescribeTable("preserves priority values",
			func(priority types.LinterPriority) {
				rec := types.LinterRecommendation{
					Name:     "test",
					Priority: priority,
					Reason:   "test",
				}
				data, err := json.Marshal(rec)
				Expect(err).NotTo(HaveOccurred())

				var rt types.LinterRecommendation
				Expect(json.Unmarshal(data, &rt)).To(Succeed())
				Expect(rt.Priority).To(Equal(priority))
			},
			Entry("Critical", types.LinterPriorityCritical),
			Entry("High", types.LinterPriorityHigh),
			Entry("Medium", types.LinterPriorityMedium),
			Entry("Optional", types.LinterPriorityOptional),
		)
	})
})
