package linter

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
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

// writeVersionProbeBinary materializes an executable stand-in that answers
// `version --json` with the minimum valid version and appends one line per
// invocation to countFile, so the test can prove the subprocess run is cached.
func writeVersionProbeBinary(t *testing.T, dir, countFile string) {
	t.Helper()

	script := "#!/bin/sh\n"
	if countFile != "" {
		script += "echo invoked >> " + countFile + "\n"
	}

	script += "cat <<'JSONEOF'\n" +
		`{"version":"` + constants.MinGolangCILintVersion + `","commit":"test","date":"2026","goVersion":"go1.27"}` +
		"\nJSONEOF\n"

	path := filepath.Join(dir, "golangci-lint")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake version binary: %v", err)
	}
}

// TestCheckVersionSpawnsOncePerAnalyzer pins the version-check memo: the
// converge pipeline called CheckVersion three to four times per fix run,
// each spawning `golangci-lint version --json` afresh.
func TestCheckVersionSpawnsOncePerAnalyzer(t *testing.T) {
	dir := t.TempDir()
	countFile := filepath.Join(dir, "invocations")

	writeVersionProbeBinary(t, dir, countFile)

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = filepath.Join(dir, "golangci-lint")

	for range 3 {
		if err := analyzer.CheckVersion(context.Background()); err != nil {
			t.Fatalf("CheckVersion failed: %v", err)
		}
	}

	data, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("read invocation count: %v", err)
	}

	if got := strings.Count(string(data), "invoked"); got != 1 {
		t.Errorf("version subprocess spawned %d times across 3 CheckVersion calls, want 1", got)
	}

	if analyzer.GetDetectedVersion() == "" {
		t.Fatal("CheckVersion left detectedVersion empty after success")
	}
}

// TestCheckVersionRetriesAfterFailedCheck pins the failure path: a failed
// version check must keep the memo empty so a later call spawns again.
func TestCheckVersionRetriesAfterFailedCheck(t *testing.T) {
	dir := t.TempDir()

	writeMemoBinary(t, dir)

	analyzer := NewAnalyzer(NewTestLogger())
	analyzer.golangciLintPath = filepath.Join(dir, "golangci-lint")

	broken := filepath.Join(dir, "golangci-lint")
	if err := os.WriteFile(broken, []byte("#!/bin/sh\necho garbage\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write broken binary: %v", err)
	}

	if err := analyzer.CheckVersion(context.Background()); err == nil {
		t.Fatal("CheckVersion succeeded on garbage output, want failure")
	}

	writeVersionProbeBinary(t, dir, "")

	if err := analyzer.CheckVersion(context.Background()); err != nil {
		t.Fatalf("CheckVersion failed after the binary was fixed: %v", err)
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
