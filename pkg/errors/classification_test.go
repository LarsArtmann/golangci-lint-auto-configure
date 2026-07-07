// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package apperrors_test

import (
	stderrors "errors"
	"fmt"
	"os/exec"

	errorfamily "github.com/larsartmann/go-error-family"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Classification", func() {
	DescribeTable(
		"sentinel errors classify to the correct family",
		func(err error, expected errorfamily.Family) {
			Expect(errorfamily.Classify(err)).To(Equal(expected))
		},
		Entry("ErrNotGitRepository", apperrors.ErrNotGitRepository, errorfamily.Rejection),
		Entry("ErrNotInGitWorkingTree", apperrors.ErrNotInGitWorkingTree, errorfamily.Rejection),
		Entry("ErrUnknownPreset", apperrors.ErrUnknownPreset, errorfamily.Rejection),
		Entry("ErrInvalidActivityContext", apperrors.ErrInvalidActivityContext, errorfamily.Rejection),
		Entry("ErrVersionTooOld", apperrors.ErrVersionTooOld, errorfamily.Rejection),
		Entry("ErrConfigValidationFailed", apperrors.ErrConfigValidationFailed, errorfamily.Rejection),
		Entry("ErrNoConfigFiles", apperrors.ErrNoConfigFiles, errorfamily.Rejection),
		Entry("types.ErrConfigNil", types.ErrConfigNil, errorfamily.Rejection),
		Entry("types.ErrVersionRequired", types.ErrVersionRequired, errorfamily.Rejection),
		Entry("types.ErrVersionInvalid", types.ErrVersionInvalid, errorfamily.Rejection),
		Entry("types.ErrTimeoutRequired", types.ErrTimeoutRequired, errorfamily.Rejection),
		Entry("types.ErrIssuesExitCode", types.ErrIssuesExitCode, errorfamily.Rejection),
		Entry("types.ErrConcurrency", types.ErrConcurrency, errorfamily.Rejection),
		Entry("types.ErrMaxIssues", types.ErrMaxIssues, errorfamily.Rejection),
		Entry("types.ErrMaxSameIssues", types.ErrMaxSameIssues, errorfamily.Rejection),
		Entry("types.ErrInvalidLinterPriority", types.ErrInvalidLinterPriority, errorfamily.Rejection),
		Entry("ErrHookAlreadyExists", apperrors.ErrHookAlreadyExists, errorfamily.Conflict),
		Entry("ErrChangesNeeded", apperrors.ErrChangesNeeded, errorfamily.Conflict),
		Entry("ErrVersionParse", apperrors.ErrVersionParse, errorfamily.Corruption),
		Entry("ErrInvalidVersionFormat", apperrors.ErrInvalidVersionFormat, errorfamily.Corruption),
		Entry("exec.ErrNotFound", exec.ErrNotFound, errorfamily.Infrastructure),
	)

	DescribeTable(
		"sentinel exit codes match BSD sysexits",
		func(err error, expectedCode int) {
			Expect(errorfamily.ExitCode(err)).To(Equal(expectedCode))
		},
		Entry("Rejection → exit 1", apperrors.ErrNotGitRepository, 1),
		Entry("Conflict → exit 1", apperrors.ErrChangesNeeded, 1),
		Entry("Corruption → exit 65 (EX_DATAERR)", apperrors.ErrVersionParse, 65),
		Entry("Infrastructure → exit 69 (EX_UNAVAILABLE)", exec.ErrNotFound, 69),
	)

	It("should classify wrapped sentinels through the error chain", func() {
		wrapped := fmt.Errorf("outer: %w", apperrors.ErrVersionTooOld)
		Expect(errorfamily.Classify(wrapped)).To(Equal(errorfamily.Rejection))

		doubleWrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", apperrors.ErrVersionParse))
		Expect(errorfamily.Classify(doubleWrapped)).To(Equal(errorfamily.Corruption))
	})

	It("should classify ConfigError as Rejection", func() {
		err := apperrors.NewConfigError("failed to load", "/path/to/config.yml", stderrors.New("io error"))
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Rejection))
		Expect(errorfamily.ExitCode(err)).To(Equal(1))
	})

	It("should classify ConfigError as Rejection even when wrapping an Infrastructure cause", func() {
		err := apperrors.NewConfigError("config issue", "/path", exec.ErrNotFound)
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Rejection))
	})

	It("should classify ReportError as Rejection", func() {
		err := apperrors.NewReportError("html generation failed", "/out/report.html", stderrors.New("template error"))
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Rejection))
		Expect(errorfamily.ExitCode(err)).To(Equal(1))
	})

	It("should classify MigrationError as Rejection", func() {
		err := apperrors.NewMigrationError(
			"v1 config invalid",
			".golangci.yml",
			stderrors.New("yaml: line 5: bad mapping"),
		)
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Rejection))
		Expect(errorfamily.ExitCode(err)).To(Equal(1))
	})

	It("should classify AnalysisError via sentinel chain", func() {
		err := apperrors.NewAnalysisError("too old", "", fmt.Errorf("%w: need v2.10+", apperrors.ErrVersionTooOld))
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Rejection))
	})

	It("should return exit 0 for nil errors", func() {
		Expect(errorfamily.ExitCode(nil)).To(Equal(0))
	})
})
