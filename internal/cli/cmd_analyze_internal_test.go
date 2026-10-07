package cli

import (
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
				if !contains(got, want) {
					t.Fatalf("doctorLine(%q, %q) = %q, want substring %q", test.localGo, test.golangciGo, got, want)
				}
			}
		})
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}

	return -1
}
