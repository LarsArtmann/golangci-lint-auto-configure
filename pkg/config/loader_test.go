package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Loader", func() {
	var (
		loader     *config.Loader
		testDir    string
		testConfig string
	)

	BeforeEach(func() {
		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		loader = config.NewLoader(logger)
		testDir = GinkgoT().TempDir()
		testConfig = filepath.Join(testDir, ".golangci.yml")
	})

	Context("LoadConfig", func() {
		It("should return error for non-existent file", func() {
			_, err := loader.LoadConfig("/non/existent/path.yml")
			Expect(err).To(HaveOccurred())
		})

		It("should load valid config file", func() {
			configContent := `
version: "1"
linters:
  enable:
    - gosec
    - errcheck
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			cfg, err := loader.LoadConfig(testConfig)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Version).To(Equal("1"))
			Expect(cfg.Linters.Enable).To(ContainElements("gosec", "errcheck"))
		})

		It("should parse all config sections", func() {
			configContent := `
version: "2"
run:
  timeout: 5m
  go: "1.21"
linters:
  enable:
    - gosec
  disable:
    - unused
output:
  formats:
    text:
      path: stdout
      colors: true
`
			Expect(os.WriteFile(testConfig, []byte(configContent), 0o644)).To(Succeed())

			cfg, err := loader.LoadConfig(testConfig)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Run.Timeout).To(Equal("5m"))
			Expect(cfg.Run.Go).To(Equal("1.21"))
			Expect(cfg.Output.Formats).ToNot(BeNil())
			Expect(cfg.Output.Formats).To(HaveKey("text"))
		})
	})

	Context("SaveConfig", func() {
		It("should save config to file", func() {
			cfg := &config.Config{
				Version: "1",
				Linters: config.LintersConfig{
					Enable: []string{"gosec", "errcheck"},
				},
			}

			err := loader.SaveConfig(cfg, testConfig)

			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("gosec"))
			Expect(string(content)).To(ContainSubstring("errcheck"))
		})

		It("should preserve complex config structure", func() {
			cfg := &config.Config{
				Version: "1",
				Run: config.RunConfig{
					Timeout: "5m",
					Go:      "1.21",
				},
				Linters: config.LintersConfig{
					Enable:  []string{"gosec", "errcheck"},
					Disable: []string{"unused"},
				},
			}

			err := loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			loaded, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded.Run.Timeout).To(Equal("5m"))
			Expect(loaded.Linters.Enable).To(HaveLen(2))
			Expect(loaded.Linters.Disable).To(HaveLen(1))
		})
	})

	Context("EnsureGitRepo", func() {
		It("should succeed when in a git repository", func() {
			err := loader.EnsureGitRepo(context.Background(), ".")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail when not in a git repository", func() {
			err := loader.EnsureGitRepo(context.Background(), "/tmp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not in a git repository"))
		})
	})

	Context("ConfigFormatSupport", func() {
		It("should load and save TOML config files", func() {
			tomlConfig := filepath.Join(testDir, ".golangci.toml")
			configContent := `
version = "2"
[linters]
enable = ["gosec", "errcheck"]
[run]
timeout = "5m"
`
			Expect(os.WriteFile(tomlConfig, []byte(configContent), 0o644)).To(Succeed())

			cfg, err := loader.LoadConfig(tomlConfig)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Version).To(Equal("2"))
			Expect(cfg.Linters.Enable).To(ContainElements("gosec", "errcheck"))
			Expect(cfg.Run.Timeout).To(Equal("5m"))
		})

		It("should load and save JSON config files", func() {
			jsonConfig := filepath.Join(testDir, ".golangci.json")
			configContent := `{
  "version": "2",
  "linters": {
    "enable": ["gosec", "errcheck"]
  },
  "run": {
    "timeout": "5m"
  }
}`
			Expect(os.WriteFile(jsonConfig, []byte(configContent), 0o644)).To(Succeed())

			cfg, err := loader.LoadConfig(jsonConfig)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Version).To(Equal("2"))
			Expect(cfg.Linters.Enable).To(ContainElements("gosec", "errcheck"))
		})

		It("should save config in TOML format", func() {
			tomlConfig := filepath.Join(testDir, ".golangci.toml")
			cfg := &config.Config{
				Version: "2",
				Linters: config.LintersConfig{
					Enable: []string{"gosec"},
				},
			}

			err := loader.SaveConfig(cfg, tomlConfig)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(tomlConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("version"))
			Expect(string(content)).To(ContainSubstring("gosec"))
		})
	})

	Context("FindConfigFile", func() {
		It("should find config files with different extensions", func() {
			testCases := []struct {
				ext      string
				filename string
			}{
				{"yml", ".golangci.yml"},
				{"yaml", ".golangci.yaml"},
				{"toml", ".golangci.toml"},
				{"json", ".golangci.json"},
			}

			for _, testCase := range testCases {
				Expect(
					os.WriteFile(filepath.Join(testDir, testCase.filename), []byte("version: 1"), 0o644),
				).To(Succeed())

				found, err := loader.FindConfigFile(testDir)

				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(Equal(filepath.Join(testDir, testCase.filename)))

				Expect(os.Remove(filepath.Join(testDir, testCase.filename))).To(Succeed())
			}
		})

		It("should return error when no config file exists", func() {
			_, err := loader.FindConfigFile(testDir)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("FindAllConfigFiles", func() {
		It("should find all config files", func() {
			Expect(os.WriteFile(filepath.Join(testDir, ".golangci.yml"), []byte("version: 1"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(testDir, ".golangci.yaml"), []byte("version: 1"), 0o644)).To(Succeed())

			found := loader.FindAllConfigFiles(testDir)

			Expect(found).To(HaveLen(2))
			Expect(found).To(ContainElement(filepath.Join(testDir, ".golangci.yml")))
			Expect(found).To(ContainElement(filepath.Join(testDir, ".golangci.yaml")))

			Expect(os.Remove(filepath.Join(testDir, ".golangci.yml"))).To(Succeed())
			Expect(os.Remove(filepath.Join(testDir, ".golangci.yaml"))).To(Succeed())
		})

		It("should return empty when no config files exist", func() {
			found := loader.FindAllConfigFiles(testDir)
			Expect(found).To(BeEmpty())
		})
	})

	Context("HasMultipleConfigFiles", func() {
		It("should return false when only one config exists", func() {
			Expect(os.WriteFile(filepath.Join(testDir, ".golangci.yml"), []byte("version: 1"), 0o644)).To(Succeed())

			result := loader.HasMultipleConfigFiles(testDir)

			Expect(result).To(BeFalse())

			Expect(os.Remove(filepath.Join(testDir, ".golangci.yml"))).To(Succeed())
		})

		It("should return true and log warning when multiple configs exist", func() {
			Expect(os.WriteFile(filepath.Join(testDir, ".golangci.yml"), []byte("version: 1"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(testDir, ".golangci.yaml"), []byte("version: 1"), 0o644)).To(Succeed())

			result := loader.HasMultipleConfigFiles(testDir)

			Expect(result).To(BeTrue())

			Expect(os.Remove(filepath.Join(testDir, ".golangci.yml"))).To(Succeed())
			Expect(os.Remove(filepath.Join(testDir, ".golangci.yaml"))).To(Succeed())
		})
	})

	Context("ValidateConfig", func() {
		It("should return error for empty timeout", func() {
			cfg := &config.Config{}
			errs := loader.ValidateConfig(cfg)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs[0].Error()).To(ContainSubstring("validation"))
		})

		It("should validate valid config", func() {
			cfg := &config.Config{
				Version: "2",
				Run: config.RunConfig{
					Timeout: "5m",
				},
			}
			errors := loader.ValidateConfig(cfg)
			Expect(errors).To(BeEmpty())
		})
	})

	Context("GetLintersEnabled", func() {
		It("should return enabled linters", func() {
			cfg := &config.Config{
				Linters: config.LintersConfig{
					Enable: []string{"gosec", "errcheck", "staticcheck"},
				},
			}

			enabled := loader.GetLintersEnabled(cfg)

			Expect(enabled).To(HaveLen(3))
			Expect(enabled).To(ContainElements("gosec", "errcheck", "staticcheck"))
		})
	})

	Context("GetLintersDisabled", func() {
		It("should return disabled linters", func() {
			cfg := &config.Config{
				Linters: config.LintersConfig{
					Disable: []string{"unused", "gocyclo"},
				},
			}

			disabled := loader.GetLintersDisabled(cfg)

			Expect(disabled).To(HaveLen(2))
			Expect(disabled).To(ContainElements("unused", "gocyclo"))
		})
	})
})
