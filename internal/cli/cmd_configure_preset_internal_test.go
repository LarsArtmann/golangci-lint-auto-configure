package cli

import (
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Multi-preset merge semantics: repeated --preset flags must UNION linter
// and formatter sets with dedup, in sorted order.
var _ = Describe("multi-preset merge", func() {
	var (
		logger       *log.Logger
		configLoader *config.Loader
		configPath   string
	)

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		configLoader = config.NewLoader(logger)
		configPath = filepath.Join(GinkgoT().TempDir(), ".golangci.yml")

		const minimalConfig = "version: \"2\"\nlinters:\n  enable:\n    - errcheck\n"
		Expect(os.WriteFile(configPath, []byte(minimalConfig), 0o644)).To(Succeed())
	})

	It("unions linters from two presets without duplicates", func() {
		_, linters, err := loadPresetConfig(logger, configLoader, configPath,
			[]string{"minimal", "performance"}, false)
		Expect(err).NotTo(HaveOccurred())

		minimal := constants.PresetLinters["minimal"]
		performance := constants.PresetLinters["performance"]

		expected := types.NewSet[types.LinterName]()
		for _, l := range minimal {
			expected.Add(l)
		}
		for _, l := range performance {
			expected.Add(l)
		}

		Expect(linters).To(Equal(types.ToSortedSlice(expected)))
		// Sorted, deduped: length matches the set union, not the concatenation.
		Expect(len(linters)).To(Equal(expected.Len()))
		Expect(len(linters)).To(BeNumerically("<", len(minimal)+len(performance)))
	})

	It("rejects an unknown preset with the valid list", func() {
		_, _, err := loadPresetConfig(logger, configLoader, configPath,
			[]string{"minimal", "bogus"}, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("bogus"))
		Expect(err.Error()).To(ContainSubstring(string(constants.ValidPresets)))
	})

	It("unions formatters across presets and dedupes extras against preset formatters", func() {
		cfg := &types.Config{}

		// "format" and "house" both carry the validated formatter quadruple;
		// the union must stay exactly 4 entries (no duplicates).
		applyPresetFormatters(logger, cfg, []string{"format", "house"}, nil)
		Expect(cfg.Formatters.Enable).To(HaveLen(4))

		expectedSet := types.NewSet[types.FormatterName]()
		for _, f := range constants.PresetFormatters["format"] {
			expectedSet.Add(f)
		}

		Expect(cfg.Formatters.Enable).To(Equal(types.ToSortedSlice(expectedSet)))

		// A preset formatter repeated as an extra must NOT appear twice.
		applyPresetFormatters(logger, cfg, []string{"format"},
			[]types.FormatterName{"gci", "swaggo"})
		Expect(cfg.Formatters.Enable).To(ContainElements("gci", "swaggo"))
		Expect(cfg.Formatters.Enable).To(HaveLen(5))
	})
})
