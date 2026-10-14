package linter

import (
	"testing"
)

func TestParseBuiltWithGoVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "real version output",
			in:   "golangci-lint has version 2.13.2 built with go1.27.1 from v2.13.2 on 1970-01-01T00:00:00Z\n",
			want: "go1.27.1",
		},
		{name: "no go version", in: "golangci-lint has version 2.13.2", want: ""},
		{name: "empty", in: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parseBuiltWithGoVersion(test.in); got != test.want {
				t.Fatalf("parseBuiltWithGoVersion(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}
