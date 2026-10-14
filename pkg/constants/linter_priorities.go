package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// LinterPriorities maps linter names to their priority levels.
// Derived from the policy table (linter_policy_table.go).
var LinterPriorities = linterPrioritiesView()

// CriticalLinters returns the names of the core critical linters that MUST
// be configured in every golangci-lint config for correctness checking.
// These are the fundamental linters (errcheck, staticcheck, govet) that
// the health check enforces. Other critical-priority linters (gosec, musttag,
// etc.) are also critical but the health check focuses on these three as
// the absolute minimum baseline.
func CriticalLinters() []types.LinterName {
	return []types.LinterName{"errcheck", "staticcheck", "govet"}
}
