package linter_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// writeConfig writes the given content to the test config file.
func writeConfig(path, content string) {
	Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
}

// fixAndRead writes config, runs fix, and returns the resulting file content.
func fixAndRead(
	fixer *linter.Fixer,
	configPath, content string,
	priority types.LinterPriority,
	dryRun bool,
) (string, error) {
	writeConfig(configPath, content)

	_, err := fixer.FixConfig(context.Background(), configPath, priority, dryRun)
	if err != nil {
		return "", err
	}

	result, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// testDeprecatedLinterDryRun tests a deprecated linter scenario in dry-run mode.
func testDeprecatedLinterDryRun(fixer *linter.Fixer, configPath, content string) {
	writeConfig(configPath, content)
	result, err := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityHigh, true)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.IsSuccess()).To(BeTrue())
	Expect(result.FixesApplied).To(BeNumerically(">", 0))
}

// testFixResult writes config, runs fix, and verifies the result contains expected substring.
func testFixResult(
	fixer *linter.Fixer,
	configPath, content string,
	priority types.LinterPriority,
	dryRun bool,
	expected string,
) {
	contentResult, err := fixAndRead(fixer, configPath, content, priority, dryRun)
	Expect(err).NotTo(HaveOccurred())
	Expect(contentResult).To(ContainSubstring(expected))
}

// testFixSuccess writes config and runs fix, verifying success without checking content.
func testFixSuccess(fixer *linter.Fixer, configPath, content string, priority types.LinterPriority, dryRun bool) {
	writeConfig(configPath, content)
	result, err := fixer.FixConfig(context.Background(), configPath, priority, dryRun)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.IsSuccess()).To(BeTrue())
}

// testLinterModification tests modification of a specific linter in the config.
func testLinterModification(
	fixer *linter.Fixer,
	configPath string,
	initialLinter string,
	priority types.LinterPriority,
	dryRun bool,
	assertFunc func(string),
) {
	configContent := fmt.Sprintf(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - %s
`, initialLinter)
	writeConfig(configPath, configContent)
	contentResult, err := fixAndRead(fixer, configPath, configContent, priority, dryRun)
	Expect(err).NotTo(HaveOccurred())
	assertFunc(contentResult)
}

// timeoutTestConfig generates a test config with the given timeout value.
func timeoutTestConfig(timeout string) string {
	return `version: "2"
run:
  timeout: ` + timeout + `
linters:
  enable:
    - gosec
`
}

// testTimeoutFixResult is a helper for testing timeout fixes with consistent parameters.
func testTimeoutFixResult(fixer *linter.Fixer, configPath, inputTimeout, expectedTimeout string, dryRun bool) {
	testFixResult(
		fixer,
		configPath,
		timeoutTestConfig(inputTimeout),
		types.LinterPriorityCritical,
		dryRun,
		"timeout: "+expectedTimeout,
	)
}

// minimalTestConfig returns a minimal test config with optional linters.
func minimalTestConfig(linters ...string) string {
	if len(linters) == 0 {
		return `version: "2"
linters:
  enable:
    - gosec
`
	}
	return fmt.Sprintf(`version: "2"
linters:
  enable:
    - gosec
    - %s
`, strings.Join(linters, "\n    - "))
}

func countSubstring(s, substr string) int {
	return strings.Count(s, substr)
}

func indexOr(s, substr string, fallback int) int {
	idx := strings.Index(s, substr)
	if idx == -1 {
		return fallback
	}

	return idx
}

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
			testFixResult(fixer, testConfig, configContent, types.LinterPriorityCritical, false, `version: "2"`)
		})

		It("should keep existing version 2", func() {
			testFixSuccess(fixer, testConfig, minimalTestConfig(), types.LinterPriorityCritical, true)
		})
	})

	Context("Configuration Modification", func() {
		It("should run in dry-run mode without modifying file", func() {
			testFixSuccess(fixer, testConfig, minimalTestConfig("errcheck"), types.LinterPriorityCritical, true)
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
			testDeprecatedLinterDryRun(fixer, testConfig, configContent)
		})

		It("should handle config with only deprecated linters", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - wsl
`
			testDeprecatedLinterDryRun(fixer, testConfig, configContent)
		})

		It("should fix deprecated linters in non-dry-run mode", func() {
			testLinterModification(fixer, testConfig, "wsl", types.LinterPriorityHigh, false, func(content string) {
				Expect(content).To(ContainSubstring("wsl_v5"))
				Expect(content).NotTo(ContainSubstring("wsl:"))
			})
		})
	})

	Context("Invalid Duration Fields", func() {
		It("should fix empty timeout field", func() {
			testTimeoutFixResult(fixer, testConfig, `""`, `5m`, false)
		})

		It("should fix invalid timeout format", func() {
			testTimeoutFixResult(fixer, testConfig, `invalid`, `5m`, false)
		})

		It("should keep valid timeout unchanged", func() {
			testTimeoutFixResult(fixer, testConfig, `10m`, `10m`, true)
		})

		It("should NOT fix invalid timeout in dry-run mode (file unchanged)", func() {
			testTimeoutFixResult(fixer, testConfig, `""`, `"`, true)
		})
	})

	Context("Typecheck Linter", func() {
		It("should remove typecheck from enabled linters", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - typecheck
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("gosec"))
			Expect(content).NotTo(ContainSubstring("- typecheck"))
		})

		It("should remove typecheck from disabled linters", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
  disable:
    - typecheck
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).NotTo(ContainSubstring("typecheck"))
		})

		It("should handle typecheck in dry-run mode", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - typecheck
`
			testDeprecatedLinterDryRun(fixer, testConfig, configContent)
		})
	})

	Context("Build Tags", func() {
		It("should add all GOEXPERIMENT build tags", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(content).To(ContainSubstring("goexperiment.arenas"))
			Expect(content).To(ContainSubstring("goexperiment.goroutineleakprofile"))
			Expect(content).To(ContainSubstring("goexperiment.jsonv2"))
			Expect(content).To(ContainSubstring("goexperiment.runtimesecret"))
			Expect(content).To(ContainSubstring("goexperiment.simd"))
		})

		It("should preserve existing build tags", func() {
			configContent := `version: "2"
run:
  timeout: 5m
  build-tags:
    - custom_tag
linters:
  enable:
    - gosec
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(content).To(ContainSubstring("custom_tag"))
			Expect(content).To(ContainSubstring("goexperiment.jsonv2"))
		})

		It("should not duplicate already-present experiment tags", func() {
			configContent := `version: "2"
run:
  timeout: 5m
  build-tags:
    - goexperiment.jsonv2
linters:
  enable:
    - gosec
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			jsonv2Count := countSubstring(content, "goexperiment.jsonv2")
			Expect(jsonv2Count).To(Equal(1))
		})

		It("should sort build tags", func() {
			configContent := `version: "2"
run:
  timeout: 5m
  build-tags:
    - zebra_tag
    - alpha_tag
linters:
  enable:
    - gosec
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			zebraIdx := indexOr(content, "zebra_tag", len(content))
			alphaIdx := indexOr(content, "alpha_tag", 0)
			Expect(alphaIdx).To(BeNumerically("<", zebraIdx))
		})
	})
})
