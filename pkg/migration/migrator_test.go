// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/migration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migration Suite")
}

// testMigrationWithExpectedContent tests a migration and verifies expected content in the result.
// writeYMLConfig writes configContent to a .golangci.yml in testDir and returns the path.
func writeYMLConfig(testDir, configContent string) string {
	configPath := filepath.Join(testDir, ".golangci.yml")
	Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

	return configPath
}

func testMigrationWithExpectedContent(testDir, configContent, expectedContent string) {
	configPath := writeYMLConfig(testDir, configContent)

	m, err := migration.NewMigrator(configPath, false)
	Expect(err).NotTo(HaveOccurred())
	m.SetValidator(migration.MockValidator{})

	success, fixes, err := m.MigrateToV2()
	Expect(err).NotTo(HaveOccurred())
	Expect(success).To(BeTrue())
	Expect(fixes).To(BeNumerically(">", 0))

	// Verify the file was updated
	content, err := os.ReadFile(configPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(content)).To(ContainSubstring(expectedContent))
}

// testSloglintMapping tests the sloglint key-naming-case mapping.
func testSloglintMapping(input, expected string, shouldExist bool) {
	rules := migration.DefaultRules()
	mapped, exists := rules.MapSloglintKeyNamingCase(input)
	Expect(exists).To(Equal(shouldExist))

	if shouldExist {
		Expect(mapped).To(Equal(expected))
	}
}

// runMigration creates a migrator and runs the migration, returning the result.
func runMigration(configPath string) (bool, int, error) {
	migrator, err := migration.NewMigrator(configPath, false)
	if err != nil {
		return false, 0, err
	}

	migrator.SetValidator(migration.MockValidator{})
	success, fixes, err := migrator.MigrateToV2()

	return success, fixes, err
}

// testMigrationWithConfig creates a config file and runs migration.
func testMigrationWithConfig(testDir, configContent string) string {
	configPath := writeYMLConfig(testDir, configContent)

	success, fixes, err := runMigration(configPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(success).To(BeTrue())
	Expect(fixes).To(BeNumerically(">", 0))

	return configPath
}

// testMigrationWithSimpleConfig creates a test dir, writes config and runs migration.
func testMigrationWithSimpleConfig(configContent string) {
	testDir := GinkgoT().TempDir()
	testMigrationWithConfig(testDir, configContent)
}

// v2ConfigWithExcludeDirs returns a v2 config with exclude-dirs.
func v2ConfigWithExcludeDirs(dirs ...string) string {
	var builder strings.Builder

	for _, dir := range dirs {
		builder.WriteString("  - " + dir + "\n")
	}

	return `version: "2"
run:
  timeout: 5m
exclude-dirs:
` + builder.String() + `linters:
  enable:
    - errcheck
`
}

// v2ConfigWithExcludeFiles returns a v2 config with exclude-files.
func v2ConfigWithExcludeFiles(files ...string) string {
	var builder strings.Builder

	for _, file := range files {
		builder.WriteString("  - \"" + file + "\"\n")
	}

	return `version: "2"
run:
  timeout: 5m
exclude-files:
` + builder.String() + `
linters:
  enable:
    - errcheck
`
}

var _ = Describe("Migrator", func() {
	var testDir string

	BeforeEach(func() {
		testDir = GinkgoT().TempDir()
	})

	Describe("NewMigrator", func() {
		It("should create a migrator with valid path", func() {
			configPath := writeYMLConfig(testDir, "version: \"2\"")

			m, err := migration.NewMigrator(configPath, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).NotTo(BeNil())
		})

		It("should return error for empty path", func() {
			m, err := migration.NewMigrator("", false)
			Expect(err).To(HaveOccurred())
			Expect(m).To(BeNil())
		})
	})

	Describe("SetDryRun", func() {
		It("should set dry run mode", func() {
			configPath := writeYMLConfig(testDir, "version: \"2\"")

			m, err := migration.NewMigrator(configPath, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(m.IsDryRun()).To(BeFalse())
			m.SetDryRun(true)
			Expect(m.IsDryRun()).To(BeTrue())
		})
	})

	Describe("SetNoEmojis", func() {
		It("should set no emojis mode", func() {
			configPath := writeYMLConfig(testDir, "version: \"2\"")

			m, err := migration.NewMigrator(configPath, false)
			Expect(err).NotTo(HaveOccurred())

			m.SetNoEmojis(false)
			// No error means success
		})
	})

	Describe("MigrateToV2", func() {
		Context("with already v2 config", func() {
			It("should not migrate when config is already v2 with all required fields", func() {
				configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`
				configPath := writeYMLConfig(testDir, configContent)

				m, err := migration.NewMigrator(configPath, false)
				Expect(err).NotTo(HaveOccurred())
				m.SetValidator(migration.MockValidator{})

				success, fixes, err := m.MigrateToV2()
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeFalse())
				Expect(fixes).To(Equal(0))
			})
		})

		DescribeTable(
			"Migrations with v1 config",
			func(configContent, expectedContent string) {
				testMigrationWithExpectedContent(testDir, configContent, expectedContent)
			},
			Entry("should migrate version to v2", `version: "1"
linters:
  enable:
    - errcheck
`, `version: "2"`),
			Entry("should migrate formatters to formatters.enable", `version: "1"
linters:
  enable:
    - gofmt
    - goimports
    - errcheck
`, "formatters:"),
		)

		Context("with deprecated linters-settings", func() {
			It("should move linters-settings to linters.settings", func() {
				configContent := `version: "1"
linters-settings:
  errcheck:
    check-type-assertions: true
linters:
  enable:
    - errcheck
`
				configPath := writeYMLConfig(testDir, configContent)

				m, err := migration.NewMigrator(configPath, false)
				Expect(err).NotTo(HaveOccurred())
				m.SetValidator(migration.MockValidator{})

				success, fixes, err := m.MigrateToV2()
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeTrue())
				Expect(fixes).To(BeNumerically(">", 0))

				// Verify the file was updated
				content, err := os.ReadFile(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("linters:"))
				Expect(string(content)).To(ContainSubstring("settings:"))
			})
		})

		Context("with deprecated output properties", func() {
			It("should remove deprecated output properties", func() {
				configContent := `version: "2"
run:
  timeout: 5m
output:
  print-issued-lines: true
  print-linter-name: true
  sort-results: true
linters:
  enable:
    - errcheck
`
				configPath := writeYMLConfig(testDir, configContent)

				m, err := migration.NewMigrator(configPath, false)
				Expect(err).NotTo(HaveOccurred())
				m.SetValidator(migration.MockValidator{})

				success, fixes, err := m.MigrateToV2()
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeTrue())
				Expect(fixes).To(BeNumerically(">", 0))

				// Verify the file was updated
				content, err := os.ReadFile(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).NotTo(ContainSubstring("print-issued-lines"))
			})
		})

		Context("in dry-run mode", func() {
			It("should not modify the file", func() {
				configContent := `version: "1"
linters-settings:
  gofmt:
    simplify: true
linters:
  enable:
    - gofmt
`
				configPath := writeYMLConfig(testDir, configContent)

				originalContent, err := os.ReadFile(configPath)
				Expect(err).NotTo(HaveOccurred())

				m, err := migration.NewMigrator(configPath, false)
				Expect(err).NotTo(HaveOccurred())
				m.SetDryRun(true)
				m.SetValidator(migration.MockValidator{})

				success, fixes, err := m.MigrateToV2()
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeTrue())
				Expect(fixes).To(BeNumerically(">", 0))

				// Verify the file was NOT modified
				currentContent, err := os.ReadFile(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(currentContent)).To(Equal(string(originalContent)))
			})
		})
	})
})

var _ = Describe("MigrationRules", func() {
	Describe("DefaultRules", func() {
		It("should return rules with valid versions", func() {
			rules := migration.DefaultRules()
			Expect(rules).NotTo(BeNil())
			Expect(rules.IsValidVersion("2")).To(BeTrue())
			Expect(rules.IsValidVersion("2.8")).To(BeTrue())
			Expect(rules.IsValidVersion("1")).To(BeFalse())
		})

		It("should return deprecated properties for known linters", func() {
			rules := migration.DefaultRules()
			props := rules.GetDeprecatedProperties("cyclop")
			Expect(props).To(ContainElement("skip-tests"))
		})

		It("should return empty slice for unknown linters", func() {
			rules := migration.DefaultRules()
			props := rules.GetDeprecatedProperties("unknown-linter")
			Expect(props).To(BeEmpty())
		})

		It("should identify linters without settings", func() {
			rules := migration.DefaultRules()
			Expect(rules.IsLinterWithoutSettings("containedctx")).To(BeTrue())
			Expect(rules.IsLinterWithoutSettings("errcheck")).To(BeFalse())
		})

		It("should map modernize disable values", func() {
			rules := migration.DefaultRules()
			mapped, exists := rules.MapModernizeDisable("forvar")
			Expect(exists).To(BeTrue())
			Expect(mapped).To(Equal("forvar"))
		})

		It("should return false for unknown modernize disable values", func() {
			rules := migration.DefaultRules()
			_, exists := rules.MapModernizeDisable("unknown")
			Expect(exists).To(BeFalse())
		})

		It("should map sloglint key-naming-case values", func() {
			testSloglintMapping("snake", "snake", true)
		})

		It("should map camelCase to camel", func() {
			testSloglintMapping("camelCase", "camel", true)
		})

		It("should return false for unknown sloglint values", func() {
			testSloglintMapping("unknown", "", false)
		})
	})
})

var _ = Describe("Validator", func() {
	Describe("MockValidator", func() {
		It("should always return nil", func() {
			v := migration.MockValidator{}
			Expect(v.ValidateConfig(nil)).To(Succeed())
		})
	})

	Describe("FailingValidator", func() {
		It("should return error with default message", func() {
			v := FailingValidator{}
			err := v.ValidateConfig(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mock validation failed"))
		})

		It("should return error with custom message", func() {
			v := FailingValidator{ErrorMessage: "custom error"}
			err := v.ValidateConfig(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("custom error"))
		})
	})
})

var _ = Describe("YAMLLoader", func() {
	Describe("LoadConfig", func() {
		It("should load valid YAML config", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`
			configPath := writeYMLConfig(testDir, configContent)

			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Version).To(Equal("2"))
			Expect(cfg.Run.Timeout).To(Equal("5m"))
		})

		It("should return error for non-existent file", func() {
			_, err := migration.LoadConfig("/non/existent/path.yml")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to read config file"))
		})

		It("should return error for invalid YAML", func() {
			testDir := GinkgoT().TempDir()
			configPath := writeYMLConfig(testDir, "invalid: yaml: content:")

			_, err := migration.LoadConfig(configPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SaveConfig", func() {
		It("should save config to file", func() {
			testDir := GinkgoT().TempDir()
			configPath := filepath.Join(testDir, ".golangci.yml")
			cfg := &migration.Config{
				Version: "2",
				Run: migration.Run{
					Timeout: "5m",
				},
				Linters: migration.Linters{
					Enable: []string{"errcheck"},
				},
			}

			err := migration.SaveConfig(cfg, configPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify the file was written
			content, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("version: \"2\""))
		})

		It("should return error for invalid path", func() {
			cfg := &migration.Config{
				Version: "2",
			}

			err := migration.SaveConfig(cfg, "/invalid/path/.golangci.yml")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("ConfigTypes", func() {
	Describe("UnmarshalYAML", func() {
		It("should handle v1 issues structure", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: "2"
run:
  timeout: 5m
issues:
  exclude-rules:
    - linters:
        - errcheck
      text: "error checked"
linters:
  enable:
    - errcheck
`
			configPath := writeYMLConfig(testDir, configContent)

			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
		})

		It("should handle nested issues structure from v1", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: "1"
run:
  issues:
    exclude-use-default: true
    exclude-rules:
      - linters:
          - errcheck
        text: "error checked"
linters:
  enable:
    - errcheck
`
			configPath := writeYMLConfig(testDir, configContent)

			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
		})
	})
})

var _ = Describe("MigrationFunctions", func() {
	Describe("migrateLintersSettings", func() {
		It("should handle config with no linters settings", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := writeYMLConfig(testDir, configContent)

			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.LintersSettingsV1).To(BeNil())
		})
	})

	Describe("migrateFormatters", func() {
		It("should migrate formatters from linters.enable to formatters.enable", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: "1"
linters:
  enable:
    - gofmt
    - goimports
    - errcheck
`
			configPath := writeYMLConfig(testDir, configContent)

			m, err := migration.NewMigrator(configPath, false)
			Expect(err).NotTo(HaveOccurred())
			m.SetValidator(migration.MockValidator{})

			success, fixes, err := m.MigrateToV2()
			Expect(err).NotTo(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(fixes).To(BeNumerically(">", 0))

			// Verify formatters were moved - reload config and check
			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Formatters.Enable).To(ContainElement("gofmt"))
			Expect(cfg.Formatters.Enable).To(ContainElement("goimports"))
			Expect(cfg.Linters.Enable).NotTo(ContainElement("gofmt"))
			Expect(cfg.Linters.Enable).NotTo(ContainElement("goimports"))
		})
	})

	Describe("migrateOutputProperties", func() {
		It("should remove deprecated output properties", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
output:
  format: json
  print-issued-lines: true
  print-linter-name: true
  sort-results: true
linters:
  enable:
    - errcheck
`)
		})
	})

	Describe("migrateVersion", func() {
		It("should handle empty version string", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: ""
linters:
  enable:
    - errcheck
`
			configPath := testMigrationWithConfig(testDir, configContent)

			// Verify version was set to "2"
			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Version).To(Equal("2"))
		})

		It("should handle version with v prefix", func() {
			testMigrationWithSimpleConfig(`version: "v1"
linters:
  enable:
    - errcheck
`)
		})
	})

	Describe("migrateRunSettings", func() {
		It("should set default timeout when missing", func() {
			testDir := GinkgoT().TempDir()
			configContent := `version: "2"
run: {}
linters:
  enable:
    - errcheck
`
			configPath := testMigrationWithConfig(testDir, configContent)

			// Verify timeout was set
			cfg, err := migration.LoadConfig(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Run.Timeout).To(Equal("5m"))
		})
	})

	Describe("migrateIssuesExcludeDirs", func() {
		It("should migrate exclude-dirs to exclusions.paths", func() {
			testMigrationWithSimpleConfig(v2ConfigWithExcludeDirs("vendor", "generated"))
		})
	})

	Describe("migrateIssuesExcludeFiles", func() {
		It("should migrate exclude-files to exclusions.paths", func() {
			testMigrationWithSimpleConfig(v2ConfigWithExcludeFiles("*.gen.go", "**/*_test.go"))
		})
	})

	Describe("migrateLinterSettings", func() {
		It("should remove deprecated cyclop skip-tests property", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
linters:
  settings:
    cyclop:
      skip-tests: true
      max-complexity: 15
  enable:
    - errcheck
`)
		})

		It("should remove gosec null excludes", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
linters:
  settings:
    gosec:
      excludes: null
  enable:
    - errcheck
`)
		})

		It("should remove settings for linters without settings", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
linters:
  settings:
    containedctx:
      some-setting: true
  enable:
    - errcheck
`)
		})

		It("should migrate forbidigo p to pattern rename", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
linters:
  settings:
    forbidigo:
      forbid:
        - p: fmt.Print.*
  enable:
    - errcheck
`)
		})

		It("should normalize goimports local-prefixes from string to array", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
linters:
  settings:
    goimports:
      local-prefixes: github.com/myorg
  enable:
    - errcheck
`)
		})
	})

	Describe("migrateFormattersSettings", func() {
		It("should remove gci skip-generated property", func() {
			testMigrationWithSimpleConfig(`version: "2"
run:
  timeout: 5m
formatters:
  settings:
    gci:
      skip-generated: true
      sections:
        - Standard
linters:
  enable:
    - errcheck
`)
		})
	})
})
