package types_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func TestConfigConstructor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Constructor Suite")
}

var _ = Describe("NewConfig", func() {
	It("sets Version to ConfigVersionV2", func() {
		cfg := types.NewConfig()
		Expect(cfg.Version).To(Equal(types.ConfigVersionV2))
	})

	It("sets Run.Timeout to 5m", func() {
		cfg := types.NewConfig()
		Expect(cfg.Run.Timeout).To(Equal("5m"))
	})

	It("sets Run.Tests to true", func() {
		cfg := types.NewConfig()
		Expect(cfg.Run.Tests).To(BeTrue())
	})

	It("sets Run.IssuesExitCode to 1", func() {
		cfg := types.NewConfig()
		Expect(cfg.Run.IssuesExitCode).To(Equal(1))
	})

	It("initializes Output.Formats to empty map", func() {
		cfg := types.NewConfig()
		Expect(cfg.Output.Formats).ToNot(BeNil())
		Expect(cfg.Output.Formats).To(BeEmpty())
	})

	It("sets Issues defaults", func() {
		cfg := types.NewConfig()
		Expect(cfg.Issues.MaxIssuesPerLinter).To(Equal(50))
		Expect(cfg.Issues.MaxSameIssues).To(Equal(10))
	})

	It("sets exclusions generated to lax", func() {
		cfg := types.NewConfig()
		Expect(cfg.Linters.Exclusions.Generated).To(Equal("lax"))
		Expect(cfg.Formatters.Exclusions.Generated).To(Equal("lax"))
	})

	It("passes ValidateConfig", func() {
		cfg := types.NewConfig()
		Expect(types.ValidateConfig(cfg)).To(Succeed())
	})

	Context("with options", func() {
		It("applies WithTimeout", func() {
			cfg := types.NewConfig(types.WithTimeout("10m"))
			Expect(cfg.Run.Timeout).To(Equal("10m"))
		})

		It("applies WithGoVersion", func() {
			cfg := types.NewConfig(types.WithGoVersion("1.26"))
			Expect(cfg.Run.Go).To(Equal("1.26"))
		})

		It("applies WithLinters", func() {
			cfg := types.NewConfig(types.WithLinters([]string{"gosec", "errcheck"}))
			Expect(cfg.Linters.Enable).To(Equal([]string{"gosec", "errcheck"}))
		})

		It("applies WithFormatters", func() {
			cfg := types.NewConfig(types.WithFormatters([]string{"gci", "gofumpt"}))
			Expect(cfg.Formatters.Enable).To(Equal([]string{"gci", "gofumpt"}))
		})

		It("applies multiple options in order", func() {
			cfg := types.NewConfig(
				types.WithTimeout("10m"),
				types.WithGoVersion("1.25"),
				types.WithLinters([]string{"gosec"}),
			)
			Expect(cfg.Run.Timeout).To(Equal("10m"))
			Expect(cfg.Run.Go).To(Equal("1.25"))
			Expect(cfg.Linters.Enable).To(Equal([]string{"gosec"}))
		})

		It("later options override earlier ones", func() {
			cfg := types.NewConfig(
				types.WithTimeout("10m"),
				types.WithTimeout("3m"),
			)
			Expect(cfg.Run.Timeout).To(Equal("3m"))
		})
	})
})
