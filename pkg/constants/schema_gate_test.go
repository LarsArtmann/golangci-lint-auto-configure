package constants_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The schema fixture (testdata/schema-fixture/.golangci.yml) is the CI
// schema-compat gate's input: it contains every default the fixer injects and
// must stay in lockstep with the constants. Committed to git so the gate is a
// visible diff, not a hidden regeneration.
const schemaFixturePath = "testdata/schema-fixture/.golangci.yml"

var _ = Describe("Schema fixture gate", func() {
	Describe("drift guard", func() {
		It("committed fixture matches the generator output for the current constants", func() {
			repoRoot := findRepoRoot()
			generated := filepath.Join(GinkgoT().TempDir(), ".golangci.yml")

			cmd := exec.Command("go", "run", "./cmd/generate-schema-fixture", "-output="+generated)
			cmd.Dir = repoRoot

			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(),
				"fixture generator failed: %s", strings.TrimSpace(string(output)))

			committed, err := os.ReadFile(filepath.Join(repoRoot, schemaFixturePath))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(committed)).To(Equal(readFileString(generated)),
				"committed schema fixture is stale — run `go run ./cmd/generate-schema-fixture` "+
					"and commit the regenerated file so new settings cannot skip the CI gate")
		})

		It("fixture contains an entry for every curated default setting", func() {
			content := readFileString(findRepoRoot() + "/" + schemaFixturePath)

			for linter := range constants.DefaultLinterSettings {
				Expect(content).To(ContainSubstring("    "+string(linter)+":\n"),
					"linter %s has curated defaults but is missing from the schema fixture", linter)
			}

			for formatter := range constants.DefaultFormatterSettings {
				Expect(content).To(ContainSubstring("    "+string(formatter)+":\n"),
					"formatter %s has curated defaults but is missing from the schema fixture", formatter)
			}
		})
	})

	Describe("schema verify", func() {
		It("golangci-lint accepts every injected default", func() {
			binary, err := exec.LookPath("golangci-lint")
			if err != nil {
				Skip("golangci-lint not in PATH — the dedicated CI schema-verify job is the authoritative gate")
			}

			cmd := exec.Command(binary, "config", "verify")
			cmd.Dir = findRepoRoot()
			cmd.Env = append(os.Environ(), "WORKING_DIR="+filepath.Dir(schemaFixturePath))

			cmd.Dir = filepath.Join(findRepoRoot(), filepath.Dir(schemaFixturePath))

			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(),
				"schema-invalid default injected: %s", strings.TrimSpace(string(output)))
		})
	})
})

func findRepoRoot() string {
	wd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			Fail("go.mod not found above " + wd)
		}

		dir = parent
	}
}

func readFileString(path string) string {
	content, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())

	return string(content)
}
