package config

import (
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func testMerger() *Merger {
	return &Merger{
		logger: log.NewWithOptions(nil, log.Options{Level: log.FatalLevel}),
		fs:     osFS{},
	}
}

// FuzzMergeConfigInto verifies that merging two configs never panics and
// always preserves primary linters while absorbing secondary ones.
func FuzzMergeConfigInto(f *testing.F) {
	f.Add("gosec,errcheck", "govet,gosec")
	f.Add("", "gosec,errcheck")
	f.Add("gosec,errcheck", "")
	f.Add("gosec,gosec,errcheck", "gosec,dupl")
	f.Add("errcheck", "gosec,errcheck,govet")

	f.Fuzz(func(t *testing.T, primaryLinters, secondaryLinters string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("mergeConfigInto panicked: %v", r)
			}
		}()

		primary := configFromLinters(primaryLinters)
		secondary := configFromLinters(secondaryLinters)

		originalPrimary := types.NewSet(primary.Linters.Enable...)

		cm := testMerger()
		cm.mergeConfigInto(primary, secondary)

		merged := types.NewSet(primary.Linters.Enable...)

		for _, linter := range types.ToSortedSlice(originalPrimary) {
			if !merged.Contains(linter) {
				t.Errorf("primary linter %q lost after merge", linter)
			}
		}

		for _, linter := range parseLinters(secondaryLinters) {
			if !merged.Contains(linter) {
				t.Errorf("secondary linter %q not absorbed", linter)
			}
		}
	})
}

// FuzzMergeIdempotent verifies that merging the same secondary config twice
// produces the same result as merging once (merge is idempotent).
func FuzzMergeIdempotent(f *testing.F) {
	f.Add("gosec,errcheck", "govet,gosec")
	f.Add("", "gosec")
	f.Add("errcheck", "errcheck")

	f.Fuzz(func(t *testing.T, primaryLinters, secondaryLinters string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("merge panicked: %v", r)
			}
		}()

		cfg1 := configFromLinters(primaryLinters)
		cfg2 := configFromLinters(primaryLinters)
		secondary := configFromLinters(secondaryLinters)

		cm := testMerger()
		cm.mergeConfigInto(cfg1, secondary)
		cm.mergeConfigInto(cfg1, secondary)

		cm.mergeConfigInto(cfg2, secondary)

		set1 := types.NewSet(cfg1.Linters.Enable...)
		set2 := types.NewSet(cfg2.Linters.Enable...)

		if !set1.Equal(set2) {
			t.Errorf("merge is not idempotent: once=%v, twice=%v",
				types.ToSortedSlice(set1), types.ToSortedSlice(set2))
		}
	})
}

func configFromLinters(linters string) *Config {
	return &Config{
		Linters: types.LintersConfig{
			Enable: parseLinters(linters),
		},
	}
}

func parseLinters(s string) []types.LinterName {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")

	var result []types.LinterName

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, types.LinterName(p))
		}
	}

	return result
}
