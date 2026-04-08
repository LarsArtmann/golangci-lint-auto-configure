package linter

import (
	"os"

	"charm.land/log/v2"
)

// NewTestLogger creates a logger configured for test output (suppresses info/debug messages).
func NewTestLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{
		Level:            log.ErrorLevel,
		TimeFunction:     nil,
		TimeFormat:       "",
		Prefix:           "",
		ReportTimestamp:  false,
		ReportCaller:     false,
		CallerFormatter:  nil,
		CallerOffset:     0,
		Fields:           nil,
		Formatter:        nil,
	})
}
