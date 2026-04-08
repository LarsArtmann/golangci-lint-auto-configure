package cli

import (
	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/spf13/cobra"
)

// CommandBuilder helps build cobra commands with common dependencies.
type CommandBuilder struct {
	logger       *log.Logger
	analyzer     *linter.Analyzer
	configLoader *config.Loader
}

// NewCommandBuilder creates a new CommandBuilder with the common dependencies.
func NewCommandBuilder(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) *CommandBuilder {
	return &CommandBuilder{
		logger:       logger,
		analyzer:     analyzer,
		configLoader: configLoader,
	}
}

// Build creates a cobra.Command with the given options.
func (b *CommandBuilder) Build(
	use string,
	short string,
	runE func(*cobra.Command, []string) error,
	options ...func(*cobra.Command),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  runE,
	}

	for _, opt := range options {
		opt(cmd)
	}

	return cmd
}

// WithLong sets the long description for a command.
func WithLong(long string) func(*cobra.Command) {
	return func(cmd *cobra.Command) {
		cmd.Long = long
	}
}

// WithStringFlag adds a string flag to the command.
func (b *CommandBuilder) WithStringFlag(name, _, value, usage string) func(*cobra.Command) {
	return func(cmd *cobra.Command) {
		cmd.Flags().
			StringVar(new(string), name, value, usage)
	}
}

// Logger returns the builder's logger.
func (b *CommandBuilder) Logger() *log.Logger {
	return b.logger
}

// Analyzer returns the builder's analyzer.
func (b *CommandBuilder) Analyzer() *linter.Analyzer {
	return b.analyzer
}

// ConfigLoader returns the builder's config loader.
func (b *CommandBuilder) ConfigLoader() *config.Loader {
	return b.configLoader
}
