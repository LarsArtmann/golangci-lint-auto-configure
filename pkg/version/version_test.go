package version_test

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version Suite")
}

var _ = Describe("Info", func() {
	Describe("Short", func() {
		It("returns the version string", func() {
			info := version.Info{Version: "v1.2.3"}
			Expect(info.Short()).To(Equal("v1.2.3"))
		})
	})

	Describe("String", func() {
		It("returns formatted single-line version info", func() {
			info := version.Info{
				Version: "v1.2.3",
				Commit:  "abc1234",
				Date:    "2026-05-06",
			}
			Expect(info.String()).To(Equal("v1.2.3 (commit: abc1234, built: 2026-05-06)"))
		})
	})

	Describe("Full", func() {
		It("returns multi-line version info", func() {
			info := version.Info{
				Version:   "v1.2.3",
				Commit:    "abc1234",
				Date:      "2026-05-06",
				TreeState: "clean",
			}
			expected := "Version:    v1.2.3\nCommit:     abc1234\nBuilt:      2026-05-06\nTree:       clean"
			Expect(info.Full()).To(Equal(expected))
		})

		It("shows dirty tree state", func() {
			info := version.Info{
				Version:   "v1.2.3-dirty",
				Commit:    "abc1234",
				Date:      "2026-05-06",
				TreeState: "dirty",
			}
			Expect(info.Full()).To(ContainSubstring("Tree:       dirty"))
		})
	})
})

var _ = Describe("Get", func() {
	It("returns a non-empty Info", func() {
		info := version.Get()
		Expect(info.Version).NotTo(BeEmpty())
		Expect(info.Commit).NotTo(BeEmpty())
		Expect(info.Date).NotTo(BeEmpty())
		Expect(info.TreeState).NotTo(BeEmpty())
	})

	It("returns consistent results on repeated calls", func() {
		info1 := version.Get()
		info2 := version.Get()
		Expect(info1).To(Equal(info2))
	})
})
