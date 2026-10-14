package constants_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// linterSettingsMap returns the base curated settings of a linter as a map,
// failing the spec when the table carries none.
func linterSettingsMap(name types.LinterName) map[string]any {
	settings, ok := constants.LinterSettings(name)
	Expect(ok).To(BeTrue(), "linter %q has no curated settings block in the policy table", name)

	return settings.ToMap()
}

// formatterSettingsMap returns the base curated settings of a formatter as a
// map, failing the spec when the table carries none.
func formatterSettingsMap(name types.FormatterName) map[string]any {
	settings, ok := constants.FormatterSettings(name)
	Expect(ok).To(BeTrue(), "formatter %q has no curated settings block in the policy table", name)

	return settings.ToMap()
}

// settingsBearingLinterPolicies returns every linter record with a curated
// settings block.
func settingsBearingLinterPolicies() []types.LinterPolicy {
	var records []types.LinterPolicy

	for _, record := range constants.LinterPolicies {
		if len(record.Settings) > 0 {
			records = append(records, record)
		}
	}

	return records
}

// deprecatedLinterRecords returns every linter record carrying a supersession
// reason (the FEATURES migration-map contract).
func deprecatedLinterRecords() []types.LinterPolicy {
	var records []types.LinterPolicy

	for _, record := range constants.LinterPolicies {
		if record.SupersessionReason != "" {
			records = append(records, record)
		}
	}

	return records
}

var _ = Describe("Reference preset", func() {
	It("should only contain linters present in LinterPriorities", func() {
		referenceLinters, ok := constants.PresetLinters["reference"]
		Expect(ok).To(BeTrue(), "reference preset must exist")

		for _, linter := range referenceLinters {
			_, exists := constants.PolicyRecord(linter)
			Expect(exists).
				To(BeTrue(), "reference preset contains %q which is missing from the policy table", linter)
		}
	})

	It("should contain all critical and high priority linters", func() {
		referenceLinters := constants.PresetLinters["reference"]
		referenceSet := make(map[string]struct{}, len(referenceLinters))

		for _, l := range referenceLinters {
			referenceSet[string(l)] = struct{}{}
		}

		for _, record := range constants.LinterPolicies {
			if record.Priority == types.LinterPriorityCritical || record.Priority == types.LinterPriorityHigh {
				_, exists := referenceSet[string(record.Name)]
				Expect(exists).
					To(BeTrue(), "%s (priority=%s) is missing from reference preset", record.Name, record.Priority)
			}
		}
	})
})

var _ = Describe("Formatter table report data", func() {
	It("should give every formatter record a priority and a rationale (report data)", func() {
		for _, record := range constants.FormatterPolicies {
			Expect(record.Rationale).ToNot(BeEmpty(),
				"formatter %q has no recommendation rationale", record.Name)
		}
	})
})

var _ = Describe("Linter settings blocks", func() {
	It("should produce non-empty ToMap output for every settings-bearing record", func() {
		for _, record := range settingsBearingLinterPolicies() {
			m := record.Settings[0].Settings.ToMap()
			Expect(m).
				ToNot(BeEmpty(), "linter %q carries an empty curated settings block", record.Name)
		}
	})

	It("should have cyclop max-complexity as int in ToMap output", func() {
		m := linterSettingsMap("cyclop")
		maxComplexity, ok := m["max-complexity"]
		Expect(ok).To(BeTrue(), "cyclop settings missing max-complexity key")
		Expect(maxComplexity).To(Equal(12))
	})

	It("should never carry curated settings on a disabled-tier linter", func() {
		for _, record := range constants.LinterPolicies {
			if record.Tier.Kind != types.TierDisabled {
				continue
			}

			Expect(record.Settings).To(BeEmpty(),
				"disabled linter %q carries curated settings — dead weight the prune pass must chase", record.Name)
		}
	})
})

var _ = Describe("Formatter settings blocks", func() {
	It("should produce non-empty ToMap output for every settings-bearing record", func() {
		for _, record := range constants.FormatterPolicies {
			if len(record.Settings) == 0 {
				continue
			}

			m := record.Settings[0].Settings.ToMap()
			Expect(m).
				ToNot(BeEmpty(), "formatter %q carries an empty curated settings block", record.Name)
		}
	})

	It("should have golines max-len as int in ToMap output", func() {
		m := formatterSettingsMap("golines")
		maxLen, ok := m["max-len"]
		Expect(ok).To(BeTrue(), "golines settings missing max-len key")
		Expect(maxLen).To(Equal(120))
	})
})

var _ = Describe("Format preset", func() {
	It("should exist in PresetLinters", func() {
		_, ok := constants.PresetLinters["format"]
		Expect(ok).To(BeTrue(), "format preset missing from PresetLinters")
	})

	It("should exist in PresetFormatters with core formatters", func() {
		formatters, ok := constants.PresetFormatters["format"]
		Expect(ok).To(BeTrue(), "format preset missing from PresetFormatters")
		Expect(formatters).To(ContainElement(types.FormatterName("gofumpt")))
		Expect(formatters).To(ContainElement(types.FormatterName("goimports")))
		Expect(formatters).NotTo(ContainElement(types.FormatterName("gci")),
			"gci was removed from the default stack 2026-10-11")
	})

	It("should exist in PresetDescriptions", func() {
		desc, ok := constants.PresetDescriptions["format"]
		Expect(ok).To(BeTrue(), "format preset missing from PresetDescriptions")
		Expect(desc).ToNot(BeEmpty())
	})

	It("should be listed in ValidPresets", func() {
		Expect(constants.ValidPresets).To(ContainSubstring("format"))
	})

	It("should have formatters in canonical order matching FormatterOrder", func() {
		formatters := constants.PresetFormatters["format"]
		Expect(formatters).To(HaveLen(3))

		positions := make([]int, 0, len(formatters))
		for _, f := range formatters {
			positions = append(positions, slices.Index(constants.FormatterOrder, f))
		}

		Expect(positions).ToNot(ContainElement(-1),
			"format preset contains a formatter missing from FormatterOrder")
		Expect(slices.IsSorted(positions)).To(BeTrue(),
			"format preset formatters must appear in FormatterOrder-relative order")
	})
})

var _ = Describe("DefaultLinterSettings ToMap equivalence", func() {
	It("ireturn should produce correct allow list", func() {
		m := linterSettingsMap("ireturn")
		allow, ok := m["allow"]
		Expect(ok).To(BeTrue(), "ireturn settings missing allow key")
		allowSlice, ok := allow.([]any)
		Expect(ok).To(BeTrue(), "ireturn allow is not []any")
		Expect(allowSlice).To(HaveLen(5))
	})

	It("gocritic should produce disabled-checks list", func() {
		m := linterSettingsMap("gocritic")
		checks, ok := m["disabled-checks"]
		Expect(ok).To(BeTrue(), "gocritic settings missing disabled-checks key")
		checksSlice, ok := checks.([]any)
		Expect(ok).To(BeTrue(), "gocritic disabled-checks is not []any")
		Expect(checksSlice).To(ContainElement(ContainSubstring("ifElseChain")))
	})

	It("exhaustruct_v5 should produce ignore-patterns list", func() {
		m := linterSettingsMap("exhaustruct_v5")
		patterns, ok := m["ignore-patterns"]
		Expect(ok).To(BeTrue(), "exhaustruct_v5 settings missing ignore-patterns key")
		patternsSlice, ok := patterns.([]any)
		Expect(ok).To(BeTrue(), "exhaustruct_v5 ignore-patterns is not []any")
		Expect(patternsSlice).To(ContainElement(ContainSubstring("os/exec.Cmd")))
	})

	It("exhaustruct_v5 ignore-patterns should contain the canonical stdlib structs", func() {
		m := linterSettingsMap("exhaustruct_v5")
		patterns, ok := m["ignore-patterns"]
		Expect(ok).To(BeTrue(), "exhaustruct_v5 settings missing ignore-patterns key")
		patternsSlice, ok := patterns.([]any)
		Expect(ok).To(BeTrue(), "exhaustruct_v5 ignore-patterns is not []any")

		// These stdlib structs are the top sources of exhaustruct false-positive
		// noise (954 nolints across 146 configs in the cross-project audit).
		required := []string{
			"net/http.Client", "net/http.Server", "net/http.Request",
			"net/http.Response", "net/http.Transport", "net/http.Cookie",
			"net.TCPAddr", "net.Dialer", "log/slog.HandlerOptions",
			"sync.WaitGroup", "bytes.Buffer", "time.Ticker", "time.Timer",
			"os/exec.Cmd",
		}

		patternStrings := make([]string, 0, len(patternsSlice))
		for _, e := range patternsSlice {
			patternStrings = append(patternStrings, e.(string))
		}

		for _, r := range required {
			Expect(patternStrings).To(ContainElement(r),
				"exhaustruct_v5 ignore-patterns must contain canonical stdlib struct %q", r)
		}
	})

	It("revive should produce rules with disabled entries", func() {
		m := linterSettingsMap("revive")
		rules, ok := m["rules"]
		Expect(ok).To(BeTrue(), "revive settings missing rules key")
		rulesSlice, ok := rules.([]any)
		Expect(ok).To(BeTrue(), "revive rules is not []any")
		Expect(rulesSlice).To(HaveLen(2))
	})

	It("varnamelen should produce ignore flags and names", func() {
		m := linterSettingsMap("varnamelen")
		Expect(m["ignore-map-index-ok"]).To(BeTrue())
		Expect(m["ignore-type-assert-ok"]).To(BeTrue())
		names, ok := m["ignore-names"]
		Expect(ok).To(BeTrue())
		namesSlice, ok := names.([]any)
		Expect(ok).To(BeTrue())
		Expect(namesSlice).To(HaveLen(11))
	})

	It("gomoddirectives should produce replace-local bool", func() {
		m := linterSettingsMap("gomoddirectives")
		Expect(m["replace-local"]).To(BeTrue())
	})

	It("ginkgolinter should produce forbid flags", func() {
		m := linterSettingsMap("ginkgolinter")
		Expect(m["forbid-focus-container"]).To(BeTrue())
		Expect(m["forbid-spec-pollution"]).To(BeTrue())
	})

	It("testifylint should produce enable-all and disable list", func() {
		m := linterSettingsMap("testifylint")
		Expect(m["enable-all"]).To(BeTrue())
		disable, ok := m["disable"]
		Expect(ok).To(BeTrue())
		disableSlice, ok := disable.([]any)
		Expect(ok).To(BeTrue())
		Expect(disableSlice).To(ContainElement(ContainSubstring("go-require")))
	})

	It("makezero should produce always bool", func() {
		m := linterSettingsMap("makezero")
		Expect(m["always"]).To(BeTrue())
	})

	It("funlen should produce lines and statements thresholds", func() {
		m := linterSettingsMap("funlen")
		Expect(m["lines"]).To(Equal(200))
		Expect(m["statements"]).To(Equal(100))
	})

	It("gosec should produce excludes list", func() {
		m := linterSettingsMap("gosec")
		excludes, ok := m["excludes"]
		Expect(ok).To(BeTrue(), "gosec settings missing excludes key")
		excludesSlice, ok := excludes.([]any)
		Expect(ok).To(BeTrue(), "gosec excludes is not []any")
		Expect(excludesSlice).To(ContainElement(Equal("G304")))
		Expect(excludesSlice).To(ContainElement(Equal("G115")))
		// G104 (unhandled errors) is intentionally NOT excluded here —
		// errcheck's exclude-functions handles known-benign unhandled errors
		// surgically, while gosec catches security-relevant ones.
		Expect(excludesSlice).ToNot(ContainElement(Equal("G104")))
	})

	It("errcheck should produce exclude-functions list", func() {
		m := linterSettingsMap("errcheck")
		fns, ok := m["exclude-functions"]
		Expect(ok).To(BeTrue(), "errcheck settings missing exclude-functions key")
		fnsSlice, ok := fns.([]any)
		Expect(ok).To(BeTrue(), "errcheck exclude-functions is not []any")
		Expect(fnsSlice).To(ContainElement(Equal("(*os.File).Close")))
		Expect(fnsSlice).To(ContainElement(Equal("fmt.Fprintf")))
	})

	It("mnd should produce ignored-files list for test exclusion", func() {
		m := linterSettingsMap("mnd")
		files, ok := m["ignored-files"]
		Expect(ok).To(BeTrue(), "mnd settings missing ignored-files key")

		filesSlice, ok := files.([]any)
		Expect(ok).To(BeTrue(), "mnd ignored-files is not []any")
		Expect(filesSlice).To(ContainElement(Equal("_test\\.go")))
	})
})

var _ = Describe("DefaultExclusionRules", func() {
	It("should reference linters that exist in the policy table", func() {
		for _, rule := range constants.DefaultExclusionRules {
			for _, linter := range rule.Linters {
				_, exists := constants.PolicyRecord(types.LinterName(linter))
				Expect(exists).
					To(BeTrue(), "DefaultExclusionRules references %q which is missing from the policy table", linter)
			}
		}
	})

	It("should not include any disabled-tier linters", func() {
		for _, rule := range constants.DefaultExclusionRules {
			for _, linter := range rule.Linters {
				record, known := constants.PolicyRecord(types.LinterName(linter))
				Expect(known && record.Tier.Kind == types.TierDisabled).
					To(BeFalse(), "DefaultExclusionRules includes %q which is tool-disabled — excluded linters should be active", linter)
			}
		}
	})

	It("should not have duplicate linter names within a single rule", func() {
		for i, rule := range constants.DefaultExclusionRules {
			seen := make(map[string]bool, len(rule.Linters))

			for _, linter := range rule.Linters {
				Expect(seen[linter]).
					To(BeFalse(), "DefaultExclusionRules[%d] has duplicate linter %q", i, linter)
				seen[linter] = true
			}
		}
	})

	It("should have a non-empty path pattern for every rule", func() {
		for i, rule := range constants.DefaultExclusionRules {
			Expect(rule.Path).
				ToNot(BeEmpty(), "DefaultExclusionRules[%d] has an empty path pattern", i)
		}
	})

	It("should have at least one linter per rule", func() {
		for i, rule := range constants.DefaultExclusionRules {
			Expect(rule.Linters).
				ToNot(BeEmpty(), "DefaultExclusionRules[%d] has no linters", i)
		}
	})
})

var _ = Describe("Pragmatic-noise tier", func() {
	It("should have a non-empty rationale for every member", func() {
		for _, linter := range constants.LintersInTier(types.TierPragmaticNoise) {
			record, known := constants.PolicyRecord(linter)
			Expect(known).To(BeTrue(), "noise linter %q missing from the table", linter)
			Expect(record.Tier.Rationale).
				ToNot(BeEmpty(), "pragmatic-noise linter %q has an empty tier rationale", linter)
		}
	})

	It("should be disjoint from the disabled tier", func() {
		for _, linter := range constants.LintersInTier(types.TierPragmaticNoise) {
			record, known := constants.PolicyRecord(linter)
			Expect(known && record.Tier.Kind == types.TierDisabled).
				To(BeFalse(), "pragmatic-noise linter %q is also disabled-tier — already permanently disabled", linter)
		}
	})
})

var _ = Describe("Core formatter stack alignment", func() {
	It("should contain exactly the validated formatter stack", func() {
		Expect(constants.CoreFormattersInOrder()).
			To(Equal([]types.FormatterName{"goimports", "gofumpt", "golines"}))
	})

	It("should match the house preset formatters", func() {
		houseFormatters := constants.PresetFormatters["house"]

		coreFormatters := constants.CoreFormattersInOrder()
		coreSet := make(map[types.FormatterName]struct{}, len(coreFormatters))
		for _, f := range coreFormatters {
			coreSet[f] = struct{}{}
		}

		for _, f := range houseFormatters {
			_, ok := coreSet[f]
			Expect(ok).
				To(BeTrue(), "house preset formatter %q is missing from the core stack", f)
		}

		Expect(houseFormatters).
			To(HaveLen(len(coreFormatters)),
				"house preset and the core stack should have the same number of formatters")
	})
})

// Docs-integrity derivation (M34): the "Deprecated Linter Migration Map"
// table in FEATURES.md must stay in lockstep with the policy table's
// supersession records — one row per record carrying a supersession reason,
// no extra rows, and the replacement column must match SupersededBy
// ("—" for entries with no direct replacement).
var _ = Describe("FEATURES deprecation migration map", func() {
	parseMigrationTable := func() map[string]string {
		content, err := os.ReadFile(filepath.Join("..", "..", "FEATURES.md"))
		Expect(err).ToNot(HaveOccurred(), "FEATURES.md must be readable from pkg/constants")

		lines := strings.Split(string(content), "\n")
		start := -1

		for i, line := range lines {
			if strings.HasPrefix(line, "### Deprecated Linter Migration Map") {
				start = i

				break
			}
		}

		Expect(start).To(BeNumerically(">=", 0),
			"FEATURES.md is missing the 'Deprecated Linter Migration Map' heading")

		rows := map[string]string{}

		for _, line := range lines[start+1:] {
			if strings.HasPrefix(line, "## ") {
				break
			}

			if !strings.HasPrefix(line, "| ") {
				continue
			}

			cells := strings.Split(strings.Trim(line, "| "), "|")
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}

			if cells[0] == "Deprecated" || strings.Contains(cells[0], "--") {
				continue
			}

			replacement := cells[1]
			if replacement == "—" {
				replacement = ""
			}

			rows[cells[0]] = replacement
		}

		return rows
	}

	It("has exactly one row per supersession record", func() {
		rows := parseMigrationTable()

		Expect(rows).To(HaveLen(len(deprecatedLinterRecords())),
			"FEATURES migration table and the policy table must cover the same number of deprecations")
	})

	It("agrees with the policy table on every replacement", func() {
		rows := parseMigrationTable()

		for _, record := range deprecatedLinterRecords() {
			rowReplacement, mapped := rows[string(record.Name)]
			Expect(mapped).
				To(BeTrue(), "deprecated linter %q is missing a row in the FEATURES migration table", record.Name)
			Expect(rowReplacement).
				To(Equal(string(record.SupersededBy)),
					"FEATURES migration table row %q disagrees with the policy table's SupersededBy", record.Name)
		}
	})

	It("has no rows outside the policy table's supersession records", func() {
		rows := parseMigrationTable()

		for deprecated := range rows {
			record, known := constants.PolicyRecord(types.LinterName(deprecated))
			Expect(known && record.SupersessionReason != "").
				To(BeTrue(), "FEATURES migration table row %q has no supersession record in the policy table", deprecated)
		}
	})
})
