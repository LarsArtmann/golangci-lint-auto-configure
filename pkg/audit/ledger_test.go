package audit_test

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func newTestLogger() *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{Level: log.ErrorLevel})
}

var _ = Describe("Ledger", func() {
	var (
		tmpDir string
		path   string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = os.MkdirTemp("", "audit-ledger-test")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "audit.jsonl")
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("recording entries", func() {
		It("appends entries to the ledger file in order", func() {
			runCtx := audit.RunContext{
				RunID:    "20260720-120000-abcdef12",
				RepoHash: "deadbeefdeadbeef",
				RepoPath: "/repo",
			}
			ledger := audit.NewLedger(newTestLogger(), runCtx, path)

			ledger.Record(audit.ActionAddedToEnable, "gosec", "recommended: security")
			ledger.Record(audit.ActionPreservedDisable, "mnd", "user intent")

			entries, err := audit.ReadAll(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Action).To(Equal(audit.ActionAddedToEnable))
			Expect(entries[0].Linter).To(Equal("gosec"))
			Expect(entries[0].RunID).To(Equal("20260720-120000-abcdef12"))
			Expect(entries[0].RepoHash).To(Equal("deadbeefdeadbeef"))
			Expect(entries[1].Action).To(Equal(audit.ActionPreservedDisable))
			Expect(entries[1].Linter).To(Equal("mnd"))
		})

		It("stamps every entry with the shared run context", func() {
			runCtx := audit.RunContext{
				RunID:    "run-x",
				RepoHash: "hash-x",
				RepoPath: "/repo-x",
				GitHead:  "abc1234",
			}
			ledger := audit.NewLedger(newTestLogger(), runCtx, path)

			ledger.Record(audit.ActionMovedToDisable, "depguard", "superseded")

			entries, err := audit.ReadAll(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].GitHead).To(Equal("abc1234"))
			Expect(entries[0].RepoPath).To(Equal("/repo-x"))
		})

		It("preserves entries across multiple records (append-only)", func() {
			runCtx := audit.RunContext{RunID: "run-1", RepoHash: "h", RepoPath: "/r"}
			ledger := audit.NewLedger(newTestLogger(), runCtx, path)

			for range 5 {
				ledger.Record(audit.ActionAddedToEnable, "gosec", "")
			}

			entries, err := audit.ReadAll(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(5))
		})
	})

	Describe("graceful degradation", func() {
		It("is a no-op when the path is empty", func() {
			runCtx := audit.RunContext{RunID: "run", RepoHash: "h", RepoPath: "/r"}
			ledger := audit.NewLedger(newTestLogger(), runCtx, "")

			Expect(ledger.Enabled()).To(BeFalse())

			ledger.Record(audit.ActionAddedToEnable, "gosec", "reason")

			Expect(ledger.Path()).To(BeEmpty())
		})

		It("satisfies the Recorder interface", func() {
			var recorder audit.Recorder = audit.NoopRecorder{}
			recorder.Record(audit.ActionAddedToEnable, "gosec", "reason")

			var noop audit.Recorder = audit.NoopRecorder{}
			Expect(noop).NotTo(BeNil())
		})
	})

	Describe("ReadAll resilience", func() {
		It("skips malformed lines and returns valid ones", func() {
			content := strings.Join([]string{
				`{"timestamp":"2026-07-20T12:00:00Z","run_id":"r","repo_hash":"h","repo_path":"/r","linter":"gosec","action":"added-to-enable"}`,
				`not valid json`,
				``,
				`{"timestamp":"2026-07-20T12:00:01Z","run_id":"r","repo_hash":"h","repo_path":"/r","linter":"mnd","action":"preserved-disable"}`,
			}, "\n") + "\n"
			Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())

			entries, err := audit.ReadAll(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Linter).To(Equal("gosec"))
			Expect(entries[1].Linter).To(Equal("mnd"))
		})

		It("returns an error when the file does not exist", func() {
			_, err := audit.ReadAll(filepath.Join(tmpDir, "missing.jsonl"))
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Run and repo identifiers", func() {
	It("NewRunID produces a timestamped id with a hex suffix", func() {
		id := audit.NewRunID()
		Expect(id).To(MatchRegexp(`^\d{8}-\d{6}-[0-9a-f]{8}$`))
	})

	It("NewRunID disambiguates calls in the same second", func() {
		ids := map[string]struct{}{}
		for range 20 {
			ids[audit.NewRunID()] = struct{}{}
		}

		Expect(ids).To(HaveLen(20))
	})

	It("RepoHashOf is deterministic for the same path", func() {
		Expect(audit.RepoHashOf("/repo")).To(Equal(audit.RepoHashOf("/repo")))
	})

	It("RepoHashOf differs for different repos", func() {
		Expect(audit.RepoHashOf("/repo-a")).NotTo(Equal(audit.RepoHashOf("/repo-b")))
	})

	It("RepoHashOf returns 16 hex characters", func() {
		Expect(audit.RepoHashOf("/repo")).To(MatchRegexp(`^[0-9a-f]{16}$`))
	})
})

var _ = Describe("DefaultLedgerPath", func() {
	It("resolves a path ending in golangci-lint-auto-configure/audit.jsonl", func() {
		path := audit.DefaultLedgerPath()
		if path == "" {
			Skip("OS cache dir unavailable in this environment")
		}

		Expect(filepath.Base(path)).To(Equal("audit.jsonl"))
		Expect(path).To(ContainSubstring("golangci-lint-auto-configure"))
	})
})

var _ = Describe("PurgeOlder", func() {
	var (
		tmpDir string
		path   string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = os.MkdirTemp("", "audit-purge-test")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "audit.jsonl")
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	writeTestEntries := func(entries ...audit.Entry) {
		lines := make([]string, 0, len(entries))

		for _, entry := range entries {
			data, err := json.Marshal(entry)
			Expect(err).NotTo(HaveOccurred())

			lines = append(lines, string(data))
		}

		Expect(os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)).To(Succeed())
	}

	It("removes entries older than the cutoff and keeps recent ones", func() {
		old := time.Now().UTC().Add(-100 * 24 * time.Hour)
		recent := time.Now().UTC().Add(-1 * time.Hour)

		writeTestEntries(
			audit.Entry{
				Timestamp: old, RunID: "r", RepoHash: "h",
				RepoPath: "/r", Linter: "old1", Action: "added-to-enable",
			},
			audit.Entry{
				Timestamp: old, RunID: "r", RepoHash: "h",
				RepoPath: "/r", Linter: "old2", Action: "added-to-enable",
			},
			audit.Entry{
				Timestamp: recent, RunID: "r", RepoHash: "h",
				RepoPath: "/r", Linter: "new", Action: "added-to-enable",
			},
		)

		purged, err := audit.PurgeOlder(path, 90*24*time.Hour)
		Expect(err).NotTo(HaveOccurred())
		Expect(purged).To(Equal(2))

		entries, err := audit.ReadAll(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Linter).To(Equal("new"))
	})

	It("returns zero purged when all entries are recent", func() {
		recent := time.Now().UTC().Add(-1 * time.Hour)

		writeTestEntries(
			audit.Entry{
				Timestamp: recent, RunID: "r", RepoHash: "h",
				RepoPath: "/r", Linter: "a", Action: "added-to-enable",
			},
		)

		purged, err := audit.PurgeOlder(path, 90*24*time.Hour)
		Expect(err).NotTo(HaveOccurred())
		Expect(purged).To(Equal(0))
	})

	It("returns an error when the file does not exist", func() {
		_, err := audit.PurgeOlder(filepath.Join(tmpDir, "missing.jsonl"), time.Hour)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Clear", func() {
	var (
		tmpDir string
		path   string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = os.MkdirTemp("", "audit-clear-test")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "audit.jsonl")
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	It("truncates a non-empty ledger to zero entries", func() {
		runCtx := audit.RunContext{RunID: "r", RepoHash: "h", RepoPath: "/r"}
		ledger := audit.NewLedger(newTestLogger(), runCtx, path)
		ledger.Record(audit.ActionAddedToEnable, "gosec", "test")

		entries, err := audit.ReadAll(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))

		Expect(audit.Clear(path)).To(Succeed())

		entries, err = audit.ReadAll(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})

	It("succeeds even when the file does not exist yet", func() {
		Expect(audit.Clear(path)).To(Succeed())

		info, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Size()).To(BeZero())
	})
})

var _ = Describe("DefaultRetention", func() {
	It("returns a duration of approximately 90 days", func() {
		retention := audit.DefaultRetention()
		Expect(retention).To(BeNumerically(">", 89*24*time.Hour))
		Expect(retention).To(BeNumerically("<", 91*24*time.Hour))
	})
})

var _ = Describe("PreviouslyAutoEnabled", func() {
	var (
		tmpDir string
		path   string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = os.MkdirTemp("", "audit-previously-enabled-test")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "audit.jsonl")
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	It("returns linters auto-enabled for the matching repoHash", func() {
		runCtx := audit.RunContext{RunID: "r", RepoHash: "hash-a", RepoPath: "/repo-a"}
		ledger := audit.NewLedger(newTestLogger(), runCtx, path)

		ledger.Record(audit.ActionAddedToEnable, "godoclint", "recommended")
		ledger.Record(audit.ActionAddedToEnable, "ireturn", "recommended")
		ledger.Record(audit.ActionPreservedDisable, "mnd", "user intent")

		result := audit.PreviouslyAutoEnabled(path, "hash-a")
		Expect(result).To(HaveLen(2))
		Expect(result["godoclint"]).To(BeTrue())
		Expect(result["ireturn"]).To(BeTrue())
		Expect(result["mnd"]).To(BeFalse())
	})

	It("filters by repoHash (ignores entries from other repos)", func() {
		runCtxA := audit.RunContext{RunID: "r1", RepoHash: "hash-a", RepoPath: "/repo-a"}
		ledgerA := audit.NewLedger(newTestLogger(), runCtxA, path)
		ledgerA.Record(audit.ActionAddedToEnable, "godoclint", "recommended")

		runCtxB := audit.RunContext{RunID: "r2", RepoHash: "hash-b", RepoPath: "/repo-b"}
		ledgerB := audit.NewLedger(newTestLogger(), runCtxB, path)
		ledgerB.Record(audit.ActionAddedToEnable, "ireturn", "recommended")

		resultA := audit.PreviouslyAutoEnabled(path, "hash-a")
		Expect(resultA).To(HaveLen(1))
		Expect(resultA["godoclint"]).To(BeTrue())
		Expect(resultA["ireturn"]).To(BeFalse())

		resultB := audit.PreviouslyAutoEnabled(path, "hash-b")
		Expect(resultB).To(HaveLen(1))
		Expect(resultB["ireturn"]).To(BeTrue())
		Expect(resultB["godoclint"]).To(BeFalse())
	})

	It("returns nil for empty path", func() {
		Expect(audit.PreviouslyAutoEnabled("", "hash-a")).To(BeNil())
	})

	It("returns nil for empty repoHash", func() {
		Expect(audit.PreviouslyAutoEnabled(path, "")).To(BeNil())
	})

	It("returns nil when the file does not exist", func() {
		Expect(audit.PreviouslyAutoEnabled(filepath.Join(tmpDir, "missing.jsonl"), "hash-a")).To(BeNil())
	})

	It("returns empty map when no entries match", func() {
		runCtx := audit.RunContext{RunID: "r", RepoHash: "hash-a", RepoPath: "/r"}
		ledger := audit.NewLedger(newTestLogger(), runCtx, path)
		ledger.Record(audit.ActionPreservedDisable, "mnd", "user intent")

		result := audit.PreviouslyAutoEnabled(path, "hash-a")
		Expect(result).To(BeEmpty())
	})

	Describe("Ledger.PreviouslyAutoEnabled", func() {
		It("returns the set for this ledger's repo", func() {
			runCtx := audit.RunContext{RunID: "r", RepoHash: "hash-x", RepoPath: "/repo-x"}
			ledger := audit.NewLedger(newTestLogger(), runCtx, path)
			ledger.Record(audit.ActionAddedToEnable, "testableexamples", "recommended")

			result := ledger.PreviouslyAutoEnabled()
			Expect(result).To(HaveLen(1))
			Expect(result["testableexamples"]).To(BeTrue())
		})

		It("returns nil when disabled (empty path)", func() {
			runCtx := audit.RunContext{RunID: "r", RepoHash: "hash-x", RepoPath: "/repo-x"}
			ledger := audit.NewLedger(newTestLogger(), runCtx, "")

			Expect(ledger.PreviouslyAutoEnabled()).To(BeNil())
		})
	})

	Describe("NoopRecorder.PreviouslyAutoEnabled", func() {
		It("returns nil", func() {
			Expect(audit.NoopRecorder{}.PreviouslyAutoEnabled()).To(BeNil())
		})
	})
})
