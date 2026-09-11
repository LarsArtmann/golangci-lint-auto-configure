package cli_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/internal/cli"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// executeInProcess runs the root command in-process with the given args and
// captured os.Stdout, returning everything written (charm log, cobra output,
// and report output all write to os.Stdout directly).
func executeInProcess(args ...string) string {
	cmd := cli.NewRootCommand(&cli.Flags{})

	original := os.Stdout

	reader, writer, pipeErr := os.Pipe()
	Expect(pipeErr).NotTo(HaveOccurred())

	os.Stdout = writer

	done := make(chan string, 1)
	go func() {
		captured, _ := io.ReadAll(reader)
		done <- string(captured)
	}()

	cmd.SetArgs(args)
	cmd.SetOut(writer)
	cmd.SetErr(writer)
	execErr := cmd.ExecuteContext(context.Background())

	Expect(os.Stdout.Close()).To(Succeed())
	os.Stdout = original

	captured := <-done
	Expect(execErr).NotTo(HaveOccurred(), "in-process execution failed; output:\n%s", captured)

	return captured
}

var _ = Context("analyze output modes (in-process)", func() {
	const validConfig = `version: "2"
linters:
  enable:
    - errcheck
`

	It("renders JSON output for --format json", func() {
		configPath := writeConfig(validConfig)

		output := executeInProcess("analyze", "--config", configPath, "--format", "json")

		jsonStart := strings.Index(output, "{")
		Expect(jsonStart).To(BeNumerically(">=", 0), "no JSON object in output:\n%s", output)

		Expect(output).To(ContainSubstring(`"ConfigPath"`))
		Expect(json.Valid([]byte(output[jsonStart:]))).To(BeTrue())
	})

	It("renders SARIF output for --format sarif", func() {
		configPath := writeConfig(validConfig)

		output := executeInProcess("analyze", "--config", configPath, "--format", "sarif")

		Expect(output).To(ContainSubstring(`"runs"`))
		Expect(output).To(ContainSubstring("2.1.0"))
	})

	It("renders finding-report output for --format finding", func() {
		configPath := writeConfig(validConfig)

		output := executeInProcess("analyze", "--config", configPath, "--format", "finding")

		Expect(output).To(ContainSubstring(`"findings"`))
	})

	It("falls back to text output for an unknown format", func() {
		configPath := writeConfig(validConfig)

		output := executeInProcess("analyze", "--config", configPath, "--format", "xml")

		Expect(output).NotTo(ContainSubstring(`"ConfigPath"`))
	})
})

var _ = Context("presets --json (in-process)", func() {
	It("emits preset entries as JSON", func() {
		output := executeInProcess("presets", "--json")

		Expect(output).To(ContainSubstring(`"Presets"`))
		Expect(output).To(ContainSubstring(`"Linters"`))
		Expect(json.Valid([]byte(strings.TrimSpace(output)))).To(BeTrue())
	})
})

var _ = Context("completion command (in-process)", func() {
	It("generates a bash completion script", func() {
		output := executeInProcess("completion", "bash")

		Expect(output).To(ContainSubstring("__start_golangci-lint-auto-configure"))
	})
})
