// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Git Utils", func() {
	Context("IsGitRepo", func() {
		It("should return true when in a git repository", func() {
			ctx := context.Background()
			Expect(utils.IsGitRepo(ctx, ".")).To(BeTrue())
		})

		It("should return false for non-existent directory", func() {
			ctx := context.Background()
			Expect(utils.IsGitRepo(ctx, "/non/existent/path")).To(BeFalse())
		})
	})

	Context("CheckGitRepo", func() {
		It("should return nil when in a git repository", func() {
			ctx := context.Background()
			Expect(utils.CheckGitRepo(ctx, ".")).To(Succeed())
		})

		It("should return error for non-existent directory", func() {
			ctx := context.Background()
			err := utils.CheckGitRepo(ctx, "/non/existent/path")
			Expect(err).To(MatchError(apperrors.ErrNotGitRepository))
		})
	})

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
