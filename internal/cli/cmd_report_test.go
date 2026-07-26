package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("report command", func() {
	It("should generate JSON report", func() {
		binaryPath := buildBinary()
		configPath := writeConfig(`version: "2"
linters:
  enable:
    - errcheck
`)

		reportPath := filepath.Join(testDir, "report.json")
		cmd := exec.Command(
			binaryPath,
			"report",
			"--config",
			configPath,
			"--output",
			reportPath,
			"--format",
			"json",
		)
		_, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		_, err = os.Stat(reportPath)
		Expect(err).NotTo(HaveOccurred())
	})

	// testReportFormat verifies a report format generates output containing all expected substrings.
	testReportFormat := func(ext, format string, expected []string) {
		binaryPath := buildBinary()
		configPath := writeConfig(`version: "2"
linters:
  enable:
    - errcheck
`)

		data := generateReport(binaryPath, configPath, ext, format)
		for _, e := range expected {
			Expect(data).To(ContainSubstring(e))
		}
	}

	It("should generate HTML report", func() {
		testReportFormat("html", "html", []string{"<!doctype html>", "golangci-lint"})
	})

	It("should generate SARIF report", func() {
		testReportFormat(
			"sarif.json",
			"sarif",
			[]string{`"$schema"`, `"golangci-lint-auto-configure"`},
		)
	})

	It("should generate finding report", func() {
		testReportFormat(
			"finding.json",
			"finding",
			[]string{`"golangci-lint-auto-configure"`, `"findings"`, `"summary"`},
		)
	})

	It("should handle missing config file", func() {
		binaryPath := buildBinary()
		cmd := exec.Command(
			binaryPath,
			"report",
			"--config",
			"/non/existent/path.yml",
			"--output",
			filepath.Join(testDir, "report.json"),
			"--format",
			"json",
		)
		_, err := cmd.CombinedOutput()
		Expect(err).To(HaveOccurred())
	})
})
