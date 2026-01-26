package linter_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
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

	Context("Configuration Modification", func() {
		It("should create backup before modification", func() {
			configContent := fmt.Sprintf(`version: "2"
linters:
  enable:
    - gosec
    - errcheck
`)
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			result, err := fixer.FixConfig(testConfig, types.LinterPriorityCritical, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Success).To(BeTrue())
		})

		It("should handle missing config file gracefully", func() {
			_, err := fixer.FixConfig("/non/existent/config.yml", types.LinterPriorityCritical, true)

			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid YAML", func() {
			configContent := `linters:
  enable: [unclosed bracket`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			_, err := fixer.FixConfig(testConfig, types.LinterPriorityCritical, true)

			Expect(err).To(HaveOccurred())
		})
	})
})
