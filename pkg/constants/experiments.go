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
