package linter_test

import (
	"context"
	"testing"

	"charm.land/log/v2"
	linterpkg "github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/stretchr/testify/assert"
)

func TestCheckVersion_Success(t *testing.T) {
	// This test requires golangci-lint v2.10.1+ to be installed
	analyzer := linterpkg.NewAnalyzer(log.Default())

	// First find the binary
	err := analyzer.FindBinary(context.Background())
	if err != nil {
		t.Skipf("golangci-lint not found in PATH: %v", err)
	}

	// Then check version (should pass with v2.10.1+)
	err = analyzer.CheckVersion(context.Background())
	assert.NoError(t, err, "Version check should pass with golangci-lint v2.10.1 or newer")
}
