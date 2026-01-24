package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Loader", func() {
	var loader *config.Loader

	BeforeEach(func() {
		logger := log.NewWithOptions(log.Options{Level: log.ErrorLevel})
		loader = config.NewLoader(logger)
	})

	Context("LoadConfig", func() {
		It("should return error for non-existent file", func() {
			_, err := loader.LoadConfig("/non/existent/path.yml")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("FindConfigFile", func() {
		It("should return error when no config file exists", func() {
			_, err := loader.FindConfigFile("/tmp/nonexistent")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ValidateConfig", func() {
		It("should validate empty config", func() {
			cfg := &config.Config{}
			errors := loader.ValidateConfig(cfg)
			Expect(errors).To(BeEmpty())
		})
	})
})
