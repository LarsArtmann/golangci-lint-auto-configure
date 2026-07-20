package audit_test

import (
	"os"
	"path/filepath"
	"strings"

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
