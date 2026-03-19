package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
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

	Context("FindConfigFile", func() {
		It("should find config files with different extensions", func() {
			testCases := []struct {
				ext      string
				filename string
			}{
				{"yml", ".golangci.yml"},
				{"yaml", ".golangci.yaml"},
			}

			for _, tc := range testCases {
				Expect(os.WriteFile(filepath.Join(testDir, tc.filename), []byte("version: 1"), 0o644)).To(Succeed())

				found, err := loader.FindConfigFile(testDir)

				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(Equal(filepath.Join(testDir, tc.filename)))

				Expect(os.Remove(filepath.Join(testDir, tc.filename))).To(Succeed())
			}
		})

		It("should return error when no config file exists", func() {
			_, err := loader.FindConfigFile(testDir)
			Expect(err).To(HaveOccurred())
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
