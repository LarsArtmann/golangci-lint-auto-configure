// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"bytes"
	"os"

	errorfamily "github.com/larsartmann/go-error-family"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"go.yaml.in/yaml/v3"
)

const permOwnerOnly = 0o600 // rw-------

// LoadConfig loads a golangci-lint configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, apperrors.WrapClassifiedf(err, "migration.read_config",
			"failed to read config file %s", path)
	}

	var config Config

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false)

	err = decoder.Decode(&config)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "migration.parse_yaml",
			"failed to parse YAML %s", path)
	}

	return &config, nil
}

// SaveConfig saves a golangci-lint configuration to a YAML file.
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return errorfamily.WrapCorruptionf(err, "migration.encode_yaml",
			"failed to encode YAML %s", path)
	}

	err = os.WriteFile(path, data, permOwnerOnly)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "migration.write_config",
			"failed to write config file %s", path)
	}

	return nil
}
