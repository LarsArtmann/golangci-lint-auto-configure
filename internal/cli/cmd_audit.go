package cli

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	// auditEnvVar mirrors pkg/audit.EnvVarDisable for help text; the behavior
	// lives in audit.OpenRunLedger (single source of truth).
	auditEnvVar     = audit.EnvVarDisable
	cliHoursPerDay  = 24
	tabwriterPad    = 2
	runIDMinParts   = 3
	runIDHexPrefix  = 4
	runIDPartSep    = "-"
	emptyReasonDash = "—"
)

var errLedgerPathUnavailable = errors.New(
	"cannot resolve audit ledger path: OS cache directory unavailable",
)

const auditLong = `Display the audit trail of configuration changes recorded by configure runs.

The ledger lives in the OS cache directory (~/.cache/golangci-lint-auto-configure/)
and records every linter enable, disable, settings prune, and re-enable this tool
has performed. Use this to review what an automated run (e.g. a pre-commit hook)
changed in your .golangci.yml.

Filters:
  --since <dur>   Only show entries newer than this duration (e.g. 24h, 7d, 720h)
  --linter <name> Only show entries for the given linter

Output:
  --json          Output as a JSON array instead of a table

Maintenance:
  --clear         Truncate the ledger file and exit`

func newAuditCommand(builder *CommandBuilder) *cobra.Command {
	var (
		jsonOutput   bool
		sinceFilter  string
		linterFilter string
		clearLedger  bool
	)

	cmd := builder.Build(
		"audit",
		"Show the audit trail of config changes",
		func(_ *cobra.Command, _ []string) error {
			return runAuditCommand(
				builder.Logger(),
				jsonOutput,
				sinceFilter,
				linterFilter,
				clearLedger,
			)
		},
		WithLong(auditLong),
	)

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as a JSON array instead of a table")
	cmd.Flags().
		StringVar(&sinceFilter, "since", "", "Only show entries newer than this duration (e.g. 24h, 7d)")
	cmd.Flags().StringVar(&linterFilter, "linter", "", "Only show entries for this linter name")
	cmd.Flags().BoolVar(&clearLedger, "clear", false, "Truncate the ledger file and exit")

	return cmd
}

func runAuditCommand(
	logger *log.Logger,
	jsonOutput bool,
	sinceFilter, linterFilter string,
	clearLedger bool,
) error {
	path := audit.DefaultLedgerPath()
	if path == "" {
		return errorfamily.WrapRejection(errLedgerPathUnavailable, "audit.ledger_path",
			"cannot resolve audit ledger path: OS cache directory unavailable")
	}

	if clearLedger {
		return clearAuditLedger(logger, path)
	}

	return displayAuditEntries(logger, path, jsonOutput, sinceFilter, linterFilter)
}

func displayAuditEntries(
	logger *log.Logger,
	path string,
	jsonOutput bool,
	sinceFilter, linterFilter string,
) error {
	entries, err := audit.ReadAll(path)
	if err != nil {
		// errors.Is unwraps the errorfamily.WrapTransientf chain that audit.ReadAll adds;
		// os.IsNotExist does not unwrap and would miss the wrapped *PathError.
		if errors.Is(err, os.ErrNotExist) {
			logger.Infof("No audit ledger found at %s", path)

			return nil
		}

		return apperrors.WrapClassified(err, "audit.read_ledger", "read audit ledger")
	}

	entries, err = filterAuditEntries(entries, sinceFilter, linterFilter)
	if err != nil {
		return err
	}

	return outputEntries(logger, entries, jsonOutput, path)
}

func outputEntries(
	logger *log.Logger,
	entries []audit.Entry,
	jsonOutput bool,
	path string,
) error {
	if len(entries) == 0 {
		logger.Infof("No audit entries found (ledger: %s)", path)

		return nil
	}

	if jsonOutput {
		return outputAuditJSON(entries)
	}

	outputAuditTable(entries)

	return nil
}

func clearAuditLedger(logger *log.Logger, path string) error {
	err := audit.Clear(path)
	if err != nil {
		return apperrors.WrapClassified(err, "audit.clear_ledger", "clear audit ledger")
	}

	logger.Infof("Cleared audit ledger at %s", path)

	return nil
}

func filterAuditEntries(
	entries []audit.Entry,
	sinceFilter, linterFilter string,
) ([]audit.Entry, error) {
	maxAge, err := parseSinceDuration(sinceFilter)
	if err != nil {
		return nil, err
	}

	cutoff := time.Time{}
	if maxAge > 0 {
		cutoff = time.Now().UTC().Add(-maxAge)
	}

	filtered := make([]audit.Entry, 0, len(entries))

	for _, entry := range entries {
		if !entryMatchesFilters(entry, linterFilter, cutoff) {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered, nil
}

func entryMatchesFilters(entry audit.Entry, linterFilter string, cutoff time.Time) bool {
	if linterFilter != "" && entry.Linter != linterFilter {
		return false
	}

	if !cutoff.IsZero() && entry.Timestamp.Before(cutoff) {
		return false
	}

	return true
}

func parseSinceDuration(since string) (time.Duration, error) {
	if since == "" {
		return 0, nil
	}

	if dayStr, ok := strings.CutSuffix(since, "d"); ok {
		dayCount, err := strconv.Atoi(dayStr)
		if err != nil {
			return 0, errorfamily.WrapRejectionf(err, "audit.parse_since_days",
				"invalid --since %q: expected a day count (e.g. 7d)", since)
		}

		return time.Duration(dayCount) * cliHoursPerDay * time.Hour, nil
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		return 0, errorfamily.WrapRejectionf(err, "audit.parse_since_duration",
			"invalid --since %q: use Go duration (24h) or days (7d)", since)
	}

	return duration, nil
}

func outputAuditJSON(entries []audit.Entry) error {
	data, err := json.Marshal(
		entries,
		jsontext.WithIndentPrefix(""),
		jsontext.WithIndent("  "),
		json.Deterministic(true),
	)
	if err != nil {
		return errorfamily.WrapCorruptionf(err, "audit.marshal_json", "marshal audit entries")
	}

	printBytesToStdout(data)

	return nil
}

func outputAuditTable(entries []audit.Entry) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, tabwriterPad, ' ', 0)
	defer writer.Flush()

	fmt.Fprintln(writer, "TIMESTAMP\tRUN ID\tLINTER\tACTION\tREASON")

	for _, entry := range entries {
		writeAuditRow(writer, entry)
	}
}

func writeAuditRow(writer *tabwriter.Writer, entry audit.Entry) {
	reason := entry.Reason
	if reason == "" {
		reason = emptyReasonDash
	}

	fmt.Fprintf(
		writer, "%s\t%s\t%s\t%s\t%s\n",
		entry.Timestamp.Format(time.DateTime),
		shortRunID(entry.RunID),
		entry.Linter,
		string(entry.Action),
		reason,
	)
}

// shortRunID truncates the run ID for table display (keeps timestamp + first hex chars).
func shortRunID(runID string) string {
	parts := strings.Split(runID, runIDPartSep)
	if len(parts) < runIDMinParts || len(parts[2]) < runIDHexPrefix {
		return runID
	}

	return parts[0] + runIDPartSep + parts[1] + runIDPartSep + parts[2][:runIDHexPrefix]
}
