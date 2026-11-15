package linter

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/log/v2"
)

// writeMemoBinary materializes an executable stand-in for golangci-lint so
// FindBinary resolves from a controlled PATH without the real tool.
func writeMemoBinary(t *testing.T, dir string) {
	t.Helper()

	path := filepath.Join(dir, "golangci-lint")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
}

// TestFindBinaryWarnsOnceAcrossRepeatedCalls pins the duplicate-binary
// warning to fire exactly once per Analyzer: the converge pipeline calls
// FindBinary five times per fix run and each call used to re-scan PATH and
// re-emit the warning (2026-11-15 BuildFlow DiscordSync report showed five
// identical WARN lines back to back).
func TestFindBinaryWarnsOnceAcrossRepeatedCalls(t *testing.T) {
	dirOne := filepath.Join(t.TempDir(), "one")
	dirTwo := filepath.Join(t.TempDir(), "two")

	for _, dir := range []string{dirOne, dirTwo} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}

		writeMemoBinary(t, dir)
	}

	t.Setenv("PATH", dirOne+string(os.PathListSeparator)+dirTwo)

	var logBuf bytes.Buffer

	logger := log.NewWithOptions(&logBuf, log.Options{
		Level:           log.WarnLevel,
		ReportTimestamp: false,
	})

	analyzer := NewAnalyzer(logger)

	for range 5 {
		if err := analyzer.FindBinary(context.Background()); err != nil {
			t.Fatalf("FindBinary failed: %v", err)
		}
	}

	warnings := strings.Count(logBuf.String(), "Multiple golangci-lint binaries")
	if warnings != 1 {
		t.Errorf("FindBinary warned %d times across 5 calls, want exactly 1: %s", warnings, logBuf.String())
	}

	if analyzer.golangciLintPath == "" {
		t.Fatal("FindBinary left golangciLintPath empty after success")
	}
}

// TestFindBinaryRetriesAfterFailedLookup pins the failure path: a failed
// lookup must keep the memo empty so a later call scans PATH again instead
// of short-circuiting on a stale decision.
func TestFindBinaryRetriesAfterFailedLookup(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	var logBuf bytes.Buffer

	logger := log.NewWithOptions(&logBuf, log.Options{
		Level:           log.WarnLevel,
		ReportTimestamp: false,
	})

	analyzer := NewAnalyzer(logger)

	if err := analyzer.FindBinary(context.Background()); err == nil {
		t.Fatal("FindBinary succeeded on an empty PATH, want failure")
	}

	writeMemoBinary(t, dir)

	if err := analyzer.FindBinary(context.Background()); err != nil {
		t.Fatalf("FindBinary failed after the binary appeared: %v", err)
	}
}
