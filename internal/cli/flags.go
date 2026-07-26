package cli

// Flags holds all CLI flag values, bound by cobra at parse time.
// Replaces package-level global variables for testability and encapsulation.
type Flags struct {
	ConfigPath   string
	DryRun       bool
	Verbose      bool
	Quiet        bool
	OutputReport string
	Priority     string
	ReportFormat string
	NoAutoMerge  bool
	ShowDiff     bool
	JSONErrors   bool
	NoColor      bool
	NoAudit      bool
	Pragmatic    bool
}
