package constants_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// v0110WrapcheckBlock re-expresses the v0.11.x wrapcheck curated default
// independently of constants.wrapcheckV0110Block: the test pin is only as
// strong as its independence from the code under test.
func v0110WrapcheckBlock() map[string]any {
	return map[string]any{
		"ignore-sigs": []any{
			".Errorf(", "errors.New(", "errors.Unwrap(", "errors.Join(",
			".Wrap(", ".Wrapf(", ".WithMessage(", ".WithMessagef(", ".WithStack(",
		},
		"ignore-sig-regexps": []any{
			"encoding/json", "fmt\\.", "errors\\.", "slices\\.", "maps\\.",
			"sort\\.", "time\\.", "net/http", "os\\.", "io\\.", "strings\\.",
			"strconv\\.", "path\\.", "filepath\\.", "compress/gzip",
			"\\(context\\.Context\\)\\.Err",
		},
	}
}

func currentCuratedLinterBlock(name string) map[string]any {
	block, ok := constants.LinterSettings(types.LinterName(name))
	if !ok {
		return nil
	}

	return block
}

var _ = Describe("Managed settings provenance", func() {
	Describe("HashSettingsBlock", func() {
		It("hashes structurally equal blocks identically regardless of key insertion order", func() {
			a := map[string]any{
				"ignore-sigs":        []any{".Errorf(", "errors.New("},
				"ignore-sig-regexps": []any{"fmt\\.", "os\\."},
			}
			b := map[string]any{
				"ignore-sig-regexps": []any{"fmt\\.", "os\\."},
				"ignore-sigs":        []any{".Errorf(", "errors.New("},
			}

			Expect(constants.HashSettingsBlock(a)).To(Equal(constants.HashSettingsBlock(b)))
		})

		It("distinguishes blocks that differ in any value", func() {
			a := map[string]any{"min-len": 4}
			b := map[string]any{"min-len": 5}

			Expect(constants.HashSettingsBlock(a)).NotTo(Equal(constants.HashSettingsBlock(b)))
		})

		It("preserves list order in the hash (order is semantic)", func() {
			a := map[string]any{"ignore-sigs": []any{"x", "y"}}
			b := map[string]any{"ignore-sigs": []any{"y", "x"}}

			Expect(constants.HashSettingsBlock(a)).NotTo(Equal(constants.HashSettingsBlock(b)))
		})
	})

	Describe("ManagedSettingsHistory", func() {
		It("records the v0.11.0 wrapcheck generation as tool-written but stale", func() {
			Expect(constants.IsHistoricalCuratedLinterBlock("wrapcheck", v0110WrapcheckBlock())).To(BeTrue())
		})

		It("does not claim user-tuned or unknown blocks", func() {
			tuned := v0110WrapcheckBlock()
			tuned["ignore-sigs"] = []any{".Errorf(", "my custom sig("}

			Expect(constants.IsHistoricalCuratedLinterBlock("wrapcheck", tuned)).To(BeFalse())

			gosecBlock := map[string]any{"excludes": []any{"G304"}}

			Expect(constants.IsHistoricalCuratedLinterBlock("gosec", gosecBlock)).To(BeFalse())
			Expect(constants.IsHistoricalCuratedLinterBlock("wrapcheck", map[string]any{})).To(BeFalse())
		})

		It("never records a stale entry that equals the live curated block (it would be noise)", func() {
			for _, generation := range constants.ManagedSettingsHistory {
				for name := range generation.LinterSettings {
					current := currentCuratedLinterBlock(name)
					Expect(current).NotTo(BeNil(), "history names %q, which must exist in the live table", name)
					Expect(constants.IsHistoricalCuratedLinterBlock(name, current)).
						To(BeFalse(), "history entry %s@%s equals the live curated block; "+
							"the entry is dead weight and the content change it recorded never shipped",
							name, generation.Since)
				}
			}
		})
	})
})
