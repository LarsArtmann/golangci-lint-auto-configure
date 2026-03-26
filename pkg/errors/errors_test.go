// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package apperrors_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
)

func TestErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errors Suite")
}

var _ = Describe("Sentinel Errors", func() {
	It("should have correct error messages", func() {
		Expect(apperrors.ErrNotGitRepository.Error()).To(ContainSubstring("not a git repository"))
		Expect(apperrors.ErrHookAlreadyExists.Error()).To(ContainSubstring("hook already exists"))
		Expect(apperrors.ErrUnknownPreset.Error()).To(ContainSubstring("unknown preset"))
		Expect(apperrors.ErrInvalidActivityContext.Error()).To(ContainSubstring("invalid activity context"))
		Expect(apperrors.ErrVersionParse.Error()).To(ContainSubstring("could not parse version"))
		Expect(apperrors.ErrInvalidVersionFormat.Error()).To(ContainSubstring("invalid version format"))
		Expect(apperrors.ErrVersionTooOld.Error()).To(ContainSubstring("version is too old"))
		Expect(apperrors.ErrConfigValidationFailed.Error()).To(ContainSubstring("configuration validation failed"))
	})

	It("should be comparable with errors.Is", func() {
		wrapped := fmt.Errorf("wrapped: %w", apperrors.ErrNotGitRepository)
		Expect(stderrors.Is(wrapped, apperrors.ErrNotGitRepository)).To(BeTrue())
	})
})

var _ = Describe("ConfigError", func() {
	It("should create error with message and path", func() {
		err := apperrors.NewConfigError("failed to load", "/path/to/config.yml", nil)
		Expect(err.Error()).To(Equal("failed to load (path: /path/to/config.yml)"))
		Expect(err.Message).To(Equal("failed to load"))
		Expect(err.Path).To(Equal("/path/to/config.yml"))
	})

	It("should include cause in error message", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewConfigError("failed to load", "/path/to/config.yml", cause)
		Expect(err.Error()).To(ContainSubstring("underlying error"))
	})

	It("should unwrap to cause", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewConfigError("failed to load", "/path/to/config.yml", cause)
		Expect(stderrors.Unwrap(err)).To(Equal(cause))
	})

	It("should be detectable with IsConfigError", func() {
		err := apperrors.NewConfigError("failed", "/path", nil)
		Expect(apperrors.IsConfigError(err)).To(BeTrue())
		Expect(apperrors.IsConfigError(fmt.Errorf("other error"))).To(BeFalse())
	})
})

var _ = Describe("AnalysisError", func() {
	It("should create error with message and file", func() {
		err := apperrors.NewAnalysisError("analysis failed", "analyzer.go", nil)
		Expect(err.Error()).To(Equal("analysis failed (file: analyzer.go)"))
		Expect(err.Message).To(Equal("analysis failed"))
		Expect(err.File).To(Equal("analyzer.go"))
	})

	It("should include cause in error message", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewAnalysisError("analysis failed", "analyzer.go", cause)
		Expect(err.Error()).To(ContainSubstring("underlying error"))
	})

	It("should unwrap to cause", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewAnalysisError("analysis failed", "analyzer.go", cause)
		Expect(stderrors.Unwrap(err)).To(Equal(cause))
	})

	It("should be detectable with IsAnalysisError", func() {
		err := apperrors.NewAnalysisError("failed", "file.go", nil)
		Expect(apperrors.IsAnalysisError(err)).To(BeTrue())
		Expect(apperrors.IsAnalysisError(fmt.Errorf("other error"))).To(BeFalse())
	})
})

var _ = Describe("ReportError", func() {
	It("should create error with message and path", func() {
		err := apperrors.NewReportError("report failed", "/path/to/report.html", nil)
		Expect(err.Error()).To(Equal("report failed (path: /path/to/report.html)"))
		Expect(err.Message).To(Equal("report failed"))
		Expect(err.Path).To(Equal("/path/to/report.html"))
	})

	It("should include cause in error message", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewReportError("report failed", "/path/to/report.html", cause)
		Expect(err.Error()).To(ContainSubstring("underlying error"))
	})

	It("should unwrap to cause", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewReportError("report failed", "/path/to/report.html", cause)
		Expect(stderrors.Unwrap(err)).To(Equal(cause))
	})

	It("should be detectable with IsReportError", func() {
		err := apperrors.NewReportError("failed", "/path", nil)
		Expect(apperrors.IsReportError(err)).To(BeTrue())
		Expect(apperrors.IsReportError(fmt.Errorf("other error"))).To(BeFalse())
	})
})

var _ = Describe("MigrationError", func() {
	It("should create error with message and config", func() {
		err := apperrors.NewMigrationError("migration failed", ".golangci.yml", nil)
		Expect(err.Error()).To(Equal("migration failed (config: .golangci.yml)"))
		Expect(err.Message).To(Equal("migration failed"))
		Expect(err.Config).To(Equal(".golangci.yml"))
	})

	It("should include cause in error message", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewMigrationError("migration failed", ".golangci.yml", cause)
		Expect(err.Error()).To(ContainSubstring("underlying error"))
	})

	It("should unwrap to cause", func() {
		cause := fmt.Errorf("underlying error")
		err := apperrors.NewMigrationError("migration failed", ".golangci.yml", cause)
		Expect(stderrors.Unwrap(err)).To(Equal(cause))
	})

	It("should be detectable with IsMigrationError", func() {
		err := apperrors.NewMigrationError("failed", ".golangci.yml", nil)
		Expect(apperrors.IsMigrationError(err)).To(BeTrue())
		Expect(apperrors.IsMigrationError(fmt.Errorf("other error"))).To(BeFalse())
	})
})

var _ = Describe("Error Chaining", func() {
	It("should support wrapping multiple levels", func() {
		cause := fmt.Errorf("root cause")
		configErr := apperrors.NewConfigError("config error", "/path", cause)
		wrapped := fmt.Errorf("operation failed: %w", configErr)

		// The chain is: wrapped -> configErr -> cause
		// So errors.Is should find cause through the chain
		Expect(stderrors.Is(wrapped, cause)).To(BeTrue())
		Expect(apperrors.IsConfigError(wrapped)).To(BeTrue())

		var cfgErr *apperrors.ConfigError
		Expect(stderrors.As(wrapped, &cfgErr)).To(BeTrue())
		Expect(cfgErr.Path).To(Equal("/path"))
	})

	It("should distinguish between error types", func() {
		configErr := apperrors.NewConfigError("config", "/path", nil)
		analysisErr := apperrors.NewAnalysisError("analysis", "file.go", nil)
		reportErr := apperrors.NewReportError("report", "/report", nil)
		migrationErr := apperrors.NewMigrationError("migration", ".golangci.yml", nil)

		Expect(apperrors.IsConfigError(configErr)).To(BeTrue())
		Expect(apperrors.IsAnalysisError(analysisErr)).To(BeTrue())
		Expect(apperrors.IsReportError(reportErr)).To(BeTrue())
		Expect(apperrors.IsMigrationError(migrationErr)).To(BeTrue())

		// Cross-checks should be false
		Expect(apperrors.IsConfigError(analysisErr)).To(BeFalse())
		Expect(apperrors.IsAnalysisError(reportErr)).To(BeFalse())
		Expect(apperrors.IsReportError(migrationErr)).To(BeFalse())
		Expect(apperrors.IsMigrationError(configErr)).To(BeFalse())
	})
})
