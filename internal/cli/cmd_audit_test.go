package cli

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
)

func testAuditLogger() *log.Logger {
	return log.NewWithOptions(io.Discard, log.Options{Level: log.ErrorLevel})
}

// captureStdout runs fn while redirecting os.Stdout, returning whatever was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w

	defer func() { os.Stdout = orig }()

	done := make(chan string)

	go func() {
		var buf bytes.Buffer

		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()

	return <-done
}

// writeTestLedger writes the given entries as JSONL to path.
func writeTestLedger(t *testing.T, path string, entries ...audit.Entry) {
	t.Helper()

	var buf bytes.Buffer

	for _, e := range entries {
		data, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("marshal entry: %v", err)
		}

		buf.Write(data)
		buf.WriteByte('\n')
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write ledger: %v", err)
	}
}

func sampleEntry(linter string, action audit.Action, ts time.Time) audit.Entry {
	return audit.Entry{
		Timestamp: ts,
		RunID:     "20260101-120000-abcdef12",
		RepoHash:  "0123456789abcdef",
		RepoPath:  "/tmp/repo",
		Linter:    linter,
		Action:    action,
		Reason:    "test reason",
	}
}

func TestParseSinceDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"empty returns zero", "", 0, false},
		{"hours", "24h", 24 * time.Hour, false},
		{"days", "7d", 7 * cliHoursPerDay * time.Hour, false},
		{"single day", "1d", cliHoursPerDay * time.Hour, false},
		{"minutes", "30m", 30 * time.Minute, false},
		{"composite", "1h30m", 90 * time.Minute, false},
		{"invalid duration", "garbage", 0, true},
		{"invalid day count", "xd", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSinceDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSinceDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("parseSinceDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestShortRunID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"full id truncated to first 4 hex of third part",
			"20260101-120000-abcdef123456",
			"20260101-120000-abcd",
		},
		{
			"short id returned as-is",
			"abc",
			"abc",
		},
		{
			"two-part id returned as-is",
			"20260101-120000",
			"20260101-120000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shortRunID(tt.input); got != tt.want {
				t.Errorf("shortRunID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEntryMatchesFilters(t *testing.T) {
	now := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)
	recentEntry := sampleEntry("errcheck", audit.ActionAddedToEnable, now.Add(-1*time.Hour))
	oldEntry := sampleEntry("gofmt", audit.ActionMovedToDisable, now.Add(-400*24*time.Hour))
	recentCutoff := now.Add(-48 * time.Hour) // 2 days ago

	tests := []struct {
		name         string
		entry        audit.Entry
		linterFilter string
		cutoff       time.Time
		want         bool
	}{
		{"no filters", recentEntry, "", time.Time{}, true},
		{"linter filter matches", recentEntry, "errcheck", time.Time{}, true},
		{"linter filter mismatch", recentEntry, "gofmt", time.Time{}, false},
		{"within cutoff", recentEntry, "", recentCutoff, true},
		{"before cutoff", oldEntry, "", recentCutoff, false},
		{"zero cutoff admits all", oldEntry, "", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entryMatchesFilters(tt.entry, tt.linterFilter, tt.cutoff); got != tt.want {
				t.Errorf("entryMatchesFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterAuditEntries(t *testing.T) {
	now := time.Now()
	entries := []audit.Entry{
		sampleEntry("errcheck", audit.ActionAddedToEnable, now.Add(-1*time.Hour)),
		sampleEntry("gofmt", audit.ActionMovedToDisable, now.Add(-2*time.Hour)),
		sampleEntry("govet", audit.ActionPrunedSettings, now.Add(-400*24*time.Hour)),
	}

	t.Run("no filters returns all", func(t *testing.T) {
		got, err := filterAuditEntries(entries, "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 3 {
			t.Errorf("expected 3 entries, got %d", len(got))
		}
	})

	t.Run("linter filter", func(t *testing.T) {
		got, err := filterAuditEntries(entries, "", "errcheck")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 || got[0].Linter != "errcheck" {
			t.Errorf("expected only errcheck, got %v", got)
		}
	})

	t.Run("since filter", func(t *testing.T) {
		// 1 day keeps the two ~hours-old entries, drops the ~400-day-old one.
		got, err := filterAuditEntries(entries, "1d", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 2 {
			t.Errorf("expected 2 entries within 1 day, got %d", len(got))
		}
	})

	t.Run("invalid since returns error", func(t *testing.T) {
		_, err := filterAuditEntries(entries, "not-a-duration", "")
		if err == nil {
			t.Fatal("expected error for invalid duration")
		}
	})
}

func TestOutputEntries_EmptyReturnsNoError(t *testing.T) {
	logger := testAuditLogger()

	if err := outputEntries(logger, nil, false, "/tmp/ledger.jsonl"); err != nil {
		t.Errorf("outputEntries with empty slice should not error, got %v", err)
	}
}

func TestOutputEntries_Table(t *testing.T) {
	logger := testAuditLogger()
	entries := []audit.Entry{
		sampleEntry("errcheck", audit.ActionAddedToEnable, time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)),
	}

	out := captureStdout(t, func() {
		if err := outputEntries(logger, entries, false, "/tmp/ledger.jsonl"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "TIMESTAMP") || !strings.Contains(out, "errcheck") {
		t.Errorf("expected table header and linter name in output, got:\n%s", out)
	}

	if !strings.Contains(out, "added-to-enable") {
		t.Errorf("expected action in output, got:\n%s", out)
	}
}

func TestOutputEntries_JSON(t *testing.T) {
	logger := testAuditLogger()
	entries := []audit.Entry{
		sampleEntry("errcheck", audit.ActionReEnabled, time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)),
	}

	out := captureStdout(t, func() {
		if err := outputEntries(logger, entries, true, "/tmp/ledger.jsonl"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	trimmed := strings.TrimSpace(out)

	var parsed []audit.Entry
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		t.Fatalf("JSON output should be valid JSON, got error: %v\noutput:\n%s", err, out)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 parsed entry, got %d", len(parsed))
	}

	if parsed[0].Linter != "errcheck" {
		t.Errorf("expected linter errcheck, got %q", parsed[0].Linter)
	}

	if parsed[0].Action != audit.ActionReEnabled {
		t.Errorf("expected action re-enabled, got %q", parsed[0].Action)
	}
}

func TestDisplayAuditEntries_NoLedgerReturnsNoError(t *testing.T) {
	logger := testAuditLogger()
	missingPath := filepath.Join(t.TempDir(), "does-not-exist.jsonl")

	if err := displayAuditEntries(logger, missingPath, false, "", ""); err != nil {
		t.Errorf("missing ledger should return nil (info message), got %v", err)
	}
}

func TestDisplayAuditEntries_ReadsAndFilters(t *testing.T) {
	logger := testAuditLogger()
	dir := t.TempDir()
	ledgerPath := filepath.Join(dir, "audit.jsonl")

	writeTestLedger(
		t, ledgerPath,
		sampleEntry("errcheck", audit.ActionAddedToEnable, time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)),
		sampleEntry("gofmt", audit.ActionMovedToDisable, time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)),
	)

	out := captureStdout(t, func() {
		if err := displayAuditEntries(logger, ledgerPath, false, "", "errcheck"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "errcheck") {
		t.Errorf("expected errcheck in filtered output, got:\n%s", out)
	}

	if strings.Contains(out, "gofmt") {
		t.Errorf("gofmt should be filtered out, got:\n%s", out)
	}
}

func TestClearAuditLedger(t *testing.T) {
	logger := testAuditLogger()
	dir := t.TempDir()
	ledgerPath := filepath.Join(dir, "audit.jsonl")

	writeTestLedger(
		t, ledgerPath,
		sampleEntry("errcheck", audit.ActionAddedToEnable, time.Now()),
	)

	if err := clearAuditLedger(logger, ledgerPath); err != nil {
		t.Fatalf("clearAuditLedger failed: %v", err)
	}

	info, err := os.Stat(ledgerPath)
	if err != nil {
		t.Fatalf("ledger file should still exist after clear: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("ledger should be empty after clear, got size %d", info.Size())
	}
}

func TestAuditDisabled_FlagAndEnv(t *testing.T) {
	original := noAudit

	t.Cleanup(func() { noAudit = original })

	t.Run("enabled by default", func(t *testing.T) {
		noAudit = false

		t.Setenv(auditEnvVar, "")

		if auditDisabled() {
			t.Error("expected auditDisabled() to be false")
		}
	})

	t.Run("disabled via flag", func(t *testing.T) {
		noAudit = true

		t.Setenv(auditEnvVar, "")

		if !auditDisabled() {
			t.Error("expected auditDisabled() to be true when flag set")
		}
	})

	t.Run("disabled via env var", func(t *testing.T) {
		noAudit = false

		t.Setenv(auditEnvVar, "1")

		if !auditDisabled() {
			t.Error("expected auditDisabled() to be true when env var set")
		}
	})
}

func TestNewRunLedger_DisabledReturnsNoop(t *testing.T) {
	original := noAudit

	t.Cleanup(func() { noAudit = original })

	noAudit = true

	configFile := filepath.Join(t.TempDir(), ".golangci.yml")

	recorder := newRunLedger(context.Background(), testAuditLogger(), configFile)

	if _, ok := recorder.(audit.NoopRecorder); !ok {
		t.Errorf("expected audit.NoopRecorder when disabled, got %T", recorder)
	}
}

func TestNewRunLedger_EnabledReturnsLedger(t *testing.T) {
	original := noAudit

	t.Cleanup(func() { noAudit = original })

	noAudit = false

	t.Setenv(auditEnvVar, "")

	// Force a deterministic, writable cache dir so DefaultLedgerPath resolves.
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	configFile := filepath.Join(t.TempDir(), ".golangci.yml")

	recorder := newRunLedger(context.Background(), testAuditLogger(), configFile)

	ledger, ok := recorder.(*audit.Ledger)
	if !ok {
		t.Fatalf("expected *audit.Ledger when enabled, got %T", recorder)
	}

	if !ledger.Enabled() {
		t.Error("expected ledger to be enabled")
	}

	if ledger.Path() == "" {
		t.Error("expected non-empty ledger path")
	}
}

func TestRunAuditCommand_ClearLedger(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	ledgerPath := audit.DefaultLedgerPath()
	if ledgerPath == "" {
		t.Skip("ledger path unavailable on this platform")
	}

	if err := os.MkdirAll(filepath.Dir(ledgerPath), 0o755); err != nil {
		t.Fatalf("mkdir ledger dir: %v", err)
	}

	writeTestLedger(
		t, ledgerPath,
		sampleEntry("errcheck", audit.ActionAddedToEnable, time.Now()),
	)

	if err := runAuditCommand(testAuditLogger(), false, "", "", true); err != nil {
		t.Fatalf("runAuditCommand clear failed: %v", err)
	}

	if info, err := os.Stat(ledgerPath); err != nil || info.Size() != 0 {
		t.Errorf("ledger should be empty after clear via runAuditCommand")
	}
}
