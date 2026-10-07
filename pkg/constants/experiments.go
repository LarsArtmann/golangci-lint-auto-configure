package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// GoExperiments lists Go runtime experiments that expose new standard library packages.
// These require build tags for golangci-lint to analyze code using them.
//
// See: https://go.dev/src/internal/goexperiment/flags.go
//
// Only experiments that expose new user-visible packages/APIs are included.
// Internal-only experiments (FieldTrack, PreemptibleLoops, etc.) and
// experiments that are already on by default (RegabiWrappers, LoopVar, etc.)
// do not need build tags.
var GoExperiments = []types.GoExperiment{
	{
		Tag:         "goexperiment.arenas",
		Package:     "arena",
		Description: "Experimental arena-based allocator for reducing GC pressure",
	},
	{
		Tag:         "goexperiment.goroutineleakprofile",
		Package:     "runtime/pprof",
		Description: "Enables goroutine leak profiling in runtime/pprof",
	},
	{
		Tag:         "goexperiment.jsonv2",
		Package:     "encoding/json/v2",
		Description: "New JSON API with improved performance and correctness",
		// Graduated in Go 1.27: encoding/json/v2 is on by default and the
		// build tag is inert there. Kept for Go 1.26 toolchains, where the
		// experiment gate still applies.
		GraduatedIn: "1.27",
	},
	{
		Tag:         "goexperiment.runtimesecret",
		Package:     "runtime/secret",
		Description: "Enables the runtime/secret package for secret handling",
	},
	{
		Tag:         "goexperiment.simd",
		Package:     "simd",
		Description: "SIMD intrinsics and the simd package for vectorized operations",
	},
}

// GoExperimentTags returns the build tags for all Go experiments.
func GoExperimentTags() []string {
	tags := make([]string, 0, len(GoExperiments))
	for _, exp := range GoExperiments {
		tags = append(tags, exp.Tag)
	}

	return tags
}

// GoExperimentTagsFor returns the build tags still relevant for the given
// local Go major.minor version. Experiments that graduated at or before the
// local toolchain are excluded (their build tags are inert there); an empty
// or unparsable version returns all tags.
func GoExperimentTagsFor(localGoMajorMinor string) []string {
	tags := make([]string, 0, len(GoExperiments))

	local, ok := types.NormalizeGoMajorMinor(localGoMajorMinor)

	for _, exp := range GoExperiments {
		if exp.GraduatedIn != "" && ok && types.CompareGoMajorMinor(local, exp.GraduatedIn) >= 0 {
			continue
		}

		tags = append(tags, exp.Tag)
	}

	return tags
}
