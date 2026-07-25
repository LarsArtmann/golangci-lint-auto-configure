package cli_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// pathWithoutGolangciLint returns the current PATH with every directory that
// contains a golangci-lint executable removed. Used to force the
// "binary not found" (Infrastructure, exit 69) condition deterministically,
// regardless of whether golangci-lint is installed on the host.
func pathWithoutGolangciLint() string {
	var kept []string

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}

		if isExecutableOnPath(filepath.Join(dir, "golangci-lint")) {
			continue
		}

		kept = append(kept, dir)
	}

	return strings.Join(kept, string(filepath.ListSeparator))
}

func isExecutableOnPath(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir() && info.Mode().Perm()&0o111 != 0
}

// envWithPATH clones the current environment, replacing PATH with newPath.
func envWithPATH(newPath string) []string {
	out := make([]string, 0, len(os.Environ())+1)

	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "PATH=") {
			continue
		}

		out = append(out, e)
	}

	return append(out, "PATH="+newPath)
}

// writeFakeGolangciLint creates a fake golangci-lint in dir that always prints
// unparseable output and exits 0. Placing dir first on PATH makes the tool
// resolve this fake, triggering the version-parse failure (Corruption, exit 65).
func writeFakeGolangciLint(dir string) {
	script := "#!/bin/sh\necho \"garbage-not-a-version\"\n"

	ExpectWithOffset(1, os.WriteFile(filepath.Join(dir, "golangci-lint"), []byte(script), 0o755)).
		To(Succeed())
}

var _ = Context("exit codes", func() {
	It("should exit 0 on successful configure", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		err := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "critical").Run()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should exit non-zero with invalid priority", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		err := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "invalid-priority").Run()
		Expect(err).To(HaveOccurred())

		exitErr, ok := errors.AsType[*exec.ExitError](err)
		Expect(ok).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(1))
	})

	It("should output JSON error with --json-errors flag", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		output, err := exec.Command(
			binaryPath, "configure", "--config", configPath, "--priority", "invalid-priority", "--json-errors",
		).CombinedOutput()

		Expect(err).To(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring(`"message"`))
		Expect(outputStr).To(ContainSubstring(`"family"`))
		Expect(outputStr).To(ContainSubstring(`"exit_code"`))
	})

	It("should exit 69 (Infrastructure) when golangci-lint binary is not found", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "critical")
		cmd.Env = envWithPATH(pathWithoutGolangciLint())

		err := cmd.Run()
		Expect(err).To(HaveOccurred())

		exitErr, ok := errors.AsType[*exec.ExitError](err)
		Expect(ok).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(69))
	})

	It("should exit 65 (Corruption) when golangci-lint emits unparseable version output", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		fakeDir := GinkgoT().TempDir()
		writeFakeGolangciLint(fakeDir)

		shadowedPath := fakeDir + string(filepath.ListSeparator) + os.Getenv("PATH")
		cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "critical")
		cmd.Env = envWithPATH(shadowedPath)

		err := cmd.Run()
		Expect(err).To(HaveOccurred())

		exitErr, ok := errors.AsType[*exec.ExitError](err)
		Expect(ok).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(65))
	})
})
