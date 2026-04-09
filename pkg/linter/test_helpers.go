package linter

import (
	"os"

	"charm.land/log/v2"
)

// NewTestLogger creates a logger configured for test output (suppresses info/debug messages).
func NewTestLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{
		Level:           log.ErrorLevel,
		ReportCaller:    false,
		TimeFormat:      "15:04:05",
		TimeFunction:    nil,
		Prefix:          "",
		ReportTimestamp: false,
		CallerFormatter: nil,
		CallerOffset:    0,
		Fields:          nil,
		Formatter:       log.TextFormatter,
	})
}
