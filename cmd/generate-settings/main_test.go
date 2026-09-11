package main

import (
	"go/format"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGenerateSettings(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Generate Settings Suite")
}

var _ = Describe("generateStruct", func() {
	It("emits a gofmt-clean struct with fields and ToMap", func() {
		def := schemaDef{
			Properties: map[string]schemaDef{
				"max-issues-per-linter": {Type: "integer"},
				"fix":                   {Type: "boolean"},
			},
		}

		var b strings.Builder
		generateStruct(&b, "dummylinter", def)

		_, err := format.Source([]byte(b.String()))
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(ContainSubstring("MaxIssuesPerLinter int `yaml:\"max-issues-per-linter\"`"))
		Expect(b.String()).To(ContainSubstring(`"max-issues-per-linter": s.MaxIssuesPerLinter,`))
	})

	It("emits an empty struct body that satisfies gofumpt (struct{})", func() {
		var b strings.Builder
		generateStruct(&b, "emptylinter", schemaDef{})

		output := b.String()
		Expect(output).To(ContainSubstring("type EmptylinterSettings struct{}"))
		Expect(output).NotTo(ContainSubstring("struct {\n}"))
	})
})
