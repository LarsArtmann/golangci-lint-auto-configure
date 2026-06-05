package types_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config.Clone", func() {
	It("should return nil for nil receiver", func() {
		var cfg *types.Config
		Expect(cfg.Clone()).To(BeNil())
	})

	It("should produce an independent deep copy", func() {
		original := &types.Config{
			Version: "2",
			Run: types.RunConfig{
				Timeout:   "5m",
				BuildTags: []string{"tag1", "tag2"},
			},
			Output: types.OutputConfig{
				Formats:   map[string]any{"json": map[string]any{"path": "stdout"}},
				SortOrder: []string{"linter", "severity"},
			},
			Linters: types.LintersConfig{
				Enable:  []string{"gosec", "errcheck"},
				Disable: []string{"typecheck"},
				Settings: map[string]any{
					"funlen": map[string]any{"lines": 80},
				},
				Exclusions: types.LintersExclusionsConfig{
					Generated: "lax",
					Paths:     []string{"vendor"},
					Rules: []types.ExclusionRuleConfig{
						{Path: "gen.go", Text: "SA1000"},
					},
				},
			},
			Formatters: types.FormattersConfig{
				Enable: []string{"gofmt"},
				Exclusions: types.FormattersExclusionsConfig{
					Paths: []string{"generated"},
				},
			},
			Issues: types.IssuesConfig{
				MaxIssuesPerLinter: 50,
			},
		}

		cloned := original.Clone()

		Expect(cloned).ToNot(BeIdenticalTo(original))
		Expect(cloned.Version).To(Equal("2"))
		Expect(cloned.Run.Timeout).To(Equal("5m"))

		Expect(cloned.Run.BuildTags).To(Equal(original.Run.BuildTags))
		cloned.Run.BuildTags[0] = "modified"
		Expect(original.Run.BuildTags[0]).To(Equal("tag1"))

		Expect(cloned.Linters.Enable).To(Equal(original.Linters.Enable))
		cloned.Linters.Enable[0] = "modified"
		Expect(original.Linters.Enable[0]).To(Equal("gosec"))

		Expect(cloned.Linters.Settings).To(Equal(original.Linters.Settings))
		cloned.Linters.Settings["funlen"] = map[string]any{"lines": 200}
		funlenOriginal, ok := original.Linters.Settings["funlen"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(funlenOriginal["lines"]).To(Equal(80))

		Expect(cloned.Linters.Exclusions.Paths).To(Equal(original.Linters.Exclusions.Paths))
		cloned.Linters.Exclusions.Paths[0] = "modified"
		Expect(original.Linters.Exclusions.Paths[0]).To(Equal("vendor"))

		Expect(cloned.Output.Formats).To(Equal(original.Output.Formats))
		cloned.Output.Formats["json"] = "modified"
		Expect(original.Output.Formats["json"]).ToNot(Equal("modified"))
	})
})
