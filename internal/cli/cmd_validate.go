package cli

import (
	"fmt"
	"os/exec"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/spf13/cobra"
)

// newValidateCommand creates the validate command.
func newValidateCommand(builder *CommandBuilder) *cobra.Command {
	var skipGolangciLint bool

	cmd := builder.Build("validate", "Validate golangci-lint configuration",
		func(cmd *cobra.Command, _ []string) error {
			return runValidate(cmd, builder.Logger(), builder.ConfigLoader(), skipGolangciLint)
		},
		WithLong(`Validates the golangci-lint configuration file.

This command performs two levels of validation:
1. Basic YAML parsing and structure validation
2. Schema validation using golangci-lint config verify

Use --skip-golangci-lint to skip the schema validation (faster).
Use --verbose to see detailed validation output.`),
	)

	cmd.Flags().
		BoolVar(&skipGolangciLint, "skip-golangci-lint", false, "Skip golangci-lint schema validation")

	return cmd
}

func runValidate(
	cmd *cobra.Command,
	logger *log.Logger,
	configLoader *config.Loader,
	skipGolangciLint bool,
) error {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	configFile, err := resolveConfigPath(cmd.Context(), configLoader, logger, configPath, dryRun)
	if err != nil {
		return err
	}

	logger.Infof("Validating configuration: %s", configFile)

	err = validateBasicStructure(configLoader, configFile, logger)
	if err != nil {
		return err
	}

	if !skipGolangciLint {
		return runSchemaValidation(cmd, configFile, logger)
	}

	logger.Infof("✅ Configuration is valid")

	return nil
}

func validateBasicStructure(
	configLoader *config.Loader,
	configFile string,
	logger *log.Logger,
) error {
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return logAndFailLoad(logger, err)
	}

	logger.Infof("✓ Basic structure valid")

	validationErrors := configLoader.ValidateConfig(cfg)
	if len(validationErrors) > 0 {
		return logAndFailValidation(logger, validationErrors)
	}

	logger.Infof("✓ Internal validation passed")

	return nil
}

func logAndFailLoad(logger *log.Logger, err error) error {
	logger.Errorf("❌ Failed to load configuration")

	return fmt.Errorf("failed to load config: %w", err)
}

func logAndFailValidation(logger *log.Logger, validationErrors []error) error {
	logger.Errorf("❌ Internal validation failed:")

	for _, e := range validationErrors {
		logger.Errorf("  - %v", e)
	}

	return fmt.Errorf(
		"%w: %d validation errors",
		apperrors.ErrConfigValidationFailed,
		len(validationErrors),
	)
}

func runSchemaValidation(cmd *cobra.Command, configFile string, logger *log.Logger) error {
	logger.Infof("Running golangci-lint schema validation...")

	verifyCmd := exec.CommandContext(
		cmd.Context(),
		"golangci-lint",
		"config",
		"verify",
		"--config",
		configFile,
	)

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
	logger.Infof("✅ Configuration is valid")

	return nil
}
