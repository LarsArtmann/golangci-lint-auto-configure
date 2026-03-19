package cli

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/spf13/cobra"
)

// newValidateCommand creates the validate command.
func newValidateCommand(
	logger *log.Logger,
	configLoader *config.Loader,
) *cobra.Command {
	var skipGolangciLint bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate golangci-lint configuration",
		Long: `Validates the golangci-lint configuration file.

This command performs two levels of validation:
1. Basic YAML parsing and structure validation
2. Schema validation using golangci-lint config verify

Use --skip-golangci-lint to skip the schema validation (faster).
Use --verbose to see detailed validation output.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile := configPath
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf("failed to find config file: %w", err)
				}
			}

			logger.Infof("Validating configuration: %s", configFile)

			// Level 1: Basic load and validate
			cfg, err := configLoader.LoadConfig(configFile)
			if err != nil {
				logger.Errorf("❌ Failed to load configuration")

				return fmt.Errorf("failed to load config: %w", err)
			}

			logger.Infof("✓ Basic structure valid")

			// Run internal validation
			validationErrors := configLoader.ValidateConfig(cfg)
			if len(validationErrors) > 0 {
				logger.Errorf("❌ Internal validation failed:")

				for _, err := range validationErrors {
					logger.Errorf("  - %v", err)
				}

				return fmt.Errorf("configuration validation failed with %d errors", len(validationErrors))
			}

			logger.Infof("✓ Internal validation passed")

			// Level 2: Schema validation via golangci-lint
			if !skipGolangciLint {
				logger.Infof("Running golangci-lint schema validation...")

				verifyCmd := exec.Command("golangci-lint", "config", "verify", "--config", configFile)

				output, err := verifyCmd.CombinedOutput()
				if err != nil {
					logger.Errorf("❌ Schema validation failed:")
					logger.Errorf("%s", string(output))

					return fmt.Errorf("golangci-lint config verify failed: %w", err)
				}

				if len(output) > 0 {
					logger.Infof("%s", string(output))
				}

				logger.Infof("✓ Schema validation passed")
			}

			logger.Infof("✅ Configuration is valid")

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipGolangciLint, "skip-golangci-lint", false, "Skip golangci-lint schema validation")

	return cmd
}
