package config

import (
	"testing"
)

func TestDetectYAMLIndent(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			"2-space indentation",
			[]byte("version: \"2\"\nlinters:\n  enable:\n    - gosec\n"),
			2,
		},
		{
			"4-space indentation",
			[]byte("version: \"2\"\nlinters:\n    enable:\n        - gosec\n"),
			4,
		},
		{
			"defaults to 2 for root-only YAML",
			[]byte("version: \"2\"\n"),
			2,
		},
		{
			"skips comments and document markers",
			[]byte("# comment\n---\nversion: \"2\"\nlinters:\n  enable:\n    - gosec\n"),
			2,
		},
		{
			"defaults to 2 for empty data",
			[]byte(""),
			2,
		},
		{
			"defaults to 2 for whitespace-only data",
			[]byte("\n\n\n"),
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectYAMLIndent(tt.data); got != tt.want {
				t.Errorf("detectYAMLIndent() = %d, want %d", got, tt.want)
			}
		})
	}
}
