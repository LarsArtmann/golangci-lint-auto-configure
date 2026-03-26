// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

// FormatterType represents a code formatter.
type FormatterType int

// FormatterMode represents the execution mode.
type FormatterMode int

const (
	// FormatterGoFmt is the standard Go formatter (always available).
	FormatterGoFmt FormatterType = iota

	// FormatterGoImports organizes imports (requires external tool).
	FormatterGoImports

	// FormatterGoFumpt is a stricter Go formatter (requires external tool).
	FormatterGoFumpt

	// ModeCheck lists files that need formatting (uses -l flag).
	ModeCheck FormatterMode = iota

	// ModeFix formats files in-place (uses -w flag).
	ModeFix

	// ModeDiff shows formatting changes (uses -d flag).
	ModeDiff
)
