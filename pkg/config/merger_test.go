package config_test

import (
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testConfigYML = `
version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`

var _ = Describe("Merger", func() {
	var (
		merger  *config.Merger
		loader  *config.Loader
		testDir string
	)

	BeforeEach(func() {
		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		merger = config.NewMerger(logger)
		loader = config.NewLoader(logger)
		testDir = GinkgoT().TempDir()
	})

	Context("MergeConfigs", func() {
		It("should return error when no configs provided", func() {
			_, _, err := merger.MergeConfigs([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no config files"))
		})

		It("should return single config unchanged when only one file", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			configContent := `
version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			Expect(os.WriteFile(ymlPath, []byte(configContent), 0o644)).To(Succeed())

			cfg, result, err := merger.MergeConfigs([]string{ymlPath})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			Expect(result.PrimaryConfig).To(Equal(ymlPath))
			Expect(cfg.Version).To(Equal("2"))
			Expect(cfg.Run.Timeout).To(Equal("5m"))
			Expect(cfg.Linters.Enable).To(ContainElement("gosec"))
		})

		It("should merge linters from secondary config", func() {
			// Primary config (.golangci.yml has higher priority)
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			primaryContent := testConfigYML
			Expect(os.WriteFile(ymlPath, []byte(primaryContent), 0o644)).To(Succeed())

			// Secondary config (.golangci.yaml)
			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			secondaryContent := `
version: "2"
linters:
  enable:
    - errcheck
    - staticcheck
`
			Expect(os.WriteFile(yamlPath, []byte(secondaryContent), 0o644)).To(Succeed())

			cfg, result, err := merger.MergeConfigs([]string{ymlPath, yamlPath})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.IsSuccess()).To(BeTrue())
			Expect(result.PrimaryConfig).To(Equal(ymlPath))
			Expect(result.MergedConfigs).To(ContainElement(yamlPath))

			// Should have merged linters
			Expect(cfg.Linters.Enable).To(ContainElements("gosec", "errcheck", "staticcheck"))
		})

		It("should merge run settings from secondary config when primary is empty", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			primaryContent := testConfigYML
			Expect(os.WriteFile(ymlPath, []byte(primaryContent), 0o644)).To(Succeed())

			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			secondaryContent := `
version: "2"
run:
  timeout: 10m
  go: "1.23"
  tests: true
`
			Expect(os.WriteFile(yamlPath, []byte(secondaryContent), 0o644)).To(Succeed())

			cfg, _, err := merger.MergeConfigs([]string{ymlPath, yamlPath})

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Run.Timeout).To(Equal("10m"))
			Expect(cfg.Run.Go).To(Equal("1.23"))
			Expect(cfg.Run.Tests).To(BeTrue())
		})

		It("should keep primary values when both configs have settings", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			primaryContent := testConfigYML
			Expect(os.WriteFile(ymlPath, []byte(primaryContent), 0o644)).To(Succeed())

			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			secondaryContent := `
version: "2"
run:
  timeout: 10m
  go: "1.23"
linters:
  enable:
    - errcheck
`
			Expect(os.WriteFile(yamlPath, []byte(secondaryContent), 0o644)).To(Succeed())

			cfg, _, err := merger.MergeConfigs([]string{ymlPath, yamlPath})

			Expect(err).NotTo(HaveOccurred())
			// Primary timeout should be kept
			Expect(cfg.Run.Timeout).To(Equal("5m"))
			// Secondary Go version should be merged
			Expect(cfg.Run.Go).To(Equal("1.23"))
			// Both linters should be present
			Expect(cfg.Linters.Enable).To(ContainElements("gosec", "errcheck"))
		})

		It("should respect golangci-lint priority order", func() {
			// Create configs in different formats
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(ymlPath, []byte(`version: "2"
run:
  timeout: 5m
`), 0o644)).To(Succeed())

			tomlPath := filepath.Join(testDir, ".golangci.toml")
			Expect(os.WriteFile(tomlPath, []byte(`version = "2"
[run]
timeout = "10m"
`), 0o644)).To(Succeed())

			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			Expect(os.WriteFile(yamlPath, []byte(`version: "2"
run:
  timeout: 15m
`), 0o644)).To(Succeed())

			// Pass configs in random order
			_, result, err := merger.MergeConfigs([]string{tomlPath, yamlPath, ymlPath})

			Expect(err).NotTo(HaveOccurred())
			// .golangci.yml should be primary (highest priority)
			Expect(result.PrimaryConfig).To(Equal(ymlPath))
		})

		It("should merge disabled linters", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			primaryContent := `
version: "2"
linters:
  disable:
    - unused
`
			Expect(os.WriteFile(ymlPath, []byte(primaryContent), 0o644)).To(Succeed())

			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			secondaryContent := `
version: "2"
linters:
  disable:
    - gocyclo
`
			Expect(os.WriteFile(yamlPath, []byte(secondaryContent), 0o644)).To(Succeed())

			cfg, _, err := merger.MergeConfigs([]string{ymlPath, yamlPath})

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Linters.Disable).To(ContainElements("unused", "gocyclo"))
		})

		It("should merge issues settings", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			Expect(
				os.WriteFile(ymlPath, []byte("version: \"2\"\nlinters:\n  enable:\n    - gosec\n"), 0o644),
			).To(Succeed())

			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			Expect(
				os.WriteFile(
					yamlPath,
					[]byte("version: \"2\"\nissues:\n  max-issues-per-linter: 100\n  max-same-issues: 5\n"),
					0o644,
				),
			).To(Succeed())

			cfg, _, err := merger.MergeConfigs([]string{ymlPath, yamlPath})

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Issues.MaxIssuesPerLinter).To(Equal(100))
			Expect(cfg.Issues.MaxSameIssues).To(Equal(5))
		})
	})

	Context("SaveMergedConfig", func() {
		It("should save merged config to primary file", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			yamlPath := filepath.Join(testDir, ".golangci.yaml")

			Expect(
				os.WriteFile(ymlPath, []byte("version: \"2\"\nlinters:\n  enable:\n    - gosec\n"), 0o644),
			).To(Succeed())

			Expect(
				os.WriteFile(yamlPath, []byte("version: \"2\"\nlinters:\n  enable:\n    - errcheck\n"), 0o644),
			).To(Succeed())

			cfg, result, err := merger.MergeConfigs([]string{ymlPath, yamlPath})
			Expect(err).NotTo(HaveOccurred())

			// Save without removing secondary
			err = merger.SaveMergedConfig(cfg, result, false)
			Expect(err).NotTo(HaveOccurred())

			// Verify primary has merged content
			loaded, err := loader.LoadConfig(ymlPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded.Linters.Enable).To(ContainElements("gosec", "errcheck"))

			// Verify secondary still exists
			_, err = os.Stat(yamlPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should remove secondary configs when requested", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			yamlPath := filepath.Join(testDir, ".golangci.yaml")

			Expect(
				os.WriteFile(ymlPath, []byte("version: \"2\"\nlinters:\n  enable:\n    - gosec\n"), 0o644),
			).To(Succeed())

			Expect(
				os.WriteFile(yamlPath, []byte("version: \"2\"\nlinters:\n  enable:\n    - errcheck\n"), 0o644),
			).To(Succeed())

			cfg, result, err := merger.MergeConfigs([]string{ymlPath, yamlPath})
			Expect(err).NotTo(HaveOccurred())

			// Save and remove secondary
			err = merger.SaveMergedConfig(cfg, result, true)
			Expect(err).NotTo(HaveOccurred())

			// Verify secondary was removed
			_, err = os.Stat(yamlPath)
			Expect(os.IsNotExist(err)).To(BeTrue())

			// Verify it was tracked in result
			Expect(result.RemovedConfigs).To(ContainElement(yamlPath))
		})
	})

	Context("mergeFormattersConfig", func() {
		It("should merge formatters from secondary config", func() {
			ymlPath := filepath.Join(testDir, ".golangci.yml")
			primaryContent := `
version: "2"
formatters:
  enable:
    - gofmt
`
			Expect(os.WriteFile(ymlPath, []byte(primaryContent), 0o644)).To(Succeed())

			yamlPath := filepath.Join(testDir, ".golangci.yaml")
			secondaryContent := `
version: "2"
formatters:
  enable:
    - goimports
`
			Expect(os.WriteFile(yamlPath, []byte(secondaryContent), 0o644)).To(Succeed())

			cfg, _, err := merger.MergeConfigs([]string{ymlPath, yamlPath})

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Formatters.Enable).To(ContainElements("gofmt", "goimports"))
		})
	})
})

var _ = Describe("GetUniqueStrings", func() {
	It("should return unique sorted strings", func() {
		input := []string{"b", "a", "b", "c", "a"}
		result := config.GetUniqueStrings(input)
		Expect(result).To(Equal([]string{"a", "b", "c"}))
	})

	It("should return nil for empty input", func() {
		result := config.GetUniqueStrings([]string{})
		Expect(result).To(BeNil())
	})

	It("should handle single element", func() {
		result := config.GetUniqueStrings([]string{"a"})
		Expect(result).To(Equal([]string{"a"}))
	})
})
