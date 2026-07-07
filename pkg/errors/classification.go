package apperrors

import (
	"os/exec"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Classification mapping rationale:
//
// Rejection (exit 1) — user's fault: bad input, missing prerequisite,
//   invalid config. The user can fix it by changing their input or environment.
// Conflict (exit 1) — user must resolve: a state conflict (hook already exists)
//   or a check-mode signal (changes needed). Not an error per se, but non-zero exit.
// Corruption (exit 65) — data integrity: version output is unparseable, suggesting
//   a broken golangci-lint installation or corrupted output.
// Infrastructure (exit 69) — system cannot serve: golangci-lint binary not found.

// init registers all sentinel errors with their behavioral Families. This
// follows the go-error-family registration pattern and must run before Main().
//
//nolint:gochecknoinits // required by go-error-family registration pattern
func init() {
	errorfamily.RegisterStdlibDefaults(errorfamily.DefaultRegistry)

	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		// User-fault errors — bad input, missing prerequisites.
		ErrNotGitRepository:       errorfamily.Rejection,
		ErrNotInGitWorkingTree:    errorfamily.Rejection,
		ErrUnknownPreset:          errorfamily.Rejection,
		ErrInvalidActivityContext: errorfamily.Rejection,
		ErrVersionTooOld:          errorfamily.Rejection,
		ErrConfigValidationFailed: errorfamily.Rejection,
		ErrNoConfigFiles:          errorfamily.Rejection,

		// Config validation sentinels — all user-fault.
		types.ErrConfigNil:       errorfamily.Rejection,
		types.ErrVersionRequired: errorfamily.Rejection,
		types.ErrVersionInvalid:  errorfamily.Rejection,
		types.ErrTimeoutRequired: errorfamily.Rejection,
		types.ErrIssuesExitCode:  errorfamily.Rejection,
		types.ErrConcurrency:     errorfamily.Rejection,
		types.ErrMaxIssues:       errorfamily.Rejection,
		types.ErrMaxSameIssues:   errorfamily.Rejection,

		// Linter priority parse errors — user-fault.
		types.ErrInvalidLinterPriority: errorfamily.Rejection,

		// State conflicts — user must resolve before proceeding.
		ErrHookAlreadyExists: errorfamily.Conflict,
		ErrChangesNeeded:     errorfamily.Conflict,

		// Data integrity — version output is unparseable or malformed.
		ErrVersionParse:         errorfamily.Corruption,
		ErrInvalidVersionFormat: errorfamily.Corruption,

		// Infrastructure — golangci-lint binary not found in PATH.
		exec.ErrNotFound: errorfamily.Infrastructure,
	})
}

// ErrorFamily classifies ConfigError as Rejection.
// Config errors are always user-fault: file not found, parse failure,
// validation failure. The user must fix their config file.
func (*ConfigError) ErrorFamily() errorfamily.Family {
	return errorfamily.Rejection
}

// ErrorFamily classifies ReportError as Rejection.
// Report failures stem from user-provided config paths or output options
// (invalid format, unwritable directory, missing template variables).
func (*ReportError) ErrorFamily() errorfamily.Family {
	return errorfamily.Rejection
}

// ErrorFamily classifies MigrationError as Rejection.
// Migration failures are user-fault: unsupported config version, malformed
// YAML, missing required fields. The user must fix their v1 config.
func (*MigrationError) ErrorFamily() errorfamily.Family {
	return errorfamily.Rejection
}

// AnalysisError does NOT implement Classified — it represents heterogeneous
// failures (binary-not-found → Infrastructure, version-too-old → Rejection,
// unparseable-output → Corruption). Its cause-chain sentinels handle
// fine-grained classification. See classification_test.go for verification.
