// HTML report golden snapshot test.
//
// Workflow: when you change pkg/report/report.templ, the golden file will
// fail to match. To regenerate it:
//
//	UPDATE_GOLDEN=1 go test ./pkg/report/...
//
// Then review the diff in the committed golden file (testdata/golden/report.html)
// to ensure the change is intentional before committing.
package report_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func renderAnalysisToHTML(a *types.ConfigAnalysis) string {
	data := report.ReportData{Analysis: a}

	var buf bytes.Buffer

	err := report.Report(data).Render(context.Background(), &buf)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

var _ = Describe("HTML Report Golden Snapshot", func() {
	var (
		analysis   *types.ConfigAnalysis
		goldenPath string
	)

	BeforeEach(func() {
		goldenPath = filepath.Join("testdata", "golden", "report.html")

		analysis = &types.ConfigAnalysis{
			ConfigPath: ".golangci.yml",
			EnabledLinters: []types.LinterInfo{
				{Name: "govet", Description: "Vet", Since: "v1.0.0"},
				{Name: "errcheck", Description: "Errcheck", Since: "v1.0.0"},
			},
			DisabledLinters: []types.LinterInfo{
				{Name: "gosec", Description: "Security", Since: "v1.0.0"},
			},
			LinterRecommendations: []types.LinterRecommendation{
				{Name: "gosec", Priority: types.LinterPriorityCritical, Reason: "Security"},
				{Name: "funlen", Priority: types.LinterPriorityHigh, Reason: "Long functions"},
			},
			CriticalCount:    1,
			HighValueCount:   1,
			MediumValueCount: 0,
			OptionalCount:    0,
			DeprecatedCount:  0,
		}
	})

	Context("golden file snapshot", func() {
		It("matches the committed golden file", func() {
			actual := renderAnalysisToHTML(analysis)

			if os.Getenv("UPDATE_GOLDEN") == "1" {
				Expect(os.MkdirAll(filepath.Dir(goldenPath), 0o755)).To(Succeed())
				Expect(os.WriteFile(goldenPath, []byte(actual), 0o644)).To(Succeed())
				Skip("golden file updated")
			}

			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				Fail("golden file not found — run UPDATE_GOLDEN=1 to create it")
			}

			Expect(string(expected)).To(Equal(actual))
		})

		It("creates the golden file if missing", func() {
			if _, err := os.Stat(goldenPath); err == nil {
				Skip("golden file already exists")
			}

			actual := renderAnalysisToHTML(analysis)

			Expect(os.MkdirAll(filepath.Dir(goldenPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(goldenPath, []byte(actual), 0o644)).To(Succeed())
		})
	})

	Context("structural invariants", func() {
		It("always renders the doctype declaration", func() {
			html := renderAnalysisToHTML(analysis)
			Expect(strings.ToLower(html[:15])).To(ContainSubstring("<!doctype html>"))
		})

		It("always renders the config path in the header", func() {
			html := renderAnalysisToHTML(analysis)
			Expect(html).To(ContainSubstring(".golangci.yml"))
		})

		It("renders the success message when no recommendations", func() {
			noRecs := &types.ConfigAnalysis{
				ConfigPath:       "clean.yml",
				EnabledLinters:   []types.LinterInfo{{Name: "govet"}},
				CriticalCount:    0,
				HighValueCount:   0,
				MediumValueCount: 0,
				OptionalCount:    0,
			}
			html := renderAnalysisToHTML(noRecs)
			Expect(html).To(ContainSubstring("Perfect Configuration"))
			Expect(html).To(ContainSubstring("All recommended linters are already enabled"))
		})

		It("renders linter names from enabled and disabled lists", func() {
			html := renderAnalysisToHTML(analysis)
			Expect(html).To(ContainSubstring("govet"))
			Expect(html).To(ContainSubstring("errcheck"))
			Expect(html).To(ContainSubstring("gosec"))
		})

		It("renders priority sections when recommendations exist", func() {
			html := renderAnalysisToHTML(analysis)
			Expect(html).To(ContainSubstring("Critical Linters"))
			Expect(html).To(ContainSubstring("High Priority Linters"))
			Expect(html).To(ContainSubstring("Medium Priority Linters"))
			Expect(html).To(ContainSubstring("Optional Linters"))
		})

		It("contains valid HTML structure (open and close tags balanced)", func() {
			html := renderAnalysisToHTML(analysis)
			Expect(strings.Count(html, "<html")).To(Equal(strings.Count(html, "</html>")))
			Expect(strings.Count(html, "<head")).To(Equal(strings.Count(html, "</head>")))
			Expect(strings.Count(html, "<body")).To(Equal(strings.Count(html, "</body>")))
		})
	})
})
