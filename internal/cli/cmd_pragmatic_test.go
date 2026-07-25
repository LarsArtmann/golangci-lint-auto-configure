package cli_test

import (
	"os"
	"os/exec"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v3"
)

var _ = Describe("--pragmatic flag", func() {
	// parseEnableList reads a golangci-lint config YAML file and returns the
	// linters.enable list.
	parseEnableList := func(configPath string) []string {
		data, err := os.ReadFile(configPath)
		Expect(err).NotTo(HaveOccurred())

		var cfg types.Config
		Expect(yaml.Unmarshal(data, &cfg)).To(Succeed())

		return cfg.Linters.Enable
	}

	// containsAny checks whether any element of needles appears in haystack.
	containsAny := func(haystack, needles []string) bool {
		set := make(map[string]struct{}, len(haystack))
		for _, h := range haystack {
			set[h] = struct{}{}
		}

		for _, n := range needles {
			if _, ok := set[n]; ok {
				return true
			}
		}

		return false
	}

	// pragmaticNoiseNames extracts linter names from PragmaticNoiseLinters.
	pragmaticNoiseNames := func() []string {
		names := make([]string, 0, len(constants.PragmaticNoiseLinters))
		for name := range constants.PragmaticNoiseLinters {
			names = append(names, string(name))
		}

		return names
	}

	It("should exclude the 5 noise linters from the enable set", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--pragmatic")
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))

		enableList := parseEnableList(configPath)

		noiseNames := pragmaticNoiseNames()

		Expect(containsAny(enableList, noiseNames)).
			To(BeFalse(),
				"--pragmatic should exclude all PragmaticNoiseLinters from enable, but found some: enable=%v noise=%v",
				enableList, noiseNames)
	})

	It("should include noise linters without the flag (default behavior)", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		cmd := exec.Command(binaryPath, "configure", "--config", configPath)
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))

		enableList := parseEnableList(configPath)

		noiseNames := pragmaticNoiseNames()

		Expect(containsAny(enableList, noiseNames)).
			To(BeTrue(),
				"without --pragmatic at least one noise linter should be enabled, but none found: enable=%v",
				enableList)
	})
})
