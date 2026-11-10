package constants

import (
	"crypto/sha256"
	"encoding/json/v2"
)

// Managed-settings provenance: when the tool injects a curated settings
// block, the block is tool-owned until the user edits it. When the curation
// later evolves, an unedited (tool-written) block is safe to refresh to the
// new default, while an edited block is user tuning and must never be
// touched by the default strategy.
//
// The provenance mechanism is a historical-hash table rather than the
// originally sketched `# managed-by:` comment stamp: the YAML writer drops
// comments on every rewrite (a fleet-visible quirk observed in the
// 2026-11-02 sweep), so a comment stamp would not survive the first
// save-after-write, and the sidecar carries policy, not settings state. A
// block whose canonical hash matches a recorded generation is proven
// tool-written; anything else is user-owned.
//
// New generations are APPENDED here whenever a curated block's content
// changes in a release (the replaced content gets a new entry, keyed by the
// release that injected it). The current generation is never listed — a
// block equal to the current curated default is a no-op before provenance
// is even consulted.

// ManagedSettingsGeneration pins one historical curated-defaults generation:
// the canonical hash of each tool-owned settings block exactly as that
// release wrote it into fleet configs (YAML shape: kebab-case keys, string
// list values).
type ManagedSettingsGeneration struct {
	// Since is the tool release whose configure runs injected this
	// generation.
	Since string
	// LinterSettings maps a linter name to the canonical hash of the block
	// that release injected for it.
	LinterSettings map[string]string
}

// wrapcheckV0110Block is the wrapcheck curated default as shipped through
// v0.11.x: the stdlib signature regexps WITHOUT the errorfamily org
// exemption promoted on 2026-11-02 (commit 1cff856). Fleet configs converged
// before that promotion carry exactly this block, which is tool-written but
// stale.
func wrapcheckV0110Block() map[string]any {
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

// ManagedSettingsHistory lists every superseded curated generation, oldest
// intent last. Data-integrity specs pin the entries against the live
// curated table (a stale entry whose block equals the current default is
// noise and fails the spec).
var ManagedSettingsHistory = []ManagedSettingsGeneration{
	{
		Since:          "v0.11.0",
		LinterSettings: map[string]string{"wrapcheck": HashSettingsBlock(wrapcheckV0110Block())},
	},
}

// HashSettingsBlock returns a canonical content hash for a settings block.
// Canonical form is compact JSON: encoding/json sorts map keys at every
// level, so structurally equal blocks hash identically regardless of Go map
// iteration or YAML document order; list order is preserved because it is
// semantic.
func HashSettingsBlock(block map[string]any) string {
	canonical, err := json.Marshal(block)
	if err != nil {
		// map[string]any with JSON-representable values cannot fail to
		// marshal; a failure means a caller built an impossible block and
		// the only safe answer is a key no real block can collide with.
		return "unmarshalable"
	}

	sum := sha256.Sum256(canonical)

	return string(sum[:])
}

// IsHistoricalCuratedLinterBlock reports whether the block equals a settings
// block an older tool release injected as its curated default for the
// linter — i.e. the block is tool-written and stale, eligible for a
// managed refresh.
func IsHistoricalCuratedLinterBlock(name string, block map[string]any) bool {
	if len(block) == 0 {
		return false
	}

	hash := HashSettingsBlock(block)

	for _, generation := range ManagedSettingsHistory {
		if generation.LinterSettings[name] == hash {
			return true
		}
	}

	return false
}
