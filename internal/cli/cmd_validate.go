package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"charm.land/log/v2"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
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
Use --verbose to see detailed validation output.
Use --format sarif to output results as SARIF.`),
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

	cfg, loadErr := configLoader.LoadConfig(configFile)
	if loadErr != nil {
		if reportFormat == formatSARIF {
			return outputValidationSARIF(nil, configFile, []error{loadErr})
		}

		return logAndFailLoad(logger, loadErr)
	}

	logger.Infof("✓ Basic structure valid")

	validationErrors := configLoader.ValidateConfig(cfg)
	if len(validationErrors) > 0 {
		if reportFormat == formatSARIF {
			return outputValidationSARIF(cfg, configFile, validationErrors)
		}

		return logAndFailValidation(logger, validationErrors)
	}

	logger.Infof("✓ Internal validation passed")

	if !skipGolangciLint {
		return runSchemaValidation(cmd, configFile, logger)
	}

	logger.Infof("✅ Configuration is valid")

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

func outputValidationSARIF(_ *types.Config, configFile string, errors []error) error {
	report := finding.NewReport(finding.ToolInfo{
		Name:    "golangci-lint-auto-configure",
		Version: Version,
	})

	for _, err := range errors {
		pos := finding.Position{File: configFile}
		f, buildErr := finding.NewBuilder(
			"validation-error",
			"golangci-lint-auto-configure",
			err.Error(),
			finding.SeverityError,
			pos,
		).Build()
		if buildErr != nil {
			return fmt.Errorf("failed to build validation finding: %w", buildErr)
		}

		report.AddFinding(f)
	}

	report.ComputeSummary()

	sarif, sarifErr := report.ToSARIF()
	if sarifErr != nil {
		return fmt.Errorf("failed to generate SARIF: %w", sarifErr)
	}

	var raw json.RawMessage = sarif

	pretty, prettyErr := json.MarshalIndent(raw, "", "  ")
	if prettyErr != nil {
		return fmt.Errorf("failed to format SARIF: %w", prettyErr)
	}

	fmt.Println(string(pretty))

	return nil
}
