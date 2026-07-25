package linter_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
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

// fixHighPriorityAndContain writes config, runs fix at high priority (non-dry-run),
// asserts no error, and verifies the result contains all expected substrings.
func fixHighPriorityAndContain(
	fixer *linter.Fixer,
	configPath, content string,
	expected ...string,
) {
	fixHighPriority(fixer, configPath, content, func(result string) {
		for _, e := range expected {
			Expect(result).To(ContainSubstring(e))
		}
	})
}

// fixHighPriorityContainAndNotContain is like fixHighPriorityAndContain but also
// verifies the result does NOT contain the specified substrings.
func fixHighPriorityContainAndNotContain(
	fixer *linter.Fixer,
	configPath, content string,
	contain []string,
	notContain []string,
) {
	fixHighPriority(fixer, configPath, content, func(result string) {
		for _, e := range contain {
			Expect(result).To(ContainSubstring(e))
		}

		for _, e := range notContain {
			Expect(result).NotTo(ContainSubstring(e))
		}
	})
}

// fixHighPriority writes config, runs fix at high priority (non-dry-run), asserts no error,
// then calls assertFunc with the result.
func fixHighPriority(
	fixer *linter.Fixer,
	configPath, content string,
	assertFunc func(string),
) {
	fixAndAssert(fixer, configPath, content, types.LinterPriorityHigh, false, assertFunc)
}

// fixAndAssert writes config, runs fix, asserts no error, then calls assertFunc with the result.
func fixAndAssert(
	fixer *linter.Fixer,
	configPath, content string,
	priority types.LinterPriority,
	dryRun bool,
	assertFunc func(string),
) {
	contentResult, err := fixAndRead(fixer, configPath, content, priority, dryRun)
	Expect(err).NotTo(HaveOccurred())
	assertFunc(contentResult)
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

// fullyPreparedConfig builds a "fully prepared" golangci-lint config with all
// default exclusion paths, rules, runner settings, and build tags pre-populated.
// It is the shared scaffolding for tests that verify what happens when only one
// specific piece of the config (e.g. issues limits, default linter settings)
// is the focus of the test.
//
// extraLinters is appended to the always-present `gosec` enable entry.
// issuesBlock is the full literal YAML for the `issues:` block (key + value),
// for example `issues: {}` or
// `issues:\n  max-issues-per-linter: 50\n  max-same-issues: 10`.
// Use `emptyIssuesBlock` for an empty object and `defaultIssuesBlock` for the
// fixer-injected defaults.
func fullyPreparedConfig(issuesBlock string, extraLinters ...string) string {
	linterList := strings.Join(append([]string{"gosec"}, extraLinters...), "\n    - ")

	return fmt.Sprintf(`version: "2"
run:
  timeout: 5m
  go: "1.26.0"
  build-tags:
    - goexperiment.arenas
    - goexperiment.goroutineleakprofile
    - goexperiment.jsonv2
    - goexperiment.runtimesecret
    - goexperiment.simd
  allow-parallel-runners: true
  allow-serial-runners: true
linters:
  enable:
    - %s
  exclusions:
    generated: lax
    rules:
      - path: _test\.go
        linters:
          - exhaustruct
          - testpackage
          - gochecknoglobals
          - funlen
          - cyclop
          - goconst
          - forcetypeassert
          - gosec
          - errcheck
          - wrapcheck
          - ireturn
          - recvcheck
          - contextcheck
          - exhaustive
      - path: _test\.go
        text: unused
        linters:
          - unused
    paths:
      - _templ\.go$
      - \.gen\.go$
      - vendor/
%s
`, linterList, issuesBlock)
}

// emptyIssuesBlock is an empty `issues:` object — used to test the fixer
// injects default issue limits when none are present.
const emptyIssuesBlock = `issues: {}`

// defaultIssuesBlock is the issues block produced by the fixer when
// `issues: {}` or no issues block is present.
const defaultIssuesBlock = `issues:
  max-issues-per-linter: 50
  max-same-issues: 10`

// recordedAction captures one ledger Record call for test assertions.
type recordedAction struct {
	action audit.Action
	linter string
	reason string
}

// captureRecorder is a test fake for audit.Recorder that collects all Record calls.
type captureRecorder struct {
	actions []recordedAction
}

func (c *captureRecorder) Record(action audit.Action, linter, reason string) {
	c.actions = append(c.actions, recordedAction{action: action, linter: linter, reason: reason})
}

func (c *captureRecorder) hasAction(action audit.Action, linter string) bool {
	for _, a := range c.actions {
		if a.action == action && a.linter == linter {
			return true
		}
	}

	return false
}

func (c *captureRecorder) countByAction(action audit.Action) int {
	count := 0

	for _, a := range c.actions {
		if a.action == action {
			count++
		}
	}

	return count
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

		DescribeTable(
			"should fix deprecated linters in non-dry-run mode",
			func(linterName, replacement, absentSubstr string) {
				testLinterModification(
					fixer, testConfig, linterName,
					types.LinterPriorityHigh, false,
					func(content string) {
						Expect(content).To(ContainSubstring(replacement))
						Expect(content).NotTo(ContainSubstring(absentSubstr))
					},
				)
			},
			Entry("wsl -> wsl_v5", "wsl", "wsl_v5", "wsl:"),
			Entry("gomodguard -> gomodguard_v2", "gomodguard", "gomodguard_v2", "- gomodguard\n"),
		)

		It("should detect deprecated linter gomodguard and suggest gomodguard_v2", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - gomodguard
`
			testDeprecatedLinterDryRun(fixer, testConfig, configContent)
		})

		It("should migrate gomodguard settings to gomodguard_v2", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - gomodguard
  settings:
    gomodguard:
      allowed:
        modules:
          - golang.org/x/mod
`
			writeConfig(testConfig, configContent)
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("gomodguard_v2"))
			Expect(content).To(ContainSubstring("golang.org/x/mod"))
			Expect(content).NotTo(ContainSubstring("- gomodguard\n"))
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
			fixHighPriorityContainAndNotContain(fixer, testConfig, `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - typecheck
`, []string{"gosec"}, []string{"- typecheck"})
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

	Context("Disabled Linters", func() {
		It("should move noinlineerr from enable to disable", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - noinlineerr
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Linters.Enable).NotTo(ContainElement("noinlineerr"))
			Expect(parsed.Linters.Disable).To(ContainElement("noinlineerr"))
		})

		It("should move depguard from enable to disable", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - depguard
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Linters.Enable).NotTo(ContainElement("depguard"))
			Expect(parsed.Linters.Disable).To(ContainElement("depguard"))
		})
	})

	Context("User-Disabled Linters", func() {
		It("should preserve the user's disable list across a configure run", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
  disable:
    - mnd
    - varnamelen`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Linters.Disable).To(ContainElement("mnd"))
			Expect(parsed.Linters.Disable).To(ContainElement("varnamelen"))
		})

		It("should not re-add a recommended linter that is in the disable list", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
  disable:
    - ireturn`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Linters.Enable).NotTo(ContainElement("ireturn"))
			Expect(parsed.Linters.Disable).To(ContainElement("ireturn"))
		})

		It("should preserve the disable list across repeated runs (idempotency)", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
  disable:
    - mnd
    - tagalign
    - varnamelen`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			afterFirst, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			_, err = fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			afterSecond, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(afterSecond.Linters.Disable).To(ContainElements("mnd", "tagalign", "varnamelen"))
			Expect(afterSecond.Linters.Enable).NotTo(ContainElement("mnd"))
			Expect(afterSecond.Linters.Disable).To(Equal(afterFirst.Linters.Disable))
		})

		It("should prune orphaned settings blocks for disabled linters", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
  disable:
    - mnd
  settings:
    mnd:
      checks:
        - argument
        - case
    gosec:
      excludes:
        - G104`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Linters.Settings).NotTo(HaveKey("mnd"))
			Expect(parsed.Linters.Settings).To(HaveKey("gosec"))
		})

		It("should resolve a linter present in both enable and disable", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - mnd
  disable:
    - mnd`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := configTypes.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Linters.Enable).To(ContainElement("mnd"))
			Expect(parsed.Linters.Disable).NotTo(ContainElement("mnd"))
		})
	})

	Context("Audit Ledger Recording", func() {
		It("records added-to-enable entries when recommenders add linters", func() {
			recorder := &captureRecorder{}
			fixer.SetLedger(recorder)

			configContent := `version: "2"
linters:
  enable:
    - gosec
`
			writeConfig(testConfig, configContent)
			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())

			Expect(recorder.countByAction(audit.ActionAddedToEnable)).To(BeNumerically(">", 0))
		})

		It("records moved-to-disable and removed-from-enable when a tool-disabled linter is enabled", func() {
			recorder := &captureRecorder{}
			fixer.SetLedger(recorder)

			configContent := `version: "2"
linters:
  enable:
    - gosec
    - noinlineerr
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(recorder.hasAction(audit.ActionMovedToDisable, "noinlineerr")).To(BeTrue())
			Expect(recorder.hasAction(audit.ActionRemovedFromEnable, "noinlineerr")).To(BeTrue())
		})

		It("records pruned-settings when an orphaned settings block is removed", func() {
			recorder := &captureRecorder{}
			fixer.SetLedger(recorder)

			configContent := `version: "2"
linters:
  enable:
    - gosec
  disable:
    - mnd
  settings:
    mnd:
      checks:
        - argument
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(recorder.hasAction(audit.ActionPrunedSettings, "mnd")).To(BeTrue())
		})

		It("does not record anything in dry-run mode", func() {
			recorder := &captureRecorder{}
			fixer.SetLedger(recorder)

			configContent := `version: "2"
linters:
  enable:
    - gosec
    - noinlineerr
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(recorder.actions).To(BeEmpty())
		})
	})

	Context("Default Linter Settings", func() {
		It("should inject ireturn defaults when ireturn is enabled without settings", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - ireturn
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("ireturn:"))
			Expect(content).To(ContainSubstring("generic"))
		})

		It("should not overwrite existing linter settings", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - ireturn
  settings:
    ireturn:
      accept:
        - error
        - empty
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("ireturn:"))
			Expect(content).NotTo(ContainSubstring("generic"))
		})

		It("should inject defaults for multiple linters simultaneously", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - ireturn
    - makezero
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("ireturn:"))
			Expect(content).To(ContainSubstring("generic"))
			Expect(content).To(ContainSubstring("makezero:"))
			Expect(content).To(ContainSubstring("always: true"))
		})

		It("should inject revive defaults when revive is enabled without settings", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
    - revive
`, "revive:", "exported", "package-comments")
		})

		It("should inject varnamelen defaults when varnamelen is enabled", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - varnamelen
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("varnamelen:"))
			Expect(content).To(ContainSubstring("ignore-map-index-ok"))
			Expect(content).To(ContainSubstring("ignore-type-assert-ok"))
		})

		It("should inject gomoddirectives defaults when enabled", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
    - gomoddirectives
`, "gomoddirectives:", "replace-local")
		})

		It("should inject cyclop defaults when cyclop is enabled", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
    - cyclop
`, "cyclop:", "max-complexity")
		})

		It("should inject makezero defaults when makezero is enabled", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - makezero
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("makezero:"))
			Expect(content).To(ContainSubstring("always: true"))
		})

		It("should not overwrite existing makezero settings", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - makezero
  settings:
    makezero:
      always: false
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("makezero:"))
			Expect(content).To(ContainSubstring("always: false"))
			Expect(content).NotTo(ContainSubstring("always: true"))
		})

		It("should inject exhaustruct stdlib excludes when exhaustruct is enabled", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - exhaustruct
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("exhaustruct:"))
			Expect(content).To(ContainSubstring("net/http.Client"))
			Expect(content).To(ContainSubstring("net/http.Server"))
			Expect(content).To(ContainSubstring("os/exec.Cmd"))
		})

		It("should inject gosec excludes when gosec is enabled without settings", func() {
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - gosec
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("gosec:"))
			Expect(content).To(ContainSubstring("G304"))
			Expect(content).To(ContainSubstring("G104"))
		})

		It("should inject errcheck exclude-functions when errcheck is enabled without settings", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - errcheck
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("errcheck:"))
			Expect(content).To(ContainSubstring("exclude-functions"))
			Expect(content).To(ContainSubstring("(*os.File).Close"))
			Expect(content).To(ContainSubstring("fmt.Fprintf"))
		})
	})

	Context("Default Formatter Settings", func() {
		It("should inject golines max-len when golines formatter is enabled", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
    - lll
`, "golines:", "max-len")
		})

		It("should not overwrite existing golines settings", func() {
			fixHighPriorityContainAndNotContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
    - lll
formatters:
  settings:
    golines:
      max-len: 100
`, []string{"max-len: 100"}, []string{"max-len: 120"})
		})
	})

	Context("Default Exclusion Rules", func() {
		It("should add test file exclusion rules", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("_test\\.go"))
			Expect(content).To(ContainSubstring("exhaustruct"))
			Expect(content).To(ContainSubstring("funlen"))
			Expect(content).To(ContainSubstring("cyclop"))
		})

		It("should not duplicate existing test exclusion rules", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
  exclusions:
    rules:
      - path: _test\.go
        linters:
          - exhaustruct
          - testpackage
          - gochecknoglobals
          - funlen
          - cyclop
          - goconst
      - path: _test\.go
        text: unused
        linters:
          - unused
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			testRuleCount := countSubstring(content, "_test\\.go")
			Expect(testRuleCount).To(Equal(2))
		})

		It("should add .gen.go to default exclusion paths", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
`, `\.gen\.go$`)
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
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
run:
  timeout: 5m
  build-tags:
    - custom_tag
linters:
  enable:
    - gosec
`, "custom_tag", "goexperiment.jsonv2")
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

	Context("Default Exclusion Paths", func() {
		It("should add _templ.go$ and vendor/ to linters exclusions", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
`, "_templ\\.go$", "vendor/")
		})

		It("should add _templ.go$ to formatters exclusions", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
`, "vendor/")
		})

		It("should not duplicate already-present default exclusion paths", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
  exclusions:
    paths:
      - _templ\.go$
      - vendor/
formatters:
  exclusions:
    paths:
      - _templ\.go$
`
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			templCount := countSubstring(content, "_templ\\.go$")
			Expect(templCount).To(Equal(2)) // once in linters, once in formatters

			vendorCount := countSubstring(content, "vendor/")
			Expect(vendorCount).To(Equal(1)) // only in linters
		})

		It("should preserve existing exclusion paths while adding defaults", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
  exclusions:
    paths:
      - custom_path/
`, "custom_path/", "_templ\\.go$", "vendor/")
		})
	})

	Context("Issues Block Normalization", func() {
		It("should add max-issues-per-linter and max-same-issues when issues block is empty", func() {
			content, err := fixAndRead(
				fixer, testConfig,
				fullyPreparedConfig(emptyIssuesBlock),
				types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("max-issues-per-linter:"))
			Expect(content).To(ContainSubstring("max-same-issues:"))
		})

		It("should not overwrite existing issue limits", func() {
			content, err := fixAndRead(
				fixer, testConfig,
				fullyPreparedConfig(`issues:
  max-issues-per-linter: 200
  max-same-issues: 50`),
				types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("max-issues-per-linter: 200"))
			Expect(content).To(ContainSubstring("max-same-issues: 50"))
		})
	})

	Context("Settings Injection Without Other Fixes", func() {
		// These tests verify the P0 silent-drop fix: when a config already has
		// all exclusion paths, rules, runner settings, and build tags, but is
		// missing default settings for an enabled linter, the fixer must still
		// inject and save those settings.
		DescribeTable(
			"should inject <linter> default settings even when no other fixes are needed",
			func(linter, expectedKey, expectedSetting string) {
				content, err := fixAndRead(
					fixer, testConfig,
					fullyPreparedConfig(defaultIssuesBlock, linter),
					types.LinterPriorityHigh, false,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(ContainSubstring(expectedKey))
				Expect(content).To(ContainSubstring(expectedSetting))
			},
			Entry("ginkgolinter", "ginkgolinter", "ginkgolinter:", "forbid-focus-container"),
			Entry("testifylint", "testifylint", "testifylint:", "enable-all"),
		)
	})

	Context("Dry-Run Accuracy", func() {
		It("should count config-level fixes in dry-run mode", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - ginkgolinter
`
			writeConfig(testConfig, configContent)
			result, err := fixer.FixConfig(
				context.Background(), testConfig, types.LinterPriorityHigh, true,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			Expect(result.FixesApplied).To(BeNumerically(">", 0))
		})

		It("should not modify the file in dry-run mode", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(
				context.Background(), testConfig, types.LinterPriorityHigh, true,
			)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("version: \"2\""))
			Expect(string(content)).NotTo(ContainSubstring("max-issues-per-linter"))
			Expect(string(content)).NotTo(ContainSubstring("goexperiment"))
		})
	})

	Context("Issues Exit Code Normalization", func() {
		DescribeTable(
			"should normalize issues-exit-code",
			func(configContent, expectedSubstring string) {
				content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(ContainSubstring(expectedSubstring))
			},
			Entry(
				"set to 1 when missing/zero",
				`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`,
				"issues-exit-code: 1",
			),
			Entry(
				"preserve existing non-zero value",
				`version: "2"
run:
  timeout: 5m
  issues-exit-code: 2
linters:
  enable:
    - gosec
`,
				"issues-exit-code: 2",
			),
		)
	})

	Context("Output Formats Normalization", func() {
		It("should set output.formats when missing", func() {
			fixHighPriorityAndContain(fixer, testConfig, `version: "2"
linters:
  enable:
    - gosec
`, "output:", "formats:")
		})
	})

	Context("Idempotency", func() {
		It("should report 0 fixes on second run", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
`
			// First run — applies all fixes
			content, err := fixAndRead(fixer, testConfig, configContent, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			// Second run — should be a no-op
			writeConfig(testConfig, content)
			result, err := fixer.FixConfig(
				context.Background(), testConfig, types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.FixesApplied).To(Equal(0))
		})
	})
})
