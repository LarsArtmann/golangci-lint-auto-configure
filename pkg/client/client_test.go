package client_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/client"
)

func TestNewCreatesValidClient(t *testing.T) {
	c := client.New(client.Options{})

	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	c := client.New(client.Options{})

	cfg, err := c.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != "2" {
		t.Errorf("expected version %q, got %q", "2", cfg.Version)
	}
}

func TestValidateConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	c := client.New(client.Options{})

	cfg, err := c.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	errs := c.ValidateConfig(cfg)

	if len(errs) > 0 {
		t.Errorf("expected no validation errors for valid config, got %d", len(errs))
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	c := client.New(client.Options{})

	cfg, err := c.LoadConfig(writeTempConfig(t, dir))
	if err != nil {
		t.Fatal(err)
	}

	err = c.SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestSimpleFixDryRun(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".golangci.yml")

	err := os.WriteFile(configPath, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.SimpleFix(context.Background(), client.Options{}, configPath, true)
	if err != nil {
		t.Skipf("skipping SimpleFix test (requires golangci-lint): %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result")
	}
}

func writeTempConfig(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, "input.yml")

	err := os.WriteFile(path, []byte(`version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	return path
}
