// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package formatters

// GoFumptFormatter implements gofumpt formatter.
type GoFumptFormatter struct {
	baseFormatter
}

// NewGoFumpt creates a new gofumpt formatter.
func NewGoFumpt() *GoFumptFormatter {
	return &GoFumptFormatter{
		baseFormatter: newBaseFormatter("gofumpt", "gofumpt", ErrFormatterNotFound),
	}
}

// Type returns the formatter type.
func (f *GoFumptFormatter) Type() FormatterType {
	return FormatterGoFumpt
}
