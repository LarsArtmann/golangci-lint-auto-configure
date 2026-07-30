package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "types.Config Suite")
}

func writeTestConfigFile(dir, filename string) {
	Expect(os.WriteFile(filepath.Join(dir, filename), []byte("version: 1"), 0o644)).To(Succeed())
}

func writeConfigContent(path, content string) {
	Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
}

func removeTestConfigFile(dir, filename string) {
	Expect(os.Remove(filepath.Join(dir, filename))).To(Succeed())
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
			writeConfigContent(testConfig, configContent)

			cfg, err := loader.LoadConfig(testConfig)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Version).To(Equal(types.Version("1")))
			Expect(cfg.Linters.Enable).To(ContainElements(types.LinterName("gosec"), types.LinterName("errcheck")))
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
			writeConfigContent(testConfig, configContent)

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
			cfg := &types.Config{
				Version: "1",
				Linters: types.LintersConfig{
					Enable: []types.LinterName{"gosec", "errcheck"},
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
			cfg := &types.Config{
				Version: "1",
				Run: types.RunConfig{
					Timeout: "5m",
					Go:      "1.21",
				},
				Linters: types.LintersConfig{
					Enable:  []types.LinterName{"gosec", "errcheck"},
					Disable: []types.LinterName{"unused"},
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

		It("should preserve 2-space indentation from existing file", func() {
			original := "version: \"2\"\nlinters:\n  enable:\n    - gosec\n  disable:\n    - unused\n"
			Expect(os.WriteFile(testConfig, []byte(original), 0o644)).To(Succeed())

			cfg := &types.Config{
				Version: "2",
				Linters: types.LintersConfig{
					Enable:  []types.LinterName{"gosec", "errcheck"},
					Disable: []types.LinterName{"unused"},
				},
			}
			err := loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("\n  enable:\n"))
		})

		It("should preserve 4-space indentation from existing file", func() {
			original := "version: \"2\"\nlinters:\n    enable:\n        - gosec\n    disable:\n        - unused\n"
			Expect(os.WriteFile(testConfig, []byte(original), 0o644)).To(Succeed())

			cfg := &types.Config{
				Version: "2",
				Linters: types.LintersConfig{
					Enable:  []types.LinterName{"gosec", "errcheck"},
					Disable: []types.LinterName{"unused"},
				},
			}
			err := loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("\n    enable:\n"))
		})

		It("should default to 2-space indent for new file", func() {
			cfg := &types.Config{
				Version: "2",
				Linters: types.LintersConfig{
					Enable: []types.LinterName{"gosec"},
				},
			}
			err := loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("\n  enable:\n"))
			Expect(string(content)).NotTo(ContainSubstring("\n    enable:\n"))
		})
	})

	Context("detectYAMLIndent", func() {
		It("should detect 2-space indentation", func() {
			data := []byte("version: \"2\"\nlinters:\n  enable:\n    - gosec\n")
			Expect(detectYAMLIndent(data)).To(Equal(2))
		})

		It("should detect 4-space indentation", func() {
			data := []byte("version: \"2\"\nlinters:\n    enable:\n        - gosec\n")
			Expect(detectYAMLIndent(data)).To(Equal(4))
		})

		It("should default to 2 for root-only YAML", func() {
			data := []byte("version: \"2\"\n")
			Expect(detectYAMLIndent(data)).To(Equal(2))
		})

		It("should skip comments and document markers", func() {
			data := []byte("# comment\n---\nversion: \"2\"\nlinters:\n  enable:\n    - gosec\n")
			Expect(detectYAMLIndent(data)).To(Equal(2))
		})

		It("should default to 2 for empty data", func() {
			Expect(detectYAMLIndent([]byte(""))).To(Equal(2))
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
			Expect(cfg.Version).To(Equal(types.ConfigVersionV2))
			Expect(cfg.Linters.Enable).To(ContainElements(types.LinterName("gosec"), types.LinterName("errcheck")))
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
			Expect(cfg.Version).To(Equal(types.ConfigVersionV2))
			Expect(cfg.Linters.Enable).To(ContainElements(types.LinterName("gosec"), types.LinterName("errcheck")))
		})

		It("should save config in TOML format", func() {
			tomlConfig := filepath.Join(testDir, ".golangci.toml")
			cfg := &types.Config{
				Version: "2",
				Linters: types.LintersConfig{
					Enable: []types.LinterName{"gosec"},
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
			writeTestConfigFile(testDir, ".golangci.yml")
			writeTestConfigFile(testDir, ".golangci.yaml")

			found := loader.FindAllConfigFiles(testDir)

			Expect(found).To(HaveLen(2))
			Expect(found).To(ContainElement(filepath.Join(testDir, ".golangci.yml")))
			Expect(found).To(ContainElement(filepath.Join(testDir, ".golangci.yaml")))

			removeTestConfigFile(testDir, ".golangci.yml")
			removeTestConfigFile(testDir, ".golangci.yaml")
		})

		It("should return empty when no config files exist", func() {
			found := loader.FindAllConfigFiles(testDir)
			Expect(found).To(BeEmpty())
		})
	})

	Context("HasMultipleConfigFiles", func() {
		It("should return false when only one config exists", func() {
			writeTestConfigFile(testDir, ".golangci.yml")

			result := loader.HasMultipleConfigFiles(testDir)

			Expect(result).To(BeFalse())

			removeTestConfigFile(testDir, ".golangci.yml")
		})

		It("should return true and log warning when multiple configs exist", func() {
			writeTestConfigFile(testDir, ".golangci.yml")
			writeTestConfigFile(testDir, ".golangci.yaml")

			result := loader.HasMultipleConfigFiles(testDir)

			Expect(result).To(BeTrue())

			removeTestConfigFile(testDir, ".golangci.yml")
			removeTestConfigFile(testDir, ".golangci.yaml")
		})
	})

	Context("ValidateConfig", func() {
		It("should return error for empty timeout", func() {
			cfg := &types.Config{}
			errs := loader.ValidateConfig(cfg)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs[0].Error()).To(ContainSubstring("validation"))
		})

		It("should validate valid config", func() {
			cfg := &types.Config{
				Version: "2",
				Run: types.RunConfig{
					Timeout: "5m",
				},
			}
			errors := loader.ValidateConfig(cfg)
			Expect(errors).To(BeEmpty())
		})
	})

	Context("RoundtripFidelity", func() {
		It("should preserve top-level linters-settings (v1) via auto-migration", func() {
			configContent := `version: "1"
linters:
  enable:
    - gosec
    - depguard
linters-settings:
  depguard:
    rules:
      main:
        allow:
          - $gostd
          - github.com/myproject
`
			writeConfigContent(testConfig, configContent)

			cfg, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.LintersSettingsV1).To(BeNil(), "v1 linters-settings should be migrated, not kept")
			Expect(cfg.Linters.Settings).To(HaveKey("depguard"))

			depguardCfg, ok := cfg.Linters.Settings["depguard"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(depguardCfg).To(HaveKey("rules"))

			err = loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			reloaded, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(reloaded.Linters.Settings).To(HaveKey("depguard"))
		})

		It("should preserve nested linters.settings (v2) through load-save-load", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - depguard
  settings:
    depguard:
      rules:
        main:
          allow:
            - $gostd
            - github.com/myproject
`
			writeConfigContent(testConfig, configContent)

			cfg, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.Linters.Settings).To(HaveKey("depguard"))

			err = loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			reloaded, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(reloaded.Linters.Settings).To(HaveKey("depguard"))

			depguardCfg, ok := reloaded.Linters.Settings["depguard"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(depguardCfg).To(HaveKey("rules"))
		})

		It("should not add linters-settings when config has none", func() {
			configContent := `version: "2"
linters:
  enable:
    - gosec
    - errcheck
`
			writeConfigContent(testConfig, configContent)

			cfg, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			err = loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(testConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).NotTo(ContainSubstring("linters-settings"))
		})

		It("should preserve full realistic config with multiple linter settings", func() {
			configContent := `version: "2"
run:
  timeout: 5m
  go: "1.26"
linters:
  enable:
    - gosec
    - errcheck
    - depguard
    - gomodguard
  settings:
    depguard:
      rules:
        main:
          allow:
            - $gostd
            - github.com/myproject
    gomodguard:
      blocked:
        - modules:
            - github.com/pkg/errors:
                recommendations:
                  - fmt
                  - errors
issues:
  max-issues-per-linter: 50
  max-same-issues: 10
output:
  formats:
    text:
      path: stdout
      colors: true
`
			writeConfigContent(testConfig, configContent)

			cfg, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.Linters.Settings).To(HaveKey("depguard"))
			Expect(cfg.Linters.Settings).To(HaveKey("gomodguard"))
			Expect(cfg.Run.Timeout).To(Equal("5m"))
			Expect(cfg.Output.Formats).To(HaveKey("text"))

			err = loader.SaveConfig(cfg, testConfig)
			Expect(err).NotTo(HaveOccurred())

			reloaded, err := loader.LoadConfig(testConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(reloaded.Linters.Settings).To(HaveKey("depguard"))
			Expect(reloaded.Linters.Settings).To(HaveKey("gomodguard"))
			Expect(reloaded.Run.Timeout).To(Equal("5m"))
		})
	})

	Context("GetLintersEnabled", func() {
		It("should return enabled linters", func() {
			cfg := &types.Config{
				Linters: types.LintersConfig{
					Enable: []types.LinterName{"gosec", "errcheck", "staticcheck"},
				},
			}

			enabled := loader.GetLintersEnabled(cfg)

			Expect(enabled).To(HaveLen(3))
			Expect(
				enabled,
			).To(ContainElements(types.LinterName("gosec"), types.LinterName("errcheck"), types.LinterName("staticcheck")))
		})
	})

	Context("GetLintersDisabled", func() {
		It("should return disabled linters", func() {
			cfg := &types.Config{
				Linters: types.LintersConfig{
					Disable: []types.LinterName{"unused", "gocyclo"},
				},
			}

			disabled := loader.GetLintersDisabled(cfg)

			Expect(disabled).To(HaveLen(2))
			Expect(disabled).To(ContainElements(types.LinterName("unused"), types.LinterName("gocyclo")))
		})
	})
})
