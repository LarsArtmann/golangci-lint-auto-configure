package linter

import (
	"bytes"

	"charm.land/log/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			err := analyzer.validateVersion("v2.12.2")
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Context("when version is older than expected but meets minimum", func() {
		It("should warn about unexpected version", func() {
			err := analyzer.validateVersion("v2.11.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring("v2.11.0"))
			Expect(buf.String()).To(ContainSubstring("v2.12.2"))
			Expect(buf.String()).To(ContainSubstring("recommended"))
		})
	})

	Context("when version is newer than expected", func() {
		It("should warn about unexpected version", func() {
			err := analyzer.validateVersion("v2.13.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(ContainSubstring("v2.13.0"))
			Expect(buf.String()).To(ContainSubstring("v2.12.2"))
		})
	})

	Context("when version is below minimum", func() {
		It("should return an error without warning", func() {
			err := analyzer.validateVersion("v2.9.0")
			Expect(err).To(HaveOccurred())
			Expect(buf.String()).To(BeEmpty())
		})
	})
})
