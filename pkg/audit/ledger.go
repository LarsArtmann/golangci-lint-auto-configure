// Package audit provides an append-only ledger that records every config change
// golangci-lint-auto-configure makes to a .golangci.yml file. The ledger lives
// in the OS cache dir (never in the git tree) so it is tamper-resistant across
// branch switches, mirroring BuildFlow's persistence model.
//
// All operations are best-effort: if the ledger file cannot be opened or
// written, the failure is logged and Record returns silently. A run never fails
// because the audit trail could not be persisted.
package audit

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"charm.land/log/v2"
)

const (
	appName           = "golangci-lint-auto-configure"
	ledgerFileName    = "audit.jsonl"
	dirPermissions    = 0o755
	filePermissions   = 0o644
	scannerMinBuffer  = 64 * 1024
	scannerMaxBuffer  = 1024 * 1024
	repoHashHexDigits = 16
	runIDSuffixBytes  = 4
)

// Action describes what happened to a linter (or the config) in a run.
type Action string

const (
	ActionAddedToEnable      Action = "added-to-enable"
	ActionRemovedFromEnable  Action = "removed-from-enable"
	ActionPreservedDisable   Action = "preserved-disable"
	ActionMovedToDisable     Action = "moved-to-disable"
	ActionRemovedFromDisable Action = "removed-from-disable"
	ActionPrunedSettings     Action = "pruned-settings"
	// ActionReEnabled records that an unjustified disable was undone (Pillar C enforcement).
	ActionReEnabled Action = "re-enabled"
)

// Entry is a single audit record, serialized as one JSONL line.
type Entry struct {
	Timestamp      time.Time `json:"timestamp"`
	RunID          string    `json:"run_id"`
	RepoHash       string    `json:"repo_hash"`
	RepoPath       string    `json:"repo_path"`
	GitHead        string    `json:"git_head,omitempty"`
	Linter         string    `json:"linter"`
	Action         Action    `json:"action"`
	Reason         string    `json:"reason,omitempty"`
	FindingsHidden int       `json:"findings_hidden,omitempty"`
}

// RunContext holds the immutable per-run metadata stamped on every entry.
type RunContext struct {
	RunID    string
	RepoHash string
	RepoPath string
	GitHead  string
}

// Recorder records audit entries. Implementations must be safe for concurrent use.
type Recorder interface {
	Record(action Action, linter, reason string)
}

// NoopRecorder discards all records. The zero value is ready to use.
type NoopRecorder struct{}

// Record implements Recorder by discarding the entry.
func (NoopRecorder) Record(Action, string, string) {}

// Ledger is an append-only audit log writer backed by a JSONL file.
// All operations are best-effort: failures are logged via the provided logger
// and never surface to the caller.
type Ledger struct {
	runCtx  RunContext
	path    string
	logger  *log.Logger
	mu      sync.Mutex
	enabled bool
}

// DefaultLedgerDir resolves the cache directory for the ledger
// ($XDG_CACHE_HOME/golangci-lint-auto-configure or the OS equivalent).
// Returns "" if the OS cache dir cannot be resolved.
func DefaultLedgerDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		return ""
	}

	return filepath.Join(cacheDir, appName)
}

// DefaultLedgerPath returns the default ledger file path, or "" if unavailable.
func DefaultLedgerPath() string {
	dir := DefaultLedgerDir()
	if dir == "" {
		return ""
	}

	return filepath.Join(dir, ledgerFileName)
}

// NewLedger creates a ledger for the given run. If path is empty, or its parent
// directory cannot be created, the returned ledger is disabled (Record is a
// silent no-op).
func NewLedger(logger *log.Logger, runCtx RunContext, path string) *Ledger {
	enabled := path != ""

	if enabled {
		if dir := filepath.Dir(path); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, dirPermissions); err != nil {
				logger.Warnf("Audit ledger disabled: cannot create dir %s: %v", dir, err)

				enabled = false
			}
		}
	}

	return &Ledger{
		runCtx:  runCtx,
		path:    path,
		logger:  logger,
		mu:      sync.Mutex{},
		enabled: enabled,
	}
}

// Enabled reports whether the ledger is actively persisting to disk.
func (l *Ledger) Enabled() bool {
	return l != nil && l.enabled
}

// Path returns the ledger file path, or "" if disabled.
func (l *Ledger) Path() string {
	if l == nil {
		return ""
	}

	return l.path
}

// Record appends an entry to the ledger. Best-effort: failures are logged, never returned.
func (l *Ledger) Record(action Action, linter, reason string) {
	if l == nil || !l.enabled {
		return
	}

	entry := Entry{
		Timestamp:      time.Now().UTC(),
		RunID:          l.runCtx.RunID,
		RepoHash:       l.runCtx.RepoHash,
		RepoPath:       l.runCtx.RepoPath,
		GitHead:        l.runCtx.GitHead,
		Linter:         linter,
		Action:         action,
		Reason:         reason,
		FindingsHidden: 0,
	}

	line, err := json.Marshal(entry)
	if err != nil {
		l.logger.Warnf("Audit ledger: cannot marshal entry: %v", err)

		return
	}

	l.appendLine(append(line, '\n'))
}

// appendLine writes one JSONL line under the mutex. Best-effort.
func (l *Ledger) appendLine(line []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		l.logger.Warnf("Audit ledger: cannot open %s: %v", l.path, err)

		return
	}
	defer file.Close()

	if _, err := file.Write(line); err != nil {
		l.logger.Warnf("Audit ledger: write failed: %v", err)
	}
}

// NewRunID generates a timestamped run ID: YYYYMMDD-HHMMSS-<8 hex>.
// The random suffix disambiguates runs that start within the same second.
func NewRunID() string {
	b := make([]byte, runIDSuffixBytes)
	_, _ = rand.Read(b)

	return time.Now().UTC().Format("20060102-150405") + "-" + hex.EncodeToString(b)
}

// RepoHashOf returns the first 16 hex chars of SHA-256 of the absolute repo path,
// isolating per-repo entries in the shared ledger (mirrors BuildFlow's repo_hash).
func RepoHashOf(repoPath string) string {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		abs = repoPath
	}

	sum := sha256.Sum256([]byte(abs))

	return hex.EncodeToString(sum[:])[:repoHashHexDigits]
}

// ReadAll reads all entries from the ledger file at path.
// Malformed lines are skipped (crash resilience, mirroring BuildFlow's loader).
func ReadAll(path string) ([]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open audit ledger %q: %w", path, err)
	}
	defer file.Close()

	var entries []Entry

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, scannerMinBuffer), scannerMaxBuffer)

	for scanner.Scan() {
		if entry, ok := parseEntry(scanner.Text()); ok {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan audit ledger %q: %w", path, err)
	}

	return entries, nil
}

// parseEntry decodes one JSONL line into an Entry. Returns ok=false for blank or
// malformed lines (malformed lines are skipped for crash resilience).
func parseEntry(line string) (Entry, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return Entry{}, false
	}

	var entry Entry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return Entry{}, false
	}

	return entry, true
}
