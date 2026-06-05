package types_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigHealth", func() {
	Describe("CheckConfigHealth", func() {
		When("config is nil", func() {
			It("returns empty health", func() {
				health := types.CheckConfigHealth(nil)
				Expect(health.Issues).To(BeEmpty())
				Expect(health.IsHealthy()).To(BeTrue())
			})
		})

		When("config has no issues", func() {
			It("returns healthy", func() {
				cfg := validConfig()
				health := types.CheckConfigHealth(cfg)
				Expect(health.Issues).To(BeEmpty())
				Expect(health.IsHealthy()).To(BeTrue())
			})
		})
	})

	Describe("duplicate linters", func() {
		When("enable list has duplicates", func() {
			It("reports critical issue", func() {
				cfg := validConfig()
				cfg.Linters.Enable = append(cfg.Linters.Enable, "errcheck")

				health := types.CheckConfigHealth(cfg)

				Expect(health.Issues).To(HaveLen(1))
				Expect(health.Issues[0].Severity).To(Equal(types.HealthSeverityCritical))
				Expect(health.Issues[0].Rule).To(Equal(types.RuleDuplicateLinter))
				Expect(health.Issues[0].Message).To(ContainSubstring("errcheck"))
				Expect(health.Issues[0].Message).To(ContainSubstring("2 times"))
			})
		})

		When("disable list has duplicates", func() {
			It("reports warning", func() {
				cfg := validConfig()
				cfg.Linters.Disable = []string{"nlreturn", "nlreturn"}

				health := types.CheckConfigHealth(cfg)

				Expect(health.Issues).To(HaveLen(1))
				Expect(health.Issues[0].Severity).To(Equal(types.HealthSeverityWarning))
				Expect(health.Issues[0].Rule).To(Equal(types.RuleDuplicateLinter))
			})
		})

		When("both enable and disable have duplicates", func() {
			It("reports both", func() {
				cfg := validConfig()
				cfg.Linters.Enable = append(cfg.Linters.Enable, "errcheck", "staticcheck")
				cfg.Linters.Disable = []string{"nlreturn", "nlreturn"}

				health := types.CheckConfigHealth(cfg)

				Expect(health.Issues).To(HaveLen(3))
			})
		})
	})

	Describe("enable+disable overlap", func() {
		When("a linter is in both enable and disable", func() {
			It("reports warning", func() {
				cfg := validConfig()
				cfg.Linters.Enable = append(cfg.Linters.Enable, "nlreturn")
				cfg.Linters.Disable = []string{"nlreturn"}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleEnableDisableOverlap)
				Expect(issues).To(HaveLen(1))
				Expect(issues[0].Severity).To(Equal(types.HealthSeverityWarning))
				Expect(issues[0].Message).To(ContainSubstring("nlreturn"))
			})
		})

		When("multiple linters overlap", func() {
			It("reports each", func() {
				cfg := validConfig()
				cfg.Linters.Enable = append(cfg.Linters.Enable, "nlreturn", "godox")
				cfg.Linters.Disable = []string{"nlreturn", "godox"}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleEnableDisableOverlap)
				Expect(issues).To(HaveLen(2))
			})
		})
	})

	Describe("missing critical linters", func() {
		When("errcheck is missing", func() {
			It("reports warning", func() {
				cfg := validConfig()
				cfg.Linters.Enable = []string{"staticcheck", "govet"}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleMissingCriticalLinter)
				Expect(issues).To(HaveLen(1))
				Expect(issues[0].Message).To(ContainSubstring("errcheck"))
			})
		})

		When("all critical linters are missing", func() {
			It("reports each one", func() {
				cfg := validConfig()
				cfg.Linters.Enable = []string{"gosec"}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleMissingCriticalLinter)
				Expect(issues).To(HaveLen(3))
			})
		})

		When("critical linter is explicitly disabled", func() {
			It("does not report it as missing", func() {
				cfg := validConfig()
				cfg.Linters.Enable = []string{"staticcheck", "govet"}
				cfg.Linters.Disable = []string{"errcheck"}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleMissingCriticalLinter)
				Expect(issues).To(BeEmpty())
			})
		})
	})

	Describe("v1/v2 syntax mixing", func() {
		When("v2 config uses linters-settings", func() {
			It("reports warning", func() {
				cfg := validConfig()
				cfg.LintersSettingsV1 = map[string]any{
					"cyclop": map[string]any{"max-complexity": 15},
				}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleV1SyntaxInV2)
				Expect(issues).To(HaveLen(1))
				Expect(issues[0].Severity).To(Equal(types.HealthSeverityWarning))
			})
		})

		When("v1 config uses linters-settings", func() {
			It("does not report", func() {
				cfg := validConfig()
				cfg.Version = "1"
				cfg.LintersSettingsV1 = map[string]any{
					"cyclop": map[string]any{"max-complexity": 15},
				}

				health := types.CheckConfigHealth(cfg)

				issues := health.IssuesByRule(types.RuleV1SyntaxInV2)
				Expect(issues).To(BeEmpty())
			})
		})
	})

	Describe("IsHealthy", func() {
		When("no issues", func() {
			It("returns true", func() {
				health := &types.ConfigHealth{}
				Expect(health.IsHealthy()).To(BeTrue())
			})
		})

		When("only info issues", func() {
			It("returns true", func() {
				health := &types.ConfigHealth{
					Issues: []types.HealthIssue{
						{Severity: types.HealthSeverityInfo, Rule: "test"},
					},
				}
				Expect(health.IsHealthy()).To(BeTrue())
			})
		})

		When("has warning issues", func() {
			It("returns false", func() {
				health := &types.ConfigHealth{
					Issues: []types.HealthIssue{
						{Severity: types.HealthSeverityWarning, Rule: "test"},
					},
				}
				Expect(health.IsHealthy()).To(BeFalse())
			})
		})

		When("has critical issues", func() {
			It("returns false", func() {
				health := &types.ConfigHealth{
					Issues: []types.HealthIssue{
						{Severity: types.HealthSeverityCritical, Rule: "test"},
					},
				}
				Expect(health.IsHealthy()).To(BeFalse())
			})
		})
	})

	Describe("CriticalIssues", func() {
		It("filters to critical only", func() {
			health := &types.ConfigHealth{
				Issues: []types.HealthIssue{
					{Severity: types.HealthSeverityCritical, Rule: "a"},
					{Severity: types.HealthSeverityWarning, Rule: "b"},
					{Severity: types.HealthSeverityCritical, Rule: "c"},
					{Severity: types.HealthSeverityInfo, Rule: "d"},
				},
			}
			Expect(health.CriticalIssues()).To(HaveLen(2))
		})
	})

	Describe("WarningIssues", func() {
		It("filters to warning only", func() {
			health := &types.ConfigHealth{
				Issues: []types.HealthIssue{
					{Severity: types.HealthSeverityCritical, Rule: "a"},
					{Severity: types.HealthSeverityWarning, Rule: "b"},
					{Severity: types.HealthSeverityWarning, Rule: "c"},
				},
			}
			Expect(health.WarningIssues()).To(HaveLen(2))
		})
	})

	Describe("IssuesByRule", func() {
		It("filters issues by rule name", func() {
			health := &types.ConfigHealth{
				Issues: []types.HealthIssue{
					{Severity: types.HealthSeverityWarning, Rule: types.RuleDuplicateLinter},
					{Severity: types.HealthSeverityInfo, Rule: types.RuleV1SyntaxInV2},
					{Severity: types.HealthSeverityWarning, Rule: types.RuleDuplicateLinter},
				},
			}
			Expect(health.IssuesByRule("duplicate-linter")).To(HaveLen(2))
			Expect(health.IssuesByRule(types.RuleV1SyntaxInV2)).To(HaveLen(1))
			Expect(health.IssuesByRule("unknown")).To(BeEmpty())
		})
	})

	Describe("HasRule", func() {
		It("returns true when rule exists", func() {
			health := &types.ConfigHealth{
				Issues: []types.HealthIssue{
					{Severity: types.HealthSeverityWarning, Rule: types.RuleEnableDisableOverlap},
				},
			}
			Expect(health.HasRule(types.RuleEnableDisableOverlap)).To(BeTrue())
			Expect(health.HasRule(types.RuleDuplicateLinter)).To(BeFalse())
		})
	})

	Describe("CountBySeverity", func() {
		It("returns count for each severity", func() {
			health := &types.ConfigHealth{
				Issues: []types.HealthIssue{
					{Severity: types.HealthSeverityCritical, Rule: "a"},
					{Severity: types.HealthSeverityCritical, Rule: "b"},
					{Severity: types.HealthSeverityWarning, Rule: "c"},
					{Severity: types.HealthSeverityInfo, Rule: "d"},
				},
			}
			Expect(health.CountBySeverity(types.HealthSeverityCritical)).To(Equal(2))
			Expect(health.CountBySeverity(types.HealthSeverityWarning)).To(Equal(1))
			Expect(health.CountBySeverity(types.HealthSeverityInfo)).To(Equal(1))
		})
	})
})

func validConfig() *types.Config {
	return &types.Config{
		Version: "2",
		Run: types.RunConfig{
			Timeout: "5m",
		},
		Linters: types.LintersConfig{
			Enable: []string{"errcheck", "staticcheck", "govet"},
		},
	}
}
