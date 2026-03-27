package linter_test

import (
	"context"
	"testing"

	"charm.land/log/v2"
	linterpkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLinter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linter Suite")
}

var _ = Describe("Version Check", func() {
	var analyzer *linterpkg.Analyzer

	BeforeEach(func() {
		analyzer = linterpkg.NewAnalyzer(log.Default())
	})

	Context("When golangci-lint is installed", func() {
		It("should find the binary", func() {
			err := analyzer.FindBinary(context.Background())
			if err != nil {
				Skip("golangci-lint not found in PATH")
			}

			Expect(err).ToNot(HaveOccurred())
		})

		It("should pass version check with v2.10.1 or newer", func() {
			err := analyzer.FindBinary(context.Background())
			if err != nil {
				Skip("golangci-lint not found in PATH")
			}

			err = analyzer.CheckVersion(context.Background())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
