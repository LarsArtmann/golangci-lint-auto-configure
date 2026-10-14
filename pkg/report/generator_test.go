package report_test

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReport(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Report Suite")
}

// evaluationFixture builds the shared report-test evaluation: one missing
// linter deviation, one enabled, one disabled.
func evaluationFixture() *linter.Evaluation {
	return &linter.Evaluation{
		ConfigPath: ".golangci.yml",
		Observed: &types.Config{
			Linters: types.LintersConfig{
				Enable:  []types.LinterName{"gosec"},
				Disable: []types.LinterName{"funlen"},
			},
		},
		Deviations: []linter.Change{
			{
				Path: linter.PathLintersEnable, Element: "funlen", Kind: linter.ChangeAdd,
				To:     "funlen",
				Reason: "Detects long functions",
			},
		},
	}
}

var _ = Describe("Report Generator", func() {
	var (
		logger     *log.Logger
		tmpDir     string
		evaluation *linter.Evaluation
	)

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		tmpDir = GinkgoT().TempDir()
		evaluation = evaluationFixture()
	})

	Context("HTML Report", func() {
		It("generates an HTML report file", func() {
			gen := report.NewGenerator(logger)
			outputPath := filepath.Join(tmpDir, "report.html")

			err := gen.GenerateReport(context.Background(), evaluation, outputPath)
			Expect(err).NotTo(HaveOccurred())

			content, readErr := os.ReadFile(outputPath)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("gosec"))
		})

		It("returns error for invalid output path", func() {
			gen := report.NewGenerator(logger)

			err := gen.GenerateReport(context.Background(), evaluation, "/nonexistent/dir/report.html")
			Expect(err).To(HaveOccurred())
		})

		It("propagates render errors through the named return", func() {
			if _, statErr := os.Stat("/dev/full"); statErr != nil {
				Skip("/dev/full not available on this platform")
			}

			gen := report.NewGenerator(logger)
			err := gen.GenerateReport(context.Background(), evaluation, "/dev/full")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("JSON Report", func() {
		It("generates a valid JSON report file", func() {
			gen := report.NewJSONGenerator(logger)
			outputPath := filepath.Join(tmpDir, "report.json")

			err := gen.GenerateJSONReport(evaluation, outputPath)
			Expect(err).NotTo(HaveOccurred())

			data, readErr := os.ReadFile(outputPath)
			Expect(readErr).NotTo(HaveOccurred())

			var jsonReport report.JSONReport
			Expect(json.Unmarshal(data, &jsonReport)).To(Succeed())
			Expect(jsonReport.ConfigPath).To(Equal(".golangci.yml"))
			Expect(jsonReport.Summary.EnabledCount).To(Equal(1))
			Expect(jsonReport.Summary.DisabledCount).To(Equal(1))
			Expect(jsonReport.Summary.DeviationsCount).To(Equal(1))
		})

		It("returns error for invalid output path", func() {
			gen := report.NewJSONGenerator(logger)

			err := gen.GenerateJSONReport(evaluation, "/nonexistent/dir/report.json")
			Expect(err).To(HaveOccurred())
		})
	})
})
