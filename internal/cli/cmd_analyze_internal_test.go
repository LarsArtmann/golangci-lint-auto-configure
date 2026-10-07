package cli

import (
	"strings"
	"testing"
)

func TestDoctorLine(t *testing.T) {
	tests := []struct {
		name          string
		localGo       string
		golangciGo    string
		wantSubstring []string
	}{
		{
			name:          "both known",
			localGo:       "1.27.1",
			golangciGo:    "1.27",
			wantSubstring: []string{"local go1.27.1", "golangci-lint built with go1.27", "capped at go1.27"},
		},
		{
			name:          "local unknown",
			localGo:       "",
			golangciGo:    "1.27",
			wantSubstring: []string{"local unknown", "built with go1.27"},
		},
		{
			name:          "binary unknown",
			localGo:       "1.27.1",
			golangciGo:    "",
			wantSubstring: []string{"local go1.27.1", "built with unknown", "capped at unknown"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := doctorLine(test.localGo, test.golangciGo)
			for _, want := range test.wantSubstring {
				if !strings.Contains(got, want) {
					t.Fatalf("doctorLine(%q, %q) = %q, want substring %q", test.localGo, test.golangciGo, got, want)
				}
			}
		})
	}
}
