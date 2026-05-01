// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"bytes"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

const permUserRead = 0o644 // rw-r--r--

// LoadConfig loads a golangci-lint configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false)

	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML %s: %w", path, err)
	}

	return &config, nil
}

// SaveConfig saves a golangci-lint configuration to a YAML file.
func SaveConfig(config *Config, path string) error {
	var buf bytes.Buffer

	encoder := yaml.NewEncoder(&buf)
	//nolint:mnd // 2-space indentation is standard for YAML
	encoder.SetIndent(2)

	err := encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to encode YAML %s: %w", path, err)
	}

	err = os.WriteFile(path, buf.Bytes(), permUserRead)
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}
