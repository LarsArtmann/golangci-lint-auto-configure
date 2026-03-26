// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

// GoImportsFormatter implements goimports formatter.
type GoImportsFormatter struct {
	baseFormatter
}

// NewGoImports creates a new goimports formatter.
func NewGoImports() *GoImportsFormatter {
	return &GoImportsFormatter{
		baseFormatter: newBaseFormatter("goimports", "goimports", ErrFormatterNotFound),
	}
}

// Type returns the formatter type.
func (f *GoImportsFormatter) Type() FormatterType {
	return FormatterGoImports
}
