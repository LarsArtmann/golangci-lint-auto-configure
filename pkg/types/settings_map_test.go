package types_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SettingsMap", func() {
	Describe("AsSettingsMap", func() {
		It("converts a valid map[string]any", func() {
			sm, ok := types.AsSettingsMap(map[string]any{"key": "value"})
			Expect(ok).To(BeTrue())
			Expect(sm).To(HaveKey("key"))
			Expect(sm["key"]).To(Equal("value"))
		})

		It("returns false for nil", func() {
			sm, ok := types.AsSettingsMap(nil)
			Expect(ok).To(BeFalse())
			Expect(sm).To(BeNil())
		})

		It("returns false for non-map types", func() {
			for _, input := range []any{
				"string",
				42,
				true,
				[]any{"a", "b"},
				map[string]string{"key": "value"},
			} {
				sm, ok := types.AsSettingsMap(input)
				Expect(ok).To(BeFalse(), "expected false for %v", input)
				Expect(sm).To(BeNil())
			}
		})
	})

	Describe("IsEmpty", func() {
		It("returns true for nil", func() {
			var sm types.SettingsMap
			Expect(sm.IsEmpty()).To(BeTrue())
		})

		It("returns true for an empty map", func() {
			Expect(types.SettingsMap{}.IsEmpty()).To(BeTrue())
		})

		It("returns false for a populated map", func() {
			sm := types.SettingsMap{"key": "value"}
			Expect(sm.IsEmpty()).To(BeFalse())
		})
	})

	Describe("Clone", func() {
		It("returns nil for a nil receiver", func() {
			var sm types.SettingsMap
			Expect(sm.Clone()).To(BeNil())
		})

		It("returns an independent empty map for an initialized empty map", func() {
			cloned := types.SettingsMap{}.Clone()
			Expect(cloned).ToNot(BeNil())
			Expect(cloned.IsEmpty()).To(BeTrue())
		})

		It("deep-copies nested maps", func() {
			original := types.SettingsMap{
				"gocritic": map[string]any{"enabled-tags": []any{"diagnostic"}},
			}

			cloned := original.Clone()

			clonedNested, ok := cloned["gocritic"].(map[string]any)
			Expect(ok).To(BeTrue())

			clonedNested["enabled-tags"] = []any{"performance"}

			originalNested := original["gocritic"].(map[string]any)
			Expect(originalNested["enabled-tags"]).To(Equal([]any{"diagnostic"}))
		})

		It("deep-copies nested slices", func() {
			original := types.SettingsMap{
				"govet": map[string]any{
					"enable": []any{"shadow", "printf"},
				},
			}

			cloned := original.Clone()

			clonedSlice := cloned["govet"].(map[string]any)["enable"].([]any)
			clonedSlice[0] = "modified"

			originalSlice := original["govet"].(map[string]any)["enable"].([]any)
			Expect(originalSlice[0]).To(Equal("shadow"))
		})

		It("preserves primitive values", func() {
			original := types.SettingsMap{
				"timeout": "5m",
				"count":   42,
				"flag":    true,
			}

			cloned := original.Clone()

			Expect(cloned["timeout"]).To(Equal("5m"))
			Expect(cloned["count"]).To(Equal(42))
			Expect(cloned["flag"]).To(BeTrue())
		})

		It("produces a map that is not the same reference", func() {
			original := types.SettingsMap{"key": "value"}
			cloned := original.Clone()

			Expect(cloned).ToNot(BeIdenticalTo(original))

			cloned["new"] = "entry"

			Expect(original).ToNot(HaveKey("new"))
		})
	})
})
