package linter_test

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// newSpecFixer builds a fixer wired to a quiet logger, the real analyzer,
// and the real config loader.
func newSpecFixer() *linter.Fixer {
	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})

	return linter.NewFixer(logger, linter.NewAnalyzer(logger), config.NewLoader(logger))
}

// specConfigPath returns a unique config path inside the spec's temp dir.
func specConfigPath() string {
	return filepath.Join(GinkgoT().TempDir(), ".golangci.yml")
}

// configLoaderForSpecs re-parses a saved config for structural assertions.
func configLoaderForSpecs() types.ConfigLoader {
	logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})

	return config.NewLoader(logger)
}

// fixExclusionsAndParse runs a non-dry-run fix and returns the re-parsed config.
func fixExclusionsAndParse(fixer *linter.Fixer, configPath, content string) *types.Config {
	writeConfig(configPath, content)

	_, err := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityMedium, false)
	Expect(err).NotTo(HaveOccurred())

	return loadParsedConfig(configPath)
}

func loadParsedConfig(configPath string) *types.Config {
	configTypes := configLoaderForSpecs()

	parsed, err := configTypes.LoadConfig(configPath)
	Expect(err).NotTo(HaveOccurred())

	return parsed
}

// exclusionLinterNames flattens the linter lists of all exclusion rules.
func exclusionLinterNames(cfg *types.Config) [][]string {
	names := make([][]string, 0, len(cfg.Linters.Exclusions.Rules))

	for _, rule := range cfg.Linters.Exclusions.Rules {
		names = append(names, rule.Linters)
	}

	return names
}

var _ = Describe("Deprecated linter exclusion cleanup", func() {
	var (
		fixer      *linter.Fixer
		testConfig string
	)

	BeforeEach(func() {
		fixer = newSpecFixer()
		testConfig = specConfigPath()
	})

	Context("when a config already migrated its enable list", func() {
		It("renames the dead predecessor inside exclusion rules", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - exhaustruct_v5
  exclusions:
    rules:
      - linters:
          - exhaustruct
        text: "is missing fields"
`
			parsed := fixExclusionsAndParse(fixer, testConfig, configContent)

			Expect(exclusionLinterNames(parsed)).To(ContainElement([]string{"exhaustruct_v5"}))

			for _, names := range exclusionLinterNames(parsed) {
				Expect(names).NotTo(ContainElement("exhaustruct"))
			}
		})

		It("deduplicates when a rule lists both the predecessor and the successor", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - exhaustruct_v5
  exclusions:
    rules:
      - linters:
          - exhaustruct
          - exhaustruct_v5
        text: "is missing fields"
`
			parsed := fixExclusionsAndParse(fixer, testConfig, configContent)

			Expect(exclusionLinterNames(parsed)).To(ContainElement([]string{"exhaustruct_v5"}))
		})

		It("records an exclusion-linter-renamed audit entry", func() {
			recorder := &captureRecorder{}
			fixer.SetLedger(recorder)

			configContent := `version: "2"
linters:
  default: none
  enable:
    - exhaustruct_v5
  exclusions:
    rules:
      - linters:
          - exhaustruct
        text: "is missing fields"
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(recorder.hasAction(audit.ActionExclusionLinterRenamed, "exhaustruct")).To(BeTrue())
		})
	})

	Context("when a linter was removed from golangci-lint without a replacement", func() {
		It("drops text-scoped rules left without any linter", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - gosec
  exclusions:
    rules:
      - linters:
          - ifshort
        text: "unnecessary if"
`
			parsed := fixExclusionsAndParse(fixer, testConfig, configContent)

			for _, rule := range parsed.Linters.Exclusions.Rules {
				Expect(rule.Linters).NotTo(ContainElement("ifshort"))
				Expect(rule.Text).NotTo(Equal("unnecessary if"))
			}
		})

		It("keeps path-scoped rules and clears the removed linter from them", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - gosec
  exclusions:
    rules:
      - linters:
          - ifshort
        path: _test\.go
`
			parsed := fixExclusionsAndParse(fixer, testConfig, configContent)

			pathOnlyRuleSurvived := false

			for _, rule := range parsed.Linters.Exclusions.Rules {
				Expect(rule.Linters).NotTo(ContainElement("ifshort"))

				if rule.Path == `_test\.go` && rule.Text == "" && rule.Source == "" {
					pathOnlyRuleSurvived = true
				}
			}

			Expect(pathOnlyRuleSurvived).To(BeTrue(),
				"the user's path-only rule must survive the removed-linter cleanup")
		})

		It("keeps only the surviving linters of a mixed rule", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - gosec
  exclusions:
    rules:
      - linters:
          - ifshort
          - gosec
        text: "G402"
`
			parsed := fixExclusionsAndParse(fixer, testConfig, configContent)

			Expect(exclusionLinterNames(parsed)).To(ContainElement([]string{"gosec"}))
		})

		It("records an exclusion-linter-removed audit entry", func() {
			recorder := &captureRecorder{}
			fixer.SetLedger(recorder)

			configContent := `version: "2"
linters:
  default: none
  enable:
    - gosec
  exclusions:
    rules:
      - linters:
          - ifshort
        text: "unnecessary if"
`
			writeConfig(testConfig, configContent)
			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(recorder.hasAction(audit.ActionExclusionLinterRemoved, "ifshort")).To(BeTrue())
		})
	})

	Context("when nothing needs cleaning", func() {
		It("leaves healthy exclusion rules untouched and reports no fixes on a second run", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - exhaustruct_v5
  exclusions:
    rules:
      - linters:
          - exhaustruct_v5
        text: "is missing fields"
`
			writeConfig(testConfig, configContent)

			_, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())

			parsed := loadParsedConfig(testConfig)
			Expect(exclusionLinterNames(parsed)).To(ContainElement([]string{"exhaustruct_v5"}))

			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.FixesApplied).To(BeZero())

			second := loadParsedConfig(testConfig)
			Expect(second.Linters.Exclusions.Rules).To(HaveLen(len(parsed.Linters.Exclusions.Rules)))
		})
	})

	Context("in dry-run mode", func() {
		It("reports the cleanup without touching the file", func() {
			configContent := `version: "2"
linters:
  default: none
  enable:
    - exhaustruct_v5
  exclusions:
    rules:
      - linters:
          - exhaustruct
        text: "is missing fields"
`
			writeConfig(testConfig, configContent)
			result, err := fixer.FixConfig(context.Background(), testConfig, types.LinterPriorityMedium, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.FixesApplied).To(BeNumerically(">", 0))

			parsed := loadParsedConfig(testConfig)
			Expect(exclusionLinterNames(parsed)).To(ContainElement([]string{"exhaustruct"}))
		})
	})
})
