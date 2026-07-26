// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Git Utils", func() {
	DescribeTable(
		"IsGitRepo",
		func(path string, expected bool) {
			ctx := context.Background()
			Expect(utils.IsGitRepo(ctx, path)).To(Equal(expected))
		},
		Entry("should return true when in a git repository", ".", true),
		Entry("should return false for non-existent directory", "/non/existent/path", false),
	)

	DescribeTable(
		"CheckGitRepo",
		func(path string, expectSuccess bool) {
			ctx := context.Background()
			if expectSuccess {
				Expect(utils.CheckGitRepo(ctx, path)).To(Succeed())
			} else {
				Expect(utils.CheckGitRepo(ctx, path)).To(MatchError(apperrors.ErrNotGitRepository))
			}
		},
		Entry("should return nil when in a git repository", ".", true),
		Entry("should return error for non-existent directory", "/non/existent/path", false),
	)

	Context("CheckGitRepoWithTimeout", func() {
		It("should return nil when in a git repository", func() {
			Expect(utils.CheckGitRepoWithTimeout(".")).To(Succeed())
		})
	})

	Context("Constants", func() {
		It("should have reasonable timeout value", func() {
			Expect(utils.GitCheckTimeout).To(Equal(5 * time.Second))
		})
	})
})
