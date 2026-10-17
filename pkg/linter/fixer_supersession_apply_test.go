package linter_test

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v3"
)

// These specs pin detect/apply parity for Converge-only normalizer repairs:
// the redundancy pass (lll dies with golines) runs ONLY inside Converge, so
// apply must apply what the report-only path plans. The overwrite it replaces
// discarded those repairs, reporting "already converged" while detect still
// flagged the drift — a permanent finding that no repair could clear
// (2026-10-17 BuildFlow incident).
var _ = Describe("FixConfig supersession apply parity", func() {
	var (
		logger     *log.Logger
		fixer      *linter.Fixer
		testDir    string
		configPath string
	)

	const lllEnabled = `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - lll
formatters:
  enable:
    - goimports
    - gofumpt
    - golines
`

	enableList := func() []string {
		fixed, readErr := os.ReadFile(configPath)
		Expect(readErr).NotTo(HaveOccurred())

		var parsed struct {
			Linters struct {
				Enable []string `yaml:"enable"`
			} `yaml:"linters"`
		}

		Expect(yaml.Unmarshal(fixed, &parsed)).To(Succeed())

		return parsed.Linters.Enable
	}

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer := linter.NewAnalyzer(logger)
		configLoader := config.NewLoader(logger)
		fixer = linter.NewFixer(logger, analyzer, configLoader)

		testDir = GinkgoT().TempDir()
		configPath = filepath.Join(testDir, ".golangci.yml")
		Expect(os.WriteFile(configPath, []byte(lllEnabled), 0o644)).To(Succeed())
	})

	It("dry run and report-only plan agree on the supersession removal", func() {
		planned, planErr := fixer.PlanDeviations(context.Background(), configPath, types.LinterPriorityMedium)
		Expect(planErr).NotTo(HaveOccurred())
		Expect(planned).NotTo(BeEmpty(), "detect must plan the lll supersession removal")

		result, fixErr := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityMedium, true)
		Expect(fixErr).NotTo(HaveOccurred())
		Expect(result.FixesApplied).To(Equal(len(planned)),
			"dry run must report exactly the changes the report-only plan carries")
	})

	It("applies the supersession removal the report-only path planned", func() {
		planned, planErr := fixer.PlanDeviations(context.Background(), configPath, types.LinterPriorityMedium)
		Expect(planErr).NotTo(HaveOccurred())
		Expect(planned).NotTo(BeEmpty())

		result, fixErr := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityMedium, false)
		Expect(fixErr).NotTo(HaveOccurred())
		Expect(result.IsSuccess()).To(BeTrue())
		Expect(result.FixesApplied).To(BeNumerically(">=", 1))
		Expect(enableList()).NotTo(ContainElement("lll"))
	})

	It("converges in one round: detect is quiet after apply", func() {
		_, fixErr := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityMedium, false)
		Expect(fixErr).NotTo(HaveOccurred())

		planned, planErr := fixer.PlanDeviations(context.Background(), configPath, types.LinterPriorityMedium)
		Expect(planErr).NotTo(HaveOccurred())
		Expect(planned).To(BeEmpty(),
			"apply must clear every deviation detect planned; leftovers are detect/apply drift")
	})
})
