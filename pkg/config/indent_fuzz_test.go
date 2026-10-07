package config

import "testing"

// FuzzDetectYAMLIndent pins the indent-detection contract for arbitrary input:
// it never panics, always returns the default (2) or a valid width in
// [1, maxValidYAMLIndent], and is deterministic.
func FuzzDetectYAMLIndent(f *testing.F) {
	f.Add([]byte("linters:\n  enable:\n    - errcheck\n"))
	f.Add([]byte("a:\n b:\n  c: 1\n"))
	f.Add([]byte("a:\n     b: 1\n"))
	f.Add([]byte("# comment only\n---\n"))
	f.Add([]byte("a:\n\tb: 1\n"))
	f.Add([]byte(""))
	f.Add([]byte("\n\n\n"))
	f.Add([]byte("a:                     1\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		first := detectYAMLIndent(data)

		if first != defaultYAMLIndent && (first < 1 || first > maxValidYAMLIndent) {
			t.Fatalf("out-of-range indent %d for input %q", first, data)
		}

		if second := detectYAMLIndent(data); second != first {
			t.Fatalf("non-deterministic: %d then %d for input %q", first, second, data)
		}
	})
}
