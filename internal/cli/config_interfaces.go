package cli

import (
	"context"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Narrow config interfaces (M30): each CLI helper declares the Loader
// methods it actually uses, so tests can mock the slice instead of
// constructing a full *config.Loader. *config.Loader satisfies every one of
// these implicitly; pass-through functions that only forward the loader
// keep the concrete type on purpose.

// configReader reads existing golangci-lint config files.
type configReader interface {
	LoadConfig(path string) (*types.Config, error)
}

// configWriter persists a golangci-lint config.
type configWriter interface {
	SaveConfig(config *types.Config, path string) error
}

// configValidator runs struct + business validation on a loaded config.
type configValidator interface {
	ValidateConfig(config *types.Config) []error
}

// configLocator resolves config file paths.
type configLocator interface {
	FindConfigFile(startDir string) (string, error)
	HasMultipleConfigFiles(startDir string) bool
}

// configLister enumerates config candidates for auto-merge.
type configLister interface {
	FindAllConfigFiles(startDir string) []string
}

// defaultConfigCreator builds a fresh default config.
type defaultConfigCreator interface {
	CreateDefaultConfig(ctx context.Context) *types.Config
}

// loadedConfigValidator loads a config and validates it.
type loadedConfigValidator interface {
	configReader
	configValidator
}

// defaultConfigWriter creates and persists a default config.
type defaultConfigWriter interface {
	defaultConfigCreator
	configWriter
}
