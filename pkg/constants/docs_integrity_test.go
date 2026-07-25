package constants_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// findRepoRoot walks upward from the current directory until it finds go.mod.
// Go test sets the working directory to the package source directory.
func findRepoRoot() string {
	GinkgoHelper()

	dir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		Expect(parent).ToNot(Equal(dir), "cannot find go.mod — not running inside a Go module")
		dir = parent
	}
}

// readFeaturesMD reads FEATURES.md from the repo root.
func readFeaturesMD() string {
	GinkgoHelper()

	path := filepath.Join(findRepoRoot(), "FEATURES.md")
	data, err := os.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())

	return string(data)
}

// extractPresetCount parses a line like "`strict` preset (20 linters)" and returns 20.
func extractPresetCount(content, preset string) int {
	GinkgoHelper()

	pattern := regexp.MustCompile("`" + regexp.QuoteMeta(preset) + "` preset \\((\\d+) linters")
	match := pattern.FindStringSubmatch(content)
	Expect(match).ToNot(BeNil(), "preset %q linter count not found in FEATURES.md", preset)

	count, err := strconv.Atoi(match[1])
	Expect(err).ToNot(HaveOccurred())

	return count
}

var _ = Describe("FEATURES.md documentation integrity", func() {
	var content string

	BeforeEach(func() {
		content = readFeaturesMD()
	})

	DescribeTable("preset linter count should match pkg/constants/presets.go",
		func(preset string) {
			documented := extractPresetCount(content, preset)
			actual := len(constants.PresetLinters[preset])

			Expect(documented).To(Equal(actual),
				"FEATURES.md documents %s preset with %d linters, but presets.go has %d",
				preset, documented, actual)
		},
		Entry("minimal", "minimal"),
		Entry("standard", "standard"),
		Entry("strict", "strict"),
		Entry("reference", "reference"),
	)

	It("format preset linter and formatter counts should match presets.go", func() {
		pattern := regexp.MustCompile("`format` preset \\((\\d+) linters \\+ (\\d+) formatters\\)")
		match := pattern.FindStringSubmatch(content)
		Expect(match).ToNot(BeNil(), "format preset count pattern not found in FEATURES.md")

		documentedLinters, err := strconv.Atoi(match[1])
		Expect(err).ToNot(HaveOccurred())

		documentedFormatters, err := strconv.Atoi(match[2])
		Expect(err).ToNot(HaveOccurred())

		Expect(documentedLinters).To(Equal(len(constants.PresetLinters["format"])))
		Expect(documentedFormatters).To(Equal(len(constants.PresetFormatters["format"])))
	})

	It("should document test exclusion rule count accurately", func() {
		// FEATURES.md says "Default test exclusion rules (7 linters for _test.go)"
		// The count must match len(DefaultExclusionRules[0].Linters)
		pattern := regexp.MustCompile(`(\d+) linters for \\?_test\.go`)
		match := pattern.FindStringSubmatch(content)
		if match == nil {
			Skip("test exclusion count not documented in FEATURES.md")
		}

		documented, err := strconv.Atoi(match[1])
		Expect(err).ToNot(HaveOccurred())

		actual := len(constants.DefaultExclusionRules[0].Linters)
		Expect(documented).To(Equal(actual),
			"FEATURES.md documents %d test exclusion linters, but code has %d (%s)",
			documented, actual, strings.Join(constants.DefaultExclusionRules[0].Linters, ", "))
	})

	It("should document all test exclusion linter names accurately", func() {
		for _, linter := range constants.DefaultExclusionRules[0].Linters {
			Expect(content).To(ContainSubstring(linter),
				fmt.Sprintf("test exclusion linter %q is missing from FEATURES.md", linter))
		}
	})
})
