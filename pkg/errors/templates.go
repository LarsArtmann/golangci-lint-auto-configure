package apperrors

import (
	errorfamily "github.com/larsartmann/go-error-family"
)

// RegisterDomainTemplates registers Wix-style MessageTemplates for every domain
// error code used in errorfamily.Wrap* calls across the codebase. This enables
// errorfamily.TemplateForCode(code) to resolve a user-friendly What/Why/Fix/WayOut
// presentation for structured error rendering.
//
// Called from init() in classification.go alongside the sentinel registrations.
//
//nolint:gochecknoinits // required by go-error-family registration pattern
func RegisterDomainTemplates() {
	errorfamily.DefaultRegistry.RegisterTemplates(domainMessageTemplates)
}

var domainMessageTemplates = map[string]errorfamily.MessageTemplate{
	// ── Git ──────────────────────────────────────────────────────────────────
	"git.not_repository": {
		What:   "Not a git repository.",
		Why:    "golangci-lint-auto-configure requires a git repository for version-control safety before modifying config files.",
		Fix:    "Run this command from the root of a git repository (a directory containing a .git folder).",
		WayOut: "Initialize a repo with 'git init' if this is a new project, or cd to your project root.",
	},
	"git.not_work_tree": {
		What:   "Not inside a git working tree.",
		Why:    "The current directory is part of a git repository but is not inside the working tree.",
		Fix:    "Move to a directory within the checked-out working tree.",
		WayOut: "Run 'git checkout <branch>' or 'git worktree list' to find a valid working directory.",
	},

	// ── Version checks ───────────────────────────────────────────────────────
	"version.too_old": {
		What:   "golangci-lint version is too old.",
		Why:    "This tool requires a newer version of golangci-lint to support all features.",
		Fix:    "Upgrade golangci-lint to the latest release.",
		WayOut: "See https://golangci-lint.run/usage/install/ for installation instructions.",
	},
	"version.invalid_format": {
		What:   "Could not parse the golangci-lint version.",
		Why:    "The version output does not match the expected semantic version format.",
		Fix:    "Ensure golangci-lint is properly installed and its version command works.",
		WayOut: "Reinstall golangci-lint or set GOLANGCI_LINT_BIN to a known-good binary.",
	},
	"version.parse_json": {
		What:   "Failed to parse golangci-lint JSON version output.",
		Why:    "The JSON output from 'golangci-lint version --short' is malformed or empty.",
		Fix:    "Verify golangci-lint is installed and functional.",
		WayOut: "Run 'golangci-lint version' manually to inspect the output.",
	},
	"version.parse_text": {
		What:   "Failed to parse golangci-lint text version output.",
		Why:    "The text version string does not contain a recognizable version number.",
		Fix:    "Ensure a supported version of golangci-lint is in PATH.",
		WayOut: "Reinstall golangci-lint from the official source.",
	},

	// ── Config loading ───────────────────────────────────────────────────────
	"config.unsupported_format": {
		What:   "Unsupported configuration file format.",
		Why:    "Only YAML (.yml/.yaml) and JSON (.json) config files are supported.",
		Fix:    "Convert your config to YAML or JSON format.",
		WayOut: "Use 'golangci-lint-auto-configure configure' to generate a fresh YAML config.",
	},
	"config.linters_parse": {
		What:   "Failed to parse the linters section of your config.",
		Why:    "The 'linters' key contains invalid or malformed data.",
		Fix:    "Check the linters.enable and linters.disable arrays for syntax errors.",
		WayOut: "Back up the current config and regenerate with 'golangci-lint-auto-configure configure'.",
	},

	// ── Config validation ────────────────────────────────────────────────────
	"config.validation.version": {
		What:   "Invalid 'run.go-version' in config.",
		Why:    "The Go version string does not follow the required '1.x.y' format.",
		Fix:    "Set run.go-version to a valid Go version (e.g., '1.26.0').",
		WayOut: "Remove the run.go-version field to use the system default.",
	},
	"config.validation.issues_exit_code": {
		What:   "Invalid 'run.issues-exit-code' in config.",
		Why:    "The value must be a positive integer (typically 1).",
		Fix:    "Set run.issues-exit-code to a valid exit code (e.g., 1).",
		WayOut: "",
	},
	"config.validation.concurrency": {
		What:   "Invalid 'run.concurrency' in config.",
		Why:    "Concurrency must be a positive integer.",
		Fix:    "Set run.concurrency to a positive value (e.g., 4) or remove it for auto-detection.",
		WayOut: "",
	},
	"config.validation.max_issues_per_linter": {
		What:   "Invalid 'issues.max-issues-per-linter' in config.",
		Why:    "The value must be a positive integer.",
		Fix:    "Set issues.max-issues-per-linter to a positive value (e.g., 50).",
		WayOut: "",
	},
	"config.validation.max_same_issues": {
		What:   "Invalid 'issues.max-same-issues' in config.",
		Why:    "The value must be a positive integer.",
		Fix:    "Set issues.max-same-issues to a positive value (e.g., 10).",
		WayOut: "",
	},

	// ── Migration ────────────────────────────────────────────────────────────
	"migration.parse_yaml": {
		What:   "Failed to parse the v1 config YAML.",
		Why:    "The YAML syntax is invalid or contains unsupported constructs.",
		Fix:    "Fix the YAML syntax errors in your v1 config file.",
		WayOut: "Validate with 'yamllint' or an online YAML parser.",
	},
	"migration.encode_yaml": {
		What:   "Failed to encode the migrated v2 config.",
		Why:    "The migrated config could not be serialized to YAML.",
		Fix:    "This is likely a tool bug. Report it with the input config attached.",
		WayOut: "",
	},
	"migration.validate_config": {
		What:   "The migrated v2 config failed validation.",
		Why:    "The v1-to-v2 conversion produced an invalid configuration.",
		Fix:    "Review the validation errors and adjust the v1 config.",
		WayOut: "Report this as a bug if the v1 config is valid.",
	},

	// ── Report generation ────────────────────────────────────────────────────
	"report.create_output": {
		What:   "Failed to create the report output file.",
		Why:    "The output directory may not exist or you lack write permissions.",
		Fix:    "Ensure the output directory exists and is writable.",
		WayOut: "Choose a different output path with --output.",
	},
	"report.render": {
		What:   "Failed to render the report.",
		Why:    "The template engine encountered an error during rendering.",
		Fix:    "This is likely a tool bug. Report it with the config and analysis output.",
		WayOut: "",
	},
	"report.json_marshal": {
		What:   "Failed to marshal the JSON report.",
		Why:    "The report data could not be serialized to JSON.",
		Fix:    "This is a tool bug. Report it with the input config.",
		WayOut: "",
	},
	"report.json_write": {
		What:   "Failed to write the JSON report to disk.",
		Why:    "The output path may not be writable.",
		Fix:    "Ensure the output directory exists and is writable.",
		WayOut: "",
	}, ────────────────────────────────────────────────────────────
	"detector.open_file": {
		What:   "Failed to open a source file during detection.",
		Why:    "A file detected during the scan could not be read.",
		Fix:    "Check file permissions and ensure no files are locked.",
		WayOut: "",
	},
	"detector.walk_dir": {
		What:   "Failed to walk the project directory.",
		Why:    "A directory access error occurred during the scan.",
		Fix:    "Check directory permissions.",
		WayOut: "",
	},
	"detector.open_gomod": {
		What:   "Failed to open go.mod.",
		Why:    "The go.mod file exists but could not be read.",
		Fix:    "Check file permissions on go.mod.",
		WayOut: "",
	},
	"detector.scan_gomod": {
		What:   "Failed to scan go.mod for module dependencies.",
		Why:    "The go.mod file could not be parsed.",
		Fix:    "Ensure go.mod is valid (run 'go mod tidy').",
		WayOut: "",
	},
	"detector.walk_failed": {
		What:   "Failed to walk the directory tree.",
		Why:    "An I/O error occurred while scanning directories.",
		Fix:    "Check directory permissions and disk health.",
		WayOut: "",
	},

	// ── Retry ────────────────────────────────────────────────────────────────
	"retry.interrupted": {
		What:   "Operation was interrupted.",
		Why:    "The context was cancelled while retrying a transient failure.",
		Fix:    "Retry the command.",
		WayOut: "",
	},
	"retry.exhausted": {
		What:   "Operation failed after exhausting all retries.",
		Why:    "A transient failure persisted beyond the retry limit.",
		Fix:    "Check network connectivity and golangci-lint availability, then retry.",
		WayOut: "",
	},
}

// init wires domain templates into the DefaultRegistry at package load time.
//
//nolint:gochecknoinits // required by go-error-family registration pattern
func init() {
	RegisterDomainTemplates()
}
