package utils_test

import (
	"context"
	"errors"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WithRetry", func() {
	var (
		ctx    context.Context
		config utils.Config
	)

	BeforeEach(func() {
		ctx = context.Background()
		config = utils.Config{
			MaxRetries:     2,
			InitialBackoff: 10 * time.Millisecond,
		}
	})

	Context("Success cases", func() {
		It("should return immediately on success", func() {
			callCount := 0
			//nolint:varnamelen // operation is clear in test context
			op := func() ([]byte, error) {
				callCount++

				return []byte("success"), nil
			}

			shouldRetry := func(_ error, _ string) bool { return false }

			result, err := utils.WithRetry(ctx, config, "test", shouldRetry, op)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("success")))
			Expect(callCount).To(Equal(1))
		})
	})

	Context("Retry cases", func() {
		It("should retry on retryable errors", func() {
			callCount := 0
			//nolint:varnamelen // operation is clear in test context
			op := func() ([]byte, error) {
				callCount++
				if callCount < 3 {
					return []byte("error output"), errors.New("retryable error")
				}

				return []byte("success"), nil
			}

			shouldRetry := func(err error, output string) bool {
				return err != nil && output == "error output"
			}

			result, err := utils.WithRetry(ctx, config, "test", shouldRetry, op)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("success")))
			Expect(callCount).To(Equal(3))
		})

		It("should fail after max retries", func() {
			callCount := 0
			//nolint:varnamelen // operation is clear in test context
			op := func() ([]byte, error) {
				callCount++

				return nil, errors.New("persistent error")
			}

			shouldRetry := func(err error, _ string) bool {
				return err != nil
			}

			_, err := utils.WithRetry(ctx, config, "test", shouldRetry, op)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed after 2 retries"))
			Expect(callCount).To(Equal(3)) // Initial (0) + 2 retries (1, 2)
		})
	})

	Context("Context cancellation", func() {
		It("should return context error when canceled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			callCount := 0

			//nolint:varnamelen // operation is clear in test context
			op := func() ([]byte, error) {
				callCount++
				if callCount == 1 {
					cancel() // Cancel context on first call
				}

				return nil, errors.New("retryable")
			}

			shouldRetry := func(_ error, _ string) bool { return true }

			_, err := utils.WithRetry(ctx, config, "test", shouldRetry, op)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("retry interrupted"))
		})
	})

	Context("Non-retryable errors", func() {
		It("should not retry on non-retryable errors", func() {
			callCount := 0
			//nolint:varnamelen // operation is clear in test context
			op := func() ([]byte, error) {
				callCount++

				return nil, errors.New("fatal error")
			}

			shouldRetry := func(_ error, _ string) bool { return false }

			_, err := utils.WithRetry(ctx, config, "test", shouldRetry, op)
			Expect(err).To(HaveOccurred())
			Expect(callCount).To(Equal(1))
		})
	})
})

var _ = Describe("DefaultConfig", func() {
	It("should return config with sensible defaults", func() {
		config := utils.DefaultConfig()
		Expect(config.MaxRetries).To(Equal(3))
		Expect(config.InitialBackoff).To(Equal(500 * time.Millisecond))
	})
})

var _ = Describe("IsContextCanceled", func() {
	It("should return true for context.Canceled", func() {
		err := context.Canceled
		Expect(utils.IsContextCanceled(err)).To(BeTrue())
	})

	It("should return true for context.DeadlineExceeded", func() {
		err := context.DeadlineExceeded
		Expect(utils.IsContextCanceled(err)).To(BeTrue())
	})

	It("should return false for other errors", func() {
		err := errors.New("some error")
		Expect(utils.IsContextCanceled(err)).To(BeFalse())
	})

	It("should return false for nil", func() {
		Expect(utils.IsContextCanceled(nil)).To(BeFalse())
	})
})
