package report_test

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReport(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Report Suite")
}

var _ = Describe("Report Generator", func() {
	var (
		logger   *log.Logger
		tmpDir   string
		analysis *types.ConfigAnalysis
	)

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		tmpDir = GinkgoT().TempDir()

		analysis = &types.ConfigAnalysis{
			ConfigPath: ".golangci.yml",
			EnabledLinters: []types.LinterInfo{
				{Name: "gosec", Description: "Security"},
			},
			DisabledLinters: []types.LinterInfo{
				{Name: "funlen", Description: "Function length"},
			},
			LinterRecommendations: []types.LinterRecommendation{
				{Name: "funlen", Priority: types.LinterPriorityHigh, Reason: "Detects long functions"},
			},
			CriticalCount:  0,
			HighValueCount: 1,
		}
	})

	Context("HTML Report", func() {
		It("generates an HTML report file", func() {
			gen := report.NewGenerator(logger)
			outputPath := filepath.Join(tmpDir, "report.html")

			err := gen.GenerateReport(context.Background(), analysis, outputPath)
			Expect(err).NotTo(HaveOccurred())

			content, readErr := os.ReadFile(outputPath)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("gosec"))
		})

		It("returns error for invalid output path", func() {
			gen := report.NewGenerator(logger)

			err := gen.GenerateReport(context.Background(), analysis, "/nonexistent/dir/report.html")
			Expect(err).To(HaveOccurred())
		})

		It("propagates render errors through the named return", func() {
			if _, statErr := os.Stat("/dev/full"); statErr != nil {
				Skip("/dev/full not available on this platform")
			}

			gen := report.NewGenerator(logger)
			err := gen.GenerateReport(context.Background(), analysis, "/dev/full")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("JSON Report", func() {
		It("generates a valid JSON report file", func() {
			gen := report.NewJSONGenerator(logger)
			outputPath := filepath.Join(tmpDir, "report.json")

			err := gen.GenerateJSONReport(analysis, outputPath)
			Expect(err).NotTo(HaveOccurred())

			data, readErr := os.ReadFile(outputPath)
			Expect(readErr).NotTo(HaveOccurred())

			var jsonReport report.JSONReport
			Expect(json.Unmarshal(data, &jsonReport)).To(Succeed())
			Expect(jsonReport.ConfigPath).To(Equal(".golangci.yml"))
			Expect(jsonReport.Summary.EnabledCount).To(Equal(1))
			Expect(jsonReport.Summary.DisabledCount).To(Equal(1))
			Expect(jsonReport.Summary.RecommendationsCount).To(Equal(1))
		})

		It("returns error for invalid output path", func() {
			gen := report.NewJSONGenerator(logger)

			err := gen.GenerateJSONReport(analysis, "/nonexistent/dir/report.json")
			Expect(err).To(HaveOccurred())
		})
	})
})
