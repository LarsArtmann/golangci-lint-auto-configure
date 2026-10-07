package linter

import (
	"bytes"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/semver"
)

var _ = Describe("validateVersion", func() {
	var (
		analyzer *Analyzer
		buf      *bytes.Buffer
	)

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		logger := log.NewWithOptions(buf, log.Options{
			Level:           log.WarnLevel,
			ReportTimestamp: false,
		})
		analyzer = NewAnalyzer(logger)
	})

	Context("when version equals expected", func() {
		It("should not warn", func() {
			err := analyzer.validateVersion("v2.14.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Context("when version is older than expected but meets minimum", func() {
		It("should warn about unexpected version", func() {
			between := "v2.13.0"
			// Guards rot loudly: bump this spec constant together with the
			// constants when the minimum or recommendation moves.
			Expect(semver.Compare(between, constants.MinGolangCILintVersion)).
				To(BeNumerically(">=", 0), "spec constant must be >= MinGolangCILintVersion — update it when the minimum bumps")
			Expect(semver.Compare(between, constants.ExpectedGolangCILintVersion)).
				To(BeNumerically("<", 0), "spec constant must be < ExpectedGolangCILintVersion — update it when the recommendation bumps")

			err := analyzer.validateVersion(between)
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring(between))
			Expect(buf.String()).To(ContainSubstring(constants.ExpectedGolangCILintVersion))
			Expect(buf.String()).To(ContainSubstring("recommended"))
		})
	})

	Context("when version is newer than expected", func() {
		It("should warn about unexpected version", func() {
			err := analyzer.validateVersion("v2.15.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring("v2.15.0"))
			Expect(buf.String()).To(ContainSubstring(constants.ExpectedGolangCILintVersion))
		})
	})

	Context("when version is below minimum", func() {
		It("should return an error without warning", func() {
			below := "v2.9.0"
			Expect(semver.Compare(below, constants.MinGolangCILintVersion)).
				To(BeNumerically("<", 0), "spec constant must be < MinGolangCILintVersion — update it when the minimum bumps")

			err := analyzer.validateVersion(below)
			Expect(err).To(HaveOccurred())
			Expect(buf.String()).To(BeEmpty())
		})
	})
})
