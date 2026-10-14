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
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const outputPath = "pkg/constants/linter_policy_table.go"

// curatedData holds the hand-curated facts that live only in this generator:
// stated deprecation versions, measured friction ratios, and language
// relevance bounds.
type curatedData struct {
	validToolMax    map[string]string
	friction        map[string]float64
	exhaustructAsOf time.Time
	langMin         map[string]string
}

// loadCurated returns the curated facts.
//
// validToolMax encodes the deprecation version where it is a stated fact in
// the legacy reasons ("deprecated since golangci-lint vX"). Linters removed
// in v2 have no precise last-good version in the data; their interval stays
// unbounded and supersession drives behavior.
//
// friction encodes the measured ratios from the cross-project nolint audit.
// AsOf is unknown for all but exhaustruct; the zero time means "measurement
// date not recorded". exhaustruct is the one measurement whose raw counts
// survive in the curated docs: 954 nolints across 146 configs.
func loadCurated() curatedData {
	return curatedData{
		validToolMax: map[string]string{
			"wsl":         "v2.2.0",
			"gomodguard":  "v2.12.0",
			"exhaustruct": "v2.13.0",
		},
		friction: map[string]float64{
			"exhaustruct_v5":   6.5,  //nolint:mnd // measured friction ratio (curated audit data)
			"gochecknoglobals": 5.1,  //nolint:mnd // measured friction ratio (curated audit data)
			"wrapcheck":        1.8,  //nolint:mnd // measured friction ratio (curated audit data)
			"ireturn":          1.3,  //nolint:mnd // measured friction ratio (curated audit data)
			"funlen":           0.83, //nolint:mnd // measured friction ratio (curated audit data)
		},
		exhaustructAsOf: time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC),
		langMin: map[string]string{
			"copyloopvar": "1.22",
		},
	}
}

func main() {
	records := collect()

	write(render(records))

	fmt.Fprintln(os.Stdout, "wrote", outputPath, "with", len(records), "policy records")
}

func write(source string) {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		const debugPath = "/tmp/linter_policy_table.go.raw"

		if writeErr := os.WriteFile(debugPath, []byte(source), 0o600); writeErr != nil {
			fmt.Fprintln(os.Stderr, "write raw debug source:", writeErr)
		}

		fmt.Fprintf(os.Stderr, "format generated source: %v (raw source at %s)\n", err, debugPath)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, formatted, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
}

func collect() []types.LinterPolicy {
	legacy := readGoldenMaps()
	curated := loadCurated()

	sorted := collectNames(legacy)

	records := make([]types.LinterPolicy, 0, len(sorted))

	for _, name := range sorted {
		records = append(records, build(types.LinterName(name), legacy, curated))
	}

	return records
}

func collectNames(legacy goldenMaps) []string {
	names := map[string]bool{}

	for _, source := range []map[string]string{
		legacy.reasons,
		legacy.disabled,
		legacy.neverAutoEnable,
		legacy.pragmaticNoise,
		legacy.minVersions,
		legacy.projectSpecific,
	} {
		for name := range source {
			names[name] = true
		}
	}

	for name := range legacy.priorities {
		names[name] = true
	}

	for name := range legacy.deprecated {
		names[name] = true
	}

	for name := range constants.DefaultLinterSettings {
		names[string(name)] = true
	}

	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}

	sort.Strings(sorted)

	return sorted
}

type goldenMaps struct {
	priorities      map[string]int
	reasons         map[string]string
	deprecated      map[string]goldenReplacement
	disabled        map[string]string
	neverAutoEnable map[string]string
	pragmaticNoise  map[string]string
	minVersions     map[string]string
	projectSpecific map[string]string
}

type goldenReplacement struct {
	Replacement string
	Reason      string
	MinVersion  string
}

// readGoldenMaps parses the frozen views goldens: the byte-exact pre-table
// truth. Regenerating from the package vars would be circular (they are
// views over this table) and has already proven lossy once.
func readGoldenMaps() goldenMaps {
	read := func(file string, into any) {
		data, err := os.ReadFile("pkg/constants/testdata/views-golden/" + file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read golden", file, err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, into); err != nil {
			fmt.Fprintln(os.Stderr, "parse golden", file, err)
			os.Exit(1)
		}
	}

	var legacy goldenMaps

	read("linter_priorities.json", &legacy.priorities)
	read("linter_reasons.json", &legacy.reasons)
	read("deprecated_linters.json", &legacy.deprecated)
	read("disabled_linters.json", &legacy.disabled)
	read("never_auto_enable_linters.json", &legacy.neverAutoEnable)
	read("pragmatic_noise_linters.json", &legacy.pragmaticNoise)
	read("linter_min_versions.json", &legacy.minVersions)
	read("project_specific_linters.json", &legacy.projectSpecific)

	return legacy
}

func build(name types.LinterName, legacy goldenMaps, curated curatedData) types.LinterPolicy {
	record := buildIdentity(name, legacy)
	buildValidity(&record, legacy, curated)
	buildTier(&record, legacy)
	buildEvidence(&record, curated)
	buildSettings(&record)

	return record
}

func buildIdentity(name types.LinterName, legacy goldenMaps) types.LinterPolicy {
	record := types.LinterPolicy{Name: name}

	// Priorities come from the frozen golden JSON (the pre-table truth), NOT
	// from constants.LinterPriorities: since the var swap that map is a view
	// over THIS table, and regenerating from it would be circular.
	if raw, ok := legacy.priorities[string(name)]; ok {
		record.Priority = types.LinterPriority(raw)
	} else {
		// Explicit Optional: the zero value of LinterPriority is Critical, so
		// "no recommendation callout" must be stated, never defaulted.
		record.Priority = types.LinterPriorityOptional
	}

	record.Rationale = legacy.reasons[string(name)]

	return record
}

func buildValidity(record *types.LinterPolicy, legacy goldenMaps, curated curatedData) {
	if replacement, deprecated := legacy.deprecated[string(record.Name)]; deprecated {
		record.SupersededBy = types.LinterName(replacement.Replacement)
		record.SupersessionReason = replacement.Reason
		record.SupersessionSince = parseVersion(replacement.MinVersion)

		if maxVer, hasMax := curated.validToolMax[string(record.Name)]; hasMax {
			record.ValidTool = types.Interval[types.GolangciVersion]{
				Max: types.MustParseGolangciVersion(maxVer),
			}
		}
	}

	if minVer, hasMin := legacy.minVersions[string(record.Name)]; hasMin {
		record.ValidTool.Min = parseVersion(minVer)
	}

	if lang, hasLang := curated.langMin[string(record.Name)]; hasLang {
		record.Lang = types.Interval[types.GoVersion]{
			Min: types.MustParseGoVersion(lang),
		}
	}

	if tech, isSpecific := legacy.projectSpecific[string(record.Name)]; isSpecific {
		record.Gate = types.RequiresTech(tech)
	}
}

func buildTier(record *types.LinterPolicy, legacy goldenMaps) {
	if reason, disabled := legacy.disabled[string(record.Name)]; disabled {
		record.Tier = types.Tier{Kind: types.TierDisabled, Rationale: reason}
		record.Priority = types.LinterPriorityOptional
		record.Rationale = ""

		return
	}

	if reason, never := legacy.neverAutoEnable[string(record.Name)]; never {
		record.Tier = types.Tier{Kind: types.TierNeverAutoEnable, Rationale: reason}

		return
	}

	if reason, noise := legacy.pragmaticNoise[string(record.Name)]; noise {
		record.Tier = types.Tier{Kind: types.TierPragmaticNoise, Rationale: reason}
	}
}

func buildEvidence(record *types.LinterPolicy, curated curatedData) {
	ratio, measured := curated.friction[string(record.Name)]
	if !measured {
		return
	}

	evidence := &types.Evidence{Ratio: ratio}

	if record.Name == "exhaustruct_v5" {
		evidence.Nolints = 954
		evidence.Configs = 146
		evidence.AsOf = curated.exhaustructAsOf
	}

	record.Friction = evidence
}

func buildSettings(record *types.LinterPolicy) {
	settings, hasSettings := constants.DefaultLinterSettings[record.Name]
	if !hasSettings {
		return
	}

	since := settingsSince(record.Name)

	record.Settings = []types.VersionedSettings{{Settings: settings, Since: since}}
}

func settingsSince(name types.LinterName) types.GolangciVersion {
	switch name {
	case "goconst":
		return types.MustParseGolangciVersion("v2.12.0")
	case "exhaustruct_v5":
		return types.MustParseGolangciVersion("v2.13.0")
	default:
		return types.GolangciVersion{}
	}
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
		renderRecord(&b, record)
	}

	b.WriteString("}\n")

	return b.String()
}

func renderRecord(b *strings.Builder, record types.LinterPolicy) {
	b.WriteString("\t{\n")

	fmt.Fprintf(b, "\t\tName: %q,\n", string(record.Name))

	b.WriteString(intervalField(&record))
	b.WriteString(langField(&record))
	b.WriteString(supersessionFields(&record))
	b.WriteString(priorityField(&record))

	if record.Rationale != "" {
		fmt.Fprintf(b, "\t\tRationale: %q,\n", record.Rationale)
	}

	if record.Friction != nil {
		fmt.Fprintf(b, "\t\tFriction: &types.Evidence{Ratio: %#v, Nolints: %d, Configs: %d, AsOf: %s},\n",
			record.Friction.Ratio, record.Friction.Nolints, record.Friction.Configs, timeLit(record.Friction.AsOf))
	}

	for _, entry := range record.Settings {
		fmt.Fprintf(b, "\t\tSettings: []types.VersionedSettings{{Settings: %s, Since: %s}},\n",
			structLiteral(reflect.ValueOf(entry.Settings)), sinceLit(entry.Since))
	}

	if record.Tier.Kind != types.TierRecommendable {
		fmt.Fprintf(b, "\t\tTier: types.Tier{Kind: types.%s, Rationale: %q},\n",
			tierIdent(record.Tier.Kind), record.Tier.Rationale)
	}

	if record.Gate.Kind == types.GateRequiresTech {
		fmt.Fprintf(b, "\t\tGate: types.RequiresTech(%q),\n", record.Gate.Tech)
	}

	b.WriteString("\t},\n")
}

func supersessionFields(record *types.LinterPolicy) string {
	var b strings.Builder

	if record.SupersededBy != "" {
		fmt.Fprintf(&b, "\t\tSupersededBy: %q,\n", string(record.SupersededBy))
	}

	if record.SupersessionReason != "" {
		fmt.Fprintf(&b, "\t\tSupersessionReason: %q,\n", record.SupersessionReason)
	}

	if record.SupersessionSince.String() != "" {
		fmt.Fprintf(&b, "\t\tSupersessionSince: types.MustParseGolangciVersion(%q),\n", record.SupersessionSince.String())
	}

	return b.String()
}

func priorityField(record *types.LinterPolicy) string {
	if record.Priority == types.LinterPriorityOptional {
		return "\t\tPriority: types.LinterPriorityOptional,\n"
	}

	return fmt.Sprintf("\t\tPriority: %s,\n", priorityIdent(record.Priority))
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

func priorityIdent(priority types.LinterPriority) string {
	identifiers := map[types.LinterPriority]string{
		types.LinterPriorityCritical: "types.LinterPriorityCritical",
		types.LinterPriorityHigh:     "types.LinterPriorityHigh",
		types.LinterPriorityMedium:   "types.LinterPriorityMedium",
		types.LinterPriorityOptional: "types.LinterPriorityOptional",
	}

	return identifiers[priority]
}

func tierIdent(kind types.TierKind) string {
	identifiers := map[types.TierKind]string{
		types.TierRecommendable:   "TierRecommendable",
		types.TierDisabled:        "TierDisabled",
		types.TierNeverAutoEnable: "TierNeverAutoEnable",
		types.TierPragmaticNoise:  "TierPragmaticNoise",
	}

	return identifiers[kind]
}

func sinceLit(version types.GolangciVersion) string {
	if version.IsZero() {
		return "types.GolangciVersion{}"
	}

	return fmt.Sprintf("types.MustParseGolangciVersion(%q)", version.String())
}

func timeLit(moment time.Time) string {
	if moment.IsZero() {
		return "time.Time{}"
	}

	return fmt.Sprintf("time.Date(%d, %d, %d, 0, 0, 0, 0, time.UTC)", moment.Year(), moment.Month(), moment.Day())
}

// structLiteral renders a settings struct value as a composite literal.
func structLiteral(value reflect.Value) string {
	switch value.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", value.String())
	case reflect.Int:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Slice:
		return sliceLiteral(value)
	case reflect.Struct:
		return structFields(value)
	default:
		panic(fmt.Sprintf("gen-policy-table: unsupported settings field kind %s", value.Kind()))
	}
}

func sliceLiteral(value reflect.Value) string {
	elements := make([]string, 0, value.Len())

	for i := range value.Len() {
		elements = append(elements, structLiteral(value.Index(i)))
	}

	return fmt.Sprintf("[]%s{%s}", typeName(value.Type().Elem()), strings.Join(elements, ", "))
}

func structFields(value reflect.Value) string {
	fields := make([]string, 0, value.Type().NumField())

	for i := range value.Type().NumField() {
		field := value.Type().Field(i)
		if !ast.IsExported(field.Name) {
			continue
		}

		fields = append(fields, fmt.Sprintf("%s: %s", field.Name, structLiteral(value.Field(i))))
	}

	return fmt.Sprintf("%s{%s}", typeName(value.Type()), strings.Join(fields, ", "))
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
