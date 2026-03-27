// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/migration"
)

func TestMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migration Suite")
}

// testMigrationWithExpectedContent tests a migration and verifies expected content in the result.
func testMigrationWithExpectedContent(testDir, configContent string, expectedContent string) {
	configPath := filepath.Join(testDir, ".golangci.yml")
	Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

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

var _ = Describe("Migrator", func() {
	var testDir string

	BeforeEach(func() {
		testDir = GinkgoT().TempDir()
	})

	Describe("NewMigrator", func() {
		It("should create a migrator with valid path", func() {
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte("version: \"2\""), 0o644)).To(Succeed())

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
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte("version: \"2\""), 0o644)).To(Succeed())

			m, err := migration.NewMigrator(configPath, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(m.IsDryRun()).To(BeFalse())
			m.SetDryRun(true)
			Expect(m.IsDryRun()).To(BeTrue())
		})
	})

	Describe("SetNoEmojis", func() {
		It("should set no emojis mode", func() {
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte("version: \"2\""), 0o644)).To(Succeed())

			m, err := migration.NewMigrator(configPath, false)
			Expect(err).NotTo(HaveOccurred())

			m.SetNoEmojis(false)
			// No error means success
		})
	})

	Describe("MigrateToV2", func() {
		Context("with already v2 config", func() {
			It("should not migrate when config is already v2 with all required fields", func() {
				configPath := filepath.Join(testDir, ".golangci.yml")
				configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`
				Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

				m, err := migration.NewMigrator(configPath, false)
				Expect(err).NotTo(HaveOccurred())
				m.SetValidator(migration.MockValidator{})

				success, fixes, err := m.MigrateToV2()
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeFalse())
				Expect(fixes).To(Equal(0))
			})
		})

		Context("with v1 config", func() {
			It("should migrate version to v2", func() {
				configContent := `version: "1"
linters:
  enable:
    - errcheck
`
				testMigrationWithExpectedContent(testDir, configContent, `version: "2"`)
			})
		})

		Context("with formatters in linters.enable", func() {
			It("should migrate formatters to formatters.enable", func() {
				configContent := `version: "1"
linters:
  enable:
    - gofmt
    - goimports
    - errcheck
`
				testMigrationWithExpectedContent(testDir, configContent, "formatters:")
			})
		})

		Context("with deprecated linters-settings", func() {
			It("should move linters-settings to linters.settings", func() {
				configPath := filepath.Join(testDir, ".golangci.yml")
				configContent := `version: "1"
linters-settings:
  errcheck:
    check-type-assertions: true
linters:
  enable:
    - errcheck
`
				Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

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
				configPath := filepath.Join(testDir, ".golangci.yml")
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
				Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

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
				configPath := filepath.Join(testDir, ".golangci.yml")
				configContent := `version: "1"
linters-settings:
  gofmt:
    simplify: true
linters:
  enable:
    - gofmt
`
				Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

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
			v := migration.FailingValidator{}
			err := v.ValidateConfig(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mock validation failed"))
		})

		It("should return error with custom message", func() {
			v := migration.FailingValidator{ErrorMessage: "custom error"}
			err := v.ValidateConfig(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("custom error"))
		})
	})
})
