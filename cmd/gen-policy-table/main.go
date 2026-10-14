// Command gen-policy-table transcribes the legacy hand-maintained linter maps
// (LinterPriorities, LinterReasons, DeprecatedLinters, tier maps,
// LinterMinVersions, ProjectSpecificLinters, DefaultLinterSettings) into the
// declarative policy table at pkg/constants/linter_policy_table.go.
//
// It ran once to bootstrap the table (2026-10-14); the table is the
// hand-maintained source of truth afterwards. Keep the generator around for
// provenance and re-runnability during the migration window.
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const outputPath = "pkg/constants/linter_policy_table.go"

// validToolMax encodes the deprecation version where it is a stated fact in
// the legacy reasons ("deprecated since golangci-lint vX"). Linters removed
// in v2 have no precise last-good version in the data; their interval stays
// unbounded and supersession drives behavior.
var validToolMax = map[string]string{
	"wsl":         "v2.2.0",
	"gomodguard":  "v2.12.0",
	"exhaustruct": "v2.13.0",
}

// friction encodes the measured ratios documented in the tier comments
// (cross-project nolint audit). AsOf is unknown for all but exhaustruct;
// the zero time means "measurement date not recorded".
var friction = map[string]float64{
	"exhaustruct_v5":   6.5,
	"gochecknoglobals": 5.1,
	"wrapcheck":        1.8,
	"ireturn":          1.3,
	"funlen":           0.83,
}

// exhaustructEvidence is the one measurement whose raw counts survive in the
// curated docs: 954 nolints across 146 configs.
var exhaustructAsOf = time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)

// langMin encodes linguistic-relevance lower bounds that are language facts.
var langMin = map[string]string{
	"copyloopvar": "1.22",
}

func main() {
	records := collect()

	src := render(records)

	formatted, err := format.Source([]byte(src))
	if err != nil {
		const debugPath = "/tmp/linter_policy_table.go.raw"

		if writeErr := os.WriteFile(debugPath, []byte(src), 0o644); writeErr != nil {
			fmt.Fprintln(os.Stderr, "write raw debug source:", writeErr)
		}

		fmt.Fprintf(os.Stderr, "format generated source: %v (raw source at %s)\n", err, debugPath)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, formatted, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s with %d policy records\n", outputPath, len(records))
}

func collect() []types.LinterPolicy {
	names := map[types.LinterName]bool{}

	for name := range constants.LinterPriorities {
		names[name] = true
	}

	for name := range constants.LinterReasons {
		names[name] = true
	}

	for name := range constants.DeprecatedLinters {
		names[name] = true
	}

	for name := range constants.DisabledLinters {
		names[name] = true
	}

	for name := range constants.NeverAutoEnableLinters {
		names[name] = true
	}

	for name := range constants.PragmaticNoiseLinters {
		names[name] = true
	}

	for name := range constants.DefaultLinterSettings {
		names[name] = true
	}

	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, string(name))
	}

	sort.Strings(sorted)

	records := make([]types.LinterPolicy, 0, len(sorted))

	for _, name := range sorted {
		records = append(records, build(types.LinterName(name)))
	}

	return records
}

func build(name types.LinterName) types.LinterPolicy {
	record := types.LinterPolicy{Name: name}

	if priority, ok := constants.LinterPriorities[name]; ok && priority != types.LinterPriorityOptional {
		record.Priority = priority
	} else {
		// Explicit Optional: the zero value of LinterPriority is Critical, so
		// "no recommendation callout" must be stated, never defaulted.
		record.Priority = types.LinterPriorityOptional
	}

	record.Rationale = constants.LinterReasons[name]

	if replacement, deprecated := constants.DeprecatedLinters[name]; deprecated {
		record.SupersededBy = replacement.Replacement
		record.SupersessionReason = replacement.Reason
		record.SupersessionSince = parseVersion(replacement.MinVersion)

		if maxVer, hasMax := validToolMax[string(name)]; hasMax {
			record.ValidTool = types.Interval[types.GolangciVersion]{
				Max: types.MustParseGolangciVersion(maxVer),
			}
		}
	}

	if minVer, hasMin := constants.LinterMinVersions[name]; hasMin {
		record.ValidTool.Min = parseVersion(minVer)
	}

	if lang, hasLang := langMin[string(name)]; hasLang {
		record.Lang = types.Interval[types.GoVersion]{
			Min: types.MustParseGoVersion(lang),
		}
	}

	if tech, isSpecific := constants.ProjectSpecificLinters[name]; isSpecific {
		record.Gate = types.RequiresTech(tech)
	}

	if reason, disabled := constants.DisabledLinters[name]; disabled {
		record.Tier = types.Tier{Kind: types.TierDisabled, Rationale: reason}
		record.Priority = types.LinterPriorityOptional
		record.Rationale = ""
	} else if reason, never := constants.NeverAutoEnableLinters[name]; never {
		record.Tier = types.Tier{Kind: types.TierNeverAutoEnable, Rationale: reason}
	} else if reason, noise := constants.PragmaticNoiseLinters[name]; noise {
		record.Tier = types.Tier{Kind: types.TierPragmaticNoise, Rationale: reason}
	}

	if ratio, measured := friction[string(name)]; measured {
		evidence := &types.Evidence{Ratio: ratio}
		if name == "exhaustruct_v5" {
			evidence.Nolints = 954
			evidence.Configs = 146
			evidence.AsOf = exhaustructAsOf
		}

		record.Friction = evidence
	}

	if settings, hasSettings := constants.DefaultLinterSettings[name]; hasSettings {
		since := types.GolangciVersion{}
		if name == "goconst" {
			since = types.MustParseGolangciVersion("v2.12.0")
		}

		if name == "exhaustruct_v5" {
			since = types.MustParseGolangciVersion("v2.13.0")
		}

		record.Settings = []types.VersionedSettings{{Settings: settings, Since: since}}
	}

	return record
}

func parseVersion(raw string) types.GolangciVersion {
	if raw == "" {
		return types.GolangciVersion{}
	}

	return types.MustParseGolangciVersion(raw)
}

func render(records []types.LinterPolicy) string {
	var b strings.Builder

	b.WriteString(header)

	for _, record := range records {
		b.WriteString("\t{\n")

		fmt.Fprintf(&b, "\t\tName: %q,\n", string(record.Name))

		b.WriteString(intervalField(&record))
		b.WriteString(langField(&record))

		if record.SupersededBy != "" {
			fmt.Fprintf(&b, "\t\tSupersededBy: %q,\n", string(record.SupersededBy))
			fmt.Fprintf(&b, "\t\tSupersessionReason: %q,\n", record.SupersessionReason)

			if record.SupersessionSince.String() != "" {
				fmt.Fprintf(&b, "\t\tSupersessionSince: types.MustParseGolangciVersion(%q),\n", record.SupersessionSince.String())
			}
		}

		if record.Priority != types.LinterPriorityOptional {
			fmt.Fprintf(&b, "\t\tPriority: %s,\n", priorityIdent(record.Priority))
		}

		if record.Rationale != "" {
			fmt.Fprintf(&b, "\t\tRationale: %q,\n", record.Rationale)
		}

		if record.Friction != nil {
			fmt.Fprintf(&b, "\t\tFriction: &types.Evidence{Ratio: %s, Nolints: %d, Configs: %d, AsOf: %s},\n",
				floatLit(record.Friction.Ratio), record.Friction.Nolints, record.Friction.Configs, timeLit(record.Friction.AsOf))
		}

		for _, entry := range record.Settings {
			fmt.Fprintf(&b, "\t\tSettings: []types.VersionedSettings{{Settings: %s, Since: %s}},\n",
				structLiteral(reflect.ValueOf(entry.Settings), 3), sinceLit(entry.Since))
		}

		if record.Tier.Kind != types.TierRecommendable {
			fmt.Fprintf(&b, "\t\tTier: types.Tier{Kind: types.%s, Rationale: %q},\n",
				tierIdent(record.Tier.Kind), record.Tier.Rationale)
		}

		if record.Gate.Kind == types.GateRequiresTech {
			fmt.Fprintf(&b, "\t\tGate: types.RequiresTech(%q),\n", record.Gate.Tech)
		}

		b.WriteString("\t},\n")
	}

	b.WriteString("}\n")

	return b.String()
}

func intervalField(record *types.LinterPolicy) string {
	minStr := record.ValidTool.Min.String()
	maxStr := record.ValidTool.Max.String()

	switch {
	case minStr != "" && maxStr != "":
		return fmt.Sprintf("\t\tValidTool: types.NewInterval(types.MustParseGolangciVersion(%q), types.MustParseGolangciVersion(%q)),\n",
			minStr, maxStr)
	case minStr != "":
		return fmt.Sprintf("\t\tValidTool: types.Interval[types.GolangciVersion]{Min: types.MustParseGolangciVersion(%q)},\n", minStr)
	case maxStr != "":
		return fmt.Sprintf("\t\tValidTool: types.Interval[types.GolangciVersion]{Max: types.MustParseGolangciVersion(%q)},\n", maxStr)
	default:
		return ""
	}
}

func langField(record *types.LinterPolicy) string {
	if record.Lang.Min.IsZero() {
		return ""
	}

	return fmt.Sprintf("\t\tLang: types.Interval[types.GoVersion]{Min: types.MustParseGoVersion(%q)},\n", record.Lang.Min.String())
}

func priorityIdent(p types.LinterPriority) string {
	return map[types.LinterPriority]string{
		types.LinterPriorityCritical: "types.LinterPriorityCritical",
		types.LinterPriorityHigh:     "types.LinterPriorityHigh",
		types.LinterPriorityMedium:   "types.LinterPriorityMedium",
		types.LinterPriorityOptional: "types.LinterPriorityOptional",
	}[p]
}

func tierIdent(k types.TierKind) string {
	return map[types.TierKind]string{
		types.TierRecommendable:   "TierRecommendable",
		types.TierDisabled:        "TierDisabled",
		types.TierNeverAutoEnable: "TierNeverAutoEnable",
		types.TierPragmaticNoise:  "TierPragmaticNoise",
	}[k]
}

func floatLit(f float64) string {
	return fmt.Sprintf("%#v", f)
}

func sinceLit(v types.GolangciVersion) string {
	if v.IsZero() {
		return "types.GolangciVersion{}"
	}

	return fmt.Sprintf("types.MustParseGolangciVersion(%q)", v.String())
}

func timeLit(t time.Time) string {
	if t.IsZero() {
		return "time.Time{}"
	}

	return fmt.Sprintf("time.Date(%d, %d, %d, 0, 0, 0, 0, time.UTC)", t.Year(), t.Month(), t.Day())
}

// structLiteral renders a settings struct value as a composite literal.
func structLiteral(value reflect.Value, depth int) string {
	t := value.Type()

	switch t.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", value.String())
	case reflect.Int:
		return fmt.Sprintf("%d", value.Int())
	case reflect.Bool:
		return fmt.Sprintf("%v", value.Bool())
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Struct {
			parts := make([]string, 0, value.Len())
			for i := 0; i < value.Len(); i++ {
				parts = append(parts, structLiteral(value.Index(i), depth+1))
			}

			return fmt.Sprintf("[]%s{%s}", typeName(t.Elem()), strings.Join(parts, ", "))
		}

		parts := make([]string, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			parts = append(parts, structLiteral(value.Index(i), depth+1))
		}

		return fmt.Sprintf("[]%s{%s}", typeName(t.Elem()), strings.Join(parts, ", "))
	case reflect.Struct:
		fields := make([]string, 0, t.NumField())

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !ast.IsExported(field.Name) {
				continue
			}

			fields = append(fields, fmt.Sprintf("%s: %s", field.Name, structLiteral(value.Field(i), depth+1)))
		}

		return fmt.Sprintf("%s{%s}", typeName(t), strings.Join(fields, ", "))
	default:
		panic(fmt.Sprintf("gen-policy-table: unsupported settings field kind %s", t.Kind()))
	}
}

func typeName(t reflect.Type) string {
	if t.Name() == "" {
		return t.String()
	}

	return t.Name()
}

const header = `// Code generated by cmd/gen-policy-table from the legacy linter maps
// (2026-10-14); curated by hand afterwards. This table is the single source
// of truth about linter management: priority, rationale, validity, tier,
// gate, supersession, friction, and safe-default settings. The legacy maps
// in this package are views computed from this table.
package constants

import (
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// LinterPolicies is the policy table: one record per known linter.
var LinterPolicies = []types.LinterPolicy{
`
