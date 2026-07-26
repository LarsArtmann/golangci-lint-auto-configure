package types_test

import (
	"encoding/json/v2"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON Tag Serialization", func() {
	Describe("Report types use PascalCase JSON keys", func() {
		Context("LinterInfo", func() {
			It("marshals all fields as PascalCase", func() {
				linter := types.LinterInfo{
					Name:        "gosec",
					Description: "Security checks",
					Groups:      []string{"security"},
					Fast:        true,
					AutoFix:     true,
					Deprecated:  false,
					Since:       "v1.0.0",
					OriginalURL: "https://example.com",
				}

				data, err := json.Marshal(linter)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("Name"))
				Expect(raw).To(HaveKey("Description"))
				Expect(raw).To(HaveKey("Groups"))
				Expect(raw).To(HaveKey("Fast"))
				Expect(raw).To(HaveKey("AutoFix"))
				Expect(raw).To(HaveKey("Deprecated"))
				Expect(raw).To(HaveKey("Since"))
				Expect(raw).To(HaveKey("OriginalURL"))

				Expect(raw).NotTo(HaveKey("name"))
				Expect(raw).NotTo(HaveKey("autoFix"))
				Expect(raw).NotTo(HaveKey("originalURL"))
			})

			It("omits empty optional fields", func() {
				linter := types.LinterInfo{
					Name:        "govet",
					Description: "Vet",
					Deprecated:  false,
					Since:       "v1.0.0",
				}

				data, err := json.Marshal(linter)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).NotTo(HaveKey("Groups"))
				Expect(raw).NotTo(HaveKey("Fast"))
				Expect(raw).NotTo(HaveKey("AutoFix"))
			})
		})

		Context("ConfigAnalysis", func() {
			It("marshals all fields as PascalCase", func() {
				analysis := types.ConfigAnalysis{
					ConfigPath:       ".golangci.yml",
					EnabledLinters:   []types.LinterInfo{{Name: "govet"}},
					DisabledLinters:  []types.LinterInfo{{Name: "gosec"}},
					CriticalCount:    2,
					HighValueCount:   3,
					MediumValueCount: 1,
					OptionalCount:    0,
					DeprecatedCount:  0,
				}

				data, err := json.Marshal(analysis)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("ConfigPath"))
				Expect(raw).To(HaveKey("EnabledLinters"))
				Expect(raw).To(HaveKey("DisabledLinters"))
				Expect(raw).To(HaveKey("CriticalCount"))
				Expect(raw).To(HaveKey("HighValueCount"))
				Expect(raw).To(HaveKey("MediumValueCount"))
				Expect(raw).To(HaveKey("OptionalCount"))
				Expect(raw).To(HaveKey("DeprecatedCount"))

				Expect(raw).NotTo(HaveKey("config_path"))
				Expect(raw).NotTo(HaveKey("enabled_linters"))
				Expect(raw).NotTo(HaveKey("critical_count"))
			})
		})

		Context("LinterReplacement", func() {
			It("marshals PascalCase and omits empty MinVersion", func() {
				rep := types.LinterReplacement{
					Replacement: "gocritic",
					Reason:      "old linter deprecated",
				}

				data, err := json.Marshal(rep)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("Replacement"))
				Expect(raw).To(HaveKey("Reason"))
				Expect(raw).NotTo(HaveKey("MinVersion"))
				Expect(raw).NotTo(HaveKey("minVersion"))
			})

			It("includes MinVersion when set", func() {
				rep := types.LinterReplacement{
					Replacement: "gocritic",
					Reason:      "replaced",
					MinVersion:  "v1.50.0",
				}

				data, err := json.Marshal(rep)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("MinVersion"))
				Expect(raw["MinVersion"]).To(Equal("v1.50.0"))
			})
		})

		Context("MigrationResult", func() {
			It("marshals PascalCase and omits Error from JSON", func() {
				result := types.MigrationResult{
					FixesApplied: 3,
					Message:      "success",
					NextSteps:    []string{"step1", "step2"},
					DryRun:       false,
				}

				data, err := json.Marshal(result)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("FixesApplied"))
				Expect(raw).To(HaveKey("Message"))
				Expect(raw).To(HaveKey("NextSteps"))
				Expect(raw).To(HaveKey("DryRun"))
				Expect(raw).NotTo(HaveKey("Error"))
				Expect(raw).NotTo(HaveKey("fixes_applied"))
				Expect(raw).NotTo(HaveKey("dry_run"))
			})

			It("omits empty NextSteps", func() {
				result := types.MigrationResult{
					FixesApplied: 0,
					Message:      "nothing to do",
				}

				data, err := json.Marshal(result)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).NotTo(HaveKey("NextSteps"))
			})
		})

		Context("ValidationResult", func() {
			It("marshals PascalCase and omits empty Errors", func() {
				result := types.ValidationResult{
					Valid: true,
				}

				data, err := json.Marshal(result)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("Valid"))
				Expect(raw).NotTo(HaveKey("Errors"))
				Expect(raw).NotTo(HaveKey("valid"))
			})
		})

		Context("HealthIssue", func() {
			It("marshals PascalCase and omits empty Suggestion", func() {
				issue := types.HealthIssue{
					Rule:    "duplicate-linter",
					Message: "found duplicate",
					Field:   "linters.enable",
				}

				data, err := json.Marshal(issue)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("Severity"))
				Expect(raw).To(HaveKey("Rule"))
				Expect(raw).To(HaveKey("Message"))
				Expect(raw).To(HaveKey("Field"))
				Expect(raw).NotTo(HaveKey("Suggestion"))
			})
		})
	})

	Describe("Config types use kebab-case JSON keys for round-trip safety", func() {
		Context("Config", func() {
			It("marshals JSON keys as kebab-case", func() {
				cfg := types.Config{
					Version: "2",
					Run: types.RunConfig{
						Timeout:              "5m",
						BuildTags:            []string{"foo"},
						ModulesDownloadMode:  "mod",
						AllowParallelRunners: true,
						AllowSerialRunners:   true,
						IssuesExitCode:       1,
						Tests:                true,
						Concurrency:          4,
						RelativePathMode:     "gomod",
					},
					Output: types.OutputConfig{
						Formats:    map[string]any{"json": nil},
						PathPrefix: "",
						PathMode:   "",
						SortOrder:  []string{"linter"},
						ShowStats:  true,
					},
					Linters: types.LintersConfig{
						Enable:  []types.LinterName{"govet"},
						Disable: []types.LinterName{"gosec"},
						Default: "standard",
					},
					Issues: types.IssuesConfig{
						MaxIssuesPerLinter: 50,
						MaxSameIssues:      10,
					},
				}

				data, err := json.Marshal(cfg)
				Expect(err).NotTo(HaveOccurred())

				var raw map[string]any
				Expect(json.Unmarshal(data, &raw)).To(Succeed())

				Expect(raw).To(HaveKey("version"))
				Expect(raw).To(HaveKey("run"))
				Expect(raw).To(HaveKey("output"))
				Expect(raw).To(HaveKey("linters"))
				Expect(raw).To(HaveKey("issues"))

				runRaw, ok := raw["run"].(map[string]any)
				Expect(ok).To(BeTrue())
				Expect(runRaw).To(HaveKey("timeout"))
				Expect(runRaw).To(HaveKey("build-tags"))
				Expect(runRaw).To(HaveKey("modules-download-mode"))
				Expect(runRaw).To(HaveKey("allow-parallel-runners"))
				Expect(runRaw).To(HaveKey("allow-serial-runners"))
				Expect(runRaw).To(HaveKey("issues-exit-code"))
				Expect(runRaw).To(HaveKey("relative-path-mode"))

				issuesRaw, ok := raw["issues"].(map[string]any)
				Expect(ok).To(BeTrue())
				Expect(issuesRaw).To(HaveKey("max-issues-per-linter"))
				Expect(issuesRaw).To(HaveKey("max-same-issues"))
			})
		})
	})
})
