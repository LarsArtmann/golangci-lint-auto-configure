package linter_test

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fixer", func() {
	var (
		fixer      *linter.Fixer
		analyzer   *linter.Analyzer
		testConfig string
		logger     *log.Logger
	)

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer = linter.NewAnalyzer(logger)
		fixer = linter.NewFixer(logger, analyzer)

		testDir := GinkgoT().TempDir()
		testConfig = filepath.Join(testDir, ".golangci.yml")
	})

	Context("Version Field", func() {
		It("should fix empty version field", func() {
			configContent := `version: ""
run:
  timeout: 10m
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, false)

			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(`version: "2"`))
		})

		It("should keep existing version 2", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Success).To(BeTrue())
		})
	})

	Context("Configuration Modification", func() {
		It("should run in dry-run mode without modifying file", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - errcheck
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Success).To(BeTrue())
		})

		It("should handle missing config file gracefully", func() {
			_, err := fixer.FixConfig(
				context.Background(),
				"/non/existent/config.yml",
				types.LinterPriorityCritical,
				true,
			)

			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid YAML", func() {
			configContent := `linters:
  enable: [unclosed bracket`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, true)

			Expect(err).To(HaveOccurred())
		})
	})
})
