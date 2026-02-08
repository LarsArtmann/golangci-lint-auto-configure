package linter

import (
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "standard format",
			output:   "golangci-lint has version 2.8.0 built with go1.25.5",
			expected: "2.8.0",
		},
		{
			name:     "with v prefix",
			output:   "golangci-lint has version v1.23.1 built with go1.19.1",
			expected: "v1.23.1",
		},
		{
			name:     "multiple spaces",
			output:   "golangci-lint has version  2.10.5  built with go1.26.0",
			expected: "2.10.5",
		},
		{
			name:     "no version found",
			output:   "golangci-lint has no version",
			expected: "",
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
		},
	}

	analyzer := NewAnalyzer(log.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.parseVersionText(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckVersion_Success(t *testing.T) {
	// This test requires golangci-lint v2.8.0+ to be installed
	analyzer := NewAnalyzer(log.Default())

	// First find the binary
	err := analyzer.FindBinary()
	if err != nil {
		t.Skipf("golangci-lint not found in PATH: %v", err)
	}

	// Then check version (should pass with v2.8.0+)
	err = analyzer.CheckVersion()
	assert.NoError(t, err, "Version check should pass with golangci-lint v2.8.0 or newer")
}
