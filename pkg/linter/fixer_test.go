package linter_test

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fixer", func() {
	var (
		fixer       *linter.Fixer
		analyzer    *linter.Analyzer
		testConfig  string
		logger      *log.Logger
		configTypes types.ConfigLoader
	)

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer = linter.NewAnalyzer(logger)
		configTypes = config.NewLoader(logger)
		fixer = linter.NewFixer(logger, analyzer, configTypes)

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
			Expect(result.IsSuccess()).To(BeTrue())
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
			Expect(result.IsSuccess()).To(BeTrue())
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

	Context("Deprecated Linters", func() {
		It("should detect deprecated linter wsl and suggest wsl_v5", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - wsl
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			// Dry-run should report fixes but not modify
			Expect(result.FixesApplied).To(BeNumerically(">", 0))
		})

		It("should handle config with only deprecated linters", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - wsl
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			Expect(result.FixesApplied).To(BeNumerically(">", 0))
		})

		It("should fix deprecated linters in non-dry-run mode", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - wsl
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			Expect(result.FixesApplied).To(BeNumerically(">", 0))

			// Verify file was modified - wsl should be replaced with wsl_v5
			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("wsl_v5"))
			Expect(string(content)).NotTo(ContainSubstring("wsl:"))
		})
	})

	Context("Invalid Duration Fields", func() {
		It("should fix empty timeout field", func() {
			configContent := `version: "2"
run:
  timeout: ""
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, false)

			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(`timeout: 5m`))
		})

		It("should fix invalid timeout format", func() {
			configContent := `version: "2"
run:
  timeout: invalid
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, false)

			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(`timeout: 5m`))
		})

		It("should keep valid timeout unchanged", func() {
			configContent := `version: "2"
run:
  timeout: 10m
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, true)

			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(`timeout: 10m`))
		})

		It("should NOT fix invalid timeout in dry-run mode (file unchanged)", func() {
			configContent := `version: "2"
run:
  timeout: ""
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityCritical, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			Expect(result.FixesApplied).To(Equal(1))

			// File should NOT be modified in dry-run mode
			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(`timeout: ""`))
		})
	})
})
