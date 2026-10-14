package linter

import (
	"errors"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Property specs for the merge/diff engine (plan phase P4). These are the
// structural guarantees the imperative fixer could only approximate with
// counters and guards.
var _ = Describe("PlanDesired", func() {
	basicEnv := func() Env {
		return Env{
			Tool:        types.MustParseGolangciVersion("v2.14.0"),
			ToolBuildGo: types.MustParseGoVersion("1.27"),
			LocalGo:     types.MustParseGoVersion("1.27"),
			Tech:        map[string]bool{},
		}
	}

	emptyConfig := func() *types.Config {
		return &types.Config{}
	}

	It("plans adds for every missing desired linter", func() {
		plan := PlanDesired(emptyConfig(), Desired(basicEnv()), StrategiesFor(MergeOptions{}))

		enableAdds := make(map[string]bool)

		for _, change := range plan {
			if change.Path == "linters.enable" {
				enableAdds[change.Element] = true
			}
		}

		Expect(enableAdds).To(HaveKey("errcheck"))
		Expect(enableAdds).To(HaveKey("wrapcheck"))
		Expect(enableAdds).NotTo(HaveKey("depguard"))
	})

	It("plans nothing for a config already at the desired state", func() {
		once := ApplyPlan(emptyConfig(), PlanDesired(emptyConfig(), Desired(basicEnv()), StrategiesFor(MergeOptions{})))

		secondPlan := PlanDesired(once, Desired(basicEnv()), StrategiesFor(MergeOptions{}))
		Expect(secondPlan).To(BeEmpty(), "a converged config must re-plan to nothing")
	})

	It("is idempotent: ApplyPlan(Plan(x)) twice equals once (merge(merge(x)) == merge(x))", func() {
		strategies := StrategiesFor(MergeOptions{})
		desired := Desired(basicEnv())
		observed := emptyConfig()

		first := ApplyPlan(observed, PlanDesired(observed, desired, strategies))
		second := ApplyPlan(first, PlanDesired(first, desired, strategies))

		Expect(second.Linters.Enable).To(Equal(first.Linters.Enable))
		Expect(second.Formatters.Enable).To(Equal(first.Formatters.Enable))
		Expect(second.Run.Go).To(Equal(first.Run.Go))
		Expect(second.Issues).To(Equal(first.Issues))
	})

	It("never mutates the observed config", func() {
		observed := emptyConfig()
		observed.Run.Go = "1.26"
		before := observed.Clone()

		_ = PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{}))
		_ = ApplyPlan(observed, PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{})))

		Expect(observed).To(Equal(before))
	})

	It("keeps user settings untouched under PreferUser", func() {
		observed := emptyConfig()
		observed.Linters.Enable = []types.LinterName{"errcheck"}
		observed.Linters.Settings = map[string]any{
			"goconst": map[string]any{"min-len": 7, "min-occurrences": 9},
		}

		plan := PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{}))

		for _, change := range plan {
			if change.Path == "linters.settings" && change.Element == "goconst" {
				Fail("existing goconst settings must never be re-planned under prefer-user")
			}
		}
	})

	It("flips settings to PreferDesired under ForceSettings", func() {
		observed := emptyConfig()
		observed.Linters.Settings = map[string]any{
			"goconst": map[string]any{"min-len": 7, "min-occurrences": 9},
		}

		gentle := PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{}))
		forced := PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{ForceSettings: true}))

		gentleTouches := false

		for _, change := range gentle {
			if change.Path == "linters.settings" && change.Element == "goconst" {
				gentleTouches = true
			}
		}

		forcedTouches := false

		for _, change := range forced {
			if change.Path == "linters.settings" && change.Element == "goconst" {
				forcedTouches = true

				Expect(change.Kind).To(Equal(ChangeUpdate))
			}
		}

		Expect(gentleTouches).To(BeFalse(), "prefer-user must leave existing settings alone")
		Expect(forcedTouches).To(BeTrue(), "prefer-desired must rewrite diverging settings")
	})

	It("plans a run.go update when the observed value is empty or patch-form", func() {
		observed := emptyConfig()
		observed.Run.Go = "1.27.1"

		plan := PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{}))

		var runGo *Change

		for i := range plan {
			if plan[i].Path == "run.go" {
				runGo = &plan[i]
			}
		}

		Expect(runGo).NotTo(BeNil())
		Expect(runGo.To).To(Equal("1.27"))
	})

	It("tracks run.go to the capped local Go, matching legacy behavior", func() {
		observed := emptyConfig()
		observed.Run.Go = "1.24"

		plan := PlanDesired(observed, Desired(basicEnv()), StrategiesFor(MergeOptions{}))

		var runGo *Change

		for i := range plan {
			if plan[i].Path == "run.go" {
				runGo = &plan[i]
			}
		}

		Expect(runGo).NotTo(BeNil(), "run.go must track the local toolchain (parity corpus contract)")
		Expect(runGo.To).To(Equal("1.27"))
		Expect(runGo.From).To(Equal("1.24"))
	})

	It("plans issues and runner defaults on an empty config", func() {
		plan := PlanDesired(emptyConfig(), Desired(basicEnv()), StrategiesFor(MergeOptions{}))

		paths := make(map[string]bool)
		for _, change := range plan {
			paths[change.Path] = true
		}

		Expect(paths).To(HaveKey("issues.max-issues-per-linter"))
		Expect(paths).To(HaveKey("issues.max-same-issues"))
		Expect(paths).To(HaveKey("run.allow-parallel-runners"))
		Expect(paths).To(HaveKey("linters.exclusions.paths"))
	})

	It("validates a clean plan without conflicts", func() {
		plan := PlanDesired(emptyConfig(), Desired(basicEnv()), StrategiesFor(MergeOptions{}))
		Expect(ValidatePlan(plan)).To(Succeed())
	})

	It("detects conflicting writers on the same stable key", func() {
		conflicted := []Change{
			{Path: "run.go", Kind: ChangeUpdate, To: "1.26"},
			{Path: "run.go", Kind: ChangeUpdate, To: "1.27"},
		}

		err := ValidatePlan(conflicted)
		Expect(err).To(HaveOccurred())

		var conflict *ConflictError
		Expect(asConflict(err, &conflict)).To(BeTrue())
		Expect(conflict.Key).To(Equal("run.go"))
	})

	It("detects kind conflicts on the same stable key", func() {
		conflicted := []Change{
			{Path: "linters.enable", Element: "godot", Kind: ChangeAdd},
			{Path: "linters.enable", Element: "godot", Kind: ChangeRemove},
		}

		Expect(ValidatePlan(conflicted)).To(HaveOccurred())
	})

	It("accepts duplicate identical changes as non-conflicting", func() {
		duplicated := []Change{
			{Path: "run.go", Kind: ChangeUpdate, To: "1.27"},
			{Path: "run.go", Kind: ChangeUpdate, To: "1.27"},
		}

		Expect(ValidatePlan(duplicated)).To(Succeed())
	})

	It("produces stable change keys across runs", func() {
		first := PlanDesired(emptyConfig(), Desired(basicEnv()), StrategiesFor(MergeOptions{}))
		second := PlanDesired(emptyConfig(), Desired(basicEnv()), StrategiesFor(MergeOptions{}))

		firstKeys := make([]string, 0, len(first))
		for _, change := range first {
			firstKeys = append(firstKeys, change.Key())
		}

		secondKeys := make([]string, 0, len(second))
		for _, change := range second {
			secondKeys = append(secondKeys, change.Key())
		}

		Expect(secondKeys).To(Equal(firstKeys))
	})
})

func asConflict(err error, target **ConflictError) bool {
	conflict := &ConflictError{}
	ok := errors.As(err, &conflict)
	if ok {
		*target = conflict
	}

	return ok
}
