package constants_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the default exclusion paths that provide baseline coverage
// independent of gogenfilter scan results. Removing any of them silently
// un-excludes classes of generated files for every consumer.
var _ = Describe("DefaultExclusionPaths", func() {
	Describe("DefaultLinterExclusionPaths", func() {
		It("excludes templ generated files", func() {
			Expect(constants.DefaultLinterExclusionPaths).To(ContainElement(`_templ\.go$`))
		})

		It("excludes legacy oapi-codegen .gen.go files", func() {
			// api.gen.go files with legacy text markers may classify as generic
			// in gogenfilter (module-path markers are required for oapi-codegen
			// detection). The `.gen\.go$` default keeps them excluded regardless
			// of which generator gogenfilter attributes them to.
			Expect(constants.DefaultLinterExclusionPaths).To(ContainElement(`\.gen\.go$`))
		})

		It("excludes vendored dependencies", func() {
			Expect(constants.DefaultLinterExclusionPaths).To(ContainElement("vendor/"))
		})
	})

	Describe("DefaultFormatterExclusionPaths", func() {
		It("excludes templ generated files", func() {
			Expect(constants.DefaultFormatterExclusionPaths).To(ContainElement(`_templ\.go$`))
		})
	})
})
