package policy_test

import (
	"os"
	"path/filepath"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy", func() {
	var (
		tmpDir string
		path   string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = os.MkdirTemp("", "policy-test")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, policy.SidecarFileName)
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("Load", func() {
		It("returns nil when the sidecar does not exist", func() {
			pol, err := policy.Load(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(pol).To(BeNil())
		})

		It("parses a valid sidecar with justified disables", func() {
			content := `disabled:
  mnd:
    reason: "false-positives in file permissions"
    category: false-positives
  varnamelen:
    reason: "short names are idiomatic"
    category: convention
`
			Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())

			pol, err := policy.Load(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(pol).NotTo(BeNil())
			Expect(pol.Disabled).To(HaveLen(2))
			Expect(pol.Disabled["mnd"].Category).To(Equal(policy.CategoryFalsePositives))
			Expect(pol.Disabled["varnamelen"].Category).To(Equal(policy.CategoryConvention))
		})

		It("returns an error for malformed YAML", func() {
			Expect(os.WriteFile(path, []byte("{{invalid yaml"), 0o600)).To(Succeed())

			_, err := policy.Load(path)
			Expect(err).To(HaveOccurred())
		})

		It("handles an empty sidecar gracefully", func() {
			Expect(os.WriteFile(path, []byte(""), 0o600)).To(Succeed())

			pol, err := policy.Load(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(pol).NotTo(BeNil())
			Expect(pol.Disabled).To(BeEmpty())
		})
	})

	Describe("IsJustified", func() {
		It("returns true for a linter with a justification", func() {
			pol := &policy.Policy{
				Disabled: map[types.LinterName]policy.DisableJustification{
					"mnd": {Reason: "test", Category: policy.CategoryFalsePositives},
				},
			}

			Expect(pol.IsJustified("mnd")).To(BeTrue())
		})

		It("returns false for a linter without a justification", func() {
			pol := &policy.Policy{
				Disabled: map[types.LinterName]policy.DisableJustification{
					"mnd": {Reason: "test", Category: policy.CategoryFalsePositives},
				},
			}

			Expect(pol.IsJustified("ireturn")).To(BeFalse())
		})

		It("returns false when the policy is nil", func() {
			var pol *policy.Policy
			Expect(pol.IsJustified("mnd")).To(BeFalse())
		})
	})

	Describe("Justification", func() {
		It("returns the justification and true for a justified linter", func() {
			expected := policy.DisableJustification{
				Reason:   "test reason",
				Category: policy.CategoryPerformance,
			}
			pol := &policy.Policy{
				Disabled: map[types.LinterName]policy.DisableJustification{
					"mnd": expected,
				},
			}

			just, ok := pol.Justification("mnd")
			Expect(ok).To(BeTrue())
			Expect(just).To(Equal(expected))
		})

		It("returns zero and false for an unjustified linter", func() {
			pol := &policy.Policy{}

			_, ok := pol.Justification("mnd")
			Expect(ok).To(BeFalse())
		})
	})
})
