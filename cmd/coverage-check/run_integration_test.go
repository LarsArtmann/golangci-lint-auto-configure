package main

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("run (integration)", func() {
	var (
		workDir     string
		profilePath string
		origDir     string
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		profilePath = filepath.Join(workDir, "coverage.out")

		origDir, _ = os.Getwd()

		Expect(os.Chdir(workDir)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(origDir) })

		// Create a minimal Go module with two functions, only one tested.
		Expect(os.WriteFile(
			filepath.Join(workDir, "go.mod"),
			[]byte("module testcov\n\ngo 1.21\n"),
			0o644,
		)).To(Succeed())

		Expect(os.WriteFile(filepath.Join(workDir, "cov.go"), []byte(`package testcov

func Covered(x int) int { return x * 2 }

func Uncovered(x int) int { return x * 3 }
`), 0o644)).To(Succeed())

		Expect(os.WriteFile(filepath.Join(workDir, "cov_test.go"), []byte(`package testcov

import "testing"

func TestCovered(t *testing.T) {
	if Covered(2) != 4 {
		t.Fatal("expected 4")
	}
}
`), 0o644)).To(Succeed())

		// Generate a real coverage profile (50% — 1 of 2 functions covered).
		cmd := exec.Command("go", "test", "-coverprofile="+profilePath, "./...")
		cmd.Dir = workDir
		Expect(cmd.Run()).To(Succeed())
	})

	It("passes when coverage meets the threshold", func() {
		Expect(run(40.0, profilePath)).To(Succeed())
	})

	It("fails when coverage is below the threshold", func() {
		err := run(90.0, profilePath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("coverage below minimum threshold"))
	})

	It("returns an error for a missing profile file", func() {
		err := run(60.0, filepath.Join(workDir, "nonexistent.out"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("coverage profile not found"))
	})
})
