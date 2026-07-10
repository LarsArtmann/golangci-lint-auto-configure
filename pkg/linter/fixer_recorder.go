package linter

// configChangeRecorder ensures every config mutation is counted by wrapping
// mutations in closures. The recorder controls both execution and counting,
// preventing the footgun where a mutation runs but its count is not incremented
// (which would cause the total()==0 guard to silently discard all changes).
// Used consistently in both applyAllFixes and applyAndSave.
type configChangeRecorder struct {
	counts fixCounts
}

func (r *configChangeRecorder) deprecation(fn func() int) {
	r.counts.deprecation += fn()
}

func (r *configChangeRecorder) enable(fn func() int) {
	r.counts.enable += fn()
}

func (r *configChangeRecorder) formatter(fn func() int) {
	r.counts.formatter += fn()
}

func (r *configChangeRecorder) generated(fn func() int) {
	r.counts.generated += fn()
}

func (r *configChangeRecorder) normalize(fn func() int) {
	r.counts.normalization += fn()
}

func (r *configChangeRecorder) redundant(fn func() int) {
	r.counts.redundant += fn()
}
