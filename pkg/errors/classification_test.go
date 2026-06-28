// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package apperrors_test

import (
	stderrors "errors"
	"fmt"
	"os/exec"

	errorfamily "github.com/larsartmann/go-error-family"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
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

	It("should classify AnalysisError via sentinel chain", func() {
		err := apperrors.NewAnalysisError("too old", "", fmt.Errorf("%w: need v2.10+", apperrors.ErrVersionTooOld))
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Rejection))
	})

	It("should return exit 0 for nil errors", func() {
		Expect(errorfamily.ExitCode(nil)).To(Equal(0))
	})
})
