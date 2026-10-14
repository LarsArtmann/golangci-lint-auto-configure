package types

import (
	"fmt"
	"time"
)

// TierKind is the closed union of linter-management decisions. A linter lives
// in exactly one tier; the zero value (Recommendable) is the default for
// ordinary linters, so most table records never mention Tier at all.
type TierKind uint8

const (
	// TierRecommendable is the default tier: the linter may be recommended and
	// auto-enabled subject to its Gate, priority, and validity interval.
	TierRecommendable TierKind = iota
	// TierDisabled is forcibly moved to linters.disable on every run, never
	// recommended, and excluded from the recommendation pipeline entirely.
	TierDisabled
	// TierNeverAutoEnable is never added to an enable list by the tool, but
	// never stripped either: manual enables are respected and still receive
	// safe default settings and test-file exclusions.
	TierNeverAutoEnable
	// TierPragmaticNoise is enabled by default and dropped from the dynamic
	// enable set only when the user passes --pragmatic.
	TierPragmaticNoise
)

// String renders the tier kind for logs and audit reasons.
func (t TierKind) String() string {
	switch t {
	case TierRecommendable:
		return "recommendable"
	case TierDisabled:
		return "disabled"
	case TierNeverAutoEnable:
		return "never-auto-enable"
	case TierPragmaticNoise:
		return "pragmatic-noise"
	default:
		return "unknown-tier"
	}
}

// Tier is the management decision for one linter, carrying the per-decision
// rationale (the reason a linter sits in this tier — distinct from the
// record's Rationale, which says why the linter is worth recommending).
type Tier struct {
	Kind      TierKind
	Rationale string
}

// GateKind is the closed union of recommendation gates: the conditions under
// which a Recommendable linter may actually be recommended.
type GateKind uint8

const (
	// GateRecommendAlways means the linter is recommended whenever it is not
	// already enabled (subject to validity intervals).
	GateRecommendAlways GateKind = iota
	// GateRequiresTech means the linter is only recommended when the project
	// actually uses the named technology (Gate.Tech, the analyzer's
	// hasTechnology key).
	GateRequiresTech
)

// Gate is the environment-dependent recommendation gate for one linter.
// Relations between linters (redundancy, supersession chains) are NOT gates —
// they are separate relations evaluated from the table.
type Gate struct {
	Kind GateKind
	// Tech is the technology key checked by the analyzer's hasTechnology.
	// Only meaningful when Kind == GateRequiresTech.
	Tech string
}

// RequiresTech builds a GateRequiresTech gate for the given technology key.
func RequiresTech(tech string) Gate {
	return Gate{Kind: GateRequiresTech, Tech: tech}
}

// VersionedSettings is one settings block a linter should receive, with the
// golangci-lint version it became valid in. Since the zero value means "all
// versions", the curated minimum of the tool DERIVES from the max Since over
// the settings that get rendered — no hand-maintained constant.
type VersionedSettings struct {
	// Settings is the typed settings value (SettingsConverter.ToMap renders it).
	Settings SettingsConverter
	// Since is the golangci-lint version this settings block is valid from.
	// Zero means the block is valid for every version.
	Since GolangciVersion
}

// SettingsConverter converts typed settings to map[string]any for config
// injection. Implemented by the curated linter and formatter settings structs
// (constants package) to provide compile-time safety for default setting keys
// and value types.
type SettingsConverter interface {
	ToMap() map[string]any
}

// Evidence is the measured friction data behind tier decisions. Tiers DERIVE
// from this (a linter's tier payload cites the measurement, and --pragmatic
// becomes a threshold over Ratio), replacing opinions with numbers.
type Evidence struct {
	// Nolints is the count of //nolint suppressions observed for the linter.
	Nolints int
	// Configs is the count of configs in the measurement corpus.
	Configs int
	// Ratio is Nolints per config — the friction metric.
	Ratio float64
	// AsOf is when the measurement was taken.
	AsOf time.Time
}

// LinterPolicy is the single source of truth about one linter. Every fact the
// tool acts on lives here: why it is recommended (Rationale), when it exists
// (ValidTool), which language it speaks about (Lang), what replaced it
// (SupersededBy), how annoying it is (Friction), what settings it needs
// (Settings), how it is managed (Tier), and when it may be recommended (Gate).
// Adding a linter = adding one record.
//
// Invariants (enforced by the table authoring specs, and by construction when
// records are built through ValidatedPolicy):
//   - Tier Rationale is required for every non-Recommendable tier.
//   - SupersededBy must reference a linter that exists in the table.
//   - A disabled-tier linter carries no Priority (it is never recommended).
//   - GateRequiresTech requires a non-empty Tech key.
type LinterPolicy struct {
	// Name is the golangci-lint linter identifier.
	Name LinterName
	// ValidTool is the tool-version interval where the linter exists and is
	// loadable. Deprecation is a Max bound; the zero interval means "all
	// versions". Emitted findings when violated come from this bound, not a
	// separate deprecation map.
	ValidTool Interval[GolangciVersion]
	// Lang is the Go-version interval where the linter is linguistically
	// relevant (e.g. copyloopvar from Go 1.22). Zero means "always".
	Lang Interval[GoVersion]
	// SupersededBy names the replacement linter; zero means none.
	SupersededBy LinterName
	// SupersessionReason explains the replacement (verbatim from the old
	// deprecation map).
	SupersessionReason string
	// SupersessionSince is the golangci-lint version where the replacement
	// became available. Zero means "no stated availability bound" (the
	// replacement is treated as available whenever the tool runs).
	SupersessionSince GolangciVersion
	// Priority is the recommendation priority. Optional priority is the zero
	// value — records omit it unless the linter is worth calling out.
	Priority LinterPriority
	// Rationale is the one human-readable reason the table owns for
	// recommending this linter (shown in reports and audit reasons).
	Rationale string
	// Friction is the measured evidence; nil for unmeasured linters.
	Friction *Evidence
	// Settings are the safe-default settings blocks injected when the linter
	// is enabled and unconfigured.
	Settings []VersionedSettings
	// Tier is the management decision; zero = Recommendable with no
	// rationale needed.
	Tier Tier
	// Gate is the recommendation gate; zero = RecommendAlways.
	Gate Gate
}

// Validate checks the record's structural invariants. It is called on every
// table entry at package init; a violation is a programming error in curated
// data and panics there.
func (p LinterPolicy) Validate() error {
	if p.Name == "" {
		return errPolicy("policy record without a name")
	}

	if p.Tier.Kind != TierRecommendable && p.Tier.Rationale == "" {
		return errPolicy("linter %s: tier %s requires a rationale", p.Name, p.Tier.Kind)
	}

	if p.SupersededBy != "" && p.SupersessionReason == "" {
		return errPolicy("linter %s: superseded by %s requires a reason", p.Name, p.SupersededBy)
	}

	if p.SupersededBy == "" && p.SupersessionReason != "" {
		return errPolicy("linter %s: supersession reason without SupersededBy", p.Name)
	}

	if p.Gate.Kind == GateRequiresTech && p.Gate.Tech == "" {
		return errPolicy("linter %s: RequiresTech gate requires a technology key", p.Name)
	}

	if p.Tier.Kind == TierDisabled && p.Priority != LinterPriorityOptional {
		return errPolicy("linter %s: disabled tier must not carry a recommendation priority", p.Name)
	}

	if !p.ValidTool.IsEmpty() {
		p.ValidTool.MustValidate()
	}

	if !p.Lang.IsEmpty() {
		p.Lang.MustValidate()
	}

	return nil
}

func errPolicy(format string, args ...any) error { return fmt.Errorf(format, args...) }
