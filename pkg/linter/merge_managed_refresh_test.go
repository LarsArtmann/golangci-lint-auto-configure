package linter_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Managed-settings refresh: a tool-written settings block from an older
// curated generation refreshes to the current curated default under the
// default strategy; user tuning never moves; --force-settings is unchanged.
var _ = Describe("Managed settings refresh", func() {
	env := func() linter.Env {
		return linter.Env{
			Tool:        types.MustParseGolangciVersion("v2.15.0"),
			ToolBuildGo: types.MustParseGoVersion("1.27"),
			LocalGo:     types.MustParseGoVersion("1.27"),
			Tech:        map[string]bool{},
		}
	}

	v0110Wrapcheck := func() map[string]any {
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

	configWithSettings := func(settings map[string]any) *types.Config {
		cfg := &types.Config{}
		cfg.Linters.Settings = map[string]any{"wrapcheck": settings}

		return cfg
	}

	findSettingsChange := func(plan []linter.Change) (linter.Change, bool) {
		for _, change := range plan {
			if change.Path == linter.PathLintersSettings && change.Element == "wrapcheck" {
				return change, true
			}
		}

		return linter.Change{}, false
	}

	It("refreshes a v0.11.0-era tool-written wrapcheck block to the current curated default", func() {
		plan := linter.PlanDesired(
			configWithSettings(v0110Wrapcheck()), linter.Desired(env()),
			linter.StrategiesFor(linter.MergeOptions{}),
		)

		change, found := findSettingsChange(plan)
		Expect(found).To(BeTrue(), "the stale tool-written block must be planned for refresh")

		Expect(change.Kind).To(Equal(linter.ChangeUpdate))
		Expect(change.Reason).To(Equal("refreshed-managed-settings"))

		current, ok := constants.LinterSettings("wrapcheck")
		Expect(ok).To(BeTrue())
		Expect(change.To).To(Equal(current))
	})

	It("plans no settings change when the block is already current", func() {
		current, ok := constants.LinterSettings("wrapcheck")
		Expect(ok).To(BeTrue())

		plan := linter.PlanDesired(
			configWithSettings(current), linter.Desired(env()),
			linter.StrategiesFor(linter.MergeOptions{}),
		)

		_, found := findSettingsChange(plan)
		Expect(found).To(BeFalse(), "a current curated block is a no-op")
	})

	It("never touches a user-tuned block (provenance mismatch)", func() {
		tuned := v0110Wrapcheck()
		tuned["ignore-sigs"] = []any{".Errorf(", "errors.New(", "myHouseSig("}

		plan := linter.PlanDesired(
			configWithSettings(tuned), linter.Desired(env()),
			linter.StrategiesFor(linter.MergeOptions{}),
		)

		_, found := findSettingsChange(plan)
		Expect(found).To(BeFalse(), "user tuning is preserved under the default strategy")
	})

	It("does not fire under --force-settings (the table flip owns that path)", func() {
		plan := linter.PlanDesired(
			configWithSettings(v0110Wrapcheck()), linter.Desired(env()),
			linter.StrategiesFor(linter.MergeOptions{ForceSettings: true}),
		)

		change, found := findSettingsChange(plan)
		Expect(found).To(BeTrue())

		Expect(change.Reason).To(Equal("force-settings"),
			"force-settings has its own reason; the refresh path must not shadow it")
	})
})
