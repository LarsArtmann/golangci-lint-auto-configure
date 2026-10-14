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

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func renderEvaluationToHTML(e *linter.Evaluation) string {
	data := report.ReportData{Evaluation: e}

	var buf bytes.Buffer

	err := report.Report(data).Render(context.Background(), &buf)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

var _ = Describe("HTML Report Golden Snapshot", func() {
	var (
		evaluation *linter.Evaluation
		goldenPath string
	)

	BeforeEach(func() {
		goldenPath = filepath.Join("testdata", "golden", "report.html")

		evaluation = &linter.Evaluation{
			ConfigPath: ".golangci.yml",
			Observed: &types.Config{
				Linters: types.LintersConfig{
					Enable:  []types.LinterName{"govet", "errcheck"},
					Disable: []types.LinterName{"gosec"},
				},
			},
			Deviations: []linter.Change{
				{
					Path: linter.PathLintersEnable, Element: "gosec", Kind: linter.ChangeAdd,
					To:     "gosec",
					Reason: "Inspects source code for security problems",
				},
				{
					Path: "run.timeout", Kind: linter.ChangeUpdate,
					From: "1s", To: "5m",
					Reason: "invalid duration",
				},
			},
		}
	})

	Context("golden file snapshot", func() {
		It("matches the committed golden file", func() {
			actual := renderEvaluationToHTML(evaluation)

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

			actual := renderEvaluationToHTML(evaluation)

			Expect(os.MkdirAll(filepath.Dir(goldenPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(goldenPath, []byte(actual), 0o644)).To(Succeed())
		})
	})

	Context("structural invariants", func() {
		It("always renders the doctype declaration", func() {
			html := renderEvaluationToHTML(evaluation)
			Expect(strings.ToLower(html[:15])).To(ContainSubstring("<!doctype html>"))
		})

		It("always renders the config path in the header", func() {
			html := renderEvaluationToHTML(evaluation)
			Expect(html).To(ContainSubstring(".golangci.yml"))
		})

		It("renders the converged message when no deviations exist", func() {
			converged := &linter.Evaluation{
				ConfigPath: "clean.yml",
				Observed: &types.Config{
					Linters: types.LintersConfig{
						Enable: []types.LinterName{"govet"},
					},
				},
				Converged: true,
			}
			html := renderEvaluationToHTML(converged)
			Expect(html).To(ContainSubstring("Converged Configuration"))
		})

		It("renders linter names from enabled and disabled lists", func() {
			html := renderEvaluationToHTML(evaluation)
			Expect(html).To(ContainSubstring("govet"))
			Expect(html).To(ContainSubstring("errcheck"))
			Expect(html).To(ContainSubstring("gosec"))
		})

		It("renders deviation sections when deviations exist", func() {
			html := renderEvaluationToHTML(evaluation)
			Expect(html).To(ContainSubstring("Missing Linters"))
			Expect(html).To(ContainSubstring("Value Updates"))
			Expect(html).To(ContainSubstring("deviation-table"))
		})

		It("contains valid HTML structure (open and close tags balanced)", func() {
			html := renderEvaluationToHTML(evaluation)
			Expect(strings.Count(html, "<html")).To(Equal(strings.Count(html, "</html>")))
			Expect(strings.Count(html, "<head")).To(Equal(strings.Count(html, "</head>")))
			Expect(strings.Count(html, "<body")).To(Equal(strings.Count(html, "</body>")))
		})
	})
})
