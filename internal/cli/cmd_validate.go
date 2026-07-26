package cli

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"os/exec"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	appfinding "github.com/larsartmann/golangci-lint-auto-configure/pkg/finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/spf13/cobra"
)

func newValidateCommand(builder *CommandBuilder) *cobra.Command {
	var skipGolangciLint bool

	cmd := builder.Build(
		"validate", "Validate golangci-lint configuration",
		func(cmd *cobra.Command, _ []string) error {
			return runValidate(
				cmd,
				builder.Logger(),
				builder.ConfigLoader(),
				builder.Flags(),
				skipGolangciLint,
			)
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
	flags *Flags,
	skipGolangciLint bool,
) error {
	setLogLevel(logger, flags.Verbose)

	configFile, err := resolveValidateConfig(cmd, configLoader, logger, flags, skipGolangciLint)
	if err != nil {
		return apperrors.WrapClassified(err, "validate.resolve_config", "resolve config")
	}

	logger.Infof("Validating configuration: %s", configFile)

	err = validateConfig(configLoader, logger, configFile, flags.ReportFormat)
	if err != nil {
		return apperrors.WrapClassified(err, "validate.config", "validate config")
	}

	if skipGolangciLint {
		logger.Infof("✅ Configuration is valid")

		return nil
	}

	return runSchemaValidation(cmd, configFile, logger)
}

func resolveValidateConfig(
	cmd *cobra.Command,
	configLoader *config.Loader,
	logger *log.Logger,
	flags *Flags,
	_ bool,
) (string, error) {
	return resolveConfig(cmd.Context(), configLoader, logger, flags, "resolve config path")
}

func validateConfig(
	configLoader *config.Loader,
	logger *log.Logger,
	configFile string,
	reportFormat string,
) error {
	err := validateLoadedConfig(
		configLoader,
		logger,
		configFile,
		reportFormat,
	) //art-dupl:accept standard Go early-return idiom
	if err == nil {
		return nil
	}

	return apperrors.WrapClassified(
		err,
		"validate.loaded_config",
		"validate loaded config",
	)
}

func validateLoadedConfig(
	configLoader *config.Loader,
	logger *log.Logger,
	configFile string,
	reportFormat string,
) error {
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

	return checkConfigHealth(cfg, logger, configFile, reportFormat)
}

func checkConfigHealth(
	cfg *types.Config,
	logger *log.Logger,
	configFile string,
	reportFormat string,
) error {
	health := types.CheckConfigHealthWithCriticalLinters(cfg, constants.CriticalLinters())

	if health.IsHealthy() {
		logger.Infof("✓ Config health check passed")

		return nil
	}

	if reportFormat == formatSARIF {
		return outputHealthSARIF(health, configFile, logger)
	}

	logHealthIssues(logger, health)

	return errorfamily.WrapRejectionf(apperrors.ErrConfigValidationFailed, "validate.health",
		"%d health issues (%d critical, %d warning)",
		len(health.Issues), len(health.CriticalIssues()), len(health.WarningIssues()))
}

func logHealthIssues(logger *log.Logger, health *types.ConfigHealth) {
	logger.Warnf("⚠️  Config health check found issues:")

	for _, issue := range health.CriticalIssues() {
		logger.Errorf("  🔴 [%s] %s (%s)", issue.Severity, issue.Message, issue.Rule)
	}

	for _, issue := range health.WarningIssues() {
		logger.Warnf("  🟡 [%s] %s (%s)", issue.Severity, issue.Message, issue.Rule)
	}

	for _, issue := range health.Issues {
		if issue.Severity == types.HealthSeverityInfo {
			logger.Infof("  ℹ️  [%s] %s (%s)", issue.Severity, issue.Message, issue.Rule)
		}
	}

	logger.Warnf("  Suggestions:")

	for _, issue := range health.Issues {
		if issue.Suggestion != "" {
			logger.Warnf("    - %s: %s", issue.Rule, issue.Suggestion)
		}
	}
}

func outputHealthSARIF(health *types.ConfigHealth, configFile string, logger *log.Logger) error {
	report := finding.NewReport(finding.ToolInfo{
		Name:    constants.ToolName,
		Version: Version,
	})

	report.AddFindings(healthIssuesToFindings(health, configFile, logger))
	report.ComputeSummary()

	sarif, err := report.ToSARIF()
	if err != nil {
		return errorfamily.WrapCorruption(err, "validate.sarif_health",
			"failed to generate SARIF")
	}

	var raw jsontext.Value = sarif

	pretty, err := json.Marshal(raw, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return errorfamily.WrapCorruption(err, "validate.sarif_format_health",
			"failed to format SARIF")
	}

	_, err = os.Stdout.Write(pretty)
	if err != nil {
		return errorfamily.WrapRejection(err, "validate.sarif_write_health",
			"failed to write SARIF output")
	}

	return nil
}

func healthIssuesToFindings(
	health *types.ConfigHealth,
	configFile string,
	logger *log.Logger,
) []finding.Finding {
	result := make([]finding.Finding, 0, len(health.Issues))

	for _, issue := range health.Issues {
		pos := finding.Position{File: finding.FilePath(configFile), Line: 1}

		severity := appfinding.SeverityFromHealthSeverity(issue.Severity)

		findingObj, err := finding.NewBuilder(
			finding.RuleName(issue.Rule),
			finding.ToolName(constants.ToolName),
			issue.Message,
			severity,
			pos,
		).
			WithCategory(finding.CategoryConfiguration).
			WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion(issue.Suggestion).
			Build()
		if err != nil { //nolint:erraudit // best-effort: skip one finding that fails to build, continue emitting the rest
			logger.Warnf("⚠️  Failed to build finding for rule %q: %v", issue.Rule, err)

			continue
		}

		result = append(result, findingObj)
	}

	return result
}

func logAndFailLoad(logger *log.Logger, err error) error {
	logger.Errorf("❌ Failed to load configuration")

	return apperrors.WrapClassified(err, "validate.load_config", "failed to load config")
}

func logAndFailValidation(logger *log.Logger, validationErrors []error) error {
	logger.Errorf("❌ Internal validation failed:")

	for _, e := range validationErrors {
		logger.Errorf("  - %v", e)
	}

	return errorfamily.WrapRejectionf(apperrors.ErrConfigValidationFailed, "validate.internal",
		"%d validation errors", len(validationErrors))
}

func runSchemaValidation(cmd *cobra.Command, configFile string, logger *log.Logger) error {
	logger.Infof("Running golangci-lint schema validation...")

	verifyCmd := exec.CommandContext( //nolint:gosec // configFile is resolved and validated by config loader before reaching this point
		cmd.Context(),
		constants.GolangciLintBinaryName,
		"config",
		"verify",
		"--config",
		configFile,
	)

	output, err := verifyCmd.CombinedOutput()
	if err != nil {
		logger.Errorf("❌ Schema validation failed:")
		logger.Errorf("%s", string(output))

		return apperrors.WrapClassified(err, "validate.schema_verify",
			"golangci-lint config verify failed")
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
		Name:    constants.ToolName,
		Version: Version,
	})

	findings, err := appfinding.ErrorsToFindings(errors, configFile)
	if err != nil {
		return apperrors.WrapClassified(err, "validate.errors_to_findings",
			"failed to convert errors to findings")
	}

	report.AddFindings(findings)
	report.ComputeSummary()

	sarif, sarifErr := report.ToSARIF()
	if sarifErr != nil {
		return errorfamily.WrapCorruptionf(sarifErr, "validate.sarif",
			"failed to generate SARIF")
	}

	var raw jsontext.Value = sarif

	pretty, prettyErr := json.Marshal(raw, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if prettyErr != nil {
		return errorfamily.WrapCorruptionf(prettyErr, "validate.sarif_format",
			"failed to format SARIF")
	}

	_, writeErr := os.Stdout.Write(pretty) //nolint:erraudit // _ is the byte count, not an error; writeErr is checked below
	if writeErr != nil {
		return errorfamily.WrapRejectionf(writeErr, "validate.sarif_write",
			"failed to write SARIF output")
	}

	return nil
}
